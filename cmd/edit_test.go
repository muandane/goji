package cmd

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
	"github.com/spf13/cobra"
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

	result := constructCommitMessage(cfg, "feat", "", "add feature")
	expected := "feat: add feature"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
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
	w.Close()
	os.Stdout = old
	io.Copy(io.Discard, r)
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
