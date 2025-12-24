package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss" // Ensure lipgloss is imported for styles
	"github.com/fatih/color"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"
	"github.com/spf13/cobra"

	"github.com/carapace-sh/carapace"
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
			v := getVersion()
			color.Green("goji version: v%s", v)
			return nil
		}

		// Check if we're in a git repository before proceeding
		if !isInGitRepo() {
			return showNotInGitRepoMessage(cmd)
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

// getVersion returns the version string, falling back to git tag detection if version is empty
func getVersion() string {
	if version != "" {
		return version
	}
	
	// Try to get version from git describe
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err == nil {
		v := strings.TrimSpace(string(output))
		// Remove 'v' prefix if present
		v = strings.TrimPrefix(v, "v")
		// Clean up any newlines
		v = strings.TrimRight(v, "\n\r")
		// If the version contains a dash (e.g., "0.1.8-2-g035f84e-dirty"), 
		// extract just the version part before the first dash for cleaner output
		if idx := strings.Index(v, "-"); idx > 0 {
			// Keep the full string if it's just a commit hash, otherwise use prefix
			if strings.HasPrefix(v, "0.") || strings.HasPrefix(v, "1.") || strings.HasPrefix(v, "2.") {
				v = v[:idx]
			}
		}
		if v != "" {
			return v
		}
	}
	
	// Fallback to "dev" if git is not available or not in a git repo
	return "dev"
}

// isInGitRepo checks if the current directory is in a git repository
func isInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// showNotInGitRepoMessage displays a helpful message when not in a git repository
func showNotInGitRepoMessage(cmd *cobra.Command) error {
	color.Set(color.FgYellow)
	fmt.Println("\n⚠️  You're not in a git repository.")
	color.Unset()
	
	fmt.Println("\nGoji is a CLI tool for generating conventional commit messages with emojis.")
	fmt.Println("To use goji, you need to be in a git repository.")
	
	fmt.Println("\n💡 Quick tips:")
	fmt.Println("   1. Initialize a git repository:  git init")
	fmt.Println("   2. Initialize goji config:       goji init --global")
	fmt.Println("   3. Or initialize repo config:     goji init --repo (requires git repo)")
	
	fmt.Println("\n📚 For more information:")
	fmt.Println("   • View help:                      goji --help")
	fmt.Println("   • Initialize config:               goji init --help")
	fmt.Println("   • Read the README:                 https://github.com/muandane/goji")
	
	fmt.Println()
	
	// Show help output
	_ = cmd.Help()
	
	// Return nil (no error) so we exit with code 0
	return nil
}

func Execute() {
	carapace.Gen(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
