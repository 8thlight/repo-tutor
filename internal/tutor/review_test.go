package tutor_test

import (
	"testing"

	"repo-tutor/internal/tutor"
)

func TestTheWeakestTopicHasTheLowestFinishedQuizUnderTheBar(t *testing.T) {
	course := []tutor.Topic{{ID: "a", Title: "A"}, {ID: "b", Title: "B"}, {ID: "c", Title: "C"}, {ID: "d", Title: "D"}}
	progress := map[string]tutor.Progress{
		"a": {Lessons: 1, QuizCredit: 10, QuizGoal: 5},
		"b": {Lessons: 1, QuizCredit: 4, QuizGoal: 5},
		"c": {Lessons: 1, QuizCredit: 2, QuizGoal: 5},
		"d": {Lessons: 1},
	}
	topic, found := tutor.WeakTopic(course, progress, nil)
	if !found || topic.ID != "c" {
		t.Fatalf("got %+v %v", topic, found)
	}
}

func TestNoTopicNeedsReviewWithoutAFinishedQuizUnderTheBar(t *testing.T) {
	course := []tutor.Topic{{ID: "a", Title: "A"}, {ID: "b", Title: "B"}}
	progress := map[string]tutor.Progress{"a": {Lessons: 1, Missed: 3}, "b": {Lessons: 1, QuizCredit: 6, QuizGoal: 5}}
	if _, found := tutor.WeakTopic(course, progress, nil); found {
		t.Fatal("answers outside a finished quiz, and a quiz at 60%, need no review")
	}
	if _, found := tutor.WeakTopic(course, nil, nil); found {
		t.Fatal("a course with no answers needs no review")
	}
}
