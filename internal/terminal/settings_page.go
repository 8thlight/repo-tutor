package terminal

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"repo-tutor/internal/session"
	"repo-tutor/internal/settings"
)

const (
	settingsNarrowWidth   = 80
	settingsSavedNotice   = "Saved"
	settingsUnsavedNotice = "Could not save. This change lasts until you quit."
	powerlineNote         = "Needs a Powerline or Nerd Font in your terminal."
	settingsKeysNote      = "↑↓ move · Space chooses · F2 or Esc closes"
	noColorNote           = "Color is off, so tab style does not apply."
	noColorHighlightNote  = "Color is off, so highlight color does not apply."
	highlightTerminalNote = "Drawn in your terminal's own colors."
	settingsLabelWidth    = 14
	examLengthNote        = "An exam has one question for each topic you have read, up to this many."
)

type settingChoice struct {
	group, label, note string
	selected           func(settings.Settings) bool
	apply              func(settings.Settings) settings.Settings
	style, highlight   string
}

func tabStyleChoice(style, label, note string) settingChoice {
	return settingChoice{
		group: "TAB STYLE", label: label, note: note, style: style,
		selected: func(current settings.Settings) bool { return current.TabStyle == style },
		apply: func(current settings.Settings) settings.Settings {
			current.TabStyle = style
			return current
		},
	}
}

func markdownChoice(rendered bool, label string) settingChoice {
	return settingChoice{
		group: "MARKDOWN FILES OPEN", label: label,
		selected: func(current settings.Settings) bool { return current.MarkdownOpensRendered == rendered },
		apply: func(current settings.Settings) settings.Settings {
			current.MarkdownOpensRendered = rendered
			return current
		},
	}
}

func quizLengthChoice(length int) settingChoice {
	return settingChoice{
		group: "QUIZ LENGTH", label: fmt.Sprintf("%d questions", length),
		selected: func(current settings.Settings) bool { return current.QuizLength == length },
		apply: func(current settings.Settings) settings.Settings {
			current.QuizLength = length
			return current
		},
	}
}

func examLengthLabel(length int) string {
	if length == 0 {
		return "Every topic read"
	}
	return fmt.Sprintf("%d questions", length)
}

func examLengthChoice(length int) settingChoice {
	return settingChoice{
		group: "EXAM LENGTH", label: examLengthLabel(length),
		selected: func(current settings.Settings) bool { return current.ExamLength == length },
		apply: func(current settings.Settings) settings.Settings {
			current.ExamLength = length
			return current
		},
	}
}

func highlightChoice(brand highlightColor) settingChoice {
	return settingChoice{
		group: "HIGHLIGHT COLOR", label: brand.label, highlight: brand.id,
		selected: func(current settings.Settings) bool { return current.Highlight == brand.id },
		apply: func(current settings.Settings) settings.Settings {
			current.Highlight = brand.id
			return current
		},
	}
}

func settingChoices() []settingChoice {
	choices := []settingChoice{
		tabStyleChoice(settings.TabStyleNone, "None", ""),
		tabStyleChoice(settings.TabStyleBlock, "Block", ""),
		tabStyleChoice(settings.TabStyleSlanted, "Slanted", powerlineNote),
		tabStyleChoice(settings.TabStyleRounded, "Rounded", powerlineNote),
		tabStyleChoice(settings.TabStyleJoinedSlanted, "Joined slants", powerlineNote),
		tabStyleChoice(settings.TabStyleJoinedRounded, "Joined rounds", powerlineNote),
		tabStyleChoice(settings.TabStyleJoinedArrows, "Joined arrows", powerlineNote),
		markdownChoice(true, "Rendered"),
		markdownChoice(false, "Source"),
	}
	for _, length := range settings.QuizLengths {
		choices = append(choices, quizLengthChoice(length))
	}
	for _, length := range settings.ExamLengths {
		choices = append(choices, examLengthChoice(length))
	}
	for _, brand := range highlightColors {
		choices = append(choices, highlightChoice(brand))
	}
	return choices
}

