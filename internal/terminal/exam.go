package terminal

import (
	"fmt"

	"repo-tutor/internal/tutor"
)

const examMessage = "Ask the first question of a course exam, about this topic only."

func (model interfaceModel) quizBlockTitle() string {
	if quiz := model.application.state.Quiz; quiz != nil && quiz.Exam {
		return "EXAM"
	}
	return "QUIZ GOAL"
}

func (model interfaceModel) examPlanLength() int {
	return len(tutor.ExamPlan(model.application.course, model.application.state.Progress, model.application.settings.ExamLength))
}

func (model interfaceModel) examCompleteText() string {
	state := model.application.state
	text := "EXAM COMPLETE"
	if state.LastExam != nil {
		percent := state.LastExam.Percent()
		text = fmt.Sprintf("EXAM COMPLETE · %d%% · %s", percent, tutor.Band(percent))
	}
	line := "Ctrl+P Progress for each topic · Ctrl+B results"
	if title, found := model.nextTopicLabel(); found {
		line += " · Ctrl+N Next: " + title
	}
	return text + "\n" + line
}
