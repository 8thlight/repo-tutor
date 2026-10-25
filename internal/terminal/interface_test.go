package terminal

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func interfaceFixture() interfaceModel {
	snapshot := repository.Analyze([]repository.File{
		{Path: "go.mod", Content: "module learning\n"},
		{Path: "README.md", Content: "# Learning\nA repository for developers.\n"},
		{Path: "main.go", Content: "package main\nfunc main() {}\n"},
		{Path: "core/value.go", Content: "package core\ntype Value struct{}\n"},
	})
	snapshot.Root = "/work/repowise"
	reply := tutor.Reply{Explanation: "The **CLI** is the front door to the repository. It hands work to the `core` package.\n\n- Start with the entry point.\n- Follow one operation into the core.", Question: "Which package would you inspect to change a calculation?", Assessment: "none", Citations: []repository.Location{{Path: "main.go", Line: 2}}}
	state := session.State{Version: 1, Repository: snapshot.Root, Agent: "codex", Goal: "Understand the architecture and make a small change.", Topic: tutor.Overview, Question: reply.Question, Progress: map[string]tutor.Progress{tutor.Overview: {Lessons: 1}}, Turns: []session.Turn{{Topic: tutor.Overview, Action: "lesson", Reply: reply}}}
	application := application{snapshot: snapshot, course: tutor.Course(snapshot), state: state, timeout: 3 * time.Minute}
	return newInterface(context.Background(), application, func(context.Context, session.State, string, string, string) (session.State, error) { return state, nil }, true)
}

func press(model interfaceModel, code rune, modifier tea.KeyMod) (interfaceModel, tea.Cmd) {
	next, command := model.Update(tea.KeyPressMsg{Code: code, Mod: modifier})
	return next.(interfaceModel), command
}

func TestInterfaceFitsTerminalAndKeepsExitDiscoverable(t *testing.T) {
	for _, size := range []struct{ width, height int }{{44, 18}, {60, 24}, {80, 24}, {100, 32}, {120, 36}, {160, 48}} {
		model := interfaceFixture()
		resized, _ := model.Update(tea.WindowSizeMsg{Width: size.width, Height: size.height})
		model = resized.(interfaceModel)
		pages := []string{"lesson", "topics", "quiz", "sources", "hotspots", "progress", "help"}
		for _, page := range pages {
			if page == "topics" || page == "sources" || page == "hotspots" {
				opened, _ := model.openPicker(page, "")
				model = opened.(interfaceModel)
			} else {
				model.page = page
				if page == "help" {
					model.pageContent = interfaceHelp
				}
				model.rebuildContent()
			}
			content := ansi.Strip(model.View().Content)
			lines := strings.Split(content, "\n")
			if len(lines) > size.height {
				t.Fatalf("%s overflows %dx%d vertically: %d", page, size.width, size.height, len(lines))
			}
			for _, line := range lines {
				if ansi.StringWidth(line) > size.width {
					t.Fatalf("%s overflows %dx%d: %q", page, size.width, size.height, line)
				}
			}
			if !strings.Contains(content, "Ctrl+Q") {
				t.Fatalf("exit shortcut disappeared on %s at %dx%d:\n%s", page, size.width, size.height, content)
			}
			if strings.Contains(model.View().Content, "\x1b[") {
				t.Fatal("NO_COLOR output retained style sequences")
			}
		}
	}
}

func TestNavigationRetainsDraftAndDoesNotCallAgent(t *testing.T) {
	model := interfaceFixture()
	calls := 0
	model.run = func(context.Context, session.State, string, string, string) (session.State, error) {
		calls++
		return session.State{}, nil
	}
	model.input.SetValue("My unfinished answer")
	model, _ = press(model, 's', tea.ModCtrl)
	if model.page != "sources" {
		t.Fatal("source shortcut did not open the source browser")
	}
	model, _ = press(model, tea.KeyEnter, 0)
	if model.page != "source" || !strings.Contains(ansi.Strip(model.pageContent), "func main()") {
		t.Fatalf("citation did not open actual source: %s", model.pageContent)
	}
	model, _ = press(model, tea.KeyEscape, 0)
	if model.page != "lesson" || model.input.Value() != "My unfinished answer" || calls != 0 {
		t.Fatal("browsing lost the draft or called an agent")
	}
	opened, _ := model.openPicker("topics", "core")
	model = opened.(interfaceModel)
	selected := model.picker.SelectedItem().(browserItem)
	resized, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 38})
	model = resized.(interfaceModel)
	if model.picker.FilterValue() != "core" || model.picker.SelectedItem().(browserItem).value != selected.value {
		t.Fatal("resize lost the topic filter or selection")
	}
}

