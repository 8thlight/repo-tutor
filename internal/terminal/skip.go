package terminal

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

var skipWindow = 5 * time.Second

const (
	skipPrompt = "Press Ctrl+N again to skip this question."
	skipText   = "I would like to skip this question."
)

type skipExpired struct{ generation int }

func (model interfaceModel) skipOrArm() (tea.Model, tea.Cmd) {
	if model.skipArmed {
		model.skipArmed = false
		model.notice = ""
		return model.begin(turnIntent{"answer", skipText, model.application.state.Topic})
	}
	model.skipArmed = true
	model.skipGeneration++
	model.notice = skipPrompt
	generation := model.skipGeneration
	return model, tea.Tick(skipWindow, func(time.Time) tea.Msg { return skipExpired{generation} })
}

func (model interfaceModel) disarmSkip() interfaceModel {
	if model.skipArmed && model.notice == skipPrompt {
		model.notice = ""
	}
	model.skipArmed = false
	return model
}
