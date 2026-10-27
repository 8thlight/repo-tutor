package session

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	lessonPrefix  = "lesson-"
	goalKeyLength = 16
)

func (store *Store) PruneLessons(currentGoalKey string, maxAge time.Duration, now time.Time) int {
	paths, err := filepath.Glob(filepath.Join(store.Directory, lessonPrefix+"*.md"))
	if err != nil {
		return 0
	}
	removed := 0
	for _, path := range paths {
		name := strings.TrimPrefix(filepath.Base(path), lessonPrefix)
		if len(name) < goalKeyLength || name[:goalKeyLength] == currentGoalKey {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || now.Sub(info.ModTime()) <= maxAge {
			continue
		}
		if os.Remove(path) == nil {
			removed++
		}
	}
	return removed
}
