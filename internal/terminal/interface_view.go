package terminal

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/tutor"
)

type interfacePalette struct{ accent, foreground, muted, border, warning, cutout, rule color.Color }

func (model interfaceModel) palette() interfacePalette {
	if model.noColor {
		return interfacePalette{}
	}
	if model.dark {
		return interfacePalette{model.highlight(model.application.settings.Highlight), lipgloss.Color("#E4E4E7"), lipgloss.Color("#A1A1AA"), lipgloss.Color("#3F3F46"), lipgloss.Color("#FBBF24"), model.cutoutColor("#09090B"), lipgloss.Color("#2E2E33")}
	}
	return interfacePalette{model.highlight(model.application.settings.Highlight), lipgloss.Color("#18181B"), lipgloss.Color("#52525B"), lipgloss.Color("#D4D4D8"), lipgloss.Color("#92400E"), model.cutoutColor("#FAFAFA"), lipgloss.Color("#D4D4D8")}
}

func (model interfaceModel) bodyWidth() int {
	width := max(1, model.width-4)
	if model.width >= 108 {
		width -= 29
	}
	return max(1, width)
}
func (model interfaceModel) bodyHeight() int {
	return max(1, model.height-8-lipgloss.Height(model.composer()))
}

func (model *interfaceModel) resize() {
	palette := model.palette()
	model.input.SetWidth(max(1, model.bodyWidth()-4))
	model.input.SetHeight(2)
	inputStyles := textarea.DefaultStyles(model.dark)
	inputStyles.Focused.Prompt = lipgloss.NewStyle().Foreground(palette.accent)
	inputStyles.Focused.Text = lipgloss.NewStyle().Foreground(palette.foreground)
	inputStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(palette.muted)
	inputStyles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(palette.muted)
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

func (model interfaceModel) markdown(text string) string {
	configuration := styles.DarkStyleConfig
	if !model.dark {
		configuration = styles.LightStyleConfig
	}
	if model.noColor {
		configuration = styles.ASCIIStyleConfig
	}
	zero := uint(0)
	configuration.Document.Margin = &zero
	configuration.Document.BlockPrefix = ""
	configuration.Document.BlockSuffix = ""
	if !model.noColor {
		codeColor := "#5EEAD4"
		if !model.dark {
			codeColor = "#0F766E"
		}
		configuration.Code.Color = &codeColor
		configuration.Code.BackgroundColor = nil
	}
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

func (model interfaceModel) lessonContent() string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Foreground(palette.muted)
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
		blocks = append(blocks, lipgloss.NewStyle().Foreground(palette.warning).Bold(true).Render("Turn not recorded · Ctrl+R to retry")+"\n"+ansi.Wrap(model.errorText, model.bodyWidth(), ""))
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
		blocks = append(blocks, heading.Render(title)+"\n"+model.markdown(turn.Reply.Feedback))
	} else if turn.Reply.Feedback != "" {
		blocks = append(blocks, model.markdown(turn.Reply.Feedback))
	}
	if turn.Action != "answer" || turn.Reply.Feedback == "" {
		blocks = append(blocks, model.markdown(turn.Reply.Explanation))
	}
	if turn.Action == "answer" && (turn.Reply.Assessment != "none" || turn.Reply.Question == "") {
		blocks = append(blocks, model.sectionDivider("END OF QUESTION"))
	}
	if turn.Reply.Question != "" && turn.Action != "answer" && index != len(turns)-1 {
		question := heading.Render("NEXT QUESTION") + "\n" + lipgloss.NewStyle().Width(max(1, model.bodyWidth()-6)).Render(tutor.Plain(turn.Reply.Question))
		blocks = append(blocks, lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(palette.border).Padding(1, 2).Width(max(1, model.bodyWidth()-2)).Render(question))
	}
	var sources []string
	seen := map[string]bool{}
	for _, citation := range turn.Reply.Citations {
		name := fmt.Sprintf("%s:%d", citation.Path, citation.Line)
		if !seen[name] {
			sources = append(sources, tutor.Plain(name))
			seen[name] = true
		}
	}
	blocks = append(blocks, muted.Render("EVIDENCE · Ctrl+S to open source")+"\n"+muted.Render(strings.Join(sources, "\n")))
	return strings.Join(blocks, "\n\n")
}

