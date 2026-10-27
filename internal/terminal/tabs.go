package terminal

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/settings"
)

type tabHint struct{ key, label string }

var (
	tabKey      = tabHint{"Tab", "Next"}
	helpKey     = tabHint{"F1", "Help"}
	aboutKey    = tabHint{"F2", "About"}
	settingsKey = tabHint{"F3", "Settings"}
	tabHintSets = [][]tabHint{{tabKey, helpKey, aboutKey, settingsKey}, {helpKey, aboutKey, settingsKey}, {helpKey, aboutKey}, {helpKey}}
)

func (model interfaceModel) cutoutColor(fallback string) color.Color {
	if model.background != nil {
		return model.background
	}
	return lipgloss.Color(fallback)
}

var tabEdges = map[string][2]string{
	settings.TabStyleSlanted: {"\ue0ba", "\ue0b8"},
	settings.TabStyleRounded: {"\ue0b6", "\ue0b4"},
}

func (model interfaceModel) tabs(style, active string) string {
	palette := model.palette()
	pages := []struct{ page, label string }{{"home", "Home"}, {"course", "Course"}, {"quiz", "Quiz"}, {"progress", "Progress"}}
	rendered := make([]string, len(pages))
	for index, page := range pages {
		current := active == page.page
		label := " " + page.label + " "
		fill := palette.muted
		if current {
			fill = palette.accent
		}
		switch {
		case model.noColor && current:
			rendered[index] = lipgloss.NewStyle().Underline(true).Render("[" + page.label + "]")
		case model.noColor:
			rendered[index] = label
		case style == settings.TabStyleNone && current:
			rendered[index] = lipgloss.NewStyle().Foreground(palette.accent).Bold(true).Underline(true).Render(label)
		case style == settings.TabStyleNone:
			rendered[index] = lipgloss.NewStyle().Foreground(palette.muted).Render(label)
		default:
			block := lipgloss.NewStyle().Background(fill).Foreground(palette.cutout).Bold(current).Render(label)
			if edges, found := tabEdges[style]; found {
				edge := lipgloss.NewStyle().Foreground(fill)
				block = edge.Render(edges[0]) + block + edge.Render(edges[1])
			}
			rendered[index] = block
		}
	}
	return strings.Join(rendered, " ")
}

func (model interfaceModel) tabBar(width int) string {
	bar := model.tabs(model.application.settings.TabStyle, model.activeTab())
	for _, set := range tabHintSets {
		hints := model.renderTabHints(set)
		if lipgloss.Width(bar)+2+lipgloss.Width(hints) <= width {
			return bar + strings.Repeat(" ", width-lipgloss.Width(bar)-lipgloss.Width(hints)) + hints
		}
	}
	return ansi.Truncate(bar, width, "…")
}

func (model interfaceModel) renderTabHints(hints []tabHint) string {
	palette := model.palette()
	key := lipgloss.NewStyle().Foreground(palette.foreground).Bold(true)
	label := lipgloss.NewStyle().Foreground(palette.muted)
	rendered := make([]string, len(hints))
	for index, hint := range hints {
		rendered[index] = key.Render(hint.key) + " " + label.Render(hint.label)
	}
	return strings.Join(rendered, "  ")
}
