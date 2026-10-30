package tutor

import (
	"regexp"
	"strconv"
	"strings"
)

const MaxCitations = 9

var (
	fencedCode    = regexp.MustCompile("(?s)```.*?```")
	inlineCode    = regexp.MustCompile("`[^`\n]*`")
	citationMarks = regexp.MustCompile(`\[(\d+)\]`)
)

type markerSpan struct{ start, end, number int }

func markerSpans(text string) []markerSpan {
	masked := text
	var codeRanges [][]int
	for _, found := range fencedCode.FindAllStringIndex(text, -1) {
		codeRanges = append(codeRanges, found)
		masked = masked[:found[0]] + strings.Repeat(" ", found[1]-found[0]) + masked[found[1]:]
	}
	codeRanges = append(codeRanges, inlineCode.FindAllStringIndex(masked, -1)...)
	inCode := func(position int) bool {
		for _, found := range codeRanges {
			if position >= found[0] && position < found[1] {
				return true
			}
		}
		return false
	}
	var spans []markerSpan
	for _, match := range citationMarks.FindAllStringSubmatchIndex(text, -1) {
		if inCode(match[0]) {
			continue
		}
		if end := match[1]; end < len(text) && (text[end] == '(' || text[end] == ':') {
			continue
		}
		if number, err := strconv.Atoi(text[match[2]:match[3]]); err == nil {
			spans = append(spans, markerSpan{match[0], match[1], number})
		}
	}
	return spans
}

func Markers(text string) []int {
	var numbers []int
	for _, span := range markerSpans(text) {
		numbers = append(numbers, span.number)
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

func ReplaceMarkers(text string, citations int, replace func(number int) string) string {
	var output strings.Builder
	last := 0
	for _, span := range markerSpans(text) {
		if span.number < 1 || span.number > citations {
			continue
		}
		output.WriteString(text[last:span.start])
		output.WriteString(replace(span.number))
		last = span.end
	}
	output.WriteString(text[last:])
	return output.String()
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
