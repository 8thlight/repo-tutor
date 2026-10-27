package settings

import (
	"bytes"
	"encoding/json"
	"slices"
)

const Version = 1

const (
	TabStyleNone    = "none"
	TabStyleBlock   = "block"
	TabStyleSlanted = "slanted"
	TabStyleRounded = "rounded"

	TabStyleJoinedSlanted = "joined_slanted"
	TabStyleJoinedRounded = "joined_rounded"
	TabStyleJoinedArrows  = "joined_arrows"

	HighlightGreen = "green"
)

var (
	TabStyles  = []string{TabStyleNone, TabStyleBlock, TabStyleSlanted, TabStyleRounded, TabStyleJoinedSlanted, TabStyleJoinedRounded, TabStyleJoinedArrows}
	Highlights = []string{"red", "orange", "yellow", HighlightGreen, "blue", "dark_blue", "purple"}
)

type Settings struct {
	TabStyle              string
	MarkdownOpensRendered bool
	Highlight             string
}

type Status struct {
	Unreadable bool
	ReadOnly   bool
}

type file struct {
	Version               int     `json:"version"`
	TabStyle              *string `json:"tab_style,omitempty"`
	MarkdownOpensRendered *bool   `json:"markdown_opens_rendered,omitempty"`
	Highlight             *string `json:"highlight,omitempty"`
}

func Defaults() Settings {
	return Settings{TabStyle: TabStyleNone, MarkdownOpensRendered: true, Highlight: HighlightGreen}
}

func Decode(data []byte) (Settings, Status) {
	settings := Defaults()
	var decoded file
	if err := json.Unmarshal(data, &decoded); err != nil {
		return settings, Status{Unreadable: true}
	}
	if decoded.TabStyle != nil && slices.Contains(TabStyles, *decoded.TabStyle) {
		settings.TabStyle = *decoded.TabStyle
	}
	if decoded.MarkdownOpensRendered != nil {
		settings.MarkdownOpensRendered = *decoded.MarkdownOpensRendered
	}
	if decoded.Highlight != nil && slices.Contains(Highlights, *decoded.Highlight) {
		settings.Highlight = *decoded.Highlight
	}
	return settings, Status{ReadOnly: decoded.Version > Version}
}

func Encode(settings Settings) []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "  ")
	encoder.Encode(file{Version: Version, TabStyle: &settings.TabStyle, MarkdownOpensRendered: &settings.MarkdownOpensRendered, Highlight: &settings.Highlight})
	return buffer.Bytes()
}
