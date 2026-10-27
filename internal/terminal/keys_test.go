package terminal

import (
	"repo-tutor/internal/repository"
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
	if line != "F1 Help · F2 Settings · F3 About · Esc back or stop · Ctrl+Q quit" {
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
		90: "Ctrl+K topics · Ctrl+S sources · F1 Help · F2 Settings · F3 About · Esc back or stop · Ctrl+Q quit",
		80: "Ctrl+K topics · Ctrl+S sources · F1 Help · F2 Settings · F3 About · Ctrl+Q quit",
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
	if lines := frameLines(reviewed); !strings.Contains(lines[0], "Ask about this answer, or continue") || !strings.Contains(lines[len(lines)-1], "Enter or Ctrl+N continue") {
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

func TestABlankRowSeparatesTheRuleFromTheInput(t *testing.T) {
	for name, model := range map[string]interfaceModel{"a question box": readingFixture(), "the goal prompt": pendingGoalFixture(t)} {
		lines := strings.Split(ansi.Strip(model.composer()), "\n")
		rule := -1
		for index, line := range lines {
			if strings.HasPrefix(line, "────") {
				rule = index
				break
			}
		}
		if rule < 0 || rule+2 >= len(lines) || strings.TrimSpace(lines[rule+1]) != "" || strings.TrimSpace(lines[rule+2]) == "" {
			t.Fatalf("%s: one blank row should sit under the rule:\n%s", name, strings.Join(lines, "\n"))
		}
	}
}

func TestABlankRowAlsoSeparatesAQuizQuestionFromTheInput(t *testing.T) {
	model := readingFixture()
	model.application.state.Question = "Which package owns the calculation?"
	lines := strings.Split(ansi.Strip(model.composer()), "\n")
	question := -1
	for index, line := range lines {
		if strings.Contains(line, "Which package owns the calculation?") {
			question = index
		}
	}
	if question < 0 || strings.TrimSpace(lines[question+1]) != "" || !strings.HasPrefix(lines[question+2], "╭") {
		t.Fatalf("a blank row should sit between the question and the box:\n%s", strings.Join(lines, "\n"))
	}
	if strings.TrimSpace(lines[1]) != "" {
		t.Fatalf("a blank row should also sit under the rule:\n%s", strings.Join(lines, "\n"))
	}
}

func TestTheFooterLeadsWithTheKeysOfTheCurrentPage(t *testing.T) {
	model := interfaceFixture()
	model.width = 120
	opened, _ := model.openPicker("topics", "")
	list := ansi.Strip(opened.(interfaceModel).footerKeys(120))
	if !strings.HasPrefix(list, "↑↓ move · / search · Enter open · F1 Help") {
		t.Fatalf("a list should lead with its own keys: %q", list)
	}
	if strings.Contains(list, "A show answer") {
		t.Fatal("only the question index offers A")
	}
	quiz := revealFixture()
	quiz.width = 120
	index, _ := quiz.openPicker("quiz", "")
	if line := ansi.Strip(index.(interfaceModel).footerKeys(140)); !strings.Contains(line, "Enter open · A show answer · F1 Help") {
		t.Fatalf("the question index should offer A: %q", line)
	}
	file, _ := model.openFile(repository.Location{Path: "main.go", Line: 1})
	if line := ansi.Strip(file.(interfaceModel).footerKeys(120)); !strings.HasPrefix(line, "Enter or Esc back · T teach this file · F1 Help") || strings.Contains(line, "rendered or source") {
		t.Fatalf("a code file should offer back and teach: %q", line)
	}
	markdown, _ := model.openFile(repository.Location{Path: "README.md", Line: 1})
	if line := ansi.Strip(markdown.(interfaceModel).footerKeys(140)); !strings.Contains(line, "T teach this file · v rendered or source · F1 Help") {
		t.Fatalf("a Markdown file should also offer v: %q", line)
	}
}

func TestANarrowFooterDropsPageKeysBeforeItDropsQuit(t *testing.T) {
	model := interfaceFixture()
	file, _ := model.openFile(repository.Location{Path: "README.md", Line: 1})
	opened := file.(interfaceModel)
	for _, width := range []int{80, 50, 40, 30} {
		line := ansi.Strip(opened.footerKeys(width))
		if ansi.StringWidth(line) > width || !strings.Contains(line, "Ctrl+Q quit") && width >= 30 && !strings.Contains(line, "F1") {
			t.Fatalf("width %d: %q", width, line)
		}
	}
	if line := ansi.Strip(opened.footerKeys(40)); !strings.HasSuffix(line, "Ctrl+Q quit") {
		t.Fatalf("at 40 columns the quit key should survive: %q", line)
	}
}

func TestNavigationHelpNeverAppearsInTheNoticeLine(t *testing.T) {
	model := interfaceFixture()
	for name, open := range map[string]func() interfaceModel{
		"a list": func() interfaceModel { next, _ := model.openPicker("sources", ""); return next.(interfaceModel) },
		"a code file": func() interfaceModel {
			next, _ := model.openFile(repository.Location{Path: "main.go", Line: 1})
			return next.(interfaceModel)
		},
		"a Markdown file": func() interfaceModel {
			next, _ := model.openFile(repository.Location{Path: "README.md", Line: 1})
			return next.(interfaceModel)
		},
	} {
		notice := open().notice
		for _, forbidden := range []string{"Enter or Esc", "↑↓", "/ search", "T teaches", "Enter open"} {
			if strings.Contains(notice, forbidden) {
				t.Fatalf("%s: the notice should say what happened and not how to navigate: %q", name, notice)
			}
		}
	}
}
