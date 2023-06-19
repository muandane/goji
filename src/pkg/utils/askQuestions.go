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
	err := survey.AskOne(promptType, &commitType)
	if err != nil {
		return "", err
	}
	for _, ct := range config.Types {
		if fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description) == commitType {
			commitType = fmt.Sprintf("%s %s", ct.Name, ct.Emoji)
			break
		}
	}

	promptScope := &survey.Input{
		Message: "Enter the scope of the change:",
	}
	err = survey.AskOne(promptScope, &commitScope)
	if err != nil {
		return "", err
	}

	promptSubject := &survey.Input{
		Message: "Enter a short description of the change:",
	}
	err = survey.AskOne(promptSubject, &commitSubject)
	if err != nil {
		return "", err
	}

	commitMessage := fmt.Sprintf("%s (%s): %s", commitType, commitScope, commitSubject)
	return commitMessage, nil
}
