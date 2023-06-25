package config

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	rootDirBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error finding git root directory: %v", err)
	}
	rootDir := string(rootDirBytes)
	rootDir = strings.TrimSpace(rootDir) // Remove newline character at the end

	// Prepare a temporary configuration file for testing
	filename := "test_config.json"
	content := `{
		"Types": [
			{
				"Emoji": "✨",
				"Code": ":sparkles:",
				"Description": "Introducing new features.",
				"Name": "feat"
			},
			{
				"Emoji": "🐛",
				"Code": ":bug:",
				"Description": "Fixing a bug.",
				"Name": "fix"
			},
			{
				"Emoji": "🧹",
				"Code": ":broom:",
				"Description": "A chore change.",
				"Name": "chore"
			}
		],
		"Scopes": ["home", "accounts", "ci"],
		"Symbol": true,
		"SkipQuestions": [],
		"SubjectMaxLength": 50
	}
	`

	testConfigPath := filepath.Join(rootDir, filename)
	err = ioutil.WriteFile(testConfigPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(testConfigPath)

	// Test the LoadConfig function
	config, err := LoadConfig(filename)
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("config is nil")
	}

	expectedTypeName := "feat"
	if config.Types[0].Name != expectedTypeName {
		t.Errorf("Expected type name %s, got %s", expectedTypeName, config.Types[0].Name)
	}
}
