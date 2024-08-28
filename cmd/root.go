package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	gitFlags                                      []string
	version, typeFlag, messageFlag, scopeFlag     string
	versionFlag, noVerifyFlag, amendFlag, addFlag bool
)

var rootCmd = &cobra.Command{
	Use:          "goji",
	Short:        "Goji CLI",
	Long:         `Goji is a CLI tool to generate conventional commits with emojis`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag {
			color.Green("goji version: v%s", version)
			return nil
		}

		cfg, err := config.ViperConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		typeFlag, _ := cmd.Flags().GetString("type")
		messageFlag, _ := cmd.Flags().GetString("message")

		var commitMessage, commitBody string
		if typeFlag != "" && messageFlag != "" {
			commitMessage = constructCommitMessage(cfg, typeFlag, "", messageFlag)
		} else {
			messages, err := utils.AskQuestions(cfg, typeFlag, messageFlag)
			if err != nil {
				return fmt.Errorf("failed to get commit details: %w", err)
			}
			commitMessage, commitBody = messages[0], messages[1]
		}

		if commitMessage == "" {
			return fmt.Errorf("commit message cannot be empty")
		}

		return executeGitCommit(commitMessage, commitBody, cfg.SignOff)
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
	rootCmd.Flags().BoolVarP(&addFlag, "add", "a", false, "Automatically stage files that have been modified and deleted")
	rootCmd.Flags().BoolVar(&amendFlag, "amend", false, "Change last commit")

	rootCmd.Flags().StringArrayVar(&gitFlags, "git-flag", []string{}, "Git flags (can be used multiple times)")
}

func constructCommitMessage(cfg *config.Config, typeFlag, scopeFlag, messageFlag string) string {
	typeMatch := typeFlag
	for _, t := range cfg.Types {
		if typeFlag == t.Name {
			if !cfg.NoEmoji {
				typeMatch = fmt.Sprintf("%s %s", t.Name, t.Emoji)
			}
			break
		}
	}

	if scopeFlag != "" {
		return fmt.Sprintf("%s(%s): %s", typeMatch, scopeFlag, messageFlag)
	}
	return fmt.Sprintf("%s: %s", typeMatch, messageFlag)
}

// commit executes a git commit with the given message and body.
//
// Parameters:
// - message: the commit message.
// - body: the commit body.
// - sign: a boolean indicating whether to add a Signed-off-by trailer.
// - amend: a boolean indicating whether to amend the last commit.
// - commits the changes to git
//
// Returns:
// - error: an error if the git commit execution fails.
func executeGitCommit(message, body string, signOff bool) error {
	args := []string{"commit", "-m", message}
	if body != "" {
		args = append(args, "-m", body)
	}
	if signOff {
		args = append(args, "--signoff")
	}
	var extraArgs []string
	if noVerifyFlag {
		extraArgs = append(extraArgs, "--no-verify")
	}
	if addFlag {
		extraArgs = append(extraArgs, "-a")
	}
	if amendFlag {
		extraArgs = append(extraArgs, "--amend")
	}
	// Add all dynamically passed Git flags
	baseArgs := append(args, extraArgs...)
	args = append(baseArgs, gitFlags...)

	gitCmd := exec.Command("git", args...)
	output, err := gitCmd.CombinedOutput()
	// Print the executed command
	fmt.Println("Executing command:", strings.Join(append([]string{"git"}, args...), " "))
	// Print the command output
	if err != nil {
		return fmt.Errorf("git command failed: %v\nOutput: %s", err, output)
	}
	fmt.Printf("Git command output:\n%s", output)

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
