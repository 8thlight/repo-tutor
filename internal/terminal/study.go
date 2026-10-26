package terminal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"repo-tutor/internal/agent"
	"repo-tutor/internal/repository"
	"repo-tutor/internal/session"
	"repo-tutor/internal/tutor"
)

type application struct {
	snapshot    repository.Snapshot
	course      []tutor.Topic
	state       session.State
	store       *session.Store
	output      io.Writer
	diagnostics io.Writer
	timeout     time.Duration
}

func (application *application) turn(ctx context.Context, action, message, topic string) error {
	fmt.Fprintf(application.diagnostics, "\nWaiting for %s (turn limit %s; Ctrl-C to exit)...\n", application.state.Agent, application.timeout)
	updated, err := application.completeTurn(ctx, action, message, topic)
	if err != nil {
		return err
	}
	application.state = updated
	application.show(updated.Turns[len(updated.Turns)-1].Reply)
	return nil
}

func (application application) completeTurn(ctx context.Context, action, message, topic string) (session.State, error) {
	if len(message) > 16000 {
		return session.State{}, fmt.Errorf("message is too long; keep each turn under 16000 bytes")
	}
	request := tutor.Request{Topic: topic, Action: action, Message: message, Goal: application.state.Goal}
	if action == "quiz-next" || action == "quiz-extend" {
		request.Action = "quiz"
	}
	request.Lesson = lastLesson(application.state.Turns, topic)
	if topic == application.state.Topic {
		if !reviewing(application.state) {
			request.Question = application.state.Question
		}
		if len(application.state.Turns) > 0 {
			previous := application.state.Turns[len(application.state.Turns)-1].Reply
			request.Previous = &previous
		}
	}
	prepared := tutor.Prepare(application.snapshot, request)
	turnContext, cancel := context.WithTimeout(ctx, application.timeout)
	defer cancel()
	result, err := agent.Run(turnContext, agent.Request{
		Agent: application.state.Agent, Model: application.state.Model,
		Workspace: filepath.Join(application.store.Directory, "agent"),
		SessionID: application.state.SessionID, Prompt: prepared.Prompt,
	})
	if err != nil {
		return session.State{}, fmt.Errorf("%w\nProgress is saved. Retry the same command; use --new if the agent session cannot resume", err)
	}
	reply, err := tutor.DecodeReply(result.Content, prepared, request)
	if err != nil {
		return session.State{}, fmt.Errorf("%w\nThis turn was not recorded. Retry your message", err)
	}
	updated := application.recordReply(action, message, topic, request.Question, reply)
	updated.SessionID = result.SessionID
	if err := application.store.Save(updated); err != nil {
		return session.State{}, fmt.Errorf("save tutor turn: %w", err)
	}
	return updated, nil
}

func (application application) recordReply(action, message, topic, pending string, reply tutor.Reply) session.State {
	if action == "answer" && reply.Assessment != "none" {
		reply.Question = ""
	}
	if action == "question" && reviewing(application.state) {
		reply.Question = ""
	}
	updated := application.state
	updated.Topic = topic
	updated.Question = reply.Question
	updated.Review = action == "answer" && (reply.Assessment != "none" || reply.Question == "")
	if action == "question" && reviewing(application.state) && reply.Question == "" {
		updated.Review = true
	}
	updated.Quiz = tutor.RecordQuiz(updated.Quiz, topic, action, pending, reply)
	updated.Progress = tutor.Record(updated.Progress, topic, action, reply)
	updated.Probe, updated.Progress = tutor.RecordProbe(updated.Probe, application.course, updated.Progress, action, topic, reply)
	updated.Turns = append(append([]session.Turn(nil), updated.Turns...), session.Turn{Topic: topic, Action: action, Message: message, Reply: reply})
	updated.Quiz = tutor.RecordAnswerTurn(application.state.Quiz, updated.Quiz, len(updated.Turns)-1)
	if trimmed := len(updated.Turns) - 200; trimmed > 0 {
		updated.Turns = updated.Turns[trimmed:]
		updated.Quiz = tutor.ShiftQuizTurns(updated.Quiz, trimmed)
	}
	return updated
}

