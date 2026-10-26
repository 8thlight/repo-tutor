package tutor

import (
	"strings"
	"testing"

	"repo-tutor/internal/repository"
)

func TestLessonRoundTripsAsMarkdownAndRejectsDamage(t *testing.T) {
	reply := Reply{Explanation: "# Map\n\nThe CLI starts at main.go:2. A note about --> in text.", Question: "What starts it?", Assessment: "none", Citations: []repository.Location{{Path: "main.go", Line: 2}}}
	overview := SavedLesson{Fingerprint: "abc", Goal: GoalKey("Trace the entry point"), Topic: "core", Reply: reply}
	data := EncodeLesson(overview)
	if !strings.HasSuffix(string(data), "A note about --> in text.\n") {
		t.Fatalf("the explanation should be readable Markdown after the header:\n%s", data)
	}
	decoded, ok := DecodeLesson(data)
	if !ok || decoded.Fingerprint != "abc" || decoded.Topic != "core" || decoded.Reply.Explanation != reply.Explanation || decoded.Reply.Question != reply.Question || decoded.Reply.Citations[0] != reply.Citations[0] {
		t.Fatalf("overview did not survive a round trip: %+v", decoded)
	}
	for name, damaged := range map[string][]byte{"empty": nil, "no header": []byte("# Map\n"), "bad json": []byte("<!-- repo-tutor lesson {nope} -->\n\nText\n"), "no body": []byte(strings.Split(string(data), "-->")[0] + "-->\n\n")} {
		if _, ok := DecodeLesson(damaged); ok {
			t.Fatalf("%s should count as a cache miss", name)
		}
	}
	if !decoded.Fresh("abc") || decoded.Fresh("other") || GoalKey("a") == GoalKey("b") || LessonKey("a", "x") == LessonKey("a", "y") || LessonKey("a", "x") == LessonKey("b", "x") {
		t.Fatal("freshness must follow the fingerprint and the goal key must differ per goal")
	}
}
