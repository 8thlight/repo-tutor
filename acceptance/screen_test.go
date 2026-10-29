package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"
)

func launchScreen(t *testing.T) *tuitest.Client {
	t.Helper()
	return launchScreenWith(t, fixture(t), t.TempDir())
}

func launchScreenWith(t *testing.T, repository, stateDirectory string, extra ...string) *tuitest.Client {
	t.Helper()
	terminal := launchTerminal(t, 120, 36, nil, append([]string{repository, "--agent", "codex", "--state-dir", stateDirectory, "--goal", "learn it"}, extra...)...)
	waitForText(t, terminal, "Understanding")
	return terminal
}

func launchTerminal(t *testing.T, columns, rows uint16, environment map[string]string, arguments ...string) *tuitest.Client {
	t.Helper()
	variables := map[string]string{"PATH": agentDirectory + ":" + "/usr/bin:/bin", "HOME": t.TempDir(), "FAKE_AGENT_PID_FILE": filepath.Join(t.TempDir(), "agent.pid")}
	for name, value := range environment {
		variables[name] = value
	}
	return tuitesttest.New(t, tuitesttest.Options{
		Program: binary,
		Args:    arguments,
		Spawn: tuitest.SpawnOptions{
			Cols: tuitest.Ptr(columns), Rows: tuitest.Ptr(rows), WaitReady: tuitest.Ptr(false),
			Env: variables,
		},
	})
}

func waitForText(t *testing.T, terminal *tuitest.Client, text string) {
	t.Helper()
	if err := terminal.GetByText(text, tuitest.TextSelectorOptions{}).Wait(tuitest.LocatorWaitOptions{}); err != nil {
		screen, _ := terminal.Screenshot("", tuitest.ScreenshotOptions{})
		t.Fatalf("never saw %q: %v\n%s", text, err, screen)
	}
}

func TestSettingsOpensOnF2AndCtrlQExitsCleanly(t *testing.T) {
	terminal := launchScreen(t)
	if err := terminal.Press("F2"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "TAB STYLE")
	if err := terminal.Press("Escape"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Understanding")
	if running, err := terminal.State(); err != nil || !running.Modes["alternate_screen"] {
		t.Fatalf("the workspace should own the alternate screen: %v %v", err, running.Modes)
	}
	if err := terminal.Press("Ctrl+Q"); err != nil {
		t.Fatal(err)
	}
	state := waitForExit(t, terminal)
	if *state.Exited != 0 {
		t.Fatalf("exit code %d, want 0", *state.Exited)
	}
	if state.Modes["alternate_screen"] || state.Modes["bracketed_paste"] || !state.Modes["cursor_visible"] {
		t.Fatalf("the terminal was not restored: %v", state.Modes)
	}
}

func waitForExit(t *testing.T, terminal *tuitest.Client) tuitest.State {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		state, err := terminal.State()
		if err != nil {
			t.Fatal(err)
		}
		if state.Exited != nil {
			return state
		}
	}
	t.Fatal("the process did not exit")
	return tuitest.State{}
}

func openCourse(t *testing.T, terminal *tuitest.Client) {
	t.Helper()
	waitForText(t, terminal, "Lesson ready")
	if err := terminal.Press("Enter"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Execution starts at main.go:2")
}

func waitUntilGone(t *testing.T, terminal *tuitest.Client, text string) {
	t.Helper()
	var screen string
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		screen, _ = terminal.Screenshot("", tuitest.ScreenshotOptions{})
		if !strings.Contains(screen, text) {
			return
		}
	}
	t.Fatalf("%q never went away:\n%s", text, screen)
}

func agentCalls(t *testing.T, stateDirectory string) int {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(stateDirectory, "*", "agent", "invocations.jsonl"))
	if len(files) == 0 {
		return 0
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	return len(strings.Split(strings.TrimSpace(string(data)), "\n"))
}

