package terminal

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

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

func frameLines(model interfaceModel) []string {
	var lines []string
	for _, line := range strings.Split(ansi.Strip(model.composer()), "\n") {
		if strings.HasPrefix(line, "╭") || strings.HasPrefix(line, "│") || strings.HasPrefix(line, "╰") {
			lines = append(lines, line)
		}
	}
	return lines
}

func TestTheInputFrameCarriesItsTitleAndItsKeys(t *testing.T) {
	model := readingFixture()
	lines := frameLines(model)
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "╭─ Ask about this topic ─") || !strings.HasSuffix(lines[len(lines)-1], "Enter send ─╯") {
		t.Fatalf("a question box should have a title on top and its key below:\n%s", strings.Join(lines, "\n"))
	}
	for _, line := range lines {
		if ansi.StringWidth(line) != model.bodyWidth() {
			t.Fatalf("every frame row should span the body (%d), got %d: %q", model.bodyWidth(), ansi.StringWidth(line), line)
		}
	}
	busy := readingFixture()
	busy.busy = true
	busy.intent = turnIntent{"lesson", "go", tutor.Overview}
	if lines := frameLines(busy); !strings.Contains(lines[0], "Preparing your lesson") || !strings.Contains(lines[len(lines)-1], "Esc stops it") {
		t.Fatalf("a running request should say so in its frame:\n%s", strings.Join(lines, "\n"))
	}
	reviewed := readingFixture()
	reviewed.application.state.Review = true
	if lines := frameLines(reviewed); !strings.Contains(lines[0], "Ask about this answer, or continue") || !strings.Contains(lines[len(lines)-1], "Enter or Ctrl+F continue") {
		t.Fatalf("feedback should offer Continue in its frame:\n%s", strings.Join(lines, "\n"))
	}
	answering := readingFixture()
	answering.application.state.Question = "Which package owns the calculation?"
	var trimmed []string
	for _, line := range strings.Split(ansi.Strip(answering.composer()), "\n") {
		trimmed = append(trimmed, strings.TrimRight(line, " "))
	}
	composer := strings.Join(trimmed, "\n")
	if !strings.Contains(composer, "NEXT QUESTION\nWhich package owns the calculation?") || !strings.Contains(frameLines(answering)[0], "Your answer") {
		t.Fatalf("a quiz question stays above a box titled Your answer:\n%s", composer)
	}
}

func TestTheGoalPromptNamesItsEnterKeyFromWhatIsTyped(t *testing.T) {
	model := pendingGoalFixture(t)
	if lines := frameLines(model); !strings.Contains(lines[0], goalQuestion) || !strings.Contains(lines[len(lines)-1], "Enter keeps the default") {
		t.Fatalf("an empty goal box keeps the default:\n%s", strings.Join(lines, "\n"))
	}
	typed := typeText(model, "Trace the persistence pipeline")
	if lines := frameLines(typed); !strings.Contains(lines[len(lines)-1], "Enter saves your goal") {
		t.Fatalf("a typed goal is saved:\n%s", strings.Join(lines, "\n"))
	}
}

func TestTheBoxGrowsWithTheTextAndNeverTakesANewLine(t *testing.T) {
	model := readingFixture()
	short := len(frameLines(model))
	long := model
	for range 40 {
		long = typeText(long, "word ")
	}
	if got := len(frameLines(long)); got <= short || got > inputMaxRows+2 {
		t.Fatalf("a long answer should grow the box up to %d rows of text: %d vs %d lines", inputMaxRows, got, short)
	}
	broken, _ := long.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
	if strings.Contains(broken.(interfaceModel).input.Value(), "\n") {
		t.Fatal("Alt+Enter should not add a line")
	}
}

func TestPastedLineBreaksBecomeSpaces(t *testing.T) {
	model := readingFixture()
	pasted, _ := model.Update(tea.PasteMsg{Content: "first line\r\nsecond   line\nthird"})
	if got := pasted.(interfaceModel).input.Value(); got != "first line second line third" {
		t.Fatalf("got %q", got)
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
