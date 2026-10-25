package acceptance_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var binary string
var agentDirectory string

func TestMain(tests *testing.M) {
	directory, err := os.MkdirTemp("", "repo-tutor-acceptance-*")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(directory, "repo-tutor")
	fake := filepath.Join(directory, "fake-agent")
	for _, build := range []struct{ target, source string }{{binary, "../binary"}, {fake, "./testdata/agent"}} {
		if output, err := exec.Command("go", "build", "-o", build.target, build.source).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "build acceptance executable: %v\n%s", err, output)
			os.RemoveAll(directory)
			os.Exit(1)
		}
	}
	agentDirectory = filepath.Join(directory, "agents")
	os.MkdirAll(agentDirectory, 0700)
	for _, name := range []string{"claude", "codex"} {
		if err := os.Symlink(fake, filepath.Join(agentDirectory, name)); err != nil {
			panic(err)
		}
	}
	status := tests.Run()
	os.RemoveAll(directory)
	os.Exit(status)
}

func fixture(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	for name, data := range map[string]string{"go.mod": "module learning\n", "main.go": "package main\nfunc main() {}\n", "README.md": "# Learning\nA test repository.\n"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func run(t *testing.T, input string, arguments ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Stdin = strings.NewReader(input)
	command.Env = append(os.Environ(), "PATH="+agentDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := command.CombinedOutput()
	return string(output), err
}

func saved(t *testing.T, base string) map[string]any {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(base, "*", "session.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one saved session: %v %v", files, err)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestTerminalCourseWithBothAgents(t *testing.T) {
	for _, name := range []string{"claude", "codex"} {
		t.Run(name, func(t *testing.T) {
			repository, stateDirectory := fixture(t), t.TempDir()
			arguments := []string{repository, "--agent", name, "--state-dir", stateDirectory}
			if name == "claude" {
				arguments = append(arguments, "--plain")
			}
			output, err := run(t, "/topics\n/quiz\nIt reads files.\n/progress\n/read main.go:1\n/next\n/quit\n", arguments...)
			if err != nil {
				t.Fatalf("session failed: %v\n%s", err, output)
			}
			for _, expected := range []string{"Question:", "Assessment: partial", "1 partial", "main.go:2  func main()", "Session saved"} {
				if !strings.Contains(output, expected) {
					t.Fatalf("missing %q in terminal output:\n%s", expected, output)
				}
			}
			state := saved(t, stateDirectory)
			if state["topic"] != "learning" || len(state["turns"].([]any)) != 4 {
				t.Fatalf("wrong saved progress: %+v", state)
			}
			output, err = run(t, "/quit\n", arguments...)
			if err != nil || !strings.Contains(output, "Resuming your last turn") || strings.Contains(output, "Waiting for") {
				t.Fatalf("resume made an unnecessary agent call: %v\n%s", err, output)
			}
			files, _ := filepath.Glob(filepath.Join(stateDirectory, "*", "agent", "invocations.jsonl"))
			data, _ := os.ReadFile(files[0])
			calls := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(calls) != 4 || !strings.Contains(calls[1], "11111111-1111-4111-8111-111111111111") || strings.Contains(calls[0], "11111111-1111-4111-8111-111111111111") {
				t.Fatalf("explicit session continuation was not used: %s", data)
			}
		})
	}
}

func TestInvalidOrFailedTurnDoesNotChangeProgress(t *testing.T) {
	repository, stateDirectory := fixture(t), t.TempDir()
	arguments := []string{repository, "--agent", "codex", "--state-dir", stateDirectory}
	if output, err := run(t, "/quit\n", arguments...); err != nil {
		t.Fatalf("initial lesson: %v\n%s", err, output)
	}
	for _, question := range []string{"invent a citation", "simulate failure", "wait forever"} {
		output, err := run(t, "", append(arguments, "--timeout", "50ms", "--ask", question)...)
		if err == nil {
			t.Fatalf("invalid turn succeeded: %s\n%s", question, output)
		}
		if len(saved(t, stateDirectory)["turns"].([]any)) != 1 {
			t.Fatal("failed turn changed the saved lesson")
		}
	}
}

func TestQuizGoalWaitsForContinueAndSurvivesResume(t *testing.T) {
	repository, stateDirectory := fixture(t), t.TempDir()
	arguments := []string{repository, "--agent", "codex", "--state-dir", stateDirectory, "--plain"}
	input := "/quiz\nFirst attempt\n" + strings.Repeat("/continue\nAnother attempt\n", 4) + "/continue\n/questions\n/quit\n"
	output, err := run(t, input, arguments...)
	if err != nil || !strings.Contains(output, "QUIZ · 5 / 5 answered") || !strings.Contains(output, "Goal reached") {
		t.Fatalf("five-question goal failed: %v\n%s", err, output)
	}
	state := saved(t, stateDirectory)
	if state["review"] != true || state["question"] != "" || len(state["turns"].([]any)) != 11 {
		t.Fatalf("quiz advanced automatically or did not save review: %+v", state)
	}
	output, err = run(t, "/questions\n/quit\n", arguments...)
	if err != nil || strings.Contains(output, "Waiting for") || !strings.Contains(output, "QUIZ · 5 / 5 answered") {
		t.Fatalf("resume lost the quiz goal or called the agent: %v\n%s", err, output)
	}
	output, err = run(t, "Can you clarify that answer?\n/quit\n", arguments...)
	if err != nil || saved(t, stateDirectory)["review"] != true || saved(t, stateDirectory)["question"] != "" {
		t.Fatalf("follow-up question advanced past review: %v\n%s", err, output)
	}
	output, err = run(t, "/extend\n/questions\n/quit\n", arguments...)
	if err != nil || !strings.Contains(output, "QUIZ · 5 / 10 answered") {
		t.Fatalf("quiz extension lost outcomes: %v\n%s", err, output)
	}
	state = saved(t, stateDirectory)
	quiz := state["quiz"].(map[string]any)
	if quiz["goal"] != float64(10) || len(quiz["questions"].([]any)) != 6 || len(state["turns"].([]any)) != 13 {
		t.Fatalf("extension should generate only the sixth question: %+v", state)
	}
}

func TestChangedSourceStartsFreshAndArchivesProgress(t *testing.T) {
	repository, stateDirectory := fixture(t), t.TempDir()
	arguments := []string{repository, "--agent", "codex", "--state-dir", stateDirectory, "--goal", "Trace the entry point"}
	if output, err := run(t, "/quit\n", arguments...); err != nil {
		t.Fatalf("initial lesson: %v\n%s", err, output)
	}
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("# Updated course\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := run(t, "/quit\n", repository, "--state-dir", stateDirectory)
	if err != nil || !strings.Contains(output, "Repository sources changed") {
		t.Fatalf("source changes were not detected: %v\n%s", err, output)
	}
	files, _ := filepath.Glob(filepath.Join(stateDirectory, "*", "previous.json"))
	if len(files) != 1 || saved(t, stateDirectory)["goal"] != "Trace the entry point" {
		t.Fatal("source refresh lost the previous session or learning goal")
	}
}

func TestScanNeedsNoAgent(t *testing.T) {
	output, err := run(t, "", "scan", fixture(t), "--json")
	if err != nil || !json.Valid([]byte(output)) || !strings.Contains(output, `"entry_points"`) || strings.Contains(output, "Waiting for") {
		t.Fatalf("offline scan failed: %v\n%s", err, output)
	}
}

func TestInterruptAtPromptPreservesCompletedLesson(t *testing.T) {
	repository, stateDirectory := fixture(t), t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, repository, "--agent", "codex", "--state-dir", stateDirectory)
	command.Env = append(os.Environ(), "PATH="+agentDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	ready := make(chan error, 1)
	go func() {
		_, err := bufio.NewReader(output).ReadString('>')
		ready <- err
	}()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("terminal never reached its prompt")
	}
	if err := command.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	command.Wait()
	if ctx.Err() != nil {
		t.Fatal("Ctrl-C did not exit while waiting for terminal input")
	}
	if len(saved(t, stateDirectory)["turns"].([]any)) != 1 {
		t.Fatal("Ctrl-C lost the completed lesson")
	}
}
