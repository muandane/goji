package cmd

import (
	"fmt"
	"os"
	"regexp"

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
		// if len(args) == 1 {
		// 	// If an argument is provided, treat it as the path to the commit message file
		// 	content, err := os.ReadFile(args[0])
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	commitMessage = strings.TrimSpace(string(content))
		// } else {
		// 	// If no argument is provided, fallback to getting the last commit message
		// 	out, err := exec.Command("sh", "-c", "git log -1 --pretty=%B").Output()
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	commitMessage = strings.Split(string(out), "\n")[0]
		// }
		if len(os.Args) < 2 {
			fmt.Println("No commit message provided.")
			os.Exit(1)
		}

		commitMessage = os.Args[2]
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
