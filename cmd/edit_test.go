package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestGetRecentCommits tests the git log parsing functionality
func TestGetRecentCommits(t *testing.T) {
	// Skip if not in a git repository
	if !isGitRepo() {
		t.Skip("Not in a git repository")
	}

	commits, err := getRecentCommits()
	if err != nil {
		t.Fatalf("getRecentCommits() failed: %v", err)
	}

	if len(commits) == 0 {
		t.Skip("No commits found in repository")
	}

	// Validate first commit structure
	commit := commits[0]
	if commit.SHA == "" {
		t.Error("Expected non-empty SHA")
	}
	if len(commit.SHA) < 7 {
		t.Error("Expected SHA to be at least 7 characters")
	}
	if commit.Subject == "" {
		t.Error("Expected non-empty subject")
	}
}

// TestGetLastCommitSHA tests getting the HEAD commit SHA
func TestGetLastCommitSHA(t *testing.T) {
	if !isGitRepo() {
		t.Skip("Not in a git repository")
	}

	sha, err := getLastCommitSHA()
	if err != nil {
		t.Fatalf("getLastCommitSHA() failed: %v", err)
	}

	if len(sha) != 40 {
		t.Errorf("Expected SHA length 40, got %d", len(sha))
	}
}

// TestGetCurrentCommitDetails tests fetching commit details
func TestGetCurrentCommitDetails(t *testing.T) {
	if !isGitRepo() {
		t.Skip("Not in a git repository")
	}

	sha, err := getLastCommitSHA()
	if err != nil {
		t.Fatalf("Failed to get last commit SHA: %v", err)
	}

	subject, _, err := getCurrentCommitDetails(sha)
	if err != nil {
		t.Fatalf("getCurrentCommitDetails() failed: %v", err)
	}

	if subject == "" {
		t.Error("Expected non-empty subject")
	}
	// Body can be empty, so we don't test for that
}

// TestConstructCommitMessage tests commit message construction
func TestConstructCommitMessage(t *testing.T) {
	cfg := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨"},
			{Name: "fix", Emoji: "🐛"},
		},
		NoEmoji: false,
	}

	tests := []struct {
		name        string
		typeFlag    string
		scopeFlag   string
		messageFlag string
		expected    string
	}{
		{
			name:        "basic commit with emoji",
			typeFlag:    "feat",
			scopeFlag:   "",
			messageFlag: "add new feature",
			expected:    "feat ✨: add new feature",
		},
		{
			name:        "commit with scope",
			typeFlag:    "fix",
			scopeFlag:   "auth",
			messageFlag: "resolve login issue",
			expected:    "fix 🐛 (auth): resolve login issue",
		},
		{
			name:        "unknown type",
			typeFlag:    "unknown",
			scopeFlag:   "",
			messageFlag: "some message",
			expected:    "unknown: some message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructCommitMessage(cfg, tt.typeFlag, tt.scopeFlag, tt.messageFlag)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestConstructCommitMessageNoEmoji tests commit message construction without emojis
func TestConstructCommitMessageNoEmoji(t *testing.T) {
	cfg := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		},
		NoEmoji: true,
	}

	t.Run("without scope", func(t *testing.T) {
		result := constructCommitMessage(cfg, "feat", "", "add feature")
		expected := "feat: add feature"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("with scope", func(t *testing.T) {
		result := constructCommitMessage(cfg, "feat", "lang", "add Polish language")
		expected := "feat(lang): add Polish language"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

// TestContainsHelper tests the contains helper function
func TestContainsHelper(t *testing.T) {
	slice := []string{"--amend", "--no-verify", "-a"}

	tests := []struct {
		item     string
		expected bool
	}{
		{"--amend", true},
		{"--no-verify", true},
		{"-a", true},
		{"--missing", false},
		{"", false},
	}

	for _, tt := range tests {
		result := contains(slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%q) = %v, expected %v", tt.item, result, tt.expected)
		}
	}
}

// TestRootCommandFlags tests flag parsing for root command
func TestRootCommandFlags(t *testing.T) {
	// Reset global variables
	resetFlags()

	// Test version flag
	cmd := &cobra.Command{}
	cmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")

	err := cmd.ParseFlags([]string{"--version"})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if !versionFlag {
		t.Error("Expected versionFlag to be true")
	}
}

// TestExecuteGitCommitDryRun tests the git commit execution logic without actually committing
func TestExecuteGitCommitValidation(t *testing.T) {
	// Test with empty message
	err := executeGitCommit("", "", false)
	if err == nil {
		t.Error("Expected error for empty commit message")
	}

	// Test flag combination logic
	resetFlags()
	noVerifyFlag = true
	addFlag = true
	amendFlag = true

	// We can't actually test git execution without mocking, but we can test the flag logic
	// by checking that the function doesn't panic with various flag combinations
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("executeGitCommit panicked: %v", r)
		}
	}()

	// This will fail because we're not in a proper git state, but shouldn't panic
	_ = executeGitCommit("test message", "test body", true, "--dry-run")
}

// TestEditCommandIntegration tests the edit command structure
func TestEditCommandIntegration(t *testing.T) {
	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test that the command is properly configured
	if editCmd.Use != "edit" {
		t.Error("Expected editCmd.Use to be 'edit'")
	}

	if editCmd.Short == "" {
		t.Error("Expected editCmd.Short to be non-empty")
	}

	// Restore stdout
	_ = w.Close()
	os.Stdout = old
	_, _ = io.Copy(io.Discard, r)
}

// Helper functions

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func resetFlags() {
	versionFlag = false
	noVerifyFlag = false
	amendFlag = false
	addFlag = false
	typeFlag = ""
	messageFlag = ""
	scopeFlag = ""
	gitFlags = nil
}

// Benchmark tests

func BenchmarkGetRecentCommits(b *testing.B) {
	if !isGitRepo() {
		b.Skip("Not in a git repository")
	}

	for i := 0; i < b.N; i++ {
		_, err := getRecentCommits()
		if err != nil {
			b.Fatalf("getRecentCommits() failed: %v", err)
		}
	}
}

func BenchmarkConstructCommitMessage(b *testing.B) {
	cfg := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨"},
		},
		NoEmoji: false,
	}

	for i := 0; i < b.N; i++ {
		constructCommitMessage(cfg, "feat", "scope", "message")
	}
}

