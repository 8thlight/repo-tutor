package tutor

const ReviewMessage = "Review this topic with a different example, focusing on earlier confusion."

func WeakTopic(course []Topic, progress map[string]Progress) (Topic, bool) {
	var selected Topic
	highest := 0
	for _, topic := range course {
		current := progress[topic.ID]
		if score := current.Missed + current.Partial - current.Correct; score > highest {
			selected, highest = topic, score
		}
	}
	return selected, highest > 0
}
