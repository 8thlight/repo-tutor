package repository_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"repo-tutor/internal/repository"
)

func TestReadingHintsUseChangesAndBranchPoints(t *testing.T) {
	snapshot := repository.Analyze([]repository.File{
		{Path: "go.mod", Content: "module learning\n"},
		{Path: "busy.go", Content: "package learning\nfunc Busy(values []bool) {\n for _, value := range values {\n  if value || len(values) == 1 { println(value) }\n }\n}\n"},
		{Path: "quiet.go", Content: "package learning\nfunc Quiet() {}\n"},
		{Path: "busy_test.go", Content: "package learning\nfunc Example() { if true {} }\n"},
	})
	changes := repository.ChangeCounts("busy.go\x00quiet.go\x00busy.go\x00busy_test.go\x00busy.go\x00")
	if !reflect.DeepEqual(changes, map[string]int{"busy.go": 3, "quiet.go": 1, "busy_test.go": 1}) {
		t.Fatalf("wrong change counts: %v", changes)
	}
	hints := repository.ReadingHints(snapshot, changes)
	if len(hints) != 2 || hints[0].Path != "busy.go" || hints[0].Commits != 3 || hints[0].Branches != 3 || hints[0].Line != 3 || hints[0].Score != 16 {
		t.Fatalf("wrong reading hints: %+v", hints)
	}
	withoutHistory := repository.ReadingHints(snapshot, nil)
	if len(withoutHistory) != 1 || withoutHistory[0].Path != "busy.go" || withoutHistory[0].Commits != 0 {
		t.Fatalf("missing history was treated as measured churn: %+v", withoutHistory)
	}
}

func TestHistoryCountsLocalCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is optional")
	}
	root := t.TempDir()
	git := func(arguments ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-c", "user.name=Tutor Test", "-c", "user.email=tutor@example.invalid", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"}, arguments...)...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}
	git("init", "-q")
	for _, name := range []string{"alpha.py", "beta.py"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("if True:\n    pass\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-qm", "initial source")
	if err := os.WriteFile(filepath.Join(root, "alpha.py"), []byte("if False:\n    pass\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-qm", "change alpha")
	snapshot, err := repository.Read(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.History || len(snapshot.Hotspots) != 2 || snapshot.Hotspots[0].Path != "alpha.py" || snapshot.Hotspots[0].Commits != 2 || snapshot.Hotspots[1].Commits != 1 {
		t.Fatalf("local Git history did not reach the reading hints: %+v", snapshot.Hotspots)
	}
}