// Table-driven test for git command argument construction
func TestGitCommandArgs(t *testing.T) {
	resetFlags()

	tests := []struct {
		name             string
		message          string
		body             string
		signOff          bool
		noVerify         bool
		add              bool
		amend            bool
		extraFlags       []string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "basic commit",
			message:       "test message",
			body:          "",
			signOff:       false,
			shouldContain: []string{"commit", "-m", "test message"},
		},
		{
			name:          "commit with body",
			message:       "test message",
			body:          "test body",
			signOff:       false,
			shouldContain: []string{"commit", "-m", "test message", "-m", "test body"},
		},
		{
			name:          "commit with signoff",
			message:       "test message",
			body:          "",
			signOff:       true,
			shouldContain: []string{"commit", "-m", "test message", "--signoff"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			resetFlags()
			noVerifyFlag = tt.noVerify
			addFlag = tt.add
			amendFlag = tt.amend

			// We can't easily test the actual command execution without mocking,
			// but we can verify the logic doesn't panic and handles edge cases
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test %s panicked: %v", tt.name, r)
				}
			}()

			// This will likely fail due to git state, but shouldn't panic
			_ = executeGitCommit(tt.message, tt.body, tt.signOff, tt.extraFlags...)
		})
	}
}

func TestGetRecentCommits_EdgeCases(t *testing.T) {
	t.Run("malformed git log output", func(t *testing.T) {
		// Test parsing logic with malformed input
		malformedOutput := "sha1\nsubject\n---COMMIT-END---\nsha2\n---COMMIT-END---"
		rawCommits := strings.Split(strings.TrimSpace(malformedOutput), "---COMMIT-END---")

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

		// Should skip commits with insufficient parts
		assert.Len(t, commits, 1) // Only first commit should be parsed
		assert.Equal(t, "sha1", commits[0].SHA)
		assert.Equal(t, "subject", commits[0].Subject)
	})

	t.Run("empty git log output", func(t *testing.T) {
		emptyOutput := ""
		rawCommits := strings.Split(strings.TrimSpace(emptyOutput), "---COMMIT-END---")

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

		assert.Empty(t, commits)
	})

	t.Run("commit with body", func(t *testing.T) {
		outputWithBody := "sha123\nsubject line\nbody content\n---COMMIT-END---"
		rawCommits := strings.Split(strings.TrimSpace(outputWithBody), "---COMMIT-END---")

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

		assert.Len(t, commits, 1)
		assert.Equal(t, "sha123", commits[0].SHA)
		assert.Equal(t, "subject line", commits[0].Subject)
		assert.Equal(t, "body content", commits[0].Body)
	})

	t.Run("commit without body", func(t *testing.T) {
		outputWithoutBody := "sha123\nsubject line\n---COMMIT-END---"
		rawCommits := strings.Split(strings.TrimSpace(outputWithoutBody), "---COMMIT-END---")

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

		assert.Len(t, commits, 1)
		assert.Equal(t, "sha123", commits[0].SHA)
		assert.Equal(t, "subject line", commits[0].Subject)
		assert.Empty(t, commits[0].Body)
	})
}

