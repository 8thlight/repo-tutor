package tutor

const ReviewMessage = "Review this topic with a different example, focusing on earlier confusion."

const weakBelow = 60

func WeakTopic(course []Topic, progress map[string]Progress, exam *ExamResult) (Topic, bool) {
	var selected Topic
	lowest, found := 100, false
	for _, topic := range course {
		current := progress[topic.ID]
		if current.QuizGoal == 0 {
			continue
		}
		if percent, _ := Understanding(current); percent < weakBelow && percent < lowest {
			selected, lowest, found = topic, percent, true
		}
	}
	if found || exam == nil {
		return selected, found
	}
	missed := map[string]bool{}
	for _, result := range exam.Topics {
		if result.Credit == 0 {
			missed[result.Topic] = true
		}
	}
	for _, topic := range course {
		if missed[topic.ID] {
			return topic, true
		}
	}
	return Topic{}, false
}
