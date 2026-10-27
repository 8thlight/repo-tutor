package terminal

import (
	tea "charm.land/bubbletea/v2"
	"strconv"
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
	if !strings.Contains(line, "Ctrl+N Next: "+title) || !strings.Contains(line, "Ctrl+F More questions (5)") || strings.Contains(line, "/extend") {
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

func TestTheSidebarCarriesTheSameChoicesAsTheActionLine(t *testing.T) {
	model := readingFixture()
	model.width, model.height = 120, 36
	model.resize()
	content := ansi.Strip(model.View().Content)
	for _, action := range model.courseActions(false) {
		if row := action[0] + "  " + action[1]; !strings.Contains(content, row) {
			t.Fatalf("the sidebar is missing %q:\n%s", row, content)
		}
	}
	bar, sidebar := model.courseActions(true), model.courseActions(false)
	if bar[0] != sidebar[0] || bar[1] != sidebar[1] {
		t.Fatalf("Quiz and More should read the same in both places: %v vs %v", bar, sidebar)
	}
	for _, row := range strings.Split(ansi.Strip(model.sidebarShortcuts()), "\n") {
		if len([]rune(row)) > sidebarWidth-3 {
			t.Fatalf("%q would wrap in the sidebar", row)
		}
	}
	for _, expected := range []string{"Ctrl+K  Choose a topic", "Ctrl+S  Follow a citation"} {
		if !strings.Contains(ansi.Strip(model.sidebarShortcuts()), expected) {
			t.Fatalf("the sidebar should keep %q", expected)
		}
	}
}

func TestTheSidebarOffersOnlyNextQuestionDuringAQuiz(t *testing.T) {
	model := readingFixture()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "partial"}}}
	shortcuts := ansi.Strip(model.sidebarShortcuts())
	if !strings.Contains(shortcuts, "Ctrl+N  Next question") || strings.Contains(shortcuts, "Ctrl+F") || strings.Contains(shortcuts, "Ctrl+G") {
		t.Fatalf("during a quiz Ctrl+N asks the next question and nothing offers more:\n%s", shortcuts)
	}
}

func TestControlFExtendsAFinishedQuiz(t *testing.T) {
	model := readingFixture()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 3, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}, {Text: "B?", Assessment: "partial"}, {Text: "C?", Assessment: "correct"}}}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "answer", Reply: tutor.Reply{Assessment: "correct", Feedback: "Right."}})
	model.application.state.Review = true
	if !quizFinished(model.application.state) {
		t.Fatal("every question is answered and reviewed, so the quiz is finished")
	}
	started, _ := press(model, 'f', tea.ModCtrl)
	defer started.cancel()
	if !started.busy || started.intent.action != "quiz-extend" {
		t.Fatalf("Ctrl+F on a finished quiz should ask for more questions: %+v", started.intent)
	}
	unfinished := model
	unfinished.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}}}
	if quizFinished(unfinished.application.state) {
		t.Fatal("an unfinished quiz should not extend")
	}
}

func TestSourcesEndsWithEveryOtherFileInPathOrder(t *testing.T) {
	model := interfaceFixture()
	model.application.snapshot.Files = append(model.application.snapshot.Files,
		repository.File{Path: "zeta/notes.txt", Content: "later\n"}, repository.File{Path: "alpha/notes.txt", Content: "earlier\n"})
	opened, _ := model.openPicker("sources", "")
	var paths, descriptions []string
	for _, item := range opened.(interfaceModel).picker.Items() {
		entry := item.(browserItem)
		paths, descriptions = append(paths, entry.value), append(descriptions, entry.description)
	}
	seen := map[string]int{}
	for _, path := range paths {
		seen[path]++
	}
	for _, file := range model.application.snapshot.Files {
		if file.Content != "" && seen[file.Path] != 1 {
			t.Fatalf("%s should appear exactly once, appeared %d times: %v", file.Path, seen[file.Path], paths)
		}
	}
	var rest []string
	for index, description := range descriptions {
		if strings.HasPrefix(description, "Repository file") {
			rest = append(rest, paths[index])
		}
	}
	if len(rest) == 0 {
		t.Fatalf("files not cited or excerpted should still be listed: %v", descriptions)
	}
	for index := 1; index < len(rest); index++ {
		if rest[index-1] > rest[index] {
			t.Fatalf("the remaining files should be in path order: %v", rest)
		}
	}
}

func TestFinishingAQuizScoresTheTopicAndARetakeReplacesIt(t *testing.T) {
	model := readingFixture()
	quiz := &tutor.Quiz{Topic: tutor.Overview, Goal: 2, Active: 1, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}, {Text: "B?"}}}
	model.application.state.Quiz, model.application.state.Question = quiz, "B?"
	updated := model.application.recordReply("answer", "my answer", tutor.Overview, "B?", tutor.Reply{Assessment: "partial", Feedback: "Close.", Explanation: "Close."})
	got := updated.Progress[tutor.Overview]
	if got.QuizGoal != 2 || got.QuizCredit != 3 {
		t.Fatalf("a correct and a partial answer on a two-question quiz earn 3 half points: %+v", got)
	}
	unfinished := readingFixture()
	unfinished.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 3, Active: 0, Questions: []tutor.QuizQuestion{{Text: "A?"}}}
	unfinished.application.state.Question = "A?"
	again := unfinished.application.recordReply("answer", "x", tutor.Overview, "A?", tutor.Reply{Assessment: "correct", Feedback: "Yes.", Explanation: "Yes."})
	if again.Progress[tutor.Overview].QuizGoal != 0 {
		t.Fatalf("a quiz with questions left must not be scored yet: %+v", again.Progress[tutor.Overview])
	}
}