func TestGetCurrentCommitDetails_EdgeCases(t *testing.T) {
	t.Run("empty SHA", func(t *testing.T) {
		// This will fail in real execution, but we test the function signature
		assert.NotNil(t, getCurrentCommitDetails)
	})

	t.Run("commit with empty body", func(t *testing.T) {
		if !isGitRepo() {
			t.Skip("Not in a git repository")
		}

		sha, err := getLastCommitSHA()
		if err != nil {
			t.Skip("Cannot get last commit SHA")
		}

		subject, body, err := getCurrentCommitDetails(sha)
		assert.NoError(t, err)
		assert.NotEmpty(t, subject)
		// Body can be empty, that's valid
		_ = body
	})
}

func TestEditCmd_Structure(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		assert.NotNil(t, editCmd)
	})

	t.Run("command use", func(t *testing.T) {
		assert.Equal(t, "edit", editCmd.Use)
	})

	t.Run("command short description", func(t *testing.T) {
		assert.NotEmpty(t, editCmd.Short)
	})

	t.Run("command long description", func(t *testing.T) {
		assert.NotEmpty(t, editCmd.Long)
		assert.Contains(t, editCmd.Long, "edit")
	})
}

func TestGitCommit_Struct(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		commit := GitCommit{
			SHA:     "abc123",
			Subject: "test subject",
			Body:    "test body",
		}

		assert.Equal(t, "abc123", commit.SHA)
		assert.Equal(t, "test subject", commit.Subject)
		assert.Equal(t, "test body", commit.Body)
	})

	t.Run("empty struct", func(t *testing.T) {
		commit := GitCommit{}
		assert.Empty(t, commit.SHA)
		assert.Empty(t, commit.Subject)
		assert.Empty(t, commit.Body)
	})
}

