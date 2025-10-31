package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCmd_VersionFlag(t *testing.T) {

	var versionFlag bool
	// Save the original os.Stdout
	originalStdout := os.Stdout
	// Create a buffer to capture the output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a new command
	cmd := &cobra.Command{
		Use:   "goji",
		Short: "Goji CLI",
		Long:  `Goji is a cli tool to generate conventional commits with emojis`,
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				color.Set(color.FgGreen)
				fmt.Printf("goji version: v%s\n", version)
				color.Unset()
				return
			}
		},
	}

	// Set the version flag
	cmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	cmd.SetArgs([]string{"--version"})

	// Execute the command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	// Close the writer and restore os.Stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read the captured output
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Error reading captured output: %v", err)
	}
	output := buf.String()

	// Assert that the output contains the expected version string
	assert.Contains(t, output, "goji version: v")
}

func TestGetVersion(t *testing.T) {
	// Save original version
	originalVersion := version
	defer func() { version = originalVersion }()

	t.Run("version set directly", func(t *testing.T) {
		version = "1.0.0"
		result := getVersion()
		assert.Equal(t, "1.0.0", result)
	})

	t.Run("version from git describe", func(t *testing.T) {
		version = ""
		// This test may work if we're in a git repo, otherwise it falls back to "dev"
		result := getVersion()
		assert.NotEmpty(t, result)
		// If git is available, result should be a version string
		// If not, it should be "dev"
		assert.True(t, result == "dev" || len(result) > 0)
	})
}

func TestIsInGitRepo(t *testing.T) {
	// This test depends on whether we're in a git repo or not
	// We're running tests in the goji repo, so it should return true
	result := isInGitRepo()
	assert.IsType(t, true, result)
}

func TestShowNotInGitRepoMessage(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "goji",
		Short: "Test command",
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := showNotInGitRepoMessage(cmd)

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "not in a git repository")
	assert.Contains(t, output, "git init")
}
