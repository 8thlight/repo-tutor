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
		return "A turn is already running. Esc stops it; Topics (Ctrl+K) stay open."
	}
	return "A turn is already running. Esc stops it; Topics (Ctrl+K) and sources (Ctrl+S) stay open."
}
