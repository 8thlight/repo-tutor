package terminal

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTabTextKeepsReadableContrastOnBothFills(t *testing.T) {
	for _, dark := range []bool{true, false} {
		model := interfaceFixture()
		model.noColor, model.dark = false, dark
		palette := model.palette()
		for name, fill := range map[string]color.Color{"active": palette.accent, "inactive": palette.muted} {
			if ratio := contrastBetween(palette.cutout, fill); ratio < 4.5 {
				t.Fatalf("dark=%t %s tab text contrast is %.2f, want at least 4.5", dark, name, ratio)
			}
		}
	}
}

func TestTabTextUsesTheReportedTerminalBackground(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	model.background = color.RGBA{R: 30, G: 30, B: 30, A: 255}
	if model.palette().cutout != model.background {
		t.Fatal("tab text should use the background the terminal reported")
	}
}

func TestTheTabLineKeepsOnlyTheTabHint(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for _, width := range []int{120, 80, 44} {
		bar := ansi.Strip(model.tabBar(width))
		if ansi.StringWidth(bar) > width || !strings.HasSuffix(bar, "Tab Next") || strings.Contains(bar, "F1") {
			t.Fatalf("width %d: the tab line should end with only the Tab hint:\n%q", width, bar)
		}
	}
	if bar := ansi.Strip(model.tabBar(34)); strings.Contains(bar, "Tab") {
		t.Fatalf("a bar with no room should drop the hint:\n%q", bar)
	}
}

func TestTabBarMarksTheActiveTabWithoutColor(t *testing.T) {
	model := interfaceFixture()
	model.noColor = true
	bar := ansi.Strip(model.tabBar(120))
	if !strings.Contains(bar, "[Course]") || strings.Contains(bar, "[Home]") {
		t.Fatalf("only the active tab should be bracketed without color:\n%q", bar)
	}
}

func withTerminalColors(model interfaceModel, background, foreground string) interfaceModel {
	model.noColor = false
	model.background, model.foreground = lipgloss.Color(background), lipgloss.Color(foreground)
	return model
}

func TestJoinedTabsTouchAndEndWithTheirSeparator(t *testing.T) {
	model := withTerminalColors(interfaceFixture(), "#09090B", "#E4E4E7")
	for style, glyph := range map[string]string{"joined_slanted": "", "joined_rounded": "", "joined_arrows": ""} {
		want := " Home " + glyph + " Course " + glyph + " Quiz " + glyph + " Progress " + glyph
		if got := ansi.Strip(model.tabs(style, "course")); got != want {
			t.Fatalf("%s: got %q want %q", style, got, want)
		}
	}
}

func TestEveryBoundaryInAJoinedBarShowsAColorChange(t *testing.T) {
	for _, pair := range [][2]string{{"#09090B", "#E4E4E7"}, {"#FAFAFA", "#18181B"}} {
		model := withTerminalColors(interfaceFixture(), pair[0], pair[1])
		for _, page := range []string{"home", "course", "quiz", "progress", "settings"} {
			fills := model.joinedFills(page)
			for index := 1; index < len(fills); index++ {
				if hexOf(fills[index-1]) == hexOf(fills[index]) {
					t.Fatalf("%v active=%s: tabs %d and %d share %s, so the separator would vanish", pair, page, index-1, index, hexOf(fills[index]))
				}
			}
		}
	}
}

func TestJoinedFillsStepFromTheFirstShareToTheStrongestAlongTheReportedColors(t *testing.T) {
	model := withTerminalColors(interfaceFixture(), "#000000", "#FFFFFF")
	fills := model.joinedFills("settings")
	for index, want := range []string{"#262626", "#373737", "#484848", "#595959"} {
		if got := hexOf(fills[index]); got != want {
			t.Fatalf("tab %d is %s, want %s", index, got, want)
		}
	}
}

func TestEveryInactiveFillKeepsTheTerminalForegroundReadable(t *testing.T) {
	for _, pair := range [][2]string{{"#000000", "#FFFFFF"}, {"#09090B", "#E4E4E7"}, {"#1E1E1E", "#FFFFFF"}, {"#282C34", "#FFFFFF"}, {"#FFFFFF", "#000000"}, {"#FAFAFA", "#18181B"}, {"#EEEEEE", "#000000"}} {
		model := withTerminalColors(interfaceFixture(), pair[0], pair[1])
		for index, fill := range model.joinedFills("settings") {
			if ratio := contrastBetween(fill, model.foreground); ratio < 4.5 {
				t.Fatalf("%v tab %d (%s) has contrast %.2f with the foreground", pair, index, hexOf(fill), ratio)
			}
		}
	}
}

