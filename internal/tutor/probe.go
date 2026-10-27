package tutor

const probeLength = 3

type Probe struct {
	Topics   []string `json:"topics"`
	Asked    int      `json:"asked"`
	Answered int      `json:"answered"`
}

func ProbeTopics(course []Topic, progress map[string]Progress) []string {
	var topics []string
	for _, topic := range course {
		if topic.ID != Overview && !progress[topic.ID].Held && len(topics) < probeLength {
			topics = append(topics, topic.ID)
		}
	}
	return topics
}

func (probe Probe) Current() (string, bool) {
	if probe.Asked == 0 || probe.Asked > len(probe.Topics) {
		return "", false
	}
	return probe.Topics[probe.Asked-1], true
}

func (probe Probe) Next() (string, bool) {
	if probe.Asked >= len(probe.Topics) {
		return "", false
	}
	return probe.Topics[probe.Asked], true
}

func FirstUnheld(course []Topic, progress map[string]Progress) (Topic, bool) {
	for _, topic := range course {
		if topic.ID != Overview && !progress[topic.ID].Held {
			return topic, true
		}
	}
	return Topic{}, false
}

func NextUnheld(course []Topic, progress map[string]Progress, current string) (Topic, bool) {
	found := false
	for _, topic := range course {
		if found && !progress[topic.ID].Held {
			return topic, true
		}
		found = found || topic.ID == current
	}
	if found {
		return Topic{}, false
	}
	for _, topic := range course {
		if progress[topic.ID].Lessons == 0 && !progress[topic.ID].Held {
			return topic, true
		}
	}
	return Topic{}, false
}

func RecordProbe(previous *Probe, course []Topic, progress map[string]Progress, action, topic string, reply Reply) (*Probe, map[string]Progress) {
	switch {
	case action == "probe":
		if previous != nil {
			if following, more := previous.Next(); more && following == topic {
				next := *previous
				next.Asked++
				return &next, progress
			}
		}
		return &Probe{Topics: ProbeTopics(course, progress), Asked: 1}, progress
	case previous == nil || action == "lesson" || action == "quiz" || action == "quiz-next" || action == "quiz-extend":
		return nil, progress
	}
	current, asking := previous.Current()
	if action != "answer" || !asking || current != topic || previous.Answered >= previous.Asked || (reply.Assessment == "none" && reply.Question != "") {
		return previous, progress
	}
	next := *previous
	next.Answered++
	result := make(map[string]Progress, len(progress)+1)
	for name, value := range progress {
		result[name] = value
	}
	held := result[topic]
	held.Held = reply.Assessment == "correct"
	result[topic] = held
	return &next, result
}
