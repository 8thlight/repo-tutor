package tutor

type Progress struct {
	Lessons int `json:"lessons"`
	Correct int `json:"correct"`
	Partial int `json:"partial"`
	Missed  int `json:"missed"`
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
	}
	result[topic] = progress
	return result
}