func (model interfaceModel) sectionDivider(label string) string {
	text := ""
	if label != "" {
		text = "── " + label + " "
	}
	text += strings.Repeat("─", max(0, model.bodyWidth()-ansi.StringWidth(text)))
	return lipgloss.NewStyle().Foreground(model.palette().muted).Render(text)
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
	for _, topic := range model.application.course {
		if topic.ID == identifier {
			return tutor.Plain(topic.Title)
		}
	}
	return identifier
}

func (model interfaceModel) composer() string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Foreground(palette.muted)
	if model.page != "lesson" && !model.askingGoal() {
		content := muted.Render("Esc goes back")
		if model.input.Value() != "" {
			content += "\n" + muted.Render(ansi.Truncate("Draft kept: "+tutor.Plain(model.input.Value()), model.bodyWidth(), "…"))
		}
		return content
	}
	label := "Ask about this topic"
	border := palette.accent
	if model.application.state.Question != "" {
		question := ansi.Wrap(tutor.Plain(model.application.state.Question), model.bodyWidth(), "")
		lines := strings.Split(question, "\n")
		if len(lines) > 2 {
			lines = lines[:2]
			lines[1] = ansi.Truncate(lines[1]+" …", model.bodyWidth(), "…")
		}
		label = "NEXT QUESTION\n" + strings.Join(lines, "\n")
	}
	if model.history < len(model.application.state.Turns)-1 {
		label = "Earlier step · Esc returns to the current question"
	} else if reviewing(model.application.state) {
		label = "[ Continue → ] Enter empty / Ctrl+F\nAsk about this answer, or continue."
		if quiz := model.application.state.Quiz; quiz != nil && quiz.Completed() >= quiz.Goal {
			label = fmt.Sprintf("QUIZ COMPLETE · %d / %d answered\n%s", quiz.Completed(), quiz.Goal, model.quizCompleteLine())
		}
	}
	if model.askingGoal() {
		label = goalComposerLabel
	}
	if model.busy {
		label = "Preparing your next turn · Esc stops it"
		border = palette.border
	}
	editor := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Width(max(1, model.bodyWidth()-2)).Render(model.input.View())
	actions := ""
	if line := model.actionLine(); line != "" {
		actions = line + "\n"
	}
	return actions + model.sectionDivider("") + "\n" + muted.Render(ansi.Wrap(label, model.bodyWidth(), "")) + "\n" + editor
}

func (model interfaceModel) sidebar() string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Foreground(palette.muted)
	title := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	visited, correct, attempts := 0, 0, 0
	for _, progress := range model.application.state.Progress {
		if progress.Lessons > 0 {
			visited++
		}
		correct += progress.Correct
		attempts += progress.Correct + progress.Partial + progress.Missed
	}
	total := len(model.application.course)
	filled := 0
	if total > 0 {
		filled = min(20, visited*20/total)
	}
	content := title.Render("YOUR COURSE") + "\n\n" + fmt.Sprintf("%d / %d topics read", visited, total) + "\n" + title.Render(strings.Repeat("━", filled)) + muted.Render(strings.Repeat("─", 20-filled)) + "\n\n" + fmt.Sprintf("%d assessed answers\n%d correct", attempts, correct) + "\n\n" + muted.Render("CURRENT TOPIC") + "\n" + lipgloss.NewStyle().Width(23).Render(model.topicTitle())
	if quiz := model.application.state.Quiz; quiz != nil {
		content += "\n\n" + title.Render("QUIZ GOAL") + fmt.Sprintf("\n%d / %d answered\nCtrl+B  Question index", quiz.Completed(), quiz.Goal)
	}
	if model.bodyHeight() >= 19 {
		content += "\n\n" + muted.Render("KEEP GOING") + "\n" + "Ctrl+N  Next topic\nCtrl+G  Quiz me\nCtrl+K  Choose a topic\nCtrl+S  Follow a citation"
	}
	return lipgloss.NewStyle().Width(26).Height(model.bodyHeight()).MaxHeight(model.bodyHeight()).BorderRight(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(palette.border).Render(content)
}

