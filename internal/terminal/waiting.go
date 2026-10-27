package terminal

const sourcesWaitNotice = "Sources open again when the next lesson arrives."

func (model interfaceModel) movingToAnotherTopic() bool {
	return model.busy && model.intent.topic != model.application.state.Topic
}

func (model interfaceModel) browsingHint() string {
	if model.movingToAnotherTopic() {
		return "You can browse topics (Ctrl+K) while you wait."
	}
	return "You can browse topics (Ctrl+K) or this lesson's sources (Ctrl+S) while you wait."
}

func (model interfaceModel) runningNotice() string {
	if model.movingToAnotherTopic() {
		return "Something is already loading. Esc stops it; Topics (Ctrl+K) stay open."
	}
	return "Something is already loading. Esc stops it; Topics (Ctrl+K) and sources (Ctrl+S) stay open."
}

func finishedNotice(action string) string {
	switch action {
	case "quiz", "quiz-next", "quiz-extend", "probe":
		return "Question ready"
	case "answer":
		return "Answer reviewed"
	case "question":
		return "Reply ready"
	default:
		return "Lesson ready"
	}
}
