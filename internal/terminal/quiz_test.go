package terminal

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func TestContinueRequiresReviewAndStopsAtQuizGoal(t *testing.T) {
	model := interfaceFixture()
	if _, err := continueIntent(model.application.state, model.application.course, false); err == nil {
		t.Fatal("continued past an unanswered question")
	}
	model.application.state.Review = true
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "partial"}}}
	calls := 0
	model.run = func(_ context.Context, state session.State, action, _, _ string) (session.State, error) {
		calls++
		if action != "quiz-next" {
			t.Fatalf("continue should request the next quiz question, got %s", action)
		}
		return state, nil
	}
	model.rebuildContent()
	model.View()
	if calls != 0 || !strings.Contains(model.composer(), "Continue") {
		t.Fatal("viewing feedback should wait for Continue")
	}
	started, command := model.submit()
	if command == nil || !started.(interfaceModel).busy {
		t.Fatal("empty Enter on review should continue")
	}
	finishedCommand(t, command)
	started.(interfaceModel).cancel()
	if calls != 1 {
		t.Fatal("one explicit continue should produce one request")
	}
	quiz := model.application.state.Quiz
	for len(quiz.Questions) < 5 {
		quiz.Questions = append(quiz.Questions, tutor.QuizQuestion{Text: "Next?", Assessment: "correct"})
	}
	if _, err := continueIntent(model.application.state, model.application.course, false); err == nil {
		t.Fatal("continued beyond the quiz goal without an extension")
	}
	intent, err := continueIntent(model.application.state, model.application.course, true)
	if err != nil || intent.action != "quiz-extend" || !strings.Contains(quizText(quiz), "Goal reached") {
		t.Fatal("completed quiz should offer an explicit extension")
	}
}

func TestLoadingReplacesPreviousLessonAndKeepsBrowsingAvailable(t *testing.T) {
	model := interfaceFixture()
	started, _ := model.begin(turnIntent{"quiz", "Start a quiz", tutor.Overview})
	model = started.(interfaceModel)
	defer model.cancel()
	content := ansi.Strip(model.View().Content)
	if !strings.Contains(content, "Loading your next exercise") || strings.Contains(content, "The CLI is the front door") {
		t.Fatalf("loading should replace the stale lesson:\n%s", content)
	}
	model, _ = press(model, 's', tea.ModCtrl)
	if model.page != "sources" || !model.busy {
		t.Fatal("loading should permit source browsing")
	}
}

func TestSkippedQuestionHasBoundaryAndFailedContinueKeepsReview(t *testing.T) {
	model := interfaceFixture()
	reply := tutor.Reply{Assessment: "none", Explanation: "The core package contains the calculation."}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "answer", Message: "Please skip this question", Reply: reply})
	model.application.state.Question, model.application.state.Review = "", true
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "Where?", Assessment: "skipped"}}}
	model.history = 1
	model.rebuildContent()
	content := ansi.Strip(model.lessonContent())
	if !strings.Contains(content, "── FEEDBACK") || !strings.Contains(content, "── END OF QUESTION") || !strings.Contains(model.composer(), "Continue") {
		t.Fatalf("ungraded skip lost its question boundary:\n%s", content)
	}
	model.run = func(context.Context, session.State, string, string, string) (session.State, error) {
		return session.State{}, errors.New("test connection failed")
	}
	started, command := model.continueQuiz(false)
	next, _ := started.(interfaceModel).Update(finishedCommand(t, command))
	model = next.(interfaceModel)
	if !reviewing(model.application.state) || model.application.state.Quiz.Completed() != 1 || len(model.application.state.Turns) != 2 || !strings.Contains(model.lessonContent(), reply.Explanation) {
		t.Fatal("failed continue lost saved feedback or advanced the quiz")
	}
	if model.intent.action != "quiz-next" || !strings.Contains(model.errorText, "test connection failed") {
		t.Fatal("failed continue must offer retry for the same next-question request")
	}
}

