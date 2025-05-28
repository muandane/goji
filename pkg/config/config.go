package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

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

	return &config, nil
}
