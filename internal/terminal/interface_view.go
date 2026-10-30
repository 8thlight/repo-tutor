package terminal

import (
	"fmt"
	"image/color"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	glamouransi "charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/tutor"
)

type interfacePalette struct{ accent, warning, cutout color.Color }

func (model interfaceModel) palette() interfacePalette {
	if model.noColor {
		return interfacePalette{}
	}
	return interfacePalette{model.highlight(model.application.settings.Highlight), lipgloss.ANSIColor(ansiYellow), model.background}
}

func (model interfaceModel) bodyWidth() int {
	width := max(1, model.width-4)
	if model.width >= 108 {
		width -= sidebarOffset
	}
	return max(1, width)
}
func (model interfaceModel) bodyHeight() int {
	return max(1, model.height-8-lipgloss.Height(model.composer()))
}

func (model *interfaceModel) resize() {
	palette := model.palette()
	model.input.SetWidth(max(1, model.bodyWidth()-4))
	inputStyles := textarea.DefaultStyles(true)
	withoutStyleColors(&inputStyles)
	inputStyles.Focused.Prompt = lipgloss.NewStyle().Foreground(palette.accent)
	inputStyles.Focused.Text = lipgloss.NewStyle()
	inputStyles.Focused.Placeholder = lipgloss.NewStyle().Faint(true)
	inputStyles.Blurred.Placeholder = lipgloss.NewStyle().Faint(true)
	model.input.SetStyles(inputStyles)
	model.spinner.Style = lipgloss.NewStyle().Foreground(palette.accent)
	model.viewport.SetWidth(model.bodyWidth())
	model.viewport.SetHeight(max(1, model.bodyHeight()-2))
	if model.page == "about" {
		model.pageContent = model.aboutText()
	}
	if file, found := model.application.snapshot.File(model.sourcePath); found && model.page == "source" && model.sourceRendered {
		model.pageContent, _, _ = model.markdownPage(file, model.sourceLine)
	}
	if model.isPicker() {
		model.configurePicker(model.picker.FilterValue())
	}
	model.rebuildContent()
}

func (model *interfaceModel) rebuildContent() {
	model.viewport.SetHeight(max(1, model.bodyHeight()-2))
	if model.isPicker() {
		return
	}
	if model.page == "lesson" {
		model.viewport.SetContent(model.lessonContent())
	} else if model.page == "home" {
		model.viewport.SetContent(model.homeContent())
	} else if model.page == "settings" {
		model.viewport.SetContent(model.settingsContent())
		model.followSettingsFocus()
	} else {
		model.viewport.SetContent(model.pageContent)
	}
}

func (model interfaceModel) markdownStyle() glamouransi.StyleConfig {
	choice := highlightByID(model.application.settings.Highlight)
	configuration := terminalMarkdownStyle(choice.index, choice.index == noHighlight)
	if model.noColor {
		configuration = styles.ASCIIStyleConfig
	}
	zero := uint(0)
	configuration.Document.Margin = &zero
	configuration.Document.BlockPrefix = ""
	configuration.Document.BlockSuffix = ""
	return configuration
}

func (model interfaceModel) markdown(text string) string {
	configuration := model.markdownStyle()
	options := []glamour.TermRendererOption{glamour.WithStyles(configuration), glamour.WithWordWrap(max(12, min(92, model.bodyWidth()-3)))}
	renderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return tutor.Plain(text)
	}
	rendered, err := renderer.Render(tutor.Plain(text))
	if err != nil {
		return tutor.Plain(text)
	}
	return strings.TrimSpace(rendered)
}

const (
	markerOpen  = "\ue000"
	markerClose = "\ue001"
)

var placeholderMarker = regexp.MustCompile(markerOpen + `(\d)` + markerClose)

func (model interfaceModel) markedMarkdown(text string, citations int) string {
	held := tutor.ReplaceMarkers(text, citations, func(number int) string {
		return markerOpen + strconv.Itoa(number) + markerClose
	})
	style := lipgloss.NewStyle().Foreground(model.palette().accent).Bold(true)
	return placeholderMarker.ReplaceAllStringFunc(model.markdown(held), func(found string) string {
		return style.Render("[" + placeholderMarker.FindStringSubmatch(found)[1] + "]")
	})
}