func TestProgressShowsCoverageUnderstandingAndEachTopicsLatestQuiz(t *testing.T) {
	model := interfaceFixture()
	core := model.application.course[1]
	model.application.state.Progress = map[string]tutor.Progress{
		tutor.Overview: {Lessons: 2, QuizCredit: 9, QuizGoal: 5},
		core.ID:        {Lessons: 1},
	}
	text := model.progressText()
	total := len(model.application.course)
	for _, expected := range []string{
		"Goal\n" + model.application.state.Goal,
		"Coverage        2 of " + strconv.Itoa(total) + " topics read",
		"Understanding   45% across the topics you have read",
		"Repository overview\n  read · quiz 90% · solid",
		tutor.Plain(core.Title) + "\n  read · no quiz yet",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("Progress is missing %q:\n%s", expected, text)
		}
	}
}

func TestProgressNamesAKnownTopicAndAQuizInProgress(t *testing.T) {
	model := interfaceFixture()
	core := model.application.course[1]
	model.application.state.Progress = map[string]tutor.Progress{core.ID: {Held: true}}
	model.application.state.Quiz = &tutor.Quiz{Topic: core.ID, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}}}
	text := model.progressText()
	if !strings.Contains(text, "not read · known from the probe · 100% · quiz in progress (1 of 5)") {
		t.Fatalf("got:\n%s", text)
	}
}

func TestTheSidebarShowsUnderstandingInsteadOfAnswerCounts(t *testing.T) {
	model := interfaceFixture()
	model.width, model.height = 120, 36
	model.resize()
	model.application.state.Progress = map[string]tutor.Progress{tutor.Overview: {Lessons: 1, QuizCredit: 4, QuizGoal: 5}}
	content := ansi.Strip(model.View().Content)
	if !strings.Contains(content, "Understanding 40%") || strings.Contains(content, "assessed answers") {
		t.Fatalf("the sidebar should lead with the understanding figure:\n%s", content)
	}
}

func TestTheQuestionIndexKeyLivesInTheSidebarAndNotInTheTitle(t *testing.T) {
	model := readingFixture()
	model.width, model.height = 120, 36
	model.resize()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?"}}}
	model.application.state.Question = "First?"
	content := ansi.Strip(model.View().Content)
	if !strings.Contains(content, "Question 1 / 5 · ") || strings.Contains(content, "Ctrl+B index") {
		t.Fatalf("the title should name the page and no keys:\n%s", content)
	}
	if !strings.Contains(ansi.Strip(model.sidebarShortcuts()), "Ctrl+B  Question index") {
		t.Fatalf("the sidebar should list Ctrl+B during a quiz:\n%s", ansi.Strip(model.sidebarShortcuts()))
	}
	if strings.Contains(ansi.Strip(readingFixture().sidebarShortcuts()), "Ctrl+B") {
		t.Fatal("with no quiz there is no index to open")
	}
}

func quizFeedbackFixture() interfaceModel {
	model := readingFixture()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?", Assessment: "partial"}}}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "answer", Reply: tutor.Reply{Assessment: "partial", Feedback: "Close."}})
	model.application.state.Review = true
	return model
}

func TestControlNAsksTheNextQuestionWhileAQuizHasQuestionsLeft(t *testing.T) {
	model := quizFeedbackFixture()
	started, _ := press(model, 'n', tea.ModCtrl)
	defer started.cancel()
	if !started.busy || started.intent.action != "quiz-next" || started.intent.topic != tutor.Overview {
		t.Fatalf("Ctrl+N after feedback should ask the next question: %+v", started.intent)
	}
}

func TestControlNWaitsForAnAnswerBeforeAskingTheNextQuestion(t *testing.T) {
	model := quizFeedbackFixture()
	model.application.state.Review = false
	model.application.state.Turns = model.application.state.Turns[:1]
	model.application.state.Question = "First?"
	next, _ := press(model, 'n', tea.ModCtrl)
	if next.busy || !strings.Contains(next.notice, "answer or skip") {
		t.Fatalf("an open question should hold Ctrl+N: %q", next.notice)
	}
}

func TestControlFPointsAtControlNDuringAQuiz(t *testing.T) {
	model := quizFeedbackFixture()
	next, _ := press(model, 'f', tea.ModCtrl)
	if next.busy || next.notice != "Ctrl+N asks the next question." {
		t.Fatalf("Ctrl+F is not used while a quiz has questions left: busy=%v %q", next.busy, next.notice)
	}
}

func TestAfterAFinishedQuizControlFAddsQuestionsAndTheSidebarOffersNext(t *testing.T) {
	finished := quizFeedbackFixture()
	finished.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 2, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}, {Text: "B?", Assessment: "partial"}}}
	more, _ := press(finished, 'f', tea.ModCtrl)
	defer more.cancel()
	if more.intent.action != "quiz-extend" {
		t.Fatalf("Ctrl+F on a finished quiz should add questions: %+v", more.intent)
	}
	shortcuts := ansi.Strip(finished.sidebarShortcuts())
	if !strings.Contains(shortcuts, "Ctrl+F  More questions (5)") || !strings.Contains(shortcuts, "Ctrl+N  Next topic") {
		t.Fatalf("a finished quiz offers more questions and the next topic:\n%s", shortcuts)
	}
}
