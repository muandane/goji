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
