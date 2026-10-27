package tutor

type QuizQuestion struct {
	Text       string `json:"text"`
	Assessment string `json:"assessment,omitempty"`
	Turn       int    `json:"turn,omitempty"`
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

const DefaultQuizLength = 5

func RecordQuiz(previous *Quiz, topic, action, pending string, reply Reply, length int) *Quiz {
	if length < 1 {
		length = DefaultQuizLength
	}
	if action == "quiz" {
		return &Quiz{Topic: topic, Goal: length, Questions: []QuizQuestion{{Text: reply.Question}}}
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
		quiz.Goal += length
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

func RecordAnswerTurn(before, after *Quiz, turn int) *Quiz {
	if before == nil || after == nil {
		return after
	}
	recorded := *after
	recorded.Questions = append([]QuizQuestion(nil), after.Questions...)
	for slot := range recorded.Questions {
		answered := recorded.Questions[slot].Assessment != "" && recorded.Questions[slot].Assessment != "skipped"
		open := slot >= len(before.Questions) || before.Questions[slot].Assessment == "" || before.Questions[slot].Assessment == "skipped"
		if answered && open {
			recorded.Questions[slot].Turn = turn
		}
	}
	return &recorded
}

func ShiftQuizTurns(quiz *Quiz, trimmed int) *Quiz {
	if quiz == nil {
		return nil
	}
	shifted := *quiz
	shifted.Questions = append([]QuizQuestion(nil), quiz.Questions...)
	for slot := range shifted.Questions {
		shifted.Questions[slot].Turn = max(0, shifted.Questions[slot].Turn-trimmed)
	}
	return &shifted
}
