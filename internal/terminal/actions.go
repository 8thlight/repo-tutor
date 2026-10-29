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
	key := lipgloss.NewStyle().Bold(true)
	label := lipgloss.NewStyle().Faint(true)
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
	return state.Question == "" && !reviewing(state) && (state.Quiz == nil || !state.Quiz.Covers(state.Topic))
}

const (
	sidebarWidth  = 30
	sidebarOffset = sidebarWidth + 3
)

func (model interfaceModel) courseActions(titled bool) [][2]string {
	state := model.application.state
	if questionOpen(state) {
		return [][2]string{{"Ctrl+N", "Skip question"}}
	}
	if inQuestionFlow(state) {
		return [][2]string{{"Ctrl+N", "Next question"}}
	}
	var actions [][2]string
	if state.Quiz != nil && state.Quiz.Exam && state.Quiz.Finished() {
		actions = [][2]string{{"Ctrl+P", "Exam results"}}
	} else if state.Quiz != nil && state.Quiz.Finished() {
		actions = [][2]string{{"Ctrl+F", fmt.Sprintf("More questions (%d)", model.quizLength())}}
	} else {
		actions = [][2]string{
			{"Ctrl+G", fmt.Sprintf("Quiz this topic (%d)", model.quizLength())},
			{"Ctrl+F", "More on this topic"},
		}
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

func (model interfaceModel) sidebarShortcutsTitle() string {
	if len(model.pageActions()) > 0 {
		return "ON THIS PAGE"
	}
	return "KEEP GOING"
}

func (model interfaceModel) sidebarShortcuts() string {
	if model.movingToAnotherTopic() {
		return lipgloss.NewStyle().Bold(true).Render("Ctrl+K") + "  " + lipgloss.NewStyle().Faint(true).Render("Choose a topic")
	}
	actions := model.pageActions()
	if len(actions) > 0 {
		actions = append(actions, [2]string{"Ctrl+K", "Choose a topic"}, [2]string{"Ctrl+S", "Follow a citation"})
		return model.renderSidebarRows(actions)
	}
	actions = model.courseActions(false)
	if state := model.application.state; state.Quiz != nil {
		actions = append(actions, [2]string{"Ctrl+B", "Question index"})
	}
	if state := model.application.state; model.leavable() || (inQuestionFlow(state) && model.page == "lesson") {
		actions = append(actions, [2]string{"Ctrl+L", "Leave " + model.flowName()})
	}
	actions = append(actions, [2]string{"Ctrl+K", "Choose a topic"}, [2]string{"Ctrl+S", "Follow a citation"})
	return model.renderSidebarRows(actions)
}

func (model interfaceModel) renderSidebarRows(actions [][2]string) string {
	key := lipgloss.NewStyle().Bold(true)
	label := lipgloss.NewStyle().Faint(true)
	rows := make([]string, len(actions))
	for index, action := range actions {
		rows[index] = key.Render(action[0]) + "  " + label.Render(action[1])
	}
	return strings.Join(rows, "\n")
}

func (model interfaceModel) quizCompleteLine() string {
	line := fmt.Sprintf("Ctrl+F More questions (%d)", model.quizLength())
	if title, found := model.nextTopicLabel(); found {
		line += " · Ctrl+N Next: " + title
	}
	return line + " · Ctrl+B results"
}
