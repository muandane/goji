package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCmd_Structure(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		assert.NotNil(t, completionCmd)
	})

	t.Run("command use", func(t *testing.T) {
		assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use)
	})

	t.Run("command short description", func(t *testing.T) {
		assert.NotEmpty(t, completionCmd.Short)
	})

	t.Run("command long description", func(t *testing.T) {
		assert.NotEmpty(t, completionCmd.Long)
		assert.Contains(t, completionCmd.Long, "bash")
		assert.Contains(t, completionCmd.Long, "zsh")
		assert.Contains(t, completionCmd.Long, "fish")
		assert.Contains(t, completionCmd.Long, "powershell")
	})

	t.Run("valid args", func(t *testing.T) {
		expectedArgs := []string{"bash", "zsh", "fish", "powershell"}
		assert.Equal(t, expectedArgs, completionCmd.ValidArgs)
	})

	t.Run("disable flags in use line", func(t *testing.T) {
		assert.True(t, completionCmd.DisableFlagsInUseLine)
	})
}

func TestCompletionCmd_GenerateBash(t *testing.T) {
	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a test command to generate completion for
	testCmd := &cobra.Command{
		Use: "testcmd",
	}
	testCmd.AddCommand(&cobra.Command{Use: "subcmd"})

	// Replace rootCmd temporarily
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	// Test bash completion generation
	err := testCmd.GenBashCompletion(os.Stdout)
	
	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "testcmd")
}

func TestCompletionCmd_GenerateZsh(t *testing.T) {
	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a test command
	testCmd := &cobra.Command{
		Use: "testcmd",
	}
	testCmd.AddCommand(&cobra.Command{Use: "subcmd"})

	// Test zsh completion generation
	err := testCmd.GenZshCompletion(os.Stdout)
	
	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "testcmd")
}

func TestCompletionCmd_GenerateFish(t *testing.T) {
	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a test command
	testCmd := &cobra.Command{
		Use: "testcmd",
	}
	testCmd.AddCommand(&cobra.Command{Use: "subcmd"})

	// Test fish completion generation
	err := testCmd.GenFishCompletion(os.Stdout, true)
	
	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "testcmd")
}

func TestCompletionCmd_GeneratePowerShell(t *testing.T) {
	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a test command
	testCmd := &cobra.Command{
		Use: "testcmd",
	}
	testCmd.AddCommand(&cobra.Command{Use: "subcmd"})

	// Test powershell completion generation
	err := testCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	
	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, strings.ToLower(output), "testcmd")
}

func TestCompletionCmd_ArgsValidation(t *testing.T) {
	t.Run("valid shell names", func(t *testing.T) {
		validShells := []string{"bash", "zsh", "fish", "powershell"}
		for _, shell := range validShells {
			// Test that the args validator accepts valid shells
			// The Args field uses cobra.MatchAll with OnlyValidArgs
			err := completionCmd.Args(completionCmd, []string{shell})
			assert.NoError(t, err, "Shell %s should be valid", shell)
		}
	})

	t.Run("invalid shell name", func(t *testing.T) {
		err := completionCmd.Args(completionCmd, []string{"invalid"})
		assert.Error(t, err)
	})

	t.Run("exact args requirement", func(t *testing.T) {
		// Test with no args
		err := completionCmd.Args(completionCmd, []string{})
		assert.Error(t, err)

		// Test with too many args
		err = completionCmd.Args(completionCmd, []string{"bash", "extra"})
		assert.Error(t, err)
	})
}

func TestCompletionCmd_Run(t *testing.T) {
	t.Run("run bash completion", func(t *testing.T) {
		// Capture stdout
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Create a test command to avoid side effects
		testCmd := &cobra.Command{
			Use: "testcmd",
		}
		testCmd.AddCommand(&cobra.Command{Use: "subcmd"})

		// Create a temporary completion command for testing

		// Temporarily create a completion command that uses testCmd
		tempCompletionCmd := &cobra.Command{
			Use:       "completion [bash|zsh|fish|powershell]",
			ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
			Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
			Run: func(cmd *cobra.Command, args []string) {
				switch args[0] {
				case "bash":
					_ = testCmd.GenBashCompletion(os.Stdout)
				case "zsh":
					_ = testCmd.GenZshCompletion(os.Stdout)
				case "fish":
					_ = testCmd.GenFishCompletion(os.Stdout, true)
				case "powershell":
					_ = testCmd.GenPowerShellCompletionWithDesc(os.Stdout)
				}
			},
		}

		// Test bash
		tempCompletionCmd.Run(tempCompletionCmd, []string{"bash"})

		// Close writer and restore stdout
		_ = w.Close()
		os.Stdout = originalStdout

		// Read captured output
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "testcmd")
	})

	t.Run("run zsh completion", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		testCmd := &cobra.Command{Use: "testcmd"}
		testCmd.GenZshCompletion(os.Stdout)

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, buf.String(), "testcmd")
	})

	t.Run("run fish completion", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		testCmd := &cobra.Command{Use: "testcmd"}
		testCmd.GenFishCompletion(os.Stdout, true)

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, buf.String(), "testcmd")
	})

	t.Run("run powershell completion", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		testCmd := &cobra.Command{Use: "testcmd"}
		testCmd.GenPowerShellCompletionWithDesc(os.Stdout)

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, strings.ToLower(buf.String()), "testcmd")
	})
}

func TestCompletionCmd_ActualRun(t *testing.T) {
	// Test the actual completionCmd.Run function
	t.Run("run bash completion through actual command", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call the actual Run function
		completionCmd.Run(completionCmd, []string{"bash"})

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		// Should contain goji command references
		assert.Contains(t, output, "goji")
	})

	t.Run("run zsh completion through actual command", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		completionCmd.Run(completionCmd, []string{"zsh"})

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, buf.String(), "goji")
	})

	t.Run("run fish completion through actual command", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		completionCmd.Run(completionCmd, []string{"fish"})

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, buf.String(), "goji")
	})

	t.Run("run powershell completion through actual command", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		completionCmd.Run(completionCmd, []string{"powershell"})

		_ = w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, strings.ToLower(buf.String()), "goji")
	})
}

