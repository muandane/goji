package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AddCustomCommitTypes(gitmojis []Gitmoji) []Gitmoji {
	customGitmojis := []Gitmoji{
		{Emoji: "✨", Code: ":sparkles:", Description: "Introduce new features.", Name: "feat"},
		{Emoji: "🐛", Code: ":bug:", Description: "Fix a bug.", Name: "fix"},
		{Emoji: "📚", Code: ":books:", Description: "Documentation change.", Name: "docs"},
		{Emoji: "🎨", Code: ":art:", Description: "Improve structure/format of the code.", Name: "refactor"},
		{Emoji: "🧹", Code: ":broom:", Description: "A chore change.", Name: "chore"},
		{Emoji: "🧪", Code: ":test_tube:", Description: "Add a test.", Name: "test"},
		{Emoji: "🚑️", Code: ":ambulance:", Description: "Critical hotfix.", Name: "hotfix"},
		{Emoji: "⚰️", Code: ":coffin:", Description: "Remove dead code.", Name: "deprecate"},
		{Emoji: "⚡️", Code: ":zap:", Description: "Improve performance.", Name: "perf"},
		{Emoji: "🚧", Code: ":construction:", Description: "Work in progress.", Name: "wip"},
		{Emoji: "📦", Code: ":package:", Description: "Add or update compiled files or packages.", Name: "package"},
	}

	return append(gitmojis, customGitmojis...)
}
func GetGitRootDir() (string, error) {
	gitRoot := exec.Command("git", "rev-parse", "--show-toplevel")
	gitDirBytes, err := gitRoot.Output()
	if err != nil {
		return "", fmt.Errorf("error finding git root directory: %v", err)
	}
	gitDir := string(gitDirBytes)
	gitDir = strings.TrimSpace(gitDir) // Remove newline character at the end

	return gitDir, nil
}

func SaveGitmojisToFile(config initConfig, filename string, dir string) error {
	configFile := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
func InitRepoConfig(global bool, repo bool) error {
	gitmojis := AddCustomCommitTypes([]Gitmoji{})
	config := initConfig{
		Types:            gitmojis,
		Scopes:           []string{"home", "accounts", "ci"},
		Symbol:           true,
		SkipQuestions:    []string{},
		SubjectMaxLength: 50,
	}

	var dir string
	var err error
	if global {
		dir, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	} else if repo {
		dir, err = GetGitRootDir()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no flag set for location to save configuration file")
	}

	err = SaveGitmojisToFile(config, ".goji.json", dir)

	if err != nil {
		return fmt.Errorf("error saving gitmojis to file: %v", err)
	} else {
		fmt.Println("Gitmojis saved to .goji.json 🎊")
	}

	return nil
}
