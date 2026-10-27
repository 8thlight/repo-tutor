package terminal

import (
	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func readingAction(action string) bool {
	return action == "lesson" || action == "question"
}

func latestReading(state session.State, topic string) (int, bool) {
	for index := len(state.Turns) - 1; index >= 0; index-- {
		if state.Turns[index].Topic == topic && readingAction(state.Turns[index].Action) {
			return index, true
		}
	}
	return -1, false
}

func showAgain(state session.State, source int) session.State {
	copied := state.Turns[source]
	copied.Resumed = true
	state.Topic = copied.Topic
	state.Question = copied.Reply.Question
	state.Review = false
	state.Turns = append(append([]session.Turn(nil), state.Turns...), copied)
	if trimmed := len(state.Turns) - 200; trimmed > 0 {
		state.Turns = state.Turns[trimmed:]
		state.Quiz = tutor.ShiftQuizTurns(state.Quiz, trimmed)
	}
	return state
}

func resumeTopic(state session.State, topic string) (session.State, bool) {
	source, found := latestReading(state, topic)
	if !found {
		return state, false
	}
	if state.Topic == topic {
		return state, true
	}
	if state.Quiz != nil && state.Quiz.Topic != topic {
		state.Quiz = nil
	}
	state.Probe = nil
	return showAgain(state, source), true
}

func leaveQuestions(state session.State) (session.State, bool) {
	source, found := latestReading(state, state.Topic)
	if !found {
		return state, false
	}
	state.Quiz, state.Probe = nil, nil
	return showAgain(state, source), true
}

func (model interfaceModel) openTopic(topic string) (tea.Model, tea.Cmd) {
	resumed, found := resumeTopic(model.application.state, topic)
	if !found {
		return model.openFirstLesson(topic)
	}
	if model.application.store != nil {
		if err := model.application.store.Save(resumed); err != nil {
			model.notice = "Could not save: " + tutor.Plain(err.Error())
			return model, nil
		}
	}
	model.application.state = resumed
	model.history = len(resumed.Turns) - 1
	model.notice = "Resumed this topic. Answer, or continue when you are ready."
	return model.learn()
}

func (model interfaceModel) openFirstLesson(topic string) (tea.Model, tea.Cmd) {
	lesson, cached := model.application.cachedLesson(topic)
	if !cached || !lesson.Fresh(model.application.snapshot.Fingerprint) {
		return model.begin(turnIntent{"lesson", firstLessonMessage, topic})
	}
	opened, err := model.application.openWithLesson(lesson, firstLessonMessage)
	if err != nil {
		model.notice = tutor.Plain(err.Error())
		return model, nil
	}
	model.application.state = opened
	model.history = len(opened.Turns) - 1
	model.notice = "Opened a lesson prepared earlier"
	return model.learn()
}
