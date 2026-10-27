package terminal

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func reviewing(state session.State) bool {
	if state.Review {
		return true
	}
	if len(state.Turns) == 0 {
		return false
	}
	turn := state.Turns[len(state.Turns)-1]
	return turn.Action == "answer" && (turn.Reply.Assessment != "none" || turn.Reply.Question == "")
}

func probeIntent(state session.State, course []tutor.Topic) (turnIntent, error) {
	topics := tutor.ProbeTopics(course, state.Progress)
	if len(topics) == 0 {
		return turnIntent{}, fmt.Errorf("every topic is already marked as known; use /topic or /quiz")
	}
	return turnIntent{"probe", "Ask one short question that tests whether I already know this topic.", topics[0]}, nil
}

func continueIntent(state session.State, course []tutor.Topic, extend bool) (turnIntent, error) {
	quiz := state.Quiz
	if extend {
		if quiz == nil || quiz.Completed() < quiz.Goal {
			return turnIntent{}, fmt.Errorf("finish the current quiz before extending it")
		}
		return turnIntent{"quiz-extend", "Extend this quiz with more questions. Ask the next question without repeating earlier exercises.", state.Topic}, nil
	}
	if state.Question != "" && !reviewing(state) {
		return turnIntent{}, fmt.Errorf("answer or skip the current question before continuing")
	}
	if !reviewing(state) {
		return turnIntent{"lesson", "Continue with the next small idea in this topic.", state.Topic}, nil
	}
	if probe := state.Probe; probe != nil {
		if next, more := probe.Next(); more {
			return turnIntent{"probe", "Ask one short question that tests whether I already know this topic.", next}, nil
		}
		if topic, found := tutor.FirstUnheld(course, state.Progress); found {
			return turnIntent{"lesson", "Teach this topic and connect it to its foundations.", topic.ID}, nil
		}
		return turnIntent{}, fmt.Errorf("every topic is marked as known; use /topic or /quiz")
	}
	if quiz != nil {
		if quiz.Completed() >= quiz.Goal {
			return turnIntent{}, fmt.Errorf("quiz complete: %d / %d answered. Use /extend for more or choose another topic", quiz.Completed(), quiz.Goal)
		}
		return turnIntent{"quiz-next", fmt.Sprintf("Ask question %d of %d for this quiz. Choose a different exercise from previous questions.", len(quiz.Questions)+1, quiz.Goal), state.Topic}, nil
	}
	return turnIntent{"lesson", "Continue with the next small idea in this topic and one new question.", state.Topic}, nil
}

func quizText(quiz *tutor.Quiz) string {
	if quiz == nil {
		return "Start a quiz with Ctrl+G or /quiz.\n\nEach answer is a step toward the goal. Use /extend after finishing for more."
	}
	var text strings.Builder
	fmt.Fprintf(&text, "QUIZ · %d / %d answered\n\n", quiz.Completed(), quiz.Goal)
	for index := 0; index < quiz.Goal; index++ {
		status, question := "upcoming", "Generated when you continue"
		if index < len(quiz.Questions) {
			current := quiz.Questions[index]
			question, status = tutor.Plain(current.Text), current.Assessment
			if status == "" {
				status = "current"
			}
		}
		fmt.Fprintf(&text, "%d. [%s] %s\n\n", index+1, status, question)
	}
	if quiz.Completed() >= quiz.Goal {
		text.WriteString("Goal reached. /extend adds more questions; /quiz starts a new quiz.\n")
	}
	return text.String()
}

func inQuestionFlow(state session.State) bool {
	return (state.Quiz != nil && !state.Quiz.Finished()) || state.Probe != nil
}

func quizFinished(state session.State) bool {
	quiz := state.Quiz
	return quiz != nil && reviewing(state) && quiz.Completed() >= quiz.Goal
}

func (model interfaceModel) continueQuiz(extend bool) (tea.Model, tea.Cmd) {
	extend = extend || quizFinished(model.application.state)
	intent, err := continueIntent(model.application.state, model.application.course, extend)
	if err != nil {
		model.notice = err.Error()
		return model, nil
	}
	return model.begin(intent)
}

func quizSlotItems(state session.State) []browserItem {
	quiz := state.Quiz
	if quiz == nil {
		return []browserItem{{"No quiz yet", "Start a quiz with Ctrl+G or /quiz.", "", -1}}
	}
	items := make([]browserItem, 0, quiz.Goal)
	for slot := 0; slot < max(quiz.Goal, len(quiz.Questions)); slot++ {
		title, description := fmt.Sprintf("%d  Generated when you get there", slot+1), "Later"
		if slot < len(quiz.Questions) {
			question := quiz.Questions[slot]
			title = fmt.Sprintf("%d  %s", slot+1, tutor.Plain(question.Text))
			switch question.Assessment {
			case "":
				description = "Current · Enter returns to this question"
			case "skipped":
				description = "Skipped · Enter answers it again"
			default:
				description = "Answered " + question.Assessment + " · Enter shows the feedback"
			}
		} else if slot == len(quiz.Questions) && reviewing(state) && quiz.Completed() < quiz.Goal {
			description = "Next · Enter generates this question"
		}
		items = append(items, browserItem{title, description, "", slot})
	}
	return items
}

func reopenQuestion(state session.State, slot int) (session.State, bool) {
	reopened, ok := tutor.Reopen(state.Quiz, slot)
	if !ok {
		return state, false
	}
	text := reopened.Questions[slot].Text
	state.Quiz, state.Question, state.Review = reopened, text, false
	reply := tutor.Reply{Assessment: "none", Explanation: fmt.Sprintf("Returning to question %d of %d.", slot+1, reopened.Goal), Question: text}
	state.Turns = append(append([]session.Turn(nil), state.Turns...), session.Turn{Topic: state.Topic, Action: "quiz-retry", Reply: reply})
	return state, true
}

func (model interfaceModel) selectQuizSlot(slot int) (tea.Model, tea.Cmd) {
	state := model.application.state
	quiz := state.Quiz
	switch {
	case quiz == nil || slot < 0:
		model.notice = "Start a quiz with Ctrl+G or /quiz."
		return model, nil
	case slot == len(quiz.Questions) && reviewing(state):
		return model.continueQuiz(false)
	case slot >= len(quiz.Questions):
		model.notice = "This question is generated when you reach it."
		return model, nil
	}
	reopened, ok := reopenQuestion(state, slot)
	if !ok {
		return model.viewAnswer(quiz.Questions[slot].Turn)
	}
	if model.application.store != nil {
		if err := model.application.store.Save(reopened); err != nil {
			model.notice = "Could not save: " + tutor.Plain(err.Error())
			return model, nil
		}
	}
	model.application.state = reopened
	model.history = len(reopened.Turns) - 1
	return model.learn()
}

func (model interfaceModel) viewAnswer(turn int) (tea.Model, tea.Cmd) {
	turns := model.application.state.Turns
	if turn <= 0 || turn >= len(turns) || turns[turn].Action != "answer" {
		model.notice = "The feedback for this answer is no longer saved."
		return model, nil
	}
	model.page, model.history = "lesson", turn
	model.rebuildContent()
	model.viewport.GotoTop()
	return model, nil
}
