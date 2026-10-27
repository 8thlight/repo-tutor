package tutor

const ReviewMessage = "Review this topic with a different example, focusing on earlier confusion."

const weakBelow = 60

func WeakTopic(course []Topic, progress map[string]Progress) (Topic, bool) {
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
	return selected, found
}
