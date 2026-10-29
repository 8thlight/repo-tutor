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
	terminal := tuitesttest.New(t, tuitesttest.Options{
		Program: binary,
		Args:    append([]string{repository, "--agent", "codex", "--state-dir", stateDirectory, "--goal", "learn it"}, extra...),
		Spawn: tuitest.SpawnOptions{
			Cols: tuitest.Ptr(uint16(120)), Rows: tuitest.Ptr(uint16(36)), WaitReady: tuitest.Ptr(false),
			Env: map[string]string{"PATH": agentDirectory + ":" + "/usr/bin:/bin", "HOME": t.TempDir(), "FAKE_AGENT_PID_FILE": filepath.Join(t.TempDir(), "agent.pid")},
		},
	})
	waitForText(t, terminal, "Understanding")
	return terminal
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
