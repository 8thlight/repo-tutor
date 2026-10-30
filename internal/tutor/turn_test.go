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
	request := tutor.Request{Topic: tutor.Overview, Action: "answer", Question: "Where does it start?"}
	valid := tutor.Reply{Explanation: "Execution begins in main.go:3.", Question: "What is delegated?", Assessment: "correct", Feedback: "You found the entry point.", Citations: []repository.Location{{Path: "main.go", Line: 3}}}
	for _, test := range []struct {
		name   string
		change func(*tutor.Reply, *tutor.Request)
		valid  bool
	}{
		{name: "supported answer", valid: true},
		{name: "invented file", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Citations[0].Path = "imaginary.go" }},
		{name: "unprovided line", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Citations[0].Line = 1 }},
		{name: "marked claim", valid: true, change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Explanation = "Execution begins in main.go [1]." }},
		{name: "marker with no citation", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Explanation = "Execution begins in main.go [2]." }},
		{name: "feedback marker with no citation", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Feedback = "You found it [3]." }},
		{name: "marker zero", change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Explanation = "Execution begins [0]." }},
		{name: "array index in code is not a marker", valid: true, change: func(reply *tutor.Reply, _ *tutor.Request) { reply.Explanation = "Execution begins in `books[7]` [1]." }},
		{name: "ten citations", change: func(reply *tutor.Reply, _ *tutor.Request) {
			reply.Citations = nil
			for range 10 {
				reply.Citations = append(reply.Citations, repository.Location{Path: "main.go", Line: 3})
			}
		}},
		{name: "nine citations", valid: true, change: func(reply *tutor.Reply, _ *tutor.Request) {
			reply.Citations = nil
			for range 9 {
				reply.Citations = append(reply.Citations, repository.Location{Path: "main.go", Line: 3})
			}
		}},
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
	prepared := tutor.Prepare(snapshot, tutor.Request{Action: "question", Topic: tutor.Overview, Message: "Explain target/target.go."})
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

func TestQuizContextRetainsTeachingSeparateFromLastFeedback(t *testing.T) {
	snapshot := repository.Analyze([]repository.File{
		{Path: "go.mod", Content: "module learning\n"},
		{Path: "core/value.go", Content: "package core\nfunc Calculate() int { return 42 }\n"},
		{Path: "cli/main.go", Content: "package cli\nfunc Run() {}\n"},
	})
	lesson := tutor.Reply{Explanation: "Core owns calculations; CLI handles commands.", Assessment: "none", Citations: []repository.Location{{Path: "core/value.go", Line: 2}}}
	feedback := tutor.Reply{Explanation: "That is correct.", Assessment: "correct", Feedback: "You identified core."}
	prepared := tutor.Prepare(snapshot, tutor.Request{Topic: tutor.Overview, Action: "quiz", Lesson: &lesson, Previous: &feedback})
	_, payload, found := strings.Cut(prepared.Prompt, "Turn and source excerpts:\n")
	if !found {
		t.Fatal("missing structured tutor context")
	}
	var context struct {
		Lesson   tutor.Reply `json:"last_lesson"`
		Previous tutor.Reply `json:"previous_reply"`
	}
	if err := json.Unmarshal([]byte(payload), &context); err != nil {
		t.Fatal(err)
	}
	if context.Lesson.Explanation != lesson.Explanation || context.Previous.Feedback != feedback.Feedback {
		t.Fatal("a subsequent quiz lost the taught material behind the latest feedback")
	}
	found = false
	for _, excerpt := range prepared.Excerpts {
		if excerpt.Path == "core/value.go" && excerpt.StartLine <= 2 && excerpt.EndLine >= 2 {
			found = true
		}
	}
	if !found {
		t.Fatal("quiz context omitted the code cited by its lesson")
	}
}