func (model interfaceModel) openSettings() (tea.Model, tea.Cmd) {
	opened, command := model.document("settings", "Settings · Saved automatically", "")
	for index, choice := range settingChoices() {
		if choice.style != "" && choice.selected(opened.application.settings) {
			opened.settingsFocus = index
		}
	}
	opened.rebuildContent()
	return opened, command
}

func (model interfaceModel) moveSetting(step int) (tea.Model, tea.Cmd) {
	count := len(settingChoices())
	model.settingsFocus = (model.settingsFocus + step + count) % count
	model.rebuildContent()
	return model, nil
}

func (model interfaceModel) chooseSetting() (tea.Model, tea.Cmd) {
	choice := settingChoices()[model.settingsFocus]
	if model.application.settingsStatus.ReadOnly {
		model.notice = settingsReadOnlyNotice
		return model, nil
	}
	changed, err := session.ChangeSettings(model.application.settingsBase, choice.apply)
	model.notice = settingsSavedNotice
	if err != nil {
		if errors.Is(err, session.ErrSettingsReadOnly) {
			model.application.settingsStatus.ReadOnly = true
			model.notice = settingsReadOnlyNotice
			return model, nil
		}
		changed = choice.apply(model.application.settings)
		model.notice = settingsUnsavedNotice
	}
	model.application.settings = changed
	model.application.settingsStatus.Unreadable = false
	model.rebuildContent()
	return model, nil
}

func (model interfaceModel) settingsContent() string {
	content, _ := model.settingsLayout()
	return content
}

func (model interfaceModel) settingsLayout() (string, int) {
	palette := model.palette()
	muted := lipgloss.NewStyle().Faint(true)
	heading := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	narrow := model.bodyWidth() < settingsNarrowWidth
	lines := []string{muted.Render(settingsKeysNote), ""}
	switch {
	case model.application.settingsStatus.ReadOnly:
		lines = append(lines, lipgloss.NewStyle().Foreground(palette.warning).Render(settingsReadOnlyNotice), "")
	case model.application.settingsStatus.Unreadable:
		lines = append(lines, lipgloss.NewStyle().Foreground(palette.warning).Render(settingsUnreadableNotice), "")
	}
	choices := settingChoices()
	focusLine := 0
	group := ""
	for index, choice := range choices {
		if choice.group != group {
			if group != "" {
				lines = append(lines, "")
			}
			group = choice.group
			lines = append(lines, heading.Render(group))
			switch {
			case model.noColor && group == "TAB STYLE":
				lines = append(lines, muted.Render(noColorNote))
			case model.noColor && group == "HIGHLIGHT COLOR":
				lines = append(lines, muted.Render(noColorHighlightNote))
			case group == "HIGHLIGHT COLOR":
				lines = append(lines, muted.Render(highlightTerminalNote))
			case group == "EXAM LENGTH":
				lines = append(lines, muted.Render(examLengthNote))
			}
		}
		marker, radio := "  ", "( )"
		if index == model.settingsFocus {
			marker = lipgloss.NewStyle().Foreground(palette.accent).Bold(true).Render("› ")
			focusLine = len(lines)
		}
		if choice.selected(model.application.settings) {
			radio = "(•)"
		}
		row := marker + radio + " "
		if choice.highlight != "" {
			row += lipgloss.NewStyle().Foreground(model.highlight(choice.highlight)).Render("██") + " "
		}
		row += choice.label
		switch {
		case choice.style != "" && narrow:
			lines = append(lines, row, "      "+model.tabs(choice.style, "course"))
		case choice.style != "":
			lines = append(lines, row+strings.Repeat(" ", max(1, settingsLabelWidth-lipgloss.Width(choice.label)))+model.tabs(choice.style, "course"))
		default:
			lines = append(lines, row)
		}
		if choice.note != "" {
			lines = append(lines, "      "+muted.Render(choice.note))
		}
	}
	return strings.Join(lines, "\n"), focusLine
}

func (model *interfaceModel) followSettingsFocus() {
	_, focusLine := model.settingsLayout()
	offset, height := model.viewport.YOffset(), model.viewport.Height()
	switch {
	case focusLine <= offset:
		model.viewport.SetYOffset(max(0, focusLine-1))
	case focusLine >= offset+height:
		model.viewport.SetYOffset(focusLine - height + 1)
	}
}
