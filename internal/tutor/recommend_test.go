package tutor

import (
	"strings"
	"testing"
)

func TestRecommendationFollowsTheLearnersState(t *testing.T) {
	course := []Topic{{Overview, "Overview"}, {"a", "Core"}, {"b", "CLI"}}
	fresh := map[string]Progress{}
	opened := map[string]Progress{Overview: {Lessons: 1}}
	studied := map[string]Progress{Overview: {Lessons: 1}, "a": {Lessons: 1, Correct: 1}}
	done := map[string]Progress{Overview: {Lessons: 1}, "a": {Lessons: 1, Held: true}, "b": {Lessons: 1}}
	running := &Quiz{Goal: 5, Questions: []QuizQuestion{{Text: "One?", Assessment: "correct"}}}
	finished := &Quiz{Goal: 1, Questions: []QuizQuestion{{Text: "One?", Assessment: "correct"}}}
	for name, test := range map[string]struct {
		progress  map[string]Progress
		topic     string
		quiz      *Quiz
		reviewing bool
		place     string
		reason    string
	}{
		"no turns":          {fresh, Overview, nil, false, PlaceLearn, "not started"},
		"only the overview": {opened, Overview, nil, false, PlaceLearn, "not started"},
		"review first":      {studied, "a", running, true, PlaceContinue, "answer to review"},
		"quiz in progress":  {studied, "a", running, false, PlaceQuiz, "1 of 5"},
		"quiz finished":     {studied, "a", finished, false, PlaceLearn, "Next topic: CLI"},
		"next topic":        {studied, "a", nil, false, PlaceLearn, "Next topic: CLI"},
		"every topic":       {done, "b", nil, false, PlaceLearn, "every topic"},
	} {
		got := Recommend(course, test.progress, test.topic, test.quiz, test.reviewing)
		if got.Place != test.place || !strings.Contains(strings.ToLower(got.Reason), strings.ToLower(test.reason)) {
			t.Fatalf("%s: got %+v", name, got)
		}
	}
}
