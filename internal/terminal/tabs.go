package terminal

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type tabHint struct{ key, label string }

var (
	tabKey      = tabHint{"Tab", "Next"}
	helpKey     = tabHint{"F1", "Help"}
	aboutKey    = tabHint{"F2", "About"}
	tabHintSets = [][]tabHint{{tabKey, helpKey, aboutKey}, {helpKey, aboutKey}, {helpKey}}
)

func (model interfaceModel) cutoutColor(fallback string) color.Color {
	if model.background != nil {
		return model.background
	}
	return lipgloss.Color(fallback)
}

func (model interfaceModel) tabBar(width int) string {
	palette := model.palette()
	pages := []struct{ page, label string }{{"home", "Home"}, {"course", "Course"}, {"quiz", "Quiz"}, {"progress", "Progress"}}
	var tabs []string
	for _, page := range pages {
		active := model.activeTab() == page.page
		switch {
		case model.noColor && active:
			tabs = append(tabs, lipgloss.NewStyle().Underline(true).Render("["+page.label+"]"))
		case model.noColor:
			tabs = append(tabs, " "+page.label+" ")
		case active:
			tabs = append(tabs, lipgloss.NewStyle().Background(palette.accent).Foreground(palette.cutout).Bold(true).Render(" "+page.label+" "))
		default:
			tabs = append(tabs, lipgloss.NewStyle().Background(palette.muted).Foreground(palette.cutout).Render(" "+page.label+" "))
		}
	}
	bar := strings.Join(tabs, " ")
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