func TestProgressIsEvidenceBasedAndImmutable(t *testing.T) {
	previous := map[string]tutor.Progress{tutor.Overview: {Lessons: 1}}
	read := tutor.Record(previous, "course/value", "lesson", tutor.Reply{Assessment: "none"})
	if read["course/value"] != (tutor.Progress{Lessons: 1}) || !reflect.DeepEqual(previous, map[string]tutor.Progress{tutor.Overview: {Lessons: 1}}) {
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
		{Action: "question", Topic: tutor.Overview, Message: "How does TraceRename work?"},
		{Action: "answer", Topic: tutor.Overview, Message: "Why does it do that?", Previous: &tutor.Reply{Citations: []repository.Location{{Path: "deep/behavior.go", Line: 2}}}},
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

func hotspotSnapshot() repository.Snapshot {
	var files []repository.File
	var hotspots []repository.Hotspot
	for index := range 7 {
		path := "core/busy" + strings.Repeat("x", index) + ".go"
		files = append(files, repository.File{Path: path, Content: "package core\n" + strings.Repeat("var value = 1\n", 80)})
		hotspots = append(hotspots, repository.Hotspot{Path: path, Line: 50, Language: "go", Commits: 9 - index, Branches: 12, Lines: 81, Score: 100 - index})
	}
	snapshot := repository.Analyze(append(files, repository.File{Path: "go.mod", Content: "module course\n"}))
	snapshot.Hotspots = hotspots
	return snapshot
}

func TestTheCourseEndsWithWhereChangeHappensOnlyWhenThereAreHotspots(t *testing.T) {
	course := tutor.Course(hotspotSnapshot())
	last := course[len(course)-1]
	if last.ID != tutor.HotspotsTopic || last.Title != "Where change happens" {
		t.Fatalf("the hotspot topic should come last: %+v", course)
	}
	plain := hotspotSnapshot()
	plain.Hotspots = nil
	for _, topic := range tutor.Course(plain) {
		if topic.ID == tutor.HotspotsTopic {
			t.Fatal("a repository with no hotspots should not offer the topic")
		}
	}
}

func TestTheHotspotsTopicQuotesTheTopFivePlacesFirst(t *testing.T) {
	snapshot := hotspotSnapshot()
	prepared := tutor.Prepare(snapshot, tutor.Request{Action: "lesson", Topic: tutor.HotspotsTopic})
	if len(prepared.Excerpts) < 5 {
		t.Fatalf("expected at least five excerpts, got %d", len(prepared.Excerpts))
	}
	for index := range 5 {
		excerpt := prepared.Excerpts[index]
		if excerpt.Path != snapshot.Hotspots[index].Path || excerpt.StartLine != 38 {
			t.Fatalf("excerpt %d should open at the line before hotspot %d: %+v", index, index, excerpt)
		}
	}
	if !strings.Contains(prepared.Prompt, "For topic hotspots") {
		t.Fatal("the prompt should explain how to teach the hotspots topic")
	}
	other := tutor.Prepare(snapshot, tutor.Request{Action: "lesson", Topic: tutor.Overview})
	if len(other.Excerpts) > 0 && other.Excerpts[0].Path == snapshot.Hotspots[0].Path && other.Excerpts[0].StartLine == 38 {
		t.Fatal("other topics should not lead with the hotspots")
	}
}

func TestALessonNeverCarriesAQuestion(t *testing.T) {
	prepared := tutor.Context{Excerpts: []tutor.Excerpt{{Path: "main.go", StartLine: 2, EndLine: 5}}}
	request := tutor.Request{Topic: tutor.Overview, Action: "lesson"}
	for name, question := range map[string]string{"without a question": "", "with a question": "What is delegated?"} {
		reply := tutor.Reply{Explanation: "Execution begins in main.go:3.", Question: question, Assessment: "none", Citations: []repository.Location{{Path: "main.go", Line: 3}}}
		data, _ := json.Marshal(reply)
		decoded, err := tutor.DecodeReply(string(data), prepared, request)
		if err != nil || decoded.Question != "" {
			t.Fatalf("%s: a lesson should be accepted with an empty question, got %q and %v", name, decoded.Question, err)
		}
	}
}

func TestThePromptKeepsQuestionsForQuizzesAndProbes(t *testing.T) {
	prompt := tutor.Prepare(hotspotSnapshot(), tutor.Request{Action: "lesson", Topic: tutor.Overview}).Prompt
	for _, expected := range []string{"only for\nactions quiz and probe", "Do not ask the learner a question"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("the prompt should say %q", expected)
		}
	}
	if strings.Contains(prompt, "end with one comprehension question") {
		t.Fatal("lessons no longer end with a question")
	}
}

func TestThePromptAsksForNumberedMarkersAndNotPathReferences(t *testing.T) {
	prompt := tutor.Prepare(hotspotSnapshot(), tutor.Request{Action: "lesson", Topic: tutor.Overview}).Prompt
	for _, expected := range []string{"mark each factual code claim with [n]", "at most nine", "Do not write path:line\nin the text"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("the prompt should say %q", expected)
		}
	}
	if strings.Contains(prompt, "use path:line references") {
		t.Fatal("the prompt should no longer ask for path:line references in the text")
	}
}
