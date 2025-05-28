package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a previous commit message",
	Long: `This command allows you to select a recent commit and edit its message.
Note: Currently, goji primarily supports amending the *last* commit directly.
For older commits, you will be advised to use 'git rebase -i'.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("📝 Edit Commit Message"))

		cfg, err := config.ViperConfig()
		if err != nil {
			printErrorAndExit("❌ Error loading config: %v", err)
		}

		fmt.Println(mutedStyle.Render("Fetching recent commits..."))
		commits, err := getRecentCommits()
		if err != nil {
			printErrorAndExit("❌ Error fetching commits: %v", err)
		}

		if len(commits) == 0 {
			fmt.Println(infoMsgStyle.Render("No commits found in this repository."))
			return
		}

		commitOptions := make([]huh.Option[string], len(commits))
		for i, commit := range commits {
			commitOptions[i] = huh.NewOption(
				fmt.Sprintf("%s %s", commit.SHA[:7], commit.Subject),
				commit.SHA,
			)
		}

		var selectedSHA string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a commit to edit:").
					Options(commitOptions...).
					Height(10).
					Value(&selectedSHA).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("no commit selected")
						}
						return nil
					}),
			),
		)

		if err := form.Run(); err != nil {
			fmt.Println(mutedStyle.Render("Commit editing cancelled."))
			os.Exit(0)
		}

		lastCommitSHA, err := getLastCommitSHA()
		if err != nil {
			printErrorAndExit("❌ Error getting last commit SHA: %v", err)
		}

		if selectedSHA != lastCommitSHA {
			fmt.Println(errorMsgStyle.Render("⚠️ Warning: Goji currently only supports amending the *last* commit directly."))
			fmt.Println(infoMsgStyle.Render(fmt.Sprintf("To edit commit %s, you would typically use: git rebase -i %s^", selectedSHA[:7], selectedSHA[:7])))
			return
		}

		currentCommitMessage, _, err := getCurrentCommitDetails(selectedSHA) // Removed unused variable
		if err != nil {
			printErrorAndExit("❌ Error fetching current commit details: %v", err)
		}

		fmt.Println(mutedStyle.Render("Please provide the new commit message and body."))

		newMessages, err := utils.AskQuestions(cfg, "", currentCommitMessage)
		if err != nil {
			printErrorAndExit("❌ Failed to get new commit details: %v", err)
		}

		newCommitMessage := newMessages[0]
		newCommitBody := ""
		if len(newMessages) > 1 {
			newCommitBody = newMessages[1]
		}

		fmt.Println(infoMsgStyle.Render(fmt.Sprintf("New commit message: %s", newCommitMessage)))
		if newCommitBody != "" {
			fmt.Println(infoMsgStyle.Render(fmt.Sprintf("New commit body: %s", newCommitBody)))
		}

		fmt.Println(mutedStyle.Render("Amending commit..."))
		// CORRECTED: Pass "--amend" as a separate string argument to the variadic parameter
		err = executeGitCommit(newCommitMessage, newCommitBody, cfg.SignOff, "--amend")
		if err != nil {
			printErrorAndExit("❌ Error amending commit: %v", err)
		}

		fmt.Println(successMsgStyle.Render("🎉 Commit message amended successfully!"))
	},
}

type GitCommit struct {
	SHA     string
	Subject string
	Body    string
}

func getRecentCommits() ([]GitCommit, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H%n%s%n%b%n---COMMIT-END---", "-n", "20")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	rawCommits := strings.Split(strings.TrimSpace(string(output)), "---COMMIT-END---")
	var commits []GitCommit
	for _, rawCommit := range rawCommits {
		if strings.TrimSpace(rawCommit) == "" {
			continue
		}
		parts := strings.SplitN(strings.TrimSpace(rawCommit), "\n", 3)
		if len(parts) < 2 {
			continue
		}

		commit := GitCommit{
			SHA:     parts[0],
			Subject: parts[1],
		}
		if len(parts) == 3 {
			commit.Body = strings.TrimSpace(parts[2])
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func getLastCommitSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit SHA: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func getCurrentCommitDetails(sha string) (subject, body string, err error) {
	cmdSubject := exec.Command("git", "log", "-1", "--pretty=format:%s", sha)
	outputSubject, err := cmdSubject.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get commit subject for %s: %w", sha, err)
	}
	subject = strings.TrimSpace(string(outputSubject))

	cmdBody := exec.Command("git", "log", "-1", "--pretty=format:%b", sha)
	outputBody, err := cmdBody.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get commit body for %s: %w", sha, err)
	}
	body = strings.TrimSpace(string(outputBody))

	return subject, body, nil
}
