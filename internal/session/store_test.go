package session_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func TestPersistenceLockAndArchive(t *testing.T) {
	base := t.TempDir()
	store, err := session.Open(base, "/repository")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.Open(base, "/repository"); err == nil {
		t.Fatal("two tutors obtained the same session lock")
	}
	state := session.State{Version: session.Version, Repository: "/repository", Agent: "codex", Topic: "overview", SessionID: "session-id", Question: "Where does it begin?", Progress: map[string]tutor.Progress{"overview": {Correct: 1}}}
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
	second, err := session.Open(base, "/repository")
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

func TestOverviewIsKeptPerGoalAndReplacedInPlace(t *testing.T) {
	store, err := session.Open(t.TempDir(), "/repository")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, found, err := store.LoadOverview("goal-a"); err != nil || found {
		t.Fatal("a missing overview should be a miss, not an error")
	}
	if err := store.SaveOverview("goal-a", []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveOverview("goal-b", []byte("other goal\n")); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveOverview("goal-a", []byte("second\n")); err != nil {
		t.Fatal(err)
	}
	data, found, err := store.LoadOverview("goal-a")
	if err != nil || !found || string(data) != "second\n" {
		t.Fatalf("the newest overview for a goal should replace the older one: %q %v", data, err)
	}
	files, _ := filepath.Glob(filepath.Join(store.Directory, "overview-*.md"))
	if len(files) != 2 {
		t.Fatalf("expected one file per goal, found %v", files)
	}
	information, err := os.Stat(files[0])
	if err != nil || information.Mode().Perm() != 0600 {
		t.Fatal("overview was not saved privately")
	}
}
