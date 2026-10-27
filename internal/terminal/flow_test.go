package terminal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func readingFixture() interfaceModel {
	model := interfaceFixture()
	model.application.state.Question, model.application.state.Review, model.application.state.Quiz = "", false, nil
	model.application.state.Turns[0].Reply.Question = ""
	return model
}

func citedReply(question string) tutor.Reply {
	return tutor.Reply{Explanation: "It starts in main.go:2.", Question: question, Assessment: "none", Citations: []repository.Location{{Path: "main.go", Line: 2}}}
}

func TestALessonReplyLeavesNoPendingQuestion(t *testing.T) {
	model := readingFixture()
	updated := model.application.recordReply("lesson", "Teach this.", tutor.Overview, "", citedReply("Why?"))
	if updated.Question != "" || reviewing(updated) {
		t.Fatalf("a lesson should leave nothing pending: %q", updated.Question)
	}
	if updated.Turns[len(updated.Turns)-1].Reply.Question != "" {
		t.Fatal("the saved lesson turn should carry no question")
	}
}

func TestAFreeQuestionOutsideAQuizGetsNoQuestionBack(t *testing.T) {
	model := readingFixture()
	updated := model.application.recordReply("question", "How does it start?", tutor.Overview, "", citedReply("Does that help?"))
	if updated.Question != "" {
		t.Fatalf("answering a free question should not open a pending question: %q", updated.Question)
	}
}

func TestAClarificationInsideAQuizKeepsThePendingQuestion(t *testing.T) {
	model := readingFixture()
	model.application.state.Question = "Which package owns the calculation?"
	updated := model.application.recordReply("question", "What do you mean?", tutor.Overview, "", citedReply("Which package owns the calculation, in other words?"))
	if updated.Question == "" {
		t.Fatal("a clarification during a quiz question should keep the question open")
	}
}

func TestMoreOnThisTopicNeedsNoReview(t *testing.T) {
	model := readingFixture()
	intent, err := continueIntent(model.application.state, model.application.course, false)
	if err != nil || intent.action != "lesson" || intent.topic != model.application.state.Topic {
		t.Fatalf("Ctrl+F after a lesson should ask for more on the topic: %+v %v", intent, err)
	}
}

func TestMoreIsRefusedWhileAQuestionIsPending(t *testing.T) {
	model := readingFixture()
	model.application.state.Question = "Which package owns the calculation?"
	if _, err := continueIntent(model.application.state, model.application.course, false); err == nil || !strings.Contains(err.Error(), "answer or skip") {
		t.Fatalf("a pending question should block Continue: %v", err)
	}
}

func TestAnAnswerIsStillReviewedBeforeTheNextQuizQuestion(t *testing.T) {
	model := readingFixture()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "partial"}}}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "answer", Reply: tutor.Reply{Assessment: "partial", Feedback: "Close."}})
	model.application.state.Review = true
	intent, err := continueIntent(model.application.state, model.application.course, false)
	if err != nil || intent.action != "quiz-next" {
		t.Fatalf("after feedback, Ctrl+F should ask the next quiz question: %+v %v", intent, err)
	}
}

func TestTheLessonComposerIsAQuestionBox(t *testing.T) {
	composer := ansi.Strip(readingFixture().composer())
	if !strings.Contains(composer, "Ask about this topic") || strings.Contains(composer, "NEXT QUESTION") {
		t.Fatalf("a lesson should offer a question box:\n%s", composer)
	}
}
