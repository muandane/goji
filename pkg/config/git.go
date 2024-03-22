package config

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func GitRepo() (string, error) {
	revParse := exec.Command("git", "rev-parse", "--show-toplevel")
	repoDirBytes, err := revParse.Output()
	if err != nil {
		log.Fatalf("Error finding git root directory: %v", err)
	}
	repoDir := strings.TrimRight(string(repoDirBytes), "\n")

	return repoDir, nil
}
func (c *Config) GitCommit(repoPath, message, description string) error {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// Get the working tree
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Commit the changes
	_, err = wt.Commit(fmt.Sprintf("%s\n\n%s", message, description), &git.CommitOptions{
		Author: &object.Signature{
			When: time.Now(),
		},
		// Parents:           []plumbing.Hash{},
	})

	return err
}