func lastLesson(turns []session.Turn, topic string) *tutor.Reply {
	for index := len(turns) - 1; index >= 0; index-- {
		if turns[index].Topic == topic && turns[index].Action == "lesson" {
			reply := turns[index].Reply
			return &reply
		}
	}
	return nil
}

func (application *application) show(reply tutor.Reply) {
	if reply.Assessment != "none" {
		fmt.Fprintf(application.output, "\nAssessment: %s\n", reply.Assessment)
	}
	if reply.Feedback != "" {
		fmt.Fprintf(application.output, "%s\n", tutor.Plain(reply.Feedback))
	}
	if reply.Assessment == "none" || reply.Feedback == "" {
		fmt.Fprintf(application.output, "\n%s\n", tutor.Plain(reply.Explanation))
	}
	fmt.Fprintln(application.output, "\nSources:")
	for _, citation := range reply.Citations {
		fmt.Fprintf(application.output, "  %s:%d\n", tutor.Plain(citation.Path), citation.Line)
	}
	if reviewing(application.state) {
		fmt.Fprintln(application.output, "\n--- End of question ---\nUse /continue when ready, or ask about this answer. /questions shows the quiz goal.")
	} else if reply.Question != "" {
		fmt.Fprintf(application.output, "\n---\nQuestion: %s\n", tutor.Plain(reply.Question))
	}
}

func (application *application) loop(ctx context.Context, input io.Reader) error {
	readingContext, stopReading := context.WithCancel(ctx)
	defer stopReading()
	type inputLine struct {
		text string
		err  error
	}
	lines := make(chan inputLine)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			select {
			case lines <- inputLine{text: scanner.Text()}:
			case <-readingContext.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			select {
			case lines <- inputLine{err: err}:
			case <-readingContext.Done():
			}
		}
	}()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		fmt.Fprint(application.output, "\nyou> ")
		var next inputLine
		select {
		case <-ctx.Done():
			return ctx.Err()
		case line, open := <-lines:
			if !open {
				fmt.Fprintln(application.output, "\nSession saved. Run the same command to resume.")
				return nil
			}
			next = line
		}
		if next.err != nil {
			return next.err
		}
		message := strings.TrimSpace(next.text)
		if message == "" {
			continue
		}
		if strings.HasPrefix(message, "/") {
			quit, err := application.command(ctx, message)
			if err != nil {
				fmt.Fprintln(application.diagnostics, tutor.Plain(err.Error()))
			}
			if quit {
				return nil
			}
			continue
		}
		action := "question"
		if application.state.Question != "" && !reviewing(application.state) {
			action = "answer"
		}
		if err := application.turn(ctx, action, message, application.state.Topic); err != nil {
			if ctx.Err() != nil {
				return err
			}
			fmt.Fprintln(application.diagnostics, tutor.Plain(err.Error()))
		}
	}
}

