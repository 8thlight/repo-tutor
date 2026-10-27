package terminal

import (
	"context"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func quizPrefetchFixture(t *testing.T) interfaceModel {
	model, _ := prefetchFixture(t)
	model.application.questionStash = &questionStash{}
	model.application.state.Quiz = &tutor.Quiz{Topic: model.application.state.Topic, Goal: 3, Active: 0, Questions: []tutor.QuizQuestion{{Text: "First?"}}}
	model.application.state.Question = "First?"
	model.prefetcher = func(context.Context, session.State, string) error { return nil }
	return model
}

func TestAnOpenQuizQuestionPrefetchesTheNextOneBeforeAnyLesson(t *testing.T) {
	model := quizPrefetchFixture(t)
	key, wanted := model.prefetchTarget()
	if !wanted || key != questionKey(model.application.state.Topic, 1) || !isQuestionKey(key) {
		t.Fatalf("a quiz with questions left should prefetch its next question, got %q %v", key, wanted)
	}
}

func TestTheLastQuestionStartsNoPrefetchOfAQuestion(t *testing.T) {
	model := quizPrefetchFixture(t)
	model.application.state.Quiz.Questions = append(model.application.state.Quiz.Questions, tutor.QuizQuestion{Text: "Second?"}, tutor.QuizQuestion{Text: "Third?"})
	if key, wanted := model.prefetchTarget(); wanted && isQuestionKey(key) {
		t.Fatalf("the last question has no next one: %q", key)
	}
}

func TestAStoredQuestionOrAFailureEndsTheWantToPrefetchIt(t *testing.T) {
	model := quizPrefetchFixture(t)
	key, _ := model.prefetchTarget()
	model.application.questionStash.put(key, tutor.Reply{Question: "Second?"})
	if next, wanted := model.prefetchTarget(); wanted && isQuestionKey(next) {
		t.Fatal("a question already stored should not be fetched again")
	}
	model.application.questionStash.clear()
	model.prefetchFailed = map[string]bool{key: true}
	if next, wanted := model.prefetchTarget(); wanted && isQuestionKey(next) {
		t.Fatal("a failed question prefetch should not retry for the same slot")
	}
}

func TestWithoutAQuizTheLessonPrefetchStillRuns(t *testing.T) {
	model, topic := prefetchFixture(t)
	model.prefetcher = func(context.Context, session.State, string) error { return nil }
	if key, wanted := model.prefetchTarget(); !wanted || key != topic {
		t.Fatalf("got %q %v", key, wanted)
	}
}

func TestThePrefetchRequestListsEarlierQuestionsSoTheyAreNotRepeated(t *testing.T) {
	quiz := &tutor.Quiz{Topic: "core", Goal: 5, Questions: []tutor.QuizQuestion{{Text: "Which package owns it?"}, {Text: "Where does it start?"}}}
	message := prefetchQuestionMessage(quiz)
	for _, expected := range []string{"Ask question 3 of 5", "1. Which package owns it?", "2. Where does it start?", "must not repeat"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("the prefetch request is missing %q:\n%s", expected, message)
		}
	}
}

func TestTheNextQuestionIsRecordedFromTheStashWithoutAskingTheAgent(t *testing.T) {
	model := quizPrefetchFixture(t)
	model.application.state.Review = true
	model.application.state.Question = ""
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: model.application.state.Topic, Action: "answer", Reply: tutor.Reply{Assessment: "correct", Feedback: "Yes."}})
	model.application.state.Quiz.Questions[0].Assessment = "correct"
	key, _ := nextQuestionKey(model.application.state)
	citations := model.application.state.Turns[0].Reply.Citations
	model.application.questionStash.put(key, tutor.Reply{Explanation: "Here is the second exercise.", Question: "Second?", Assessment: "none", Citations: citations})
	state, err := model.application.completeTurn(context.Background(), "quiz-next", nextQuestionMessage(model.application.state.Quiz), model.application.state.Topic)
	if err != nil {
		t.Fatalf("a stored question needs no agent: %v", err)
	}
	if state.Question != "Second?" || len(state.Quiz.Questions) != 2 || state.Quiz.Questions[1].Text != "Second?" || state.Quiz.Active != 1 {
		t.Fatalf("the stored question should become the open one: %+v", state.Quiz)
	}
	if model.application.questionStash.has(key) {
		t.Fatal("a used question must leave the stash")
	}
}

