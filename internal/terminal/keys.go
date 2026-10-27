package terminal

import (
	"image/color"
	"strings"
	"unicode"

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

func (model interfaceModel) pageActions() [][2]string {
	switch {
	case model.page == "source":
		actions := [][2]string{{"Enter or Esc", "Back"}, {"T", "Teach this file"}}
		if isMarkdown(model.sourcePath) {
			actions = append(actions, [2]string{"v", "Rendered or source"})
		}
		return actions
	case model.isPicker():
		actions := [][2]string{{"↑↓", "Move"}, {"/", "Search"}, {"Enter", "Open"}}
		if model.page == "quiz" {
			actions = append(actions, [2]string{"A", "Show answer"})
		}
		return actions
	}
	return nil
}

func (model interfaceModel) pageHints() []keyHint {
	if model.sidebarVisible() {
		return nil
	}
	var hints []keyHint
	for _, action := range model.pageActions() {
		label := []rune(action[1])
		label[0] = unicode.ToLower(label[0])
		hints = append(hints, keyHint{action[0], string(label), false})
	}
	return hints
}

func (model interfaceModel) footerKeys(width int) string {
	page := model.pageHints()
	pageGoesBack := false
	for _, hint := range page {
		pageGoesBack = pageGoesBack || strings.Contains(hint.key, "Esc")
	}
	for _, set := range model.footerHintSets() {
		var merged []keyHint
		merged = append(merged, page...)
		for _, hint := range set {
			if pageGoesBack && hint == escapeHint {
				continue
			}
			merged = append(merged, hint)
		}
		if line := model.renderKeyHints(merged, " · "); ansi.StringWidth(line) <= width {
			return line
		}
	}
	for count := len(page); count > 0; count-- {
		shortened := append(append([]keyHint(nil), page[:count]...), quitHint)
		if line := model.renderKeyHints(shortened, " · "); ansi.StringWidth(line) <= width {
			return line
		}
	}
	return ansi.Truncate(model.renderKeyHints([]keyHint{helpHint, quitHint}, " · "), width, "…")
}

func (model interfaceModel) typingHints() []keyHint {
	state := model.application.state
	switch {
	case model.leaveArmed:
		return []keyHint{{"Ctrl+L", "again to leave", true}}
	case model.askingGoal() && strings.TrimSpace(model.input.Value()) != "":
		return []keyHint{{"Enter", "saves your goal", false}}
	case model.askingGoal():
		return []keyHint{{"Enter", "keeps the default", false}}
	case model.busy:
		return []keyHint{{"Esc", "stops it", false}}
	case model.page == "lesson" && model.history < len(state.Turns)-1:
		return nil
	case reviewing(state) && state.Quiz != nil && state.Quiz.Exam && state.Quiz.Completed() >= state.Quiz.Goal:
		return []keyHint{{"Ctrl+P", "see results", false}}
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
