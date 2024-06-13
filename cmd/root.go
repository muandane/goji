package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"os"

	"github.com/alessio/shellescape"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	version      string
	versionFlag  bool
	noVerifyFlag bool
	typeFlag     string
	scopeFlag    string
	messageFlag  string
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
			var extraArgs []string
			if noVerifyFlag {
				extraArgs = append(extraArgs, "--no-verify")
			}
			command, commandString := buildCommitCommand(commitMessage, commitBody, signOff, extraArgs)
			textStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Background(lipgloss.Color("#000000")).
				Bold(true)
			prettyContent := textStyle.Render(fmt.Sprintf("Executing command: %s", commandString))
			fmt.Println(prettyContent)
			if err := commit(command); err != nil {
				log.Fatalf("Error committing changes: %q", err)
			}
		}

		err = spinner.New().
			Title("Committing...").
			Action(action).
			Run()
		if gitCommitError != nil {
			fmt.Println("Error committing changes:", gitCommitError)
			fmt.Println("Check the output above for details.")
		} else if err != nil {
			fmt.Println("Error committing:", err)
		}
	},
}

// init initializes the flags for the root command.
//
// This function sets up the flags for the root command, which are used to specify the type, scope, message,
// and options for the command. The flags are defined using the `rootCmd.Flags()` method.
//
// - `typeFlag` is a string flag that specifies the type from the config file.
// - `scopeFlag` is a string flag that specifies a custom scope.
// - `messageFlag` is a string flag that specifies a commit message.
// - `noVerifyFlag` is a boolean flag that bypasses pre-commit and commit-msg hooks.
// - `versionFlag` is a boolean flag that displays version information.
//
// There are no parameters or return values for this function.
func init() {
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type from the config file")
	rootCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Specify a custom scope")
	rootCmd.Flags().StringVarP(&messageFlag, "message", "m", "", "Specify a commit message")
	rootCmd.Flags().BoolVarP(&noVerifyFlag, "no-verify", "n", false, "bypass pre-commit and commit-msg hooks")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
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
func buildCommitCommand(message string, body string, sign bool, extraArgs []string) ([]string, string) {
	if sign {
		extraArgs = append(extraArgs, "--signoff")
	}
	args := append([]string{"commit", "-m", message}, extraArgs...)
	if body != "" {
		args = append(args, "-m", body)
	}
	return args, fmt.Sprintf("git %v", shellescape.QuoteCommand(args))
}

// commit commits the changes to git
func commit(command []string) error {
	cmd := exec.Command("git", command...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// Capture Stdout
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	if err != nil {
		return err
	}

	// Define a style for the text
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#000000")).
		Bold(true)

	// Apply the style to the captured Stdout
	formattedOutput := textStyle.Render(stdout.String())

	// Print the styled output
	fmt.Print(formattedOutput)

	return nil
}
