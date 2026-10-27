package terminal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func openFileFixture(t *testing.T) interfaceModel {
	t.Helper()
	model := readingFixture()
	model.width, model.height = 120, 36
	model.resize()
	opened, _ := model.openFile(repository.Location{Path: "core/value.go", Line: 1})
	return opened.(interfaceModel)
}

func TestTOnAnOpenFileStartsALessonOnThatFile(t *testing.T) {
	model := openFileFixture(t)
	started, _ := press(model, 't', 0)
	defer started.cancel()
	if !started.busy || started.intent.action != "lesson" || started.intent.topic != tutor.FileTopic("core/value.go") || started.intent.message != firstLessonMessage {
		t.Fatalf("T should start the file's first lesson: %+v", started.intent)
	}
	content := ansi.Strip(started.View().Content)
	if !strings.Contains(content, "Preparing lesson · File · core/value.go") || !strings.Contains(content, "File · core/value.go") {
		t.Fatalf("the title and sidebar should name the file:\n%s", content)
	}
}

func TestTAgainResumesTheSavedFileLessonWithoutARequest(t *testing.T) {
	model := openFileFixture(t)
	topic := tutor.FileTopic("core/value.go")
	reply := tutor.Reply{Explanation: "value.go declares one type.", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: topic, Action: "lesson", Reply: reply})
	model.application.state.Progress = map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, topic: {Lessons: 1}}
	model.application.state.Topic = tutor.Overview
	again, _ := press(model, 't', 0)
	if again.busy || again.application.state.Topic != topic || again.page != "lesson" {
		t.Fatalf("a studied file resumes with no request: busy=%v topic=%s page=%s", again.busy, again.application.state.Topic, again.page)
	}
	if again.application.state.Turns[len(again.application.state.Turns)-1].Action != "lesson" {
		t.Fatal("the lesson should be the current step")
	}
}

func TestTOutsideTheFileViewIsJustTyping(t *testing.T) {
	model := readingFixture()
	typed := typeText(model, "t")
	if typed.busy || typed.application.state.Topic != model.application.state.Topic || typed.input.Value() != "t" {
		t.Fatal("T only teaches a file when a source file is open")
	}
}

func TestTheFileViewNamesTheTeachKey(t *testing.T) {
	if notice := openFileFixture(t).notice; !strings.Contains(notice, "T teaches this file") {
		t.Fatalf("the notice should name the key: %q", notice)
	}
	if !strings.Contains(interfaceHelp, "Teach the open source file") {
		t.Fatal("the Help page should list the key")
	}
}

func TestAfterAFileLessonNextIsTheFirstUnreadCourseTopic(t *testing.T) {
	model := readingFixture()
	file := tutor.FileTopic("core/value.go")
	model.application.state.Topic = file
	model.application.state.Progress = map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, file: {Lessons: 1}}
	title, found := model.nextTopicLabel()
	if !found || title == "" {
		t.Fatal("from a file lesson the course should still offer a next topic")
	}
	first := ""
	for _, topic := range model.application.course {
		if model.application.state.Progress[topic.ID].Lessons == 0 {
			first = tutor.Plain(topic.Title)
			break
		}
	}
	if title != first {
		t.Fatalf("next should be the first unread topic %q, got %q", first, title)
	}
}

func TestAFileTopicIsNotPartOfTheCourseFigures(t *testing.T) {
	model := readingFixture()
	file := tutor.FileTopic("core/value.go")
	before := tutor.ScoreCourse(model.application.course, model.application.state.Progress)
	progress := map[string]tutor.Progress{file: {Lessons: 3, QuizCredit: 10, QuizGoal: 5}}
	for key, value := range model.application.state.Progress {
		progress[key] = value
	}
	after := tutor.ScoreCourse(model.application.course, progress)
	if before != after {
		t.Fatalf("file lessons must not change coverage or understanding: %+v vs %+v", before, after)
	}
}
