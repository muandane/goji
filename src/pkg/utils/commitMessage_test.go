package utils

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"goji/pkg/config"
	"goji/pkg/models"
	"testing"
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

// func TestAskQuestions_Failure(t *testing.T) {
// 	testCases := []struct {
// 		Name           string
// 		MockAnswers    []interface{}
// 		ExpectedOutput string
// 	}{
// 		{
// 			Name: "Valid commit type 'feat'",
// 			MockAnswers: []interface{}{
// 				"feat :sparkles:",
// 				"core",
// 				"Add new feature",
// 			},
// 			ExpectedOutput: "This will cause the test to fail",
// 		},
// 		{
// 			Name: "Valid commit type 'fix'",
// 			MockAnswers: []interface{}{
// 				"fix :bug:",
// 				"core",
// 				"Fix a bug",
// 			},
// 			ExpectedOutput: "fix :bug: (core): Fix a bug",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.Name, func(t *testing.T) {
// 			mockAnswers := tc.MockAnswers

// 			mockAskOne := func(prompt survey.Prompt, response interface{}, options ...survey.AskOpt) error {
// 				// Simulate an error when commit type is 'feat'
// 				if len(mockAnswers) > 0 && mockAnswers[0] == "feat :sparkles:" {
// 					return errors.New("simulated error")
// 				}

// 				if len(mockAnswers) == 0 {
// 					return errors.New("no more answers available")
// 				}

// 				answer := mockAnswers[0]
// 				mockAnswers = mockAnswers[1:]
// 				switch v := response.(type) {
// 				case *string:
// 					*v = answer.(string)
// 				default:
// 					return errors.New("unsupported response type")
// 				}

// 				return nil
// 			}

// 			// Override the askOneFunc with the mock function
// 			askOneFunc = mockAskOne

// 			// Restore the original askOneFunc after the test
// 			defer func() {
// 				askOneFunc = defaultAskOne
// 			}()

// 			mockConfig := &config.Config{
// 				Types: []models.CommitType{
// 					{Name: "feat", Emoji: ":sparkles:", Description: "A new feature"},
// 					{Name: "fix", Emoji: ":bug:", Description: "Fix a bug"},
// 				},
// 			}

// 			commitMessage, err := AskQuestions(mockConfig)
// 			if err != nil {
// 				t.Errorf("AskQuestions failed: %v", err)
// 			}

// 			if commitMessage != tc.ExpectedOutput {
// 				t.Errorf("Expected commit message '%s', got '%s'", tc.ExpectedOutput, commitMessage)
// 			}
// 		})
// 	}
// }
