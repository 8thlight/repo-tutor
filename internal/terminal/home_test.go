package terminal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func newLearnerFixture() interfaceModel {
	model := interfaceFixture()
	model.application.state.Turns, model.application.state.Progress, model.application.state.Question = nil, map[string]tutor.Progress{}, ""
	model.history = -1
	model.page, model.pageTitle = "home", "Welcome"
	model.rebuildContent()
	return model
}

func TestHomeRecommendsLearnFirstAndStartsTheOverviewInTheBackground(t *testing.T) {
	model := newLearnerFixture()
	content := ansi.Strip(model.View().Content)
	for _, expected := range []string{"Recommended: You have not started.", "› 1  Learn", "Browse", "Home"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("home is missing %q:\n%s", expected, content)
		}
	}
	var places []string
	for _, entry := range model.homeEntries() {
		places = append(places, entry.place)
	}
	if strings.Join(places, ",") != "learn,browse" {
		t.Fatalf("a new learner should see only places with something to show: %v", places)
	}
	started, command := model.Update(beginFirstLesson{})
	model = started.(interfaceModel)
	defer model.cancel()
	if command == nil || !model.busy || model.page != "home" || !strings.Contains(ansi.Strip(model.View().Content), "Preparing the overview in the background") {
		t.Fatal("the overview should generate while Home stays on screen")
	}
}

func TestHomeKeysOpenTheRecommendationOrAnotherPlace(t *testing.T) {
	model := newLearnerFixture()
	model, _ = press(model, 'j', 0)
	if model.homeSelected != 1 {
		t.Fatal("down should move the selection")
	}
	browsed, _ := press(model, '2', 0)
	if browsed.page != "topics" {
		t.Fatalf("number keys should open that entry, got %s", browsed.page)
	}
	model.homeSelected = 0
	begun, command := model.openSelectedHomeEntry()
	if command == nil || !begun.(interfaceModel).busy || begun.(interfaceModel).page != "lesson" {
		t.Fatal("Enter on Learn before any lesson exists should request the overview")
	}
	begun.(interfaceModel).cancel()
}

func TestHomeShowsStatusAndQuizOnceThereIsProgress(t *testing.T) {
	model := interfaceFixture()
	model.application.state.Progress = map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, "core": {Lessons: 1, Correct: 1}}
	model.application.state.Quiz = &tutor.Quiz{Topic: "core", Goal: 5, Questions: []tutor.QuizQuestion{{Text: "One?", Assessment: "correct"}}}
	model.page = "home"
	model.rebuildContent()
	entries := model.homeEntries()
	if entries[0].place != tutor.PlaceQuiz || !strings.Contains(entries[0].reason, "1 of 5") {
		t.Fatalf("an unfinished quiz should be recommended first: %+v", entries[0])
	}
	content := ansi.Strip(model.View().Content)
	if !strings.Contains(content, "quiz 1 of 5") || strings.Contains(content, "teaches you this repository") || !strings.Contains(content, "Progress") {
		t.Fatalf("a returning learner should see status, not the introduction:\n%s", content)
	}
	model.application.state.Review = true
	if first := model.homeEntries()[0]; first.place != tutor.PlaceContinue {
		t.Fatalf("a pending review should be recommended before the quiz: %+v", first)
	}
}

func TestResumedSessionOpensOnLearnNotHome(t *testing.T) {
	if model := interfaceFixture(); model.page != "lesson" {
		t.Fatalf("a session with turns should resume on Learn, got %s", model.page)
	}
	if model := newInterface(interfaceFixture().ctx, application{state: session.State{}}, nil, true); model.page != "home" {
		t.Fatalf("a session with no turns should open on Home, got %s", model.page)
	}
}
