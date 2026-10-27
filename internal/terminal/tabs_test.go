package terminal

import (
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

func TestTabHintsShortenAsTheLineNarrows(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for width, want := range map[int]string{80: "Tab Next  F1 Help  F2 About  F3 Settings", 70: "F1 Help  F2 About  F3 Settings", 56: "F1 Help  F2 About", 44: "F1 Help", 40: ""} {
		bar := ansi.Strip(model.tabBar(width))
		if ansi.StringWidth(bar) > width {
			t.Fatalf("width %d: the bar is %d columns wide", width, ansi.StringWidth(bar))
		}
		if want != "" && !strings.HasSuffix(bar, want) {
			t.Fatalf("width %d should end with %q:\n%q", width, want, bar)
		}
		if want == "" && strings.Contains(bar, "F1") {
			t.Fatalf("width %d should drop the hints:\n%q", width, bar)
		}
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

func TestJoinedTabsTouchAndEndWithTheirSeparator(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for style, glyph := range map[string]string{"joined_slanted": "", "joined_rounded": "", "joined_arrows": ""} {
		want := " Home " + glyph + " Course " + glyph + " Quiz " + glyph + " Progress " + glyph
		if got := ansi.Strip(model.tabs(style, "course")); got != want {
			t.Fatalf("%s: got %q want %q", style, got, want)
		}
	}
}

func TestEveryBoundaryInAJoinedBarShowsAColorChange(t *testing.T) {
	for _, dark := range []bool{true, false} {
		model := interfaceFixture()
		model.noColor, model.dark = false, dark
		for _, page := range []string{"home", "course", "quiz", "progress", "settings"} {
			fills := model.joinedFills(page)
			for index := 1; index < len(fills); index++ {
				if hexOf(fills[index-1]) == hexOf(fills[index]) {
					t.Fatalf("dark=%t active=%s: tabs %d and %d share %s, so the separator would vanish", dark, page, index-1, index, hexOf(fills[index]))
				}
			}
		}
	}
}

func TestEveryJoinedFillKeepsReadableTabText(t *testing.T) {
	for _, background := range terminalBackgrounds {
		model := interfaceFixture()
		model.noColor = false
		model.background = lipgloss.Color(background)
		model.dark = relativeLuminance(model.background) <= 0.179
		for _, page := range []string{"home", "course", "quiz", "progress"} {
			for index, fill := range model.joinedFills(page) {
				if ratio := contrastBetween(fill, model.background); ratio < 4.5 {
					t.Fatalf("background %s active %s tab %d (%s) has contrast %.2f", background, page, index, hexOf(fill), ratio)
				}
			}
		}
	}
}

func TestJoinedBarReachesTheTerminalEdgeAndOtherStylesKeepTheMargin(t *testing.T) {
	row := func(style string) string {
		model := interfaceFixture()
		model.noColor = false
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
