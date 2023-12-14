package utils

import (
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
)

func TestIsInSkipQuestions(t *testing.T) {
	list := []string{"Scopes", "Types"}
	if !isInSkipQuestions("Scopes", list) {
		t.Errorf("isInSkipQuestions() returned unexpected result")
	}
}

func convertGitmojiToCommitType(gitmojis []config.Gitmoji) []models.CommitType {
	commitTypes := make([]models.CommitType, len(gitmojis))
	for i, gitmoji := range gitmojis {
		commitTypes[i] = models.CommitType{
			Emoji:       gitmoji.Emoji,
			Code:        gitmoji.Code,
			Description: gitmoji.Description,
			Name:        gitmoji.Name,
		}
	}
	return commitTypes
}
