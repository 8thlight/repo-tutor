package store_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"repo-tutor/internal/settings/codec"
	"repo-tutor/internal/settings/store"
)

func withTabStyle(style string) func(codec.Settings) codec.Settings {
	return func(current codec.Settings) codec.Settings {
		current.TabStyle = style
		return current
	}
}

func TestMissingSettingsFileGivesDefaultsWithoutNotice(t *testing.T) {
	got, status, err := store.LoadSettings(t.TempDir())
	if err != nil || got != codec.Defaults() || status != (codec.Status{}) {
		t.Fatalf("got %+v %+v %v", got, status, err)
	}
}

func TestChangeSettingsSavesAndLoadsBack(t *testing.T) {
	base := t.TempDir()
	if _, err := store.ChangeSettings(base, withTabStyle("rounded")); err != nil {
		t.Fatal(err)
	}
	got, _, _ := store.LoadSettings(base)
	if got.TabStyle != "rounded" {
		t.Fatalf("got %+v", got)
	}
}

func TestChangeSettingsKeepsFieldsAnotherTutorChanged(t *testing.T) {
	base := t.TempDir()
	store.ChangeSettings(base, func(current codec.Settings) codec.Settings {
		current.MarkdownOpensRendered = false
		return current
	})
	store.ChangeSettings(base, withTabStyle("none"))
	got, _, _ := store.LoadSettings(base)
	if got.TabStyle != "none" || got.MarkdownOpensRendered {
		t.Fatalf("a save should change only its own field: %+v", got)
	}
}

func TestUnreadableSettingsAreMovedAsideOnTheNextSave(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "settings.json")
	os.WriteFile(path, []byte("{broken"), 0600)
	_, status, _ := store.LoadSettings(base)
	if !status.Unreadable {
		t.Fatal("a broken file should be reported as unreadable")
	}
	if data, _ := os.ReadFile(path); string(data) != "{broken" {
		t.Fatal("loading must not touch the file")
	}
	store.ChangeSettings(base, withTabStyle("none"))
	if data, _ := os.ReadFile(path + ".bad"); string(data) != "{broken" {
		t.Fatal("the broken file should be kept as settings.json.bad")
	}
	if got, _, _ := store.LoadSettings(base); got.TabStyle != "none" {
		t.Fatalf("a new file should hold the change: %+v", got)
	}
}

func TestSettingsFromANewerVersionAreNeverWrittenOver(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "settings.json")
	original := `{"version":2,"tab_style":"none"}`
	os.WriteFile(path, []byte(original), 0600)
	_, err := store.ChangeSettings(base, withTabStyle("block"))
	if !errors.Is(err, store.ErrSettingsReadOnly) {
		t.Fatalf("got %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != original {
		t.Fatal("the newer file must be left alone")
	}
}
