package tutor

import (
	"regexp"
	"strconv"
)

const MaxCitations = 9

var (
	fencedCode    = regexp.MustCompile("(?s)```.*?```")
	inlineCode    = regexp.MustCompile("`[^`\n]*`")
	citationMarks = regexp.MustCompile(`\[(\d+)\]`)
)

func Markers(text string) []int {
	prose := inlineCode.ReplaceAllString(fencedCode.ReplaceAllString(text, " "), " ")
	var numbers []int
	for _, match := range citationMarks.FindAllStringSubmatchIndex(prose, -1) {
		if end := match[1]; end < len(prose) && (prose[end] == '(' || prose[end] == ':') {
			continue
		}
		if number, err := strconv.Atoi(prose[match[2]:match[3]]); err == nil {
			numbers = append(numbers, number)
		}
	}
	return numbers
}

func CitedSources(text string, citations int) []int {
	var cited []int
	for _, number := range Markers(text) {
		if number >= 1 && number <= citations {
			cited = append(cited, number)
		}
	}
	return cited
}

func markerOutOfRange(text string, citations int) (int, bool) {
	for _, number := range Markers(text) {
		if number < 1 || number > citations {
			return number, true
		}
	}
	return 0, false
}

func describeCitations(citations int) string {
	if citations == 1 {
		return "1 citation"
	}
	return strconv.Itoa(citations) + " citations"
}