func TestSecondLaunchShowsTheCachedLessonWithoutAskingTheAgent(t *testing.T) {
	repository, stateDirectory := fixture(t), t.TempDir()
	first := launchScreenWith(t, repository, stateDirectory, "--no-prefetch")
	openCourse(t, first)
	if err := first.Press("Ctrl+Q"); err != nil {
		t.Fatal(err)
	}
	waitForExit(t, first)
	if calls := agentCalls(t, stateDirectory); calls != 1 {
		t.Fatalf("the first launch made %d agent calls, want 1", calls)
	}
	second := launchScreenWith(t, repository, stateDirectory, "--no-prefetch")
	waitForText(t, second, "Execution starts at main.go:2")
	if calls := agentCalls(t, stateDirectory); calls != 1 {
		t.Fatalf("the second launch made %d agent calls, want 1", calls)
	}
}

func TestNextLessonIsPreparedInTheBackgroundAndUsedWithoutAnotherCall(t *testing.T) {
	repository, stateDirectory := fixture(t), t.TempDir()
	terminal := launchScreenWith(t, repository, stateDirectory)
	openCourse(t, terminal)
	for deadline := time.Now().Add(5 * time.Second); agentCalls(t, stateDirectory) < 2; time.Sleep(20 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("the next lesson was never prepared; %d agent calls", agentCalls(t, stateDirectory))
		}
	}
	prepared := agentCalls(t, stateDirectory)
	if err := terminal.Press("Ctrl+N"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Topic 2 / 2")
	if calls := agentCalls(t, stateDirectory); calls != prepared {
		t.Fatalf("moving to the prepared topic made %d new agent calls", calls-prepared)
	}
}

func readTenTopicsOfThisRepository(t *testing.T) (string, string) {
	t.Helper()
	repository, stateDirectory := "..", t.TempDir()
	output, err := runWithEnvironment(t, []string{"FAKE_AGENT_CITED_PATH=README.md"}, strings.Repeat("/next\n", 10)+"/quit\n", repository, "--agent", "codex", "--state-dir", stateDirectory, "--plain", "--goal", "learn it")
	if err != nil {
		t.Fatalf("reading topics failed: %v\n%s", err, output)
	}
	progress := saved(t, stateDirectory)["progress"].(map[string]any)
	if len(progress) < 10 {
		t.Fatalf("only %d topics were read: %v", len(progress), progress)
	}
	return repository, stateDirectory
}