func TestPowerlineStylesDrawAsBlockUntilTheTerminalReportsItsColors(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for _, style := range []string{"slanted", "rounded", "joined_slanted", "joined_rounded", "joined_arrows"} {
		if got := ansi.Strip(model.tabs(style, "course")); got != "  Home   Course   Quiz   Progress " && got != " Home   Course   Quiz   Progress " {
			t.Fatalf("%s should fall back to the block tabs, got %q", style, got)
		}
	}
	if !strings.Contains(model.tabs("block", "course"), "\x1b[7m") {
		t.Fatalf("the fallback should use reverse video: %q", model.tabs("block", "course"))
	}
}

func TestInactiveBlockTabsUseTheReportedColorsBlended(t *testing.T) {
	model := withTerminalColors(interfaceFixture(), "#000000", "#FFFFFF")
	inactive := lipgloss.NewStyle().Background(model.mixed(inactiveTabShare)).Foreground(model.foreground).Render(" Home ")
	if !strings.Contains(model.tabs("block", "course"), inactive) {
		t.Fatalf("an inactive tab should be the background blended toward the foreground: %q", model.tabs("block", "course"))
	}
}

func TestJoinedBarReachesTheTerminalEdgeAndOtherStylesKeepTheMargin(t *testing.T) {
	row := func(style string) string {
		model := withTerminalColors(interfaceFixture(), "#09090B", "#E4E4E7")
		model.application.settings.TabStyle = style
		for _, line := range strings.Split(ansi.Strip(model.View().Content), "\n") {
			if strings.Contains(line, "Home") && strings.Contains(line, "Course") {
				return line
			}
		}
		return ""
	}
	if line := row("joined_arrows"); !strings.HasPrefix(line, " Home") {
		t.Fatalf("a joined bar should start at the first column: %q", line)
	}
	if line := row("block"); !strings.HasPrefix(line, "   Home") {
		t.Fatalf("a block bar should keep the margin: %q", line)
	}
}

func tabRows(model interfaceModel) (string, string) {
	lines := strings.Split(ansi.Strip(model.View().Content), "\n")
	for index, line := range lines {
		if strings.Contains(line, "Home") && strings.Contains(line, "Course") && index+1 < len(lines) {
			return line, lines[index+1]
		}
	}
	return "", ""
}

func rawTabRule(model interfaceModel) string {
	_, rule := tabRows(model)
	for _, line := range strings.Split(model.View().Content, "\n") {
		if strings.Contains(ansi.Strip(line), "▔") && ansi.Strip(line) == rule {
			return line
		}
	}
	return ""
}

func TestTheAccentRuleSitsUnderTheActiveLabelOnly(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	palette := model.palette()
	accentRun := lipgloss.NewStyle().Foreground(palette.accent).Render(strings.Repeat("▔", len([]rune(" Course "))))
	raw := rawTabRule(model)
	if !strings.Contains(raw, accentRun) {
		t.Fatalf("the accent run should cover exactly the Course label:\n%q", raw)
	}
	if strings.Count(raw, accentRun[:strings.Index(accentRun, "▔")]) != 1 {
		t.Fatalf("only one stretch should use the accent:\n%q", raw)
	}
	bar, rule := tabRows(model)
	if len([]rune(rule[:strings.Index(rule, "▔")])) != 2 || strings.Index(bar, " Course ") < 0 {
		t.Fatalf("the rule should start at the margin:\n%q\n%q", bar, rule)
	}
}

func TestTheRuleHasNoAccentWhenNoTabIsActive(t *testing.T) {
	model := settingsPageFixture(t)
	model.noColor = false
	accent := lipgloss.NewStyle().Foreground(model.palette().accent).Render("▔")
	if strings.Contains(rawTabRule(model), accent[:strings.Index(accent, "▔")]) {
		t.Fatal("with no active tab the rule should use only the base grey")
	}
	if _, rule := tabRows(model); !strings.Contains(rule, "▔") {
		t.Fatalf("the base rule should still be drawn:\n%q", rule)
	}
}

