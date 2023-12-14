package config

import (
	"testing"
)

func TestAddCustomCommitTypes(t *testing.T) {
	gitmojis := AddCustomCommitTypes([]Gitmoji{})
	if len(gitmojis) != 11 {
		t.Errorf("AddCustomCommitTypes() returned unexpected number of Gitmojis")
	}
}

func TestGetGitRootDir(t *testing.T) {
	_, err := GetGitRootDir()
	if err != nil {
		t.Errorf("GetGitRootDir() error = %v", err)
	}
}

func TestSaveGitmojisToFile(t *testing.T) {
	config := initConfig{
		Types:            []Gitmoji{},
		Scopes:           []string{"home", "accounts", "ci"},
		Symbol:           true,
		SkipQuestions:    []string{},
		SubjectMaxLength: 50,
	}
	err := SaveGitmojisToFile(config, ".goji", ".")
	if err != nil {
		t.Errorf("SaveGitmojisToFile() error = %v", err)
	}
}

func TestInitRepoConfig(t *testing.T) {
	err := InitRepoConfig(false, true)
	if err != nil {
		t.Errorf("InitRepoConfig() error = %v", err)
	}
}
