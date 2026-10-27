package tutor

import "testing"

func TestQuizGoalAndImmutableOutcomes(t *testing.T) {
	quiz := RecordQuiz(nil, "core", "quiz", "", Reply{Question: "First?"}, 5)
	if quiz.Goal != 5 || quiz.Completed() != 0 || len(quiz.Questions) != 1 {
		t.Fatal("quiz should begin with one current question and a five-question goal")
	}
	hint := RecordQuiz(quiz, "core", "answer", "First?", Reply{Assessment: "none", Question: "First?"}, 5)
	if hint.Completed() != 0 {
		t.Fatal("a hint counted as a completed question")
	}
	clarified := RecordQuiz(quiz, "core", "answer", "First?", Reply{Assessment: "none", Question: "Which package owns the calculation, based on the overview?"}, 5)
	if clarified.Completed() != 0 || len(clarified.Questions) != 1 || clarified.Questions[0].Text != "Which package owns the calculation, based on the overview?" || quiz.Questions[0].Text != "First?" {
		t.Fatal("clarification must revise the current question without completing it or mutating prior state")
	}
	clarified = RecordQuiz(clarified, "core", "answer", clarified.Questions[0].Text, Reply{Assessment: "correct"}, 5)
	if clarified.Completed() != 1 || clarified.Questions[0].Assessment != "correct" {
		t.Fatal("the revised question could not be answered")
	}
	graded := RecordQuiz(quiz, "core", "answer", "First?", Reply{Assessment: "incorrect"}, 5)
	if quiz.Completed() != 0 || graded.Completed() != 1 || len(graded.Questions) != 1 {
		t.Fatal("grading must record one outcome without generating the next question or mutating the prior quiz")
	}
	for index := 1; index < 5; index++ {
		graded = RecordQuiz(graded, "core", "quiz-next", "", Reply{Question: "Another?"}, 5)
		graded = RecordQuiz(graded, "core", "answer", "Another?", Reply{Assessment: "none"}, 5)
	}
	if graded.Completed() != 5 || graded.Goal != 5 || graded.Questions[4].Assessment != "skipped" {
		t.Fatal("skipped questions should count toward finishing, not correctness")
	}
	extended := RecordQuiz(graded, "core", "quiz-extend", "", Reply{Question: "Sixth?"}, 5)
	if extended.Goal != 10 || extended.Completed() != 5 || len(extended.Questions) != 6 || graded.Goal != 5 {
		t.Fatal("extension must preserve outcomes and add five to the goal")
	}
	if RecordQuiz(extended, "other", "lesson", "", Reply{}, 5) != nil {
		t.Fatal("a new topic must leave the quiz")
	}
}

func TestSkippedSlotIsReopenedAndReplacedWithoutDoubleCounting(t *testing.T) {
	quiz := RecordQuiz(nil, "core", "quiz", "", Reply{Question: "First?"}, 5)
	quiz = RecordQuiz(quiz, "core", "answer", "First?", Reply{Assessment: "none"}, 5)
	quiz = RecordQuiz(quiz, "core", "quiz-next", "", Reply{Question: "Second?"}, 5)
	if quiz.Active != 1 || quiz.Questions[0].Assessment != "skipped" {
		t.Fatal("the newest question should be active and the first should be skipped")
	}
	reopened, ok := Reopen(quiz, 0)
	if !ok || reopened.Active != 0 || quiz.Active != 1 {
		t.Fatal("reopening should select the skipped slot without mutating the saved quiz")
	}
	answered := RecordQuiz(reopened, "core", "answer", "First?", Reply{Assessment: "correct"}, 5)
	if answered.Questions[0].Assessment != "correct" || answered.Completed() != 1 || len(answered.Questions) != 2 {
		t.Fatal("answering a reopened slot should replace the skipped result once")
	}
	if _, ok := Reopen(answered, 0); ok {
		t.Fatal("an answered slot must not be reopened")
	}
	if _, ok := Reopen(answered, 5); ok {
		t.Fatal("a later slot has no question to reopen")
	}
}

func TestAQuizTakesItsGoalFromTheChosenLengthAndExtendsByIt(t *testing.T) {
	for _, length := range []int{3, 5, 10} {
		quiz := RecordQuiz(nil, "core", "quiz", "", Reply{Question: "First?"}, length)
		if quiz.Goal != length {
			t.Fatalf("goal %d, want %d", quiz.Goal, length)
		}
		extended := RecordQuiz(quiz, "core", "quiz-extend", "", Reply{Question: "More?"}, length)
		if extended.Goal != 2*length {
			t.Fatalf("extended goal %d, want %d", extended.Goal, 2*length)
		}
	}
	if RecordQuiz(nil, "core", "quiz", "", Reply{Question: "First?"}, 0).Goal != DefaultQuizLength {
		t.Fatal("a missing length should fall back to the default")
	}
}
