package terminal

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type keyHint struct {
	key, label string
	warning    bool
}

var (
	topicsHint   = keyHint{"Ctrl+K", "topics", false}
	sourcesHint  = keyHint{"Ctrl+S", "sources", false}
	helpHint     = keyHint{"F1", "Help", false}
	aboutHint    = keyHint{"F3", "About", false}
	settingsHint = keyHint{"F2", "Settings", false}
	escapeHint   = keyHint{"Esc", "back or stop", false}
	quitHint     = keyHint{"Ctrl+Q", "quit", false}
)

func (model interfaceModel) renderKeyHints(hints []keyHint, separator string) string {
	palette := model.palette()
	key := lipgloss.NewStyle().Foreground(palette.foreground).Bold(true)
	label := lipgloss.NewStyle().Foreground(palette.muted)
	alert := lipgloss.NewStyle().Foreground(palette.warning).Bold(true)
	rendered := make([]string, len(hints))
	for index, hint := range hints {
		rendered[index] = key.Render(hint.key) + " " + label.Render(hint.label)
		if hint.warning {
			rendered[index] = alert.Render(hint.key + " " + hint.label)
		}
	}
	return strings.Join(rendered, label.Render(separator))
}

func (model interfaceModel) sidebarVisible() bool {
	return model.width >= 108
}

func (model interfaceModel) footerHintSets() [][]keyHint {
	if model.sidebarVisible() {
		return [][]keyHint{
			{helpHint, settingsHint, aboutHint, escapeHint, quitHint},
			{helpHint, settingsHint, aboutHint, quitHint},
			{helpHint, quitHint},
		}
	}
	navigation := []keyHint{topicsHint, sourcesHint}
	if model.movingToAnotherTopic() {
		navigation = []keyHint{topicsHint}
	}
	return [][]keyHint{
		append(append([]keyHint(nil), navigation...), helpHint, settingsHint, aboutHint, escapeHint, quitHint),
		append(append([]keyHint(nil), navigation...), helpHint, settingsHint, aboutHint, quitHint),
		append(append([]keyHint(nil), navigation...), helpHint, quitHint),
		{helpHint, quitHint},
	}
}

func (model interfaceModel) footerKeys(width int) string {
	for _, set := range model.footerHintSets() {
		if line := model.renderKeyHints(set, " · "); ansi.StringWidth(line) <= width {
			return line
		}
	}
	return ansi.Truncate(model.renderKeyHints([]keyHint{helpHint, quitHint}, " · "), width, "…")
}

func (model interfaceModel) typingHints() []keyHint {
	state := model.application.state
	switch {
	case model.askingGoal() && strings.TrimSpace(model.input.Value()) != "":
		return []keyHint{{"Enter", "saves your goal", false}}
	case model.askingGoal():
		return []keyHint{{"Enter", "keeps the default", false}}
	case model.busy:
		return []keyHint{{"Esc", "stops it", false}}
	case model.page == "lesson" && model.history < len(state.Turns)-1:
		return nil
	case reviewing(state) && state.Quiz != nil && state.Quiz.Completed() >= state.Quiz.Goal:
		return []keyHint{{"Ctrl+F", "more questions", false}}
	case reviewing(state):
		return []keyHint{{"Enter or Ctrl+N", "continue", false}}
	}
	if questionOpen(state) {
		if model.skipArmed {
			return []keyHint{{"Ctrl+N", "again to skip", true}}
		}
		return []keyHint{{"Enter", "send", false}, {"Ctrl+N", "skip", false}}
	}
	return []keyHint{{"Enter", "send", false}}
}

func (model interfaceModel) framedInput(title string, hints []keyHint, tint color.Color) string {
	width := max(12, model.bodyWidth())
	inner := width - 2
	border := lipgloss.NewStyle().Foreground(tint)
	caption := lipgloss.NewStyle().Foreground(model.palette().muted)
	title = ansi.Truncate(title, max(1, inner-4), "…")
	top := border.Render("╭─ ") + caption.Render(title) + border.Render(" "+strings.Repeat("─", max(0, inner-3-ansi.StringWidth(title)))+"╮")
	bottom := border.Render("╰" + strings.Repeat("─", inner) + "╯")
	if len(hints) > 0 {
		rendered := model.renderKeyHints(hints, " · ")
		if dashes := inner - ansi.StringWidth(rendered) - 3; dashes >= 2 {
			bottom = border.Render("╰"+strings.Repeat("─", dashes)+" ") + rendered + border.Render(" ─╯")
		}
	}
	body := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, true, false, true).BorderForeground(tint).Width(width).Render(model.input.View())
	return top + "\n" + body + "\n" + bottom
}
