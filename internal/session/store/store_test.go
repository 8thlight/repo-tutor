package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sessionstore "repo-tutor/internal/session/store"
	"repo-tutor/internal/tutor"
)

func TestPersistenceLockAndArchive(t *testing.T) {
	base := t.TempDir()
	store, err := sessionstore.Open(base, "/repository")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sessionstore.Open(base, "/repository"); err == nil {
		t.Fatal("two tutors obtained the same session lock")
	}
	state := sessionstore.State{Version: sessionstore.Version, Repository: "/repository", Agent: "codex", Topic: tutor.Overview, SessionID: "session-id", Question: "Where does it begin?", Progress: map[string]tutor.Progress{tutor.Overview: {Correct: 1}}}
	if err := store.Save(state); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := store.Load()
	if err != nil || !found || loaded.SessionID != state.SessionID || loaded.Question != state.Question || !reflect.DeepEqual(loaded.Progress, state.Progress) {
		t.Fatalf("session did not round-trip: %+v %v", loaded, err)
	}
	information, err := os.Stat(filepath.Join(store.Directory, "session.json"))
	if err != nil || information.Mode().Perm() != 0600 {
		t.Fatal("session was not saved privately")
	}
	if err := store.Archive(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(store.Directory, "previous.json")); err != nil || len(data) == 0 {
		t.Fatal("previous session was not archived")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := sessionstore.Open(base, "/repository")
	if err != nil {
		t.Fatalf("closing the store did not release its lock: %v", err)
	}
	defer second.Close()
	if err := os.WriteFile(filepath.Join(second.Directory, "session.json"), []byte("broken JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := second.Load(); err == nil {
		t.Fatal("corrupt state was silently accepted")
	}
}

func TestLessonIsKeptPerKeyAndReplacedInPlace(t *testing.T) {
	store, err := sessionstore.Open(t.TempDir(), "/repository")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, found, err := store.LoadLesson("goal-a"); err != nil || found {
		t.Fatal("a missing overview should be a miss, not an error")
	}
	if err := store.SaveLesson("goal-a", []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveLesson("goal-b", []byte("other goal\n")); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveLesson("goal-a", []byte("second\n")); err != nil {
		t.Fatal(err)
	}
	data, found, err := store.LoadLesson("goal-a")
	if err != nil || !found || string(data) != "second\n" {
		t.Fatalf("the newest overview for a key should replace the older one: %q %v", data, err)
	}
	files, _ := filepath.Glob(filepath.Join(store.Directory, "lesson-*.md"))
	if len(files) != 2 {
		t.Fatalf("expected one file per key, found %v", files)
	}
	information, err := os.Stat(files[0])
	if err != nil || information.Mode().Perm() != 0600 {
		t.Fatal("overview was not saved privately")
	}
}

func TestBaseDirectoryPrefersTheFlagThenXDGStateHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/state")
	if base, _ := sessionstore.Base("/chosen"); base != "/chosen" {
		t.Fatalf("an explicit directory should win, got %s", base)
	}
	if base, _ := sessionstore.Base(""); base != "/state/repo-tutor" {
		t.Fatalf("XDG_STATE_HOME should hold a repo-tutor folder, got %s", base)
	}
}

func TestBaseSourceNamesWhatChoseTheDirectory(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/state")
	if source := sessionstore.BaseSource("/chosen"); source != "--state-dir" {
		t.Fatalf("a flag should be named, got %s", source)
	}
	if source := sessionstore.BaseSource(""); source != "XDG_STATE_HOME" {
		t.Fatalf("the environment should be named, got %s", source)
	}
	t.Setenv("XDG_STATE_HOME", "")
	if source := sessionstore.BaseSource(""); source != "the user configuration directory" {
		t.Fatalf("the default should be named, got %s", source)
	}
}

func TestNewerSessionIsLeftAloneAndTheMessageSaysToUpdate(t *testing.T) {
	store, err := sessionstore.Open(t.TempDir(), "/repository")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	path := filepath.Join(store.Directory, "session.json")
	newer := []byte(fmt.Sprintf(`{"version": %d, "unknown": {"shape": [1, 2]}}`, sessionstore.Version+1))
	if err := os.WriteFile(path, newer, 0600); err != nil {
		t.Fatal(err)
	}
	_, found, err := store.Load()
	if found || err == nil {
		t.Fatalf("a newer session must not load: %v", err)
	}
	for _, expected := range []string{"newer repo-tutor", fmt.Sprintf("version %d", sessionstore.Version+1), fmt.Sprintf("reads %d", sessionstore.Version), "update", "--new"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("the message should say %q: %v", expected, err)
		}
	}
	if strings.Index(err.Error(), "update") > strings.Index(err.Error(), "--new") {
		t.Fatalf("the message should offer the update before --new: %v", err)
	}
	if after, _ := os.ReadFile(path); string(after) != string(newer) {
		t.Fatalf("loading changed the session file: %s", after)
	}
}

func TestOlderAndDamagedSessionsKeepTheMessageToStartAgain(t *testing.T) {
	for name, content := range map[string]string{
		"older":   fmt.Sprintf(`{"version": %d, "repository": "/r", "topic": "t", "agent": "codex"}`, sessionstore.Version-1),
		"damaged": `{"version": `,
	} {
		store, err := sessionstore.Open(t.TempDir(), "/repository")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(store.Directory, "session.json"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		_, _, err = store.Load()
		store.Close()
		if err == nil || !strings.Contains(err.Error(), "use --new to archive it and start again") || strings.Contains(err.Error(), "newer") {
			t.Fatalf("%s session should keep the old message: %v", name, err)
		}
	}
}

func TestLessonFilesListPathAndGoalKeyAndSkipShortNames(t *testing.T) {
	store, err := sessionstore.Open(t.TempDir(), "/repository")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	key := "0123456789abcdef-core"
	if err := store.SaveLesson(key, []byte("lesson")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"lesson-short.md", "session.json", "lesson-0123456789abcdef.txt"} {
		if err := os.WriteFile(filepath.Join(store.Directory, name), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	want := []sessionstore.LessonFile{{
		Path:    filepath.Join(store.Directory, "lesson-"+key+".md"),
		GoalKey: "0123456789abcdef",
	}}
	if got := store.LessonFiles(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
