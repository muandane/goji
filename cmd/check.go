package cmd

import (
	"fmt"
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
		var commitMessage string
		fromFile, _ := cmd.Flags().GetBool("from-file")

		if fromFile {
			// Read commit message from file
			if len(os.Args) < 2 {
				fmt.Println("Please provide the path to the commit message file.")
				os.Exit(1)
			}
			commitMessage = strings.Split(string(os.Args[3]), "\n")[0]
			// commitMessage = strings.TrimSpace(string(content))
		} else {
			// Get the last commit message
			gitCmd := exec.Command("git", "log", "-1", "--pretty=%B")
			output, err := gitCmd.Output()
			if err != nil {
				fmt.Printf("Error getting last commit message: %v\n", err)
				os.Exit(1)
			}
			commitMessage = strings.TrimSpace(string(output))
		}

		// Define the regex pattern for a conventional commit message
		re := regexp.MustCompile(`^[\w\s]*?(feat|fix|docs|style|refactor|test|chore|build|ci|perf|improvement|package)(\([\w\s]*\))?[:  ].+$`)
		if !re.MatchString(commitMessage) {
			fmt.Printf("Error: Your commit message does not follow the conventional commit format. %s", commitMessage)
			os.Exit(1)
		} else {
			fmt.Printf("Success: Your commit message follows the conventional commit format. %s", commitMessage)
		}
	},
}

func init() {
	checkCmd.Flags().BoolP("from-file", "f", false, "Check the commit message from a file")
	rootCmd.AddCommand(checkCmd)
}
