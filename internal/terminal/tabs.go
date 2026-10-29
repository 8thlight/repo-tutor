package terminal

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/settings"
)

var tabHintSets = [][]keyHint{{{"Tab", "Next", false}}}

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

const (
	inactiveTabShare  = 0.15
	strongestTabShare = 0.35
	tabRuleShare      = 0.20
)

func (model interfaceModel) terminalColorsKnown() bool {
	return model.background != nil && model.foreground != nil
}

func (model interfaceModel) mixed(share float64) color.Color {
	return blend(model.background, model.foreground, share)
}

func (model interfaceModel) drawnTabStyle(style string) string {
	_, joined := joinedSeparators[style]
	_, edged := tabEdges[style]
	if (joined || edged) && !model.terminalColorsKnown() {
		return settings.TabStyleBlock
	}
	return style
}

func (model interfaceModel) tabsBleed() bool {
	_, joined := joinedSeparators[model.drawnTabStyle(model.application.settings.TabStyle)]
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
	fills := make([]color.Color, len(tabPages))
	for index, page := range tabPages {
		share := inactiveTabShare + (strongestTabShare-inactiveTabShare)*float64(index)/float64(len(tabPages)-1)
		fills[index] = model.mixed(share)
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
		text := model.foreground
		if page.page == active {
			text = palette.cutout
		}
		cells[index] = lipgloss.NewStyle().Background(fills[index]).Foreground(text).Bold(page.page == active).Render(" "+page.label+" ") + separator.Render(joinedSeparators[style])
	}
	return cells
}

func (model interfaceModel) tabs(style, active string) string {
	cells, joiner := model.tabCells(style, active)
	return strings.Join(cells, joiner)
}

func (model interfaceModel) tabCells(style, active string) ([]string, string) {
	style = model.drawnTabStyle(style)
	if _, joined := joinedSeparators[style]; joined && !model.noColor {
		return model.joinedCells(style, active), ""
	}
	palette := model.palette()
	pages := tabPages
	rendered := make([]string, len(pages))
	for index, page := range pages {
		current := active == page.page
		label := " " + page.label + " "
		switch {
		case model.noColor && current:
			rendered[index] = lipgloss.NewStyle().Underline(true).Render("[" + page.label + "]")
		case model.noColor:
			rendered[index] = label
		case style == settings.TabStyleNone && current:
			rendered[index] = lipgloss.NewStyle().Foreground(palette.accent).Bold(true).Render(label)
		case style == settings.TabStyleNone:
			rendered[index] = lipgloss.NewStyle().Foreground(palette.muted).Render(label)
		case !model.terminalColorsKnown() && current:
			rendered[index] = lipgloss.NewStyle().Foreground(palette.accent).Reverse(true).Bold(true).Render(label)
		case !model.terminalColorsKnown():
			rendered[index] = lipgloss.NewStyle().Reverse(true).Render(label)
		default:
			fill, text := model.mixed(inactiveTabShare), model.foreground
			if current {
				fill, text = palette.accent, palette.cutout
			}
			block := lipgloss.NewStyle().Background(fill).Foreground(text).Bold(current).Render(label)
			if edges, found := tabEdges[style]; found {
				edge := lipgloss.NewStyle().Foreground(fill)
				block = edge.Render(edges[0]) + block + edge.Render(edges[1])
			}
			rendered[index] = block
		}
	}
	return rendered, " "
}

type ruleReach int

const (
	ruleLabel ruleReach = iota
	ruleWholeCell
	ruleLeftSeparatorAndLabel
	ruleNone
)

var ruleReaches = map[string]ruleReach{
	settings.TabStyleSlanted:       ruleWholeCell,
	settings.TabStyleRounded:       ruleNone,
	settings.TabStyleJoinedSlanted: ruleLeftSeparatorAndLabel,
	settings.TabStyleJoinedRounded: ruleLeftSeparatorAndLabel,
	settings.TabStyleJoinedArrows:  ruleLeftSeparatorAndLabel,
}

func (model interfaceModel) ruleReach() ruleReach {
	if model.noColor {
		return ruleLabel
	}
	return ruleReaches[model.drawnTabStyle(model.application.settings.TabStyle)]
}

func (model interfaceModel) activeTabSpan(style, active string) (int, int, bool) {
	cells, joiner := model.tabCells(style, active)
	start := 0
	for index, cell := range cells {
		end := start + lipgloss.Width(cell)
		if tabPages[index].page == active {
			labelEnd := start + lipgloss.Width(" "+tabPages[index].label+" ")
			switch model.ruleReach() {
			case ruleWholeCell:
				return start, end, true
			case ruleLeftSeparatorAndLabel:
				return max(0, start-1), labelEnd, true
			}
			return start, labelEnd, true
		}
		start = end + lipgloss.Width(joiner)
	}
	return 0, 0, false
}

func (model interfaceModel) tabRule(width int) string {
	if model.ruleReach() == ruleNone {
		return strings.Repeat(" ", width)
	}
	palette := model.palette()
	start, end, found := model.activeTabSpan(model.drawnTabStyle(model.application.settings.TabStyle), model.activeTab())
	base := func(count int) string {
		if model.noColor {
			return strings.Repeat(" ", count)
		}
		rule := palette.rule
		if model.terminalColorsKnown() {
			rule = model.mixed(tabRuleShare)
		}
		return lipgloss.NewStyle().Foreground(rule).Render(strings.Repeat("▔", count))
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
		hints := model.renderKeyHints(set, "  ")
		if lipgloss.Width(bar)+2+lipgloss.Width(hints) <= width {
			return bar + strings.Repeat(" ", width-lipgloss.Width(bar)-lipgloss.Width(hints)) + hints
		}
	}
	return ansi.Truncate(bar, width, "…")
}
