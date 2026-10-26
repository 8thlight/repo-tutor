package tutor

type QuizQuestion struct {
	Text       string `json:"text"`
	Assessment string `json:"assessment,omitempty"`
}

type Quiz struct {
	Topic     string         `json:"topic"`
	Goal      int            `json:"goal"`
	Active    int            `json:"active"`
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
	if action == "lesson" || action == "probe" || (previous != nil && previous.Topic != topic) {
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
		quiz.Active = len(quiz.Questions) - 1
	}
	if action == "answer" && quiz.Active < len(quiz.Questions) {
		index := quiz.Active
		open := quiz.Questions[index].Assessment == "" || quiz.Questions[index].Assessment == "skipped"
		if quiz.Questions[index].Text == pending && open && reply.Assessment == "none" && reply.Question != "" {
			quiz.Questions[index].Text = reply.Question
		}
		if quiz.Questions[index].Text == pending && open && (reply.Assessment != "none" || reply.Question == "") {
			assessment := reply.Assessment
			if assessment == "none" {
				assessment = "skipped"
			}
			quiz.Questions[index].Assessment = assessment
		}
	}
	return &quiz
}

func Reopen(quiz *Quiz, slot int) (*Quiz, bool) {
	if quiz == nil || slot < 0 || slot >= len(quiz.Questions) {
		return nil, false
	}
	if assessment := quiz.Questions[slot].Assessment; assessment != "" && assessment != "skipped" {
		return nil, false
	}
	reopened := *quiz
	reopened.Questions = append([]QuizQuestion(nil), quiz.Questions...)
	reopened.Active = slot
	return &reopened, true
}
