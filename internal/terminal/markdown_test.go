package terminal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
)

func TestRenderedRowFindsTheCitedTextDespiteMarkup(t *testing.T) {
	source := strings.Split("# Install guide\n\n- Run **just check** before you commit.\n\nSee the [architecture notes](docs/a.md) for details.\n\nRun just check before you commit.\n\n| a | b |\n|---|---|\n| - | - |\n", "\n")
	rendered := "\x1b[1mInstall guide\x1b[0m\n\n  • Run just check before you commit.\n\nSee the architecture notes for details.\n\nRun just check before you commit.\n"
	for name, test := range map[string]struct {
		line  int
		row   int
		found bool
	}{
		"heading":           {1, 0, true},
		"list with bold":    {3, 2, true},
		"link text":         {5, 4, true},
		"repeated text":     {7, 6, true},
		"table of symbols":  {11, 0, false},
		"blank line":        {2, 0, false},
		"beyond the source": {99, 0, false},
	} {
		row, found := renderedRow(source, test.line, rendered)
		if found != test.found || (found && row != test.row) {
			t.Fatalf("%s: got row %d found %v, want row %d found %v", name, row, found, test.row, test.found)
		}
	}
}

func TestCitationScrollsRenderedMarkdownToItsText(t *testing.T) {
	model := interfaceFixture()
	var document strings.Builder
	for section := 1; section <= 60; section++ {
		fmt.Fprintf(&document, "## Section %d\n\nParagraph number %d explains one idea.\n\n", section, section)
	}
	model.application.snapshot.Files = append(model.application.snapshot.Files, repository.File{Path: "docs/guide.md", Content: document.String()})
	cited, _ := model.openSource(repository.Location{Path: "docs/guide.md", Line: 3 + 4*40})
	model = cited.(interfaceModel)
	if model.viewport.YOffset() == 0 || !strings.Contains(ansi.Strip(model.viewport.View()), "Paragraph number 41 explains one idea.") || !strings.Contains(model.notice, "marked with ›") {
		t.Fatalf("the rendered view should scroll to the cited paragraph: offset %d, notice %q", model.viewport.YOffset(), model.notice)
	}
	var marked []string
	for _, row := range strings.Split(ansi.Strip(model.pageContent), "\n") {
		if strings.HasPrefix(row, "› ") {
			marked = append(marked, row)
		}
	}
	if len(marked) != 1 || !strings.Contains(marked[0], "Paragraph number 41 explains one idea.") {
		t.Fatalf("exactly the cited row should carry the gutter marker: %q", marked)
	}
	opened, _ := interfaceFixture().openSource(repository.Location{Path: "README.md", Line: 1})
	short := opened.(interfaceModel)
	if short.viewport.YOffset() != 0 {
		t.Fatal("a short document should stay at the top")
	}
}

func TestRenderedMarkdownHasNoMarkerWhenNothingMatches(t *testing.T) {
	model := interfaceFixture()
	model.application.snapshot.Files = append(model.application.snapshot.Files, repository.File{Path: "rule.md", Content: "# Title\n\n***\n"})
	opened, _ := model.openSource(repository.Location{Path: "rule.md", Line: 3})
	page := opened.(interfaceModel)
	if strings.Contains(ansi.Strip(page.pageContent), "› ") || !strings.Contains(page.notice, "could not place it") {
		t.Fatalf("an unmatched citation should show no marker: %q", page.notice)
	}
}
