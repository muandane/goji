package utils

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/muandane/goji/pkg/config"
)

var commitType string
var commitScope string
var commitSubject string

func AskQuestions(config *config.Config, commitType, commitScope, commitSubject string) (string, error) {
	var finalCommitType, finalCommitScope, finalCommitSubject string
	commitTypeOptions := make([]huh.Option[string], len(config.Types))

	for i, ct := range config.Types {
		commitTypeOptions[i] = huh.NewOption[string](fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description), fmt.Sprintf("%s %s", ct.Name, ct.Emoji))
	}
	if commitType != "" {
		finalCommitType = commitType
	} else {
		promptType := huh.NewSelect[string]().
			Title("Select the type of change you are committing:").
			Options(commitTypeOptions...).
			Value(&commitType)
		err := promptType.Run()
		if err != nil {
			return "", err
		}
	}
	if commitScope != "" {
		finalCommitScope = commitScope
	} else {
		if !isInSkipQuestions("Scopes", config.SkipQuestions) {
			promptScope := huh.NewInput().
				Title("What is the scope of this change? (class or file name): (press [enter] to skip)").
				CharLimit(50).
				Placeholder("Example: ci, api, parser").
				Value(&commitScope)

			err := promptScope.Run()
			if err != nil {
				return "", err
			}
		}
		// Only ask for commitScope if not in SkipQuestions
	}
	if commitSubject != "" {
		finalCommitSubject = commitSubject
	} else {
		promptSubject := huh.NewInput().
			Title("Write a short and imperative summary of the code changes: (lower case and no period)").
			CharLimit(100).
			Placeholder("Short description of your commit").
			Value(&commitSubject)

		err := promptSubject.Run()
		if err != nil {
			return "", err
		}
	}

	var commitMessage string
	if commitScope == "" {
		if commitSubject != "" {
			commitMessage = fmt.Sprintf("%s: %s", finalCommitType, finalCommitSubject)
		} else {
			commitMessage = fmt.Sprintf("%s: %s", commitType, commitSubject)
		}
	} else {
		if commitSubject != "" {
			commitMessage = fmt.Sprintf("%s (%s): %s", finalCommitType, finalCommitScope, finalCommitSubject)
		} else {
			commitMessage = fmt.Sprintf("%s (%s): %s", commitType, commitScope, commitSubject)
		}
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
