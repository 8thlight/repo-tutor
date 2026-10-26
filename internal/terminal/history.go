package terminal

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

type historyPosition struct {
	topic string
	step  int
	steps int
}

func realTurns(turns []session.Turn, topic string) []int {
	var indexes []int
	for index, turn := range turns {
		if turn.Topic == topic && !turn.Resumed && turn.Action != "quiz-retry" {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func positionOf(turns []session.Turn, viewed int) (historyPosition, bool) {
	if viewed < 0 || viewed >= len(turns) {
		return historyPosition{}, false
	}
	topic := turns[viewed].Topic
	indexes := realTurns(turns, topic)
	if len(indexes) == 0 {
		return historyPosition{}, false
	}
	step := 0
	for _, index := range indexes {
		if index <= viewed {
			step++
		}
	}
	return historyPosition{topic, max(1, step), len(indexes)}, true
}

func studiedTopics(course []string, turns []session.Turn) []string {
	var studied []string
	for _, topic := range course {
		if len(realTurns(turns, topic)) > 0 {
			studied = append(studied, topic)
		}
	}
	return studied
}

type historyStep struct {
	target   int
	crossing bool
	moved    bool
}

func stepHistory(course []string, state session.State, viewed, direction int) historyStep {
	position, found := positionOf(state.Turns, viewed)
	if !found {
		return historyStep{}
	}
	indexes := realTurns(state.Turns, position.topic)
	next := position.step - 1 + direction
	if next >= 0 && next < len(indexes) {
		return historyStep{target: liveIndex(state, position.topic, indexes[next]), moved: true}
	}
	studied := studiedTopics(course, state.Turns)
	current := -1
	for index, topic := range studied {
		if topic == position.topic {
			current = index
		}
	}
	neighbor := current + direction
	if current < 0 || neighbor < 0 || neighbor >= len(studied) {
		return historyStep{}
	}
	others := realTurns(state.Turns, studied[neighbor])
	target := others[0]
	if direction < 0 {
		target = others[len(others)-1]
	}
	return historyStep{target: liveIndex(state, studied[neighbor], target), crossing: true, moved: true}
}

func liveIndex(state session.State, topic string, index int) int {
	indexes := realTurns(state.Turns, topic)
	if topic == state.Topic && len(indexes) > 0 && index == indexes[len(indexes)-1] {
		return len(state.Turns) - 1
	}
	return index
}

func (model interfaceModel) courseTopics() []string {
	topics := make([]string, 0, len(model.application.course))
	for _, topic := range model.application.course {
		topics = append(topics, topic.ID)
	}
	return topics
}

func (model interfaceModel) topicNumber(topic string) int {
	for index, current := range model.application.course {
		if current.ID == topic {
			return index + 1
		}
	}
	return 0
}

func (model interfaceModel) topicName(topic string) string {
	for _, current := range model.application.course {
		if current.ID == topic {
			return tutor.Plain(current.Title)
		}
	}
	return topic
}

func (model interfaceModel) stepTurns(direction int) (tea.Model, tea.Cmd) {
	step := stepHistory(model.courseTopics(), model.application.state, model.history, direction)
	if !step.moved {
		model.notice = "This is the first step of the course so far"
		if direction > 0 {
			model.notice = "This is the latest step"
		}
		return model, nil
	}
	if step.crossing {
		bump := fmt.Sprintf("%d:%d", model.history, direction)
		if model.edgeBump != bump {
			model.edgeBump = bump
			model.notice = "First step of this topic · press again for the previous topic"
			if direction > 0 {
				model.notice = "Last step of this topic · press again for the next topic"
			}
			return model, nil
		}
		model.edgeBump = ""
		model.history = step.target
		model.notice = "Previous topic: " + model.topicName(model.application.state.Turns[step.target].Topic)
		if direction > 0 {
			model.notice = "Next topic: " + model.topicName(model.application.state.Turns[step.target].Topic)
		}
	} else {
		model.edgeBump = ""
		model.history = step.target
	}
	model.rebuildContent()
	model.viewport.GotoTop()
	return model, nil
}

func (model interfaceModel) historyTitle() string {
	turns := model.application.state.Turns
	title := model.topicTitle()
	position, found := positionOf(turns, model.history)
	if !found {
		return title
	}
	if number := model.topicNumber(position.topic); number > 0 {
		title += fmt.Sprintf("  ·  Topic %d / %d", number, len(model.application.course))
	}
	return title + fmt.Sprintf("  ·  Step %d / %d", position.step, position.steps)
}
