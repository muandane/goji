package config

import (
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func GitRepo() (string, error) {
	revParse := exec.Command("git", "rev-parse", "--show-toplevel")
	repoDirBytes, err := revParse.Output()
	if err != nil {
		log.Fatal().Msg("Error finding git root directory")
	}
	repoDir := strings.TrimRight(string(repoDirBytes), "\n")

	return repoDir, nil
}

// func (c *Config) GitCommit(repoPath, message, description string) error {
// 	// Open the repository
// 	repo, err := git.PlainOpen(repoPath)
// 	if err != nil {
// 		return err
// 	}

// 	// Get the working tree
// 	wt, err := repo.Worktree()
// 	if err != nil {
// 		return err
// 	}

// 	// Commit the changes
// 	_, err = wt.Commit(fmt.Sprintf("%s\n\n%s", message, description), &git.CommitOptions{
// 		Parents: []plumbing.Hash{},
// 	})

// 	return err
// }
