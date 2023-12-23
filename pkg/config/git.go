package config

import (
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
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
