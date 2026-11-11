package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"repo-tutor/internal/file/writer"
	"repo-tutor/internal/tutor"
)

const (
	Version = 8

	lessonPrefix  = "lesson-"
	lessonSuffix  = ".md"
	goalKeyLength = 16
)

type Turn struct {
	Topic   string      `json:"topic"`
	Action  string      `json:"action"`
	Message string      `json:"message"`
	Reply   tutor.Reply `json:"reply"`
	Resumed bool        `json:"resumed,omitempty"`
}

type State struct {
	Version     int                       `json:"version"`
	Repository  string                    `json:"repository"`
	Fingerprint string                    `json:"fingerprint"`
	Agent       string                    `json:"agent"`
	Model       string                    `json:"model"`
	SessionID   string                    `json:"session_id"`
	Goal        string                    `json:"goal"`
	GoalChosen  bool                      `json:"goal_chosen,omitempty"`
	Topic       string                    `json:"topic"`
	Question    string                    `json:"question"`
	Review      bool                      `json:"review,omitempty"`
	Quiz        *tutor.Quiz               `json:"quiz,omitempty"`
	LastExam    *tutor.ExamResult         `json:"last_exam,omitempty"`
	Probe       *tutor.Probe              `json:"probe,omitempty"`
	Progress    map[string]tutor.Progress `json:"progress"`
	Turns       []Turn                    `json:"turns"`
	Updated     time.Time                 `json:"updated"`
}

type Store struct {
	Directory string
	lock      *os.File
}

func Base(base string) (string, error) {
	if base != "" {
		return base, nil
	}
	base = os.Getenv("XDG_STATE_HOME")
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(base, "repo-tutor"), nil
}

func BaseSource(base string) string {
	if base != "" {
		return "--state-dir"
	}
	if os.Getenv("XDG_STATE_HOME") != "" {
		return "XDG_STATE_HOME"
	}
	return "the user configuration directory"
}

func Open(base string, repository string) (*Store, error) {
	base, err := Base(base)
	if err != nil {
		return nil, err
	}
	identifier := fmt.Sprintf("%x", sha256.Sum256([]byte(repository)))
	directory := filepath.Join(base, identifier[:24])
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(filepath.Join(directory, "session.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		return nil, fmt.Errorf("another tutor is using this repository's session; close it or use a different --state-dir")
	}
	return &Store{Directory: directory, lock: lock}, nil
}

func (store *Store) Close() error {
	return store.lock.Close()
}

func (store *Store) Load() (State, bool, error) {
	file, err := os.Open(filepath.Join(store.Directory, "session.json"))
	if errors.Is(err, os.ErrNotExist) {
		return State{}, false, nil
	}
	if err != nil {
		return State{}, false, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 16*1024*1024))
	if err != nil {
		return State{}, false, err
	}
	var saved struct {
		Version int `json:"version"`
	}
	if json.Unmarshal(data, &saved) == nil && saved.Version > Version {
		return State{}, false, fmt.Errorf("saved session comes from a newer repo-tutor (version %d; this one reads %d); update repo-tutor to continue it, or use --new to archive it and start again", saved.Version, Version)
	}
	var state State
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&state); err != nil {
		return State{}, false, fmt.Errorf("read saved session: %w (use --new to archive it and start again)", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return State{}, false, fmt.Errorf("saved session contains extra data; use --new to archive it and start again")
	}
	if state.Version != Version || state.Repository == "" || state.Topic == "" || (state.Agent != "claude" && state.Agent != "codex") {
		return State{}, false, fmt.Errorf("saved session has an unsupported format; use --new to archive it and start again")
	}
	return state, true, nil
}

func (store *Store) lessonPath(key string) string {
	return filepath.Join(store.Directory, lessonPrefix+key+lessonSuffix)
}

type LessonFile struct {
	Path    string
	GoalKey string
}

func (store *Store) LessonFiles() []LessonFile {
	paths, err := filepath.Glob(filepath.Join(store.Directory, lessonPrefix+"*"+lessonSuffix))
	if err != nil {
		return nil
	}
	var files []LessonFile
	for _, path := range paths {
		name := strings.TrimPrefix(filepath.Base(path), lessonPrefix)
		if len(name) < goalKeyLength {
			continue
		}
		files = append(files, LessonFile{Path: path, GoalKey: name[:goalKeyLength]})
	}
	return files
}

func (store *Store) LoadLesson(key string) ([]byte, bool, error) {
	data, err := os.ReadFile(store.lessonPath(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (store *Store) SaveLesson(key string, data []byte) error {
	return store.write(filepath.Base(store.lessonPath(key)), data)
}

func (store *Store) Archive() error {
	data, err := os.ReadFile(filepath.Join(store.Directory, "session.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return store.write("previous.json", data)
}

func (store *Store) Save(state State) error {
	state.Updated = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if len(data) > 16*1024*1024 {
		return fmt.Errorf("session exceeds the 16 MiB storage limit; use --new to start a fresh session")
	}
	return store.write("session.json", append(data, '\n'))
}

func (store *Store) write(name string, data []byte) error {
	return writer.Write(store.Directory, name, data)
}
