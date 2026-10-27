package terminal

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/tutor"
)

const (
	goalLimit         = 500
	goalPendingNotice = "Choose your goal first. Enter keeps the default."
	goalQuestion      = "What do you want to learn?"
	goalComposerLabel = goalQuestion + "\nEnter keeps the default · type your own goal and press Enter"
	goalSavedNotice   = "Goal saved"
)

var goalEditingKeys = map[string]bool{
	"space": true, "backspace": true, "delete": true, "left": true, "right": true,
	"up": true, "down": true, "home": true, "end": true,
	"ctrl+a": true, "ctrl+e": true, "ctrl+w": true,
}

var goalPassThroughKeys = map[string]bool{
	"ctrl+q": true, "ctrl+c": true, "f1": true, "f2": true, "f3": true,
}

func (model interfaceModel) askingGoal() bool {
	return model.goalPending && model.page == "home"
}

func (model interfaceModel) goalKey(message tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	pressed := message.String()
	switch {
	case goalPassThroughKeys[pressed]:
		return model, nil, false
	case pressed == "enter":
		next, command := model.submitGoal()
		return next, command, true
	case message.Text != "" || goalEditingKeys[pressed]:
		focus := model.input.Focus()
		var command tea.Cmd
		model.input, command = model.input.Update(message)
		return model, tea.Batch(focus, command), true
	}
	model.notice = goalPendingNotice
	return model, nil, true
}

func (model interfaceModel) submitGoal() (tea.Model, tea.Cmd) {
	goal := strings.TrimSpace(model.input.Value())
	if goal == "" {
		goal = model.application.state.Goal
	}
	if utf8.RuneCountInString(goal) > goalLimit {
		model.notice = "Keep your goal under 500 characters. Your text is still here."
		return model, nil
	}
	state := model.application.state
	state.Goal, state.GoalChosen = goal, true
	if model.application.store != nil {
		if err := model.application.store.Save(state); err != nil {
			model.notice = "Could not save: " + tutor.Plain(err.Error())
			return model, nil
		}
	}
	model.application.state = state
	model.goalPending = false
	model.input.Reset()
	model.input.Placeholder = lessonPlaceholder
	model.notice = goalSavedNotice
	model.rebuildContent()
	return model, func() tea.Msg { return beginFirstLesson{} }
}
