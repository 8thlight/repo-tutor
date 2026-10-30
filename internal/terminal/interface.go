package terminal

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

type turnRunner func(context.Context, session.State, string, string, string) (session.State, error)

type turnIntent struct {
	action  string
	message string
	topic   string
}

type turnFinished struct {
	state session.State
	err   error
}

type interfaceModel struct {
	application       application
	ctx               context.Context
	run               turnRunner
	input             textarea.Model
	viewport          viewport.Model
	picker            list.Model
	spinner           spinner.Model
	width             int
	height            int
	page              string
	homeSelected      int
	origin            string
	sourceList        string
	sourceCitation    int
	lessonReturn      *lessonPlace
	goalPending       bool
	noticeGeneration  int
	skipArmed         bool
	leaveArmed        bool
	armedPrompt       string
	skipGeneration    int
	settingsFocus     int
	prefetcher        lessonPrefetcher
	prefetching       bool
	prefetchScheduled bool
	prefetchID        int
	prefetchTopic     string
	edgeBump          string
	prefetchCancel    context.CancelFunc
	prefetchFailed    map[string]bool
	prefetchSpinner   spinner.Model
	overviewStale     bool
	sourcePath        string
	sourceLine        int
	sourceRendered    bool
	markdownViews     map[string]bool
	pageTitle         string
	pageContent       string
	history           int
	richColor         bool
	noColor           bool
	background        color.Color
	foreground        color.Color
	busy              bool
	cancel            context.CancelFunc
	intent            turnIntent
	errorText         string
	notice            string
	started           time.Time
}

func runInterface(ctx context.Context, application application, input io.Reader, output io.Writer) error {
	runningContext, cancel := context.WithCancel(ctx)
	var running sync.Mutex
	defer func() {
		cancel()
		running.Lock()
		running.Unlock()
	}()
	runner := func(turnContext context.Context, state session.State, action, message, topic string) (session.State, error) {
		running.Lock()
		defer running.Unlock()
		if err := runningContext.Err(); err != nil {
			return session.State{}, err
		}
		if err := turnContext.Err(); err != nil {
			return session.State{}, err
		}
		worker := application
		worker.state = state
		return worker.completeTurn(turnContext, action, message, topic)
	}
	_, noColor := os.LookupEnv("NO_COLOR")
	model := newInterface(runningContext, application, runner, noColor)
	if application.prefetch {
		model.prefetcher = func(prefetchContext context.Context, state session.State, topic string) error {
			running.Lock()
			defer running.Unlock()
			if err := runningContext.Err(); err != nil {
				return err
			}
			if err := prefetchContext.Err(); err != nil {
				return err
			}
			worker := application
			worker.state = state
			if isQuestionKey(topic) {
				return worker.prefetchQuestion(prefetchContext)
			}
			return worker.prefetchLesson(prefetchContext, topic)
		}
	}
	environment := os.Environ()
	options := []tea.ProgramOption{
		tea.WithContext(runningContext), tea.WithInput(input), tea.WithOutput(output),
		tea.WithEnvironment(plainCursorMovement(environment)), tea.WithColorProfile(colorprofile.Detect(output, environment)),
	}
	if noColor {
		options = append(options, tea.WithColorProfile(colorprofile.NoTTY))
	}
	_, err := tea.NewProgram(model, options...).Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("terminal interface: %w; use --plain for line input", err)
	}
	fmt.Fprintln(output, "Session saved. Run the same command to resume.")
	return nil
}

func plainCursorMovement(environment []string) []string {
	adjusted := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "TERM=") {
			adjusted = append(adjusted, entry)
		}
	}
	return append(adjusted, "TERM=vt100")
}

func newInterface(ctx context.Context, application application, runner turnRunner, noColor bool) interfaceModel {
	input := textarea.New()
	input.Placeholder = lessonPlaceholder
	input.Prompt = "  "
	input.ShowLineNumbers = false
	input.CharLimit = 16000
	input.DynamicHeight = true
	input.MinHeight = 1
	input.MaxHeight = inputMaxRows
	input.MaxWidth = 0
	input.KeyMap.InsertNewline = key.NewBinding(key.WithDisabled())
	input.Focus()
	activity := spinner.New()
	activity.Spinner = spinner.MiniDot
	background := spinner.New()
	background.Spinner = spinner.MiniDot
	model := interfaceModel{
		application: application, ctx: ctx, run: runner,
		input: input, viewport: viewport.New(), spinner: activity, prefetchSpinner: background,
		width: 100, height: 32, noColor: noColor,
		page: "lesson", history: len(application.state.Turns) - 1,
		notice: "",
	}
	model.viewport.SoftWrap = true
	if len(application.state.Turns) == 0 && !application.state.GoalChosen {
		model.goalPending = true
		model.input.Placeholder = application.state.Goal
	}
	switch {
	case application.settingsStatus.Unreadable:
		model.notice = settingsUnreadableNotice
	case application.settingsStatus.ReadOnly:
		model.notice = settingsReadOnlyNotice
	}
	if len(application.state.Turns) == 0 {
		model.page, model.pageTitle = "home", "Welcome"
		model.input.Blur()
	}
	model.resize()
	return model
}