func TestExamIsOfferedAfterTenTopicsWithoutAskingTheAgent(t *testing.T) {
	repository, stateDirectory := readTenTopicsOfThisRepository(t)
	calls := agentCalls(t, stateDirectory)
	terminal := launchScreenWith(t, repository, stateDirectory, "--no-prefetch")
	waitForText(t, terminal, "topics read")
	if err := terminal.Press("Shift+Tab"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Take an exam")
	if after := agentCalls(t, stateDirectory); after != calls {
		t.Fatalf("opening a saved course made %d agent calls", after-calls)
	}
}

func TestSavedCoursePagesOpenWithoutAskingTheAgent(t *testing.T) {
	repository, stateDirectory := readTenTopicsOfThisRepository(t)
	calls := agentCalls(t, stateDirectory)
	terminal := launchScreenWith(t, repository, stateDirectory, "--no-prefetch")
	waitForText(t, terminal, "topics read")
	for _, page := range []struct{ key, heading string }{
		{"Ctrl+K", "Choose your next topic"},
		{"Ctrl+S", "Follow the evidence"},
		{"Ctrl+P", "Coverage"},
	} {
		if err := terminal.Press(page.key); err != nil {
			t.Fatal(err)
		}
		waitForText(t, terminal, page.heading)
		if err := terminal.Press("Escape"); err != nil {
			t.Fatal(err)
		}
		waitForText(t, terminal, "Ask about this topic")
	}
	if after := agentCalls(t, stateDirectory); after != calls {
		t.Fatalf("opening the pages made %d agent calls", after-calls)
	}
}

func launchWithoutGoal(t *testing.T) (*tuitest.Client, string) {
	t.Helper()
	stateDirectory := t.TempDir()
	terminal := launchTerminal(t, 120, 36, nil, fixture(t), "--agent", "codex", "--state-dir", stateDirectory, "--no-prefetch")
	waitForText(t, terminal, "What do you want to learn?")
	return terminal, stateDirectory
}

func TestEnterKeepsTheSuggestedGoalOnFirstRun(t *testing.T) {
	terminal, stateDirectory := launchWithoutGoal(t)
	if err := terminal.Press("Enter"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Lesson ready")
	if state := saved(t, stateDirectory); state["goal"] == "" || state["goal"] == nil {
		t.Fatalf("the suggested goal was not saved: %v", state["goal"])
	}
}

func TestTypedGoalIsSavedOnFirstRun(t *testing.T) {
	terminal, stateDirectory := launchWithoutGoal(t)
	if err := terminal.Type("Trace the entry point"); err != nil {
		t.Fatal(err)
	}
	if err := terminal.Press("Enter"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Lesson ready")
	if goal := saved(t, stateDirectory)["goal"]; goal != "Trace the entry point" {
		t.Fatalf("saved goal %q", goal)
	}
}

func TestOtherKeysAreHeldUntilTheGoalIsChosen(t *testing.T) {
	terminal, _ := launchWithoutGoal(t)
	if err := terminal.Press("Ctrl+N"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Choose your goal first. Enter keeps the default.")
}

func TestHelpSettingsAndAboutOpenFromHomeAndEscapeCloses(t *testing.T) {
	terminal := launchScreen(t)
	for _, page := range []struct{ key, heading string }{
		{"F1", "Show this help"},
		{"F2", "TAB STYLE"},
		{"F3", "About repo-tutor"},
	} {
		if err := terminal.Press(page.key); err != nil {
			t.Fatal(err)
		}
		waitForText(t, terminal, page.heading)
		if err := terminal.Press("Escape"); err != nil {
			t.Fatal(err)
		}
		waitUntilGone(t, terminal, page.heading)
	}
}

func TestSettingsChangeIsSavedToSettingsFile(t *testing.T) {
	stateDirectory := t.TempDir()
	terminal := launchScreenWith(t, fixture(t), stateDirectory)
	if err := terminal.Press("F2"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "TAB STYLE")
	if err := terminal.Press("Up"); err != nil {
		t.Fatal(err)
	}
	if err := terminal.Press("Space"); err != nil {
		t.Fatal(err)
	}
	var data []byte
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline) && !strings.Contains(string(data), `"highlight": "purple"`); time.Sleep(20 * time.Millisecond) {
		data, _ = os.ReadFile(filepath.Join(stateDirectory, "settings.json"))
	}
	if !strings.Contains(string(data), `"highlight": "purple"`) {
		screen, _ := terminal.Screenshot("", tuitest.ScreenshotOptions{})
		t.Fatalf("settings.json does not hold the chosen highlight color:\n%s\n%s", data, screen)
	}
}

func startQuiz(t *testing.T, terminal *tuitest.Client) {
	t.Helper()
	openCourse(t, terminal)
	if err := terminal.Press("Ctrl+G"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "What does the entry point delegate?")
}

func TestSkippingAQuizQuestionNeedsASecondPressThatExpires(t *testing.T) {
	terminal := launchScreenWith(t, fixture(t), t.TempDir(), "--no-prefetch")
	startQuiz(t, terminal)
	if err := terminal.Press("Ctrl+N"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Press Ctrl+N again to skip this question.")
	waitUntilGone(t, terminal, "Press Ctrl+N again to skip this question.")
	waitForText(t, terminal, "What does the entry point delegate?")
}

func TestLeavingAQuizNeedsASecondCtrlL(t *testing.T) {
	terminal := launchScreenWith(t, fixture(t), t.TempDir(), "--no-prefetch")
	startQuiz(t, terminal)
	if err := terminal.Press("Ctrl+L"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, terminal, "Press Ctrl+L again to leave this quiz. It will not be scored.")
	if err := terminal.Press("Ctrl+L"); err != nil {
		t.Fatal(err)
	}
	waitUntilGone(t, terminal, "Press Ctrl+L again")
	waitUntilGone(t, terminal, "What does the entry point delegate?")
}
