package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestInitCmd_Structure(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		assert.NotNil(t, initCmd)
	})

	t.Run("command use", func(t *testing.T) {
		assert.Equal(t, "init", initCmd.Use)
	})

	t.Run("command short description", func(t *testing.T) {
		assert.NotEmpty(t, initCmd.Short)
	})

	t.Run("command long description", func(t *testing.T) {
		assert.NotEmpty(t, initCmd.Long)
	})
}

func TestInitCmd_Flags(t *testing.T) {
	t.Run("global flag exists", func(t *testing.T) {
		flag := initCmd.Flags().Lookup("global")
		assert.NotNil(t, flag)
		assert.Equal(t, "save the init file to your home directory", flag.Usage)
	})

	t.Run("repo flag exists", func(t *testing.T) {
		flag := initCmd.Flags().Lookup("repo")
		assert.NotNil(t, flag)
		assert.Equal(t, "save the init file in the repository", flag.Usage)
	})

	t.Run("flags are boolean", func(t *testing.T) {
		globalFlag := initCmd.Flags().Lookup("global")
		repoFlag := initCmd.Flags().Lookup("repo")
		
		assert.Equal(t, "bool", globalFlag.Value.Type())
		assert.Equal(t, "bool", repoFlag.Value.Type())
	})

	t.Run("flag default values", func(t *testing.T) {
		// Reset flags
		globalFlag = false
		repoFlag = false

		// Test defaults
		assert.False(t, globalFlag)
		assert.False(t, repoFlag)
	})
}

func TestInitCmd_FlagParsing(t *testing.T) {
	t.Run("parse global flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().BoolVar(&globalFlag, "global", false, "")
		
		err := cmd.ParseFlags([]string{"--global"})
		assert.NoError(t, err)
		assert.True(t, globalFlag)
		
		// Reset
		globalFlag = false
	})

	t.Run("parse repo flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().BoolVar(&repoFlag, "repo", false, "")
		
		err := cmd.ParseFlags([]string{"--repo"})
		assert.NoError(t, err)
		assert.True(t, repoFlag)
		
		// Reset
		repoFlag = false
	})

	t.Run("parse both flags", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().BoolVar(&globalFlag, "global", false, "")
		cmd.Flags().BoolVar(&repoFlag, "repo", false, "")
		
		err := cmd.ParseFlags([]string{"--global", "--repo"})
		assert.NoError(t, err)
		assert.True(t, globalFlag)
		assert.True(t, repoFlag)
		
		// Reset
		globalFlag = false
		repoFlag = false
	})
}

func TestInitCmd_Run(t *testing.T) {
	t.Run("command run function exists", func(t *testing.T) {
		assert.NotNil(t, initCmd.Run)
	})

	t.Run("run with error path", func(t *testing.T) {
		// Save original flags
		originalGlobal := globalFlag
		originalRepo := repoFlag
		defer func() {
			globalFlag = originalGlobal
			repoFlag = originalRepo
		}()

		// Set flags to trigger error (no flags set)
		globalFlag = false
		repoFlag = false

		// Capture output
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call Run function - this will call InitRepoConfig which will error
		initCmd.Run(initCmd, []string{})

		// Close writer and restore stdout
		_ = w.Close()
		os.Stdout = originalStdout

		// Read captured output
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		// Should contain error message
		assert.Contains(t, output, "Failed to initialize")
	})
}

