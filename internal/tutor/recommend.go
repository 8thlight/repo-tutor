package tutor

import "fmt"

const (
	PlaceContinue = "continue"
	PlaceQuiz     = "quiz"
	PlaceLearn    = "learn"
)

type Recommendation struct {
	Place  string
	Reason string
}

func Recommend(course []Topic, progress map[string]Progress, topic string, quiz *Quiz, reviewing bool) Recommendation {
	if reviewing {
		return Recommendation{PlaceContinue, "You have an answer to review."}
	}
	if quiz != nil && quiz.Completed() < quiz.Goal {
		return Recommendation{PlaceQuiz, fmt.Sprintf("Your quiz has %d of %d answered.", quiz.Completed(), quiz.Goal)}
	}
	if !Started(progress) {
		return Recommendation{PlaceLearn, "You have not started."}
	}
	next, found := NextUnheld(course, progress, topic)
	if !found {
		return Recommendation{PlaceLearn, "You covered every topic. Review one."}
	}
	return Recommendation{PlaceLearn, "Next topic: " + Plain(next.Title) + "."}
}

func Started(progress map[string]Progress) bool {
	for topic, outcome := range progress {
		attempts := outcome.Correct + outcome.Partial + outcome.Missed
		if attempts > 0 || (topic != Overview && outcome.Lessons > 0) {
			return true
		}
	}
	return false
}
