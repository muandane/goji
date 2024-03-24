package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestViperConfig(t *testing.T) {
	gitDir, err := os.MkdirTemp("", "git")
	if err != nil {
		t.Fatalf("Failed to create temporary git directory: %v", err)
	}
	defer os.RemoveAll(gitDir)

	t.Run("Config file exists in git directory", func(t *testing.T) {
		configFile := filepath.Join(gitDir, ".goji.json")
		err = os.WriteFile(configFile, []byte(`{"key": "value"}`), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary config file: %v", err)
		}

		config, err := ViperConfig()
		if err != nil {
			t.Errorf("ViperConfig returned an error: %v", err)
		}
		if config == nil {
			t.Error("Expected a non-nil Config, but got nil")
		}
	})

	t.Run("Config file exists in home directory", func(t *testing.T) {
		homeDir, err := os.MkdirTemp("", "home")
		if err != nil {
			t.Fatalf("Failed to create temporary home directory: %v", err)
		}
		defer os.RemoveAll(homeDir)

		configFile := filepath.Join(homeDir, ".goji.json")
		err = os.WriteFile(configFile, []byte(`{"key": "value"}`), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary config file: %v", err)
		}

		os.Setenv("HOME", homeDir)

		config, err := ViperConfig()
		if err != nil {
			t.Errorf("ViperConfig returned an error: %v", err)
		}
		if config == nil {
			t.Error("Expected a non-nil Config, but got nil")
		}
	})

}
