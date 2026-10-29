package terminal

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"repo-tutor/internal/settings"
)

const (
	ansiYellow  = 3
	noHighlight = -1
)

type highlightColor struct {
	id, label string
	index     int
}

var highlightColors = []highlightColor{
	{settings.HighlightNone, "None", noHighlight},
	{"red", "Red", 1},
	{settings.HighlightGreen, "Green", 2},
	{"yellow", "Yellow", ansiYellow},
	{"blue", "Blue", 4},
	{"magenta", "Magenta", 5},
	{"cyan", "Cyan", 6},
}

func highlightByID(id string) highlightColor {
	for _, candidate := range highlightColors {
		if candidate.id == id {
			return candidate
		}
	}
	return highlightByID(settings.HighlightGreen)
}

func (model interfaceModel) highlight(id string) color.Color {
	choice := highlightByID(id)
	if choice.index == noHighlight {
		return lipgloss.NoColor{}
	}
	return lipgloss.ANSIColor(choice.index)
}
