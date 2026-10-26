package terminal

import (
	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func resumeTopic(state session.State, topic string) (session.State, bool) {
	source := -1
	for index := len(state.Turns) - 1; index >= 0; index-- {
		if state.Turns[index].Topic == topic {
			source = index
			break
		}
	}
	if source < 0 {
		return state, false
	}
	if state.Topic == topic {
		return state, true
	}
	last := state.Turns[source]
	state.Topic = topic
	state.Question = last.Reply.Question
	state.Review = last.Action == "answer" && (last.Reply.Assessment != "none" || last.Reply.Question == "")
	if state.Quiz != nil && state.Quiz.Topic != topic {
		state.Quiz = nil
	}
	state.Probe = nil
	state.Turns = append(append([]session.Turn(nil), state.Turns...), last)
	if trimmed := len(state.Turns) - 200; trimmed > 0 {
		state.Turns = state.Turns[trimmed:]
		state.Quiz = tutor.ShiftQuizTurns(state.Quiz, trimmed)
	}
	return state, true
}

func (model interfaceModel) openTopic(topic string) (tea.Model, tea.Cmd) {
	resumed, found := resumeTopic(model.application.state, topic)
	if !found {
		return model.begin(turnIntent{"lesson", "Teach the next small idea in this topic.", topic})
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
