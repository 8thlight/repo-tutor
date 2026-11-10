package analyzer_test

import (
	"reflect"
	"testing"

	"repo-tutor/internal/repository/analyzer"
)

func TestReadingHintsUseChangesAndBranchPoints(t *testing.T) {
	snapshot := analyzer.Analyze([]analyzer.File{
		{Path: "go.mod", Content: "module learning\n"},
		{Path: "busy.go", Content: "package learning\nfunc Busy(values []bool) {\n for _, value := range values {\n  if value || len(values) == 1 { println(value) }\n }\n}\n"},
		{Path: "quiet.go", Content: "package learning\nfunc Quiet() {}\n"},
		{Path: "busy_test.go", Content: "package learning\nfunc Example() { if true {} }\n"},
	})
	changes := analyzer.ChangeCounts("busy.go\x00quiet.go\x00busy.go\x00busy_test.go\x00busy.go\x00")
	if !reflect.DeepEqual(changes, map[string]int{"busy.go": 3, "quiet.go": 1, "busy_test.go": 1}) {
		t.Fatalf("wrong change counts: %v", changes)
	}
	hints := analyzer.ReadingHints(snapshot, changes)
	if len(hints) != 2 || hints[0].Path != "busy.go" || hints[0].Commits != 3 || hints[0].Branches != 3 || hints[0].Line != 3 || hints[0].Score != 16 {
		t.Fatalf("wrong reading hints: %+v", hints)
	}
	withoutHistory := analyzer.ReadingHints(snapshot, nil)
	if len(withoutHistory) != 1 || withoutHistory[0].Path != "busy.go" || withoutHistory[0].Commits != 0 {
		t.Fatalf("missing history was treated as measured churn: %+v", withoutHistory)
	}
}