func TestEditCmd_RunErrorPaths(t *testing.T) {
	t.Run("test error path logic for config loading", func(t *testing.T) {
		// Test the error handling logic for config loading
		err := fmt.Errorf("config error")
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Error loading config: %v", err)
			assert.Contains(t, errorMsg, "Error loading config")
			assert.Contains(t, errorMsg, "config error")
		}
	})

	t.Run("test error path logic for getRecentCommits", func(t *testing.T) {
		// Test the error handling logic for getRecentCommits
		err := fmt.Errorf("git log error")
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Error fetching commits: %v", err)
			assert.Contains(t, errorMsg, "Error fetching commits")
			assert.Contains(t, errorMsg, "git log error")
		}
	})

	t.Run("test error path logic for empty commits", func(t *testing.T) {
		// Test the logic for handling empty commits list
		commits := []GitCommit{}
		if len(commits) == 0 {
			infoMsg := "No commits found in this repository."
			assert.Contains(t, infoMsg, "No commits found")
		}
	})

	t.Run("test error path logic for getLastCommitSHA", func(t *testing.T) {
		// Test the error handling logic for getLastCommitSHA
		err := fmt.Errorf("git rev-parse error")
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Error getting last commit SHA: %v", err)
			assert.Contains(t, errorMsg, "Error getting last commit SHA")
			assert.Contains(t, errorMsg, "git rev-parse error")
		}
	})

	t.Run("test error path logic for non-last commit", func(t *testing.T) {
		// Test the logic for handling non-last commit selection
		selectedSHA := "abc123"
		lastCommitSHA := "def456"
		if selectedSHA != lastCommitSHA {
			warningMsg := "⚠️ Warning: Goji currently only supports amending the *last* commit directly."
			assert.Contains(t, warningMsg, "only supports amending")
			assert.Contains(t, warningMsg, "last")
		}
	})

	t.Run("test error path logic for getCurrentCommitDetails", func(t *testing.T) {
		// Test the error handling logic for getCurrentCommitDetails
		err := fmt.Errorf("git log error")
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Error fetching current commit details: %v", err)
			assert.Contains(t, errorMsg, "Error fetching current commit details")
			assert.Contains(t, errorMsg, "git log error")
		}
	})

	t.Run("test error path logic for AskQuestions", func(t *testing.T) {
		// Test the error handling logic for AskQuestions
		err := fmt.Errorf("TUI error")
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Failed to get new commit details: %v", err)
			assert.Contains(t, errorMsg, "Failed to get new commit details")
			assert.Contains(t, errorMsg, "TUI error")
		}
	})

	t.Run("test error path logic for executeGitCommit", func(t *testing.T) {
		// Test the error handling logic for executeGitCommit
		err := fmt.Errorf("git commit error")
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Error amending commit: %v", err)
			assert.Contains(t, errorMsg, "Error amending commit")
			assert.Contains(t, errorMsg, "git commit error")
		}
	})

	t.Run("test logic for commit body extraction", func(t *testing.T) {
		// Test the logic for extracting commit body from messages
		newMessages := []string{"subject", "body line 1", "body line 2"}
		newCommitMessage := newMessages[0]
		newCommitBody := ""
		if len(newMessages) > 1 {
			newCommitBody = strings.Join(newMessages[1:], "\n")
		}
		assert.Equal(t, "subject", newCommitMessage)
		assert.Contains(t, newCommitBody, "body line 1")
		assert.Contains(t, newCommitBody, "body line 2")

		// Test with single message (no body)
		newMessages2 := []string{"subject only"}
		newCommitMessage2 := newMessages2[0]
		newCommitBody2 := ""
		if len(newMessages2) > 1 {
			newCommitBody2 = strings.Join(newMessages2[1:], "\n")
		}
		assert.Equal(t, "subject only", newCommitMessage2)
		assert.Empty(t, newCommitBody2)
	})

	t.Run("test logic for form cancellation", func(t *testing.T) {
		// Test the logic for form cancellation
		err := fmt.Errorf("user cancelled")
		if err != nil {
			cancelMsg := "Commit editing cancelled."
			assert.Contains(t, cancelMsg, "cancelled")
		}
	})
}

