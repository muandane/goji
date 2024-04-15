package config

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/viper"
)

func AddCustomCommitTypes(gitmojis []Gitmoji) []Gitmoji {
	custom := []Gitmoji{
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

	return append(gitmojis, custom...)
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

func SaveConfigToFile(config initConfig, file, dir string) error {
	viper.Set("types", config.Types)
	viper.Set("scopes", config.Scopes)
	viper.Set("skipQuestions", config.SkipQuestions)
	viper.Set("subjectMaxLength", config.SubjectMaxLength)
	viper.Set("signOff", config.SignOff)
	viper.Set("noEmojis", config.NoEmoji)

	viper.SetConfigName(file)
	viper.SetConfigType("json")
	viper.AddConfigPath(dir)

	if err := viper.WriteConfigAs(path.Join(dir, file+".json")); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

func InitRepoConfig(global, repo bool) error {
	gitmojis := AddCustomCommitTypes([]Gitmoji{})
	config := initConfig{
		Types:            gitmojis,
		Scopes:           []string{"home", "accounts", "ci"},
		SkipQuestions:    nil,
		SubjectMaxLength: 100,
		SignOff:          true,
		NoEmoji:          false,
	}

	var location string
	var err error

	switch {
	case global:
		location, err = os.UserHomeDir()
	case repo:
		location, err = GetGitRootDir()
	default:
		return fmt.Errorf("no flag set for location to save configuration file")
	}

	if err != nil {
		return err
	}

	if err = SaveConfigToFile(config, ".goji", location); err != nil {
		return fmt.Errorf("error saving gitmojis to file: %v", err)
	}

	return nil
}
