package utils

import (
	"errors"
	"goji/pkg/config"
	"goji/pkg/models"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAskQuestions(t *testing.T) {
	testCases := []struct {
		Name           string
		MockAnswers    []interface{}
		ExpectedOutput string
	}{
		{
			Name: "Valid commit type 'feat'",
			MockAnswers: []interface{}{
				"feat :sparkles:",
				"core",
				"Add new feature",
			},
			ExpectedOutput: "feat :sparkles: (core): Add new feature",
		},
		{
			Name: "Valid commit type 'fix'",
			MockAnswers: []interface{}{
				"fix :bug:",
				"core",
				"Fix a bug",
			},
			ExpectedOutput: "fix :bug: (core): Fix a bug",
		},
	}
	// Create a mock survey function
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockAnswers := tc.MockAnswers

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
					{Name: "fix", Emoji: ":bug:", Description: "Fix a bug"},
				},
			}

			commitMessage, err := AskQuestions(mockConfig)
			if err != nil {
				t.Errorf("AskQuestions failed: %v", err)
			}

			if commitMessage != tc.ExpectedOutput {
				t.Errorf("Expected commit message '%s', got '%s'", tc.ExpectedOutput, commitMessage)
			}
		})
	}
}

// MockAskOneFunc is a mock type for askOneFunc
type MockAskOneFunc struct {
	mock.Mock
}

// AskOne is a mock function for askOneFunc
func (m *MockAskOneFunc) AskOne(prompt survey.Prompt, response interface{}, options ...survey.AskOpt) error {
	args := m.Called(prompt, response, options)
	return args.Error(0)
}

func TestAskQuestions_Failure(t *testing.T) {
	// Create a new mock for askOneFunc
	mockAskOne := new(MockAskOneFunc)

	// Simulate an error in the mock function
	mockAskOne.On("AskOne", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("simulated error"))

	// Override the askOneFunc with the mock function
	askOneFunc = mockAskOne.AskOne

	// Restore the original askOneFunc after the test
	defer func() {
		askOneFunc = defaultAskOne
	}()

	mockConfig := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: ":sparkles:", Description: "A new feature"},
			{Name: "fix", Emoji: ":bug:", Description: "Fix a bug"},
		},
	}

	commitMessage, err := AskQuestions(mockConfig)
	if err == nil || err.Error() != "simulated error" {
		t.Errorf("Expected error 'simulated error', got '%v'", err)
	}

	if commitMessage != "" {
		t.Errorf("Expected commit message '', got '%s'", commitMessage)
	}
}
func TestIsInSkipQuestions(t *testing.T) {
	testCases := []struct {
		skipQuestions []string
		value         string
		expected      bool
	}{
		{[]string{"types", "scopes", "message"}, "scopes", true},
		{[]string{"types", "scopes", "message"}, "notInList", false},
		{[]string{}, "scopes", false},
	}

	for _, tc := range testCases {
		result := isInSkipQuestions(tc.value, tc.skipQuestions)
		assert.Equal(t, tc.expected, result)
	}
}
