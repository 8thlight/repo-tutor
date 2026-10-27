package tutor

func SkipQuizQuestion(quiz *Quiz) (*Quiz, bool) {
	if quiz == nil || quiz.Active < 0 || quiz.Active >= len(quiz.Questions) {
		return quiz, false
	}
	if assessment := quiz.Questions[quiz.Active].Assessment; assessment != "" && assessment != "skipped" {
		return quiz, false
	}
	skipped := *quiz
	skipped.Questions = append([]QuizQuestion(nil), quiz.Questions...)
	skipped.Questions[skipped.Active].Assessment = "skipped"
	return &skipped, true
}

func SkipProbeQuestion(probe *Probe, progress map[string]Progress, topic string) (*Probe, map[string]Progress, bool) {
	if probe == nil || probe.Asked == 0 || probe.Answered >= probe.Asked {
		return probe, progress, false
	}
	next := *probe
	next.Answered++
	result := make(map[string]Progress, len(progress)+1)
	for name, value := range progress {
		result[name] = value
	}
	current := result[topic]
	current.Held = false
	result[topic] = current
	return &next, result, true
}