func (model interfaceModel) Init() tea.Cmd {
	commands := []tea.Cmd{tea.RequestBackgroundColor, tea.RequestForegroundColor, textarea.Blink}
	if model.goalPending {
		return tea.Batch(append(commands, model.input.Focus())...)
	}
	if len(model.application.state.Turns) == 0 {
		return tea.Batch(append(commands, func() tea.Msg { return beginFirstLesson{} })...)
	}
	if _, wanted := model.prefetchTarget(); wanted {
		commands = append(commands, tea.Tick(prefetchDelay, func(time.Time) tea.Msg { return prefetchDue{} }))
	}
	return tea.Batch(commands...)
}

const (
	lessonPlaceholder = ""
	inputMaxRows      = 6
)

type beginFirstLesson struct{}

func (model interfaceModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	next, command := model.update(message)
	updated := next.(interfaceModel)
	scheduled, schedule := updated.schedulePrefetch(model)
	updated, command = scheduled, tea.Batch(command, schedule)
	if model.primaryPage() && !updated.primaryPage() {
		updated.origin = model.page
	}
	updated, expiry := updated.expireNotice(model)
	return updated, tea.Batch(command, expiry)
}

func (model interfaceModel) primaryPage() bool {
	return model.page == "lesson" || model.page == "home"
}

func (model interfaceModel) back() (tea.Model, tea.Cmd) {
	if model.page == "source" && model.sourceCitation > 0 && model.lessonReturn != nil {
		return model.backToTheLesson()
	}
	if model.page == "source" && model.sourceList != "" {
		return model.openPicker(model.sourceList, "")
	}
	if model.origin == "home" {
		model.page, model.pageTitle = "home", "Welcome"
		model.input.Blur()
		model.rebuildContent()
		model.viewport.GotoTop()
		return model, nil
	}
	return model.learn()
}

