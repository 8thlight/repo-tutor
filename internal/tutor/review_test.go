package tutor_test

import (
	"testing"

	"repo-tutor/internal/tutor"
)

func TestTheWeakestTopicHasTheMostMissesAndPartialsOverCorrect(t *testing.T) {
	course := []tutor.Topic{{ID: "a", Title: "A"}, {ID: "b", Title: "B"}, {ID: "c", Title: "C"}}
	progress := map[string]tutor.Progress{
		"a": {Correct: 2, Partial: 1},
		"b": {Missed: 2, Partial: 1, Correct: 1},
		"c": {Missed: 1},
	}
	topic, found := tutor.WeakTopic(course, progress)
	if !found || topic.ID != "b" {
		t.Fatalf("got %+v %v", topic, found)
	}
}

func TestNoTopicNeedsReviewWhenCorrectAnswersKeepUp(t *testing.T) {
	course := []tutor.Topic{{ID: "a", Title: "A"}}
	if _, found := tutor.WeakTopic(course, map[string]tutor.Progress{"a": {Correct: 2, Missed: 1, Partial: 1}}); found {
		t.Fatal("a topic with as many correct answers as misses needs no review")
	}
	if _, found := tutor.WeakTopic(course, nil); found {
		t.Fatal("a topic with no answers needs no review")
	}
}
