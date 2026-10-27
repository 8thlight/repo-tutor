package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"

	"repo-tutor/internal/about"
	"repo-tutor/internal/agent"
	"repo-tutor/internal/repository"
	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

const usage = `repo-tutor: learn a code repository with a terminal tutor

Usage:
  repo-tutor [tutor] [options] [repository]
  repo-tutor scan [--json | --map] [repository]
  repo-tutor doctor

The repository defaults to the current directory. Tutor sessions resume automatically.

Tutor options:
  --agent auto|claude|codex  CLI to use (auto prefers Codex, then Claude)
  --model NAME              Model supported by your chosen CLI
  --goal TEXT               What you want to learn (skips the first-run question)
  --new                     Archive the previous session and start again
  --ask TEXT                Ask one question, save the reply, and exit
  --state-dir DIRECTORY     Store sessions in this directory
  --timeout DURATION        Limit each agent turn (default 3m)
  --plain                   Use the line-based interface instead of the full-screen UI
  --no-prefetch             Do not prepare the next lesson in the background

Examples:
  repo-tutor ../onboard-kit --agent codex
  repo-tutor scan ../onboard-kit
  repo-tutor --goal "Trace how a plan becomes a document" ../onboard-kit

Inside a session, type /help for topics, quizzes, source browsing, and progress.
`

func Run(ctx context.Context, arguments []string, input io.Reader, output io.Writer, diagnostics io.Writer) error {
	command := "tutor"
	if len(arguments) > 0 {
		switch arguments[0] {
		case "-h", "--help", "help":
			fmt.Fprint(output, usage)
			return nil
		case "--version", "version":
			if len(arguments) == 2 && arguments[1] == "--notices" {
				fmt.Fprint(output, about.Notices())
				return nil
			}
			fmt.Fprint(output, about.Read().Text())
			return nil
		case "scan", "tutor", "doctor":
			command, arguments = arguments[0], arguments[1:]
		}
	}
	if command == "doctor" {
		return doctor(ctx, arguments, output)
	}
	if command == "scan" {
		return scan(ctx, arguments, output, diagnostics)
	}
	return study(ctx, arguments, input, output, diagnostics)
}

func parse(flags *flag.FlagSet, arguments []string) (string, error) {
	var options, positional []string
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--" {
			positional = append(positional, arguments[index+1:]...)
			break
		}
		if !strings.HasPrefix(argument, "-") || argument == "-" {
			positional = append(positional, argument)
			continue
		}
		options = append(options, argument)
		name := strings.TrimLeft(strings.SplitN(argument, "=", 2)[0], "-")
		definition := flags.Lookup(name)
		if definition == nil || strings.Contains(argument, "=") {
			continue
		}
		if boolean, ok := definition.Value.(interface{ IsBoolFlag() bool }); ok && boolean.IsBoolFlag() {
			continue
		}
		if index+1 < len(arguments) {
			index++
			options = append(options, arguments[index])
		}
	}
	if err := flags.Parse(options); err != nil {
		return "", err
	}
	if len(positional) > 1 {
		return "", fmt.Errorf("expected one repository path, got %d; see repo-tutor --help", len(positional))
	}
	if len(positional) == 0 {
		return ".", nil
	}
	return positional[0], nil
}

func scan(ctx context.Context, arguments []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	asJSON := flags.Bool("json", false, "print the source inventory as JSON")
	asMap := flags.Bool("map", false, "print a Mermaid package dependency graph")
	directory, err := parse(flags, arguments)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if *asJSON && *asMap {
		return fmt.Errorf("choose one of --json and --map")
	}
	snapshot, err := repository.Read(ctx, directory)
	if err != nil {
		return err
	}
	if *asJSON {
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(snapshot)
	}
	if *asMap {
		_, err := fmt.Fprint(output, tutor.Diagram(snapshot))
		return err
	}
	summary(output, snapshot)
	fmt.Fprintln(output, "\nTopics, foundations first:")
	for index, topic := range tutor.Course(snapshot) {
		fmt.Fprintf(output, "  %d. %s\n", index+1, tutor.Plain(topic.Title))
	}
	fmt.Fprintln(output, "\nEntry points:")
	for _, entry := range snapshot.EntryPoints {
		fmt.Fprintf(output, "  %s:%d\n", tutor.Plain(entry.Path), entry.Line)
	}
	showHotspots(output, snapshot, 5)
	fmt.Fprintf(output, "\nStart a session: repo-tutor --agent auto %q\n", snapshot.Root)
	return nil
}

func doctor(ctx context.Context, arguments []string, output io.Writer) error {
	if len(arguments) > 0 {
		return fmt.Errorf("usage: repo-tutor doctor")
	}
	available := false
	for _, name := range []string{"claude", "codex"} {
		path, err := exec.LookPath(name)
		if err != nil {
			fmt.Fprintf(output, "%s: not found on PATH\n", name)
			continue
		}
		versionContext, cancel := context.WithTimeout(ctx, 10*time.Second)
		data, err := exec.CommandContext(versionContext, path, "--version").CombinedOutput()
		cancel()
		if err != nil {
			fmt.Fprintf(output, "%s: cannot run --version (%v)\n", name, err)
			continue
		}
		available = true
		fmt.Fprintf(output, "%s: %s (%s)\n", name, strings.TrimSpace(tutor.Plain(string(data))), tutor.Plain(path))
	}
	fmt.Fprintln(output, "\nLogin and model access are checked on your first tutor turn. Offline scan needs no agent.")
	if !available {
		return fmt.Errorf("install and sign in to Claude Code or Codex to start a tutor session")
	}
	return nil
}

