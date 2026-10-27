package terminal

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type keyHint struct{ key, label string }

var (
	topicsHint   = keyHint{"Ctrl+K", "topics"}
	sourcesHint  = keyHint{"Ctrl+S", "sources"}
	helpHint     = keyHint{"F1", "Help"}
	aboutHint    = keyHint{"F2", "About"}
	settingsHint = keyHint{"F3", "Settings"}
	escapeHint   = keyHint{"Esc", "back or stop"}
	quitHint     = keyHint{"Ctrl+Q", "quit"}
)

func (model interfaceModel) renderKeyHints(hints []keyHint, separator string) string {
	palette := model.palette()
	key := lipgloss.NewStyle().Foreground(palette.foreground).Bold(true)
	label := lipgloss.NewStyle().Foreground(palette.muted)
	rendered := make([]string, len(hints))
	for index, hint := range hints {
		rendered[index] = key.Render(hint.key) + " " + label.Render(hint.label)
	}
	return strings.Join(rendered, label.Render(separator))
}

func (model interfaceModel) sidebarVisible() bool {
	return model.width >= 108
}

func (model interfaceModel) footerHintSets() [][]keyHint {
	if model.sidebarVisible() {
		return [][]keyHint{
			{helpHint, aboutHint, settingsHint, escapeHint, quitHint},
			{helpHint, aboutHint, settingsHint, quitHint},
			{helpHint, quitHint},
		}
	}
	navigation := []keyHint{topicsHint, sourcesHint}
	if model.movingToAnotherTopic() {
		navigation = []keyHint{topicsHint}
	}
	return [][]keyHint{
		append(append([]keyHint(nil), navigation...), helpHint, aboutHint, settingsHint, escapeHint, quitHint),
		append(append([]keyHint(nil), navigation...), helpHint, aboutHint, settingsHint, quitHint),
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
	case model.askingGoal():
		return nil
	case model.busy:
		return []keyHint{{"Esc", "stops it"}}
	case model.page == "lesson" && model.history < len(state.Turns)-1:
		return nil
	case reviewing(state) && state.Quiz != nil && state.Quiz.Completed() >= state.Quiz.Goal:
		return []keyHint{{"Ctrl+F", "more questions"}}
	case reviewing(state):
		return []keyHint{{"Enter or Ctrl+F", "continue"}}
	}
	return []keyHint{{"Enter", "send"}, {"Alt+Enter", "new line"}}
}

func (model interfaceModel) composerDivider() string {
	hints := model.typingHints()
	if len(hints) == 0 {
		return model.sectionDivider("")
	}
	rendered := model.renderKeyHints(hints, " · ")
	room := model.bodyWidth() - ansi.StringWidth(rendered) - 4
	if room < 2 {
		return model.sectionDivider("")
	}
	rule := lipgloss.NewStyle().Foreground(model.palette().muted)
	return rule.Render(strings.Repeat("─", room)+" ") + rendered + rule.Render(" ──")
}
