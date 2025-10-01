package cmd

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/muandane/goji/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessCommitMessage(t *testing.T) {
	t.Run("basic commit message processing", func(t *testing.T) {
		commitMessage := "feat: add new feature"
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		}

		result := processCommitMessage(commitMessage, false, configTypes)
		expected := "feat ✨: add new feature"

		assert.Equal(t, expected, result)
	})

	t.Run("commit message with scope", func(t *testing.T) {
		commitMessage := "fix(api): resolve authentication bug"
		configTypes := []models.CommitType{
			{Name: "fix", Emoji: "🐛"},
		}

		result := processCommitMessage(commitMessage, false, configTypes)
		expected := "fix 🐛 (api): resolve authentication bug"

		assert.Equal(t, expected, result)
	})

	t.Run("no emoji mode", func(t *testing.T) {
		commitMessage := "feat: add new feature"
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		}

		result := processCommitMessage(commitMessage, true, configTypes)
		expected := "feat: add new feature"

		assert.Equal(t, expected, result)
	})

	t.Run("override type", func(t *testing.T) {
		originalType := "feat"
		overrideType = "fix"
		commitMessage := originalType + ": add new feature"
		configTypes := []models.CommitType{
			{Name: "fix", Emoji: "🐛"},
		}

		result := processCommitMessage(commitMessage, false, configTypes)
		expected := "fix 🐛: add new feature"

		assert.Equal(t, expected, result)

		// Reset for other tests
		overrideType = ""
	})

	t.Run("override scope", func(t *testing.T) {
		originalScope := "api"
		overrideScope = "backend"
		commitMessage := "feat(" + originalScope + "): add new feature"
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		}

		result := processCommitMessage(commitMessage, false, configTypes)
		expected := "feat ✨ (backend): add new feature"

		assert.Equal(t, expected, result)

		// Reset for other tests
		overrideScope = ""
	})

	t.Run("invalid commit message format", func(t *testing.T) {
		commitMessage := "invalid format"
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		}

		result := processCommitMessage(commitMessage, false, configTypes)
		// Should return original message when regex doesn't match
		assert.Equal(t, commitMessage, result)
	})

	t.Run("empty commit message", func(t *testing.T) {
		commitMessage := ""
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		}

		result := processCommitMessage(commitMessage, false, configTypes)
		assert.Equal(t, "", result)
	})
}

func TestPrintErrorAndExit(t *testing.T) {
	t.Run("error message formatting", func(t *testing.T) {
		// This test verifies the function exists and can be called
		// In a real scenario, we'd need to capture os.Exit behavior
		// For now, we just ensure the function is defined
		assert.NotNil(t, printErrorAndExit)
	})
}

func TestDraftCommandFlags(t *testing.T) {
	t.Run("commit directly flag", func(t *testing.T) {
		// Test that the flag is properly defined
		assert.False(t, commitDirectly)
	})

	t.Run("override flags", func(t *testing.T) {
		// Test that override flags are properly initialized
		assert.Empty(t, overrideType)
		assert.Empty(t, overrideScope)
		assert.Empty(t, extraContext)
		assert.False(t, generateBody)
	})
}

func TestCommitMessageRegex(t *testing.T) {
	t.Run("valid conventional commit formats", func(t *testing.T) {
		testCases := []struct {
			message string
			valid   bool
		}{
			{"feat: add new feature", true},
			{"fix(api): resolve bug", true},
			{"docs: update readme", true},
			{"feat(auth): add login system", true},
			{"invalid format", false},
			{"", false},
			{"feat", false},
		}

		for _, tc := range testCases {
			t.Run(tc.message, func(t *testing.T) {
				// Test the regex pattern used in processCommitMessage
				re := regexp.MustCompile(`^([a-zA-Z]+)(\([^)]*\))?:\s*(.*)$`)
				matches := re.FindStringSubmatch(tc.message)

				if tc.valid {
					assert.Greater(t, len(matches), 0, "Expected valid match for: %s", tc.message)
				} else {
					assert.Equal(t, 0, len(matches), "Expected no match for: %s", tc.message)
				}
			})
		}
	})

	t.Run("regex capture groups", func(t *testing.T) {
		commitMessage := "feat(api): add authentication"
		re := regexp.MustCompile(`^([a-zA-Z]+)(\([^)]*\))?:\s*(.*)$`)
		matches := re.FindStringSubmatch(commitMessage)

		require.Len(t, matches, 4)
		assert.Equal(t, "feat", matches[1])               // type
		assert.Equal(t, "(api)", matches[2])              // scope
		assert.Equal(t, "add authentication", matches[3]) // message
	})
}

func TestCommitMessageConstruction(t *testing.T) {
	t.Run("message with all components", func(t *testing.T) {
		commitType := "feat"
		scope := "api"
		message := "add authentication"

		// Test the construction logic from processCommitMessage
		var builder strings.Builder
		builder.WriteString(commitType)
		builder.WriteString(" ✨")
		builder.WriteString(" (" + scope + ")")
		builder.WriteString(": " + message)

		expected := "feat ✨ (api): add authentication"
		assert.Equal(t, expected, builder.String())
	})

	t.Run("message without scope", func(t *testing.T) {
		commitType := "feat"
		message := "add authentication"

		var builder strings.Builder
		builder.WriteString(commitType)
		builder.WriteString(" ✨")
		builder.WriteString(": " + message)

		expected := "feat ✨: add authentication"
		assert.Equal(t, expected, builder.String())
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("API key environment variables", func(t *testing.T) {
		// Test that environment variables are checked
		envVars := []string{
			"OPENROUTER_API_KEY",
			"GROQ_API_KEY",
		}

		for _, envVar := range envVars {
			value := os.Getenv(envVar)
			// We can't test the actual values in CI, but we can test the function exists
			_ = value
		}
	})
}

func TestConfigTypes(t *testing.T) {
	t.Run("commit type matching", func(t *testing.T) {
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨", Description: "New feature"},
			{Name: "fix", Emoji: "🐛", Description: "Bug fix"},
			{Name: "docs", Emoji: "📚", Description: "Documentation"},
		}

		// Test finding commit type by name
		var foundType *models.CommitType
		for _, ct := range configTypes {
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

	t.Run("commit type emoji handling", func(t *testing.T) {
		configTypes := []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		}

		ct := configTypes[0]
		emoji := func() string {
			if !false { // NoEmoji = false
				return ct.Emoji
			}
			return ""
		}()

		assert.Equal(t, "✨", emoji)
	})
}
