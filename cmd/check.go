package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/muandane/goji/pkg/config"
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
		config, err := config.ViperConfig()
		if err != nil {
			log.Fatalf("Error loading config file.")
		}
		if fromFile {
			// Read commit message from file
			if len(os.Args) < 2 {
				fmt.Println("Please provide the path to the commit message file.")
				os.Exit(1)
			}
			commitMessage = strings.Split(string(os.Args[3]), "\n")[0]
			fmt.Println(commitMessage)
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

		emojisToIgnore := make(map[string]string)
		for _, t := range config.Types {
			emojisToIgnore[t.Emoji] = ""
		}

		for emoji, replacement := range emojisToIgnore {
			commitMessage = strings.ReplaceAll(commitMessage, emoji, replacement)
		}

		// Define the regex pattern for a conventional commit message
		parts := strings.SplitN(commitMessage, ":", 2)
		if len(parts) != 2 {
			fmt.Println("Error: Commit message does not follow the conventional commit format.")
			return
		}
		var typeNames []string
		for _, t := range config.Types {
			typeNames = append(typeNames, t.Name)
		}
		typePattern := strings.Join(typeNames, "|")
		// Validate the type and scope
		typeScope := strings.Split(strings.TrimSpace(parts[0]), "(")
		if len(typeScope) > 2 {
			fmt.Println("Error: Commit message does not follow the conventional commit format.")
			return
		}

		// Validate the type
		typeRegex := regexp.MustCompile(`\A[\w\s]*?(` + typePattern + `)\z`)
		if !typeRegex.MatchString(typeScope[0]) {
			fmt.Println("Error: Commit message type is invalid.")
			return
		}

		// Validate the scope (optional)
		if len(typeScope) == 2 {
			scope := strings.TrimSuffix(typeScope[1], ")")
			if scope == "" {
				fmt.Println("Error: Commit message scope is empty.")
				return
			}
		}
		description := strings.TrimSpace(parts[1])
		if description == "" {
			fmt.Println("Error: Commit message description is empty.")
			return
		}
		fmt.Printf("Success: Your commit message follows the conventional commit format: \n%s", commitMessage)
	},
}

func init() {
	checkCmd.Flags().BoolP("from-file", "f", false, "Check the commit message from a file")
	rootCmd.AddCommand(checkCmd)
}
