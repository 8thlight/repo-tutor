package terminal

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
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

func homeNumber(model interfaceModel, place string) rune {
	for index, entry := range model.homeEntries() {
		if entry.place == place {
			return rune('1' + index)
		}
	}
	return '0'
}

func TestHomeRecommendsLearnFirstAndStartsTheOverviewInTheBackground(t *testing.T) {
	model := newLearnerFixture()
	content := ansi.Strip(model.View().Content)
	for _, expected := range []string{"Recommended: You have not started.", "› 1  Course", "Topics", "Home"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("home is missing %q:\n%s", expected, content)
		}
	}
	var places []string
	for _, entry := range model.homeEntries() {
		places = append(places, entry.place)
	}
	if strings.Join(places, ",") != "learn,topics,about" {
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
	browsed, _ := press(model, homeNumber(model, placeTopics), 0)
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

func TestAboutPageShowsBuildCopyrightAndSessionAgent(t *testing.T) {
	model := interfaceFixture()
	opened, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyF2})
	model = opened.(interfaceModel)
	content := ansi.Strip(model.View().Content)
	for _, expected := range []string{"About repo-tutor", "Copyright © 2026 Eighth Light LLC.", "https://8thlight.com", "This session uses codex (default model)."} {
		if !strings.Contains(content, expected) {
			t.Fatalf("about page is missing %q:\n%s", expected, content)
		}
	}
}

func TestAboutShowsThePaintingOnlyOnWideTerminalsWithRichColor(t *testing.T) {
	for name, test := range map[string]struct {
		width   int
		profile colorprofile.Profile
		noColor bool
		shown   bool
	}{
		"wide truecolor": {113, colorprofile.TrueColor, false, true},
		"wide 256 color": {113, colorprofile.ANSI256, false, true},
		"too narrow":     {112, colorprofile.TrueColor, false, false},
		"16 colors":      {140, colorprofile.ANSI, false, false},
		"NO_COLOR":       {140, colorprofile.TrueColor, true, false},
	} {
		model := interfaceFixture()
		model.width, model.height = test.width, 60
		colored, _ := model.Update(tea.ColorProfileMsg{Profile: test.profile})
		model = colored.(interfaceModel)
		model.noColor = test.noColor
		next, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyF2})
		model = next.(interfaceModel)
		if shown := strings.Contains(model.pageContent, "Jan Lievens, 1631"); shown != test.shown {
			t.Fatalf("%s: painting shown=%v, want %v", name, shown, test.shown)
		}
		if test.shown {
			for _, line := range strings.Split(ansi.Strip(model.View().Content), "\n") {
				if ansi.StringWidth(line) > test.width {
					t.Fatalf("%s: line overflows: %q", name, line)
				}
			}
		}
	}
}

func TestAboutOpensFromHomeAndFromAnyPageWithF2(t *testing.T) {
	model := newLearnerFixture()
	entries := model.homeEntries()
	if last := entries[len(entries)-1]; last.place != placeAbout {
		t.Fatalf("Home should end with About, got %+v", last)
	}
	opened, _ := press(model, rune('0'+len(entries)), 0)
	if opened.page != "about" {
		t.Fatalf("Home's About entry should open the page, got %s", opened.page)
	}
	for _, page := range []string{"topics", "quiz", "progress", "home"} {
		model := interfaceFixture()
		switch page {
		case "topics", "quiz":
			next, _ := model.openPicker(page, "")
			model = next.(interfaceModel)
		default:
			model.page = page
		}
		pressed, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyF2})
		if pressed.(interfaceModel).page != "about" {
			t.Fatalf("F2 should open About from %s", page)
		}
	}
}

func TestPressingAPageKeyAgainReturnsToTheLesson(t *testing.T) {
	for name, key := range map[string]tea.KeyPressMsg{
		"topics":   {Code: 'k', Mod: tea.ModCtrl},
		"sources":  {Code: 's', Mod: tea.ModCtrl},
		"hotspots": {Code: 'o', Mod: tea.ModCtrl},
		"progress": {Code: 'p', Mod: tea.ModCtrl},
		"quiz":     {Code: 'b', Mod: tea.ModCtrl},
		"help":     {Code: tea.KeyF1},
		"about":    {Code: tea.KeyF2},
	} {
		model := interfaceFixture()
		model.history = 0
		opened, _ := model.Update(key)
		if opened.(interfaceModel).page != name {
			t.Fatalf("%s key should open %s, got %s", name, name, opened.(interfaceModel).page)
		}
		closed, _ := opened.(interfaceModel).Update(key)
		if closed.(interfaceModel).page != "lesson" {
			t.Fatalf("pressing the %s key again should return to the lesson, got %s", name, closed.(interfaceModel).page)
		}
	}
}

