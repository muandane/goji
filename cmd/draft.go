// cmd/draft.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// extractScopeFromFiles determines the scope based on staged file paths
// Returns a short, meaningful scope like "ai", "cmd", or "pkg/ai" for nested paths
func extractScopeFromFiles() string {
	cmd := exec.Command("git", "diff", "--staged", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	files := strings.Fields(string(output))
	if len(files) == 0 {
		return ""
	}

	// Normalize file paths and extract directories
	dirPaths := make([]string, 0, len(files))
	dirCounts := make(map[string]int)

	for _, file := range files {
		// Clean the path
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		dir := filepath.Dir(file)
		if dir == "." || dir == "" {
			// Root level file - use filename without extension as scope
			base := filepath.Base(file)
			ext := filepath.Ext(base)
			if ext != "" {
				base = strings.TrimSuffix(base, ext)
			}
			dirCounts[base]++
			dirPaths = append(dirPaths, base)
		} else {
			// Normalize path separators
			dir = strings.ReplaceAll(dir, "\\", "/")
			dir = strings.TrimPrefix(dir, "./")
			dirCounts[dir]++
			dirPaths = append(dirPaths, dir)
		}
	}

	// Strategy 1: If all files are in the same directory, use the last component
	if len(dirCounts) == 1 {
		for dir := range dirCounts {
			parts := strings.Split(dir, "/")
			// Return the last component (e.g., "ai" from "pkg/ai")
			return parts[len(parts)-1]
		}
	}

	// Strategy 2: Find common path prefix
	if len(dirPaths) > 1 {
		commonParts := strings.Split(dirPaths[0], "/")
		for _, dir := range dirPaths[1:] {
			parts := strings.Split(dir, "/")
			// Find common prefix
			minLen := len(commonParts)
			if len(parts) < minLen {
				minLen = len(parts)
			}
			newCommonLen := 0
			for i := 0; i < minLen; i++ {
				if commonParts[i] == parts[i] {
					newCommonLen++
				} else {
					break
				}
			}
			commonParts = commonParts[:newCommonLen]
			if len(commonParts) == 0 {
				break
			}
		}

		// If we have a meaningful common prefix, use the last component
		if len(commonParts) > 0 {
			// For paths like "pkg/ai", return "ai" (last component)
			// For single-level paths, return as-is
			return commonParts[len(commonParts)-1]
		}
	}

	// Strategy 3: Use the most common directory's last component
	maxCount := 0
	mostCommon := ""
	for dir, count := range dirCounts {
		if count > maxCount {
			maxCount = count
			mostCommon = dir
		}
	}

	if mostCommon != "" {
		parts := strings.Split(mostCommon, "/")
		return parts[len(parts)-1]
	}

	return ""
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
			// Manual override takes precedence
			fullScopePart = "(" + overrideScope + ")"
		} else {
			// Always replace AI-generated scope with file path-based scope
			detectedScope := extractScopeFromFiles()
			if detectedScope != "" {
				fullScopePart = "(" + detectedScope + ")"
			} else if fullScopePart == "" {
				// Keep empty if no scope detected and none was in original
				fullScopePart = ""
			}
		}

		var builder strings.Builder
		builder.WriteString(commitType)

		hasEmoji := false
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
				hasEmoji = true
			}
		}

		if fullScopePart != "" {
			if hasEmoji {
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
	Long:  `This command connects to an AI provider (e.g., OpenRouter, Groq) to generate a commit message based on your staged changes.`,
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
			printErrorAndExit(
				"❌ Phind provider has been permanently shut down and is no longer supported. Please update your .goji.json to use 'openrouter', 'groq', or 'gemini' instead.",
			)
		case "openrouter":
			apiKey := os.Getenv("OPENROUTER_API_KEY")
			if apiKey == "" {
				printErrorAndExit("❌ OPENROUTER_API_KEY environment variable not set.")
			}
			provider = ai.NewOpenRouterProvider(apiKey, cfg.AIChoices.OpenRouter.Model)
		case "groq":
			apiKey := os.Getenv("GROQ_API_KEY")
			if apiKey == "" {
				printErrorAndExit("❌ GROQ_API_KEY environment variable not set.")
			}
			provider = ai.NewGroqProvider(apiKey, cfg.AIChoices.Groq.Model)
		case "gemini":
			// Gemini supports both OAuth (browser login) and API key
			apiKey := os.Getenv("GEMINI_API_KEY")
			provider = ai.NewGeminiProvider(apiKey, cfg.AIChoices.Gemini.Model)
		default:
			printErrorAndExit("❌ Unsupported AI provider: %s", cfg.AIProvider)
		}

		// Wrap provider with chunked processor for large diffs
		chunkedProcessor := ai.NewChunkedDiffProcessor(provider)

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
			// Generate detailed commit with body using chunked processor
			result, err := chunkedProcessor.ProcessChunkedDetailedCommit(diff, string(commitTypesJSON), extraContext)
			if err != nil {
				printErrorAndExit("❌ Error generating detailed commit message: %v", err)
			}

			if result.Message == "" {
				printErrorAndExit("❌ No commit message generated. The AI provider returned an empty response.")
			}

			finalCommitMessage = processCommitMessage(result.Message, cfg.NoEmoji, cfg.Types)
			commitBody = result.Body
		} else {
			// Generate simple commit message using chunked processor
			commitMessage, err := chunkedProcessor.ProcessChunkedDiff(diff, string(commitTypesJSON), extraContext)
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
