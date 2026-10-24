package agent

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestEventProtocols(t *testing.T) {
	lesson := `{"explanation":"Start at main.go:2.","question":"Why?","assessment":"none","feedback":"","citations":[{"path":"main.go","line":2}]}`
	quoted, _ := json.Marshal(lesson)
	for _, test := range []struct {
		name   string
		agent  string
		stream string
		valid  bool
	}{
		{"codex completed", "codex", `{"type":"thread.started","thread_id":"saved-id"}` + "\n" + `{"type":"item.completed","item":{"type":"agent_message","text":` + string(quoted) + `}}` + "\n" + `{"type":"turn.completed","usage":{"input_tokens":1}}`, true},
		{"claude object messages", "claude", `{"type":"system","subtype":"init","session_id":"saved-id"}` + "\n" + `{"type":"assistant","message":{"content":[{"type":"text","text":"Working"}]}}` + "\n" + `{"type":"result","subtype":"success","session_id":"saved-id","structured_output":` + lesson + `}`, true},
		{"claude fallback result", "claude", `{"type":"result","subtype":"success","session_id":"saved-id","result":` + string(quoted) + `}`, true},
		{"claude spend limit", "claude", `{"type":"result","subtype":"success","is_error":true,"session_id":"saved-id","result":"spend limit"}`, false},
		{"codex failure", "codex", `{"type":"thread.started","thread_id":"saved-id"}` + "\n" + `{"type":"turn.failed","error":{"message":"rate limit"}}`, false},
		{"truncated stream", "codex", `{"type":"thread.started","thread_id":"saved-id"}`, false},
		{"missing session", "claude", `{"type":"result","subtype":"success","structured_output":` + lesson + `}`, false},
		{"non-json output", "codex", "Sign in first", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := readEvents(strings.NewReader(test.stream), test.agent)
			if (err == nil) != test.valid {
				t.Fatalf("expected valid=%v, got %v", test.valid, err)
			}
			if test.valid && (result.SessionID != "saved-id" || result.Content != lesson) {
				t.Fatalf("wrong normalized reply: %+v", result)
			}
		})
	}
}

func TestCommandArgumentsPreserveIsolationAndExplicitResume(t *testing.T) {
	for _, name := range []string{"claude", "codex"} {
		request := Request{Agent: name, Workspace: "/a path/with spaces", SessionID: "saved-id", Model: "model; touch unwanted"}
		arguments, err := Arguments(request)
		if err != nil {
			t.Fatal(err)
		}
		contains := func(value string) bool {
			for _, argument := range arguments {
				if argument == value {
					return true
				}
			}
			return false
		}
		if !contains("saved-id") || !contains(request.Model) || contains("--last") || contains("--dangerously-bypass-approvals-and-sandbox") {
			t.Fatalf("unsafe or ambiguous arguments: %q", arguments)
		}
		if name == "codex" && (!contains(`sandbox_mode="read-only"`) || !contains("--ignore-user-config") || !contains("features.shell_tool=false") || !contains("features.multi_agent=false")) {
			t.Fatalf("Codex isolation missing: %q", arguments)
		}
		if name == "claude" && (!contains("--safe-mode") || !contains("--tools") || !contains("--strict-mcp-config")) {
			t.Fatalf("Claude isolation missing: %q", arguments)
		}
		before := request
		Arguments(request)
		if !reflect.DeepEqual(before, request) {
			t.Fatal("argument building mutated its request")
		}
	}
}
