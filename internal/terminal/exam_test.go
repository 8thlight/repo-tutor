package terminal

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func examFixture(t *testing.T, topicsRead int) interfaceModel {
	t.Helper()
	model := readingFixture()
	model.width, model.height = 120, 36
	model.resize()
	course := []tutor.Topic{{ID: tutor.Overview, Title: "Repository overview"}}
	progress := map[string]tutor.Progress{tutor.Overview: {Lessons: 1}}
	for index := 1; index < 14; index++ {
		id := fmt.Sprintf("pkg%02d", index)
		course = append(course, tutor.Topic{ID: id, Title: fmt.Sprintf("Package %d", index)})
		if index < topicsRead {
			progress[id] = tutor.Progress{Lessons: 1}
		}
	}
	model.application.course = course
	model.application.state.Progress = progress
	model.application.state.Quiz, model.application.state.Question = nil, ""
	return model
}

func startedExam(t *testing.T) interfaceModel {
	model := examFixture(t, 12)
	plan := tutor.ExamPlan(model.application.course, model.application.state.Progress, model.application.settings.ExamLength)
	updated := model.application.recordReply("exam", examMessage, plan[0], "", tutor.Reply{Explanation: "Exam.", Question: "First?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations})
	model.application.state = updated
	model.history = len(updated.Turns) - 1
	return model
}

func homeHas(model interfaceModel, place string) (homeEntry, bool) {
	for _, entry := range model.homeEntries() {
		if entry.place == place {
			return entry, true
		}
	}
	return homeEntry{}, false
}

func TestHomeOffersAnExamOnlyAfterTenTopicsAreRead(t *testing.T) {
	if _, offered := homeHas(examFixture(t, 9), placeExam); offered {
		t.Fatal("nine topics read is not enough for an exam")
	}
	entry, offered := homeHas(examFixture(t, 10), placeExam)
	if !offered || entry.label != "Take an exam" || !strings.Contains(entry.description, "10 questions") {
		t.Fatalf("ten topics read should offer a ten-question exam: %+v %v", entry, offered)
	}
	if entry, _ := homeHas(examFixture(t, 14), placeExam); !strings.Contains(entry.description, "10 questions") {
		t.Fatalf("the default length caps the exam at ten: %q", entry.description)
	}
	open := startedExam(t)
	if _, offered := homeHas(open, placeExam); offered {
		t.Fatal("an exam in progress is not offered again")
	}
}

func TestTheExamEntryStartsAnExamRequestWithNoTopicYet(t *testing.T) {
	model := examFixture(t, 12)
	entry, _ := homeHas(model, placeExam)
	started, _ := model.openHomeEntry(entry)
	exam := started.(interfaceModel)
	defer exam.cancel()
	if !exam.busy || exam.intent.action != "exam" || exam.intent.topic != "" {
		t.Fatalf("the entry should request the exam's first question: %+v", exam.intent)
	}
	short := examFixture(t, 4)
	refused, _ := short.openHomeEntry(homeEntry{place: placeExam})
	if refused.(interfaceModel).busy || !strings.Contains(refused.(interfaceModel).notice, "at least 10 topics") {
		t.Fatal("too few topics read should say so")
	}
}

func TestTheFirstExamQuestionBuildsThePlanAndSitsOnItsTopic(t *testing.T) {
	model := startedExam(t)
	state := model.application.state
	quiz := state.Quiz
	if quiz == nil || !quiz.Exam || quiz.Goal != 10 || len(quiz.Plan) != 10 || state.Topic != quiz.Plan[0] || quiz.Questions[0].Topic != quiz.Plan[0] {
		t.Fatalf("the exam should carry its plan and start on its first topic: %+v topic=%s", quiz, state.Topic)
	}
	if state.Probe != nil || state.Question != "First?" {
		t.Fatalf("the first question should be open: %q", state.Question)
	}
}

func TestTheNextExamQuestionIsAskedOnTheNextPlannedTopic(t *testing.T) {
	model := startedExam(t)
	model.application.state.Quiz.Questions[0].Assessment = "correct"
	model.application.state.Question, model.application.state.Review = "", true
	intent, err := continueIntent(model.application.state, model.application.course, false)
	plan := model.application.state.Quiz.Plan
	if err != nil || intent.action != "quiz-next" || intent.topic != plan[1] || !strings.Contains(intent.message, "course exam") {
		t.Fatalf("the next question should be on the next planned topic: %+v %v", intent, err)
	}
	next := model.application.recordReply("quiz-next", intent.message, intent.topic, "", tutor.Reply{Question: "Second?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations})
	if next.Topic != plan[1] || next.Quiz == nil || !next.Quiz.Exam || next.Quiz.Questions[1].Topic != plan[1] || next.Quiz.Active != 1 {
		t.Fatalf("recording it should move the exam on and keep it: %+v topic=%s", next.Quiz, next.Topic)
	}
}

func TestFinishingAnExamSavesItsResultAndLeavesTopicQuizScoresAlone(t *testing.T) {
	model := startedExam(t)
	quiz := model.application.state.Quiz
	quiz.Questions[0].Assessment = "correct"
	for index := 1; index < quiz.Goal; index++ {
		quiz.Questions = append(quiz.Questions, tutor.QuizQuestion{Text: "Q?", Topic: quiz.Plan[index], Assessment: "correct"})
	}
	quiz.Active = quiz.Goal - 1
	quiz.Questions[quiz.Active].Assessment = ""
	model.application.state.Question, model.application.state.Review = "Q?", false
	model.application.state.Progress[quiz.Plan[0]] = tutor.Progress{Lessons: 1, QuizCredit: 8, QuizGoal: 5}
	finished := model.application.recordReply("answer", "mine", quiz.Plan[quiz.Goal-1], "Q?", tutor.Reply{Assessment: "partial", Feedback: "Close.", Explanation: "Close."})
	if finished.LastExam == nil || finished.LastExam.Goal != 10 || finished.LastExam.Credit != 19 || finished.LastExam.Percent() != 95 {
		t.Fatalf("the finished exam should be saved: %+v", finished.LastExam)
	}
	if got := finished.Progress[quiz.Plan[0]]; got.QuizCredit != 8 || got.QuizGoal != 5 {
		t.Fatalf("an exam must not change a topic's quiz score: %+v", got)
	}
}

func TestSkippingTheLastExamQuestionFinishesItWithoutARequest(t *testing.T) {
	model := startedExam(t)
	quiz := model.application.state.Quiz
	quiz.Questions[0].Assessment = "correct"
	for index := 1; index < quiz.Goal; index++ {
		quiz.Questions = append(quiz.Questions, tutor.QuizQuestion{Text: "Q?", Topic: quiz.Plan[index], Assessment: "correct"})
	}
	quiz.Active = quiz.Goal - 1
	quiz.Questions[quiz.Active].Assessment = ""
	model.application.state.Question = "Q?"
	armed, _ := press(model, 'n', tea.ModCtrl)
	done, _ := press(armed, 'n', tea.ModCtrl)
	if done.busy || done.application.state.LastExam == nil || done.application.state.LastExam.Credit != 18 {
		t.Fatalf("the last skip should finish the exam with nine correct answers: busy=%v %+v", done.busy, done.application.state.LastExam)
	}
}

func TestAnExamCannotBeExtended(t *testing.T) {
	model := startedExam(t)
	quiz := model.application.state.Quiz
	for index := range quiz.Goal {
		if index >= len(quiz.Questions) {
			quiz.Questions = append(quiz.Questions, tutor.QuizQuestion{Text: "Q?", Topic: quiz.Plan[index]})
		}
		quiz.Questions[index].Assessment = "correct"
	}
	model.application.state.Question, model.application.state.Review = "", true
	next, _ := press(model, 'f', tea.ModCtrl)
	if next.busy || next.notice != "an exam cannot be extended" {
		t.Fatalf("Ctrl+F on a finished exam should refuse: busy=%v %q", next.busy, next.notice)
	}
}

func TestTheExamNamesItselfInTheTitleSidebarAndLeavePrompt(t *testing.T) {
	model := startedExam(t)
	content := ansi.Strip(model.View().Content)
	if !strings.Contains(content, "Exam · Question 1 / 10 · Repository overview") || !strings.Contains(content, "EXAM") {
		t.Fatalf("the exam should name itself:\n%s", content)
	}
	if model.activeTab() != "quiz" {
		t.Fatalf("the Quiz tab is active during an exam, got %s", model.activeTab())
	}
	if !strings.Contains(ansi.Strip(model.sidebarShortcuts()), "Ctrl+L  Leave exam") {
		t.Fatal("the sidebar should offer leaving the exam")
	}
	armed, _ := press(model, 'l', tea.ModCtrl)
	if armed.notice != "Press Ctrl+L again to leave this exam. It will not be scored." {
		t.Fatalf("got %q", armed.notice)
	}
}

func TestPickingATopicDuringAnExamEndsTheExam(t *testing.T) {
	model := startedExam(t)
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: "pkg05", Action: "lesson", Reply: model.application.state.Turns[0].Reply})
	resumed, found := resumeTopic(model.application.state, "pkg05")
	if !found || resumed.Quiz != nil {
		t.Fatalf("choosing a topic should leave the exam: found=%v quiz=%v", found, resumed.Quiz)
	}
}

func TestTheNextExamQuestionIsPrefetchedForThePlannedTopic(t *testing.T) {
	model := startedExam(t)
	key, wanted := nextQuestionKey(model.application.state)
	plan := model.application.state.Quiz.Plan
	if !wanted || key != questionKey(plan[1], 1) {
		t.Fatalf("the prefetch should be for the next planned topic: %q %v", key, wanted)
	}
}

func TestProgressShowsTheLatestExamAndOffersAnotherFromHome(t *testing.T) {
	model := examFixture(t, 12)
	model.application.state.LastExam = &tutor.ExamResult{Credit: 14, Goal: 10, Topics: []tutor.ExamTopicResult{{Topic: "pkg01", Credit: 2}, {Topic: "pkg02", Credit: 1}, {Topic: "pkg03", Credit: 0}}}
	text := model.progressText()
	for _, expected := range []string{"Latest exam     70% · developing (10 questions)", "Latest exam, by topic", "Package 1 · correct", "Package 2 · partial", "Package 3 · missed", "Take an exam from Home: 10 questions"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("Progress is missing %q:\n%s", expected, text)
		}
	}
	few := examFixture(t, 4)
	if strings.Contains(few.progressText(), "Take an exam") {
		t.Fatal("too few topics read should not suggest an exam")
	}
}

func TestExamMissesSurfaceInReviewWhenNoQuizIsWeak(t *testing.T) {
	model := examFixture(t, 12)
	model.application.state.LastExam = &tutor.ExamResult{Goal: 2, Topics: []tutor.ExamTopicResult{{Topic: "pkg02", Credit: 0}, {Topic: "pkg03", Credit: 2}}}
	entry, offered := homeHas(model, placeReview)
	if !offered || !strings.Contains(entry.description, "Package 2") {
		t.Fatalf("an exam miss should be offered for review: %+v %v", entry, offered)
	}
}

func TestExamStateIsSavedWithTheSession(t *testing.T) {
	model := startedExam(t)
	data := model.application.state
	if data.Quiz == nil || !data.Quiz.Exam || len(data.Quiz.Plan) != 10 {
		t.Fatal("the exam and its plan live in the session state")
	}
}