func (model interfaceModel) update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		model.width, model.height = message.Width, message.Height
		model.resize()
		return model, nil
	case tea.PasteMsg:
		if model.askingGoal() || (model.page == "lesson" && !model.busy) {
			flattened := tea.PasteMsg{Content: strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(message.Content, "\r", " "), "\n", " ")), " ")}
			var command tea.Cmd
			model.input, command = model.input.Update(flattened)
			return model, command
		}
		return model, nil
	case tea.ColorProfileMsg:
		model.richColor = message.Profile == colorprofile.TrueColor || message.Profile == colorprofile.ANSI256
		return model, nil
	case tea.ForegroundColorMsg:
		model.foreground = message.Color
		return model, nil
	case tea.BackgroundColorMsg:
		model.background = message.Color
		model.resize()
		return model, nil
	case beginFirstLesson:
		if overview, found := model.application.cachedLesson(tutor.Overview); found {
			return model.openCached(overview)
		}
		started, command := model.begin(turnIntent{"lesson", openingMessage, model.application.state.Topic})
		return started.(interfaceModel).showingHome(), command
	case turnFinished:
		model.busy = false
		if model.cancel != nil {
			model.cancel()
			model.cancel = nil
		}
		if message.err != nil {
			model.errorText = tutor.Plain(message.err.Error())
			model.notice = "That did not finish. Your progress and draft are safe. Ctrl+R tries again."
		} else {
			model.application.state = message.state
			model.history = len(message.state.Turns) - 1
			if model.intent.message == openingMessage {
				model.overviewStale = false
			}
			model.errorText = ""
			if model.intent.action == "answer" || model.intent.action == "question" {
				model.input.Reset()
			}
			model.notice = finishedNotice(model.intent.action)
		}
		if model.page == "progress" {
			model.pageContent = model.progressText()
		}
		if model.isPicker() {
			model.configurePicker(model.picker.FilterValue())
		}
		model.rebuildContent()
		model.viewport.GotoTop()
		if model.page == "lesson" {
			return model, model.input.Focus()
		}
		return model, nil
	case prefetchDue:
		model.prefetchScheduled = false
		return model.startPrefetch()
	case prefetchFinished:
		return model.finishPrefetch(message), nil
	case spinner.TickMsg:
		var commands []tea.Cmd
		if model.busy {
			var command tea.Cmd
			model.spinner, command = model.spinner.Update(message)
			commands = append(commands, command)
		}
		if model.prefetching {
			var command tea.Cmd
			model.prefetchSpinner, command = model.prefetchSpinner.Update(message)
			commands = append(commands, command)
		}
		return model, tea.Batch(commands...)
	case noticeExpired:
		if message.generation == model.noticeGeneration && !model.noticeSticky() {
			model.notice = ""
		}
		return model, nil
	case skipExpired:
		if message.generation == model.skipGeneration {
			model = model.disarmSkip()
		}
		return model, nil
	case tea.KeyPressMsg:
		if (model.skipArmed && message.String() != "ctrl+n") || (model.leaveArmed && message.String() != "ctrl+l") {
			model = model.disarmSkip()
		}
		if model.askingGoal() {
			if next, command, handled := model.goalKey(message); handled {
				return next, command
			}
		}
		if next, command, handled := model.citationKey(message.String()); handled {
			return next, command
		}
		switch message.String() {
		case "ctrl+q":
			if model.cancel != nil {
				model.cancel()
			}
			model = model.stopPrefetch()
			return model, tea.Quit
		case "ctrl+c":
			if model.busy {
				model.cancel()
				model = model.stopPrefetch()
				model.notice = "Stopping. Your progress is saved."
				return model, nil
			}
			model = model.stopPrefetch()
			return model, tea.Quit
		case "esc":
			if !model.primaryPage() {
				return model.back()
			}
			if model.page != "lesson" || model.history != len(model.application.state.Turns)-1 {
				return model.learn()
			}
			if model.busy {
				model.cancel()
				model = model.stopPrefetch()
				model.notice = "Stopping. Your progress is saved."
				return model, nil
			}
		case "ctrl+k":
			return model.toggled("topics", func() (tea.Model, tea.Cmd) { return model.openPicker("topics", "") })
		case "ctrl+s":
			if model.movingToAnotherTopic() && model.page != "sources" {
				model.notice = sourcesWaitNotice
				return model, nil
			}
			return model.toggled("sources", func() (tea.Model, tea.Cmd) { return model.openPicker("sources", "") })
		case "ctrl+p":
			return model.toggled("progress", func() (tea.Model, tea.Cmd) {
				return model.openDocument("progress", "Progress", model.progressText())
			})
		case "ctrl+b":
			return model.toggled("quiz", func() (tea.Model, tea.Cmd) { return model.openPicker("quiz", "") })
		case "ctrl+f":
			if inQuestionFlow(model.application.state) || questionOpen(model.application.state) {
				model.notice = "Ctrl+N asks the next question."
				return model, nil
			}
			return model.continueQuiz(false)
		case "f3":
			return model.toggled("about", model.openAbout)
		case "f2":
			return model.toggled("settings", model.openSettings)
		case "space":
			if model.page == "settings" {
				return model.chooseSetting()
			}
		case "f1":
			return model.toggled("help", func() (tea.Model, tea.Cmd) {
				return model.openDocument("help", "A few ways to keep learning", interfaceHelp)
			})
		case "ctrl+l":
			if model.leavable() {
				return model.leaveOrArm()
			}
			return model.learn()
		case "ctrl+n":
			if questionOpen(model.application.state) {
				return model.skipOrArm()
			}
			if inQuestionFlow(model.application.state) {
				return model.continueQuiz(false)
			}
			return model.next()
		case "ctrl+g":
			return model.begin(turnIntent{"quiz", "Ask one practical comprehension question about this topic.", model.application.state.Topic})
		case "ctrl+r":
			if model.errorText != "" && model.page == "home" {
				retried, command := model.begin(model.intent)
				return retried.(interfaceModel).showingHome(), command
			}
			if model.errorText != "" {
				return model.begin(model.intent)
			}
		case "alt+left", "alt+right":
			if model.page == "lesson" && len(model.application.state.Turns) > 0 {
				if message.String() == "alt+left" {
					return model.stepTurns(-1)
				}
				return model.stepTurns(1)
			}
		case "t":
			if model.page == "source" && model.sourcePath != "" {
				return model.openTopic(tutor.FileTopic(model.sourcePath))
			}
		case "a":
			if model.page == "quiz" && model.isPicker() && !model.picker.SettingFilter() {
				return model.revealSlot()
			}
		case "v":
			if model.page == "source" && isMarkdown(model.sourcePath) {
				return model.toggleMarkdownView()
			}
		case "up", "down", "j", "k", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if model.page == "home" {
				return model.homeKey(message.String())
			}
			if model.page == "settings" {
				switch message.String() {
				case "up", "k":
					return model.moveSetting(-1)
				case "down", "j":
					return model.moveSetting(1)
				}
			}
		case "tab", "shift+tab":
			return model.cyclePage(message.String() == "shift+tab")
		case "enter":
			if model.page == "home" {
				return model.openSelectedHomeEntry()
			}
			if model.isPicker() {
				if model.picker.SettingFilter() {
					var command tea.Cmd
					model.picker, command = model.picker.Update(message)
					return model, command
				}
				return model.selectItem()
			}
			if model.page == "source" {
				return model.back()
			}
			if model.page == "lesson" && !model.busy {
				return model.submit()
			}
		}
	}
	if model.isPicker() {
		var command tea.Cmd
		model.picker, command = model.picker.Update(message)
		return model, command
	}
	var commands []tea.Cmd
	if pressed, ok := message.(tea.KeyPressMsg); ok {
		if model.page != "lesson" || pressed.String() == "pgup" || pressed.String() == "pgdown" {
			var command tea.Cmd
			model.viewport, command = model.viewport.Update(message)
			return model, command
		}
	}
	if _, wheel := message.(tea.MouseWheelMsg); wheel {
		var command tea.Cmd
		model.viewport, command = model.viewport.Update(message)
		commands = append(commands, command)
	} else if model.page == "lesson" && !model.busy {
		var command tea.Cmd
		model.input, command = model.input.Update(message)
		commands = append(commands, command)
	}
	return model, tea.Batch(commands...)
}

