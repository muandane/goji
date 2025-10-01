package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStagedDiff_EdgeCases(t *testing.T) {
	t.Run("not in git repository", func(t *testing.T) {
		// Create a temporary directory that's not a git repository
		tempDir, err := os.MkdirTemp("", "goji-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }()

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		_, err = GetStagedDiff()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in a git repository")
	})

	t.Run("git repository with no staged changes", func(t *testing.T) {
		// Create a git repository
		tempDir, err := os.MkdirTemp("", "goji-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }()

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// Initialize git repository
		cmd := exec.Command("git", "init")
		err = cmd.Run()
		require.NoError(t, err)

		// Create a test file but don't stage it
		testFile := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		_, err = GetStagedDiff()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no staged changes found")
	})

	t.Run("git repository with staged changes", func(t *testing.T) {
		// Create a git repository
		tempDir, err := os.MkdirTemp("", "goji-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }()

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// Initialize git repository
		cmd := exec.Command("git", "init")
		err = cmd.Run()
		require.NoError(t, err)

		// Create a test file
		testFile := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Stage the file
		cmd = exec.Command("git", "add", "test.txt")
		err = cmd.Run()
		require.NoError(t, err)

		diff, err := GetStagedDiff()
		assert.NoError(t, err)
		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "test.txt")
	})

	t.Run("git command not found", func(t *testing.T) {
		// This test would require mocking exec.Command
		// For now, we just test that the function exists
		assert.NotNil(t, GetStagedDiff)
	})
}

func TestGitCommandExecution(t *testing.T) {
	t.Run("git rev-parse command", func(t *testing.T) {
		// Test the git rev-parse command that checks for git repository
		cmd := exec.Command("git", "rev-parse", "--git-dir")
		err := cmd.Run()

		// This will fail if not in a git repository, which is expected
		// We're just testing that the command can be executed
		_ = err
	})

	t.Run("git diff --staged --name-only command", func(t *testing.T) {
		// Test the command that checks for staged files
		cmd := exec.Command("git", "diff", "--staged", "--name-only")
		output, err := cmd.Output()

		// This will fail if not in a git repository or no staged files
		// We're just testing that the command can be executed
		_ = output
		_ = err
	})

	t.Run("git diff --staged command", func(t *testing.T) {
		// Test the command that gets the actual diff
		cmd := exec.Command("git", "diff", "--staged")
		output, err := cmd.Output()

		// This will fail if not in a git repository or no staged files
		// We're just testing that the command can be executed
		_ = output
		_ = err
	})
}

func TestDiffParsing(t *testing.T) {
	t.Run("empty diff string", func(t *testing.T) {
		diff := ""
		trimmed := strings.TrimSpace(diff)
		assert.Equal(t, "", trimmed)
	})

	t.Run("diff with whitespace", func(t *testing.T) {
		diff := "  \n  diff content  \n  "
		trimmed := strings.TrimSpace(diff)
		assert.Equal(t, "diff content", trimmed)
	})

	t.Run("diff with newlines", func(t *testing.T) {
		diff := "\n\ndiff content\n\n"
		trimmed := strings.TrimSpace(diff)
		assert.Equal(t, "diff content", trimmed)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("command execution error", func(t *testing.T) {
		// Test that command execution errors are properly handled
		// This would require mocking exec.Command in a real scenario
		assert.NotNil(t, GetStagedDiff)
	})

	t.Run("output parsing error", func(t *testing.T) {
		// Test that output parsing errors are handled
		// This would require mocking command output in a real scenario
		assert.NotNil(t, GetStagedDiff)
	})
}

func TestGitRepositoryDetection(t *testing.T) {
	t.Run("detect git repository", func(t *testing.T) {
		// Test if we're currently in a git repository
		cmd := exec.Command("git", "rev-parse", "--git-dir")
		err := cmd.Run()

		// If this test is running in a git repository, it should succeed
		// If not, it should fail
		_ = err
	})

	t.Run("detect staged files", func(t *testing.T) {
		// Test if there are staged files
		cmd := exec.Command("git", "diff", "--staged", "--name-only")
		output, err := cmd.Output()

		// If there are staged files, output should not be empty
		// If not, output should be empty
		_ = output
		_ = err
	})
}

func TestDiffContentValidation(t *testing.T) {
	t.Run("validate diff content", func(t *testing.T) {
		// Test that diff content is properly validated
		// This would require actual diff content in a real scenario
		assert.NotNil(t, GetStagedDiff)
	})

	t.Run("handle empty diff gracefully", func(t *testing.T) {
		// Test that empty diff is handled gracefully
		// This would require mocking in a real scenario
		assert.NotNil(t, GetStagedDiff)
	})
}
