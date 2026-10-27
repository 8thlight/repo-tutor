package terminal

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/tutor"
)

func (model interfaceModel) quizLength() int {
	if length := model.application.settings.QuizLength; length > 0 {
		return length
	}
	return tutor.DefaultQuizLength
}

func (model interfaceModel) nextTopicLabel() (string, bool) {
	next, found := tutor.NextUnheld(model.application.course, model.application.state.Progress, model.application.state.Topic)
	return tutor.Plain(next.Title), found
}

func (model interfaceModel) styledActions(actions [][2]string) string {
	palette := model.palette()
	key := lipgloss.NewStyle().Foreground(palette.foreground).Bold(true)
	label := lipgloss.NewStyle().Foreground(palette.muted)
	rendered := make([]string, len(actions))
	for index, action := range actions {
		rendered[index] = key.Render(action[0]) + " " + label.Render(action[1])
	}
	return ansi.Truncate(strings.Join(rendered, label.Render("  ·  ")), model.bodyWidth(), "…")
}

func (model interfaceModel) readingActions() bool {
	state := model.application.state
	if model.page != "lesson" || model.busy || model.goalPending || len(state.Turns) == 0 {
		return false
	}
	if model.history != len(state.Turns)-1 {
		return false
	}
	return state.Question == "" && !reviewing(state) && (state.Quiz == nil || state.Quiz.Topic != state.Topic)
}

const (
	sidebarWidth  = 30
	sidebarOffset = sidebarWidth + 3
)

func (model interfaceModel) courseActions(titled bool) [][2]string {
	actions := [][2]string{
		{"Ctrl+G", fmt.Sprintf("Quiz this topic (%d)", model.quizLength())},
		{"Ctrl+F", "More on this topic"},
	}
	if title, found := model.nextTopicLabel(); found {
		next := "Next topic"
		if titled {
			next = "Next: " + title
		}
		actions = append(actions, [2]string{"Ctrl+N", next})
	}
	return actions
}

func (model interfaceModel) actionLine() string {
	if !model.readingActions() {
		return ""
	}
	return model.styledActions(model.courseActions(true))
}

func (model interfaceModel) sidebarShortcuts() string {
	if model.movingToAnotherTopic() {
		return lipgloss.NewStyle().Foreground(model.palette().foreground).Bold(true).Render("Ctrl+K") + "  " + lipgloss.NewStyle().Foreground(model.palette().muted).Render("Choose a topic")
	}
	actions := model.courseActions(false)
	if quiz := model.application.state.Quiz; quiz != nil && quiz.Completed() < quiz.Goal {
		actions[1][1] = "Next question"
	}
	actions = append(actions, [2]string{"Ctrl+K", "Choose a topic"}, [2]string{"Ctrl+S", "Follow a citation"})
	palette := model.palette()
	key := lipgloss.NewStyle().Foreground(palette.foreground).Bold(true)
	label := lipgloss.NewStyle().Foreground(palette.muted)
	rows := make([]string, len(actions))
	for index, action := range actions {
		rows[index] = key.Render(action[0]) + "  " + label.Render(action[1])
	}
	return strings.Join(rows, "\n")
}

func (model interfaceModel) quizCompleteLine() string {
	line := fmt.Sprintf("/extend adds %d · Ctrl+B results", model.quizLength())
	if title, found := model.nextTopicLabel(); found {
		line = "Ctrl+N Next: " + title + " · " + line
	}
	return line
}