func TestRephrasedQuestionRemainsAnswerable(t *testing.T) {
	model := interfaceFixture()
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "answer", Message: "What do you mean?", Reply: tutor.Reply{Assessment: "none", Explanation: "Consider the package responsibilities from the overview.", Question: "Which package owns calculations?"}})
	model.application.state.Question = model.application.state.Turns[1].Reply.Question
	model.history = 1
	composer := ansi.Strip(model.composer())
	if strings.Contains(composer, "Continue") || !strings.Contains(composer, model.application.state.Question) {
		t.Fatalf("clarification must keep the revised question available to answer:\n%s", composer)
	}
}

func TestLastLessonMatchesTopicAndIgnoresQuizFeedback(t *testing.T) {
	turns := []session.Turn{
		{Topic: "core", Action: "lesson", Reply: tutor.Reply{Explanation: "An older lesson."}},
		{Topic: "core", Action: "lesson", Reply: tutor.Reply{Explanation: "The calculation belongs to core."}},
		{Topic: "cli", Action: "lesson", Reply: tutor.Reply{Explanation: "CLI handles commands."}},
		{Topic: "core", Action: "quiz", Reply: tutor.Reply{Explanation: "Try an exercise."}},
		{Topic: "core", Action: "answer", Reply: tutor.Reply{Explanation: "Correct."}},
	}
	lesson := lastLesson(turns, "core")
	if lesson == nil || lesson.Explanation != "The calculation belongs to core." {
		t.Fatal("quiz must retain the latest lesson for its topic across intervening turns")
	}
	if lastLesson(turns, "unseen") != nil {
		t.Fatal("an unseen topic inherited another topic's teaching")
	}
}

func TestReviewAndLoadingFitSmallTerminals(t *testing.T) {
	for _, size := range []struct{ width, height int }{{44, 18}, {60, 24}, {80, 24}, {120, 36}} {
		for _, stage := range []string{"review", "complete", "loading"} {
			model := interfaceFixture()
			model.width, model.height = size.width, size.height
			model.application.state.Review = true
			model.application.state.Quiz = &tutor.Quiz{Goal: 5}
			if stage == "complete" {
				for index := 0; index < 5; index++ {
					model.application.state.Quiz.Questions = append(model.application.state.Quiz.Questions, tutor.QuizQuestion{Text: "Question?", Assessment: "correct"})
				}
			}
			model.busy = stage == "loading"
			model.resize()
			content := ansi.Strip(model.View().Content)
			if len(strings.Split(content, "\n")) > size.height || !strings.Contains(content, "Ctrl+Q") {
				t.Fatalf("%s at %dx%d lost exit guidance or overflowed:\n%s", stage, size.width, size.height, content)
			}
			for _, line := range strings.Split(content, "\n") {
				if ansi.StringWidth(line) > size.width {
					t.Fatalf("%s line overflows %dx%d: %q", stage, size.width, size.height, line)
				}
			}
		}
	}
}

func TestSourceHighlightingPreservesCodeAndLineNumbers(t *testing.T) {
	for _, dark := range []bool{true, false} {
		model := interfaceFixture()
		model.dark, model.noColor = dark, false
		for path, source := range map[string]string{"main.go": "package main\n/* a\ncomment */\nfunc main() {\n\tprintln(\"hello\")\n}\n", "value.py": "def value():\n    return \"hello\"\n", "main.ts": "const value: string = `hello\nworld`;\n"} {
			colored := strings.Join(model.highlightSource(path, source), "\n")
			want := strings.ReplaceAll(strings.TrimSuffix(source, "\n"), "\t", "    ")
			if ansi.Strip(colored) != want || !strings.Contains(colored, "\x1b[") {
				t.Fatalf("highlighting altered %s or lost colors:\n%q", path, colored)
			}
			model.noColor = true
			plain := strings.Join(model.highlightSource(path, source), "\n")
			if plain != want {
				t.Fatalf("NO_COLOR altered source %s", path)
			}
			model.noColor = false
		}
		opened, _ := model.openSource(model.application.state.Turns[0].Reply.Citations[0])
		content := ansi.Strip(opened.(interfaceModel).pageContent)
		if !strings.Contains(content, "›    2  func main() {}") {
			t.Fatalf("source citation marker or line number changed:\n%s", content)
		}
	}
}

