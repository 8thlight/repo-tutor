package terminal

import (
	"fmt"
	"image/color"
	"slices"
	"testing"

	"charm.land/lipgloss/v2"

	"repo-tutor/internal/settings"
)

var terminalBackgrounds = []string{"#000000", "#09090B", "#1E1E1E", "#282C34", "#FFFFFF", "#FAFAFA", "#F4F4F4", "#EEEEEE"}

func hexOf(shade color.Color) string {
	red, green, blue, _ := shade.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", red>>8, green>>8, blue>>8)
}

func TestEveryBrandColorIsReadableOnEveryBackground(t *testing.T) {
	for _, brand := range highlightColors {
		for _, background := range terminalBackgrounds {
			shifted := readableOn(lipgloss.Color(brand.hex), lipgloss.Color(background))
			if ratio := contrastBetween(shifted, lipgloss.Color(background)); ratio < 4.5 {
				t.Fatalf("%s on %s has contrast %.2f (%s)", brand.id, background, ratio, hexOf(shifted))
			}
		}
	}
}

func TestABrandColorThatAlreadyReadsIsLeftAlone(t *testing.T) {
	for id, background := range map[string]string{"green": "#09090B", "yellow": "#1E1E1E", "orange": "#09090B", "dark_blue": "#FFFFFF"} {
		brand := highlightByID(id)
		if got := hexOf(readableOn(lipgloss.Color(brand.hex), lipgloss.Color(background))); got != brand.hex {
			t.Fatalf("%s on %s should stay %s, got %s", id, background, brand.hex, got)
		}
	}
}

func TestAShiftKeepsTheHueAndMovesTowardTheContrast(t *testing.T) {
	green := lipgloss.Color(highlightByID("green").hex)
	darker := readableOn(green, lipgloss.Color("#FAFAFA"))
	baseHue, _, baseLightness := toHSL(green)
	hue, _, lightness := toHSL(darker)
	if lightness >= baseLightness || hue < baseHue-2 || hue > baseHue+2 {
		t.Fatalf("green on a light terminal should darken and keep its hue: %s", hexOf(darker))
	}
	lighter := readableOn(lipgloss.Color(highlightByID("blue").hex), lipgloss.Color("#09090B"))
	_, _, blueLightness := toHSL(lipgloss.Color(highlightByID("blue").hex))
	if _, _, shifted := toHSL(lighter); shifted <= blueLightness {
		t.Fatalf("blue on a dark terminal should lighten: %s", hexOf(lighter))
	}
}

func TestPaletteUsesTheChosenHighlight(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	model.dark = true
	if hexOf(model.palette().accent) != "#93D500" {
		t.Fatalf("green is the default and reads on a dark terminal, got %s", hexOf(model.palette().accent))
	}
	model.application.settings.Highlight = "purple"
	if hexOf(model.palette().accent) == "#93D500" {
		t.Fatal("choosing purple should change the accent")
	}
	model.dark = false
	model.application.settings.Highlight = "green"
	if hexOf(model.palette().accent) == "#93D500" {
		t.Fatal("green should darken on a light terminal")
	}
}

func TestEverySettingsHighlightHasABrandColor(t *testing.T) {
	var ids []string
	for _, brand := range highlightColors {
		ids = append(ids, brand.id)
	}
	if !slices.Equal(ids, settings.Highlights) {
		t.Fatalf("the brand table and the settings values must list the same colors in the same order: %v vs %v", ids, settings.Highlights)
	}
}
