package utils

import (
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
)

func TestAskQuestions(t *testing.T) {
	// Create a mock config for testing purposes
	mockConfig := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨", Description: "A new feature"},
			{Name: "fix", Emoji: "🐛", Description: "A bug fix"},
		},
		SkipQuestions: []string{"Scopes"},
	}

	// Call the function with mock data
	commitMessage, err := AskQuestions(mockConfig, "feat", "", "test commit message")

	// Add assertions based on expected behavior
	if err != nil {
		t.Errorf("AskQuestions returned an error: %v", err)
	}

	// Example assertions for expected output format
	expectedMessage := "feat: test commit message" // Replace with your expected message format
	if commitMessage != expectedMessage {
		t.Errorf("Expected commit message %s, but got %s", expectedMessage, commitMessage)
	}
}
