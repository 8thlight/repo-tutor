package terminal

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

const questionPrefix = "question:"

func questionKey(topic string, slot int) string {
	return fmt.Sprintf("%s%s#%d", questionPrefix, topic, slot)
}

func isQuestionKey(key string) bool {
	return strings.HasPrefix(key, questionPrefix)
}

func nextQuestionKey(state session.State) (string, bool) {
	quiz := state.Quiz
	if quiz == nil || !quiz.Covers(state.Topic) || quiz.Finished() || len(quiz.Questions) >= quiz.Goal {
		return "", false
	}
	return questionKey(quiz.NextTopic(state.Topic), len(quiz.Questions)), true
}

type questionStash struct {
	lock    sync.Mutex
	replies map[string]tutor.Reply
}

func (stash *questionStash) put(key string, reply tutor.Reply) {
	if stash == nil {
		return
	}
	stash.lock.Lock()
	defer stash.lock.Unlock()
	if stash.replies == nil {
		stash.replies = map[string]tutor.Reply{}
	}
	stash.replies[key] = reply
}

func (stash *questionStash) take(key string) (tutor.Reply, bool) {
	if stash == nil {
		return tutor.Reply{}, false
	}
	stash.lock.Lock()
	defer stash.lock.Unlock()
	reply, found := stash.replies[key]
	delete(stash.replies, key)
	return reply, found
}

func (stash *questionStash) has(key string) bool {
	if stash == nil {
		return false
	}
	stash.lock.Lock()
	defer stash.lock.Unlock()
	_, found := stash.replies[key]
	return found
}

func (stash *questionStash) clear() {
	if stash == nil {
		return
	}
	stash.lock.Lock()
	defer stash.lock.Unlock()
	stash.replies = nil
}

func nextQuestionMessage(quiz *tutor.Quiz) string {
	if quiz.Exam {
		return fmt.Sprintf("Ask question %d of %d for a course exam, about this topic only. Choose an exercise unlike the earlier ones.", len(quiz.Questions)+1, quiz.Goal)
	}
	return fmt.Sprintf("Ask question %d of %d for this quiz. Choose a different exercise from previous questions.", len(quiz.Questions)+1, quiz.Goal)
}

func prefetchQuestionMessage(quiz *tutor.Quiz) string {
	message := nextQuestionMessage(quiz) + " Earlier questions, which you must not repeat:"
	for index, question := range quiz.Questions {
		message += fmt.Sprintf("\n%d. %s", index+1, tutor.Plain(question.Text))
	}
	return message
}

func (application application) prefetchQuestion(ctx context.Context) error {
	state := application.state
	key, wanted := nextQuestionKey(state)
	if !wanted {
		return fmt.Errorf("the quiz has no next question")
	}
	topic := state.Quiz.NextTopic(state.Topic)
	request := tutor.Request{Topic: topic, Action: "quiz", Message: prefetchQuestionMessage(state.Quiz), Goal: state.Goal, Lesson: lastLesson(state.Turns, topic)}
	if len(state.Turns) > 0 && topic == state.Topic {
		previous := state.Turns[len(state.Turns)-1].Reply
		request.Previous = &previous
	}
	reply, _, err := application.ask(ctx, request, "")
	if err != nil {
		return err
	}
	application.questionStash.put(key, reply)
	return nil
}

func (model interfaceModel) prefetchLabel() string {
	if isQuestionKey(model.prefetchTopic) {
		return "Preparing next question"
	}
	return "Preparing next lesson"
}
