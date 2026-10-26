package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		"feedback": feedback, "citations": []map[string]any{{"path": "main.go", "line": line}},
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
