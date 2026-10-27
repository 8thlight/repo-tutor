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

func (model interfaceModel) actionLine() string {
	if !model.readingActions() {
		return ""
	}
	actions := [][2]string{
		{"Ctrl+G", fmt.Sprintf("Quiz this topic (%d)", model.quizLength())},
		{"Ctrl+F", "More on this topic"},
	}
	if title, found := model.nextTopicLabel(); found {
		actions = append(actions, [2]string{"Ctrl+N", "Next: " + title})
	}
	return model.styledActions(actions)
}

func (model interfaceModel) quizCompleteLine() string {
	line := fmt.Sprintf("/extend adds %d · Ctrl+B results", model.quizLength())
	if title, found := model.nextTopicLabel(); found {
		line = "Ctrl+N Next: " + title + " · " + line
	}
	return line
}
