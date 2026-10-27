package terminal

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/about"
	"repo-tutor/internal/repository"
	"repo-tutor/internal/tutor"
)

const interfaceHelp = `Start with one idea, then try explaining it back.

Enter           Send an answer, question, or slash command
Alt+Enter       Add a line to your answer
Ctrl+N          Learn the next topic
Ctrl+G          Start a five-question quiz
/probe           Check what you already know and skip those topics
Ctrl+F          Continue after reviewing an answer
Ctrl+B          Open the quiz question index and goal
Ctrl+K          Browse and search topics
Ctrl+S          Browse cited sources and open the code
Ctrl+O          Open hotspots: churn and branching
v               Switch a Markdown file between rendered and source
Ctrl+P          See lessons and recorded answer outcomes
Esc             Go back to where you opened this page (its key again also works)
Ctrl+L          Return to the current lesson
Alt+Left/Right  Step through this topic; at its ends, press again for the neighbor
PgUp / PgDown   Scroll the lesson or source
Tab / Shift+Tab Switch between Home, learning, and browsing
F1              Show this help
F2              Show the version, copyright, and credits
Ctrl+R          Retry a failed turn, keeping your draft
Esc / Ctrl+C    Stop a running turn
Ctrl+Q          Save and quit (Ctrl+C also quits when idle)

In lists: arrows move, / filters, Enter confirms a filter,
then Enter opens the selected item. Esc goes back.

Slash commands from the plain interface also work:
/topics [filter]  /topic NUMBER|PATH  /next  /quiz  /review
/continue  /questions  /extend  /refresh  /home
/sources  /read PATH:LINE  /hotspots  /progress  /glossary
/map  /help  /quit

Completed replies save before display. A failed or stopped turn
does not change your recorded outcomes. Answers are assessed by
your chosen agent; reading a lesson does not count as mastery.

Use --plain for line input, terminal scrollback, or a screen reader.
Piped input and output use plain mode automatically.
`

type browserItem struct {
	title, description, value string
	line                      int
}

func (item browserItem) Title() string       { return tutor.Plain(item.title) }
func (item browserItem) Description() string { return tutor.Plain(item.description) }
func (item browserItem) FilterValue() string { return tutor.Plain(item.title + " " + item.description) }

func (model interfaceModel) isPicker() bool {
	return model.page == "topics" || model.page == "sources" || model.page == "hotspots" || model.page == "quiz"
}

func (model interfaceModel) openPicker(page, filter string) (tea.Model, tea.Cmd) {
	model.page = page
	model.input.Blur()
	model.configurePicker(filter)
	if filter == "" {
		model.selectCurrent()
	}
	return model, nil
}

func (model *interfaceModel) selectCurrent() {
	state := model.application.state
	for index, current := range model.picker.VisibleItems() {
		item := current.(browserItem)
		switch {
		case model.page == "topics" && item.value == state.Topic:
			model.picker.Select(index)
			return
		case model.page == "quiz" && state.Quiz != nil && item.line == state.Quiz.Active:
			model.picker.Select(index)
			return
		}
	}
}