func (application *application) command(ctx context.Context, text string) (bool, error) {
	parts := strings.SplitN(text, " ", 2)
	argument := ""
	if len(parts) > 1 {
		argument = strings.TrimSpace(parts[1])
	}
	switch parts[0] {
	case "/continue", "/extend":
		intent, err := continueIntent(application.state, application.course, parts[0] == "/extend")
		if err != nil {
			return false, err
		}
		return false, application.turn(ctx, intent.action, intent.message, intent.topic)
	case "/probe":
		intent, err := probeIntent(application.state, application.course)
		if err != nil {
			return false, err
		}
		return false, application.turn(ctx, intent.action, intent.message, intent.topic)
	case "/questions":
		fmt.Fprintln(application.output, quizText(application.state.Quiz))
	case "/quit", "/exit":
		fmt.Fprintln(application.output, "Session saved. Run the same command to resume.")
		return true, nil
	case "/help":
		fmt.Fprint(application.output, `
/topics [FILTER]     List topics in dependency order, optionally matching a path
/topic NUMBER|PATH   Learn a chosen topic (number, directory, or import path)
/next                Learn the next topic
/probe               Ask three short questions and skip topics you already know
/quiz                Start a five-question quiz on this topic
/continue            Continue after reviewing an answer
/questions           Show the quiz question index and goal
/extend              Add five questions after finishing a quiz
/review              Revisit a topic with missed or partial answers
/progress            Show lessons and quiz outcomes
/sources             List the excerpts available for this topic
/read PATH[:LINE]    Show 30 source lines from the scanned snapshot
/map                 Print the package dependency graph as Mermaid
/hotspots            Show files with recent churn and branching logic
/glossary [FILTER]   List type names and their declaration locations
/help                Show these commands
/quit                Save and exit (Ctrl-D also exits)

Ask follow-up questions or answer the current question in plain text.
Each lesson is saved before it is displayed. Ctrl-C exits; completed turns remain saved.
`)
	case "/topics":
		application.topics(false, argument)
	case "/progress":
		application.topics(true, "")
	case "/sources":
		prepared := tutor.Prepare(application.snapshot, tutor.Request{Topic: application.state.Topic})
		for _, excerpt := range prepared.Excerpts {
			fmt.Fprintf(application.output, "  %s:%d-%d\n", tutor.Plain(excerpt.Path), excerpt.StartLine, excerpt.EndLine)
		}
	case "/map":
		fmt.Fprint(application.output, tutor.Diagram(application.snapshot))
	case "/hotspots":
		showHotspots(application.output, application.snapshot, 15)
	case "/glossary":
		shown, total := 0, 0
		for _, current := range repository.Foundations(application.snapshot) {
			for _, symbol := range current.Symbols {
				if symbol.Kind == "type" && strings.Contains(strings.ToLower(symbol.Name+" "+symbol.Path), strings.ToLower(argument)) {
					total++
					if shown < 40 {
						fmt.Fprintf(application.output, "  %s: %s:%d\n", tutor.Plain(symbol.Name), tutor.Plain(symbol.Path), symbol.Line)
						shown++
					}
				}
			}
		}
		fmt.Fprintf(application.output, "Showing %d of %d matching types. Use /glossary FILTER to narrow the list, or ask about a term.\n", shown, total)
	case "/read":
		return false, application.read(argument)
	case "/quiz":
		return false, application.turn(ctx, "quiz", "Ask one practical comprehension question about this topic.", application.state.Topic)
	case "/topic":
		topic, err := application.findTopic(argument)
		if err != nil {
			return false, err
		}
		return false, application.turn(ctx, "lesson", "Teach the next small idea in this topic.", topic.ID)
	case "/next":
		topic, found := tutor.NextUnheld(application.course, application.state.Progress, application.state.Topic)
		if !found {
			fmt.Fprintln(application.output, "You reached the last topic. Use /review, /quiz, or /topic to keep practicing.")
			return false, nil
		}
		return false, application.turn(ctx, "lesson", "Teach this topic and connect it to its foundations.", topic.ID)
	case "/review":
		var selected tutor.Topic
		highest := 0
		for _, topic := range application.course {
			progress := application.state.Progress[topic.ID]
			score := progress.Missed + progress.Partial - progress.Correct
			if score > highest {
				selected, highest = topic, score
			}
		}
		if highest == 0 {
			fmt.Fprintln(application.output, "No topic currently needs review from recorded answers. Try /quiz or /next.")
			return false, nil
		}
		return false, application.turn(ctx, "lesson", "Review this topic with a different example, focusing on earlier confusion.", selected.ID)
	default:
		return false, fmt.Errorf("unknown command %q; type /help", parts[0])
	}
	return false, nil
}

