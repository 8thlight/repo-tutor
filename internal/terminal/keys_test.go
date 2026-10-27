package terminal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/tutor"
)

func TestTheFooterCarriesTheAppWideKeysOnAWideTerminal(t *testing.T) {
	model := interfaceFixture()
	model.width = 120
	line := ansi.Strip(model.footerKeys(100))
	if line != "F1 Help · F2 About · F3 Settings · Esc back or stop · Ctrl+Q quit" {
		t.Fatalf("got %q", line)
	}
	if strings.Contains(line, "Ctrl+K") || strings.Contains(line, "Enter") {
		t.Fatalf("the sidebar and composer own those keys: %q", line)
	}
}

func TestANarrowFooterAddsTopicsAndSourcesAndShortensFromTheRight(t *testing.T) {
	model := interfaceFixture()
	model.width = 90
	for width, want := range map[int]string{
		90: "Ctrl+K topics · Ctrl+S sources · F1 Help · F2 About · F3 Settings · Esc back or stop · Ctrl+Q quit",
		80: "Ctrl+K topics · Ctrl+S sources · F1 Help · F2 About · F3 Settings · Ctrl+Q quit",
		60: "Ctrl+K topics · Ctrl+S sources · F1 Help · Ctrl+Q quit",
		40: "F1 Help · Ctrl+Q quit",
	} {
		got := ansi.Strip(model.footerKeys(width))
		if got != want && ansi.StringWidth(got) > width {
			t.Fatalf("width %d: %q is too wide", width, got)
		}
		if width != 90 && ansi.StringWidth(want) <= width && got != want {
			t.Fatalf("width %d: got %q want %q", width, got, want)
		}
	}
}

func TestANarrowFooterDropsSourcesWhileMoving(t *testing.T) {
	model := movingFixture(t)
	model.width = 90
	if strings.Contains(ansi.Strip(model.footerKeys(100)), "Ctrl+S") {
		t.Fatal("sources are held during a move, so the footer should not offer them")
	}
}

func TestTheComposerDividerNamesTheKeysThatActOnTheBox(t *testing.T) {
	model := readingFixture()
	if divider := ansi.Strip(model.composerDivider()); !strings.Contains(divider, "Enter send · Alt+Enter new line") || !strings.HasSuffix(divider, " ──") {
		t.Fatalf("a question box should show its typing keys:\n%s", divider)
	}
	busy := readingFixture()
	busy.busy = true
	if divider := ansi.Strip(busy.composerDivider()); !strings.Contains(divider, "Esc stops it") {
		t.Fatalf("a running request should offer Esc:\n%s", divider)
	}
	reviewed := readingFixture()
	reviewed.application.state.Review = true
	if divider := ansi.Strip(reviewed.composerDivider()); !strings.Contains(divider, "Enter or Ctrl+F continue") {
		t.Fatalf("feedback should offer Continue:\n%s", divider)
	}
	if width := ansi.StringWidth(ansi.Strip(model.composerDivider())); width != model.bodyWidth() {
		t.Fatalf("the divider should span the body, got %d of %d", width, model.bodyWidth())
	}
}

func TestTheQuestionBoxHasNoPlaceholderButTheGoalPromptDoes(t *testing.T) {
	if interfaceFixture().input.Placeholder != "" {
		t.Fatal("the label above the box already says what it is for")
	}
	if pendingGoalFixture(t).input.Placeholder != tutor.DefaultGoal {
		t.Fatal("the goal prompt shows the default goal as its placeholder")
	}
}

func TestNoticesNameWhatHappenedWithoutTheWordTurn(t *testing.T) {
	for action, want := range map[string]string{
		"lesson": "Lesson ready", "quiz": "Question ready", "quiz-next": "Question ready", "probe": "Question ready",
		"answer": "Answer reviewed", "question": "Reply ready",
	} {
		if got := finishedNotice(action); got != want {
			t.Fatalf("%s: got %q want %q", action, got, want)
		}
	}
	model := interfaceFixture()
	if model.notice != "" {
		t.Fatalf("the notice starts empty, got %q", model.notice)
	}
	model.busy = true
	if strings.Contains(strings.ToLower(model.runningNotice()), "turn") {
		t.Fatal("the running notice should not say turn")
	}
}
