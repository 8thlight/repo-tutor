package terminal

import (
	"math"
	"testing"

	"charm.land/lipgloss/v2"
)

var terminalThemes = []struct{ name, background, foreground string }{
	{"Alacritty default", "#181818", "#D8D8D8"},
	{"8th Light Black", "#000000", "#EBEBEB"},
	{"8th Light Charcoal", "#141414", "#EBEBEB"},
	{"8th Light Dark Gray", "#3C3C3C", "#F4F4F4"},
	{"8th Light Mid Gray", "#808080", "#000000"},
	{"8th Light Light Gray", "#DEDEDE", "#141414"},
	{"8th Light Lighter Gray", "#EBEBEB", "#141414"},
	{"8th Light Lightest Gray", "#F4F4F4", "#141414"},
	{"8th Light White", "#FFFFFF", "#141414"},
	{"Solarized dark", "#002B36", "#839496"},
	{"Solarized light", "#FDF6E3", "#657B83"},
	{"Dracula", "#282A36", "#F8F8F2"},
	{"Gruvbox dark", "#282828", "#EBDBB2"},
	{"Gruvbox light", "#FBF1C7", "#3C3836"},
	{"Nord", "#2E3440", "#D8DEE9"},
	{"Tokyo Night", "#1A1B26", "#A9B1D6"},
	{"One Light", "#FAFAFA", "#383A42"},
	{"Catppuccin Mocha", "#1E1E2E", "#CDD6F4"},
	{"Catppuccin Latte", "#EFF1F5", "#4C4F69"},
}

func TestJoinedFillsKeepTheTextReadableOnEveryTheme(t *testing.T) {
	for _, theme := range terminalThemes {
		model := withTerminalColors(interfaceFixture(), theme.background, theme.foreground)
		floor := math.Min(readableTabText, keptThemeContrast*contrastBetween(model.background, model.foreground))
		fills := model.joinedFills("settings")
		for index, fill := range fills {
			if ratio := contrastBetween(fill, model.foreground); ratio < floor-0.05 {
				t.Fatalf("%s tab %d (%s) has text contrast %.2f, under the floor %.2f", theme.name, index, hexOf(fill), ratio, floor)
			}
			if index > 0 && hexOf(fills[index-1]) == hexOf(fill) {
				t.Fatalf("%s tabs %d and %d share %s, so the separator would vanish", theme.name, index-1, index, hexOf(fill))
			}
		}
	}
}

func TestThemesWithRoomKeepTheFullGradient(t *testing.T) {
	keeps := map[string]bool{"Alacritty default": true, "8th Light Black": true, "8th Light Charcoal": true, "8th Light Light Gray": true, "8th Light Lighter Gray": true, "8th Light Lightest Gray": true, "8th Light White": true, "Gruvbox light": true, "One Light": true, "Solarized light": true}
	for _, theme := range terminalThemes {
		got := strongestShare(lipgloss.Color(theme.background), lipgloss.Color(theme.foreground))
		if keeps[theme.name] && got != strongestTabShare {
			t.Fatalf("%s should keep the full gradient, got a ceiling of %.3f", theme.name, got)
		}
		if !keeps[theme.name] && got >= strongestTabShare {
			t.Fatalf("%s should get a gentler gradient, got a ceiling of %.3f", theme.name, got)
		}
	}
}

func TestLowContrastThemesGetAGentlerGradient(t *testing.T) {
	for _, name := range []string{"8th Light Dark Gray", "8th Light Mid Gray"} {
		for _, theme := range terminalThemes {
			if theme.name != name {
				continue
			}
			if got := strongestShare(lipgloss.Color(theme.background), lipgloss.Color(theme.foreground)); got >= strongestTabShare || got < inactiveTabShare+smallestTabStep {
				t.Fatalf("%s should lower the ceiling but keep a visible step, got %.2f", name, got)
			}
		}
	}
}