func TestPageKeyOnAnotherPageSwitchesInsteadOfClosing(t *testing.T) {
	model := interfaceFixture()
	opened, _ := model.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	switched, _ := opened.(interfaceModel).Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if switched.(interfaceModel).page != "sources" {
		t.Fatal("a different page key should open that page")
	}
}

func TestEscAndPageKeysReturnToWhereThePageWasOpened(t *testing.T) {
	home := newLearnerFixture()
	about := len(home.homeEntries())
	opened, _ := press(home, rune('0'+about), 0)
	if opened.page != "about" || opened.origin != "home" {
		t.Fatalf("About opened from Home should remember Home: %s %q", opened.page, opened.origin)
	}
	if closed, _ := press(opened, tea.KeyEscape, 0); closed.page != "home" {
		t.Fatalf("Esc should return to Home, got %s", closed.page)
	}
	if closed, _ := opened.Update(tea.KeyPressMsg{Code: tea.KeyF2}); closed.(interfaceModel).page != "home" {
		t.Fatal("F2 again should return to Home")
	}
	if lesson, _ := press(opened, 'l', tea.ModCtrl); lesson.page != "lesson" {
		t.Fatal("Ctrl+L should always go to the lesson")
	}

	browsed, _ := press(home, homeNumber(home, placeTopics), 0)
	switched, _ := press(browsed, 's', tea.ModCtrl)
	if switched.page != "sources" || switched.origin != "home" {
		t.Fatalf("switching pages directly should keep the original origin: %s %q", switched.page, switched.origin)
	}
	if closed, _ := press(switched, tea.KeyEscape, 0); closed.page != "home" {
		t.Fatalf("one Esc should return to where the learner started, got %s", closed.page)
	}

	lesson := interfaceFixture()
	fromLesson, _ := lesson.Update(tea.KeyPressMsg{Code: tea.KeyF2})
	if fromLesson.(interfaceModel).origin != "lesson" {
		t.Fatal("About opened from the lesson should remember the lesson")
	}
	if closed, _ := press(fromLesson.(interfaceModel), tea.KeyEscape, 0); closed.page != "lesson" {
		t.Fatalf("Esc should return to the lesson, got %s", closed.page)
	}
}

func TestFailedOverviewShowsOnHomeAndRetryStaysThere(t *testing.T) {
	model := newLearnerFixture()
	model.run = func(context.Context, session.State, string, string, string) (session.State, error) {
		return session.State{}, errors.New("connection refused")
	}
	started, command := model.Update(beginFirstLesson{})
	next, _ := started.(interfaceModel).Update(finishedCommand(t, command))
	model = next.(interfaceModel)
	content := ansi.Strip(model.View().Content)
	if model.page != "home" || model.busy || !strings.Contains(content, "The overview could not be prepared. Ctrl+R tries again.") || !strings.Contains(content, "connection refused") {
		t.Fatalf("a failed overview should be explained on Home:\n%s", content)
	}
	retried, command := press(model, 'r', tea.ModCtrl)
	if retried.page != "home" || !retried.busy || command == nil {
		t.Fatalf("Ctrl+R should retry without leaving Home: page=%s busy=%v", retried.page, retried.busy)
	}
	retried.cancel()
}