func (model interfaceModel) View() tea.View {
	palette := model.palette()
	accent := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(palette.muted)
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
		title = fmt.Sprintf("Question %d / %d · %s · Ctrl+B index", quiz.Active+1, quiz.Goal, model.topicTitle())
	}
	body := model.viewport.View()
	if model.busy && model.page == "lesson" {
		title = "Preparing " + turnLabel(model.intent.action)
		for _, topic := range model.application.course {
			if topic.ID == model.intent.topic {
				title += " · " + tutor.Plain(topic.Title)
			}
		}
		body = accent.Render(model.spinner.View()+" "+waitingText(model.intent.action)) + "\n\n" + muted.Render("Grounding the reply in repository sources.\n\nYou can browse topics and sources while you wait.\nEsc stops this request; your progress is saved.")
	}
	if model.isPicker() {
		body = model.picker.View()
	}
	main := accent.Render(ansi.Truncate(title, model.bodyWidth(), "…")) + "\n\n" + body
	main = lipgloss.NewStyle().Width(model.bodyWidth()).MaxWidth(model.bodyWidth()).Height(model.bodyHeight()).MaxHeight(model.bodyHeight()).Render(main)
	if model.width >= 108 {
		main = lipgloss.JoinHorizontal(lipgloss.Top, model.sidebar(), "  ", main)
	}
	composer := model.composer()
	if model.width >= 108 {
		composer = lipgloss.NewStyle().PaddingLeft(29).Render(composer)
	}
	notice := model.notice
	if model.busy {
		notice = fmt.Sprintf("%s Preparing a grounded %s · %s elapsed · limit %s", model.spinner.View(), turnLabel(model.intent.action), time.Since(model.started).Round(time.Second), model.application.timeout)
	}
	if model.isPicker() {
		notice += " · / search · ↑↓ move · Enter open"
	}
	if model.page == "lesson" && model.history < len(model.application.state.Turns)-1 {
		notice = "Earlier step · Esc returns to the current question · Alt+←/→ steps through this topic"
	}
	keys := "Enter send   Ctrl+K topics   Ctrl+S sources   Ctrl+Q quit"
	if reviewing(model.application.state) {
		keys = "Ctrl+F continue   Ctrl+B quiz index   Ctrl+S sources   Ctrl+Q quit"
	}
	if model.askingGoal() {
		keys = "Enter keeps the default   F1 help   Ctrl+Q quit"
	}
	if width < 80 {
		keys = "Enter send · F1 help · Ctrl+Q quit"
	}
	indicator := ""
	if model.prefetching && !model.busy {
		indicator = model.prefetchSpinner.View() + " Preparing next lesson"
	}
	footer := model.footerLine(notice, indicator, width, muted) + "\n" + muted.Render(ansi.Truncate(keys, width, "…"))
	margin := lipgloss.NewStyle().Padding(0, 2)
	navigationRows := margin.Render(navigation + "\n" + model.tabRule(navigationWidth))
	if model.tabsBleed() {
		navigationRows = navigation + "\n" + model.tabRule(navigationWidth)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, margin.Render(lipgloss.JoinVertical(lipgloss.Left, header, "")), navigationRows, margin.Render(lipgloss.JoinVertical(lipgloss.Left, main, "", composer, "", footer)))
	content = lipgloss.NewStyle().MaxWidth(model.width).MaxHeight(model.height).Render(content)
	if model.noColor {
		content = ansi.Strip(content)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

func waitingText(action string) string {
	switch action {
	case "quiz", "quiz-next", "quiz-extend", "probe":
		return "Preparing your question…"
	case "answer":
		return "Reviewing your answer…"
	case "question":
		return "Thinking about your question…"
	default:
		return "Preparing your lesson…"
	}
}

func turnLabel(action string) string {
	switch action {
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
