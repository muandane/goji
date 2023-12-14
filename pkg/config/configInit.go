package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
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
	viper.Set("Types", config.Types)
	viper.Set("Scopes", config.Scopes)
	viper.Set("Symbol", config.Symbol)
	viper.Set("SkipQuestions", config.SkipQuestions)
	viper.Set("SubjectMaxLength", config.SubjectMaxLength)

	viper.SetConfigName(filename) // name of config file (without extension)
	viper.SetConfigType("json")   // specifying the config type
	viper.AddConfigPath(dir)      // path to look for the config file in

	err := viper.SafeWriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
			err = viper.WriteConfig()
			if err != nil {
				return fmt.Errorf("error writing config file: %v", err)
			}
		} else {
			return fmt.Errorf("error creating config file: %v", err)
		}
	}

	return nil
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

	err = SaveGitmojisToFile(config, ".goji", dir)

	if err != nil {
		return fmt.Errorf("error saving gitmojis to file: %v", err)
	} else {
		fmt.Println("Gitmojis saved to .goji.json 🎊")
	}

	return nil
}
