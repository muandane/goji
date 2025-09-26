// cmd/draft.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

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
	generateBody   bool
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
		case "openrouter":
			apiKey := os.Getenv("OPENROUTER_API_KEY")
			if apiKey == "" {
				printErrorAndExit("❌ OPENROUTER_API_KEY environment variable not set.")
			}
			provider = ai.NewOpenRouterProvider(apiKey, cfg.AIChoices.OpenRouter.Model)
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

		// Handle both regular and detailed commit generation
		var finalCommitMessage, commitBody string
		if generateBody {
			// Generate detailed commit with body
			result, err := provider.GenerateDetailedCommit(diff, string(commitTypesJSON), extraContext)
			if err != nil {
				printErrorAndExit("❌ Error generating detailed commit message: %v", err)
			}

			if result.Message == "" {
				printErrorAndExit("❌ No commit message generated. The AI provider returned an empty response.")
			}

			finalCommitMessage = processCommitMessage(result.Message, cfg.NoEmoji, cfg.Types)
			commitBody = result.Body
		} else {
			// Generate simple commit message
			commitMessage, err := provider.GenerateCommitMessage(diff, string(commitTypesJSON), extraContext)
			if err != nil {
				printErrorAndExit("❌ Error generating commit message: %v", err)
			}

			if commitMessage == "" {
				printErrorAndExit("❌ No commit message generated. The AI provider returned an empty response.")
			}

			finalCommitMessage = processCommitMessage(commitMessage, cfg.NoEmoji, cfg.Types)
		}

		// Validate the final commit message
		if finalCommitMessage == "" {
			printErrorAndExit("❌ Failed to process commit message. The result is empty.")
		}

		fmt.Println(successMsgStyle.Render("✅ Commit message generated successfully!"))

		if commitDirectly {
			// Display the commit message with body if available
			displayMessage := finalCommitMessage
			if commitBody != "" {
				displayMessage += "\n\n" + commitBody
			}
			fmt.Println(commitMsgStyle.Render(displayMessage))

			fmt.Println(mutedStyle.Render("📤 Committing changes..."))

			err := executeGitCommit(finalCommitMessage, commitBody, cfg.SignOff)
			if err != nil {
				printErrorAndExit("❌ Error committing changes: %v", err)
			}
			fmt.Println(successMsgStyle.Render("🎉 Successfully committed changes!"))
		} else {
			fmt.Println(infoMsgStyle.Render("Here's your generated commit message:"))
			fmt.Println(commitMsgStyle.Render(finalCommitMessage))
			if commitBody != "" {
				fmt.Println(commitMsgStyle.Render(commitBody))
			}

			bodyHint := ""
			if !generateBody {
				bodyHint = "\n    • Use --body flag to generate detailed commit body"
			}
			fmt.Println(infoMsgStyle.Render(
				"💡 Ready to commit!\n" +
					"    • Run with --commit flag to auto-commit\n" +
					"    • Use --type and --scope to override defaults\n" +
					"    • Use --context to provide additional context to the AI" + bodyHint,
			))
		}
	},
}

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.Flags().BoolVarP(&commitDirectly, "commit", "c", false, "Commit the generated message directly")
	draftCmd.Flags().StringVarP(&overrideType, "type", "t", "", "Override the commit type (e.g., feat, fix, docs)")
	draftCmd.Flags().StringVarP(&overrideScope, "scope", "s", "", "Override the commit scope (e.g., api, ui, core)")
	draftCmd.Flags().StringVarP(&extraContext, "context", "x", "", "Provide additional context for AI commit message generation")
	draftCmd.Flags().BoolVarP(&generateBody, "body", "b", false, "Generate detailed commit body with bullet points explaining the changes")
}
