package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// captureOutput captures and returns the output of a function that prints to stdout.
func captureOutput(f func()) string {
	// Save the current stdout
	oldStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the function
	f()

	// Close the writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	return buf.String()
}

// TestManCmd tests the `manCmd` to ensure it outputs the manual correctly.
func TestManCmd(t *testing.T) {
	// Capturing the output of the manCmd.
	output := captureOutput(func() {
		manCmd.Run(&cobra.Command{}, []string{})
	})

	// Check that the output contains key sections.
	sections := []string{
		"NAME",
		"SYNOPSIS",
		"DESCRIPTION",
		"OPTIONS",
		"EXAMPLES",
		"AUTHOR",
		"COPYRIGHT",
	}

	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected section %s in man page output, but it was not found.", section)
		}
	}
}
