package tutor_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/tutor"
)

func TestReplyBoundary(t *testing.T) {
	prepared := tutor.Context{Excerpts: []tutor.Excerpt{{Path: "main.go", StartLine: 2, EndLine: 5}}}
	request := tutor.Request{Topic: "overview", Action: "answer", Question: "Where does it start?"}
	valid := tutor.Reply{Explanation: "Execution begins in main.go:3.", Question: "What is delegated?", Assessment: "correct", Feedback: "You found the entry point.", Citations: []repository.Location{{Path: "main.go", Line: 3}}}
	for _, test := range []struct {
		name   string
		change func(*tutor.Reply, *tutor.Request)
		valid  bool
	}{
		{name: "supported answer", valid: true},
		{name: "invented file", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Citations[0].Path = "imaginary.go" }},
		{name: "unprovided line", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Citations[0].Line = 1 }},
		{name: "missing evidence", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Citations = nil }},
		{name: "unsolicited grade", change: func(_ *tutor.Reply, request *tutor.Request) { request.Action = "lesson" }},
		{name: "missing question to assess", change: func(_ *tutor.Reply, request *tutor.Request) { request.Question = "" }},
		{name: "missing feedback", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Feedback = "" }},
		{name: "unknown grade", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Assessment = "mastered" }},
		{name: "empty lesson", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Explanation = " " }},
		{name: "quiz without question", change: func(reply *tutor.Reply, request *tutor.Request) {
			request.Action = "quiz"
			reply.Assessment = "none"
			reply.Question = ""
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			reply := valid
			reply.Citations = append([]repository.Location(nil), valid.Citations...)
			current := request
			if test.change != nil {
				test.change(&reply, &current)
			}
			data, _ := json.Marshal(reply)
			_, err := tutor.DecodeReply(string(data), prepared, current)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v, got error %v", test.valid, err)
			}
		})
	}
	data, _ := json.Marshal(valid)
	for _, suffix := range []string{" {}", " garbage"} {
		if _, err := tutor.DecodeReply(string(data)+suffix, prepared, request); err == nil {
			t.Fatal("trailing content was accepted")
		}
	}
}

func TestContextIsBoundedAndRelevant(t *testing.T) {
	files := []repository.File{{Path: "go.mod", Content: "module course\n"}, {Path: "README.md", Content: "# Course\n"}, {Path: "target/target.go", Content: "package target\nfunc Explain() {}\n"}}
	for index := 0; index < 40; index++ {
		files = append(files, repository.File{Path: strings.Repeat("a", index+1) + ".go", Content: "package course\n" + strings.Repeat("var value = 1\n", 1000)})
	}
	snapshot := repository.Analyze(files)
	prepared := tutor.Prepare(snapshot, tutor.Request{Action: "question", Topic: "overview", Message: "Explain target/target.go."})
	bytes := 0
	found := false
	for _, excerpt := range prepared.Excerpts {
		bytes += len(excerpt.Text)
		if excerpt.Path == "target/target.go" {
			found = true
		}
		if excerpt.StartLine < 1 || excerpt.EndLine < excerpt.StartLine {
			t.Fatalf("invalid source window: %+v", excerpt)
		}
	}
	if bytes > 48000 || !found {
		t.Fatalf("source budget or relevance violated: %d bytes, target present=%v", bytes, found)
	}
}

func TestProgressIsEvidenceBasedAndImmutable(t *testing.T) {
	previous := map[string]tutor.Progress{"overview": {Lessons: 1}}
	read := tutor.Record(previous, "course/value", "lesson", tutor.Reply{Assessment: "none"})
	if read["course/value"] != (tutor.Progress{Lessons: 1}) || !reflect.DeepEqual(previous, map[string]tutor.Progress{"overview": {Lessons: 1}}) {
		t.Fatalf("reading counted as mastery or changed prior progress: %v %v", read, previous)
	}
	graded := tutor.Record(read, "course/value", "answer", tutor.Reply{Assessment: "partial"})
	if graded["course/value"] != (tutor.Progress{Lessons: 1, Partial: 1}) || read["course/value"].Partial != 0 {
		t.Fatal("grading did not preserve immutable progress")
	}
}

func TestFollowUpRetainsEvidenceAndFindsNamedSymbols(t *testing.T) {
	files := []repository.File{
		{Path: "go.mod", Content: "module course\n"},
		{Path: "README.md", Content: "# Course\n"},
		{Path: "deep/behavior.go", Content: "package deep\nfunc TraceRename() {}\n"},
	}
	for index := 0; index < 40; index++ {
		files = append(files, repository.File{Path: strings.Repeat("a", index+1) + ".go", Content: "package course\n" + strings.Repeat("var value = 1\n", 1000)})
	}
	snapshot := repository.Analyze(files)
	for _, request := range []tutor.Request{
		{Action: "question", Topic: "overview", Message: "How does TraceRename work?"},
		{Action: "answer", Topic: "overview", Message: "Why does it do that?", Previous: &tutor.Reply{Citations: []repository.Location{{Path: "deep/behavior.go", Line: 2}}}},
	} {
		prepared := tutor.Prepare(snapshot, request)
		found := false
		for _, excerpt := range prepared.Excerpts {
			if excerpt.Path == "deep/behavior.go" {
				found = true
			}
		}
		if !found {
			t.Fatalf("follow-up lost its evidence: %+v", request)
		}
	}
}
