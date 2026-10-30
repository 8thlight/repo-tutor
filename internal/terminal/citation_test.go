package terminal

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
)

func citedFixture() interfaceModel {
	model := interfaceFixture()
	model.application.state.Turns[0].Reply.Citations = []repository.Location{{Path: "main.go", Line: 2}, {Path: "README.md", Line: 1}}
	model.application.state.Question = ""
	model.application.settings.MarkdownOpensRendered = false
	model.rebuildContent()
	return model
}

func digit(model interfaceModel, number rune) interfaceModel {
	next, _ := model.Update(tea.KeyPressMsg{Code: number, Text: string(number)})
	return next.(interfaceModel)
}

func controlDigit(model interfaceModel, number rune) interfaceModel {
	pressed, _ := press(model, number, tea.ModCtrl)
	return pressed
}

func TestADigitOpensItsCitationWhenTheBoxIsEmpty(t *testing.T) {
	model := digit(citedFixture(), '1')
	if model.page != "source" || model.sourcePath != "main.go" || model.sourceLine != 2 {
		t.Fatalf("1 should open main.go:2, got page %s %s:%d", model.page, model.sourcePath, model.sourceLine)
	}
	model = digit(citedFixture(), '2')
	if model.page != "source" || model.sourcePath != "README.md" {
		t.Fatalf("2 should open README.md, got page %s %s", model.page, model.sourcePath)
	}
}

func TestThePressedDigitReturnsToTheLessonAndAnotherDigitSwitches(t *testing.T) {
	opened := digit(citedFixture(), '1')
	if back := digit(opened, '1'); back.page != "lesson" {
		t.Fatalf("1 again should return to the lesson, got %s", back.page)
	}
	switched := digit(opened, '2')
	if switched.page != "source" || switched.sourcePath != "README.md" {
		t.Fatalf("2 should switch to README.md, got %s %s", switched.page, switched.sourcePath)
	}
	if back := digit(switched, '2'); back.page != "lesson" {
		t.Fatalf("2 again should return to the lesson, got %s", back.page)
	}
	escaped, _ := press(switched, tea.KeyEscape, 0)
	if escaped.page != "lesson" {
		t.Fatalf("Esc should also return to the lesson, got %s", escaped.page)
	}
}

func TestTheLessonComesBackAtTheSameScrollPosition(t *testing.T) {
	model := citedFixture()
	model.application.state.Turns[0].Reply.Explanation = strings.Repeat("A sentence that fills a line of the lesson [1].\n\n", 60)
	model.rebuildContent()
	model.viewport.SetYOffset(23)
	before := model.viewport.YOffset()
	if before == 0 {
		t.Fatal("the fixture lesson should be long enough to scroll")
	}
	back := digit(digit(model, '1'), '1')
	if back.page != "lesson" || back.viewport.YOffset() != before {
		t.Fatalf("the lesson should return at offset %d, got page %s offset %d", before, back.page, back.viewport.YOffset())
	}
}

func TestABareDigitTypesWhenTheBoxHoldsText(t *testing.T) {
	model := citedFixture()
	model.input.SetValue("how many")
	typed := digit(model, '1')
	if typed.page != "lesson" || typed.input.Value() != "how many1" {
		t.Fatalf("the digit should type into the draft, got page %s and %q", typed.page, typed.input.Value())
	}
}

func TestControlDigitOpensACitationEvenWithADraft(t *testing.T) {
	model := citedFixture()
	model.input.SetValue("a draft")
	opened := controlDigit(model, '2')
	if opened.page != "source" || opened.sourcePath != "README.md" {
		t.Fatalf("Ctrl+2 should open README.md, got %s %s", opened.page, opened.sourcePath)
	}
	if back := controlDigit(opened, '2'); back.page != "lesson" || back.input.Value() != "a draft" {
		t.Fatalf("Ctrl+2 again should return with the draft kept, got %s %q", back.page, back.input.Value())
	}
}

func TestABareDigitTypesWhileAQuestionWaitsForAnAnswer(t *testing.T) {
	model := citedFixture()
	model.application.state.Question = "How many packages does it have?"
	typed := digit(model, '1')
	if typed.page != "lesson" || typed.input.Value() != "1" {
		t.Fatalf("an answer may start with a digit, got page %s and %q", typed.page, typed.input.Value())
	}
	if opened := controlDigit(model, '1'); opened.page != "source" {
		t.Fatalf("Ctrl+1 should still open the citation, got %s", opened.page)
	}
}

func TestADigitWithNoCitationChangesNothingAndSaysHowManyThereAre(t *testing.T) {
	model := digit(citedFixture(), '3')
	if model.page != "lesson" || model.input.Value() != "" || model.notice != "This lesson cites 2 sources." {
		t.Fatalf("got page %s, draft %q, notice %q", model.page, model.input.Value(), model.notice)
	}
	one := citedFixture()
	one.application.state.Turns[0].Reply.Citations = one.application.state.Turns[0].Reply.Citations[:1]
	if got := digit(one, '2').notice; got != "This lesson cites 1 source." {
		t.Fatalf("one citation should read as 1 source, got %q", got)
	}
}

