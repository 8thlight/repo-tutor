package tutor

import (
	"fmt"
	"strings"

	"repo-tutor/internal/repository"
)

const (
	filePrefix      = ":file:"
	fileWindowLimit = 6000
	fileWindowCount = 6
	FileTitlePrefix = "File · "
)

func FileTopic(path string) string {
	return filePrefix + path
}

func FilePath(topic string) (string, bool) {
	return strings.CutPrefix(topic, filePrefix)
}

func fileWindowStarts(snapshot repository.Snapshot, path string) []int {
	file, found := snapshot.File(path)
	if !found || file.Content == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(file.Content, "\n"), "\n")
	starts := []int{1}
	size := 0
	for index, line := range lines {
		numbered := len(fmt.Sprintf("%d: %s\n", index+1, line))
		if size+numbered > fileWindowLimit && size > 0 {
			if len(starts) == fileWindowCount {
				break
			}
			starts = append(starts, index+1)
			size = 0
		}
		size += numbered
	}
	return starts
}