func (model interfaceModel) lessonContent() string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Faint(true)
	heading := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	turns := model.application.state.Turns
	if len(turns) == 0 {
		if model.errorText != "" {
			return heading.Render("Your progress is safe. Let's try again.") + "\n\n" + ansi.Wrap(model.errorText, model.bodyWidth(), "") + "\n\n" + muted.Render("Ctrl+R retries · Ctrl+Q exits · use --plain for line input")
		}
		return heading.Render("A codebase, one idea at a time.") + "\n\n" + model.markdown("Your tutor is preparing an overview grounded in this repository's files.\n\nThen you can ask questions, try a short quiz, or follow a citation into the code.") + "\n\n" + muted.Render("Browse topics with Ctrl+K or sources with Ctrl+S while you wait.")
	}
	index := max(0, min(model.history, len(turns)-1))
	turn := turns[index]
	var blocks []string
	if model.errorText != "" {
		blocks = append(blocks, lipgloss.NewStyle().Foreground(palette.warning).Bold(true).Render("That did not finish · Ctrl+R to retry")+"\n"+ansi.Wrap(model.errorText, model.bodyWidth(), ""))
	}
	if turn.Action == "answer" {
		if index > 0 && turns[index-1].Topic == turn.Topic && turns[index-1].Reply.Question != "" {
			blocks = append(blocks, muted.Render("QUESTION YOU ANSWERED")+"\n"+model.markdown(turns[index-1].Reply.Question))
		}
		blocks = append(blocks, muted.Render("YOUR ANSWER")+"\n"+model.markdown(turn.Message))
		blocks = append(blocks, model.sectionDivider("FEEDBACK"))
	} else if turn.Action == "question" {
		blocks = append(blocks, muted.Render("YOUR QUESTION")+"\n"+model.markdown(turn.Message))
	}
	if turn.Reply.Assessment != "none" {
		title := map[string]string{"correct": "Correct", "partial": "Partly there", "incorrect": "Let's revisit this"}[turn.Reply.Assessment]
		blocks = append(blocks, heading.Render(title)+"\n"+model.markedMarkdown(turn.Reply.Feedback, len(turn.Reply.Citations)))
	} else if turn.Reply.Feedback != "" {
		blocks = append(blocks, model.markedMarkdown(turn.Reply.Feedback, len(turn.Reply.Citations)))
	}
	if turn.Action != "answer" || turn.Reply.Feedback == "" {
		blocks = append(blocks, model.markedMarkdown(turn.Reply.Explanation, len(turn.Reply.Citations)))
	}
	if turn.Action == "answer" && (turn.Reply.Assessment != "none" || turn.Reply.Question == "") {
		blocks = append(blocks, model.sectionDivider("END OF QUESTION"))
	}
	if turn.Reply.Question != "" && turn.Action != "answer" && index != len(turns)-1 {
		question := heading.Render("NEXT QUESTION") + "\n" + lipgloss.NewStyle().Width(max(1, model.bodyWidth()-6)).Render(tutor.Plain(turn.Reply.Question))
		blocks = append(blocks, lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(max(1, model.bodyWidth()-2)).Render(question))
	}
	blocks = append(blocks, model.evidenceBlock(turn.Reply.Citations))
	return strings.Join(blocks, "\n\n")
}

func (model interfaceModel) sectionDivider(label string) string {
	text := ""
	if label != "" {
		text = "── " + label + " "
	}
	text += strings.Repeat("─", max(0, model.bodyWidth()-ansi.StringWidth(text)))
	return lipgloss.NewStyle().Faint(true).Render(text)
}

func (model interfaceModel) inQuiz() bool {
	if model.page != "lesson" {
		return false
	}
	if model.busy {
		return strings.HasPrefix(model.intent.action, "quiz")
	}
	return model.application.state.Quiz != nil && model.history == len(model.application.state.Turns)-1
}

func (model interfaceModel) activeTab() string {
	if model.inQuiz() {
		return "quiz"
	}
	switch model.page {
	case "lesson", "topics", "sources", "source":
		return "course"
	}
	return model.page
}

func (model interfaceModel) topicTitle() string {
	identifier := model.application.state.Topic
	if model.page == "lesson" && model.history >= 0 && model.history < len(model.application.state.Turns) {
		identifier = model.application.state.Turns[model.history].Topic
	}
	if model.movingToAnotherTopic() {
		identifier = model.intent.topic
	}
	return model.titleOf(identifier)
}

func (model interfaceModel) titleOf(identifier string) string {
	for _, topic := range model.application.course {
		if topic.ID == identifier {
			return tutor.Plain(topic.Title)
		}
	}
	if path, isFile := tutor.FilePath(identifier); isFile {
		return tutor.FileTitlePrefix + tutor.Plain(path)
	}
	return identifier
}

