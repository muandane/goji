package cmd

import (
	"fmt"

	"log"
	"os"
	"os/exec"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version     string
	versionFlag bool
	typeFlag    string
	scopeFlag   string
	messageFlag string
)

var rootCmd = &cobra.Command{
	Use:   "goji",
	Short: "Goji CLI",
	Long:  `Goji is is a cli tool to generate conventional commits with emojis`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			color.Set(color.FgGreen)
			fmt.Printf("goji version: v%s\n", version)
			color.Unset()
			return
		}
		color.Set(color.FgGreen)
		fmt.Printf("Goji v%s is a cli tool to generate conventional commits with emojis.\n", version)
		color.Unset()

		if typeFlag != "" && messageFlag != "" {
			config, err := config.LoadConfig(".goji.json")
			if err != nil {
				log.Fatalf(color.YellowString("Error loading config file: %v"), err)
			}
			var typeMatch string
			for _, t := range config.Types {
				if typeFlag == t.Name {
					typeMatch = fmt.Sprintf("%s %s", t.Name, t.Emoji)
					break
				}
			}

			// If no match was found, use the type flag as is
			if typeMatch == "" {
				typeMatch = typeFlag
			}

			// Construct the commit message from the flags
			commitMessage := messageFlag
			if typeMatch != "" {
				commitMessage = fmt.Sprintf("%s: %s", typeMatch, commitMessage)
				if scopeFlag != "" {
					commitMessage = fmt.Sprintf("%s(%s): %s", typeMatch, scopeFlag, messageFlag)
				}
			}

			// Create the git commit command with the commit message
			cmd := exec.Command("git", "commit", "-m", commitMessage)

			// Run the command and capture the output and any error
			output, err := cmd.CombinedOutput()

			// Check if the command resulted in an error
			if err != nil {
				// If there was an error, print it and the output from the command
				fmt.Printf(color.MagentaString("Error executing git commit: %v\n"), err)
				fmt.Println("Git commit output: ", string(output))
				return
			}

			// If there was no error, print the output from the command
			fmt.Printf("Git commit output: %s\n", string(output))
		} else {
			// If not all flags were provided, fall back to your existing interactive prompt logic
			config, err := config.LoadConfig(".goji.json")
			if err != nil {
				log.Fatalf(color.YellowString("Error loading config file: %v"), err)
			}

			commitMessage, err := utils.AskQuestions(config)
			if err != nil {
				log.Fatalf(color.YellowString("Error asking questions: %v"), err)
			}
			cmd := exec.Command("git", "commit", "-m", commitMessage)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf(color.MagentaString("Error executing git commit: %v\n"), err)
				fmt.Println("Git commit output: ", string(output))
				return
			}
			fmt.Printf("Git commit output: %s\n", string(output))
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type from the config file")
	rootCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Specify a custom scope")
	rootCmd.Flags().StringVarP(&messageFlag, "message", "m", "", "Specify a commit message")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
