package tutor_test

import (
	"testing"

	"repo-tutor/internal/tutor"
)

func finishedQuiz(assessments ...string) *tutor.Quiz {
	quiz := &tutor.Quiz{Topic: "core", Goal: len(assessments)}
	for _, assessment := range assessments {
		quiz.Questions = append(quiz.Questions, tutor.QuizQuestion{Text: "Q?", Assessment: assessment})
	}
	return quiz
}

func TestACorrectAnswerEarnsOnePartialHalfAndTheRestNothing(t *testing.T) {
	scores := tutor.RecordQuizScore(nil, "core", finishedQuiz("correct", "partial", "incorrect", "skipped", "correct"))
	percent, scored := tutor.Understanding(scores["core"])
	if !scored || percent != 50 || scores["core"].QuizGoal != 5 || scores["core"].QuizCredit != 5 {
		t.Fatalf("got %d %v %+v", percent, scored, scores["core"])
	}
}

func TestAnUnfinishedQuizLeavesTheStoredScoreAlone(t *testing.T) {
	stored := map[string]tutor.Progress{"core": {QuizCredit: 8, QuizGoal: 5}}
	unfinished := &tutor.Quiz{Topic: "core", Goal: 5, Questions: []tutor.QuizQuestion{{Text: "Q?", Assessment: "incorrect"}}}
	scores := tutor.RecordQuizScore(stored, "core", unfinished)
	if percent, _ := tutor.Understanding(scores["core"]); percent != 80 {
		t.Fatalf("a retake in progress must not lower the score, got %d", percent)
	}
}

func TestAFinishedRetakeReplacesTheEarlierScoreAndKeepsTheOriginalUntouched(t *testing.T) {
	stored := map[string]tutor.Progress{"core": {QuizCredit: 2, QuizGoal: 5}}
	replaced := tutor.RecordQuizScore(stored, "core", finishedQuiz("correct", "correct", "correct", "correct", "partial"))
	if percent, _ := tutor.Understanding(replaced["core"]); percent != 90 {
		t.Fatalf("got %d", percent)
	}
	if stored["core"].QuizCredit != 2 {
		t.Fatal("recording a score must not change the previous map")
	}
}

func TestAKnownTopicWithNoQuizCountsAsFullUntilAQuizReplacesIt(t *testing.T) {
	if percent, scored := tutor.Understanding(tutor.Progress{Held: true}); !scored || percent != 100 {
		t.Fatalf("got %d %v", percent, scored)
	}
	if _, scored := tutor.Understanding(tutor.Progress{Lessons: 3}); scored {
		t.Fatal("reading alone gives no score")
	}
	if percent, _ := tutor.Understanding(tutor.Progress{Held: true, QuizCredit: 3, QuizGoal: 5}); percent != 30 {
		t.Fatalf("a quiz should replace the probe's mark, got %d", percent)
	}
}

func TestTheBandsSitAtEightyAndSixty(t *testing.T) {
	for percent, want := range map[int]string{100: "solid", 80: "solid", 79: "developing", 60: "developing", 59: "needs review", 0: "needs review"} {
		if got := tutor.Band(percent); got != want {
			t.Fatalf("%d%%: got %s want %s", percent, got, want)
		}
	}
}

func TestTheCourseFiguresAreCoverageAndTheMeanOverTopicsRead(t *testing.T) {
	course := []tutor.Topic{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
	progress := map[string]tutor.Progress{
		"a": {Lessons: 1, QuizCredit: 10, QuizGoal: 5},
		"b": {Lessons: 2},
		"c": {Lessons: 1, QuizCredit: 5, QuizGoal: 5},
	}
	score := tutor.ScoreCourse(course, progress)
	if score.Read != 3 || score.Total != 4 || score.Percent != 50 {
		t.Fatalf("got %+v", score)
	}
	if empty := tutor.ScoreCourse(course, nil); empty.Read != 0 || empty.Percent != 0 {
		t.Fatalf("a course nothing has been read in scores zero: %+v", empty)
	}
}
