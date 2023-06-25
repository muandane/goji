package utils

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"goji/pkg/config"
	"goji/pkg/models"
	"testing"
)

func TestAskQuestions(t *testing.T) {
	// Create a mock survey function
	mockAnswers := []interface{}{
		"feat :sparkles:",
		"core",
		"Add new feature",
	}

	mockAskOne := func(prompt survey.Prompt, response interface{}, options ...survey.AskOpt) error {
		if len(mockAnswers) == 0 {
			return errors.New("no more answers available")
		}

		answer := mockAnswers[0]
		mockAnswers = mockAnswers[1:]
		switch v := response.(type) {
		case *string:
			*v = answer.(string)
		default:
			return errors.New("unsupported response type")
		}

		return nil
	}

	// Override the askOneFunc with the mock function
	askOneFunc = mockAskOne

	// Restore the original askOneFunc after the test
	defer func() {
		askOneFunc = defaultAskOne
	}()

	mockConfig := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: ":sparkles:", Description: "A new feature"},
		},
	}

	expectedCommitMessage := "feat :sparkles: (core): Add new feature"
	commitMessage, err := AskQuestions(mockConfig)
	if err != nil {
		t.Errorf("AskQuestions failed: %v", err)
	}

	if commitMessage != expectedCommitMessage {
		t.Errorf("Expected commit message '%s', got '%s'", expectedCommitMessage, commitMessage)
	}
}
