package terminal

import (
	"slices"
	"testing"

	"charm.land/lipgloss/v2"

	"repo-tutor/internal/settings"
)

func TestEverySettingsHighlightHasATerminalColor(t *testing.T) {
	var ids []string
	for _, choice := range highlightColors {
		ids = append(ids, choice.id)
	}
	if !slices.Equal(ids, settings.Highlights) {
		t.Fatalf("the highlight table and the settings values must list the same colors in the same order: %v vs %v", ids, settings.Highlights)
	}
}

func TestPaletteUsesTheTerminalColorOfTheChosenHighlight(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for id, want := range map[string]any{"green": lipgloss.ANSIColor(2), "red": lipgloss.ANSIColor(1), "cyan": lipgloss.ANSIColor(6), "none": lipgloss.NoColor{}} {
		model.application.settings.Highlight = id
		if got := model.palette().accent; got != want {
			t.Fatalf("%s: accent is %v, want %v", id, got, want)
		}
	}
}

func TestEveryHighlightIsOneOfTheSixteenTerminalColors(t *testing.T) {
	for _, choice := range highlightColors {
		if choice.index != noHighlight && (choice.index < 1 || choice.index > 6) {
			t.Fatalf("%s uses color %d, outside the six hues", choice.id, choice.index)
		}
	}
}

func TestInlineCodeFollowsTheChosenHighlight(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for id, want := range map[string]string{"green": "2", "red": "1", "magenta": "5"} {
		model.application.settings.Highlight = id
		if got := *model.markdownStyle().Code.Color; got != want {
			t.Fatalf("%s: inline code is %s, want %s", id, got, want)
		}
	}
	model.noColor = true
	model.application.settings.Highlight = "red"
	if model.markdownStyle().Code.Color != nil && *model.markdownStyle().Code.Color == "1" {
		t.Fatal("with no color the ASCII style should keep its own code color")
	}
}
