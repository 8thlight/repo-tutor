package tutor

import (
	"math"
	"sort"
)

const ExamMinimum = 10

var ExamLengths = []int{10, 20, 30}

type ExamTopicResult struct {
	Topic  string `json:"topic"`
	Credit int    `json:"credit"`
}

type ExamResult struct {
	Credit int               `json:"credit"`
	Goal   int               `json:"goal"`
	Topics []ExamTopicResult `json:"topics"`
}

func (result ExamResult) Percent() int {
	if result.Goal == 0 {
		return 0
	}
	return int(math.Round(float64(result.Credit) * 100 / float64(2*result.Goal)))
}

func ExamEligible(course []Topic, progress map[string]Progress) []Topic {
	var eligible []Topic
	for _, topic := range course {
		if progress[topic.ID].Lessons > 0 {
			eligible = append(eligible, topic)
		}
	}
	return eligible
}

func ExamAvailable(course []Topic, progress map[string]Progress) bool {
	return len(ExamEligible(course, progress)) >= ExamMinimum
}

func ExamPlan(course []Topic, progress map[string]Progress, length int) []string {
	if !ExamAvailable(course, progress) {
		return nil
	}
	eligible := ExamEligible(course, progress)
	order := map[string]int{}
	for index, topic := range course {
		order[topic.ID] = index
	}
	chosen := append([]Topic(nil), eligible...)
	sort.SliceStable(chosen, func(first, second int) bool {
		left, right := progress[chosen[first].ID], progress[chosen[second].ID]
		leftQuizzed, rightQuizzed := left.QuizGoal > 0, right.QuizGoal > 0
		if leftQuizzed != rightQuizzed {
			return !leftQuizzed
		}
		if leftQuizzed {
			leftPercent, _ := Understanding(left)
			rightPercent, _ := Understanding(right)
			if leftPercent != rightPercent {
				return leftPercent < rightPercent
			}
		}
		return order[chosen[first].ID] < order[chosen[second].ID]
	})
	chosen = chosen[:min(max(length, ExamMinimum), len(chosen))]
	sort.SliceStable(chosen, func(first, second int) bool { return order[chosen[first].ID] < order[chosen[second].ID] })
	plan := make([]string, len(chosen))
	for index, topic := range chosen {
		plan[index] = topic.ID
	}
	return plan
}

func StartExam(plan []string, first Reply) *Quiz {
	return &Quiz{Exam: true, Goal: len(plan), Plan: append([]string(nil), plan...), Questions: []QuizQuestion{{Text: first.Question, Topic: plan[0]}}}
}

func ResultOf(quiz *Quiz) (ExamResult, bool) {
	if quiz == nil || !quiz.Exam || !quiz.Finished() {
		return ExamResult{}, false
	}
	result := ExamResult{Credit: quiz.Credit(), Goal: quiz.Goal}
	for _, question := range quiz.Questions {
		earned := 0
		switch question.Assessment {
		case "correct":
			earned = 2
		case "partial":
			earned = 1
		}
		result.Topics = append(result.Topics, ExamTopicResult{question.Topic, earned})
	}
	return result, true
}
