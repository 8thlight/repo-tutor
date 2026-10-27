package tutor

import "math"

func (quiz Quiz) Credit() int {
	credit := 0
	for _, question := range quiz.Questions {
		switch question.Assessment {
		case "correct":
			credit += 2
		case "partial":
			credit++
		}
	}
	return credit
}

func (quiz Quiz) Finished() bool {
	return quiz.Goal > 0 && quiz.Completed() >= quiz.Goal
}

func RecordQuizScore(previous map[string]Progress, topic string, quiz *Quiz) map[string]Progress {
	if quiz == nil || !quiz.Finished() {
		return previous
	}
	result := make(map[string]Progress, len(previous)+1)
	for name, progress := range previous {
		result[name] = progress
	}
	progress := result[topic]
	progress.QuizCredit, progress.QuizGoal = quiz.Credit(), quiz.Goal
	result[topic] = progress
	return result
}

func Understanding(progress Progress) (int, bool) {
	if progress.QuizGoal > 0 {
		return int(math.Round(float64(progress.QuizCredit) * 100 / float64(2*progress.QuizGoal))), true
	}
	if progress.Held {
		return 100, true
	}
	return 0, false
}

func Band(percent int) string {
	switch {
	case percent >= 80:
		return "solid"
	case percent >= weakBelow:
		return "developing"
	default:
		return "needs review"
	}
}

type CourseScore struct{ Read, Total, Percent int }

func ScoreCourse(course []Topic, progress map[string]Progress) CourseScore {
	score := CourseScore{Total: len(course)}
	sum := 0
	for _, topic := range course {
		current := progress[topic.ID]
		if current.Lessons == 0 {
			continue
		}
		score.Read++
		percent, _ := Understanding(current)
		sum += percent
	}
	if score.Read > 0 {
		score.Percent = int(math.Round(float64(sum) / float64(score.Read)))
	}
	return score
}
