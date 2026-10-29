package terminal

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"

	"repo-tutor/internal/tutor"
)

func (model interfaceModel) highlightSource(path, content string) []string {
	content = strings.ReplaceAll(tutor.Plain(strings.TrimSuffix(content, "\n")), "\t", "    ")
	plain := strings.Split(content, "\n")
	if model.noColor {
		return plain
	}
	lexer := lexers.Match(path)
	if lexer == nil {
		return plain
	}
	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, content)
	if err != nil {
		return plain
	}
	var highlighted strings.Builder
	for token := iterator(); token != chroma.EOF; token = iterator() {
		style := sourceTokenStyle(token.Type)
		for index, line := range strings.Split(token.Value, "\n") {
			if index > 0 {
				highlighted.WriteByte('\n')
			}
			highlighted.WriteString(style.Render(line))
		}
	}
	return strings.Split(highlighted.String(), "\n")
}

const (
	ansiRed     = 1
	ansiGreen   = 2
	ansiMagenta = 5
	ansiCyan    = 6
)

func sourceTokenStyle(kind chroma.TokenType) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch {
	case kind.InCategory(chroma.Comment):
		return style.Faint(true).Italic(true)
	case kind.InCategory(chroma.Keyword):
		return style.Foreground(lipgloss.ANSIColor(ansiMagenta))
	case kind.InSubCategory(chroma.LiteralString):
		return style.Foreground(lipgloss.ANSIColor(ansiGreen))
	case kind.InSubCategory(chroma.LiteralNumber), kind == chroma.NameBuiltin, kind == chroma.NameConstant:
		return style.Foreground(lipgloss.ANSIColor(ansiCyan))
	case kind == chroma.NameFunction, kind == chroma.NameClass:
		return style.Foreground(lipgloss.ANSIColor(ansiBlue)).Bold(true)
	case kind.InCategory(chroma.Error):
		return style.Foreground(lipgloss.ANSIColor(ansiRed))
	}
	return style
}
