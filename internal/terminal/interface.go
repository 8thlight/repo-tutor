package terminal

import (
	"context"
	"fmt"
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
	application    application
	ctx            context.Context
	run            turnRunner
	input          textarea.Model
	viewport       viewport.Model
	picker         list.Model
	spinner        spinner.Model
	width          int
	height         int
	page           string
	homeSelected   int
	origin         string
	overviewStale  bool
	sourcePath     string
	sourceLine     int
	sourceRendered bool
	markdownViews  map[string]bool
	pageTitle      string
	pageContent    string
	history        int
	dark           bool
	richColor      bool
	noColor        bool
	busy           bool
	cancel         context.CancelFunc
	intent         turnIntent
	errorText      string
	notice         string
	started        time.Time
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
		worker := application
		worker.state = state
		return worker.completeTurn(turnContext, action, message, topic)
	}
	_, noColor := os.LookupEnv("NO_COLOR")
	model := newInterface(runningContext, application, runner, noColor)
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
	input.Placeholder = "Answer the question, ask a follow-up, or type /help"
	input.Prompt = "  "
	input.ShowLineNumbers = false
	input.CharLimit = 16000
	input.MaxHeight = 2
	input.MaxWidth = 0
	input.SetHeight(2)
	input.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("alt+enter", "shift+enter"))
	input.Focus()
	activity := spinner.New()
	activity.Spinner = spinner.MiniDot
	model := interfaceModel{
		application: application, ctx: ctx, run: runner,
		input: input, viewport: viewport.New(), spinner: activity,
		width: 100, height: 32, dark: true, noColor: noColor,
		page: "lesson", history: len(application.state.Turns) - 1,
		notice: "Completed turns save automatically",
	}
	model.viewport.SoftWrap = true
	if len(application.state.Turns) == 0 {
		model.page, model.pageTitle = "home", "Welcome"
		model.input.Blur()
	}
	model.resize()
	return model
}

func (model interfaceModel) Init() tea.Cmd {
	commands := []tea.Cmd{tea.RequestBackgroundColor, textarea.Blink}
	if len(model.application.state.Turns) == 0 {
		return tea.Batch(append(commands, func() tea.Msg { return beginFirstLesson{} })...)
	}
	return tea.Batch(commands...)
}

type beginFirstLesson struct{}

func (model interfaceModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	next, command := model.update(message)
	updated := next.(interfaceModel)
	if model.primaryPage() && !updated.primaryPage() {
		updated.origin = model.page
	}
	return updated, command
}

func (model interfaceModel) primaryPage() bool {
	return model.page == "lesson" || model.page == "home"
}

func (model interfaceModel) back() (tea.Model, tea.Cmd) {
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
	case tea.ColorProfileMsg:
		model.richColor = message.Profile == colorprofile.TrueColor || message.Profile == colorprofile.ANSI256
		return model, nil
	case tea.BackgroundColorMsg:
		model.dark = message.IsDark()
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
			model.notice = "Turn not recorded. Your completed progress and draft are safe. Ctrl+R retries."
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
			model.notice = "Saved automatically"
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
	case spinner.TickMsg:
		if model.busy {
			var command tea.Cmd
			model.spinner, command = model.spinner.Update(message)
			return model, command
		}
		return model, nil
	case tea.KeyPressMsg:
		switch message.String() {
		case "ctrl+q":
			if model.cancel != nil {
				model.cancel()
			}
			return model, tea.Quit
		case "ctrl+c":
			if model.busy {
				model.cancel()
				model.notice = "Stopping this turn. Completed progress is saved."
				return model, nil
			}
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
				model.notice = "Stopping this turn. Completed progress is saved."
				return model, nil
			}
		case "ctrl+k":
			return model.toggled("topics", func() (tea.Model, tea.Cmd) { return model.openPicker("topics", "") })
		case "ctrl+s":
			return model.toggled("sources", func() (tea.Model, tea.Cmd) { return model.openPicker("sources", "") })
		case "ctrl+o":
			return model.toggled("hotspots", func() (tea.Model, tea.Cmd) { return model.openPicker("hotspots", "") })
		case "ctrl+p":
			return model.toggled("progress", func() (tea.Model, tea.Cmd) {
				return model.openDocument("progress", "Progress", model.progressText())
			})
		case "ctrl+b":
			return model.toggled("quiz", func() (tea.Model, tea.Cmd) { return model.openPicker("quiz", "") })
		case "ctrl+f":
			return model.continueQuiz(false)
		case "f2":
			return model.toggled("about", model.openAbout)
		case "f1":
			return model.toggled("help", func() (tea.Model, tea.Cmd) {
				return model.openDocument("help", "A few ways to keep learning", interfaceHelp)
			})
		case "ctrl+l":
			return model.learn()
		case "ctrl+n":
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
				step := 1
				if message.String() == "alt+left" {
					step = -1
				}
				model.history = max(0, min(model.history+step, len(model.application.state.Turns)-1))
				model.rebuildContent()
				model.viewport.GotoTop()
				return model, nil
			}
		case "v":
			if model.page == "source" && isMarkdown(model.sourcePath) {
				return model.toggleMarkdownView()
			}
		case "up", "down", "j", "k", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if model.page == "home" {
				return model.homeKey(message.String())
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
		model.notice = "A turn is already running. Esc stops it; you can still browse topics and sources."
		return model, nil
	}
	if len(intent.message) > 16000 {
		model.notice = "Keep each message under 16000 bytes. Your draft is still here."
		return model, nil
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
	if strings.HasPrefix(text, "/") {
		return model.slash(text)
	}
	if model.history != len(model.application.state.Turns)-1 {
		model.notice = "You are viewing an earlier turn. Esc returns to the current question; your draft stays."
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
