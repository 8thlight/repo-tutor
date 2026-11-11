package pruner

import (
	"os"
	"time"

	sessionstore "repo-tutor/internal/session/store"
)

func Prune(store *sessionstore.Store, currentGoalKey string, maxAge time.Duration, now time.Time) int {
	removed := 0
	for _, file := range store.LessonFiles() {
		if file.GoalKey == currentGoalKey {
			continue
		}
		info, err := os.Stat(file.Path)
		if err != nil || now.Sub(info.ModTime()) <= maxAge {
			continue
		}
		if os.Remove(file.Path) == nil {
			removed++
		}
	}
	return removed
}
