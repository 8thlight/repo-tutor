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
	if turn.Action != "answer" {
		return false
	}
	if turn.Reply.Assessment != "none" || turn.Reply.Question == "" {
		return true
	}
	if len(state.Turns) > 1 {
		previous := state.Turns[len(state.Turns)-2]
		return previous.Topic == turn.Topic && previous.Reply.Question != "" && previous.Reply.Question != turn.Reply.Question
	}
	return false
}

func continueIntent(state session.State, extend bool) (turnIntent, error) {
	quiz := state.Quiz
	if extend {
		if quiz == nil || quiz.Completed() < quiz.Goal {
			return turnIntent{}, fmt.Errorf("finish the current quiz before extending it")
		}
		return turnIntent{"quiz-extend", "Extend this quiz with five more questions. Ask the next question without repeating earlier exercises.", state.Topic}, nil
	}
	if !reviewing(state) {
		return turnIntent{}, fmt.Errorf("answer the current question before continuing")
	}
	if quiz != nil {
		if quiz.Completed() >= quiz.Goal {
			return turnIntent{}, fmt.Errorf("quiz complete: %d / %d answered. Use /extend for five more or choose another topic", quiz.Completed(), quiz.Goal)
		}
		return turnIntent{"quiz-next", fmt.Sprintf("Ask question %d of %d for this quiz. Choose a different exercise from previous questions.", len(quiz.Questions)+1, quiz.Goal), state.Topic}, nil
	}
	return turnIntent{"lesson", "Continue with the next small idea in this topic and one new question.", state.Topic}, nil
}

func quizText(quiz *tutor.Quiz) string {
	if quiz == nil {
		return "Start a five-question quiz with Ctrl+G or /quiz.\n\nEach answer is a step toward the goal. Use /extend after finishing for five more."
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
		text.WriteString("Goal reached. /extend adds five more questions; /quiz starts a new quiz.\n")
	}
	return text.String()
}

func (model interfaceModel) continueQuiz(extend bool) (tea.Model, tea.Cmd) {
	intent, err := continueIntent(model.application.state, extend)
	if err != nil {
		model.notice = err.Error()
		return model, nil
	}
	return model.begin(intent)
}
