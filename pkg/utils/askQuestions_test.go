package utils

import (
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
)

type Gitmoji struct {
	Name        string
	Emoji       string
	Description string
}
type CommitType struct {
	Emoji       string
	Code        string
	Description string
	Name        string
}

func TestAskQuestions(t *testing.T) {
	config := &config.Config{
		Types: []models.CommitType{
			{
				Name:        "bug",
				Emoji:       "🐛",
				Description: "Fixing a bug.",
			},
			{
				Name:        "feat",
				Emoji:       "✨",
				Description: "Introducing new features.",
			},
		},
		SkipQuestions: []string{},
	}

	_, err := AskQuestions(config)
	if err != nil {
		if err.Error() == "open /dev/tty: device not configured" {
			t.Skip("Skipping test due to missing tty")
		} else {
			t.Errorf("AskQuestions() error = %v", err)
		}
	}
}

func TestIsInSkipQuestions(t *testing.T) {
	list := []string{"Scopes", "Types"}
	if !isInSkipQuestions("Scopes", list) {
		t.Errorf("isInSkipQuestions() returned unexpected result")
	}
}