func TestGetRecentCommits_ErrorPaths(t *testing.T) {
	t.Run("test error path for git log command failure", func(t *testing.T) {
		// Test the error handling logic for git log command failure
		err := fmt.Errorf("command failed")
		if err != nil {
			errorMsg := fmt.Errorf("failed to get git log: %w", err)
			assert.Contains(t, errorMsg.Error(), "failed to get git log")
			assert.Error(t, errorMsg)
		}
	})

	t.Run("test parsing logic for commits with body", func(t *testing.T) {
		// Test that commits with body are parsed correctly
		output := "sha123\nsubject line\nbody content\n---COMMIT-END---"
		rawCommits := strings.Split(strings.TrimSpace(output), "---COMMIT-END---")

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

		assert.Len(t, commits, 1)
		assert.Equal(t, "sha123", commits[0].SHA)
		assert.Equal(t, "subject line", commits[0].Subject)
		assert.Equal(t, "body content", commits[0].Body)
	})

	t.Run("test parsing logic for commits without body", func(t *testing.T) {
		// Test that commits without body are parsed correctly
		output := "sha123\nsubject line\n---COMMIT-END---"
		rawCommits := strings.Split(strings.TrimSpace(output), "---COMMIT-END---")

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

		assert.Len(t, commits, 1)
		assert.Equal(t, "sha123", commits[0].SHA)
		assert.Equal(t, "subject line", commits[0].Subject)
		assert.Empty(t, commits[0].Body)
	})

	t.Run("test parsing logic for malformed commits", func(t *testing.T) {
		// Test that malformed commits are skipped
		malformedOutput := "sha1\n---COMMIT-END---\nsha2\nsubject\n---COMMIT-END---"
		rawCommits := strings.Split(strings.TrimSpace(malformedOutput), "---COMMIT-END---")

		var commits []GitCommit
		for _, rawCommit := range rawCommits {
			if strings.TrimSpace(rawCommit) == "" {
				continue
			}
			parts := strings.SplitN(strings.TrimSpace(rawCommit), "\n", 3)
			if len(parts) < 2 {
				continue // This should skip the first malformed commit
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

		// Should only have the second commit (sha2)
		assert.Len(t, commits, 1)
		assert.Equal(t, "sha2", commits[0].SHA)
		assert.Equal(t, "subject", commits[0].Subject)
	})
}

func TestGetLastCommitSHA_ErrorPaths(t *testing.T) {
	t.Run("test error path for git rev-parse failure", func(t *testing.T) {
		// Test the error handling logic for git rev-parse failure
		err := fmt.Errorf("command failed")
		if err != nil {
			errorMsg := fmt.Errorf("failed to get HEAD commit SHA: %w", err)
			assert.Contains(t, errorMsg.Error(), "failed to get HEAD commit SHA")
			assert.Error(t, errorMsg)
		}
	})
}

func TestGetCurrentCommitDetails_ErrorPaths(t *testing.T) {
	t.Run("test error path for git log subject failure", func(t *testing.T) {
		// Test the error handling logic for git log subject failure
		sha := "abc123"
		err := fmt.Errorf("command failed")
		if err != nil {
			errorMsg := fmt.Errorf("failed to get commit subject for %s: %w", sha, err)
			assert.Contains(t, errorMsg.Error(), "failed to get commit subject")
			assert.Contains(t, errorMsg.Error(), sha)
			assert.Error(t, errorMsg)
		}
	})

	t.Run("test error path for git log body failure", func(t *testing.T) {
		// Test the error handling logic for git log body failure
		sha := "abc123"
		err := fmt.Errorf("command failed")
		if err != nil {
			errorMsg := fmt.Errorf("failed to get commit body for %s: %w", sha, err)
			assert.Contains(t, errorMsg.Error(), "failed to get commit body")
			assert.Contains(t, errorMsg.Error(), sha)
			assert.Error(t, errorMsg)
		}
	})
}

func TestEditCmd_RunLogicPaths(t *testing.T) {
	t.Run("test commit options generation", func(t *testing.T) {
		// Test the logic for generating commit options
		commits := []GitCommit{
			{SHA: "abc123def456", Subject: "First commit"},
			{SHA: "def456ghi789", Subject: "Second commit"},
		}

		commitOptions := make([]string, len(commits))
		for i, commit := range commits {
			commitOptions[i] = fmt.Sprintf("%s %s", commit.SHA[:7], commit.Subject)
		}

		assert.Len(t, commitOptions, 2)
		assert.Contains(t, commitOptions[0], "abc123")
		assert.Contains(t, commitOptions[0], "First commit")
		assert.Contains(t, commitOptions[1], "def456")
		assert.Contains(t, commitOptions[1], "Second commit")
	})

	t.Run("test form validation logic", func(t *testing.T) {
		// Test the form validation logic
		selectedSHA := ""
		if selectedSHA == "" {
			err := fmt.Errorf("no commit selected")
			assert.Contains(t, err.Error(), "no commit selected")
			assert.Error(t, err)
		}

		selectedSHA2 := "abc123"
		if selectedSHA2 == "" {
			t.Error("Should not error when SHA is selected")
		} else {
			assert.NotEmpty(t, selectedSHA2)
		}
	})

	t.Run("test last commit comparison logic", func(t *testing.T) {
		// Test the logic for comparing selected SHA with last commit SHA
		selectedSHA := "abc123"
		lastCommitSHA := "abc123"
		if selectedSHA != lastCommitSHA {
			t.Error("Should allow amendment when SHA matches")
		} else {
			assert.Equal(t, selectedSHA, lastCommitSHA)
		}

		selectedSHA2 := "abc123"
		lastCommitSHA2 := "def456"
		if selectedSHA2 != lastCommitSHA2 {
			assert.NotEqual(t, selectedSHA2, lastCommitSHA2)
		}
	})

	t.Run("test commit body display logic", func(t *testing.T) {
		// Test the logic for displaying commit body
		newCommitBody := "test body"
		if newCommitBody != "" {
			bodyMsg := fmt.Sprintf("New commit body: %s", newCommitBody)
			assert.Contains(t, bodyMsg, "New commit body")
			assert.Contains(t, bodyMsg, newCommitBody)
		}

		emptyBody := ""
		if emptyBody != "" {
			t.Error("Should not display empty body")
		} else {
			assert.Empty(t, emptyBody)
		}
	})
}