func TestDigitsAreHeldWhileARequestMovesToAnotherTopic(t *testing.T) {
	model := citedFixture()
	model.busy = true
	model.intent.topic = "somewhere-else"
	held := digit(model, '1')
	if held.page != "lesson" || held.notice != sourcesWaitNotice {
		t.Fatalf("the digit should be held with the wait notice, got %s %q", held.page, held.notice)
	}
}

func TestADigitOnAnEarlierStepOpensThatStepsCitation(t *testing.T) {
	model := citedFixture()
	later := model.application.state.Turns[0]
	later.Reply.Citations = []repository.Location{{Path: "go.mod", Line: 1}}
	model.application.state.Turns = append(model.application.state.Turns, later)
	model.history = 0
	if opened := digit(model, '1'); opened.sourcePath != "main.go" {
		t.Fatalf("step one cites main.go first, got %s", opened.sourcePath)
	}
	model.history = 1
	if opened := digit(model, '1'); opened.sourcePath != "go.mod" {
		t.Fatalf("step two cites go.mod, got %s", opened.sourcePath)
	}
}

func TestDigitsOnAFileOpenedFromSourcesDoNotToggle(t *testing.T) {
	model := citedFixture()
	opened, _ := model.openFile(repository.Location{Path: "main.go", Line: 1})
	file := opened.(interfaceModel)
	if file.sourceCitation != 0 {
		t.Fatalf("a file opened from Sources has no citation number, got %d", file.sourceCitation)
	}
	if still := digit(file, '1'); still.page != "source" {
		t.Fatalf("1 should not close a file that did not come from a citation, got %s", still.page)
	}
}

func TestTheSidebarOffersTheDigitsOnlyWhenTheyCanAct(t *testing.T) {
	offered := func(model interfaceModel) bool {
		return strings.Contains(ansi.Strip(model.sidebarShortcuts()), "1-9  Open a source")
	}
	if !offered(citedFixture()) {
		t.Fatal("the sidebar should offer the digits on a lesson with citations and an empty box")
	}
	typing := citedFixture()
	typing.input.SetValue("a draft")
	if offered(typing) {
		t.Fatal("the digits type while the box holds text, so the sidebar should not offer them")
	}
	asking := citedFixture()
	asking.application.state.Question = "How many?"
	if offered(asking) {
		t.Fatal("the digits type while a question waits, so the sidebar should not offer them")
	}
	bare := citedFixture()
	bare.application.state.Turns[0].Reply.Citations = nil
	if offered(bare) {
		t.Fatal("with no citations there is nothing to open")
	}
}

func TestACitationSourceNamesItsDigitAsTheWayBack(t *testing.T) {
	text := ansi.Strip(digit(citedFixture(), '2').sidebarShortcuts())
	for _, expected := range []string{"2  Back to the lesson", "1-9  Another source"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("the sidebar should say %q:\n%s", expected, text)
		}
	}
}

func TestHelpExplainsTheCitationDigitsAndTheKeyThatAlwaysWorks(t *testing.T) {
	for _, expected := range []string{"1-9             Open a cited source when the box is empty", "Ctrl+1-9        Open a cited source with a draft", "Ctrl+S          Browse cited sources"} {
		if !strings.Contains(interfaceHelp, expected) {
			t.Fatalf("help should say %q", expected)
		}
	}
}

func openSources(model interfaceModel) interfaceModel {
	pressed, _ := press(model, 's', tea.ModCtrl)
	return pressed
}

func TestTheSourcesListNumbersEachCitationAndAdmitsTwoInOneFile(t *testing.T) {
	model := citedFixture()
	model.application.state.Turns[0].Reply.Citations = []repository.Location{{Path: "main.go", Line: 2}, {Path: "main.go", Line: 1}, {Path: "README.md", Line: 1}}
	view := ansi.Strip(openSources(model).picker.View())
	for _, expected := range []string{"[1]  main.go", "[2]  main.go", "[3]  README.md", "Cited in this lesson · line 2", "Cited in this lesson · line 1"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("the Sources list is missing %q:\n%s", expected, view)
		}
	}
}

func TestADigitInTheSourcesListOpensThatCitationAndEscComesBackToTheList(t *testing.T) {
	list := openSources(citedFixture())
	opened := digit(list, '2')
	if opened.page != "source" || opened.sourcePath != "README.md" {
		t.Fatalf("2 should open README.md, got %s %s", opened.page, opened.sourcePath)
	}
	if back, _ := press(opened, tea.KeyEscape, 0); back.page != "sources" {
		t.Fatalf("Esc should return to the Sources list, got %s", back.page)
	}
}

func TestACitationDigitInTheSourcesListIgnoresWhatIsNotThere(t *testing.T) {
	missing := digit(openSources(citedFixture()), '5')
	if missing.page != "sources" || missing.notice != "This lesson cites 2 sources." {
		t.Fatalf("got page %s and notice %q", missing.page, missing.notice)
	}
}

func TestWhileFilteringTheSourcesListADigitTypesIntoTheFilter(t *testing.T) {
	list := openSources(citedFixture())
	filtering, _ := press(list, '/', 0)
	typed := digit(filtering, '1')
	if typed.page != "sources" {
		t.Fatalf("a digit typed into the filter must not open a source, got %s", typed.page)
	}
}
