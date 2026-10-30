package terminal

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
)

func beforeEvidence(content string) string {
	before, _, _ := strings.Cut(content, "── EVIDENCE")
	return before
}

func fixtureWithExplanation(explanation string) interfaceModel {
	model := interfaceFixture()
	model.noColor = false
	model.application.state.Turns[0].Reply.Explanation = explanation
	model.rebuildContent()
	return model
}

func TestMarkersDrawInTheAccentColorAndBold(t *testing.T) {
	model := fixtureWithExplanation("Execution starts in main.go [1].")
	styled := lipgloss.NewStyle().Foreground(model.palette().accent).Bold(true).Render("[1]")
	if content := model.lessonContent(); !strings.Contains(content, styled) {
		t.Fatalf("the marker should be styled, got %q", content)
	}
}

func TestACodeSpanThatLooksLikeAMarkerIsNotStyled(t *testing.T) {
	model := fixtureWithExplanation("The slice `items[1]` holds a book [1].")
	content := beforeEvidence(model.lessonContent())
	if strings.Count(content, lipgloss.NewStyle().Foreground(model.palette().accent).Bold(true).Render("[1]")) != 1 {
		t.Fatalf("only the real marker should be styled: %q", content)
	}
	if !strings.Contains(ansi.Strip(content), "items[1]") {
		t.Fatalf("the code span should be kept: %q", ansi.Strip(content))
	}
}

func TestNoPlaceholderCharacterReachesTheScreen(t *testing.T) {
	content := beforeEvidence(fixtureWithExplanation("One [1] and two [1][1] markers.").lessonContent())
	if strings.ContainsAny(content, markerOpen+markerClose) {
		t.Fatalf("a placeholder leaked: %q", content)
	}
	if got := strings.Count(ansi.Strip(content), "[1]"); got != 3 {
		t.Fatalf("every marker should still read [1], found %d in %q", got, ansi.Strip(content))
	}
}

func TestAMarkerWithNoCitationStaysPlainText(t *testing.T) {
	model := fixtureWithExplanation("Odd [4].")
	if content := model.lessonContent(); strings.Contains(content, lipgloss.NewStyle().Foreground(model.palette().accent).Bold(true).Render("[4]")) {
		t.Fatal("a marker with no citation should not be styled as one")
	}
}

func TestMarkersKeepTheirWidthWhenTheLessonWraps(t *testing.T) {
	long := strings.Repeat("The shelf is loaded and saved in main.go [1] and then read again. ", 6)
	content := ansi.Strip(fixtureWithExplanation(long).lessonContent())
	for _, line := range strings.Split(content, "\n") {
		if ansi.StringWidth(line) > 100 {
			t.Fatalf("a wrapped line is too wide: %q", line)
		}
	}
}

func TestEvidenceIsARuledSectionWithOneNumberedRowPerCitation(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	model.application.state.Turns[0].Reply.Citations = []repository.Location{{Path: "main.go", Line: 2}, {Path: "main.go", Line: 9}, {Path: "README.md", Line: 1}}
	content := ansi.Strip(model.lessonContent())
	for _, expected := range []string{"── EVIDENCE · 3 sources ─", "[1]  main.go:2", "[2]  main.go:9", "[3]  README.md:1"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("the evidence block is missing %q:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "Ctrl+S to open source") {
		t.Fatal("the old hint line should be gone")
	}
}

func TestOneCitationReadsAsOneSource(t *testing.T) {
	if content := ansi.Strip(interfaceFixture().lessonContent()); !strings.Contains(content, "EVIDENCE · 1 source ─") {
		t.Fatalf("a single citation should read as 1 source:\n%s", content)
	}
}

func TestTheEvidencePathIsPlainAndTheLineNumberFaint(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	content := model.lessonContent()
	if !strings.Contains(content, "main.go"+lipgloss.NewStyle().Faint(true).Render(":2")) {
		t.Fatalf("the path should be unstyled and the line faint: %q", content)
	}
}
