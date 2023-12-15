package utils

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muandane/goji/pkg/config"
)

func AskQuestions(config *config.Config) (string, error) {
	var commitType string
	var commitScope string
	var commitSubject string
	commitTypeOptions := make([]huh.Option[string], len(config.Types))

	nameStyle := lipgloss.NewStyle().
		Width(15).
		Align(lipgloss.Left)

	emojiStyle := lipgloss.NewStyle().
		Width(5).
		PaddingRight(5).
		Align(lipgloss.Left)

	descStyle := lipgloss.NewStyle().
		Width(45).
		Align(lipgloss.Left)

	for i, ct := range config.Types {
		name := nameStyle.Render(ct.Name)
		emoji := emojiStyle.Render(ct.Emoji)
		desc := descStyle.Render(ct.Description)

		row := lipgloss.JoinHorizontal(lipgloss.Center, name, emoji, desc)
		commitTypeOptions[i] = huh.NewOption[string](row, fmt.Sprintf("%s %s", ct.Name, ct.Emoji))
	}

	promptType := huh.NewSelect[string]().
		Title("Select the type of change you are committing:").
		Options(commitTypeOptions...).
		Value(&commitType)

	err := promptType.Run()
	if err != nil {
		return "", err
	}
	// Only ask for commitScope if not in SkipQuestions
	if !isInSkipQuestions("Scopes", config.SkipQuestions) {
		promptScope := huh.NewInput().
			Title("What is the scope of this change? (class or file name): (press [enter] to skip)").
			CharLimit(50).
			Placeholder("Example: ci, api, parser").
			Value(&commitScope)

		err = promptScope.Run()
		if err != nil {
			return "", err
		}
	}

	promptSubject := huh.NewInput().
		Title("Write a short and imperative summary of the code changes: (lower case and no period)").
		CharLimit(100).
		Placeholder("Short description of your commit").
		Value(&commitSubject)

	err = promptSubject.Run()
	if err != nil {
		return "", err
	}

	var commitMessage string
	if commitScope == "" {
		commitMessage = fmt.Sprintf("%s: %s", commitType, commitSubject)
	} else {
		commitMessage = fmt.Sprintf("%s (%s): %s", commitType, commitScope, commitSubject)
	}
	return commitMessage, nil
}

func isInSkipQuestions(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
