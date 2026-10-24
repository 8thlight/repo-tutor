package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"repo-tutor/internal/tutor"
)

type Request struct {
	Agent     string
	Model     string
	Workspace string
	SessionID string
	Prompt    string
}

type Result struct {
	SessionID string
	Content   string
}

func Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func Choose(preference string) (string, error) {
	if preference != "auto" {
		if preference != "claude" && preference != "codex" {
			return "", fmt.Errorf("unknown agent %q; choose claude, codex, or auto", preference)
		}
		if !Available(preference) {
			return "", fmt.Errorf("%s is not on PATH; install and sign in to that CLI, or choose another with --agent", preference)
		}
		return preference, nil
	}
	for _, name := range []string{"codex", "claude"} {
		if Available(name) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no agent CLI found; install and sign in to Claude Code or Codex, then retry (offline scanning works with repo-tutor scan)")
}

func Run(ctx context.Context, request Request) (Result, error) {
	if err := os.MkdirAll(request.Workspace, 0700); err != nil {
		return Result{}, err
	}
	arguments, err := Arguments(request)
	if err != nil {
		return Result{}, err
	}
	if request.Agent == "codex" {
		if err := os.WriteFile(filepath.Join(request.Workspace, "reply.schema.json"), []byte(tutor.ReplySchema), 0600); err != nil {
			return Result{}, err
		}
	}
	command := exec.CommandContext(ctx, request.Agent, arguments...)
	command.Dir = request.Workspace
	command.Stdin = strings.NewReader(request.Prompt)
	command.WaitDelay = 2 * time.Second
	var diagnostics limitedBuffer
	command.Stderr = &diagnostics
	output, err := command.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	if err := command.Start(); err != nil {
		return Result{}, fmt.Errorf("start %s: %w", request.Agent, err)
	}
	result, decodeError := readEvents(output, request.Agent)
	if decodeError != nil {
		command.Process.Kill()
	}
	waitError := command.Wait()
	if ctx.Err() != nil {
		return Result{}, fmt.Errorf("%s turn interrupted: %w", request.Agent, ctx.Err())
	}
	if waitError != nil {
		message := strings.TrimSpace(tutor.Plain(diagnostics.String()))
		if decodeError != nil {
			message = decodeError.Error() + "\n" + message
		}
		if message == "" {
			message = "check the CLI login and model access in your terminal"
		}
		return Result{}, fmt.Errorf("%s failed (%v): %s", request.Agent, waitError, message)
	}
	if decodeError != nil {
		return Result{}, fmt.Errorf("%s response: %w", request.Agent, decodeError)
	}
	return result, nil
}

func Arguments(request Request) ([]string, error) {
	var arguments []string
	switch request.Agent {
	case "codex":
		arguments = []string{
			"-c", `approval_policy="never"`,
			"-c", `sandbox_mode="read-only"`,
			"-c", "features.shell_tool=false",
			"-c", "features.multi_agent=false",
			"-c", "features.multi_agent_v2=false",
			"-c", `web_search="disabled"`,
			"-c", "project_doc_max_bytes=0",
			"exec",
		}
		if request.SessionID != "" {
			arguments = append(arguments, "resume")
		}
		arguments = append(arguments, "--ignore-user-config", "--ignore-rules", "--skip-git-repo-check", "--json", "--output-schema", filepath.Join(request.Workspace, "reply.schema.json"))
		if request.Model != "" {
			arguments = append(arguments, "--model", request.Model)
		}
		if request.SessionID != "" {
			arguments = append(arguments, request.SessionID)
		}
		arguments = append(arguments, "-")
	case "claude":
		arguments = []string{
			"--safe-mode", "-p", "--output-format", "stream-json", "--verbose",
			"--tools", "", "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`,
			"--setting-sources", "", "--settings", `{"disableAllHooks":true}`,
			"--permission-mode", "dontAsk", "--json-schema", tutor.ReplySchema,
		}
		if request.Model != "" {
			arguments = append(arguments, "--model", request.Model)
		}
		if request.SessionID != "" {
			arguments = append(arguments, "--resume", request.SessionID)
		}
	default:
		return nil, fmt.Errorf("unsupported agent %q", request.Agent)
	}
	return arguments, nil
}

type event struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	SessionID        string          `json:"session_id"`
	ThreadID         string          `json:"thread_id"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output"`
	IsError          bool            `json:"is_error"`
	Message          json.RawMessage `json:"message"`
	Errors           []string        `json:"errors"`
	Error            json.RawMessage `json:"error"`
	Item             struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item"`
}

func readEvents(reader io.Reader, name string) (Result, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	var result Result
	completed := false
	var failure error
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var current event
		if err := json.Unmarshal(scanner.Bytes(), &current); err != nil {
			return Result{}, fmt.Errorf("invalid %s event stream: %w", name, err)
		}
		if current.SessionID != "" {
			result.SessionID = current.SessionID
		}
		switch current.Type {
		case "thread.started":
			result.SessionID = current.ThreadID
		case "item.completed":
			if current.Item.Type == "agent_message" {
				result.Content = current.Item.Text
			}
		case "turn.completed":
			completed = true
		case "turn.failed", "error":
			var message string
			if len(current.Message) > 0 {
				if err := json.Unmarshal(current.Message, &message); err != nil {
					message = string(current.Message)
				}
			}
			if len(current.Error) > 0 {
				message += " " + string(current.Error)
			}
			failure = fmt.Errorf("agent reported an error: %s", tutor.Plain(message))
		case "result":
			if current.IsError || current.Subtype != "success" {
				outcome := current.Subtype
				if current.IsError && outcome == "success" {
					outcome = "a failed turn"
				}
				failure = fmt.Errorf("agent reported %s: %s %s", outcome, tutor.Plain(current.Result), strings.Join(current.Errors, "; "))
				continue
			}
			completed = true
			if len(current.StructuredOutput) > 0 && string(current.StructuredOutput) != "null" {
				result.Content = string(current.StructuredOutput)
			} else {
				result.Content = current.Result
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if failure != nil {
		return Result{}, failure
	}
	if !completed || result.Content == "" || result.SessionID == "" {
		return Result{}, errors.New("agent ended without a complete reply and session ID; check your CLI version and login")
	}
	return result, nil
}

type limitedBuffer struct {
	bytes.Buffer
}

func (buffer *limitedBuffer) Write(data []byte) (int, error) {
	length := len(data)
	remaining := 8000 - buffer.Len()
	if remaining > 0 {
		buffer.Buffer.Write(data[:min(remaining, length)])
	}
	return length, nil
}
