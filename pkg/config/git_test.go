package config

import "testing"

func TestGitRepo(t *testing.T) {
	repo, err := GitRepo()
	if err != nil {
		t.Errorf("Error finding git root directory: %v", err)
	}

	// Check if the repo path is not empty
	if repo == "" {
		t.Errorf("Expected a repository path, got an empty string")
	}
}