func summary(output io.Writer, snapshot repository.Snapshot) {
	fmt.Fprintf(output, "Repository: %s\n", tutor.Plain(snapshot.Root))
	fmt.Fprintf(output, "%d module(s), %d packages, %d documents, %d entry point(s)\n", len(snapshot.Modules), len(snapshot.Packages), len(snapshot.Documents), len(snapshot.EntryPoints))
	for _, module := range snapshot.Modules {
		fmt.Fprintf(output, "  %s (%s)\n", tutor.Plain(module.Path), tutor.Plain(module.Directory))
	}
	for _, warning := range snapshot.Warnings {
		fmt.Fprintf(output, "Warning: %s\n", tutor.Plain(warning))
	}
}

func study(ctx context.Context, arguments []string, input io.Reader, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("tutor", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	preference := flags.String("agent", "auto", "agent CLI")
	model := flags.String("model", "", "agent model")
	goal := flags.String("goal", "", "learning goal")
	fresh := flags.Bool("new", false, "start a fresh session")
	ask := flags.String("ask", "", "ask once and exit")
	stateDirectory := flags.String("state-dir", "", "session storage directory")
	timeout := flags.Duration("timeout", 3*time.Minute, "turn timeout")
	plain := flags.Bool("plain", false, "use the line-based interface")
	noPrefetch := flags.Bool("no-prefetch", false, "do not prepare the next lesson in the background")
	directory, err := parse(flags, arguments)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(output, usage)
		return nil
	}
	if err != nil {
		return err
	}
	if *timeout <= 0 {
		return fmt.Errorf("--timeout must be positive")
	}
	snapshot, err := repository.Read(ctx, directory)
	if err != nil {
		return err
	}
	if *stateDirectory != "" {
		*stateDirectory, err = filepath.Abs(*stateDirectory)
		if err != nil {
			return err
		}
	}
	store, err := session.Open(*stateDirectory, snapshot.Root)
	if err != nil {
		return err
	}
	defer store.Close()
	var state session.State
	found := false
	if !*fresh {
		state, found, err = store.Load()
		if err != nil {
			return err
		}
	}
	if found && state.Repository != snapshot.Root {
		return fmt.Errorf("saved session belongs to a different repository; use --new")
	}
	var carriedProgress map[string]tutor.Progress
	if found && state.Fingerprint != snapshot.Fingerprint {
		fmt.Fprintln(output, "Repository sources changed. Archiving the old session and starting a fresh course. Your progress carries over.")
		carriedProgress = state.Progress
		if *preference == "auto" {
			*preference = state.Agent
		}
		if *model == "" {
			*model = state.Model
		}
		if *goal == "" {
			*goal = state.Goal
		}
		*fresh, found = true, false
	}
	if found {
		if (*preference != "auto" && *preference != state.Agent) || (*model != "" && *model != state.Model) || (*goal != "" && *goal != state.Goal) {
			return fmt.Errorf("agent, model, or goal differs from the saved session; use --new to start with these settings")
		}
		*preference = state.Agent
	}
	chosen, err := agent.Choose(*preference)
	if err != nil {
		return err
	}
	if !found {
		if *fresh {
			if err := store.Archive(); err != nil {
				return err
			}
		}
		learningGoal := *goal
		if learningGoal == "" {
			learningGoal = tutor.DefaultGoal
		}
		state = session.State{Version: session.Version, Repository: snapshot.Root, Fingerprint: snapshot.Fingerprint, Agent: chosen, Model: *model, Goal: learningGoal, GoalChosen: *goal != "", Topic: tutor.Overview, Progress: map[string]tutor.Progress{}}
		for topic, progress := range carriedProgress {
			state.Progress[topic] = progress
		}
		if err := store.Save(state); err != nil {
			return err
		}
	}
	loaded, loadedStatus, err := session.LoadSettings(*stateDirectory)
	if err != nil {
		return err
	}
	application := application{snapshot: snapshot, course: tutor.Course(snapshot), state: state, store: store, output: output, diagnostics: diagnostics, timeout: *timeout, prefetch: !*noPrefetch, settings: loaded, settingsStatus: loadedStatus, settingsBase: *stateDirectory}
	if !*plain && *ask == "" && interactiveStreams(input, output) {
		return runInterface(ctx, application, input, output)
	}
	summary(output, snapshot)
	fmt.Fprintf(output, "\nAgent: %s. Goal: %s\n", chosen, tutor.Plain(state.Goal))
	fmt.Fprintf(output, "Session: %s\n", tutor.Plain(filepath.Join(store.Directory, "session.json")))
	if *ask != "" {
		return application.turn(ctx, "question", *ask, state.Topic)
	}
	fmt.Fprintln(output, "Type an answer or a question. /topics lists the course; /help lists commands; /quit saves and exits.")
	if len(state.Turns) == 0 {
		if err := application.start(ctx); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(output, "\nResuming your last turn:")
		application.show(state.Turns[len(state.Turns)-1].Reply)
	}
	return application.loop(ctx, input)
}

func interactiveStreams(input io.Reader, output io.Writer) bool {
	reader, inputIsFile := input.(*os.File)
	writer, outputIsFile := output.(*os.File)
	return inputIsFile && outputIsFile && os.Getenv("TERM") != "dumb" && term.IsTerminal(reader.Fd()) && term.IsTerminal(writer.Fd())
}
