package tutor

import "testing"

func TestQuizGoalAndImmutableOutcomes(t *testing.T) {
	quiz := RecordQuiz(nil, "core", "quiz", "", Reply{Question: "First?"})
	if quiz.Goal != 5 || quiz.Completed() != 0 || len(quiz.Questions) != 1 {
		t.Fatal("quiz should begin with one current question and a five-question goal")
	}
	hint := RecordQuiz(quiz, "core", "answer", "First?", Reply{Assessment: "none", Question: "First?"})
	if hint.Completed() != 0 {
		t.Fatal("a hint counted as a completed question")
	}
	graded := RecordQuiz(quiz, "core", "answer", "First?", Reply{Assessment: "incorrect"})
	if quiz.Completed() != 0 || graded.Completed() != 1 || len(graded.Questions) != 1 {
		t.Fatal("grading must record one outcome without generating the next question or mutating the prior quiz")
	}
	for index := 1; index < 5; index++ {
		graded = RecordQuiz(graded, "core", "quiz-next", "", Reply{Question: "Another?"})
		graded = RecordQuiz(graded, "core", "answer", "Another?", Reply{Assessment: "none"})
	}
	if graded.Completed() != 5 || graded.Goal != 5 || graded.Questions[4].Assessment != "skipped" {
		t.Fatal("skipped questions should count toward finishing, not correctness")
	}
	extended := RecordQuiz(graded, "core", "quiz-extend", "", Reply{Question: "Sixth?"})
	if extended.Goal != 10 || extended.Completed() != 5 || len(extended.Questions) != 6 || graded.Goal != 5 {
		t.Fatal("extension must preserve outcomes and add five to the goal")
	}
	if RecordQuiz(extended, "other", "lesson", "", Reply{}) != nil {
		t.Fatal("a new topic must leave the quiz")
	}
}
