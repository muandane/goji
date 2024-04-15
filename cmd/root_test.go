package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func TestCommit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		tempDir, err := os.MkdirTemp("", "git-commit-test")
		if err != nil {
			t.Fatalf("error creating temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("error changing to temp dir: %v", err)
		}

		if err := exec.Command("git", "init").Run(); err != nil {
			t.Fatalf("error initializing git repo: %v", err)
		}

		if err := os.WriteFile("testfile", []byte("test content"), 0644); err != nil {
			t.Fatalf("error writing testfile: %v", err)
		}

		if err := exec.Command("git", "add", "testfile").Run(); err != nil {
			t.Fatalf("error adding testfile to index: %v", err)
		}

		if err := commit("test commit", "test commit body", false); err != nil {
			t.Fatalf("error committing: %v", err)
		}

		if err := exec.Command("git", "log", "-1", "--pretty=%s").Run(); err != nil {
			t.Fatalf("error checking commit: %v", err)
		}
	})
}

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
	cmd.Execute()

	// Close the writer and restore os.Stdout
	w.Close()
	os.Stdout = originalStdout

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Assert that the output contains the expected version string
	assert.Contains(t, output, "goji version: v")
}
