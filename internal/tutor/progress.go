package tutor

type Progress struct {
	Lessons int  `json:"lessons"`
	Correct int  `json:"correct"`
	Partial int  `json:"partial"`
	Missed  int  `json:"missed"`
	Held    bool `json:"held,omitempty"`

	QuizCredit int `json:"quiz_credit,omitempty"`
	QuizGoal   int `json:"quiz_goal,omitempty"`
}

func Record(previous map[string]Progress, topic string, action string, reply Reply) map[string]Progress {
	result := make(map[string]Progress, len(previous)+1)
	for name, progress := range previous {
		result[name] = progress
	}
	progress := result[topic]
	if action == "lesson" {
		progress.Lessons++
	}
	switch reply.Assessment {
	case "correct":
		progress.Correct++
	case "partial":
		progress.Partial++
	case "incorrect":
		progress.Missed++
		progress.Held = false
	}
	result[topic] = progress
	return result
}
