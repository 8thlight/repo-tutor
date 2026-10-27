package tutor_test

import (
	"fmt"
	"testing"

	"repo-tutor/internal/tutor"
)

func courseOf(count int) []tutor.Topic {
	course := make([]tutor.Topic, count)
	for index := range course {
		course[index] = tutor.Topic{ID: fmt.Sprintf("t%02d", index), Title: fmt.Sprintf("Topic %d", index)}
	}
	return course
}

func readAll(course []tutor.Topic, count int) map[string]tutor.Progress {
	progress := map[string]tutor.Progress{}
	for _, topic := range course[:count] {
		progress[topic.ID] = tutor.Progress{Lessons: 1}
	}
	return progress
}

func TestAnExamNeedsTenTopicsRead(t *testing.T) {
	course := courseOf(14)
	if tutor.ExamAvailable(course, readAll(course, 9)) || tutor.ExamPlan(course, readAll(course, 9), 10) != nil {
		t.Fatal("nine topics read is not enough")
	}
	if !tutor.ExamAvailable(course, readAll(course, 10)) {
		t.Fatal("ten topics read should offer an exam")
	}
	files := readAll(course, 9)
	files[tutor.FileTopic("a.go")] = tutor.Progress{Lessons: 5}
	if tutor.ExamAvailable(course, files) {
		t.Fatal("a file topic is not a course topic and does not count")
	}
}

func TestTheLengthIsTheSettingOrTheTopicsReadWhicheverIsSmaller(t *testing.T) {
	course := courseOf(30)
	for _, test := range []struct{ read, setting, want int }{{10, 10, 10}, {15, 20, 15}, {25, 20, 20}, {30, 30, 30}, {12, 10, 10}} {
		if got := len(tutor.ExamPlan(course, readAll(course, test.read), test.setting)); got != test.want {
			t.Fatalf("%d read with length %d: got %d questions, want %d", test.read, test.setting, got, test.want)
		}
	}
	if got := len(tutor.ExamPlan(course, readAll(course, 15), 3)); got != tutor.ExamMinimum {
		t.Fatalf("an exam is never shorter than ten, got %d", got)
	}
}

func TestTheExamChoosesUnquizzedThenLowestScoresAndAsksInCourseOrder(t *testing.T) {
	course := courseOf(14)
	progress := readAll(course, 14)
	for index := range 14 {
		id := course[index].ID
		current := progress[id]
		switch {
		case index < 4:
		case index < 9:
			current.QuizGoal, current.QuizCredit = 5, 2*5
		default:
			current.QuizGoal, current.QuizCredit = 5, 10-index
		}
		progress[id] = current
	}
	plan := tutor.ExamPlan(course, progress, 10)
	if len(plan) != 10 {
		t.Fatalf("got %d", len(plan))
	}
	chosen := map[string]bool{}
	for _, id := range plan {
		chosen[id] = true
	}
	for index := range 4 {
		if !chosen[course[index].ID] {
			t.Fatalf("unquizzed topic %d should be on the exam", index)
		}
	}
	for index := 9; index < 14; index++ {
		if !chosen[course[index].ID] {
			t.Fatalf("a low quiz score (topic %d) should be on the exam", index)
		}
	}
	for index := 4; index < 9; index++ {
		if index != 4 && chosen[course[index].ID] {
			t.Fatalf("a perfect quiz score (topic %d) should be left out when there is room for weaker ones", index)
		}
	}
	for index := 1; index < len(plan); index++ {
		if plan[index-1] >= plan[index] {
			t.Fatalf("questions should be asked in course order: %v", plan)
		}
	}
}

func TestAnExamStartsWithTheFirstTopicsQuestionAndCarriesItsPlan(t *testing.T) {
	plan := []string{"a", "b", "c"}
	exam := tutor.StartExam(plan, tutor.Reply{Question: "First?"})
	if !exam.Exam || exam.Goal != 3 || exam.Questions[0].Topic != "a" || exam.Covers("zzz") != true || exam.NextTopic("x") != "b" {
		t.Fatalf("got %+v", exam)
	}
	plan[0] = "changed"
	if exam.Plan[0] != "a" {
		t.Fatal("the exam must keep its own copy of the plan")
	}
}

func TestExamQuestionsRecordTheirTopicAndTheQuizIsNotDroppedByTopicChanges(t *testing.T) {
	exam := tutor.StartExam([]string{"a", "b"}, tutor.Reply{Question: "A?"})
	graded := tutor.RecordQuiz(exam, "a", "answer", "A?", tutor.Reply{Assessment: "correct"}, 5)
	next := tutor.RecordQuiz(graded, "b", "quiz-next", "", tutor.Reply{Question: "B?"}, 5)
	if next == nil || len(next.Questions) != 2 || next.Questions[1].Topic != "b" || next.Active != 1 {
		t.Fatalf("a question on a different topic should join the exam: %+v", next)
	}
	if tutor.RecordQuiz(next, "b", "lesson", "", tutor.Reply{}, 5) != nil {
		t.Fatal("starting a lesson still ends the exam")
	}
}

