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
	"github.com/spf13/cobra"
)

var (
	commitDirectly bool
	overrideType   string
	overrideScope  string
)

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Generate a commit message for staged changes using AI",
	Long:  `This command connects to an AI provider (e.g., Phind) to generate a commit message based on your staged changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ViperConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("Error getting staged diff: %v\n", err)
			os.Exit(1)
		}

		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "phind":
			provider = ai.NewPhindProvider(cfg.AIChoices.Phind.Model)
		default:
			fmt.Printf("Unsupported AI provider: %s\n", cfg.AIProvider)
			os.Exit(1)
		}

		aiCommitTypes := make(map[string]string)
		for _, t := range cfg.Types {
			aiCommitTypes[t.Name] = t.Description
		}

		commitTypesJSON, err := json.Marshal(aiCommitTypes)
		if err != nil {
			fmt.Printf("Error marshaling commit types for AI: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Generating commit message using %s...\n", provider.GetModel())
		commitMessage, err := provider.GenerateCommitMessage(diff, string(commitTypesJSON))
		if err != nil {
			fmt.Printf("Error generating commit message: %v\n", err)
			os.Exit(1)
		}

		// --- Modified logic for emoji and spacing ---
		finalCommitMessage := commitMessage
		if !cfg.NoEmoji { // Check if emojis are enabled
			// Regex to parse: <type>(<optional scope>): <message>
			re := regexp.MustCompile(`^([a-zA-Z]+)(\([^)]*\))?:\s*(.*)$`)
			matches := re.FindStringSubmatch(commitMessage)

			if len(matches) > 0 {
				commitType := matches[1]
				fullScopePart := matches[2]
				messagePart := matches[3]

				// Apply overrides
				if overrideType != "" {
					commitType = overrideType
				}

				if overrideScope != "" {
					fullScopePart = "(" + overrideScope + ")"
				}

				var emoji string
				for _, t := range cfg.Types {
					if t.Name == commitType {
						emoji = t.Emoji
						break
					}
				}

				if emoji != "" {
					var builder strings.Builder
					builder.WriteString(commitType)
					builder.WriteString(" ")
					builder.WriteString(emoji)
					builder.WriteString(" ")

					// Append the full scope part if it exists
					if fullScopePart != "" {
						builder.WriteString(fullScopePart)
					}
					builder.WriteString(": ")
					builder.WriteString(strings.TrimSpace(messagePart))

					finalCommitMessage = builder.String()
				}
			}
		} else if overrideType != "" || overrideScope != "" {
			// Handle overrides when emojis are disabled
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
				if fullScopePart != "" {
					builder.WriteString(fullScopePart)
				}
				builder.WriteString(": ")
				builder.WriteString(strings.TrimSpace(messagePart))

				finalCommitMessage = builder.String()
			}
		}
		// --- End of modified logic ---

		fmt.Println("--- Generated Commit Message ---")
		fmt.Print(finalCommitMessage)
		fmt.Println("\n------------------------------")

		if commitDirectly {
			fmt.Println("Attempting to commit directly...")
			err := executeGitCommit(finalCommitMessage, "", cfg.SignOff)
			if err != nil {
				fmt.Printf("Error committing changes: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Successfully committed changes.")
		} else {
			fmt.Println("You can now manually commit with this message, or integrate it into your commit flow.")
			fmt.Println("To commit directly next time, use the `--commit` flag.")
		}
	},
}

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.Flags().BoolVarP(&commitDirectly, "commit", "c", false, "Commit the generated message directly")
	draftCmd.Flags().StringVarP(&overrideType, "type", "t", "", "Override the commit type (e.g., feat, fix, docs)")
	draftCmd.Flags().StringVarP(&overrideScope, "scope", "s", "", "Override the commit scope (e.g., api, ui, core)")
}
