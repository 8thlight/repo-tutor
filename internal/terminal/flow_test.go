package terminal

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"strconv"
	"strings"
	"testing"
	"time"

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

func openQuestionFixture() interfaceModel {
	model := quizFeedbackFixture()
	model.application.state.Review = false
	model.application.state.Turns = model.application.state.Turns[:1]
	model.application.state.Question = "First?"
	model.application.state.Quiz.Questions[0].Assessment = ""
	return model
}

func TestControlNOnAnOpenQuestionOnlyArmsTheSkip(t *testing.T) {
	armed, command := press(openQuestionFixture(), 'n', tea.ModCtrl)
	if armed.busy || !armed.skipArmed || armed.notice != skipPrompt || command == nil {
		t.Fatalf("the first press should arm and start a timer: busy=%v armed=%v %q", armed.busy, armed.skipArmed, armed.notice)
	}
}

func TestASecondControlNSkipsLocallyAndAsksTheNextQuestion(t *testing.T) {
	model := openQuestionFixture()
	turns := len(model.application.state.Turns)
	armed, _ := press(model, 'n', tea.ModCtrl)
	skipped, _ := press(armed, 'n', tea.ModCtrl)
	defer skipped.cancel()
	state := skipped.application.state
	if state.Quiz.Questions[0].Assessment != "skipped" || state.Question != "" || !state.Review || len(state.Turns) != turns {
		t.Fatalf("the skip should be recorded in the quiz with no model message: %+v turns=%d", state.Quiz.Questions[0], len(state.Turns))
	}
	if !skipped.busy || skipped.intent.action != "quiz-next" || skipped.skipArmed {
		t.Fatalf("after the skip the next question should be requested, not an answer review: %+v", skipped.intent)
	}
}

func TestSkippingTheLastQuestionFinishesTheQuizAndScoresItWithoutARequest(t *testing.T) {
	model := openQuestionFixture()
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 2, Active: 1, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}, {Text: "B?"}}}
	model.application.state.Question = "B?"
	armed, _ := press(model, 'n', tea.ModCtrl)
	done, _ := press(armed, 'n', tea.ModCtrl)
	if done.busy || !quizFinished(done.application.state) || done.notice != "Skipped. That was the last question." {
		t.Fatalf("the last skip should finish the quiz and ask for nothing: busy=%v %q", done.busy, done.notice)
	}
	if got := done.application.state.Progress[tutor.Overview]; got.QuizGoal != 2 || got.QuizCredit != 2 {
		t.Fatalf("a correct answer and a skip on two questions earn 2 half points: %+v", got)
	}
}

func TestSkippingAProbeQuestionMovesTheProbeOnWithoutAModelMessage(t *testing.T) {
	model := readingFixture()
	model.application.course = append(model.application.course, tutor.Topic{ID: "core", Title: "core"}, tutor.Topic{ID: "cli", Title: "cli"})
	model.application.state.Quiz = nil
	model.application.state.Probe = &tutor.Probe{Topics: []string{"core", "cli"}, Asked: 1}
	model.application.state.Topic, model.application.state.Question = "core", "Do you know core?"
	armed, _ := press(model, 'n', tea.ModCtrl)
	skipped, _ := press(armed, 'n', tea.ModCtrl)
	defer skipped.cancel()
	if skipped.application.state.Probe.Answered != 1 || !skipped.busy || skipped.intent.action != "probe" || skipped.intent.topic != "cli" {
		t.Fatalf("the skip should count and the next probe question should be requested: %+v %+v", skipped.application.state.Probe, skipped.intent)
	}
}

func TestAnyOtherKeyDisarmsTheSkip(t *testing.T) {
	armed, _ := press(openQuestionFixture(), 'n', tea.ModCtrl)
	typed := typeText(armed, "a")
	if typed.skipArmed || typed.notice == skipPrompt {
		t.Fatal("typing should cancel a pending skip and clear its prompt")
	}
	again, _ := press(typed, 'n', tea.ModCtrl)
	if again.busy || !again.skipArmed {
		t.Fatal("after disarming, Ctrl+N should arm again and not skip")
	}
}

