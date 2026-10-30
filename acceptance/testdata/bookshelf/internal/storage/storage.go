package storage

import (
	"encoding/json"
	"errors"
	"os"

	"bookshelf/internal/library"
)

type Store struct{ path string }

func Open(path string) Store {
	return Store{path: path}
}

func (store Store) Load() (library.Shelf, error) {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return library.Shelf{}, nil
	}
	if err != nil {
		return library.Shelf{}, err
	}
	var shelf library.Shelf
	return shelf, json.Unmarshal(data, &shelf)
}

func (store Store) Save(shelf library.Shelf) error {
	data, err := json.MarshalIndent(shelf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(store.path, data, 0o600)
}
