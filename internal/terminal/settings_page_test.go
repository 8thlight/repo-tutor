package terminal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func settingsPageFixture(t *testing.T) interfaceModel {
	model := interfaceFixture()
	model.application.settingsBase = t.TempDir()
	opened, _ := press(model, tea.KeyF2, 0)
	return opened
}

func TestF2OpensSettingsAndPressingItAgainReturns(t *testing.T) {
	model := interfaceFixture()
	opened, _ := press(model, tea.KeyF2, 0)
	if opened.page != "settings" || opened.activeTab() == "course" {
		t.Fatalf("F2 should open Settings with no tab highlighted, got %s on %s", opened.page, opened.activeTab())
	}
	closed, _ := press(opened, tea.KeyF2, 0)
	if closed.page != "lesson" {
		t.Fatalf("F2 again should return to the origin, got %s", closed.page)
	}
}

func TestSpaceChoosesAndSavesTheFocusedChoice(t *testing.T) {
	model := settingsPageFixture(t)
	model, _ = press(model, tea.KeyDown, 0)
	model, _ = press(model, tea.KeyDown, 0)
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.TabStyle != "slanted" || model.notice != settingsSavedNotice {
		t.Fatalf("Space should choose the focused style: %+v %q", model.application.settings, model.notice)
	}
	saved, _, _ := session.LoadSettings(model.application.settingsBase)
	if saved.TabStyle != "slanted" {
		t.Fatalf("the choice should be saved at once: %+v", saved)
	}
	if !strings.Contains(ansi.Strip(model.settingsContent()), "(•) Slanted") {
		t.Fatalf("the chosen row should show (•):\n%s", ansi.Strip(model.settingsContent()))
	}
}

func TestEnterDoesNotChangeASetting(t *testing.T) {
	model := settingsPageFixture(t)
	model, _ = press(model, tea.KeyDown, 0)
	before := model.application.settings
	model, _ = press(model, tea.KeyEnter, 0)
	if model.application.settings != before || model.page != "settings" {
		t.Fatal("Enter should leave the setting and the page alone")
	}
}

func TestMarkdownChoiceIsSavedAndUsed(t *testing.T) {
	model := settingsPageFixture(t)
	for range 8 {
		model, _ = press(model, tea.KeyDown, 0)
	}
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.MarkdownOpensRendered {
		t.Fatal("choosing Source should turn rendering off")
	}
	if model.rendersMarkdown("README.md") {
		t.Fatal("a file should open as source after the choice")
	}
}

func TestReadOnlySettingsChangeNothing(t *testing.T) {
	model := settingsPageFixture(t)
	model.application.settingsStatus.ReadOnly = true
	model, _ = press(model, tea.KeyDown, 0)
	before := model.application.settings
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings != before || model.notice != settingsReadOnlyNotice {
		t.Fatalf("a read-only file must not change: %q", model.notice)
	}
	if !strings.Contains(ansi.Strip(model.settingsContent()), settingsReadOnlyNotice) {
		t.Fatal("the page should repeat the read-only notice at the top")
	}
}

func TestAFailedSaveKeepsTheChangeForTheSession(t *testing.T) {
	model := settingsPageFixture(t)
	blocker := filepath.Join(t.TempDir(), "file")
	os.WriteFile(blocker, nil, 0600)
	model.application.settingsBase = filepath.Join(blocker, "inside")
	model, _ = press(model, tea.KeyDown, 0)
	model, _ = press(model, tea.KeyDown, 0)
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.TabStyle != "slanted" || model.notice != settingsUnsavedNotice {
		t.Fatalf("a failed write should keep the change and say so: %+v %q", model.application.settings, model.notice)
	}
}

func TestPowerlineRowsAlwaysCarryTheFontNote(t *testing.T) {
	content := ansi.Strip(settingsPageFixture(t).settingsContent())
	if strings.Count(content, powerlineNote) != 5 {
		t.Fatalf("every Powerline style should say a Powerline font is needed:\n%s", content)
	}
}

func TestSamplesMatchTheRealBarAndMoveUnderLabelsWhenNarrow(t *testing.T) {
	model := settingsPageFixture(t)
	model.noColor = false
	wide := ansi.Strip(model.settingsContent())
	if !strings.Contains(wide, "Rounded"+strings.Repeat(" ", 7)+ansi.Strip(model.tabs("rounded", "course"))) {
		t.Fatalf("a sample should come from the same function as the bar:\n%s", wide)
	}
	model.width = 70
	narrow := ansi.Strip(model.settingsContent())
	if !strings.Contains(narrow, "Rounded\n      "+ansi.Strip(model.tabs("rounded", "course"))) {
		t.Fatalf("a narrow page should put the sample under its label:\n%s", narrow)
	}
}

func TestSpaceStillTypesInTheLesson(t *testing.T) {
	model := interfaceFixture()
	model.input.SetValue("one")
	next, _ := model.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	model = next.(interfaceModel)
	if model.input.Value() != "one " {
		t.Fatalf("Space in the lesson should type a space, got %q", model.input.Value())
	}
}

func TestHomeListsSettingsBeforeAbout(t *testing.T) {
	model := newLearnerFixture()
	var places []string
	for _, entry := range model.homeEntries() {
		places = append(places, entry.place)
	}
	joined := strings.Join(places, ",")
	if !strings.HasSuffix(joined, "settings,about") {
		t.Fatalf("Settings should sit just before About: %s", joined)
	}
	opened, _ := model.openHomeEntry(homeEntry{place: placeSettings})
	if opened.(interfaceModel).page != "settings" {
		t.Fatal("the Home entry should open Settings")
	}
}

