package utils

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/log"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muandane/goji/pkg/config"
)

func AskQuestions(config *config.Config) ([]string, error) {
	var commitType string
	var commitScope string
	var commitSubject string
	var commitDescription string
	commitTypeOptions := make([]huh.Option[string], len(config.Types))
	var form huh.Form
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
		commitTypeOptions[i] = huh.NewOption[string](row, fmt.Sprintf("%s %s", ct.Name, func() string {
			if !config.NoEmoji {
				return ct.Emoji
			}
			return ""
		}()))
	}

	group1 := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select the type of change you are committing:").
			Options(commitTypeOptions...).
			Height(8).
			Value(&commitType),
		huh.NewInput().
			Title("What is the scope of this change? (class or file name): (press [enter] to skip)").
			CharLimit(50).
			Suggestions(config.Scopes).
			Placeholder("Example: ci, api, parser").
			Value(&commitScope),
	)

	group2 := huh.NewGroup(
		huh.NewInput().
			Title("Write a short and imperative summary of the code changes: (lower case and no period)").
			CharLimit(70).
			Placeholder("Short description of your commit").
			Value(&commitSubject).
			Validate(func(str string) error {
				if len(str) == 0 {
					return errors.New("sorry, subject can't be empty")
				}
				return nil
			}),
		huh.NewText().
			Title("Write a Long description of the code changes: (press [enter] to skip)").
			CharLimit(config.SubjectMaxLength).
			Placeholder("Long description of your commit").
			Value(&commitDescription).
			WithHeight(4),
		huh.NewConfirm().
			Key("done").
			Title("Commit Changes ?").
			Validate(func(v bool) error {
				if !v {
					return fmt.Errorf("welp, finish up then")
				}
				return nil
			}).
			Affirmative("Yes").
			Negative("Wait, no"),
	)

	form = *huh.NewForm(group1, group2)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	var commitMessage string
	var commitBody string
	switch {
	case commitScope != "" && commitDescription != "":
		commitMessage = fmt.Sprintf("%s(%s): %s", commitType, commitScope, commitSubject)
		commitBody = commitDescription
	case commitDescription != "":
		commitMessage = fmt.Sprintf("%s: %s", commitType, commitSubject)
		commitBody = commitDescription
	case commitScope != "":
		commitMessage = fmt.Sprintf("%s(%s): %s", commitType, commitScope, commitSubject)
	default:
		commitMessage = fmt.Sprintf("%s: %s", commitType, commitSubject)
	}

	var result []string
	result = append(result, commitMessage, commitBody)

	// logging the results for debugging purposes
	// log.Infof("result: %s", result)

	return result, nil
}
