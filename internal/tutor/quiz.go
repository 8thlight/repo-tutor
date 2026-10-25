package tutor

type QuizQuestion struct {
	Text       string `json:"text"`
	Assessment string `json:"assessment,omitempty"`
}

type Quiz struct {
	Topic     string         `json:"topic"`
	Goal      int            `json:"goal"`
	Questions []QuizQuestion `json:"questions"`
}

func (quiz Quiz) Completed() int {
	count := 0
	for _, question := range quiz.Questions {
		if question.Assessment != "" {
			count++
		}
	}
	return count
}

func RecordQuiz(previous *Quiz, topic, action, pending string, reply Reply) *Quiz {
	if action == "quiz" {
		return &Quiz{Topic: topic, Goal: 5, Questions: []QuizQuestion{{Text: reply.Question}}}
	}
	if action == "lesson" || (previous != nil && previous.Topic != topic) {
		return nil
	}
	if previous == nil {
		return nil
	}
	quiz := *previous
	quiz.Questions = append([]QuizQuestion(nil), previous.Questions...)
	if action == "quiz-extend" {
		quiz.Goal += 5
	}
	if action == "quiz-next" || action == "quiz-extend" {
		quiz.Questions = append(quiz.Questions, QuizQuestion{Text: reply.Question})
	}
	if action == "answer" && len(quiz.Questions) > 0 {
		index := len(quiz.Questions) - 1
		if quiz.Questions[index].Text == pending && quiz.Questions[index].Assessment == "" && (reply.Assessment != "none" || reply.Question == "") {
			assessment := reply.Assessment
			if assessment == "none" {
				assessment = "skipped"
			}
			quiz.Questions[index].Assessment = assessment
		}
	}
	return &quiz
}