func (model interfaceModel) composer() string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Faint(true)
	if model.page != "lesson" && !model.askingGoal() {
		content := muted.Render("Esc goes back")
		if model.input.Value() != "" {
			content += "\n" + muted.Render(ansi.Truncate("Draft kept: "+tutor.Plain(model.input.Value()), model.bodyWidth(), "…"))
		}
		return content
	}
	title, tint, above := "Ask about this topic", palette.accent, ""
	state := model.application.state
	if state.Question != "" {
		question := ansi.Wrap(tutor.Plain(state.Question), model.bodyWidth(), "")
		lines := strings.Split(question, "\n")
		if len(lines) > 2 {
			lines = lines[:2]
			lines[1] = ansi.Truncate(lines[1]+" …", model.bodyWidth(), "…")
		}
		above, title = "NEXT QUESTION\n"+strings.Join(lines, "\n"), "Your answer"
	}
	if model.history < len(state.Turns)-1 {
		above, title = "Earlier step · Esc returns to the current question", "Ask about this topic"
	} else if reviewing(state) {
		above, title = "", "Ask about this answer, or continue"
		if quiz := state.Quiz; quiz != nil && quiz.Completed() >= quiz.Goal {
			above, title = fmt.Sprintf("QUIZ COMPLETE · %d / %d answered\n%s", quiz.Completed(), quiz.Goal, model.quizCompleteLine()), "Ask about this answer"
			if quiz.Exam {
				above = model.examCompleteText()
			}
		}
	}
	if model.askingGoal() {
		above, title = "", goalQuestion
	}
	if model.busy {
		above, title, tint = "", waitingPhrase(model.intent.action), lipgloss.NoColor{}
	}
	actions := ""
	if line := model.actionLine(); line != "" {
		actions = line + "\n"
	}
	if above != "" {
		above = muted.Render(ansi.Wrap(above, model.bodyWidth(), "")) + "\n\n"
	}
	return actions + model.sectionDivider("") + "\n\n" + above + model.framedInput(title, model.typingHints(), tint)
}

func (model interfaceModel) sidebar(height int) string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Faint(true)
	title := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	score := tutor.ScoreCourse(model.application.course, model.application.state.Progress)
	visited, total := score.Read, score.Total
	filled := 0
	if total > 0 {
		filled = min(20, visited*20/total)
	}
	content := title.Render("YOUR COURSE") + "\n\n" + fmt.Sprintf("%d / %d topics read", visited, total) + "\n" + title.Render(strings.Repeat("━", filled)) + muted.Render(strings.Repeat("─", 20-filled)) + "\n\n" + fmt.Sprintf("Understanding %d%%", score.Percent) + "\n\n" + muted.Render("CURRENT TOPIC") + "\n" + lipgloss.NewStyle().Width(sidebarWidth-3).Render(model.topicTitle())
	if quiz := model.application.state.Quiz; quiz != nil && !model.movingToAnotherTopic() {
		content += "\n\n" + title.Render(model.quizBlockTitle()) + fmt.Sprintf("\n%d / %d answered", quiz.Completed(), quiz.Goal)
	}
	if model.bodyHeight() >= 19 {
		content += "\n\n" + muted.Render(model.sidebarShortcutsTitle()) + "\n" + model.sidebarShortcuts()
	}
	column := lipgloss.NewStyle().Width(sidebarWidth).Height(height).MaxHeight(height).Render(content)
	rail := muted.Render(strings.TrimSuffix(strings.Repeat("│\n", height), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top, column, rail)
}