func TestHomeOffersProbeAndRefreshOnlyWhenTheyApply(t *testing.T) {
	model := newLearnerFixture()
	if strings.Contains(strings.Join(func() (names []string) {
		for _, entry := range model.homeEntries() {
			names = append(names, entry.place)
		}
		return
	}(), ","), placeProbe) {
		t.Fatal("the probe should wait until the overview has been opened")
	}
	model.application.state.Turns = interfaceFixture().application.state.Turns
	places := func() string {
		var names []string
		for _, entry := range model.homeEntries() {
			names = append(names, entry.place)
		}
		return strings.Join(names, ",")
	}
	if strings.Contains(places(), placeRefresh) {
		t.Fatal("refresh should wait for a stale overview")
	}
	model.overviewStale = true
	if !strings.Contains(places(), placeRefresh) {
		t.Fatal("a stale overview should offer a refresh")
	}
	begun, command := press(model, homeNumber(model, placeRefresh), 0)
	if command == nil || !begun.busy || begun.intent.message != openingMessage {
		t.Fatal("Refresh should request a new overview")
	}
	begun.cancel()
	probed, command := press(model, homeNumber(model, placeProbe), 0)
	if command == nil || !probed.busy || probed.intent.action != "probe" {
		t.Fatalf("Check what you know should start the probe: %+v", probed.intent)
	}
	probed.cancel()
	model.application.state.Probe = &tutor.Probe{Topics: []string{"a"}}
	if strings.Contains(places(), placeProbe) {
		t.Fatal("the probe entry should end once a probe has run")
	}
}

func TestTabsGroupPagesByActivity(t *testing.T) {
	model := interfaceFixture()
	for page, want := range map[string]string{"home": "home", "lesson": "course", "topics": "course", "quiz": "quiz", "sources": "course", "source": "course", "hotspots": "course", "progress": "progress", "help": "help", "about": "about"} {
		model.page = page
		if got := model.activeTab(); got != want {
			t.Fatalf("%s page should highlight %s, got %s", page, want, got)
		}
	}
	model.noColor = false
	opened, _ := model.openPicker("topics", "")
	content := ansi.Strip(opened.(interfaceModel).View().Content)
	if !strings.Contains(content, "Home   Course   Quiz   Progress") {
		t.Fatalf("the tab bar should show four tabs:\n%s", content)
	}
}

func TestTabKeyStepsThroughTheFourTabs(t *testing.T) {
	model := newLearnerFixture()
	var visited []string
	for index := 0; index < 4; index++ {
		model, _ = press(model, tea.KeyTab, 0)
		visited = append(visited, model.activeTab())
	}
	if strings.Join(visited, ",") != "course,quiz,progress,home" {
		t.Fatalf("Tab should step Home, Course, Quiz, Progress and wrap: %v", visited)
	}
	if model.page != "home" {
		t.Fatalf("wrapping should land on Home, got %s", model.page)
	}
	back, _ := press(model, tea.KeyTab, tea.ModShift)
	if back.activeTab() != "progress" {
		t.Fatalf("Shift+Tab should go back one tab, got %s", back.activeTab())
	}
	sources, _ := press(newLearnerFixture(), 's', tea.ModCtrl)
	next, _ := press(sources, tea.KeyTab, 0)
	if next.activeTab() != "quiz" || next.page != "quiz" {
		t.Fatalf("Tab from Sources should reach Quiz, got %s", next.page)
	}
}

func TestEscLeavesASourceFileForTheListItCameFrom(t *testing.T) {
	for _, list := range []string{"sources", "hotspots"} {
		opened, _ := newLearnerFixture().openPicker(list, "")
		model := opened.(interfaceModel)
		file, _ := model.openFile(repository.Location{Path: "main.go", Line: 1})
		inFile := file.(interfaceModel)
		if inFile.page != "source" {
			t.Fatalf("%s: Enter should open a file, got %s", list, inFile.page)
		}
		back, _ := press(inFile, tea.KeyEscape, 0)
		if back.page != list {
			t.Fatalf("Esc from a file opened from %s should return to it, got %s", list, back.page)
		}
	}
}

func TestEnterClosesASourceFileTheWayItOpenedIt(t *testing.T) {
	for _, list := range []string{"sources", "hotspots"} {
		opened, _ := newLearnerFixture().openPicker(list, "")
		file, _ := opened.(interfaceModel).openFile(repository.Location{Path: "main.go", Line: 1})
		back, _ := press(file.(interfaceModel), tea.KeyEnter, 0)
		if back.page != list {
			t.Fatalf("Enter on a file opened from %s should return to it, got %s", list, back.page)
		}
	}
	read, _ := newLearnerFixture().openFile(repository.Location{Path: "main.go", Line: 1})
	back, _ := press(read.(interfaceModel), tea.KeyEnter, 0)
	if back.page != "lesson" {
		t.Fatalf("Enter on a file opened with /read should return to the lesson, got %s", back.page)
	}
}
