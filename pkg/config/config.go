package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/muandane/goji/pkg/models"
	"github.com/spf13/viper"
)

// gitmojiToCommitType converts a Gitmoji to models.CommitType
func gitmojiToCommitType(g Gitmoji) models.CommitType {
	return models.CommitType{
		Emoji:       g.Emoji,
		Code:        g.Code,
		Description: g.Description,
		Name:        g.Name,
	}
}

// gitmojisToCommitTypes converts a slice of Gitmoji to []models.CommitType
func gitmojisToCommitTypes(gitmojis []Gitmoji) []models.CommitType {
	commitTypes := make([]models.CommitType, len(gitmojis))
	for i, g := range gitmojis {
		commitTypes[i] = gitmojiToCommitType(g)
	}
	return commitTypes
}

func ViperConfig() (*Config, error) {
	viper.SetConfigName(".goji")
	viper.SetConfigType("json")

	// Ensure Viper searches the current working directory.
	// This is crucial for tests that chdir to a temp directory.
	viper.AddConfigPath(".") // <--- THIS IS THE KEY LINE

	// Attempt to find git root directory for .goji.json
	gitDir, err := GetGitRootDir()
	if err == nil {
		if _, statErr := os.Stat(filepath.Join(gitDir, ".goji.json")); statErr == nil {
			viper.AddConfigPath(gitDir)
		}
	}

	// Attempt to find home directory for .goji.json
	homeDir, homeErr := os.UserHomeDir()
	if homeErr == nil {
		viper.AddConfigPath(homeDir)
	}

	errRead := viper.ReadInConfig()
	if errRead != nil {
		if _, ok := errRead.(viper.ConfigFileNotFoundError); ok {
			searchPaths := []string{"current directory"}
			if gitDir != "" && err == nil {
				searchPaths = append(searchPaths, gitDir)
			}
			if homeDir != "" && homeErr == nil {
				searchPaths = append(searchPaths, homeDir)
			}
			return nil, fmt.Errorf("unable to find .goji.json in specified paths: %s. Error: %w", strings.Join(searchPaths, ", "), errRead)
		} else {
			return nil, fmt.Errorf("error reading config file: %w", errRead)
		}
	}

	var config Config
	if errUnmarshal := viper.Unmarshal(&config); errUnmarshal != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", errUnmarshal)
	}

	// If Types is empty or nil, fallback to default types
	if len(config.Types) == 0 {
		defaultGitmojis := AddCustomCommitTypes([]Gitmoji{})
		config.Types = gitmojisToCommitTypes(defaultGitmojis)
	}

	return &config, nil
}
