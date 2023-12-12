package utils

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/muandane/goji/pkg/config"
)

func AskQuestions(config *config.Config) (string, error) {
	var commitType string
	var commitScope string
	var commitSubject string
	commitTypeOptions := make([]huh.Option[string], len(config.Types))

	for i, ct := range config.Types {
		commitTypeOptions[i] = huh.NewOption[string](fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description), ct.Name)
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
			Title("Enter the scope of the change:").
			Value(&commitScope)

		err = promptScope.Run()
		if err != nil {
			return "", err
		}
	}

	promptSubject := huh.NewInput().
		Title("Enter a short description of the change:").
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
