package tutor_test

import (
	"fmt"
	"strings"
	"testing"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/tutor"
)

func fileSnapshot() repository.Snapshot {
	var big strings.Builder
	big.WriteString("package big\n")
	for index := range 2000 {
		fmt.Fprintf(&big, "var value%d = %d // a line long enough to fill windows\n", index, index)
	}
	return repository.Analyze([]repository.File{
		{Path: "go.mod", Content: "module course\n"},
		{Path: "README.md", Content: "# Course\n"},
		{Path: "small/small.go", Content: "package small\nfunc Tiny() {}\n"},
		{Path: "big/big.go", Content: big.String()},
	})
}

func TestAFileTopicIdRoundTripsAndStaysOutsidePackagePaths(t *testing.T) {
	id := tutor.FileTopic("internal/tutor/turn.go")
	path, ok := tutor.FilePath(id)
	if !ok || path != "internal/tutor/turn.go" || id[0] != ':' {
		t.Fatalf("got %q %q %v", id, path, ok)
	}
	if _, isFile := tutor.FilePath("internal/tutor"); isFile {
		t.Fatal("a package path is not a file topic")
	}
}

func TestAShortFileArrivesWholeAndFirst(t *testing.T) {
	prepared := tutor.Prepare(fileSnapshot(), tutor.Request{Topic: tutor.FileTopic("small/small.go"), Action: "lesson"})
	first := prepared.Excerpts[0]
	if first.Path != "small/small.go" || first.StartLine != 1 || first.EndLine != 2 {
		t.Fatalf("the file should open the excerpts and arrive whole: %+v", first)
	}
}

func TestALongFileArrivesInConsecutiveWindowsUpToTheCap(t *testing.T) {
	prepared := tutor.Prepare(fileSnapshot(), tutor.Request{Topic: tutor.FileTopic("big/big.go"), Action: "lesson"})
	var windows []tutor.Excerpt
	for _, excerpt := range prepared.Excerpts {
		if excerpt.Path == "big/big.go" {
			windows = append(windows, excerpt)
		}
	}
	if len(windows) != 6 {
		t.Fatalf("a long file should be taught from six windows, got %d", len(windows))
	}
	for index := 1; index < len(windows); index++ {
		if windows[index].StartLine != windows[index-1].EndLine+1 {
			t.Fatalf("windows should be consecutive: %d ends at %d, next starts at %d", index-1, windows[index-1].EndLine, windows[index].StartLine)
		}
	}
	if prepared.Excerpts[0].Path != "big/big.go" {
		t.Fatal("the file should come before every other source")
	}
	total := 0
	for _, excerpt := range prepared.Excerpts {
		total += len(excerpt.Text)
	}
	if total > 48000 {
		t.Fatalf("the source budget still applies: %d", total)
	}
}

func TestOtherTopicsDoNotLeadWithAFile(t *testing.T) {
	prepared := tutor.Prepare(fileSnapshot(), tutor.Request{Topic: tutor.Overview, Action: "lesson"})
	if prepared.Excerpts[0].Path == "big/big.go" {
		t.Fatal("only a file topic opens with its file")
	}
}

func TestThePromptExplainsAFileTopic(t *testing.T) {
	prompt := tutor.Prepare(fileSnapshot(), tutor.Request{Topic: tutor.FileTopic("small/small.go"), Action: "lesson"}).Prompt
	if !strings.Contains(prompt, "begins with :file:") || !strings.Contains(prompt, "names one source file") {
		t.Fatal("the prompt should say how to teach a file topic")
	}
}

func TestNextFromAFileTopicIsTheFirstUnreadCourseTopic(t *testing.T) {
	course := []tutor.Topic{{ID: tutor.Overview, Title: "Overview"}, {ID: "a", Title: "A"}, {ID: "b", Title: "B"}, {ID: "c", Title: "C"}}
	progress := map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, "a": {Lessons: 2}, "b": {Held: true}}
	next, found := tutor.NextUnheld(course, progress, tutor.FileTopic("x.go"))
	if !found || next.ID != "c" {
		t.Fatalf("a file topic is not in the course, so next is the first unread topic, got %+v %v", next, found)
	}
	if next, found := tutor.NextUnheld(course, progress, "a"); !found || next.ID != "c" {
		t.Fatalf("a course topic still advances in order and skips known ones: %+v", next)
	}
	allRead := map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, "a": {Lessons: 1}, "b": {Lessons: 1}, "c": {Lessons: 1}}
	if _, found := tutor.NextUnheld(course, allRead, tutor.FileTopic("x.go")); found {
		t.Fatal("with every topic read there is nothing next")
	}
}
