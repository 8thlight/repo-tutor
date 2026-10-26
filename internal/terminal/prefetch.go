package terminal

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

const prefetchDelay = 3 * time.Second

type lessonPrefetcher func(context.Context, session.State, string) error

type prefetchDue struct{}

type prefetchFinished struct {
	id    int
	topic string
	err   error
}

func (model interfaceModel) prefetchTarget() (string, bool) {
	state := model.application.state
	if model.prefetcher == nil || model.busy || model.prefetching || len(state.Turns) == 0 || model.application.store == nil {
		return "", false
	}
	next, found := tutor.NextUnheld(model.application.course, state.Progress, state.Topic)
	if !found || model.prefetchFailed[next.ID] {
		return "", false
	}
	for _, turn := range state.Turns {
		if turn.Topic == next.ID {
			return "", false
		}
	}
	if lesson, cached := model.application.cachedLesson(next.ID); cached && lesson.Fresh(model.application.snapshot.Fingerprint) {
		return "", false
	}
	return next.ID, true
}

func (model interfaceModel) schedulePrefetch(before interfaceModel) (interfaceModel, tea.Cmd) {
	changed := before.application.state.Topic != model.application.state.Topic ||
		len(before.application.state.Turns) != len(model.application.state.Turns) ||
		(before.busy && !model.busy)
	if !changed || model.prefetchScheduled {
		return model, nil
	}
	if _, wanted := model.prefetchTarget(); !wanted {
		return model, nil
	}
	model.prefetchScheduled = true
	return model, tea.Tick(prefetchDelay, func(time.Time) tea.Msg { return prefetchDue{} })
}

func (model interfaceModel) startPrefetch() (interfaceModel, tea.Cmd) {
	topic, wanted := model.prefetchTarget()
	if !wanted {
		return model, nil
	}
	prefetchContext, cancel := context.WithCancel(model.ctx)
	model.prefetching, model.prefetchCancel = true, cancel
	model.prefetchID++
	id, state, prefetcher := model.prefetchID, model.application.state, model.prefetcher
	work := func() tea.Msg {
		return prefetchFinished{id, topic, prefetcher(prefetchContext, state, topic)}
	}
	return model, tea.Batch(work, model.prefetchSpinner.Tick)
}

func (model interfaceModel) stopPrefetch() interfaceModel {
	if model.prefetchCancel != nil {
		model.prefetchCancel()
		model.prefetchCancel = nil
	}
	if model.prefetching {
		model.prefetching = false
		model.prefetchID++
	}
	return model
}

func (model interfaceModel) finishPrefetch(finished prefetchFinished) interfaceModel {
	if !model.prefetching || finished.id != model.prefetchID {
		return model
	}
	if model.prefetchCancel != nil {
		model.prefetchCancel()
		model.prefetchCancel = nil
	}
	model.prefetching = false
	if finished.err != nil {
		failed := make(map[string]bool, len(model.prefetchFailed)+1)
		for topic := range model.prefetchFailed {
			failed[topic] = true
		}
		failed[finished.topic] = true
		model.prefetchFailed = failed
	}
	return model
}

func (model interfaceModel) footerLine(notice, indicator string, width int, style lipgloss.Style) string {
	if indicator == "" {
		return style.Render(ansi.Truncate(notice, width, "…"))
	}
	room := max(0, width-lipgloss.Width(indicator)-2)
	left := ansi.Truncate(notice, room, "…")
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(indicator))
	return style.Render(left + spaces(gap) + indicator)
}

func spaces(count int) string {
	line := make([]byte, count)
	for index := range line {
		line[index] = ' '
	}
	return string(line)
}
