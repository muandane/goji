package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if the commit message follows the conventional commit format",
	Long:  ``,
	// Args:  cobra.MaximumNArgs(1), // Accept one optional argument, which is the path to the commit message file
	Run: func(cmd *cobra.Command, args []string) {
		var commitMessage string

		if len(os.Args) < 2 {
			fmt.Println("No commit message provided.")
			os.Exit(1)
		}

		commitMessage = strings.Split(string(os.Args[2]), "\n")[0]
		// fmt.Println(commitMessage)
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