func TestTheSkipPromptLapsesAfterTheWindowAndStaleTimersAreIgnored(t *testing.T) {
	skipWindow = time.Millisecond
	defer func() { skipWindow = 5 * time.Second }()
	armed, command := press(openQuestionFixture(), 'n', tea.ModCtrl)
	expired, ok := command().(skipExpired)
	if !ok {
		t.Fatal("arming should schedule an expiry message")
	}
	lapsed, _ := armed.Update(expired)
	if lapsed.(interfaceModel).skipArmed || lapsed.(interfaceModel).notice != "" {
		t.Fatal("the prompt should lapse when its timer fires")
	}
	typed := typeText(armed, "a")
	rearmed, _ := press(typed, 'n', tea.ModCtrl)
	stillArmed, _ := rearmed.Update(expired)
	if !stillArmed.(interfaceModel).skipArmed {
		t.Fatal("a timer from an earlier press must not disarm a newer one")
	}
}

func TestTheFrameAndFooterShowTheSkipPromptInTheWarningColor(t *testing.T) {
	model := openQuestionFixture()
	model.noColor = false
	if lines := frameLines(model); !strings.Contains(lines[len(lines)-1], "Enter send · Ctrl+N skip") {
		t.Fatalf("an open question offers send and skip:\n%s", strings.Join(lines, "\n"))
	}
	armed, _ := press(model, 'n', tea.ModCtrl)
	if lines := frameLines(armed); !strings.Contains(lines[len(lines)-1], "Ctrl+N again to skip") || strings.Contains(lines[len(lines)-1], "Enter send") {
		t.Fatalf("an armed skip replaces the hint:\n%s", strings.Join(lines, "\n"))
	}
	warning := lipgloss.NewStyle().Foreground(armed.palette().warning).Bold(true).Render("Ctrl+N again to skip")
	if !strings.Contains(armed.composer(), warning) {
		t.Fatal("the frame's skip prompt should use the warning color")
	}
	footer := lipgloss.NewStyle().Foreground(armed.palette().warning).Bold(true).Render(skipPrompt)
	if !strings.Contains(armed.View().Content, footer[:strings.Index(footer, skipPrompt)]) {
		t.Fatal("the footer prompt should use the warning color")
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

func TestTheSidebarOffersSkipOnAnOpenQuestionAndNextAfterFeedback(t *testing.T) {
	shortcuts := ansi.Strip(openQuestionFixture().sidebarShortcuts())
	if !strings.Contains(shortcuts, "Ctrl+N  Skip question") || !strings.Contains(shortcuts, "Ctrl+B  Question index") {
		t.Fatalf("an open question can be skipped with Ctrl+N:\n%s", shortcuts)
	}
	if !strings.Contains(ansi.Strip(quizFeedbackFixture().sidebarShortcuts()), "Ctrl+N  Next question") {
		t.Fatal("after feedback the sidebar should offer the next question")
	}
}

func TestControlNWorksFromTheQuizIndexToo(t *testing.T) {
	model := quizFeedbackFixture()
	opened, _ := model.openPicker("quiz", "")
	next, _ := press(opened.(interfaceModel), 'n', tea.ModCtrl)
	defer next.cancel()
	if !next.busy || next.intent.action != "quiz-next" {
		t.Fatalf("Ctrl+N should work on the index as it does on the lesson: %+v page=%s notice=%q", next.intent, next.page, next.notice)
	}
}

func quizWithLessonFixture() interfaceModel {
	model := readingFixture()
	model.application.state.Turns[0].Action = "lesson"
	quizReply := tutor.Reply{Explanation: "Try this.", Question: "First?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "quiz", Reply: quizReply})
	model.application.state.Quiz = &tutor.Quiz{Topic: tutor.Overview, Goal: 5, Questions: []tutor.QuizQuestion{{Text: "First?"}}}
	model.application.state.Question = "First?"
	model.history = len(model.application.state.Turns) - 1
	return model
}

func TestControlLInAQuizOnlyArmsTheExit(t *testing.T) {
	armed, command := press(quizWithLessonFixture(), 'l', tea.ModCtrl)
	if !armed.leaveArmed || armed.notice != armed.leavePrompt() || command == nil || armed.application.state.Quiz == nil {
		t.Fatalf("the first press should arm and keep the quiz: armed=%v %q", armed.leaveArmed, armed.notice)
	}
	if lines := frameLines(armed); !strings.Contains(lines[len(lines)-1], "Ctrl+L again to leave") {
		t.Fatalf("the frame should name the second press:\n%s", strings.Join(lines, "\n"))
	}
}

func TestASecondControlLLeavesTheQuizAndShowsTheLesson(t *testing.T) {
	armed, _ := press(quizWithLessonFixture(), 'l', tea.ModCtrl)
	left, _ := press(armed, 'l', tea.ModCtrl)
	state := left.application.state
	last := state.Turns[len(state.Turns)-1]
	if state.Quiz != nil || state.Question != "" || reviewing(state) || last.Action != "lesson" || !last.Resumed || left.page != "lesson" || left.history != len(state.Turns)-1 {
		t.Fatalf("leaving should drop the quiz and show the lesson: quiz=%v question=%q last=%+v", state.Quiz, state.Question, last)
	}
	if left.leaveArmed || left.notice != "Left the quiz" || left.activeTab() != "course" {
		t.Fatalf("after leaving the course tab is active and nothing stays armed: %q %s", left.notice, left.activeTab())
	}
	if line := left.actionLine(); line == "" {
		t.Fatal("the reading actions should come back")
	}
}

func TestAnyOtherKeyOrTheTimerCancelsTheExit(t *testing.T) {
	armed, command := press(quizWithLessonFixture(), 'l', tea.ModCtrl)
	typed := typeText(armed, "a")
	if typed.leaveArmed || typed.notice == armed.leavePrompt() {
		t.Fatal("typing should cancel the exit prompt")
	}
	skipWindow = time.Millisecond
	defer func() { skipWindow = 5 * time.Second }()
	armed, command = press(quizWithLessonFixture(), 'l', tea.ModCtrl)
	_ = command
	lapsed, _ := armed.Update(skipExpired{armed.skipGeneration})
	if lapsed.(interfaceModel).leaveArmed || lapsed.(interfaceModel).notice != "" {
		t.Fatal("the exit prompt should lapse with its timer")
	}
}

func TestArmingASkipAndArmingAnExitCancelEachOther(t *testing.T) {
	skip, _ := press(quizWithLessonFixture(), 'n', tea.ModCtrl)
	exit, _ := press(skip, 'l', tea.ModCtrl)
	if skip.skipArmed != true || exit.skipArmed || !exit.leaveArmed {
		t.Fatalf("pressing Ctrl+L should replace a pending skip: skip=%v leave=%v", exit.skipArmed, exit.leaveArmed)
	}
	back, _ := press(exit, 'n', tea.ModCtrl)
	if back.leaveArmed || !back.skipArmed {
		t.Fatal("and Ctrl+N should replace a pending exit")
	}
}

func TestControlLStillReturnsToTheLessonFromOtherPagesAndOutsideAQuiz(t *testing.T) {
	model := quizWithLessonFixture()
	sources, _ := press(model, 's', tea.ModCtrl)
	back, _ := press(sources, 'l', tea.ModCtrl)
	if back.page != "lesson" || back.leaveArmed || back.application.state.Quiz == nil {
		t.Fatalf("from another page Ctrl+L just returns: page=%s armed=%v", back.page, back.leaveArmed)
	}
	reading := readingFixture()
	same, _ := press(reading, 'l', tea.ModCtrl)
	if same.leaveArmed || same.page != "lesson" {
		t.Fatal("with nothing to leave, Ctrl+L arms nothing")
	}
}

func TestTheSidebarOffersLeaveDuringAQuizAndAProbe(t *testing.T) {
	shortcuts := ansi.Strip(quizWithLessonFixture().sidebarShortcuts())
	if !strings.Contains(shortcuts, "Ctrl+L  Leave quiz") {
		t.Fatalf("a quiz can be left:\n%s", shortcuts)
	}
	probe := readingFixture()
	probe.application.state.Probe = &tutor.Probe{Topics: []string{"core"}, Asked: 1}
	if !strings.Contains(ansi.Strip(probe.sidebarShortcuts()), "Ctrl+L  Leave probe") {
		t.Fatal("a probe can be left too")
	}
	if strings.Contains(ansi.Strip(readingFixture().sidebarShortcuts()), "Ctrl+L") {
		t.Fatal("with nothing to leave, the sidebar offers nothing")
	}
}

func TestAnOpenQuestionWithoutAQuizStillGatesControlN(t *testing.T) {
	model := readingFixture()
	model.application.state.Question = "Which package owns the calculation?"
	model.application.state.Quiz, model.application.state.Probe = nil, nil
	next, _ := press(model, 'n', tea.ModCtrl)
	if !next.skipArmed || next.page != "lesson" || next.busy {
		t.Fatalf("an open question must ask for confirmation before Ctrl+N moves anywhere: armed=%v page=%s", next.skipArmed, next.page)
	}
	if !strings.Contains(ansi.Strip(model.sidebarShortcuts()), "Ctrl+N  Skip question") {
		t.Fatal("the sidebar should offer a skip for any open question")
	}
}

func TestTheLeavePromptNamesWhatIsBeingLeft(t *testing.T) {
	quiz := quizWithLessonFixture()
	if got := quiz.leavePrompt(); got != "Press Ctrl+L again to leave this quiz. It will not be scored." {
		t.Fatalf("got %q", got)
	}
	probe := readingFixture()
	probe.application.state.Quiz = nil
	probe.application.state.Probe = &tutor.Probe{Topics: []string{"core"}, Asked: 1}
	probe.application.state.Question = "Do you know core?"
	armed, _ := press(probe, 'l', tea.ModCtrl)
	if armed.notice != "Press Ctrl+L again to leave this probe." || armed.flowName() != "probe" {
		t.Fatalf("a probe should be named in its prompt, got %q", armed.notice)
	}
	left, _ := press(armed, 'l', tea.ModCtrl)
	if left.notice != "Left the probe" || left.application.state.Probe != nil {
		t.Fatalf("leaving a probe should say so and end it: %q", left.notice)
	}
}

func TestOneHelperArmsEverySecondPressPrompt(t *testing.T) {
	skipping, _ := openQuestionFixture().arm("Press X again.", false)
	armed := skipping.(interfaceModel)
	if !armed.skipArmed || armed.leaveArmed || armed.notice != "Press X again." || armed.armedPrompt != "Press X again." {
		t.Fatal("the helper should arm a skip with the prompt it was given")
	}
	cleared := armed.disarmSkip()
	if cleared.notice != "" || cleared.armedPrompt != "" || cleared.skipArmed {
		t.Fatal("disarming should clear exactly the prompt the helper showed")
	}
	leaving, _ := openQuestionFixture().arm("Press Y again.", true)
	if !leaving.(interfaceModel).leaveArmed || leaving.(interfaceModel).skipArmed {
		t.Fatal("the helper should arm a leave when asked")
	}
}

func TestLeavingAQuizShowsTheLessonEvenAfterHintsAndClarifications(t *testing.T) {
	model := quizWithLessonFixture()
	clarification := tutor.Reply{Explanation: "A hint about the question.", Question: "First?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: tutor.Overview, Action: "question", Message: "hint?", Reply: clarification})
	model.history = len(model.application.state.Turns) - 1
	armed, _ := press(model, 'l', tea.ModCtrl)
	left, _ := press(armed, 'l', tea.ModCtrl)
	state := left.application.state
	last := state.Turns[len(state.Turns)-1]
	if last.Action != "lesson" || state.Question != "" || state.Quiz != nil || reviewing(state) {
		t.Fatalf("leaving should land on the lesson, not on a hint that still carries the quiz question: %+v question=%q", last.Action, state.Question)
	}
	if left.activeTab() != "course" || left.actionLine() == "" || left.notice != "Left the quiz" {
		t.Fatalf("the course view and its actions should be back: tab=%s notice=%q", left.activeTab(), left.notice)
	}
}

func TestShowingATopicAgainNeverLeavesAQuestionPending(t *testing.T) {
	model := quizWithLessonFixture()
	model.application.state.Turns[0].Reply.Question = "A leftover lesson question?"
	again := showAgain(model.application.state, 0)
	if again.Question != "" || again.Review {
		t.Fatalf("a lesson shown again should leave nothing pending: %q", again.Question)
	}
}

func TestLeavingAProbeOnAnUnstudiedTopicReturnsToTheLastLesson(t *testing.T) {
	model := readingFixture()
	model.application.state.Quiz = nil
	model.application.state.Probe = &tutor.Probe{Topics: []string{"core"}, Asked: 1}
	model.application.state.Topic, model.application.state.Question = "core", "Do you know core?"
	left, found := leaveQuestions(model.application.state)
	if !found || left.Topic != tutor.Overview || left.Probe != nil || left.Question != "" {
		t.Fatalf("with no lesson on the probed topic, leaving should land on the last lesson read: found=%v topic=%s", found, left.Topic)
	}
	empty := model.application.state
	empty.Turns = nil
	if _, found := leaveQuestions(empty); found {
		t.Fatal("with no lesson anywhere there is nowhere to return to")
	}
}