func showHotspots(output io.Writer, snapshot repository.Snapshot, limit int) {
	fmt.Fprintln(output, "\nReading hints: recent changes and branching logic")
	if snapshot.History {
		fmt.Fprintln(output, "Git window: last 180 days, at most 2000 non-merge commits; renames are not followed.")
	} else {
		fmt.Fprintln(output, "Git history unavailable. Ranking uses source branching only.")
	}
	for _, hotspot := range snapshot.Hotspots[:min(limit, len(snapshot.Hotspots))] {
		fmt.Fprintf(output, "  %s:%d: %d commits, %d branch points, %d lines\n", tutor.Plain(hotspot.Path), hotspot.Line, hotspot.Commits, hotspot.Branches, hotspot.Lines)
	}
	fmt.Fprintln(output, "Reading score = (commits + 1) * (branch points + 1). These are reading hints, not defect predictions.")
	fmt.Fprintln(output, "Go counts syntax nodes; Python and JavaScript/TypeScript use text estimates. Shallow history can undercount churn.")
	fmt.Fprintln(output, "Ask about a listed file, or use /read PATH:LINE in a tutor session.")
}

func (application *application) topics(progressOnly bool, filter string) {
	shown, total := 0, 0
	for index, topic := range application.course {
		progress := application.state.Progress[topic.ID]
		if progressOnly && progress == (tutor.Progress{}) {
			continue
		}
		if !strings.Contains(strings.ToLower(topic.Title+" "+topic.ID), strings.ToLower(filter)) {
			continue
		}
		total++
		if !progressOnly && shown == 50 {
			continue
		}
		shown++
		marker := " "
		if topic.ID == application.state.Topic {
			marker = ">"
		}
		known := ""
		if progress.Held {
			known = " (known)"
		}
		fmt.Fprintf(application.output, "%s %d. %s%s: %d lessons; answers: %d correct, %d partial, %d incorrect\n", marker, index+1, tutor.Plain(topic.Title), known, progress.Lessons, progress.Correct, progress.Partial, progress.Missed)
	}
	if progressOnly {
		fmt.Fprintln(application.output, "Quiz outcomes are the agent's assessments, not a certification of mastery.")
	} else {
		fmt.Fprintf(application.output, "Showing %d of %d matching topics. Use /topics FILTER to narrow the list.\n", shown, total)
	}
}

func (application *application) findTopic(value string) (tutor.Topic, error) {
	if number, err := strconv.Atoi(value); err == nil && number > 0 && number <= len(application.course) {
		return application.course[number-1], nil
	}
	for _, topic := range application.course {
		if topic.ID == value {
			return topic, nil
		}
		if current, exists := application.snapshot.Package(topic.ID); exists && current.Directory == value {
			return topic, nil
		}
	}
	return tutor.Topic{}, fmt.Errorf("unknown topic %q; use a number or path from /topics", value)
}

func (application *application) read(argument string) error {
	name := argument
	line := 1
	if colon := strings.LastIndex(argument, ":"); colon >= 0 {
		name = argument[:colon]
		var err error
		line, err = strconv.Atoi(argument[colon+1:])
		if err != nil || line < 1 {
			return fmt.Errorf("usage: /read relative/path.go[:positive-line-number]")
		}
	}
	file, exists := application.snapshot.File(name)
	if !exists {
		return fmt.Errorf("%q is not in the source snapshot; use /sources for available paths", name)
	}
	lines := strings.Split(strings.TrimSuffix(file.Content, "\n"), "\n")
	if line > len(lines) {
		return fmt.Errorf("%s has %d lines", name, len(lines))
	}
	for index := line - 1; index < min(line+29, len(lines)); index++ {
		fmt.Fprintf(application.output, "%s:%d  %s\n", tutor.Plain(name), index+1, tutor.Plain(lines[index]))
	}
	return nil
}
