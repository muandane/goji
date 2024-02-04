package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if the commit message follows the conventional commit format",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("sh", "-c", "git log -1 --pretty=%B").Output()
		if err != nil {
			log.Fatal(err)
		}

		commitMessage := strings.Split(string(out), "\n")[0]
		fmt.Println(commitMessage)
		// Define the regex pattern for a conventional commit message
		// Include all commit types: feat, fix, docs, style, refactor, test, chore, build, ci ...
		re := regexp.MustCompile(`^[\w\s]*?(feat|fix|docs|style|refactor|test|chore|build|ci|perf|improvement|package)(\([\w\s]*\))?[: ].+$`)
		if !re.MatchString(commitMessage) {
			fmt.Println("Error: Your commit message does not follow the conventional commit format.")
			os.Exit(1)
		} else {
			fmt.Println("Success: Your commit message follows the conventional commit format.")
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
