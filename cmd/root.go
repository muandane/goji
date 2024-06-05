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
			log.Fatalf("Error: %s", err.Error())
		}

		config, err := config.ViperConfig()
		if err != nil {
			log.Fatalf("Error loading config file: %s", err.Error())
		}

		var commitMessage string
		var commitBody string
		if typeFlag != "" && messageFlag != "" {
			// If all flags are provided, construct the commit message from them
			typeMatch := ""
			for _, t := range config.Types {
				if typeFlag == t.Name {
					if !config.NoEmoji {
						typeMatch = fmt.Sprintf("%s %s", t.Name, t.Emoji)
					} else {
						typeMatch = t.Name
					}
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

			// fmt.Printf("Enter commit message:%v", signOff)
			commitMessages, err := utils.AskQuestions(config)
			if err != nil {
				log.Fatalf("Error asking questions: %s", err.Error())
			}
			commitMessage = commitMessages[0]
			commitBody = commitMessages[1]
		}
		if commitMessage == "" {
			log.Fatalf("Error: Commit message cannot be empty")
		}
		var gitCommitError error
		action := func() {
			signOff := config.SignOff
			gitCommitError = commit(commitMessage, commitBody, signOff)
		}

		err = spinner.New().
			Title("Committing...").
			Action(action).
			Run()
		if gitCommitError != nil {
			fmt.Printf("\nError committing changes: %v\n", gitCommitError)
			fmt.Println("Check the output above for details.")
		} else if err != nil {
			fmt.Printf("Error committing: %s", err)
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

// commit executes a git commit with the given message and body.
//
// Parameters:
// - message: the commit message.
// - body: the commit body.
// - sign: a boolean indicating whether to add a Signed-off-by trailer.
//
// Returns:
// - error: an error if the git commit execution fails.

func commit(message, body string, sign bool) error {
	args := []string{"commit", "-m", message}
	if body != "" {
		args = append(args, "-m", body)
	}
	if sign {
		args = append(args, "--signoff")
	}
	gitCmd := exec.Command("git", args...)

	output, err := gitCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Git commit output:\n%s\n", string(output))
		return fmt.Errorf("error executing git commit: %v", err)
	}
	fmt.Print(string(output))
	return nil
}
