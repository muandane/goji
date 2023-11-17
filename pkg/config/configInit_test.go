package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveGitmojisToFile(t *testing.T) {
	gitmojis := AddCustomCommitTypes([]Gitmoji{})
	config := initConfig{
		Types:            gitmojis,
		Scopes:           []string{"home", "accounts", "ci"},
		Symbol:           true,
		SkipQuestions:    []string{},
		SubjectMaxLength: 50,
	}

	err := SaveGitmojisToFile(config, ".goji.json")
	if err != nil {
		t.Errorf("Error saving gitmojis to file: %v", err)
	}

	// Get the git root directory
	gitDir, err := GetGitRootDir()
	if err != nil {
		t.Errorf("Error finding git root directory: %v", err)
	}

	// Check that the file was created
	_, err = os.Stat(filepath.Join(gitDir, ".goji.json"))
	if os.IsNotExist(err) {
		t.Error("Expected .goji.json file to be created, but it was not")
	}

	// Clean up
	os.Remove(filepath.Join(gitDir, ".goji.json"))
}
func TestGetGitRootDir(t *testing.T) {
	gitDir, err := GetGitRootDir()
	if err != nil {
		t.Errorf("Error finding git root directory: %v", err)
	}

	if gitDir == "" {
		t.Error("Expected git root directory to be non-empty")
	}
}

func TestInitRepoConfig(t *testing.T) {
	err := InitRepoConfig()
	if err != nil {
		t.Errorf("Error initializing repo config: %v", err)
	}

	// Get the git root directory
	gitDir, err := GetGitRootDir()
	if err != nil {
		t.Errorf("Error finding git root directory: %v", err)
	}

	// Check that the file was created
	_, err = os.Stat(filepath.Join(gitDir, ".goji.json"))
	if os.IsNotExist(err) {
		t.Error("Expected .goji.json file to be created, but it was not")
	}

	// Clean up
	os.Remove(filepath.Join(gitDir, ".goji.json"))
}

func TestAddCustomCommitTypes(t *testing.T) {
	gitmojis := []Gitmoji{}
	customGitmojis := AddCustomCommitTypes(gitmojis)

	if len(customGitmojis) != 11 {
		t.Errorf("Expected 11 gitmojis, got %d", len(customGitmojis))
	}
}
