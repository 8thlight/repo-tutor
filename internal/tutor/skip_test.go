package tutor_test

import (
	"testing"

	"repo-tutor/internal/tutor"
)

func TestSkippingMarksTheOpenQuizQuestionSkippedAndLeavesTheOriginalAlone(t *testing.T) {
	quiz := &tutor.Quiz{Topic: "core", Goal: 3, Active: 1, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}, {Text: "B?"}}}
	skipped, ok := tutor.SkipQuizQuestion(quiz)
	if !ok || skipped.Questions[1].Assessment != "skipped" || skipped.Completed() != 2 {
		t.Fatalf("got %+v %v", skipped, ok)
	}
	if quiz.Questions[1].Assessment != "" {
		t.Fatal("skipping must not change the previous quiz")
	}
}

func TestAnAnsweredQuestionCannotBeSkippedButAReopenedSkipCan(t *testing.T) {
	answered := &tutor.Quiz{Topic: "core", Goal: 3, Active: 0, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "correct"}}}
	if _, ok := tutor.SkipQuizQuestion(answered); ok {
		t.Fatal("an answered question is not open")
	}
	reopened := &tutor.Quiz{Topic: "core", Goal: 3, Active: 0, Questions: []tutor.QuizQuestion{{Text: "A?", Assessment: "skipped"}}}
	if _, ok := tutor.SkipQuizQuestion(reopened); !ok {
		t.Fatal("a reopened skipped question can be skipped again")
	}
	if _, ok := tutor.SkipQuizQuestion(nil); ok {
		t.Fatal("no quiz, nothing to skip")
	}
}

func TestSkippingAProbeQuestionCountsItAnsweredAndClearsTheKnownMark(t *testing.T) {
	probe := &tutor.Probe{Topics: []string{"core", "cli"}, Asked: 1, Answered: 0}
	progress := map[string]tutor.Progress{"core": {Held: true, Lessons: 1}}
	next, scores, ok := tutor.SkipProbeQuestion(probe, progress, "core")
	if !ok || next.Answered != 1 || scores["core"].Held || scores["core"].Lessons != 1 {
		t.Fatalf("got %+v %+v %v", next, scores, ok)
	}
	if probe.Answered != 0 || !progress["core"].Held {
		t.Fatal("skipping must not change the previous probe or progress")
	}
	if _, _, ok := tutor.SkipProbeQuestion(&tutor.Probe{Topics: []string{"core"}, Asked: 1, Answered: 1}, progress, "core"); ok {
		t.Fatal("an answered probe question is not open")
	}
}
