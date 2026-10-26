package terminal

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const matchedWords = 4

var (
	markdownLink   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	markdownMarker = regexp.MustCompile(`^\s*(?:#{1,6}\s+|>\s*|[-*+]\s+|\d+[.)]\s+)+`)
	markdownMarks  = strings.NewReplacer("**", "", "__", "", "*", "", "_", "", "`", "")
)

func markdownWords(line string) string {
	line = markdownMarker.ReplaceAllString(line, "")
	line = markdownLink.ReplaceAllString(line, "$1")
	line = markdownMarks.Replace(line)
	words := strings.Fields(line)
	if len(words) > matchedWords {
		words = words[:matchedWords]
	}
	return strings.Join(words, " ")
}

func renderedRow(source []string, line int, rendered string) (int, bool) {
	if line < 1 || line > len(source) {
		return 0, false
	}
	target := markdownWords(source[line-1])
	if len(target) < 4 {
		return 0, false
	}
	occurrence := 0
	for _, earlier := range source[:line-1] {
		if markdownWords(earlier) == target {
			occurrence++
		}
	}
	for row, text := range strings.Split(rendered, "\n") {
		if strings.Contains(strings.Join(strings.Fields(ansi.Strip(text)), " "), target) {
			if occurrence == 0 {
				return row, true
			}
			occurrence--
		}
	}
	return 0, false
}