func TestWithoutColorOnlyTheActiveLabelIsUnderlined(t *testing.T) {
	model := interfaceFixture()
	bar, rule := tabRows(model)
	start := strings.Index(bar, "[Course]")
	if strings.Count(rule, "▔") != len([]rune(" Course ")) || start < 0 {
		t.Fatalf("only the active label should carry the rule:\n%q\n%q", bar, rule)
	}
	if strings.Index(rule, "▔")+len("▔")*0 < 0 || len([]rune(rule[:strings.Index(rule, "▔")])) != len([]rune(bar[:start])) {
		t.Fatalf("the stretch should start under the active tab:\n%q\n%q", bar, rule)
	}
}

func TestTheStretchReachFollowsTheShapeOfTheTab(t *testing.T) {
	model := withTerminalColors(interfaceFixture(), "#09090B", "#E4E4E7")
	for style, want := range map[string][2]int{"block": {7, 15}, "none": {7, 15}, "slanted": {9, 19}, "joined_arrows": {6, 15}, "joined_rounded": {6, 15}, "joined_slanted": {6, 15}} {
		model.application.settings.TabStyle = style
		start, end, found := model.activeTabSpan(style, "course")
		if !found || start != want[0] || end != want[1] {
			t.Fatalf("%s: got %d..%d, want %d..%d", style, start, end, want[0], want[1])
		}
	}
	start, _, _ := model.activeTabSpan("joined_arrows", "home")
	if start != 0 {
		t.Fatalf("the flat first tab has no left separator to include, got %d", start)
	}
}

func TestRoundedTabsDrawNoRule(t *testing.T) {
	model := withTerminalColors(interfaceFixture(), "#09090B", "#E4E4E7")
	model.application.settings.TabStyle = "rounded"
	if _, rule := tabRows(model); strings.Contains(rule, "▔") {
		t.Fatalf("rounded tabs are closed shapes and should have no rule under them: %q", rule)
	}
	model.noColor = true
	if _, rule := tabRows(model); !strings.Contains(rule, "▔") {
		t.Fatal("with no color the bar is bracketed and keeps its underline")
	}
}

func TestTheAccentStaysFarEnoughFromTheBaseRuleToTellApartByLightness(t *testing.T) {
	for _, dark := range []bool{true, false} {
		for _, brand := range highlightColors {
			model := interfaceFixture()
			model.noColor, model.dark = false, dark
			model.application.settings.Highlight = brand.id
			palette := model.palette()
			if ratio := contrastBetween(palette.accent, palette.rule); ratio < 3 {
				t.Fatalf("dark=%t %s: the accent and base rule differ by only %.2f, under 3", dark, brand.id, ratio)
			}
		}
	}
}

func TestTheRuleStartsWhereTheBarStarts(t *testing.T) {
	for style, prefix := range map[string]string{"joined_arrows": "▔", "block": "  ▔"} {
		model := withTerminalColors(interfaceFixture(), "#09090B", "#E4E4E7")
		model.application.settings.TabStyle = style
		_, rule := tabRows(model)
		if !strings.HasPrefix(rule, prefix) {
			t.Fatalf("%s: the rule should start where the bar starts: %q", style, rule)
		}
	}
}

func TestTheRuleReplacesTheBlankRowSoThePageKeepsItsHeight(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	if rows := strings.Count(model.View().Content, "\n") + 1; rows > model.height {
		t.Fatalf("the view is %d rows and the terminal has %d", rows, model.height)
	}
}

func TestTheNoneStyleLeavesUnderliningToTheRule(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	underline := lipgloss.NewStyle().Underline(true).Render("x")
	underline = underline[:strings.Index(underline, "x")]
	if strings.Contains(model.tabs("none", "course"), underline) {
		t.Fatal("the rule underlines the active tab, so its label should not also be underlined")
	}
}

func TestTheTerminalForegroundReportIsKept(t *testing.T) {
	model := interfaceFixture()
	updated, _ := model.Update(tea.ForegroundColorMsg{Color: lipgloss.Color("#E4E4E7")})
	if hexOf(updated.(interfaceModel).foreground) != "#E4E4E7" {
		t.Fatal("the reported foreground should be kept for the tab colors")
	}
}

func TestTheBaseRuleIsTheReportedColorsBlended(t *testing.T) {
	model := withTerminalColors(interfaceFixture(), "#000000", "#FFFFFF")
	base := lipgloss.NewStyle().Foreground(model.mixed(tabRuleShare)).Render("▔")
	if !strings.Contains(rawTabRule(model), base[:strings.Index(base, "▔")]) {
		t.Fatalf("the base rule should be the background blended toward the foreground:\n%q", rawTabRule(model))
	}
}