func (model interfaceModel) View() tea.View {
	palette := model.palette()
	accent := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	muted := lipgloss.NewStyle().Faint(true)
	if model.width < 44 || model.height < 18 {
		view := tea.NewView(fmt.Sprintf("repo-tutor\n\nWiden to 44 columns and 18 rows.\nCurrent: %d × %d\n\nCtrl+Q quits; --plain works at any size.", model.width, model.height))
		view.AltScreen = true
		return view
	}
	width := model.width - 4
	brand := accent.Render("repo-tutor") + muted.Render("  /  ") + tutor.Plain(filepath.Base(model.application.snapshot.Root))
	status := model.application.state.Agent + " · saved"
	if model.busy {
		status = model.spinner.View() + " " + model.application.state.Agent + " is thinking"
	}
	brand = ansi.Truncate(brand, max(8, width-lipgloss.Width(status)-2), "…")
	space := max(1, width-lipgloss.Width(brand)-lipgloss.Width(status))
	header := ansi.Truncate(brand+strings.Repeat(" ", space)+muted.Render(status), width, "")
	navigationWidth := width
	if model.tabsBleed() {
		navigationWidth += 2
	}
	navigation := model.tabBar(navigationWidth)
	title := tutor.Plain(model.pageTitle)
	if model.page == "lesson" {
		title = model.topicTitle()
	}
	if model.page == "lesson" && len(model.application.state.Turns) > 0 {
		title = model.historyTitle()
	}
	if quiz := model.application.state.Quiz; model.inQuiz() && quiz != nil {
		title = fmt.Sprintf("Question %d / %d · %s", quiz.Active+1, quiz.Goal, model.topicTitle())
		if quiz.Exam {
			title = "Exam · " + title
		}
	}
	body := model.viewport.View()
	if model.busy && model.page == "lesson" {
		title = "Preparing " + turnLabel(model.intent.action)
		if named := model.titleOf(model.intent.topic); named != model.intent.topic {
			title += " · " + named
		}
		body = accent.Render(model.spinner.View()+" "+waitingText(model.intent.action)) + "\n\n" + muted.Render(groundingText(model.intent.action)+"\n\n"+model.browsingHint()+"\nEsc stops this request; your progress is saved.")
	}
	if model.isPicker() {
		body = model.picker.View()
	}
	main := accent.Render(ansi.Truncate(title, model.bodyWidth(), "…")) + "\n\n" + body
	main = lipgloss.NewStyle().Width(model.bodyWidth()).MaxWidth(model.bodyWidth()).Height(model.bodyHeight()).MaxHeight(model.bodyHeight()).Render(main)
	column := lipgloss.JoinVertical(lipgloss.Left, main, "", model.composer())
	workspace := column
	if model.width >= 108 {
		workspace = lipgloss.JoinHorizontal(lipgloss.Top, model.sidebar(lipgloss.Height(column)), "  ", column)
	}
	notice := model.notice
	if model.busy {
		notice = fmt.Sprintf("%s %s · %s elapsed · limit %s", model.spinner.View(), waitingPhrase(model.intent.action), time.Since(model.started).Round(time.Second), model.application.timeout)
	}
	if model.page == "lesson" && model.history < len(model.application.state.Turns)-1 {
		notice = "Earlier step · Esc returns to the current question · Alt+←/→ steps through this topic"
	}
	indicator := ""
	if model.prefetching && !model.busy {
		indicator = model.prefetchSpinner.View() + " " + model.prefetchLabel()
	}
	noticeStyle := muted
	if model.skipArmed || model.leaveArmed {
		noticeStyle = lipgloss.NewStyle().Foreground(palette.warning).Bold(true)
	}
	footer := model.footerLine(notice, indicator, width, noticeStyle) + "\n" + model.footerKeys(width)
	margin := lipgloss.NewStyle().Padding(0, 2)
	navigationRows := margin.Render(navigation + "\n" + model.tabRule(navigationWidth))
	if model.tabsBleed() {
		navigationRows = navigation + "\n" + model.tabRule(navigationWidth)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, margin.Render(lipgloss.JoinVertical(lipgloss.Left, header, "")), navigationRows, margin.Render(lipgloss.JoinVertical(lipgloss.Left, workspace, "", footer)))
	content = lipgloss.NewStyle().MaxWidth(model.width).MaxHeight(model.height).Render(content)
	if model.noColor {
		content = ansi.Strip(content)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

func waitingPhrase(action string) string {
	switch action {
	case "quiz", "quiz-next", "quiz-extend", "probe", "exam":
		return "Preparing your question"
	case "answer":
		return "Reviewing your answer"
	case "question":
		return "Thinking about your question"
	default:
		return "Preparing your lesson"
	}
}

func waitingText(action string) string {
	return waitingPhrase(action) + "…"
}

func groundingText(action string) string {
	switch action {
	case "quiz", "quiz-next", "quiz-extend", "probe", "exam":
		return "Writing a question from this repository's files."
	case "answer":
		return "Checking your answer against this repository's files."
	case "question":
		return "Looking for the answer in this repository's files."
	default:
		return "Writing your lesson from this repository's files."
	}
}

func turnLabel(action string) string {
	switch action {
	case "exam":
		return "exam question"
	case "quiz", "quiz-next", "quiz-extend":
		return "quiz question"
	case "probe":
		return "knowledge check"
	case "answer":
		return "answer review"
	case "question":
		return "reply"
	default:
		return "lesson"
	}
}

func (model interfaceModel) evidenceBlock(citations []repository.Location) string {
	count := sourceCount(len(citations))
	number := lipgloss.NewStyle().Foreground(model.palette().accent).Bold(true)
	faint := lipgloss.NewStyle().Faint(true)
	rows := []string{model.sectionDivider("EVIDENCE · " + count)}
	for index, citation := range citations {
		rows = append(rows, "  "+number.Render(fmt.Sprintf("[%d]", index+1))+"  "+tutor.Plain(citation.Path)+faint.Render(fmt.Sprintf(":%d", citation.Line)))
	}
	return strings.Join(rows, "\n")
}
