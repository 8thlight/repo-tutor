package terminal

import (
	tea "charm.land/bubbletea/v2"
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

func TestWaitingTextNamesWhatIsHappening(t *testing.T) {
	for action, want := range map[string]string{
		"lesson": "Preparing your lesson", "quiz": "Preparing your question", "quiz-next": "Preparing your question",
		"probe": "Preparing your question", "answer": "Reviewing your answer", "question": "Thinking about your question",
	} {
		model := interfaceFixture()
		started, _ := model.begin(turnIntent{action, "go", tutor.Overview})
		model = started.(interfaceModel)
		content := ansi.Strip(model.View().Content)
		model.cancel()
		if !strings.Contains(content, want) {
			t.Fatalf("%s should say %q:\n%s", action, want, content)
		}
	}
}

func TestAfterALessonTheActionLineOffersQuizMoreAndNext(t *testing.T) {
	model := readingFixture()
	title, found := model.nextTopicLabel()
	if !found {
		t.Fatal("the fixture course should have a next topic")
	}
	line := ansi.Strip(model.actionLine())
	for _, expected := range []string{"Ctrl+G Quiz this topic (5)", "Ctrl+F More on this topic", "Ctrl+N Next: " + title} {
		if !strings.Contains(line, expected) {
			t.Fatalf("the action line is missing %q:\n%s", expected, line)
		}
	}
	if !strings.Contains(ansi.Strip(model.composer()), line) {
		t.Fatal("the action line should sit with the composer")
	}
}

func TestTheActionLineStaysAwayWhenItWouldMislead(t *testing.T) {
	pending := readingFixture()
	pending.application.state.Question = "Which package owns the calculation?"
	busy := readingFixture()
	busy.busy = true
	earlier := readingFixture()
	earlier.application.state.Turns = append(earlier.application.state.Turns, earlier.application.state.Turns[0])
	earlier.history = 0
	inQuiz := readingFixture()
	inQuiz.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?"}}}
	inQuiz.application.state.Question = "First?"
	for name, model := range map[string]interfaceModel{"a pending question": pending, "a running turn": busy, "an earlier step": earlier, "a quiz": inQuiz} {
		if line := model.actionLine(); line != "" {
			t.Fatalf("%s should hide the action line, got %q", name, ansi.Strip(line))
		}
	}
	last := readingFixture()
	last.application.state.Topic = last.application.course[len(last.application.course)-1].ID
	if strings.Contains(ansi.Strip(last.actionLine()), "Next:") {
		t.Fatal("the last topic has no next topic to offer")
	}
}

func TestACompleteQuizOffersTheNextTopicAndExtend(t *testing.T) {
	model := readingFixture()
	line := model.quizCompleteLine()
	title, _ := model.nextTopicLabel()
	if !strings.Contains(line, "Ctrl+N Next: "+title) || !strings.Contains(line, "/extend adds 5") {
		t.Fatalf("got %q", line)
	}
}

func TestWaitingOnTheSameTopicStillOffersItsSources(t *testing.T) {
	model := interfaceFixture()
	started, _ := model.begin(turnIntent{"question", "go", tutor.Overview})
	model = started.(interfaceModel)
	defer model.cancel()
	content := ansi.Strip(model.View().Content)
	if !strings.Contains(content, "topics (Ctrl+K) or this lesson's sources (Ctrl+S)") {
		t.Fatalf("a wait on the same topic should name both keys:\n%s", content)
	}
	for _, key := range []rune{'k', 's'} {
		opened, _ := press(model, key, tea.ModCtrl)
		if opened.page == "lesson" {
			t.Fatalf("Ctrl+%c should open a page while a turn runs", key)
		}
	}
}

func movingFixture(t *testing.T) interfaceModel {
	model := interfaceFixture()
	model.width, model.height = 120, 36
	model.resize()
	next := model.application.course[1]
	started, _ := model.begin(turnIntent{"lesson", firstLessonMessage, next.ID})
	moving := started.(interfaceModel)
	t.Cleanup(moving.cancel)
	return moving
}

func TestMovingToAnotherTopicHidesTheLastLessonsSourcesAndShortcuts(t *testing.T) {
	model := movingFixture(t)
	if !model.movingToAnotherTopic() {
		t.Fatal("a lesson request for a different topic is a move")
	}
	content := ansi.Strip(model.View().Content)
	if strings.Contains(content, "Ctrl+S") || strings.Contains(content, "Follow a citation") || strings.Contains(content, "QUIZ GOAL") {
		t.Fatalf("a move should not mention the last lesson's sources or quiz:\n%s", content)
	}
	if !strings.Contains(content, "You can browse topics (Ctrl+K) while you wait.") {
		t.Fatalf("a move should still offer Topics:\n%s", content)
	}
	if !strings.Contains(content, ansi.Strip(model.topicTitle())) || model.topicTitle() != tutor.Plain(model.application.course[1].Title) {
		t.Fatalf("the sidebar should name the requested topic: %q", model.topicTitle())
	}
}

func TestSourcesStayClosedWhileMovingAndTopicsStayOpen(t *testing.T) {
	model := movingFixture(t)
	blocked, _ := press(model, 's', tea.ModCtrl)
	if blocked.page != "lesson" || blocked.notice != sourcesWaitNotice {
		t.Fatalf("Ctrl+S should be held with a notice: page %s notice %q", blocked.page, blocked.notice)
	}
	topics, _ := press(model, 'k', tea.ModCtrl)
	if topics.page != "topics" {
		t.Fatalf("Ctrl+K should still open Topics, got %s", topics.page)
	}
}

func TestSourcesReturnOnceTheMoveEnds(t *testing.T) {
	model := movingFixture(t)
	model.busy = false
	opened, _ := press(model, 's', tea.ModCtrl)
	if opened.page != "sources" {
		t.Fatalf("with no turn running, Ctrl+S should open Sources, got %s", opened.page)
	}
}

func TestTheWaitDescribesTheWorkInPlainWords(t *testing.T) {
	for action, want := range map[string]string{
		"lesson": "Writing your lesson from this repository's files.", "quiz": "Writing a question from this repository's files.",
		"answer": "Checking your answer against this repository's files.", "question": "Looking for the answer in this repository's files.",
	} {
		model := interfaceFixture()
		started, _ := model.begin(turnIntent{action, "go", tutor.Overview})
		model = started.(interfaceModel)
		content := ansi.Strip(model.View().Content)
		model.cancel()
		if !strings.Contains(content, want) {
			t.Fatalf("%s should say %q:\n%s", action, want, content)
		}
		if strings.Contains(content, "Grounding") || strings.Contains(content, "grounded") || strings.Contains(content, "next turn") {
			t.Fatalf("%s still uses the old wording:\n%s", action, content)
		}
	}
}
