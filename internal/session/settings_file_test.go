package session_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"repo-tutor/internal/session"
	"repo-tutor/internal/settings"
)

func withTabStyle(style string) func(settings.Settings) settings.Settings {
	return func(current settings.Settings) settings.Settings {
		current.TabStyle = style
		return current
	}
}

func TestMissingSettingsFileGivesDefaultsWithoutNotice(t *testing.T) {
	got, status, err := session.LoadSettings(t.TempDir())
	if err != nil || got != settings.Defaults() || status != (settings.Status{}) {
		t.Fatalf("got %+v %+v %v", got, status, err)
	}
}

func TestChangeSettingsSavesAndLoadsBack(t *testing.T) {
	base := t.TempDir()
	if _, err := session.ChangeSettings(base, withTabStyle("rounded")); err != nil {
		t.Fatal(err)
	}
	got, _, _ := session.LoadSettings(base)
	if got.TabStyle != "rounded" {
		t.Fatalf("got %+v", got)
	}
}

func TestChangeSettingsKeepsFieldsAnotherTutorChanged(t *testing.T) {
	base := t.TempDir()
	session.ChangeSettings(base, func(current settings.Settings) settings.Settings {
		current.MarkdownOpensRendered = false
		return current
	})
	session.ChangeSettings(base, withTabStyle("none"))
	got, _, _ := session.LoadSettings(base)
	if got.TabStyle != "none" || got.MarkdownOpensRendered {
		t.Fatalf("a save should change only its own field: %+v", got)
	}
}

func TestUnreadableSettingsAreMovedAsideOnTheNextSave(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "settings.json")
	os.WriteFile(path, []byte("{broken"), 0600)
	_, status, _ := session.LoadSettings(base)
	if !status.Unreadable {
		t.Fatal("a broken file should be reported as unreadable")
	}
	if data, _ := os.ReadFile(path); string(data) != "{broken" {
		t.Fatal("loading must not touch the file")
	}
	session.ChangeSettings(base, withTabStyle("none"))
	if data, _ := os.ReadFile(path + ".bad"); string(data) != "{broken" {
		t.Fatal("the broken file should be kept as settings.json.bad")
	}
	if got, _, _ := session.LoadSettings(base); got.TabStyle != "none" {
		t.Fatalf("a new file should hold the change: %+v", got)
	}
}

func TestSettingsFromANewerVersionAreNeverWrittenOver(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "settings.json")
	original := `{"version":2,"tab_style":"none"}`
	os.WriteFile(path, []byte(original), 0600)
	_, err := session.ChangeSettings(base, withTabStyle("block"))
	if !errors.Is(err, session.ErrSettingsReadOnly) {
		t.Fatalf("got %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != original {
		t.Fatal("the newer file must be left alone")
	}
}
