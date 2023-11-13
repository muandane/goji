package utils

import (
	"fmt"

	"goji/pkg/config"

	"github.com/AlecAivazis/survey/v2"
)

func AskQuestions(config *config.Config) (string, error) {
	var commitType string
	var commitScope string
	var commitSubject string

	commitTypeOptions := make([]string, len(config.Types))
	for i, ct := range config.Types {
		commitTypeOptions[i] = fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description)
	}

	promptType := &survey.Select{
		Message: "Select the type of change you are committing:",
		Options: commitTypeOptions,
	}
	err := askOneFunc(promptType, &commitType)
	if err != nil {
		return "", err
	}
	for _, ct := range config.Types {
		if fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description) == commitType {
			commitType = fmt.Sprintf("%s %s", ct.Name, ct.Emoji)
			break
		}
	}

	// Only ask for commitScope if not in SkipQuestions
	if !isInSkipQuestions("Scopes", config.SkipQuestions) {
		promptScope := &survey.Input{
			Message: "Enter the scope of the change:",
		}
		err = askOneFunc(promptScope, &commitScope)
		if err != nil {
			return "", err
		}
	}

	promptSubject := &survey.Input{
		Message: "Enter a short description of the change:",
	}
	err = askOneFunc(promptSubject, &commitSubject)
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
