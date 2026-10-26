package tutor

import (
	"strings"
	"testing"

	"repo-tutor/internal/repository"
)

func TestOverviewRoundTripsAsMarkdownAndRejectsDamage(t *testing.T) {
	reply := Reply{Explanation: "# Map\n\nThe CLI starts at main.go:2. A note about --> in text.", Question: "What starts it?", Assessment: "none", Citations: []repository.Location{{Path: "main.go", Line: 2}}}
	overview := SavedOverview{Fingerprint: "abc", Goal: GoalKey("Trace the entry point"), Reply: reply}
	data := EncodeOverview(overview)
	if !strings.HasSuffix(string(data), "A note about --> in text.\n") {
		t.Fatalf("the explanation should be readable Markdown after the header:\n%s", data)
	}
	decoded, ok := DecodeOverview(data)
	if !ok || decoded.Fingerprint != "abc" || decoded.Reply.Explanation != reply.Explanation || decoded.Reply.Question != reply.Question || decoded.Reply.Citations[0] != reply.Citations[0] {
		t.Fatalf("overview did not survive a round trip: %+v", decoded)
	}
	for name, damaged := range map[string][]byte{"empty": nil, "no header": []byte("# Map\n"), "bad json": []byte("<!-- repo-tutor overview {nope} -->\n\nText\n"), "no body": []byte(strings.Split(string(data), "-->")[0] + "-->\n\n")} {
		if _, ok := DecodeOverview(damaged); ok {
			t.Fatalf("%s should count as a cache miss", name)
		}
	}
	if !decoded.Fresh("abc") || decoded.Fresh("other") || GoalKey("a") == GoalKey("b") {
		t.Fatal("freshness must follow the fingerprint and the goal key must differ per goal")
	}
}
