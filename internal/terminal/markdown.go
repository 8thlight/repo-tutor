package terminal

import (
	"html"
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const matchedWords = 4

var (
	htmlTag        = regexp.MustCompile(`<[^>]*>`)
	markdownLink   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	markdownMarker = regexp.MustCompile(`^\s*(?:#{1,6}\s+|>\s*|[-*+]\s+|\d+[.)]\s+)+`)
	markdownMarks  = strings.NewReplacer("**", "", "__", "", "*", "", "_", "", "`", "")
	tableRules     = strings.NewReplacer("|", " ", "│", " ")
)

func markdownWords(line string) string {
	line = html.UnescapeString(htmlTag.ReplaceAllString(line, " "))
	line = markdownMarker.ReplaceAllString(line, "")
	line = markdownLink.ReplaceAllString(line, "$1")
	line = tableRules.Replace(markdownMarks.Replace(line))
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
		if strings.Contains(strings.Join(strings.Fields(tableRules.Replace(ansi.Strip(text))), " "), target) {
			if occurrence == 0 {
				return row, true
			}
			occurrence--
		}
	}
	return 0, false
}

var trailingStyles = regexp.MustCompile(`(?:\x1b\[[0-9;]*m)+$`)

func trimPadding(row string) string {
	trimmed := ansi.Truncate(row, ansi.StringWidth(strings.TrimRight(ansi.Strip(row), " ")), "")
	return trailingStyles.ReplaceAllString(trimmed, "\x1b[m")
}