func TestNoColorSaysTabStyleDoesNotApply(t *testing.T) {
	model := settingsPageFixture(t)
	model.noColor = true
	if !strings.Contains(ansi.Strip(model.settingsContent()), noColorNote) {
		t.Fatal("with color off the page should say tab style does not apply")
	}
	model.noColor = false
	if strings.Contains(ansi.Strip(model.settingsContent()), noColorNote) {
		t.Fatal("with color on the note should not appear")
	}
}

func TestHighlightChoiceIsSavedAndChangesTheAccent(t *testing.T) {
	model := settingsPageFixture(t)
	model.noColor = false
	before := model.palette().accent
	for range 17 {
		model, _ = press(model, tea.KeyDown, 0)
	}
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.Highlight != "red" || model.palette().accent == before {
		t.Fatalf("choosing Red should change the highlight: %+v", model.application.settings)
	}
	saved, _, _ := session.LoadSettings(model.application.settingsBase)
	if saved.Highlight != "red" {
		t.Fatalf("the highlight should be saved: %+v", saved)
	}
}

func TestHighlightRowsShowASwatchAndTheTerminalNote(t *testing.T) {
	model := settingsPageFixture(t)
	model.noColor = false
	content := ansi.Strip(model.settingsContent())
	if strings.Count(content, "██") != 7 || !strings.Contains(content, highlightTerminalNote) {
		t.Fatalf("each highlight should show a swatch under the terminal note:\n%s", content)
	}
	model.noColor = true
	if !strings.Contains(ansi.Strip(model.settingsContent()), noColorHighlightNote) {
		t.Fatal("with color off the page should say highlight color does not apply")
	}
}

func TestTheFocusedRowStaysInViewOnAShortScreen(t *testing.T) {
	model := settingsPageFixture(t)
	model.width, model.height = 100, 20
	model.resize()
	for range len(settingChoices()) - 2 {
		model, _ = press(model, tea.KeyDown, 0)
	}
	_, focusLine := model.settingsLayout()
	offset := model.viewport.YOffset()
	if focusLine < offset || focusLine >= offset+model.viewport.Height() {
		t.Fatalf("focus line %d should sit inside the view %d..%d", focusLine, offset, offset+model.viewport.Height())
	}
}

func TestThePageNamesItsKeysAndLeavesTheMarkdownViewKeyToTheFile(t *testing.T) {
	content := ansi.Strip(settingsPageFixture(t).settingsContent())
	if !strings.HasPrefix(content, settingsKeysNote) {
		t.Fatalf("the page should open with its keys:\n%s", content)
	}
	if strings.Contains(content, "Press v") {
		t.Fatal("the v key belongs to the file view and not to this page")
	}
}

func TestQuizLengthIsChosenSavedAndUsedForTheNextQuiz(t *testing.T) {
	model := settingsPageFixture(t)
	for range 11 {
		model, _ = press(model, tea.KeyDown, 0)
	}
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.QuizLength != 10 {
		t.Fatalf("choosing 10 questions should set the length: %+v", model.application.settings)
	}
	saved, _, _ := session.LoadSettings(model.application.settingsBase)
	if saved.QuizLength != 10 {
		t.Fatalf("the length should be saved: %+v", saved)
	}
	if !strings.Contains(ansi.Strip(model.settingsContent()), "(•) 10 questions") {
		t.Fatalf("the chosen length should show (•):\n%s", ansi.Strip(model.settingsContent()))
	}
	updated := model.application.recordReply("quiz", "go", tutor.Overview, "", tutor.Reply{Question: "First?", Assessment: "none"})
	if updated.Quiz == nil || updated.Quiz.Goal != 10 {
		t.Fatalf("a new quiz should take the chosen length: %+v", updated.Quiz)
	}
	if !strings.Contains(ansi.Strip(readingFixtureWith(model).actionLine()), "Quiz this topic (10)") {
		t.Fatal("the action line should name the chosen length")
	}
}

func readingFixtureWith(model interfaceModel) interfaceModel {
	reading := readingFixture()
	reading.application.settings = model.application.settings
	return reading
}

func TestExamLengthIsChosenSavedAndExplainedOnce(t *testing.T) {
	model := settingsPageFixture(t)
	for range 13 {
		model, _ = press(model, tea.KeyDown, 0)
	}
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.ExamLength != 20 {
		t.Fatalf("choosing 20 questions should set the exam length: %+v", model.application.settings)
	}
	saved, _, _ := session.LoadSettings(model.application.settingsBase)
	if saved.ExamLength != 20 {
		t.Fatalf("the exam length should be saved: %+v", saved)
	}
	content := ansi.Strip(model.settingsContent())
	if strings.Count(content, examLengthNote) != 1 || !strings.Contains(content, "(•) 20 questions") {
		t.Fatalf("the group should explain itself once and show the choice:\n%s", content)
	}
}

func TestEveryTopicReadIsAnExamLengthChoice(t *testing.T) {
	model := settingsPageFixture(t)
	for range 15 {
		model, _ = press(model, tea.KeyDown, 0)
	}
	model, _ = press(model, tea.KeySpace, 0)
	if model.application.settings.ExamLength != 0 {
		t.Fatalf("the last exam length means no cap: %+v", model.application.settings)
	}
	if !strings.Contains(ansi.Strip(model.settingsContent()), "(•) Every topic read") {
		t.Fatalf("the choice should read Every topic read:\n%s", ansi.Strip(model.settingsContent()))
	}
	saved, _, _ := session.LoadSettings(model.application.settingsBase)
	if saved.ExamLength != 0 {
		t.Fatalf("the choice should be saved, got %d", saved.ExamLength)
	}
}
