package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muandane/goji/pkg/config"
)

func AskQuestions(config *config.Config, presetType, presetMessage string) ([]string, error) {
	var commitType, commitScope, commitSubject, commitDescription string

	// Validate that Types is not empty
	if len(config.Types) == 0 {
		return nil, fmt.Errorf(
			"no commit types configured. Please run 'goji init --global' or 'goji init --repo' to initialize your configuration",
		)
	}

	nameStyle := lipgloss.NewStyle().Width(15).Align(lipgloss.Left)
	emojiStyle := lipgloss.NewStyle().Width(5).PaddingRight(5).Align(lipgloss.Left)
	descStyle := lipgloss.NewStyle().Width(45).Align(lipgloss.Left)

	commitTypeOptions := make([]huh.Option[string], len(config.Types))
	for i, ct := range config.Types {
		row := lipgloss.JoinHorizontal(lipgloss.Center,
			nameStyle.Render(ct.Name),
			emojiStyle.Render(ct.Emoji),
			descStyle.Render(ct.Description))

		// Build value: "type emoji" or just "type" if NoEmoji
		value := ct.Name
		if !config.NoEmoji && ct.Emoji != "" {
			value = fmt.Sprintf("%s %s", ct.Name, ct.Emoji)
		}
		commitTypeOptions[i] = huh.NewOption[string](row, value)
	}

	if presetType != "" {
		for _, option := range commitTypeOptions {
			if strings.HasPrefix(option.Value, presetType) {
				commitType = option.Value
				break
			}
		}
	}
	if presetMessage != "" {
		commitSubject = presetMessage
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select the type of change:").
				Options(commitTypeOptions...).
				Height(8).
				Value(&commitType),
			huh.NewInput().
				Title("Scope of this change (optional):").
				Placeholder("e.g., ci, api, parser").
				CharLimit(50).
				Suggestions(config.Scopes).
				Value(&commitScope),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Short summary of changes:").
				Placeholder("Short Commit description").
				CharLimit(70).
				Value(&commitSubject).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("subject cannot be empty")
					}
					return nil
				}),
			huh.NewText().
				Title("Long description (optional):").
				CharLimit(config.SubjectMaxLength).
				Placeholder("Longer Commit description").
				Value(&commitDescription).
				WithHeight(4),
			huh.NewConfirm().
				Title("Commit changes?").
				Affirmative("Yes").
				Negative("No").
				Validate(func(v bool) error {
					if !v {
						return errors.New("changes not committed")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	commitMessage := commitType
	if commitScope != "" {
		// Add space before scope if emoji is present (commitType would be "type emoji")
		if !config.NoEmoji && strings.Contains(commitType, " ") {
			commitMessage += fmt.Sprintf(" (%s)", strings.TrimSpace(commitScope))
		} else {
			commitMessage += fmt.Sprintf("(%s)", strings.TrimSpace(commitScope))
		}
	}
	commitMessage += fmt.Sprintf(": %s", strings.TrimSpace(commitSubject))

	commitBody := strings.TrimSpace(commitDescription)

	return []string{commitMessage, commitBody}, nil
}
