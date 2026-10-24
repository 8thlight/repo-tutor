package tutor

import (
	"fmt"
	"strings"

	"repo-tutor/internal/repository"
)

const Overview = "overview"

type Topic struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func Course(snapshot repository.Snapshot) []Topic {
	result := []Topic{{Overview, "Repository overview"}}
	for _, current := range repository.Foundations(snapshot) {
		if len(current.Files) > 0 {
			result = append(result, Topic{current.Path, current.Directory + " (" + current.Name + ")"})
		}
	}
	return result
}

func Diagram(snapshot repository.Snapshot) string {
	var output strings.Builder
	output.WriteString("flowchart TD\n")
	indexes := map[string]int{}
	for index, current := range snapshot.Packages {
		indexes[current.Path] = index
		fmt.Fprintf(&output, "  p%d[%q]\n", index, current.Directory)
	}
	for index, current := range snapshot.Packages {
		for _, dependency := range current.Dependencies {
			fmt.Fprintf(&output, "  p%d --> p%d\n", index, indexes[dependency])
		}
	}
	return output.String()
}