func (model *interfaceModel) configurePicker(filter string) {
	selected, selectedBefore := model.picker.SelectedItem().(browserItem)
	filterState := model.picker.FilterState()
	var items []list.Item
	switch model.page {
	case "topics":
		model.pageTitle = "Choose your next topic"
		for index, topic := range model.application.course {
			progress := model.application.state.Progress[topic.ID]
			status := "Not studied yet"
			if progress.Lessons > 0 || progress.Correct+progress.Partial+progress.Missed > 0 {
				status = fmt.Sprintf("%d lessons · %d correct · %d partial · %d incorrect", progress.Lessons, progress.Correct, progress.Partial, progress.Missed)
			}
			if topic.ID == model.application.state.Topic {
				status = "Current topic · " + status
			}
			items = append(items, browserItem{fmt.Sprintf("%d  %s", index+1, tutor.Plain(topic.Title)), status, topic.ID, 0})
		}
	case "quiz":
		model.pageTitle = "Quiz questions"
		for _, slot := range quizSlotItems(model.application.state) {
			items = append(items, slot)
		}
	case "sources":
		model.pageTitle = "Sources · Follow the evidence"
		seen := map[string]bool{}
		if len(model.application.state.Turns) > 0 {
			turn := model.application.state.Turns[max(0, model.history)]
			for _, citation := range turn.Reply.Citations {
				if !seen[citation.Path] {
					items = append(items, browserItem{citation.Path, fmt.Sprintf("Cited in this lesson · line %d", citation.Line), citation.Path, citation.Line})
					seen[citation.Path] = true
				}
			}
		}
		prepared := tutor.Prepare(model.application.snapshot, tutor.Request{Topic: model.application.state.Topic})
		for _, excerpt := range prepared.Excerpts {
			if !seen[excerpt.Path] {
				items = append(items, browserItem{excerpt.Path, fmt.Sprintf("Available excerpt · lines %d–%d", excerpt.StartLine, excerpt.EndLine), excerpt.Path, excerpt.StartLine})
				seen[excerpt.Path] = true
			}
		}
	case "hotspots":
		model.pageTitle = "Hotspots · Where to pay attention"
		for _, hotspot := range model.application.snapshot.Hotspots {
			description := fmt.Sprintf("%d recent commits · %d branch points · %d lines", hotspot.Commits, hotspot.Branches, hotspot.Lines)
			if !model.application.snapshot.History {
				description = fmt.Sprintf("%d branch points · %d lines · Git history unavailable", hotspot.Branches, hotspot.Lines)
			}
			items = append(items, browserItem{hotspot.Path, description, hotspot.Path, hotspot.Line})
		}
	}
	palette := model.palette()
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(1)
	delegate.Styles = list.NewDefaultItemStyles(model.dark)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(palette.muted)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(palette.accent).BorderForeground(palette.accent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(palette.muted).BorderForeground(palette.accent)
	model.picker = list.New(items, delegate, model.bodyWidth(), max(1, model.bodyHeight()-2))
	model.picker.SetShowTitle(false)
	model.picker.SetShowHelp(false)
	model.picker.DisableQuitKeybindings()
	model.picker.Styles = list.DefaultStyles(model.dark)
	model.picker.Styles.StatusBar = lipgloss.NewStyle().Foreground(palette.muted).PaddingBottom(1)
	model.picker.Styles.StatusBarFilterCount = lipgloss.NewStyle().Foreground(palette.muted)
	model.picker.Styles.NoItems = lipgloss.NewStyle().Foreground(palette.muted)
	if filter != "" {
		model.picker.SetFilterText(filter)
		if filterState == list.Filtering {
			model.picker.SetFilterState(list.Filtering)
		}
	}
	if selectedBefore {
		for index, current := range model.picker.VisibleItems() {
			if item := current.(browserItem); item.value == selected.value {
				model.picker.Select(index)
				break
			}
		}
	}
}

func (model interfaceModel) selectItem() (tea.Model, tea.Cmd) {
	item, found := model.picker.SelectedItem().(browserItem)
	if !found {
		return model, nil
	}
	if model.page == "topics" {
		return model.openTopic(item.value)
	}
	if model.page == "quiz" {
		return model.selectQuizSlot(item.line)
	}
	return model.openFile(repository.Location{Path: item.value, Line: item.line})
}

func (model interfaceModel) openSource(location repository.Location) (tea.Model, tea.Cmd) {
	return model.openFile(location)
}

func (model interfaceModel) openFile(location repository.Location) (tea.Model, tea.Cmd) {
	switch model.page {
	case "sources", "hotspots":
		model.sourceList = model.page
	case "source":
	default:
		model.sourceList = ""
	}
	file, found := model.application.snapshot.File(location.Path)
	if !found {
		model.notice = "This file is not in the source snapshot."
		return model, nil
	}
	if isMarkdown(file.Path) && model.rendersMarkdown(file.Path) {
		return model.showRendered(file, location.Line)
	}
	return model.showSource(file, location)
}

func (model interfaceModel) showSource(file repository.File, location repository.Location) (tea.Model, tea.Cmd) {
	lines := strings.Split(strings.TrimSuffix(file.Content, "\n"), "\n")
	if location.Line < 1 || location.Line > len(lines) {
		model.notice = fmt.Sprintf("Choose a line between 1 and %d.", len(lines))
		return model, nil
	}
	highlighted := model.highlightSource(location.Path, file.Content)
	var content strings.Builder
	palette := model.palette()
	rowsAbove := 0
	for index, line := range highlighted {
		prefix := fmt.Sprintf(" %5d  ", index+1)
		if index+1 == location.Line {
			prefix = lipgloss.NewStyle().Foreground(palette.accent).Bold(true).Render(fmt.Sprintf("›%5d  ", index+1))
		}
		content.WriteString(prefix + line + "\n")
		if index+1 < location.Line {
			rowsAbove += max(1, (ansi.StringWidth(prefix+line)+model.bodyWidth()-1)/model.bodyWidth())
		}
	}
	model, command := model.document("source", fmt.Sprintf("%s:%d", location.Path, location.Line), content.String())
	model.viewport.SetYOffset(max(0, rowsAbove-sourceContextRows))
	model.sourcePath, model.sourceLine, model.sourceRendered = file.Path, location.Line, false
	model.notice = "Source snapshot · /read PATH:LINE opens another location · Esc goes back"
	if isMarkdown(file.Path) {
		model.notice = "Source snapshot · v shows the rendered page · Esc goes back"
	}
	return model, command
}

func isMarkdown(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return extension == ".md" || extension == ".markdown"
}

func (model interfaceModel) rendersMarkdown(path string) bool {
	if rendered, chosen := model.markdownViews[path]; chosen {
		return rendered
	}
	return true
}

func (model interfaceModel) rememberMarkdownView(path string, rendered bool) interfaceModel {
	views := make(map[string]bool, len(model.markdownViews)+1)
	for name, chosen := range model.markdownViews {
		views[name] = chosen
	}
	views[path] = rendered
	model.markdownViews = views
	return model
}

func (model interfaceModel) markdownPage(file repository.File, line int) (string, int, bool) {
	rendered := model.markdown(file.Content)
	rows := strings.Split(rendered, "\n")
	row, found := renderedRow(strings.Split(file.Content, "\n"), line, rendered)
	if !found {
		row = firstTextRow(rows)
	}
	marker := lipgloss.NewStyle().Foreground(model.palette().accent).Bold(true).Render("› ")
	for index := range rows {
		gutter := "  "
		if index == row {
			gutter = marker
		}
		rows[index] = gutter + trimPadding(rows[index])
	}
	return strings.Join(rows, "\n"), row, found
}

func firstTextRow(rows []string) int {
	for index, row := range rows {
		if strings.TrimSpace(ansi.Strip(row)) != "" {
			return index
		}
	}
	return 0
}

func (model interfaceModel) showRendered(file repository.File, line int) (tea.Model, tea.Cmd) {
	page, row, found := model.markdownPage(file, line)
	model, command := model.document("source", file.Path+" · rendered", page)
	model.sourcePath, model.sourceLine, model.sourceRendered = file.Path, line, true
	model.notice = fmt.Sprintf("Line %d · could not place it, so the top is marked with › · v shows the source · Esc goes back", line)
	if found {
		model.viewport.SetYOffset(max(0, row-sourceContextRows/4))
		model.notice = fmt.Sprintf("Line %d · marked with › · v shows the source at that line · Esc goes back", line)
	}
	return model, command
}

func (model interfaceModel) toggleMarkdownView() (tea.Model, tea.Cmd) {
	file, found := model.application.snapshot.File(model.sourcePath)
	if !found {
		return model, nil
	}
	if model.sourceRendered {
		model = model.rememberMarkdownView(file.Path, false)
		return model.showSource(file, repository.Location{Path: file.Path, Line: max(1, model.sourceLine)})
	}
	model = model.rememberMarkdownView(file.Path, true)
	return model.showRendered(file, model.sourceLine)
}

func (model interfaceModel) openDocument(page, title, content string) (tea.Model, tea.Cmd) {
	return model.document(page, title, content)
}
func (model interfaceModel) document(page, title, content string) (interfaceModel, tea.Cmd) {
	model.page, model.pageTitle, model.pageContent = page, title, content
	model.input.Blur()
	model.rebuildContent()
	model.viewport.GotoTop()
	return model, nil
}

const sourceContextRows = 12

const splashMinimumWidth = 113

func (model interfaceModel) showsSplash() bool {
	return model.richColor && !model.noColor && model.width >= splashMinimumWidth
}

func (model interfaceModel) aboutText() string {
	state := model.application.state
	modelName := state.Model
	if modelName == "" {
		modelName = "default model"
	}
	text := about.Read().Text() + fmt.Sprintf("\nThis session uses %s (%s).\nBuilt with Bubble Tea, Bubbles, Lip Gloss, Glamour, and Chroma, which are MIT licensed.\nrepo-tutor --version --notices prints every third-party license.\n", state.Agent, modelName)
	if !model.showsSplash() {
		return text
	}
	return about.Splash() + "\x1b[0m\n" + ansi.Wrap(about.SplashCaption, about.SplashWidth, "") + "\n\n" + text
}

func (model interfaceModel) progressText() string {
	var text strings.Builder
	for _, topic := range model.application.course {
		progress := model.application.state.Progress[topic.ID]
		if progress == (tutor.Progress{}) {
			continue
		}
		title := topic.Title
		if progress.Held {
			title += " (known)"
		}
		fmt.Fprintf(&text, "%s\n  %d lessons   %d correct   %d partial   %d incorrect\n\n", title, progress.Lessons, progress.Correct, progress.Partial, progress.Missed)
	}
	if text.Len() == 0 {
		text.WriteString("Your first lesson is a good place to start.\n\n")
	}
	text.WriteString("Reading counts as a lesson. Only assessed answers count toward quiz outcomes.\nUse /review to revisit a topic with partial or incorrect answers.\n")
	return text.String()
}

func (model interfaceModel) next() (tea.Model, tea.Cmd) {
	if topic, found := tutor.NextUnheld(model.application.course, model.application.state.Progress, model.application.state.Topic); found {
		return model.openTopic(topic.ID)
	}
	model.notice = "You reached the last topic. Try a quiz, choose another topic, or use /review."
	return model, nil
}

func (model interfaceModel) cyclePage(backward bool) (tea.Model, tea.Cmd) {
	tabs := []string{"home", "course", "quiz", "progress"}
	selected := 0
	for index, tab := range tabs {
		if tab == model.activeTab() {
			selected = index
		}
	}
	step := 1
	if backward {
		step = -1
	}
	switch tabs[(selected+step+len(tabs))%len(tabs)] {
	case "home":
		return model.openHome()
	case "course":
		return model.learn()
	case "quiz":
		return model.openPicker("quiz", "")
	}
	return model.openDocument("progress", "Progress", model.progressText())
}

func (model interfaceModel) slash(text string) (tea.Model, tea.Cmd) {
	parts := strings.SplitN(text, " ", 2)
	argument := ""
	if len(parts) > 1 {
		argument = strings.TrimSpace(parts[1])
	}
	switch parts[0] {
	case "/continue", "/extend":
		model.input.Reset()
		return model.continueQuiz(parts[0] == "/extend")
	case "/home":
		model.input.Reset()
		return model.openHome()
	case "/refresh":
		model.input.Reset()
		return model.begin(turnIntent{"lesson", openingMessage, tutor.Overview})
	case "/probe":
		intent, err := probeIntent(model.application.state, model.application.course)
		if err != nil {
			model.notice = err.Error()
			return model, nil
		}
		model.input.Reset()
		return model.begin(intent)
	case "/questions":
		model.input.Reset()
		return model.openPicker("quiz", "")
	case "/quit", "/exit":
		return model, tea.Quit
	case "/topics", "/sources", "/hotspots":
		model.input.Reset()
		return model.openPicker(strings.TrimPrefix(parts[0], "/"), argument)
	case "/topic":
		topic, err := model.application.findTopic(argument)
		if err != nil {
			model.notice = err.Error()
			return model, nil
		}
		model.input.Reset()
		return model.begin(turnIntent{"lesson", "Teach the next small idea in this topic.", topic.ID})
	case "/next":
		model.input.Reset()
		return model.next()
	case "/quiz":
		model.input.Reset()
		return model.begin(turnIntent{"quiz", "Ask one practical comprehension question about this topic.", model.application.state.Topic})
	case "/review":
		selected, highest := "", 0
		for _, topic := range model.application.course {
			progress := model.application.state.Progress[topic.ID]
			if score := progress.Missed + progress.Partial - progress.Correct; score > highest {
				selected, highest = topic.ID, score
			}
		}
		if selected == "" {
			model.notice = "No topic currently needs review. Try a quiz or continue with Ctrl+N."
			return model, nil
		}
		model.input.Reset()
		return model.begin(turnIntent{"lesson", "Review this topic with a different example, focusing on earlier confusion.", selected})
	case "/help":
		model.input.Reset()
		return model.openDocument("help", "A few ways to keep learning", interfaceHelp)
	case "/progress":
		model.input.Reset()
		return model.openDocument("progress", "Progress", model.progressText())
	}
	var output bytes.Buffer
	local := model.application
	local.output = &output
	if _, err := local.command(model.ctx, text); err != nil {
		model.notice = tutor.Plain(err.Error())
		return model, nil
	}
	model.input.Reset()
	return model.openDocument("document", strings.TrimPrefix(text, "/"), output.String())
}
