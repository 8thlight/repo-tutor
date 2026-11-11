package terminal

import (
	"testing"

	"repo-tutor/internal/settings/codec"
)

func TestStartupShowsAnUnreadableSettingsNotice(t *testing.T) {
	model := interfaceFixture()
	model.application.settingsStatus = codec.Status{Unreadable: true}
	reopened := newInterface(model.ctx, model.application, model.run, true)
	if reopened.notice != settingsUnreadableNotice {
		t.Fatalf("got %q", reopened.notice)
	}
}

func TestStartupShowsAReadOnlySettingsNotice(t *testing.T) {
	model := interfaceFixture()
	model.application.settingsStatus = codec.Status{ReadOnly: true}
	reopened := newInterface(model.ctx, model.application, model.run, true)
	if reopened.notice != settingsReadOnlyNotice {
		t.Fatalf("got %q", reopened.notice)
	}
}
