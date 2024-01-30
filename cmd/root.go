package cmd

import (
	"fmt"
	"os/exec"

	"os"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	version     string
	versionFlag bool
	signFlag    bool
	typeFlag    string
	scopeFlag   string
	messageFlag string
)

var rootCmd = &cobra.Command{
	Use:   "goji",
	Short: "Goji CLI",
	Long:  `Goji is a cli tool to generate conventional commits with emojis`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			color.Set(color.FgGreen)
			fmt.Printf("goji version: v%s\n", version)
			color.Unset()
			return
		}

		_, err := config.GitRepo()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		config, err := config.ViperConfig()
		if err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}

		var commitMessage string
		var commitBody string
		if typeFlag != "" && messageFlag != "" {
			// If all flags are provided, construct the commit message from them
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
			commitMessage = messageFlag
			if typeMatch != "" {
				commitMessage = fmt.Sprintf("%s: %s", typeMatch, commitMessage)
				if scopeFlag != "" {
					commitMessage = fmt.Sprintf("%s(%s): %s", typeMatch, scopeFlag, messageFlag)
				}
			}
		} else {
			// If not all flags are provided, fall back to the interactive prompt logic
			commitMessages, err := utils.AskQuestions(config)
			if err != nil {
				log.Fatalf("Error asking questions: %v", err)
			}
			commitMessage = commitMessages[0]
			commitBody = commitMessages[1]
		}
		if commitMessage == "" {
			log.Fatalf("Error: Commit message cannot be empty")
		}

		err = spinner.New().
			Title("Committing...").
			Action(func() {
				args := []string{"commit", "-m", commitMessage, "-m", commitBody}

				if signFlag {
					args = append(args, "--signoff")
				}

				gitCmd := exec.Command("git", args...)

				output, err := gitCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Error executing git commit: %v\n", err)
					fmt.Println("Git commit output: ", string(output))
					os.Exit(1)
				}
				fmt.Printf("Git commit output: %s\n", string(output))
			}).
			Run()

		if err != nil {
			fmt.Println("Error committing: ", err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVarP(&signFlag, "sign-off", "S", false, "add a Signed-off-by trailer")
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
