package terminal

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func pendingGoalFixture(t *testing.T) interfaceModel {
	base := interfaceFixture()
	store, err := session.Open(t.TempDir(), "/work/repowise")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	application := base.application
	application.state.Turns, application.state.Question, application.state.GoalChosen = nil, "", false
	application.state.Goal = tutor.DefaultGoal
	application.state.Progress = map[string]tutor.Progress{}
	application.store = store
	return newInterface(context.Background(), application, base.run, true)
}

func typeText(model interfaceModel, text string) interfaceModel {
	for _, character := range text {
		next, _ := model.Update(tea.KeyPressMsg{Code: character, Text: string(character)})
		model = next.(interfaceModel)
	}
	return model
}

func startupMessages(model interfaceModel) []tea.Msg {
	messages := make(chan tea.Msg, 16)
	command := model.Init()
	var commands []tea.Cmd
	if batch, ok := command().(tea.BatchMsg); ok {
		commands = batch
	}
	for _, each := range commands {
		if each != nil {
			go func() { messages <- each() }()
		}
	}
	var collected []tea.Msg
	deadline := time.After(100 * time.Millisecond)
	for {
		select {
		case message := <-messages:
			collected = append(collected, message)
		case <-deadline:
			return collected
		}
	}
}

func hasFirstLesson(messages []tea.Msg) bool {
	for _, message := range messages {
		if _, found := message.(beginFirstLesson); found {
			return true
		}
	}
	return false
}

func TestAFirstRunAsksForTheGoalBeforeAnythingStarts(t *testing.T) {
	model := pendingGoalFixture(t)
	if !model.askingGoal() {
		t.Fatal("a session with no turns and no chosen goal should ask")
	}
	content := ansi.Strip(model.homeContent())
	for _, expected := range []string{goalQuestion, tutor.DefaultGoal + "  (default)"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("Home is missing %q:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "Topics") || strings.Contains(content, "Enter opens the selected item") {
		t.Fatalf("Home should list no entries while the goal is pending:\n%s", content)
	}
	if !strings.Contains(ansi.Strip(model.composer()), "Enter keeps the default") {
		t.Fatalf("the answer box should say Enter keeps the default:\n%s", ansi.Strip(model.composer()))
	}
	if hasFirstLesson(startupMessages(model)) {
		t.Fatal("nothing may be generated before the goal is chosen")
	}
	if !hasFirstLesson(startupMessages(newLearnerFixture())) {
		t.Fatal("a session whose goal is chosen should still start the overview at launch")
	}
}

func TestOnlyAFreshUnchosenSessionAsks(t *testing.T) {
	if interfaceFixture().goalPending {
		t.Fatal("a session with turns should not ask")
	}
	model := pendingGoalFixture(t)
	model.application.state.GoalChosen = true
	if newInterface(context.Background(), model.application, model.run, true).goalPending {
		t.Fatal("a chosen goal should not be asked again")
	}
}

func TestEnterOnAnEmptyBoxKeepsTheDefaultAndStartsTheOverview(t *testing.T) {
	model := pendingGoalFixture(t)
	chosen, command := model.submitGoal()
	next := chosen.(interfaceModel)
	if next.askingGoal() || next.application.state.Goal != tutor.DefaultGoal || !next.application.state.GoalChosen {
		t.Fatalf("Enter should choose the default: %+v", next.application.state)
	}
	if _, started := command().(beginFirstLesson); !started {
		t.Fatal("choosing the goal should start the overview")
	}
	saved, found, _ := model.application.store.Load()
	if !found || !saved.GoalChosen || saved.Goal != tutor.DefaultGoal {
		t.Fatalf("the choice should be saved: %+v", saved)
	}
	if !strings.Contains(ansi.Strip(next.homeContent()), "Topics") {
		t.Fatal("Home should show its entries once the goal is chosen")
	}
}

func TestATypedGoalIsSavedAndReplacesTheDefault(t *testing.T) {
	model := typeText(pendingGoalFixture(t), "Trace how a plan becomes a document")
	next, _ := press(model, tea.KeyEnter, 0)
	if next.application.state.Goal != "Trace how a plan becomes a document" {
		t.Fatalf("got %q", next.application.state.Goal)
	}
	saved, _, _ := model.application.store.Load()
	if saved.Goal != "Trace how a plan becomes a document" || !saved.GoalChosen {
		t.Fatalf("the typed goal should be saved: %+v", saved)
	}
	if next.input.Placeholder != lessonPlaceholder || next.input.Value() != "" {
		t.Fatal("the answer box should return to the lesson prompt")
	}
}

func TestOtherKeysAreHeldUntilTheGoalIsChosen(t *testing.T) {
	model := pendingGoalFixture(t)
	for name, move := range map[string]func(interfaceModel) interfaceModel{
		"Tab":    func(m interfaceModel) interfaceModel { next, _ := press(m, tea.KeyTab, 0); return next },
		"Esc":    func(m interfaceModel) interfaceModel { next, _ := press(m, tea.KeyEscape, 0); return next },
		"Ctrl+K": func(m interfaceModel) interfaceModel { next, _ := press(m, 'k', tea.ModCtrl); return next },
		"Ctrl+G": func(m interfaceModel) interfaceModel { next, _ := press(m, 'g', tea.ModCtrl); return next },
	} {
		next := move(model)
		if next.page != "home" || next.notice != goalPendingNotice || !next.goalPending {
			t.Fatalf("%s should be held with a notice: page %s notice %q", name, next.page, next.notice)
		}
	}
}

func TestHelpAndSettingsStillOpenAndReturnToTheQuestion(t *testing.T) {
	model := pendingGoalFixture(t)
	help, _ := press(model, tea.KeyF1, 0)
	if help.page != "help" {
		t.Fatalf("F1 should open help, got %s", help.page)
	}
	back, _ := press(help, tea.KeyEscape, 0)
	if back.page != "home" || !back.askingGoal() {
		t.Fatalf("Esc should return to the question, got %s", back.page)
	}
	_, quit := press(model, 'q', tea.ModCtrl)
	if quit == nil {
		t.Fatal("Ctrl+Q should still quit")
	}
}

func TestATooLongGoalKeepsTheBoxAndExplains(t *testing.T) {
	model := typeText(pendingGoalFixture(t), strings.Repeat("a", goalLimit+1))
	refused, command := model.submitGoal()
	next := refused.(interfaceModel)
	if !next.askingGoal() || command != nil || !strings.Contains(next.notice, "500") || next.input.Value() == "" {
		t.Fatalf("a long goal should be refused and kept: %q", next.notice)
	}
}
