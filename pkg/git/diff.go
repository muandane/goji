package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetStagedDiff() (string, error) {
	// First check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}

	// Check if there are any staged changes
	cmd = exec.Command("git", "diff", "--staged", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to check staged files: %w", err)
	}

	stagedFiles := strings.TrimSpace(string(output))
	if stagedFiles == "" {
		return "", fmt.Errorf("no staged changes found to generate a commit message")
	}

	// Get the actual diff
	cmd = exec.Command("git", "diff", "--staged")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	diff := strings.TrimSpace(string(output))
	if diff == "" {
		// This shouldn't happen if we have staged files, but handle it gracefully
		return "", fmt.Errorf("no staged changes found to generate a commit message")
	}

	return diff, nil
}
