package terminal

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/repository"
)

type lessonPlace struct{ history, offset int }

func (model interfaceModel) shownCitations() []repository.Location {
	turns := model.application.state.Turns
	if len(turns) == 0 {
		return nil
	}
	return turns[max(0, min(model.history, len(turns)-1))].Reply.Citations
}

func citationNumber(pressed string) (number int, withControl bool) {
	pressed, withControl = strings.CutPrefix(pressed, "ctrl+")
	if len(pressed) == 1 && pressed[0] >= '1' && pressed[0] <= '9' {
		return int(pressed[0] - '0'), withControl
	}
	return 0, false
}

func (model interfaceModel) citationNotice() string {
	count := len(model.shownCitations())
	switch count {
	case 0:
		return "This lesson cites no sources."
	case 1:
		return "This lesson cites 1 source."
	}
	return fmt.Sprintf("This lesson cites %d sources.", count)
}

func (model interfaceModel) citationKey(pressed string) (tea.Model, tea.Cmd, bool) {
	number, withControl := citationNumber(pressed)
	if number == 0 {
		return model, nil, false
	}
	switch {
	case model.page == "source" && model.sourceCitation > 0:
		if number == model.sourceCitation {
			next, command := model.back()
			return next, command, true
		}
	case model.page == "lesson" && model.askingGoal():
		return model, nil, false
	case model.page == "lesson":
		typing := model.input.Value() != "" && !model.busy
		if !withControl && (typing || questionOpen(model.application.state)) {
			return model, nil, false
		}
	default:
		return model, nil, false
	}
	if model.movingToAnotherTopic() && model.page == "lesson" {
		model.notice = sourcesWaitNotice
		return model, nil, true
	}
	citations := model.shownCitations()
	if number > len(citations) {
		model.notice = model.citationNotice()
		return model, nil, true
	}
	next, command := model.openCitation(number)
	return next, command, true
}

func (model interfaceModel) openCitation(number int) (tea.Model, tea.Cmd) {
	place := model.lessonPlaceToKeep()
	opened, command := model.openFile(model.shownCitations()[number-1])
	shown, ok := opened.(interfaceModel)
	if !ok || shown.page != "source" {
		return opened, command
	}
	shown.sourceCitation, shown.lessonReturn = number, place
	return shown, command
}

func (model interfaceModel) lessonPlaceToKeep() *lessonPlace {
	if model.page == "lesson" {
		return &lessonPlace{model.history, model.viewport.YOffset()}
	}
	return model.lessonReturn
}

func (model interfaceModel) backToTheLesson() (tea.Model, tea.Cmd) {
	place := model.lessonReturn
	model.page, model.sourceCitation, model.lessonReturn = "lesson", 0, nil
	model.history = place.history
	model.rebuildContent()
	model.viewport.SetYOffset(place.offset)
	if !model.busy {
		return model, model.input.Focus()
	}
	return model, nil
}
