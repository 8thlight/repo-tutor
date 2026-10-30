package tutor_test

import (
	"reflect"
	"testing"

	"repo-tutor/internal/tutor"
)

func TestMarkersAreFoundAfterTheClaimsTheySupport(t *testing.T) {
	for name, test := range map[string]struct {
		text string
		want []int
	}{
		"one marker":                             {"Execution starts in main.go [1].", []int{1}},
		"two sources for one claim":              {"It loads and saves the shelf [1][2].", []int{1, 2}},
		"markers in order of use":                {"First [2] then [1].", []int{2, 1}},
		"no markers":                             {"Execution starts at main.go:2.", nil},
		"inline code is not a marker":            {"The slice `items[1]` holds the second book.", nil},
		"fenced code is not a marker":            {"Look at this:\n```go\nbook := books[1]\n```\nThen [1].", []int{1}},
		"a link is not a marker":                 {"See [1](https://example.com).", nil},
		"a reference definition is not a marker": {"[1]: https://example.com", nil},
		"a number above nine is still read":      {"This claim [12] is out of range.", []int{12}},
		"marker at the end of the text":          {"The tool saves the list [3]", []int{3}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tutor.Markers(test.text); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Markers(%q) = %v, want %v", test.text, got, test.want)
			}
		})
	}
}

func TestCitedSourcesKeepsOnlyMarkersThatHaveACitation(t *testing.T) {
	if got := tutor.CitedSources("A [1] B [3] C [2] D [0].", 2); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("got %v, want [1 2]", got)
	}
}

func TestReplaceMarkersChangesOnlyMarkersThatHaveACitation(t *testing.T) {
	wrap := func(number int) string { return "<" + string(rune('0'+number)) + ">" }
	for name, test := range map[string]struct {
		text      string
		citations int
		want      string
	}{
		"each marker":                     {"A [1] and B [2].", 2, "A <1> and B <2>."},
		"adjacent markers":                {"Both [1][2].", 2, "Both <1><2>."},
		"a number with no citation stays": {"Odd [3].", 2, "Odd [3]."},
		"inline code is kept":             {"Use `items[1]` here [1].", 1, "Use `items[1]` here <1>."},
		"fenced code is kept":             {"```\nx[1]\n```\nThen [1].", 1, "```\nx[1]\n```\nThen <1>."},
		"a link is kept":                  {"See [1](https://example.com) [1].", 1, "See [1](https://example.com) <1>."},
		"no markers":                      {"Plain text.", 1, "Plain text."},
		"a backtick inside a fence does not pair with one outside": {"```\na`b\n```\nSee `c` and [1].", 1, "```\na`b\n```\nSee `c` and <1>."},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tutor.ReplaceMarkers(test.text, test.citations, wrap); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}
