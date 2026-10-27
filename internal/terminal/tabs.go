package terminal

import (
	"image/color"
	"math"
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

var joinedSeparators = map[string]string{
	settings.TabStyleJoinedSlanted: "\ue0bc",
	settings.TabStyleJoinedRounded: "\ue0b4",
	settings.TabStyleJoinedArrows:  "\ue0b0",
}

var tabPages = []struct{ page, label string }{{"home", "Home"}, {"course", "Course"}, {"quiz", "Quiz"}, {"progress", "Progress"}}

func (model interfaceModel) tabsBleed() bool {
	_, joined := joinedSeparators[model.application.settings.TabStyle]
	return joined && !model.noColor
}

func blend(from, to color.Color, share float64) color.Color {
	fromRed, fromGreen, fromBlue := channels(from)
	toRed, toGreen, toBlue := channels(to)
	mix := func(start, end float64) uint8 { return uint8(math.Round((start + (end-start)*share) * 255)) }
	return color.RGBA{mix(fromRed, toRed), mix(fromGreen, toGreen), mix(fromBlue, toBlue), 255}
}

func (model interfaceModel) joinedFills(active string) []color.Color {
	palette := model.palette()
	end := color.Color(lipgloss.Color("#E4E4E7"))
	if !model.dark {
		end = lipgloss.Color("#1F1F23")
	}
	fills := make([]color.Color, len(tabPages))
	for index, page := range tabPages {
		fills[index] = blend(palette.muted, end, float64(index)/float64(len(tabPages)-1))
		if page.page == active {
			fills[index] = palette.accent
		}
	}
	return fills
}

func (model interfaceModel) joinedCells(style, active string) []string {
	palette := model.palette()
	fills := model.joinedFills(active)
	cells := make([]string, len(tabPages))
	for index, page := range tabPages {
		separator := lipgloss.NewStyle().Foreground(fills[index])
		if index+1 < len(fills) {
			separator = separator.Background(fills[index+1])
		}
		cells[index] = lipgloss.NewStyle().Background(fills[index]).Foreground(palette.cutout).Bold(page.page == active).Render(" "+page.label+" ") + separator.Render(joinedSeparators[style])
	}
	return cells
}

func (model interfaceModel) tabs(style, active string) string {
	cells, joiner := model.tabCells(style, active)
	return strings.Join(cells, joiner)
}

func (model interfaceModel) tabCells(style, active string) ([]string, string) {
	if _, joined := joinedSeparators[style]; joined && !model.noColor {
		return model.joinedCells(style, active), ""
	}
	palette := model.palette()
	pages := tabPages
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
	return rendered, " "
}

func (model interfaceModel) activeTabSpan(style, active string) (int, int, bool) {
	cells, joiner := model.tabCells(style, active)
	_, drawsEdges := tabEdges[style]
	start := 0
	for index, cell := range cells {
		end := start + lipgloss.Width(cell)
		if tabPages[index].page == active {
			labelStart := start
			if drawsEdges && !model.noColor {
				labelStart++
			}
			return labelStart, labelStart + lipgloss.Width(" "+tabPages[index].label+" "), true
		}
		start = end + lipgloss.Width(joiner)
	}
	return 0, 0, false
}

func (model interfaceModel) tabRule(width int) string {
	palette := model.palette()
	start, end, found := model.activeTabSpan(model.application.settings.TabStyle, model.activeTab())
	base := func(count int) string {
		if model.noColor {
			return strings.Repeat(" ", count)
		}
		return lipgloss.NewStyle().Foreground(palette.rule).Render(strings.Repeat("▔", count))
	}
	if !found || start >= width {
		return base(width)
	}
	end = min(end, width)
	return base(start) + lipgloss.NewStyle().Foreground(palette.accent).Render(strings.Repeat("▔", end-start)) + base(width-end)
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