func TestQuizTabStaysActiveAcrossStartAndIndex(t *testing.T) {
	model := interfaceFixture()
	started, _ := model.begin(turnIntent{"quiz", "Start a quiz", tutor.Overview})
	model = started.(interfaceModel)
	defer model.cancel()
	if model.activeTab() != "quiz" {
		t.Fatalf("starting a quiz should select the Quiz tab, got %s", model.activeTab())
	}
	idle := interfaceFixture()
	idle.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?"}}}
	if idle.activeTab() != "quiz" {
		t.Fatal("an active quiz should select the Quiz tab")
	}
	index, _ := idle.openDocument("quiz", "Quiz questions", quizText(idle.application.state.Quiz))
	returned, _ := index.(interfaceModel).learn()
	if returned.(interfaceModel).activeTab() != "quiz" {
		t.Fatal("leaving the index should return to the Quiz tab")
	}
	if interfaceFixture().activeTab() != "lesson" {
		t.Fatal("a lesson should keep the Learn tab")
	}
}

func TestSelectingSkippedSlotReturnsToItsQuestion(t *testing.T) {
	model := interfaceFixture()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Active: 1, Questions: []tutor.QuizQuestion{{Text: "Where?", Assessment: "skipped"}, {Text: "Why?", Assessment: "correct"}}}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "answer", Reply: tutor.Reply{Assessment: "correct", Feedback: "Right."}})
	model.application.state.Question, model.application.state.Review = "", true
	opened, _ := model.openPicker("quiz", "")
	model = opened.(interfaceModel)
	if !model.isPicker() || model.activeTab() != "quiz" || len(model.picker.Items()) != 5 {
		t.Fatal("the index should list every goal slot as a selectable item")
	}
	selected, command := model.selectQuizSlot(0)
	model = selected.(interfaceModel)
	if command != nil && model.busy {
		t.Fatal("reopening a skipped question must not call the agent")
	}
	state := model.application.state
	if model.page != "lesson" || state.Question != "Where?" || state.Quiz.Active != 0 || reviewing(state) || strings.Contains(model.composer(), "Continue") {
		t.Fatalf("the skipped question should be current and answerable: %+v", state)
	}
	model.application.state.Quiz.Questions[1].Turn = 1
	answered, _ := model.selectQuizSlot(1)
	if viewed := answered.(interfaceModel); viewed.page != "lesson" || viewed.history != 1 || viewed.application.state.Question != "Where?" {
		t.Fatal("an answered slot should show its feedback and stay unopened")
	}
	later, _ := model.selectQuizSlot(4)
	if !strings.Contains(later.(interfaceModel).notice, "generated when you reach it") {
		t.Fatal("a later slot must explain that it does not exist yet")
	}
}

func TestSelectingNextSlotGeneratesItOnlyAfterReview(t *testing.T) {
	model := interfaceFixture()
	model.application.state.Review = true
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "partial"}}}
	model.run = func(_ context.Context, state session.State, action, _, _ string) (session.State, error) {
		if action != "quiz-next" {
			t.Fatalf("the next slot should generate the next question, got %s", action)
		}
		return state, nil
	}
	started, command := model.selectQuizSlot(1)
	if command == nil || !started.(interfaceModel).busy {
		t.Fatal("selecting the next slot should continue the quiz")
	}
	started.(interfaceModel).cancel()
}

func TestAnswerTurnIsRecordedOnceAndShiftsWithTrimmedHistory(t *testing.T) {
	before := &tutor.Quiz{Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "skipped"}, {Text: "Second?"}}}
	after := &tutor.Quiz{Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "skipped"}, {Text: "Second?", Assessment: "correct"}}}
	recorded := tutor.RecordAnswerTurn(before, after, 7)
	if recorded.Questions[1].Turn != 7 || recorded.Questions[0].Turn != 0 || after.Questions[1].Turn != 0 {
		t.Fatal("only the newly answered slot should record its turn without mutating the input")
	}
	again := tutor.RecordAnswerTurn(recorded, recorded, 9)
	if again.Questions[1].Turn != 7 {
		t.Fatal("an answered slot must keep the turn that answered it")
	}
	if shifted := tutor.ShiftQuizTurns(recorded, 10); shifted.Questions[1].Turn != 0 || recorded.Questions[1].Turn != 7 {
		t.Fatal("a trimmed turn must read as no longer saved")
	}
}