func TestAFinishedExamIsScoredPerTopicAndNeverTouchesTopicQuizScores(t *testing.T) {
	exam := tutor.StartExam([]string{"a", "b", "c"}, tutor.Reply{Question: "A?"})
	exam.Questions[0].Assessment = "correct"
	exam.Questions = append(exam.Questions, tutor.QuizQuestion{Text: "B?", Topic: "b", Assessment: "partial"}, tutor.QuizQuestion{Text: "C?", Topic: "c", Assessment: "skipped"})
	result, finished := tutor.ResultOf(exam)
	if !finished || result.Credit != 3 || result.Goal != 3 || result.Percent() != 50 {
		t.Fatalf("got %+v %v", result, finished)
	}
	if result.Topics[0] != (tutor.ExamTopicResult{Topic: "a", Credit: 2}) || result.Topics[2].Credit != 0 {
		t.Fatalf("each topic keeps its own credit: %+v", result.Topics)
	}
	before := map[string]tutor.Progress{"a": {QuizCredit: 4, QuizGoal: 5}}
	after := tutor.RecordQuizScore(before, "a", exam)
	if after["a"].QuizCredit != 4 || after["a"].QuizGoal != 5 {
		t.Fatal("an exam must not change a topic's quiz score")
	}
	unfinished := tutor.StartExam([]string{"a", "b"}, tutor.Reply{Question: "A?"})
	if _, finished := tutor.ResultOf(unfinished); finished {
		t.Fatal("an unfinished exam has no result")
	}
}

func TestWeakTopicFallsBackToATopicTheExamMissed(t *testing.T) {
	course := courseOf(4)
	progress := map[string]tutor.Progress{"t00": {Lessons: 1, QuizCredit: 10, QuizGoal: 5}}
	exam := &tutor.ExamResult{Credit: 2, Goal: 3, Topics: []tutor.ExamTopicResult{{Topic: "t00", Credit: 2}, {Topic: "t02", Credit: 0}, {Topic: "t03", Credit: 0}}}
	topic, found := tutor.WeakTopic(course, progress, exam)
	if !found || topic.ID != "t02" {
		t.Fatalf("an exam miss should be offered for review when no quiz is weak, got %+v %v", topic, found)
	}
	progress["t01"] = tutor.Progress{Lessons: 1, QuizCredit: 2, QuizGoal: 5}
	if topic, _ := tutor.WeakTopic(course, progress, exam); topic.ID != "t01" {
		t.Fatalf("a weak quiz score still comes first, got %+v", topic)
	}
	if _, found := tutor.WeakTopic(course, map[string]tutor.Progress{}, &tutor.ExamResult{Goal: 1, Topics: []tutor.ExamTopicResult{{Topic: "t01", Credit: 2}}}); found {
		t.Fatal("a passed exam topic needs no review")
	}
}

func TestAnUncappedExamAsksOneQuestionForEveryTopicRead(t *testing.T) {
	course := courseOf(40)
	progress := readAll(course, 35)
	plan := tutor.ExamPlan(course, progress, 0)
	if len(plan) != 35 {
		t.Fatalf("no cap means every topic read, got %d", len(plan))
	}
	if short := tutor.ExamPlan(course, readAll(course, 9), 0); short != nil {
		t.Fatal("even uncapped, an exam still needs ten topics read")
	}
}

func examWith(completed, goal int) *tutor.Quiz {
	plan := make([]string, goal)
	for index := range plan {
		plan[index] = fmt.Sprintf("t%02d", index)
	}
	exam := tutor.StartExam(plan, tutor.Reply{Question: "Q?"})
	for index := 0; index < completed; index++ {
		if index >= len(exam.Questions) {
			exam.Questions = append(exam.Questions, tutor.QuizQuestion{Text: "Q?", Topic: plan[index]})
		}
		exam.Questions[index].Assessment = []string{"correct", "partial", "incorrect", "skipped"}[index%4]
	}
	if completed < goal && completed >= len(exam.Questions) {
		exam.Questions = append(exam.Questions, tutor.QuizQuestion{Text: "Open?", Topic: plan[completed]})
	}
	return exam
}

func TestAnExamLeftAfterTenQuestionsKeepsAPartialResultScoredOverThoseQuestions(t *testing.T) {
	exam := examWith(12, 30)
	result, saved := tutor.PartialResult(exam)
	if !saved || result.Goal != 12 || result.Planned != 30 || len(result.Topics) != 12 {
		t.Fatalf("a twelve-question partial should be kept: %+v %v", result, saved)
	}
	if result.Credit != 3*(2+1) || result.Percent() != 38 {
		t.Fatalf("three of each of correct and partial earn 9 half points over 12 questions: %+v %d", result, result.Percent())
	}
	for _, topic := range result.Topics {
		if topic.Topic == "t12" {
			t.Fatal("the open, unanswered question is not part of the result")
		}
	}
}

func TestAnExamLeftBeforeTenQuestionsKeepsNothing(t *testing.T) {
	if _, saved := tutor.PartialResult(examWith(9, 30)); saved {
		t.Fatal("nine questions is under the threshold")
	}
	if _, saved := tutor.PartialResult(examWith(10, 30)); !saved {
		t.Fatal("exactly ten questions meets it")
	}
}

func TestAFinishedExamIsNotMarkedPartialAndAPartialIsNotAFinishedOne(t *testing.T) {
	full := examWith(10, 10)
	result, finished := tutor.ResultOf(full)
	if !finished || result.Planned != 0 || result.Goal != 10 {
		t.Fatalf("a finished exam has no planned count: %+v", result)
	}
	if _, saved := tutor.PartialResult(full); saved {
		t.Fatal("a finished exam is a result, not a partial one")
	}
	if _, finished := tutor.ResultOf(examWith(12, 30)); finished {
		t.Fatal("an unfinished exam has no finished result")
	}
}