func TestCurrentQuestionStaysVisibleWhileLessonScrolls(t *testing.T) {
	model := interfaceFixture()
	model.application.state.Turns[0].Reply.Explanation = strings.Repeat("A long lesson with enough detail to require scrolling.\n\n", 60)
	model.rebuildContent()
	for _, offset := range []int{0, 40, 200} {
		model.viewport.SetYOffset(offset)
		content := ansi.Strip(model.View().Content)
		if !strings.Contains(content, model.application.state.Question) || !strings.Contains(content, "NEXT QUESTION") {
			t.Fatalf("the current question disappeared at scroll offset %d", offset)
		}
	}
}

func TestGradedAnswerStaysWithItsQuestionAndFeedback(t *testing.T) {
	model := interfaceFixture()
	previousQuestion := model.application.state.Question
	reply := tutor.Reply{Explanation: "Next idea: how a CLI command is wired up.", Assessment: "correct", Feedback: "Core is the shared engine.", Question: "Where are commands registered?"}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: model.application.state.Topic, Action: "answer", Message: "core", Reply: reply})
	model.application.state.Question = reply.Question
	model.history = 1
	content := ansi.Strip(model.lessonContent())
	position := -1
	for _, text := range []string{"QUESTION YOU ANSWERED", previousQuestion, "YOUR ANSWER", "core", "Correct", reply.Feedback, "── END OF QUESTION"} {
		next := strings.Index(content[position+1:], text)
		if next < 0 {
			t.Fatalf("missing or misplaced %q after position %d:\n%s", text, position, content)
		}
		position += next + 1
	}
	composer := ansi.Strip(model.composer())
	if !strings.HasPrefix(composer, "────") || !strings.Contains(composer, "Continue") || strings.Contains(composer, reply.Question) || strings.Contains(content, reply.Explanation) {
		t.Fatalf("review must stop before the next idea or question:\n%s\n%s", content, composer)
	}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: "another", Action: "lesson", Reply: tutor.Reply{Assessment: "none"}})
	content = ansi.Strip(model.lessonContent())
	if strings.Contains(content, reply.Question) || !strings.Contains(content, reply.Feedback) {
		t.Fatalf("historical review should keep feedback without mixing in the next question:\n%s", content)
	}
	model.application.state.Turns[0].Topic = "another"
	if strings.Contains(ansi.Strip(model.lessonContent()), previousQuestion) {
		t.Fatal("answer was paired with a question from another topic")
	}
	model.application.state.Turns = model.application.state.Turns[1:]
	model.history = 0
	content = ansi.Strip(model.lessonContent())
	if strings.Contains(content, "QUESTION YOU ANSWERED") || !strings.Contains(content, reply.Feedback) {
		t.Fatalf("trimmed history must keep feedback without inventing a question:\n%s", content)
	}
}

func TestPlainFeedbackStopsBeforeNextIdea(t *testing.T) {
	var output bytes.Buffer
	application := application{output: &output, state: session.State{Review: true}}
	application.show(tutor.Reply{Explanation: "Next lesson content.", Assessment: "partial", Feedback: "Review the shared engine.", Question: "What calls it?"})
	content := output.String()
	position := -1
	for _, text := range []string{"Assessment: partial", "Review the shared engine.", "--- End of question ---", "/continue"} {
		next := strings.Index(content[position+1:], text)
		if next < 0 {
			t.Fatalf("missing or misplaced %q:\n%s", text, content)
		}
		position += next + 1
	}
	if strings.Contains(content, "Next lesson content.") || strings.Contains(content, "What calls it?") {
		t.Fatal("plain output advanced before an explicit continue")
	}
}

