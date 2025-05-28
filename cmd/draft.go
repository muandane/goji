// cmd/draft.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muandane/goji/pkg/ai"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/git"
	"github.com/muandane/goji/pkg/models"
	"github.com/spf13/cobra"
)

var (
	commitDirectly bool
	overrideType   string
	overrideScope  string
	extraContext   string
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

	infoMsgStyle   = lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
	commitMsgStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)
	mutedStyle = lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
)

func printErrorAndExit(format string, a ...interface{}) {
	fmt.Println(errorMsgStyle.Render(fmt.Sprintf(format, a...)))
	os.Exit(1)
}

func processCommitMessage(commitMessage string, noEmoji bool, configTypes []models.CommitType) string {
	finalCommitMessage := commitMessage
	re := regexp.MustCompile(`^([a-zA-Z]+)(\([^)]*\))?:\s*(.*)$`)
	matches := re.FindStringSubmatch(commitMessage)

	if len(matches) > 0 {
		commitType := matches[1]
		fullScopePart := matches[2]
		messagePart := matches[3]

		if overrideType != "" {
			commitType = overrideType
		}
		if overrideScope != "" {
			fullScopePart = "(" + overrideScope + ")"
		}

		var builder strings.Builder
		builder.WriteString(commitType)

		if !noEmoji {
			var emoji string
			for _, t := range configTypes {
				if t.Name == commitType {
					emoji = t.Emoji
					break
				}
			}
			if emoji != "" {
				builder.WriteString(" ")
				builder.WriteString(emoji)
			}
		}

		if fullScopePart != "" {

			if !strings.HasSuffix(builder.String(), " ") {
				builder.WriteString(" ")
			}
			builder.WriteString(fullScopePart)
		}
		builder.WriteString(": ")
		builder.WriteString(strings.TrimSpace(messagePart))

		finalCommitMessage = builder.String()
	}
	return finalCommitMessage
}

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Generate a commit message for staged changes using AI",
	Long:  `This command connects to an AI provider (e.g., Phind) to generate a commit message based on your staged changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("✨ AI Commit Message Generator"))

		cfg, err := config.ViperConfig()
		if err != nil {
			printErrorAndExit("❌ Error loading config: %v", err)
		}

		fmt.Println(mutedStyle.Render("📋 Analyzing staged changes..."))
		diff, err := git.GetStagedDiff()
		if err != nil {
			printErrorAndExit("❌ Error getting staged diff: %v", err)
		}

		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "phind":
			provider = ai.NewPhindProvider(cfg.AIChoices.Phind.Model)
		default:
			printErrorAndExit("❌ Unsupported AI provider: %s", cfg.AIProvider)
		}

		aiCommitTypes := make(map[string]string)
		for _, t := range cfg.Types {
			aiCommitTypes[t.Name] = t.Description
		}

		commitTypesJSON, err := json.Marshal(aiCommitTypes)
		if err != nil {
			printErrorAndExit("❌ Error marshaling commit types: %v", err)
		}

		fmt.Println(mutedStyle.Render(fmt.Sprintf("🤖 Generating commit message using %s...", provider.GetModel())))

		// Pass the extra context to the AI provider
		commitMessage, err := provider.GenerateCommitMessage(diff, string(commitTypesJSON), extraContext)

		if err != nil {
			printErrorAndExit("❌ Error generating commit message: %v", err)
		}

		finalCommitMessage := processCommitMessage(commitMessage, cfg.NoEmoji, cfg.Types)

		fmt.Println(successMsgStyle.Render("✅ Commit message generated successfully!"))

		if commitDirectly {
			fmt.Println(commitMsgStyle.Render(finalCommitMessage))

			fmt.Println(mutedStyle.Render("📤 Committing changes..."))

			err := executeGitCommit(finalCommitMessage, "", cfg.SignOff)

			if err != nil {
				printErrorAndExit("❌ Error committing changes: %v", err)
			}
			fmt.Println(successMsgStyle.Render("🎉 Successfully committed changes!"))
		} else {

			fmt.Println(infoMsgStyle.Render("Here's your generated commit message:"))
			fmt.Println(commitMsgStyle.Render(finalCommitMessage))
			fmt.Println(infoMsgStyle.Render(
				"💡 Ready to commit!\n" +
					"    • Run with --commit flag to auto-commit\n" +
					"    • Use --type and --scope to override defaults\n" +
					"    • Use --context to provide additional context to the AI", // Update usage info
			))
		}
	},
}

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.Flags().BoolVarP(&commitDirectly, "commit", "c", false, "Commit the generated message directly")
	draftCmd.Flags().StringVarP(&overrideType, "type", "t", "", "Override the commit type (e.g., feat, fix, docs)")
	draftCmd.Flags().StringVarP(&overrideScope, "scope", "s", "", "Override the commit scope (e.g., api, ui, core)")
	// New: Add the --context flag
	draftCmd.Flags().StringVarP(&extraContext, "context", "x", "", "Provide additional context for AI commit message generation")
}
