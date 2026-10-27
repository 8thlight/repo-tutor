package terminal

import (
	"context"
	"testing"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func twoTopicFixture() interfaceModel {
	model := interfaceFixture()
	core := tutor.Reply{Explanation: "Core owns the calculation.", Question: "Where is the calculation?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: "core", Action: "lesson", Reply: core})
	model.application.state.Topic, model.application.state.Question = "core", core.Question
	model.application.state.Progress = map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, "core": {Lessons: 1}}
	model.history = 1
	return model
}

func TestPickingAStudiedTopicResumesItWithoutARequest(t *testing.T) {
	model := twoTopicFixture()
	calls := 0
	model.run = func(_ context.Context, state session.State, _, _, _ string) (session.State, error) {
		calls++
		return state, nil
	}
	before := model.application.state.Progress
	back, command := model.openTopic(tutor.Overview)
	model = back.(interfaceModel)
	state := model.application.state
	if command != nil && model.busy || calls != 0 || model.busy {
		t.Fatal("resuming a studied topic must not call the agent")
	}
	if state.Topic != tutor.Overview || state.Question != "Which package would you inspect to change a calculation?" || len(state.Turns) != 3 || model.history != 2 || model.page != "lesson" {
		t.Fatalf("overview should be current with its question restored: %+v", state)
	}
	if state.Progress[tutor.Overview] != before[tutor.Overview] || state.Progress["core"] != before["core"] {
		t.Fatal("a resume must not change progress")
	}
	if !state.Turns[2].Resumed || state.Turns[1].Resumed {
		t.Fatal("only the appended copy should be marked resumed")
	}
	again, _ := model.openTopic("core")
	state = again.(interfaceModel).application.state
	if state.Topic != "core" || state.Question != "Where is the calculation?" || len(state.Turns) != 4 {
		t.Fatalf("returning to core should restore its question: %+v", state)
	}
	same, _ := again.(interfaceModel).openTopic("core")
	if len(same.(interfaceModel).application.state.Turns) != 4 {
		t.Fatal("picking the current topic should add no turn")
	}
}

func TestResumingAnAnsweredTopicRestoresReviewAndEndsOtherQuizzes(t *testing.T) {
	model := twoTopicFixture()
	graded := tutor.Reply{Explanation: "Right.", Feedback: "Correct.", Assessment: "correct", Citations: model.application.state.Turns[0].Reply.Citations}
	model.application.state.Turns = append(model.application.state.Turns[:1], session.Turn{Topic: tutor.Overview, Action: "answer", Message: "main", Reply: graded}, model.application.state.Turns[1])
	model.application.state.Quiz = &tutor.Quiz{Topic: "core", Goal: 5, Questions: []tutor.QuizQuestion{{Text: "Where?"}}}
	model.application.state.Probe = &tutor.Probe{Topics: []string{"core"}}
	back, _ := model.openTopic(tutor.Overview)
	state := back.(interfaceModel).application.state
	if !state.Review || state.Question != "" || state.Quiz != nil || state.Probe != nil || !reviewing(state) {
		t.Fatalf("a graded topic should resume at its review and leave other quizzes: %+v", state)
	}
}

func TestPickingAnUnstudiedTopicStillRequestsItsFirstLesson(t *testing.T) {
	model := twoTopicFixture()
	started, command := model.openTopic("cli")
	next := started.(interfaceModel)
	if command == nil || !next.busy || next.intent.topic != "cli" || next.intent.action != "lesson" {
		t.Fatalf("a new topic should request a lesson: %+v", next.intent)
	}
	next.cancel()
}

func TestNextTopicOpensAFreshCachedLessonWithoutARequestAndIgnoresAStaleOne(t *testing.T) {
	for _, fingerprint := range []string{"current", "older"} {
		model := twoTopicFixture()
		store, err := session.Open(t.TempDir(), "/work/repowise")
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()
		model.application.store = store
		model.application.snapshot.Fingerprint = "current"
		reply := tutor.Reply{Explanation: "The CLI parses flags.", Question: "Which flag sets the goal?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations}
		lesson := tutor.SavedLesson{Fingerprint: fingerprint, Goal: tutor.GoalKey(model.application.state.Goal), Topic: "cli", Reply: reply}
		if err := store.SaveLesson(tutor.LessonKey(model.application.state.Goal, "cli"), tutor.EncodeLesson(lesson)); err != nil {
			t.Fatal(err)
		}
		opened, command := model.openTopic("cli")
		next := opened.(interfaceModel)
		if fingerprint == "older" {
			if command == nil || !next.busy {
				t.Fatal("a lesson for changed source must be requested again")
			}
			next.cancel()
			continue
		}
		state := next.application.state
		if next.busy || state.Topic != "cli" || state.Question != "" || state.Progress["cli"].Lessons != 1 || len(state.Turns) != 3 {
			t.Fatalf("a fresh cached lesson should open with no request: busy=%v %+v", next.busy, state)
		}
	}
}

func TestTopicsAndQuizListsOpenOnTheCurrentItem(t *testing.T) {
	model := twoTopicFixture()
	model.application.state.Topic = "learning/core"
	opened, _ := model.openPicker("topics", "")
	topics := opened.(interfaceModel)
	if item := topics.picker.SelectedItem().(browserItem); item.value != "learning/core" {
		t.Fatalf("the topics list should open on the current topic, got %q", item.value)
	}
	model.application.state.Quiz = &tutor.Quiz{Topic: "learning/core", Goal: 5, Active: 2, Questions: []tutor.QuizQuestion{{Text: "a", Assessment: "skipped"}, {Text: "b", Assessment: "correct"}, {Text: "c"}}}
	opened, _ = model.openPicker("quiz", "")
	if item := opened.(interfaceModel).picker.SelectedItem().(browserItem); item.line != 2 {
		t.Fatalf("the quiz index should open on the active question, got slot %d", item.line)
	}
	opened, _ = model.openPicker("topics", "core")
	if _, ok := opened.(interfaceModel).picker.SelectedItem().(browserItem); !ok {
		t.Fatal("a filtered list should still select its first match")
	}
}
