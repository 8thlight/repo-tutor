package terminal

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

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
	theme := styles.Get("github-dark")
	if !model.dark {
		theme = styles.Get("github")
	}
	var highlighted strings.Builder
	for token := iterator(); token != chroma.EOF; token = iterator() {
		entry := theme.Get(token.Type)
		style := lipgloss.NewStyle().Bold(entry.Bold == chroma.Yes).Italic(entry.Italic == chroma.Yes)
		if entry.Colour.IsSet() {
			style = style.Foreground(lipgloss.Color(entry.Colour.String()))
		}
		for index, line := range strings.Split(token.Value, "\n") {
			if index > 0 {
				highlighted.WriteByte('\n')
			}
			highlighted.WriteString(style.Render(line))
		}
	}
	return strings.Split(highlighted.String(), "\n")
}
