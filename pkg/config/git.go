package config

import (
	"os/exec"

	"github.com/charmbracelet/log"
)

func GitRepo() error {
	revParse := exec.Command("git", "rev-parse", "--show-toplevel")
	_, err := revParse.Output()
	if err != nil {
		log.Fatalf("Error finding git root directory: %v", err)
	}
	return nil
}

// func GitStatus() {
// 	cmd := exec.Command("git", "status", "--porcelain")
// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Printf("Failed to run git status: %v\n", err)
// 		return
// 	}

// 	status := strings.TrimSpace(out.String())
// 	if len(status) != 0 {
// 		log.Fatalf("You have uncommitted changes in your git repository. Please commit or stash them before running goji.")
// 	}
// }
