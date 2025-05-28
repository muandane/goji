// ===== ./pkg/git/diff_test.go =====
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to set up a temporary git repository
func setupGitRepo(t *testing.T) (repoDir string, cleanup func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "testgitrepo")
	require.NoError(t, err)

	cleanupFunc := func() {
		os.RemoveAll(dir)
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"}, // Add an initial commit
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		err := cmd.Run()
		require.NoError(t, err, "Failed to run git command: %v", c)
	}
	return dir, cleanupFunc
}

func TestGetStagedDiff(t *testing.T) {
	repoDir, cleanup := setupGitRepo(t)
	defer cleanup()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(repoDir)
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	t.Run("no staged changes", func(t *testing.T) {
		diff, err := GetStagedDiff()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no staged changes found")
		assert.Equal(t, "", diff)
	})

	t.Run("with staged changes to an existing file", func(t *testing.T) {
		filePath := filepath.Join(repoDir, "testfile.txt")
		err := os.WriteFile(filePath, []byte("initial content\n"), 0644)
		require.NoError(t, err)

		cmdAddInitial := exec.Command("git", "add", "testfile.txt")
		cmdAddInitial.Dir = repoDir
		err = cmdAddInitial.Run()
		require.NoError(t, err)

		cmdCommitInitial := exec.Command("git", "commit", "-m", "add testfile.txt")
		cmdCommitInitial.Dir = repoDir
		err = cmdCommitInitial.Run()
		require.NoError(t, err)

		// Now modify the file and stage it
		err = os.WriteFile(filePath, []byte("initial content\nnew line\n"), 0644)
		require.NoError(t, err)
		cmdAddAgain := exec.Command("git", "add", "testfile.txt")
		cmdAddAgain.Dir = repoDir
		err = cmdAddAgain.Run()
		require.NoError(t, err)

		diff, err := GetStagedDiff()
		assert.NoError(t, err)
		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "--- a/testfile.txt") // File exists in HEAD, so '--- a/'
		assert.Contains(t, diff, "+++ b/testfile.txt")
		assert.Contains(t, diff, "+new line")
		assert.NotContains(t, diff, "new file mode", "Should be a modification, not a new file diff")
	})

	t.Run("staged new file", func(t *testing.T) {
		newFilePath := filepath.Join(repoDir, "newfile.txt")
		err := os.WriteFile(newFilePath, []byte("hello world\n"), 0644)
		require.NoError(t, err)

		cmdAdd := exec.Command("git", "add", "newfile.txt")
		cmdAdd.Dir = repoDir
		err = cmdAdd.Run()
		require.NoError(t, err)

		diff, err := GetStagedDiff()
		assert.NoError(t, err)
		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "diff --git a/newfile.txt b/newfile.txt")
		assert.Contains(t, diff, "new file mode")
		assert.Contains(t, diff, "--- /dev/null") // New file diff against HEAD
		assert.Contains(t, diff, "+hello world")
	})
}
