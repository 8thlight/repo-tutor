package terminal

import (
	"fmt"
	"path/filepath"
	"sort"
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

Enter           Send your question or answer
Ctrl+N          Next: the next topic, or the next question during a quiz (twice to skip an open one)
Ctrl+G          Start a quiz on this topic (length is a setting)
Ctrl+F          More: more on this topic, or more questions after a quiz
Ctrl+B          Open the quiz question index and goal
Ctrl+K          Browse and search topics
Ctrl+S          Browse cited sources and open the code
v               Switch a Markdown file between rendered and source
T               Teach the open source file (starts a lesson on it)
Ctrl+P          See lessons and recorded answer outcomes
Esc             Go back to where you opened this page (its key again also works)
Ctrl+L          Return to the current lesson (leaves a quiz after a second press)
Alt+Left/Right  Step through this topic; at its ends, press again for the neighbor
PgUp / PgDown   Scroll the lesson or source
Tab / Shift+Tab Switch between Home, Course, Quiz, and Progress
F1              Show this help
F2              Change the tab style, highlight color, quiz length, and Markdown view
F3              Show the version, copyright, and credits
Ctrl+R          Retry a failed request, keeping your draft
Esc / Ctrl+C    Stop what is loading
Ctrl+Q          Save and quit (Ctrl+C also quits when idle)

In lists: arrows move, / filters, Enter confirms a filter,
then Enter opens the selected item. Esc goes back.

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
	return model.page == "topics" || model.page == "sources" || model.page == "quiz"
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
		files := append([]repository.File(nil), model.application.snapshot.Files...)
		sort.Slice(files, func(first, second int) bool { return files[first].Path < files[second].Path })
		for _, file := range files {
			if !seen[file.Path] && file.Content != "" {
				lines := strings.Count(strings.TrimSuffix(file.Content, "\n"), "\n") + 1
				items = append(items, browserItem{file.Path, fmt.Sprintf("Repository file · %d lines", lines), file.Path, 1})
				seen[file.Path] = true
			}
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
	case "sources":
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
	model.notice = ""
	if isMarkdown(file.Path) {
		model.notice = ""
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
	return model.application.settings.MarkdownOpensRendered
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
	model.notice = fmt.Sprintf("Line %d · could not place it, so the top is marked with ›", line)
	if found {
		model.viewport.SetYOffset(max(0, row-sourceContextRows/4))
		model.notice = fmt.Sprintf("Line %d · marked with ›", line)
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
	state := model.application.state
	score := tutor.ScoreCourse(model.application.course, state.Progress)
	fmt.Fprintf(&text, "Goal\n%s\n\n", ansi.Wrap(tutor.Plain(state.Goal), model.bodyWidth(), ""))
	fmt.Fprintf(&text, "Coverage        %d of %d topics read\nUnderstanding   %d%% across the topics you have read\n", score.Read, score.Total, score.Percent)
	if exam := state.LastExam; exam != nil {
		if exam.Planned > exam.Goal {
			fmt.Fprintf(&text, "Latest exam     %d%% · %s (left early: %d of %d questions)\n", exam.Percent(), tutor.Band(exam.Percent()), exam.Goal, exam.Planned)
		} else {
			fmt.Fprintf(&text, "Latest exam     %d%% · %s (%d questions)\n", exam.Percent(), tutor.Band(exam.Percent()), exam.Goal)
		}
	}
	text.WriteString("\n")
	listed := false
	for _, topic := range model.application.course {
		progress := state.Progress[topic.ID]
		percent, scored := tutor.Understanding(progress)
		if progress.Lessons == 0 && !scored {
			continue
		}
		listed = true
		reading := "not read"
		if progress.Lessons > 0 {
			reading = "read"
		}
		result := "no quiz yet"
		switch {
		case progress.QuizGoal > 0:
			result = fmt.Sprintf("quiz %d%% · %s", percent, tutor.Band(percent))
		case scored:
			result = fmt.Sprintf("known from the probe · %d%%", percent)
		}
		if quiz := state.Quiz; quiz != nil && quiz.Topic == topic.ID && !quiz.Finished() {
			result += fmt.Sprintf(" · quiz in progress (%d of %d)", quiz.Completed(), quiz.Goal)
		}
		fmt.Fprintf(&text, "%s\n  %s · %s\n\n", topic.Title, reading, result)
	}
	if !listed {
		text.WriteString("Your first lesson is a good place to start.\n\n")
	}
	if exam := state.LastExam; exam != nil {
		text.WriteString("Latest exam, by topic\n")
		for _, result := range exam.Topics {
			outcome := map[int]string{2: "correct", 1: "partial", 0: "missed"}[result.Credit]
			fmt.Fprintf(&text, "  %s · %s\n", model.titleOf(result.Topic), outcome)
		}
		text.WriteString("\n")
	}
	if tutor.ExamAvailable(model.application.course, state.Progress) {
		fmt.Fprintf(&text, "Take an exam from Home: %d questions from the topics you have read.\n\n", model.examPlanLength())
	}
	text.WriteString("Reading is not graded. The latest finished quiz on a topic sets its score, so a retake can raise it.\nA topic under 60% shows up on Home as one to review.\n")
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
