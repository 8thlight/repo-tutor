package terminal

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func expiryOf(t *testing.T, command tea.Cmd) (noticeExpired, bool) {
	t.Helper()
	if command == nil {
		return noticeExpired{}, false
	}
	messages := make(chan tea.Msg, 32)
	var run func(tea.Cmd)
	run = func(each tea.Cmd) {
		if each == nil {
			return
		}
		go func() {
			message := each()
			if batch, ok := message.(tea.BatchMsg); ok {
				for _, child := range batch {
					run(child)
				}
				return
			}
			messages <- message
		}()
	}
	run(command)
	deadline := time.After(300 * time.Millisecond)
	for {
		select {
		case message := <-messages:
			if expired, ok := message.(noticeExpired); ok {
				return expired, true
			}
		case <-deadline:
			return noticeExpired{}, false
		}
	}
}

func withShortNotices(t *testing.T) {
	noticeLifetime = time.Millisecond
	t.Cleanup(func() { noticeLifetime = 10 * time.Second })
}

func TestANoticeClearsWhenItsTimerFires(t *testing.T) {
	withShortNotices(t)
	model := quizFeedbackFixture()
	shown, command := model.Update(tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl})
	if shown.(interfaceModel).notice != "Ctrl+N asks the next question." {
		t.Fatalf("the fixture should show a notice, got %q", shown.(interfaceModel).notice)
	}
	expired, scheduled := expiryOf(t, command)
	if !scheduled {
		t.Fatal("showing a notice should schedule its expiry")
	}
	cleared, _ := shown.Update(expired)
	if cleared.(interfaceModel).notice != "" {
		t.Fatalf("the notice should clear when its timer fires, got %q", cleared.(interfaceModel).notice)
	}
}

func TestAnOlderNoticesTimerCannotClearANewerNotice(t *testing.T) {
	withShortNotices(t)
	model := quizFeedbackFixture()
	first, command := model.Update(tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl})
	stale, _ := expiryOf(t, command)
	newer := first.(interfaceModel)
	newer.notice = "Something else"
	newer, _ = newer.expireNotice(first.(interfaceModel))
	if newer.noticeGeneration <= stale.generation {
		t.Fatal("a new notice should start a new generation")
	}
	still, _ := newer.Update(stale)
	if still.(interfaceModel).notice != "Something else" {
		t.Fatal("a stale timer must not clear a notice shown after it")
	}
}

func TestAFailureNoticeKeepsItsRetryHintUntilSomethingReplacesIt(t *testing.T) {
	withShortNotices(t)
	model := interfaceFixture()
	model.errorText = "agent unavailable"
	model.notice = "That did not finish. Your progress and draft are safe. Ctrl+R tries again."
	next, command := model.expireNotice(interfaceModel{})
	if command != nil || next.noticeGeneration != 0 {
		t.Fatal("a failure notice should schedule no expiry")
	}
	cleared, _ := next.Update(noticeExpired{generation: 0})
	if cleared.(interfaceModel).notice == "" {
		t.Fatal("a failure notice must stay while the failure is unresolved")
	}
}

func TestSkipAndLeavePromptsKeepTheirOwnTimer(t *testing.T) {
	withShortNotices(t)
	for name, model := range map[string]interfaceModel{"skip": openQuestionFixture(), "leave": quizWithLessonFixture()} {
		key := 'n'
		if name == "leave" {
			key = 'l'
		}
		armed, command := model.Update(tea.KeyPressMsg{Code: key, Mod: tea.ModCtrl})
		if _, scheduled := expiryOf(t, command); scheduled {
			t.Fatalf("the %s prompt has its own five-second timer, not the notice timer", name)
		}
		if armed.(interfaceModel).notice == "" {
			t.Fatalf("the %s prompt should be showing", name)
		}
	}
}
