// File: ./cmd/root.go
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

// rootCmd is the base command for the goji CLI.
var rootCmd = &cobra.Command{
	Use:   "goji",
	Short: "Goji CLI",
	Long:  `Goji is a CLI tool to generate conventional commits with emojis`,
	// SilenceUsage:  true, // Keep as is if desired
	// SilenceErrors: true, // Keep as is if desired
	RunE: func(cmd *cobra.Command, args []string) error {
		// Access flag values directly via the command object
		versionFlag, err := cmd.Flags().GetBool("version")
		if err != nil {
			return fmt.Errorf("failed to get version flag: %w", err)
		}
		if versionFlag {
			// Use color.Green func for simpler coloring
			color.Green("goji version: v%s", version)
			return nil
		}

		cfg, err := config.ViperConfig()
		if err != nil {
			// Error loading config is critical for most operations
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Access flags, check for errors
		typeFlag, err := cmd.Flags().GetString("type")
		if err != nil {
			return fmt.Errorf("failed to get type flag: %w", err)
		}
		messageFlag, err := cmd.Flags().GetString("message")
		if err != nil {
			return fmt.Errorf("failed to get message flag: %w", err)
		}
		scopeFlag, err := cmd.Flags().GetString("scope")
		if err != nil {
			return fmt.Errorf("failed to get scope flag: %w", err)
		}
		noVerifyFlag, err := cmd.Flags().GetBool("no-verify")
		if err != nil {
			return fmt.Errorf("failed to get no-verify flag: %w", err)
		}
		addFlag, err := cmd.Flags().GetBool("add")
		if err != nil {
			return fmt.Errorf("failed to get add flag: %w", err)
		}
		amendFlag, err := cmd.Flags().GetBool("amend")
		if err != nil {
			return fmt.Errorf("failed to get amend flag: %w", err)
		}
		gitFlags, err := cmd.Flags().GetStringArray("git-flag")
		if err != nil {
			return fmt.Errorf("failed to get git-flag: %w", err)
		}

		var commitMessage, commitBody string
		if typeFlag != "" && messageFlag != "" {
			// Use flags for non-interactive mode
			commitMessage = constructCommitMessage(cfg, typeFlag, scopeFlag, messageFlag) // Pass scopeFlag
			commitBody = ""                                                               // No body in simple non-interactive mode via flags
		} else {
			// Interactive mode
			// AskQuestions returns []string, expecting [message, body]
			// TODO: Improve AskQuestions to return a struct for better type safety
			messages, err := utils.AskQuestions(cfg, typeFlag, messageFlag) // Pass typeFlag, messageFlag for presets
			if err != nil {
				// An error from AskQuestions (like user cancelling the commit) should be handled gracefully.
				// Check if it's a specific user cancellation error or a real issue.
				// For now, just return the error.
				return fmt.Errorf("failed to get commit details from interactive prompt: %w", err)
			}
			if len(messages) != 2 {
				// Defensive check: Ensure AskQuestions returned the expected number of elements
				return fmt.Errorf("unexpected number of return values from AskQuestions: %d", len(messages))
			}
			commitMessage, commitBody = messages[0], messages[1]
		}

		if commitMessage == "" {
			return fmt.Errorf("commit message cannot be empty")
		}

		// Pass relevant flags to the git execution function
		return executeGitCommit(commitMessage, commitBody, cfg.SignOff, noVerifyFlag, addFlag, amendFlag, gitFlags)
	},
}

// init initializes the flags for the root command.
func init() {
	// Define flags using package-level variables
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type from the config file")
	rootCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Specify a custom scope")
	rootCmd.Flags().StringVarP(&messageFlag, "message", "m", "", "Specify a commit message")
	rootCmd.Flags().BoolVarP(&noVerifyFlag, "no-verify", "n", false, "bypass pre-commit and commit-msg hooks")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVarP(&addFlag, "add", "a", false, "Automatically stage files that have been modified and deleted")
	rootCmd.Flags().BoolVar(&amendFlag, "amend", false, "Change last commit")

	// gitFlags is already a slice, so StringArrayVar is correct
	rootCmd.Flags().StringArrayVar(&gitFlags, "git-flag", []string{}, "Additional Git flags (can be used multiple times)")

	// Add the 'check' and 'init' commands (assuming they are defined elsewhere and added to rootCmd)
	// rootCmd.AddCommand(checkCmd)
	// rootCmd.AddCommand(initCmd)
}

// constructCommitMessage formats the commit message string based on config and flags.
// It handles adding emoji and scope according to the configuration.
func constructCommitMessage(cfg *config.Config, typeFlag, scopeFlag, messageFlag string) string {
	typeMatch := typeFlag // Default to the raw typeFlag
	for _, t := range cfg.Types {
		if typeFlag == t.Name {
			if !cfg.NoEmoji {
				typeMatch = fmt.Sprintf("%s %s", t.Emoji, t.Name)
				break
			} else {
				typeMatch = t.Name // Use just the name if NoEmoji is true
				break
			}
		}
	}

	// Add scope if present
	commitHeader := typeMatch
	if scopeFlag != "" {
		commitHeader += fmt.Sprintf("(%s)", scopeFlag)
	}

	// Combine header and subject
	return fmt.Sprintf("%s: %s", commitHeader, messageFlag)
}

// executeGitCommit executes a git commit with the given message and body and flags.
//
// Parameters:
// - message: the commit header (type[(scope)]: subject).
// - body: the commit body (optional).
// - sign: a boolean indicating whether to add a Signed-off-by trailer (from config).
// - noVerify: a boolean indicating whether to bypass hooks (from flag).
// - add: a boolean indicating whether to auto-stage files (from flag).
// - amend: a boolean indicating whether to amend the last commit (from flag).
// - gitFlags: additional raw git flags (from flag array).
//
// Returns:
// - error: an error if the git commit execution fails.
func executeGitCommit(message, body string, signOff, noVerify, add, amend bool, extraGitFlags []string) error {
	args := []string{"commit"}

	if noVerify {
		args = append(args, "--no-verify")
	}
	if add {
		args = append(args, "-a")
	}
	if amend {
		args = append(args, "--amend")
	}

	args = append(args, "-m", message)
	if body != "" {
		args = append(args, "-m", body)
	}

	if signOff {
		args = append(args, "--signoff")
	}

	args = append(args, extraGitFlags...)

	fmt.Println("Executing command:", "git", strings.Join(args, " "))

	gitCmd := exec.Command("git", args...)
	output, err := gitCmd.CombinedOutput()

	// fmt.Printf("Git command output:\n%s\n", output)

	if err != nil {
		return fmt.Errorf("git command failed with error: %v\nCommand: git %s\nOutput: %s", err, strings.Join(args, " "), output)
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
