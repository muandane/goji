package config

import (
	"testing"
)

func TestAddCustomCommitTypes(t *testing.T) {
	gitmojis := AddCustomCommitTypes([]Gitmoji{})
	if len(gitmojis) != 11 {
		t.Errorf("Expected 11 gitmojis, got %d", len(gitmojis))
	}
}

func TestGetGitRootDir(t *testing.T) {
	dir, err := GetGitRootDir()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dir == "" {
		t.Errorf("Expected directory, got empty string")
	}
}
