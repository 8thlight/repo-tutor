package terminal

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func prefetchFixture(t *testing.T) (interfaceModel, string) {
	t.Helper()
	model := twoTopicFixture()
	model.application.state.Turns[1].Topic, model.application.state.Topic = "learning", "learning"
	model.application.state.Progress = map[string]tutor.Progress{tutor.Overview: {Lessons: 1}, "learning": {Lessons: 1}}
	store, err := session.Open(t.TempDir(), "/work/repowise")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	model.application.store = store
	model.application.snapshot.Fingerprint = "current"
	next, found := tutor.NextUnheld(model.application.course, model.application.state.Progress, model.application.state.Topic)
	if !found {
		t.Fatal("the fixture should have a topic after the current one")
	}
	return model, next.ID
}

func runPrefetchCommand(t *testing.T, command tea.Cmd) prefetchFinished {
	t.Helper()
	message := command()
	if finished, ok := message.(prefetchFinished); ok {
		return finished
	}
	if batch, ok := message.(tea.BatchMsg); ok {
		for _, child := range batch {
			if finished, ok := child().(prefetchFinished); ok {
				return finished
			}
		}
	}
	t.Fatal("no prefetch result")
	return prefetchFinished{}
}

func TestPrefetchStartsAfterATurnShowsAnIndicatorAndClearsWhenDone(t *testing.T) {
	model, topic := prefetchFixture(t)
	var prefetched []string
	model.prefetcher = func(_ context.Context, _ session.State, name string) error {
		prefetched = append(prefetched, name)
		return nil
	}
	model.busy = true
	finished, schedule := model.Update(turnFinished{state: model.application.state})
	model = finished.(interfaceModel)
	if !model.prefetchScheduled || schedule == nil {
		t.Fatal("a finished turn should schedule a prefetch of the next topic")
	}
	started, command := model.Update(prefetchDue{})
	model = started.(interfaceModel)
	if !model.prefetching || strings.Contains(ansi.Strip(model.View().Content), "Preparing next lesson") == false {
		t.Fatal("a running prefetch should show its indicator")
	}
	result := runPrefetchCommand(t, command)
	if len(prefetched) != 1 || prefetched[0] != topic || result.err != nil {
		t.Fatalf("the next topic should be prefetched once: %v %+v", prefetched, result)
	}
	done, _ := model.Update(result)
	model = done.(interfaceModel)
	if model.prefetching || strings.Contains(ansi.Strip(model.View().Content), "Preparing next lesson") {
		t.Fatal("the indicator should clear when the prefetch ends")
	}
}

func TestForegroundTurnCancelsAPrefetchAndItsResultIsIgnored(t *testing.T) {
	model, _ := prefetchFixture(t)
	cancelled := make(chan struct{})
	model.prefetcher = func(prefetchContext context.Context, _ session.State, _ string) error {
		<-prefetchContext.Done()
		close(cancelled)
		return prefetchContext.Err()
	}
	started, command := model.startPrefetch()
	model = started
	results := make(chan prefetchFinished, 1)
	go func() { results <- runPrefetchCommand(t, command) }()
	begun, _ := model.begin(turnIntent{"lesson", firstLessonMessage, "learning"})
	model = begun.(interfaceModel)
	defer model.cancel()
	<-cancelled
	if model.prefetching || len(model.prefetchFailed) != 0 {
		t.Fatal("a foreground turn should stop the prefetch without marking it failed")
	}
	ignored, _ := model.Update(<-results)
	if len(ignored.(interfaceModel).prefetchFailed) != 0 || ignored.(interfaceModel).prefetching {
		t.Fatal("the cancelled prefetch's result should be ignored")
	}
}

func TestPrefetchSkipsStudiedCachedAndFailedTopicsAndNeverRetriesAFailure(t *testing.T) {
	model, topic := prefetchFixture(t)
	model.prefetcher = func(context.Context, session.State, string) error { return errors.New("agent unavailable") }
	if name, wanted := model.prefetchTarget(); !wanted || name != topic {
		t.Fatalf("expected the next topic %s, got %q %v", topic, name, wanted)
	}
	started, command := model.startPrefetch()
	failed, _ := started.Update(runPrefetchCommand(t, command))
	model = failed.(interfaceModel)
	if _, wanted := model.prefetchTarget(); wanted || !model.prefetchFailed[topic] {
		t.Fatal("a failed prefetch must not be retried")
	}

	model, topic = prefetchFixture(t)
	model.prefetcher = func(context.Context, session.State, string) error { return nil }
	reply := tutor.Reply{Explanation: "Cached.", Question: "Q?", Assessment: "none", Citations: model.application.state.Turns[0].Reply.Citations}
	lesson := tutor.SavedLesson{Fingerprint: "current", Goal: tutor.GoalKey(model.application.state.Goal), Topic: topic, Reply: reply}
	if err := model.application.store.SaveLesson(tutor.LessonKey(model.application.state.Goal, topic), tutor.EncodeLesson(lesson)); err != nil {
		t.Fatal(err)
	}
	if _, wanted := model.prefetchTarget(); wanted {
		t.Fatal("a fresh cached lesson needs no prefetch")
	}
	lesson.Fingerprint = "older"
	if err := model.application.store.SaveLesson(tutor.LessonKey(model.application.state.Goal, topic), tutor.EncodeLesson(lesson)); err != nil {
		t.Fatal(err)
	}
	if _, wanted := model.prefetchTarget(); !wanted {
		t.Fatal("a cached lesson for changed source should be prefetched again")
	}
	model.application.state.Turns = append(model.application.state.Turns, session.Turn{Topic: topic, Action: "lesson", Reply: reply})
	if _, wanted := model.prefetchTarget(); wanted {
		t.Fatal("a studied topic needs no prefetch")
	}
	model.prefetcher = nil
	model.application.state.Turns = model.application.state.Turns[:2]
	if _, wanted := model.prefetchTarget(); wanted {
		t.Fatal("without a prefetcher nothing runs")
	}
}