func TestSuccessfulTurnAcceptsSavedStateAndClearsDraft(t *testing.T) {
	model := interfaceFixture()
	previous := model.application.state
	updated := previous
	reply := tutor.Reply{Explanation: "The calculation lives in core/value.go:2.", Assessment: "correct", Feedback: "You found the shared value type.", Question: "Who uses it?", Citations: []repository.Location{{Path: "core/value.go", Line: 2}}}
	updated.Turns = append(append([]session.Turn(nil), previous.Turns...), session.Turn{Topic: previous.Topic, Action: "answer", Message: "The core package", Reply: reply})
	updated.Question = reply.Question
	updated.Progress = tutor.Record(previous.Progress, previous.Topic, "answer", reply)
	model.run = func(context.Context, session.State, string, string, string) (session.State, error) {
		return updated, nil
	}
	model.input.SetValue("The core package")
	started, command := model.submit()
	next, _ := started.(interfaceModel).Update(finishedCommand(t, command))
	model = next.(interfaceModel)
	if model.input.Value() != "" || model.busy || model.application.state.Question != reply.Question || model.application.state.Progress[previous.Topic].Correct != 1 {
		t.Fatal("successful completion did not adopt saved progress and reset the editor")
	}
	if len(previous.Turns) != 1 || previous.Progress[previous.Topic].Correct != 0 {
		t.Fatal("the interface changed its previous state while handling a turn")
	}
}

func finishedCommand(t *testing.T, command tea.Cmd) turnFinished {
	t.Helper()
	message := command()
	if result, ok := message.(turnFinished); ok {
		return result
	}
	if batch, ok := message.(tea.BatchMsg); ok {
		for _, child := range batch {
			if result, ok := child().(turnFinished); ok {
				return result
			}
		}
	}
	t.Fatal("turn command returned no result")
	return turnFinished{}
}

func TestFailedTurnKeepsDraftAndRetryUsesSameIntent(t *testing.T) {
	model := interfaceFixture()
	model.input.SetValue("The core package")
	calls := 0
	model.run = func(_ context.Context, state session.State, action, message, topic string) (session.State, error) {
		calls++
		if action != "answer" || message != "The core package" || topic != tutor.Overview {
			t.Fatal("retry changed the pending answer")
		}
		return state, errors.New("test account needs login")
	}
	started, command := model.submit()
	model = started.(interfaceModel)
	if !model.busy || model.input.Value() != "The core package" {
		t.Fatal("answer was cleared before a successful reply")
	}
	duplicate, duplicateCommand := model.begin(turnIntent{"quiz", "another question", tutor.Overview})
	if !duplicate.(interfaceModel).busy || duplicateCommand != nil {
		t.Fatal("a second turn was scheduled while one was active")
	}
	next, _ := model.Update(finishedCommand(t, command))
	model = next.(interfaceModel)
	if model.busy || model.input.Value() != "The core package" || len(model.application.state.Turns) != 1 || !strings.Contains(ansi.Strip(model.View().Content), "Turn not recorded") {
		t.Fatal("failed turn lost progress, draft, or recovery guidance")
	}
	retried, retry := model.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	finishedCommand(t, retry)
	retried.(interfaceModel).cancel()
	if calls != 2 {
		t.Fatal("retry did not invoke the same turn")
	}
}

func TestCancellationAndFirstLessonFailureAreRecoverable(t *testing.T) {
	model := interfaceFixture()
	model.application.state.Turns = nil
	model.history = -1
	model.run = func(ctx context.Context, _ session.State, _, _, _ string) (session.State, error) {
		return session.State{}, ctx.Err()
	}
	started, command := model.begin(turnIntent{"lesson", "start", tutor.Overview})
	model = started.(interfaceModel)
	model, _ = press(model, tea.KeyEscape, 0)
	next, _ := model.Update(finishedCommand(t, command))
	model = next.(interfaceModel)
	content := ansi.Strip(model.View().Content)
	if model.busy || !strings.Contains(content, "context canceled") || !strings.Contains(content, "Ctrl+R") {
		t.Fatalf("first-lesson cancellation offered no recovery:\n%s", content)
	}
}