func TestAskingForTheNextQuestionWaitsForItsPrefetchAndOtherRequestsCancelIt(t *testing.T) {
	model := quizPrefetchFixture(t)
	key, _ := model.prefetchTarget()
	cancelled := false
	model.prefetching, model.prefetchTopic = true, key
	model.prefetchCancel = func() { cancelled = true }
	if !model.prefetchingTopic(turnIntent{"quiz-next", nextQuestionMessage(model.application.state.Quiz), model.application.state.Topic}) {
		t.Fatal("the next question's own request should wait for its prefetch")
	}
	if model.prefetchingTopic(turnIntent{"answer", "my answer", model.application.state.Topic}) {
		t.Fatal("an answer must not wait behind a prefetch")
	}
	stopped := model.stopPrefetch()
	if !cancelled || stopped.prefetching {
		t.Fatal("stopping should cancel the running prefetch")
	}
}

func TestTheFooterNamesWhichPrefetchIsRunning(t *testing.T) {
	model := quizPrefetchFixture(t)
	model.prefetchTopic = questionKey(model.application.state.Topic, 1)
	if model.prefetchLabel() != "Preparing next question" {
		t.Fatalf("got %q", model.prefetchLabel())
	}
	model.prefetchTopic = "core"
	if model.prefetchLabel() != "Preparing next lesson" {
		t.Fatalf("got %q", model.prefetchLabel())
	}
}

func TestAFinishedQuestionTurnPrefetchesTheNextQuestionAndShowsItsIndicator(t *testing.T) {
	model := quizPrefetchFixture(t)
	var prefetched []string
	model.prefetcher = func(_ context.Context, _ session.State, key string) error {
		prefetched = append(prefetched, key)
		return nil
	}
	model.busy = true
	finished, schedule := model.Update(turnFinished{state: model.application.state})
	model = finished.(interfaceModel)
	if !model.prefetchScheduled || schedule == nil {
		t.Fatal("a finished question should schedule the prefetch of the next one")
	}
	started, command := model.Update(prefetchDue{})
	model = started.(interfaceModel)
	if content := ansi.Strip(model.View().Content); !model.prefetching || !strings.Contains(content, "Preparing next question") {
		t.Fatal("a running question prefetch should say so in the footer")
	}
	result := runPrefetchCommand(t, command)
	if len(prefetched) != 1 || prefetched[0] != questionKey(model.application.state.Topic, 1) || result.err != nil {
		t.Fatalf("the next question should be prefetched once: %v %+v", prefetched, result)
	}
}

func TestAnAnswerCancelsTheQuestionPrefetchAndItIsScheduledAgainAfterFeedback(t *testing.T) {
	model := quizPrefetchFixture(t)
	model.prefetching, model.prefetchTopic = true, questionKey(model.application.state.Topic, 1)
	cancelled := false
	model.prefetchCancel = func() { cancelled = true }
	model.application.state.Question = "First?"
	model.input.SetValue("my answer")
	answering, _ := model.submit()
	defer answering.(interfaceModel).cancel()
	if !cancelled || answering.(interfaceModel).prefetching {
		t.Fatal("submitting an answer must cancel the running prefetch so feedback is not delayed")
	}
	feedback := answering.(interfaceModel)
	feedback.busy = true
	feedback.prefetchScheduled = false
	finished, schedule := feedback.Update(turnFinished{state: feedback.application.state})
	if !finished.(interfaceModel).prefetchScheduled || schedule == nil {
		t.Fatal("when feedback arrives the prefetch should start again")
	}
}
