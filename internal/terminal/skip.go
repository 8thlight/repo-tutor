package terminal

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

var skipWindow = 5 * time.Second

const (
	skipPrompt = "Press Ctrl+N again to skip this question."

	leavePrompt = "Press Ctrl+L again to leave this quiz. It will not be scored."
)

type skipExpired struct{ generation int }

func (model interfaceModel) skipOrArm() (tea.Model, tea.Cmd) {
	if model.skipArmed {
		return model.skipQuestion()
	}
	model.skipArmed = true
	model.skipGeneration++
	model.notice = skipPrompt
	generation := model.skipGeneration
	return model, tea.Tick(skipWindow, func(time.Time) tea.Msg { return skipExpired{generation} })
}

func (model interfaceModel) disarmSkip() interfaceModel {
	if model.notice == skipPrompt || model.notice == leavePrompt {
		model.notice = ""
	}
	model.skipArmed, model.leaveArmed = false, false
	return model
}

func (model interfaceModel) leavable() bool {
	state := model.application.state
	return model.page == "lesson" && !model.busy && (inQuestionFlow(state) || questionOpen(state))
}

func (model interfaceModel) leaveOrArm() (tea.Model, tea.Cmd) {
	if model.leaveArmed {
		return model.leaveQuiz()
	}
	model.skipArmed, model.leaveArmed = false, true
	model.skipGeneration++
	model.notice = leavePrompt
	generation := model.skipGeneration
	return model, tea.Tick(skipWindow, func(time.Time) tea.Msg { return skipExpired{generation} })
}

func (model interfaceModel) leaveQuiz() (tea.Model, tea.Cmd) {
	model = model.disarmSkip()
	left, found := leaveQuestions(model.application.state)
	if !found {
		model.notice = "There is no lesson to return to for this topic yet."
		return model, nil
	}
	if model.application.store != nil {
		if err := model.application.store.Save(left); err != nil {
			model.notice = "Could not save: " + tutor.Plain(err.Error())
			return model, nil
		}
	}
	model.application.state = left
	model.history = len(left.Turns) - 1
	model.notice = "Left the quiz"
	return model.learn()
}

func (application application) skipOpenQuestion() (session.State, bool) {
	state := application.state
	if quiz, skipped := tutor.SkipQuizQuestion(state.Quiz); skipped && state.Quiz != nil && state.Quiz.Topic == state.Topic {
		state.Quiz = quiz
		state.Progress = tutor.RecordQuizScore(state.Progress, state.Topic, quiz)
	} else if probe, progress, skipped := tutor.SkipProbeQuestion(state.Probe, state.Progress, state.Topic); skipped {
		state.Probe, state.Progress = probe, progress
	} else {
		return state, false
	}
	state.Question, state.Review = "", true
	return state, true
}

func (model interfaceModel) skipQuestion() (tea.Model, tea.Cmd) {
	model = model.disarmSkip()
	skipped, found := model.application.skipOpenQuestion()
	if !found {
		model.notice = "There is no open question to skip."
		return model, nil
	}
	if model.application.store != nil {
		if err := model.application.store.Save(skipped); err != nil {
			model.notice = "Could not save: " + tutor.Plain(err.Error())
			return model, nil
		}
	}
	model.application.state = skipped
	if quizFinished(skipped) {
		model.notice = "Skipped. That was the last question."
		model.rebuildContent()
		return model, nil
	}
	model.notice = "Skipped"
	return model.continueQuiz(false)
}
