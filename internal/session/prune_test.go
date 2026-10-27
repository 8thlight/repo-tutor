package session_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

func TestPruneRemovesOnlyOldLessonsOfOtherGoals(t *testing.T) {
	store, err := session.Open(t.TempDir(), "/work/repo")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	current, other := "Trace the pipeline", "Learn the settings"
	files := map[string]time.Duration{
		tutor.LessonKey(current, "overview"): 200 * 24 * time.Hour,
		tutor.LessonKey(other, "overview"):   200 * 24 * time.Hour,
		tutor.LessonKey(other, "core"):       3 * 24 * time.Hour,
	}
	now := time.Now()
	for key, age := range files {
		if err := store.SaveLesson(key, []byte("lesson")); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(store.Directory, "lesson-"+key+".md")
		if err := os.Chtimes(path, now.Add(-age), now.Add(-age)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(store.Directory, "session.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	removed := store.PruneLessons(tutor.GoalKey(current), 60*24*time.Hour, now)
	if removed != 1 {
		t.Fatalf("only the old file for the other goal should go, removed %d", removed)
	}
	for key, want := range map[string]bool{
		tutor.LessonKey(current, "overview"): true,
		tutor.LessonKey(other, "overview"):   false,
		tutor.LessonKey(other, "core"):       true,
	} {
		_, found, _ := store.LoadLesson(key)
		if found != want {
			t.Fatalf("lesson %s present=%v, want %v", key, found, want)
		}
	}
	if _, err := os.Stat(filepath.Join(store.Directory, "session.json")); err != nil {
		t.Fatal("pruning must leave the session alone")
	}
}

func TestPruneIgnoresFilesThatAreNotLessons(t *testing.T) {
	store, err := session.Open(t.TempDir(), "/work/repo")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	odd := filepath.Join(store.Directory, "lesson-short.md")
	os.WriteFile(odd, []byte("x"), 0600)
	old := time.Now().Add(-400 * 24 * time.Hour)
	os.Chtimes(odd, old, old)
	if store.PruneLessons(tutor.GoalKey("g"), time.Hour, time.Now()) != 0 {
		t.Fatal("a file without a goal key is not a lesson to prune")
	}
}
