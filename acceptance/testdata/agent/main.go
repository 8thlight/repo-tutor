package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	text := string(data)
	index := strings.Index(text, "Turn and source excerpts:\n")
	if index < 0 {
		fmt.Fprintln(os.Stderr, "missing source context")
		os.Exit(1)
	}
	var request struct {
		Action   string `json:"action"`
		Message  string `json:"learner_message"`
		Question string `json:"pending_question"`
	}
	if err := json.Unmarshal([]byte(text[index+len("Turn and source excerpts:\n"):]), &request); err != nil {
		panic(err)
	}
	if request.Message == "wait forever" {
		os.WriteFile(os.Getenv("FAKE_AGENT_PID_FILE"), []byte(strconv.Itoa(os.Getpid())), 0600)
		time.Sleep(time.Minute)
	}
	if request.Message == "simulate failure" {
		fmt.Fprintln(os.Stderr, "test account needs login")
		os.Exit(1)
	}
	assessment, feedback := "none", ""
	question := "What does the entry point delegate?"
	if request.Action == "answer" {
		if request.Question == "" {
			panic("answer without a question")
		}
		assessment, feedback = "partial", "Find the delegation in main.go:2."
		if request.Message == "that is correct" {
			assessment = "correct"
		}
		if request.Message == "Can you clarify the question?" {
			assessment, feedback = "none", ""
			question = "What does main.go:2 do when the program starts?"
		}
	}
	line := 2
	if request.Message == "invent a citation" {
		line = 9999
	}
	reply := map[string]any{
		"explanation": "Execution starts at main.go:2.",
		"question":    question, "assessment": assessment,
		"feedback": feedback, "citations": []map[string]any{{"path": citedPath(), "line": line}},
	}
	if os.Getenv("FAKE_AGENT_SCENARIO") == "screenshots" {
		reply["explanation"] = screenshotLesson
		reply["question"] = "Which function would you read first to see how the command line is parsed?"
		reply["citations"] = []map[string]any{{"path": "binary/main.go", "line": 15}, {"path": "README.md", "line": 1}}
	}
	arguments, _ := json.Marshal(os.Args[1:])
	file, err := os.OpenFile("invocations.jsonl", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(file, string(arguments))
	file.Close()
	encoder := json.NewEncoder(os.Stdout)
	id := "11111111-1111-4111-8111-111111111111"
	if filepath.Base(os.Args[0]) == "claude" {
		encoder.Encode(map[string]any{"type": "system", "subtype": "init", "session_id": id})
		encoder.Encode(map[string]any{"type": "assistant", "message": map[string]any{"content": []any{}}})
		encoder.Encode(map[string]any{"type": "result", "subtype": "success", "session_id": id, "structured_output": reply})
	} else {
		encoded, _ := json.Marshal(reply)
		encoder.Encode(map[string]any{"type": "thread.started", "thread_id": id})
		encoder.Encode(map[string]any{"type": "item.completed", "item": map[string]any{"type": "agent_message", "text": string(encoded)}})
		encoder.Encode(map[string]any{"type": "turn.completed"})
	}
}

func citedPath() string {
	if path := os.Getenv("FAKE_AGENT_CITED_PATH"); path != "" {
		return path
	}
	return "main.go"
}

const screenshotLesson = "Run `repo-tutor` on a repository. It reads the files on your machine, builds a course, and opens a full-screen workspace.\n\n- `binary/main.go` is the entry point. It listens for `SIGINT`, `SIGTERM`, and `SIGHUP`, then hands control to `terminal.Run`.\n- `terminal.Run` picks a command: `scan`, `doctor`, or the tutor itself.\n- The tool never calls a model API. It runs your installed agent CLI headless and asks it for one small lesson at a time.\n\nEvery claim cites a file and a line, so you can open the source and check it."
