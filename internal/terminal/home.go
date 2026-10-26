package terminal

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"repo-tutor/internal/tutor"
)

const (
	placeBrowse   = "browse"
	placeProgress = "progress"
	placeAbout    = "about"
	placeProbe    = "probe"
	placeRefresh  = "refresh"
)

type homeEntry struct {
	place       string
	label       string
	description string
	reason      string
}

func (model interfaceModel) homeEntries() []homeEntry {
	state := model.application.state
	recommendation := tutor.Recommend(model.application.course, state.Progress, state.Topic, state.Quiz, reviewing(state))
	var entries []homeEntry
	if recommendation.Place == tutor.PlaceContinue {
		entries = append(entries, homeEntry{tutor.PlaceContinue, "Continue", "Move on from the answer you just reviewed.", ""})
	}
	entries = append(entries, homeEntry{tutor.PlaceLearn, "Learn", "Read the repository overview, then one idea at a time with a question to answer.", ""})
	if _, err := probeIntent(state, model.application.course); err == nil && state.Probe == nil && len(state.Turns) > 0 {
		entries = append(entries, homeEntry{placeProbe, "Check what you know", "Optional. Already know this repository's foundations? Answer a few short questions to skip topics you know. New here? Skip this.", ""})
	}
	if state.Quiz != nil {
		entries = append(entries, homeEntry{tutor.PlaceQuiz, "Quiz", "Answer five questions on this topic and revisit any you skipped.", ""})
	}
	entries = append(entries, homeEntry{placeBrowse, "Browse", "Topics, cited sources, and reading hotspots.", ""})
	if model.overviewStale {
		entries = append(entries, homeEntry{placeRefresh, "Refresh the overview", "The repository changed since the overview was written. Generate a new one.", ""})
	}
	if tutor.Started(state.Progress) {
		entries = append(entries, homeEntry{placeProgress, "Progress", "Lessons read and answers assessed so far.", ""})
	}
	entries = append(entries, homeEntry{placeAbout, "About", "Version, copyright, and credits. F2 opens it from anywhere.", ""})
	for index, entry := range entries {
		if entry.place == recommendation.Place {
			entry.reason = recommendation.Reason
			entries = append([]homeEntry{entry}, append(entries[:index:index], entries[index+1:]...)...)
			break
		}
	}
	return entries
}

func (model interfaceModel) homeContent() string {
	palette := model.palette()
	muted := lipgloss.NewStyle().Foreground(palette.muted)
	accent := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	width := model.bodyWidth()
	var content strings.Builder
	content.WriteString(accent.Render("A codebase, one idea at a time.") + "\n\n")
	if tutor.Started(model.application.state.Progress) {
		content.WriteString(muted.Render(ansi.Wrap(model.homeStatus(), width, "")) + "\n\n")
	} else {
		content.WriteString(ansi.Wrap("repo-tutor teaches you this repository from its own files. It explains one small idea at a time, asks a question, and checks your answer against the code it cites.", width, "") + "\n\n")
	}
	if model.busy {
		content.WriteString(muted.Render(ansi.Wrap("Preparing the overview in the background. Browse while you wait.", width, "")) + "\n\n")
	}
	if model.errorText != "" {
		content.WriteString(ansi.Wrap("The overview could not be prepared. Ctrl+R tries again.\n"+model.errorText, width, "") + "\n\n")
	}
	for index, entry := range model.homeEntries() {
		marker, label := "  ", entry.label
		if entry.reason != "" {
			label += "  Recommended: " + entry.reason
		}
		line := fmt.Sprintf("%s%d  %s", marker, index+1, label)
		if index == model.homeSelected {
			line = accent.Render(fmt.Sprintf("› %d  %s", index+1, label))
		}
		content.WriteString(ansi.Truncate(line, width, "…") + "\n")
		content.WriteString(muted.Render(ansi.Wrap("     "+entry.description, width, "")) + "\n\n")
	}
	content.WriteString(muted.Render("Enter opens the selected item · 1-9 jump · ↑↓ move"))
	return content.String()
}

func (model interfaceModel) homeStatus() string {
	read, attempts := 0, 0
	for _, progress := range model.application.state.Progress {
		if progress.Lessons > 0 {
			read++
		}
		attempts += progress.Correct + progress.Partial + progress.Missed
	}
	status := fmt.Sprintf("%d of %d topics read · %d answers assessed", read, len(model.application.course), attempts)
	if quiz := model.application.state.Quiz; quiz != nil {
		status += fmt.Sprintf(" · quiz %d of %d", quiz.Completed(), quiz.Goal)
	}
	return status
}

func (model interfaceModel) openHome() (tea.Model, tea.Cmd) {
	model.page, model.pageTitle, model.homeSelected = "home", "Welcome", 0
	model.input.Blur()
	model.rebuildContent()
	model.viewport.GotoTop()
	return model, nil
}

func (model interfaceModel) homeKey(pressed string) (tea.Model, tea.Cmd) {
	entries := model.homeEntries()
	switch pressed {
	case "up", "k":
		model.homeSelected = (model.homeSelected + len(entries) - 1) % len(entries)
	case "down", "j":
		model.homeSelected = (model.homeSelected + 1) % len(entries)
	default:
		number := int(pressed[0] - '1')
		if number < 0 || number >= len(entries) {
			return model, nil
		}
		return model.openHomeEntry(entries[number])
	}
	model.rebuildContent()
	return model, nil
}

func (model interfaceModel) openSelectedHomeEntry() (tea.Model, tea.Cmd) {
	entries := model.homeEntries()
	return model.openHomeEntry(entries[min(model.homeSelected, len(entries)-1)])
}

func (model interfaceModel) openHomeEntry(entry homeEntry) (tea.Model, tea.Cmd) {
	switch entry.place {
	case tutor.PlaceContinue:
		return model.continueQuiz(false)
	case tutor.PlaceLearn:
		if len(model.application.state.Turns) == 0 && !model.busy {
			return model.begin(turnIntent{"lesson", openingMessage, model.application.state.Topic})
		}
		return model.learn()
	case tutor.PlaceQuiz:
		return model.learn()
	case placeBrowse:
		return model.openPicker("topics", "")
	case placeAbout:
		return model.openAbout()
	case placeProbe:
		intent, err := probeIntent(model.application.state, model.application.course)
		if err != nil {
			model.notice = err.Error()
			return model, nil
		}
		return model.begin(intent)
	case placeRefresh:
		return model.begin(turnIntent{"lesson", openingMessage, tutor.Overview})
	default:
		return model.openDocument("progress", "Progress", model.progressText())
	}
}

func (model interfaceModel) openAbout() (tea.Model, tea.Cmd) {
	return model.openDocument("about", "About repo-tutor", model.aboutText())
}
