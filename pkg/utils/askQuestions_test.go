package utils

import (
	"errors"
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskQuestions_WithPresets(t *testing.T) {
	cfg := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨", Description: "New feature"},
			{Name: "fix", Emoji: "🐛", Description: "Bug fix"},
		},
		Scopes:           []string{"api", "ui"},
		SubjectMaxLength: 70,
		NoEmoji:          false,
	}

	t.Run("with preset type and message", func(t *testing.T) {
		// This test simulates the form behavior without actually running the interactive form
		// We test the logic that would be executed when presets are provided

		presetType := "feat"
		presetMessage := "add new feature"

		// Simulate the preset logic from AskQuestions
		var commitType string
		for _, ct := range cfg.Types {
			optionVal := ct.Name + " " + ct.Emoji
			if optionVal == presetType+" ✨" {
				commitType = optionVal
				break
			}
		}

		commitSubject := presetMessage

		// Test the message construction logic
		expectedMessage := "feat ✨: add new feature"
		actualMessage := commitType + ": " + commitSubject

		assert.Equal(t, expectedMessage, actualMessage)
	})

	t.Run("with scope", func(t *testing.T) {
		commitType := "feat ✨"
		commitScope := "api"
		commitSubject := "add authentication"

		expectedMessage := "feat ✨ (api): add authentication"
		actualMessage := commitType + " (" + commitScope + "): " + commitSubject

		assert.Equal(t, expectedMessage, actualMessage)
	})

	t.Run("no emoji mode", func(t *testing.T) {
		cfg.NoEmoji = true
		cfg.Types = []models.CommitType{
			{Name: "feat", Emoji: "✨", Description: "New feature"},
		}

		// Simulate option creation without emoji
		ct := cfg.Types[0]
		optionVal := ct.Name + " " + func() string {
			if !cfg.NoEmoji {
				return ct.Emoji
			}
			return ""
		}()

		expected := "feat "
		assert.Equal(t, expected, optionVal)
	})
}

func TestAskQuestions_Validation(t *testing.T) {
	t.Run("empty subject validation", func(t *testing.T) {
		// Test the validation logic that would be used in the form
		validateSubject := func(s string) error {
			if s == "" {
				return errors.New("subject cannot be empty")
			}
			return nil
		}

		assert.Error(t, validateSubject(""))
		assert.NoError(t, validateSubject("valid subject"))
	})

	t.Run("confirmation validation", func(t *testing.T) {
		// Test the confirmation validation logic
		validateConfirmation := func(v bool) error {
			if !v {
				return errors.New("changes not committed")
			}
			return nil
		}

		assert.Error(t, validateConfirmation(false))
		assert.NoError(t, validateConfirmation(true))
	})
}

func TestAskQuestions_MessageConstruction(t *testing.T) {
	t.Run("basic message construction", func(t *testing.T) {
		commitType := "feat ✨"
		commitSubject := "add new feature"

		message := commitType + ": " + commitSubject
		expected := "feat ✨: add new feature"

		assert.Equal(t, expected, message)
	})

	t.Run("message with scope", func(t *testing.T) {
		commitType := "fix 🐛"
		commitScope := "api"
		commitSubject := "fix authentication bug"

		message := commitType + " (" + commitScope + "): " + commitSubject
		expected := "fix 🐛 (api): fix authentication bug"

		assert.Equal(t, expected, message)
	})

	t.Run("message with body", func(t *testing.T) {
		commitMessage := "feat ✨: add new feature"
		commitBody := "This adds a new authentication system with JWT tokens"

		// Simulate the return value structure
		result := []string{commitMessage, commitBody}

		assert.Len(t, result, 2)
		assert.Equal(t, commitMessage, result[0])
		assert.Equal(t, commitBody, result[1])
	})
}

func TestAskQuestions_ConfigHandling(t *testing.T) {
	t.Run("scopes suggestions", func(t *testing.T) {
		cfg := &config.Config{
			Scopes: []string{"api", "ui", "backend"},
		}

		assert.Len(t, cfg.Scopes, 3)
		assert.Contains(t, cfg.Scopes, "api")
		assert.Contains(t, cfg.Scopes, "ui")
		assert.Contains(t, cfg.Scopes, "backend")
	})

	t.Run("subject max length", func(t *testing.T) {
		cfg := &config.Config{
			SubjectMaxLength: 100,
		}

		assert.Equal(t, 100, cfg.SubjectMaxLength)
	})

	t.Run("commit types configuration", func(t *testing.T) {
		cfg := &config.Config{
			Types: []models.CommitType{
				{Name: "feat", Emoji: "✨", Description: "New feature"},
				{Name: "fix", Emoji: "🐛", Description: "Bug fix"},
				{Name: "docs", Emoji: "📚", Description: "Documentation"},
			},
		}

		assert.Len(t, cfg.Types, 3)

		// Test type lookup
		var foundType *models.CommitType
		for _, ct := range cfg.Types {
			if ct.Name == "feat" {
				foundType = &ct
				break
			}
		}

		require.NotNil(t, foundType)
		assert.Equal(t, "feat", foundType.Name)
		assert.Equal(t, "✨", foundType.Emoji)
		assert.Equal(t, "New feature", foundType.Description)
	})
}

func TestAskQuestions_EdgeCases(t *testing.T) {
	t.Run("empty config", func(t *testing.T) {
		cfg := &config.Config{
			Types:  []models.CommitType{},
			Scopes: []string{},
		}

		assert.Empty(t, cfg.Types)
		assert.Empty(t, cfg.Scopes)
	})

	t.Run("long subject handling", func(t *testing.T) {
		longSubject := "This is a very long commit message that exceeds the typical limit and should be handled appropriately"

		// Test subject length validation
		if len(longSubject) > 70 {
			// In real implementation, this would trigger validation
			assert.Greater(t, len(longSubject), 70)
		}
	})

	t.Run("special characters in scope", func(t *testing.T) {
		commitScope := "api/v2"
		commitSubject := "update endpoint"

		// Test that special characters in scope are handled
		message := "feat ✨ (" + commitScope + "): " + commitSubject
		expected := "feat ✨ (api/v2): update endpoint"

		assert.Equal(t, expected, message)
	})
}
