package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a temporary config file
func createTempGojiConfig(t *testing.T, dir string, types []config.Gitmoji) string {
	t.Helper()
	cfg := struct {
		Types []config.Gitmoji `json:"types"`
	}{
		Types: types,
	}
	content, err := json.Marshal(cfg)
	assert.NoError(t, err)

	configFile := filepath.Join(dir, ".goji.json")
	err = os.WriteFile(configFile, content, 0644)
	assert.NoError(t, err)
	return configFile
}

// Helper function to capture stdout
func captureCheckOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestCheckCmd_FromFile(t *testing.T) {
	tempDir := t.TempDir()
	// UPDATED: Add "docs" type to the temporary config for testing
	_ = createTempGojiConfig(t, tempDir, []config.Gitmoji{
		{Name: "feat", Emoji: "✨"},
		{Name: "fix", Emoji: "🐛"},
		{Name: "docs", Emoji: "📚"}, // Add docs type
	})

	// Create a dummy commit message file
	commitMsgContent := "feat: add new feature"
	commitMsgFile := filepath.Join(tempDir, "COMMIT_MSG")
	err := os.WriteFile(commitMsgFile, []byte(commitMsgContent), 0644)
	assert.NoError(t, err)

	// Store original os.Args and defer restoration [cite: 2]
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Store original working directory and defer restoration [cite: 2]
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Chdir(originalWd)) }()

	// Change working directory to tempDir so .goji.json is found
	assert.NoError(t, os.Chdir(tempDir))

	tests := []struct {
		name           string
		fileContent    string
		expectedOutput string
		setupArgs      func() // Function to set os.Args for the test case
	}{
		{
			name:        "valid commit message from file",
			fileContent: "feat: add new feature",
			setupArgs: func() {
				// Simulate command: goji check -f ./COMMIT_MSG
				// os.Args[0] is program name, os.Args[1] is subcommand, os.Args[2] is flag, os.Args[3] is filepath [cite: 3]
				os.Args = []string{"goji", "check", "-f", commitMsgFile}
			},
			expectedOutput: "Success: Your commit message follows the conventional commit format:",
		},
		{
			name:        "valid commit message with scope from file",
			fileContent: "fix(api): resolve issue",
			setupArgs: func() {
				os.Args = []string{"goji", "check", "-f", commitMsgFile}
			},
			expectedOutput: "Success: Your commit message follows the conventional commit format:",
		},
		{
			name:        "invalid type from file",
			fileContent: "invalidtype: this will fail",
			setupArgs: func() {
				os.Args = []string{"goji", "check", "-f", commitMsgFile}
			},
			expectedOutput: "Error: Commit message type is invalid.",
		},
		{
			name:        "missing colon from file",
			fileContent: "feat this will fail",
			setupArgs: func() {
				os.Args = []string{"goji", "check", "-f", commitMsgFile}
			},
			expectedOutput: "Error: Commit message does not follow the conventional commit format.",
		},
		{
			name:        "empty description from file",
			fileContent: "docs:",
			setupArgs: func() {
				os.Args = []string{"goji", "check", "-f", commitMsgFile}
			},
			expectedOutput: "Error: Commit message description is empty.",
		},
		{
			name:        "empty scope from file",
			fileContent: "feat(): empty scope",
			setupArgs: func() {
				os.Args = []string{"goji", "check", "-f", commitMsgFile}
			},
			expectedOutput: "Error: Commit message scope is empty.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update commit message file content for current test
			err := os.WriteFile(commitMsgFile, []byte(tt.fileContent), 0644)
			assert.NoError(t, err)

			tt.setupArgs() // Set up os.Args for this specific test case [cite: 5]

			cmd := &cobra.Command{Use: "check"}
			cmd.Flags().BoolP("from-file", "f", false, "") // Ensure flag is registered

			// Simulate flag parsing
			err = cmd.Flags().Parse(os.Args[1:]) // Parse flags starting from "check"
			assert.NoError(t, err)

			output := captureCheckOutput(func() {
				checkCmd.Run(cmd, []string{}) // args to Run are positional, not flags
			})
			assert.Contains(t, output, tt.expectedOutput)
		})
	}
}
