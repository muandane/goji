package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func LoadConfig(filename string) (*Config, error) {
	// Get the root directory of the Git project
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	rootDirBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error finding git root directory: %v", err)
	}
	rootDir := string(rootDirBytes)
	rootDir = strings.TrimSpace(rootDir) // Remove newline character at the end

	// Try to load the config file from the root of the Git project
	configFile := filepath.Join(rootDir, filename)
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		// If not found, try to load it from the home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error finding home directory: %v", err)
		}
		configFile = filepath.Join(homeDir, filename)
		data, err = ioutil.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