func (model interfaceModel) begin(intent turnIntent) (tea.Model, tea.Cmd) {
	if model.busy {
		model.notice = model.runningNotice()
		return model, nil
	}
	if len(intent.message) > 16000 {
		model.notice = "Keep each message under 16000 bytes. Your draft is still here."
		return model, nil
	}
	if !model.prefetchingTopic(intent) {
		model = model.stopPrefetch()
	}
	model.busy = true
	model.started = time.Now()
	model.intent = intent
	model.errorText = ""
	model.page = "lesson"
	model.input.Blur()
	model.rebuildContent()
	model.viewport.GotoTop()
	turnContext, cancel := context.WithCancel(model.ctx)
	model.cancel = cancel
	state, runner := model.application.state, model.run
	return model, tea.Batch(model.spinner.Tick, func() tea.Msg {
		state, err := runner(turnContext, state, intent.action, intent.message, intent.topic)
		return turnFinished{state, err}
	})
}

func (model interfaceModel) submit() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(model.input.Value())
	if text == "" {
		if reviewing(model.application.state) && model.history == len(model.application.state.Turns)-1 {
			return model.continueQuiz(false)
		}
		return model, nil
	}
	if model.history != len(model.application.state.Turns)-1 {
		model.notice = "You are viewing an earlier step. Esc returns to the current question; your draft stays."
		return model, nil
	}
	action := "question"
	if model.application.state.Question != "" && !reviewing(model.application.state) {
		action = "answer"
	}
	return model.begin(turnIntent{action, text, model.application.state.Topic})
}

func (model interfaceModel) learn() (tea.Model, tea.Cmd) {
	model.page = "lesson"
	model.history = len(model.application.state.Turns) - 1
	model.rebuildContent()
	model.viewport.GotoTop()
	if !model.busy {
		return model, model.input.Focus()
	}
	return model, nil
}

func (model interfaceModel) openCached(overview tutor.SavedLesson) (tea.Model, tea.Cmd) {
	opened, err := model.application.openWithLesson(overview, openingMessage)
	if err != nil {
		model.errorText = tutor.Plain(err.Error())
		return model, nil
	}
	model.application.state = opened
	model.history = len(opened.Turns) - 1
	model.notice = "Loaded the saved overview"
	model.overviewStale = !overview.Fresh(model.application.snapshot.Fingerprint)
	if model.overviewStale {
		model.notice = staleOverviewNotice
	}
	model.rebuildContent()
	model.viewport.GotoTop()
	if model.page == "lesson" {
		return model, model.input.Focus()
	}
	return model, nil
}

func (model interfaceModel) showingHome() interfaceModel {
	model.page, model.pageTitle = "home", "Welcome"
	model.input.Blur()
	model.rebuildContent()
	return model
}

func (model interfaceModel) toggled(page string, open func() (tea.Model, tea.Cmd)) (tea.Model, tea.Cmd) {
	if model.page == page {
		return model.back()
	}
	return open()
}
