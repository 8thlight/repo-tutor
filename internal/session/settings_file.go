package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"repo-tutor/internal/settings/codec"
)

const settingsName = "settings.json"

var ErrSettingsReadOnly = errors.New("settings come from a newer repo-tutor and are read-only")

func LoadSettings(base string) (codec.Settings, codec.Status, error) {
	base, err := Base(base)
	if err != nil {
		return codec.Defaults(), codec.Status{}, err
	}
	data, err := os.ReadFile(filepath.Join(base, settingsName))
	if errors.Is(err, os.ErrNotExist) {
		return codec.Defaults(), codec.Status{}, nil
	}
	if err != nil {
		return codec.Defaults(), codec.Status{Unreadable: true}, nil
	}
	loaded, status := codec.Decode(data)
	return loaded, status, nil
}

func ChangeSettings(base string, change func(codec.Settings) codec.Settings) (codec.Settings, error) {
	base, err := Base(base)
	if err != nil {
		return codec.Defaults(), err
	}
	current, status, err := LoadSettings(base)
	if err != nil {
		return current, err
	}
	if status.ReadOnly {
		return current, ErrSettingsReadOnly
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return current, err
	}
	path := filepath.Join(base, settingsName)
	if status.Unreadable {
		if err := os.Rename(path, path+".bad"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return current, fmt.Errorf("could not move the unreadable settings aside: %w", err)
		}
	}
	changed := change(current)
	return changed, writeFile(base, settingsName, codec.Encode(changed))
}
