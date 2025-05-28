package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss" // Ensure lipgloss is imported for styles
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

var (
	primaryColor = lipgloss.Color("#7C3AED")
	successColor = lipgloss.Color("#10B981")
	errorColor   = lipgloss.Color("#EF4444")
	mutedColor   = lipgloss.Color("#6B7280")
	accentColor  = lipgloss.Color("#EC4899")

	headerStyle     = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	successMsgStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	errorMsgStyle   = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	infoMsgStyle    = lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
	commitMsgStyle  = lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	mutedStyle      = lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
)

var rootCmd = &cobra.Command{
	Use:           "goji",
	Short:         "Goji CLI",
	Long:          `Goji is a CLI tool to generate conventional commits with emojis`,
	SilenceUsage:  true,
	SilenceErrors: true,
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
		scopeFlag, _ := cmd.Flags().GetString("scope")

		var commitMessage, commitBody string
		if typeFlag != "" && messageFlag != "" {
			commitMessage = constructCommitMessage(cfg, typeFlag, scopeFlag, messageFlag)
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

		// CORRECTED: Pass "--amend" as a variadic argument if amendFlag is set
		var additionalFlags []string
		if amendFlag { // This `amendFlag` is the global one from root command's flags
			additionalFlags = append(additionalFlags, "--amend")
		}
		// Now, executeGitCommit will combine these with gitFlags and other internal flags.
		return executeGitCommit(commitMessage, commitBody, cfg.SignOff, additionalFlags...)
	},
}

// init initializes the flags for the root command.
func init() {
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type from the config file")
	rootCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Specify a custom scope")
	rootCmd.Flags().StringVarP(&messageFlag, "message", "m", "", "Specify a commit message")
	rootCmd.Flags().BoolVarP(&noVerifyFlag, "no-verify", "n", false, "bypass pre-commit and commit-msg hooks")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVarP(&addFlag, "add", "a", false, "Automatically stage files that have been modified and deleted")
	rootCmd.Flags().BoolVar(&amendFlag, "amend", false, "Change last commit")

	rootCmd.Flags().StringArrayVar(&gitFlags, "git-flag", []string{}, "Additional Git flags (can be used multiple times)")
	rootCmd.AddCommand(editCmd) // Ensure editCmd is added
	// Add existing commands
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(draftCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(manCmd)
}

func constructCommitMessage(cfg *config.Config, typeFlag, scopeFlag, messageFlag string) string {
	typeMatch := typeFlag
	for _, t := range cfg.Types {
		if typeFlag == t.Name {
			if !cfg.NoEmoji {
				typeMatch = fmt.Sprintf("%s %s", t.Name, t.Emoji)
				break
			} else {
				typeMatch = t.Name
				break
			}
		}
	}

	commitHeader := typeMatch
	if scopeFlag != "" {
		commitHeader += fmt.Sprintf(" (%s)", scopeFlag)
	}

	return fmt.Sprintf("%s: %s", commitHeader, messageFlag)
}

// executeGitCommit executes a git commit with the given message and body.
// It now accepts variadic `extraGitFlags` which will be appended to the command.
func executeGitCommit(message, body string, signOff bool, extraGitFlags ...string) error {
	args := []string{"commit", "-m", message}
	if body != "" {
		args = append(args, "-m", body)
	}
	if signOff {
		args = append(args, "--signoff")
	}
	var rootCommandFlags []string
	if noVerifyFlag {
		rootCommandFlags = append(rootCommandFlags, "--no-verify")
	}
	if addFlag {
		rootCommandFlags = append(rootCommandFlags, "-a")
	}
	// Only apply the root-level amendFlag if it's set AND it's not already
	// included in the extraGitFlags (e.g., from `goji edit`).
	if amendFlag && !contains(extraGitFlags, "--amend") {
		rootCommandFlags = append(rootCommandFlags, "--amend")
	}

	// Combine all flags
	allArgs := append(args, rootCommandFlags...)
	allArgs = append(allArgs, gitFlags...)      // Flags passed via --git-flag
	allArgs = append(allArgs, extraGitFlags...) // Flags passed dynamically from other commands (e.g., "edit")

	gitCmd := exec.Command("git", allArgs...)
	// fmt.Println("Executing command:", strings.Join(append([]string{"git"}, allArgs...), " "))
	output, err := gitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %v\nOutput: %s", err, output)
	}
	fmt.Print(string(output))

	return nil
}

// Helper function to check if a slice contains a string (for executeGitCommit)
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
