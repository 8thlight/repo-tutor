package terminal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func historyState() session.State {
	answer := tutor.Reply{Explanation: "Right.", Feedback: "Correct.", Assessment: "correct"}
	lesson := tutor.Reply{Explanation: "A lesson.", Question: "Why?", Assessment: "none"}
	return session.State{Topic: tutor.Overview, Review: true, Turns: []session.Turn{
		{Topic: tutor.Overview, Action: "lesson", Reply: lesson},
		{Topic: tutor.Overview, Action: "answer", Reply: answer},
		{Topic: "learning", Action: "lesson", Reply: lesson},
		{Topic: "learning", Action: "answer", Reply: answer},
		{Topic: tutor.Overview, Action: "answer", Reply: answer, Resumed: true},
	}}
}

func TestStepsAreCountedPerTopicAndSkipResumedCopies(t *testing.T) {
	state := historyState()
	if got, _ := positionOf(state.Turns, 4); got.topic != tutor.Overview || got.step != 2 || got.steps != 2 {
		t.Fatalf("a resumed copy should read as its topic's last step: %+v", got)
	}
	if got, _ := positionOf(state.Turns, 2); got.topic != "learning" || got.step != 1 || got.steps != 2 {
		t.Fatalf("learning step 1 of 2 expected: %+v", got)
	}
	course := []string{tutor.Overview, "learning", "learning/core"}
	for name, test := range map[string]struct {
		from, direction, target int
		crossing, moved         bool
	}{
		"back within a topic":           {4, -1, 0, false, true},
		"forward lands on the live one": {0, 1, 4, false, true},
		"before the first of all":       {0, -1, 0, false, false},
		"forward over a topic edge":     {4, 1, 2, true, true},
		"back over a topic edge":        {2, -1, 4, true, true},
		"after the last of all":         {3, 1, 3, false, false},
	} {
		got := stepHistory(course, state, test.from, test.direction)
		if got.moved != test.moved || got.crossing != test.crossing || (got.moved && got.target != test.target) {
			t.Fatalf("%s: got %+v", name, got)
		}
	}
}

func TestStepKeysNeedASecondPressToCrossATopicAndTitleShowsPosition(t *testing.T) {
	model := interfaceFixture()
	model.application.state = historyState()
	model.application.course = []tutor.Topic{{ID: tutor.Overview, Title: "Repository overview"}, {ID: "learning", Title: ". (main)"}}
	model.history = 4
	model.rebuildContent()
	if title := ansi.Strip(model.View().Content); !strings.Contains(title, "Repository overview  ·  Topic 1 / 2  ·  Step 2 / 2") || strings.Contains(title, "turn ") {
		t.Fatalf("the title should give topic and step, not a turn count:\n%s", title)
	}
	first, _ := model.stepTurns(1)
	held := first.(interfaceModel)
	if held.history != 4 || !strings.Contains(held.notice, "press again for the next topic") {
		t.Fatalf("the first press at an edge should only explain: %d %q", held.history, held.notice)
	}
	second, _ := held.stepTurns(1)
	crossed := second.(interfaceModel)
	if crossed.history != 2 || !strings.Contains(crossed.notice, "Next topic: . (main)") {
		t.Fatalf("the second press should cross to the next topic: %d %q", crossed.history, crossed.notice)
	}
	if !strings.Contains(ansi.Strip(crossed.View().Content), ". (main)  ·  Topic 2 / 2  ·  Step 1 / 2") {
		t.Fatal("the title should follow into the new topic")
	}
	back, _ := crossed.stepTurns(1)
	inside := back.(interfaceModel)
	if inside.history != 3 || inside.edgeBump != "" {
		t.Fatal("a normal step inside a topic should clear any pending crossing")
	}
}
