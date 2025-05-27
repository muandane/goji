package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetStagedDiff returns the staged git diff.
func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged") // [cite: 331]
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	diff := strings.TrimSpace(string(output))
	if diff == "" {
		return "", fmt.Errorf("no staged changes found to generate a commit message") // [cite: 332]
	}

	return diff, nil
}
