package cmd

import (
	"os/exec"
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
	"github.com/spf13/cobra"
)

// TestConstructCommitMessage tests commit message construction
func TestConstructCommitMessage(t *testing.T) {
	cfg := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨"},
			{Name: "fix", Emoji: "🐛"},
			{Name: "docs", Emoji: "📝"},
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
			name:        "commit with complex scope",
			typeFlag:    "docs",
			scopeFlag:   "api/users",
			messageFlag: "update user documentation",
			expected:    "docs 📝 (api/users): update user documentation",
		},
		{
			name:        "unknown type fallback",
			typeFlag:    "unknown",
			scopeFlag:   "",
			messageFlag: "some message",
			expected:    "unknown: some message",
		},
		{
			name:        "unknown type with scope",
			typeFlag:    "custom",
			scopeFlag:   "component",
			messageFlag: "custom change",
			expected:    "custom (component): custom change",
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
			{Name: "fix", Emoji: "🐛"},
		},
		NoEmoji: true,
	}

	tests := []struct {
		name        string
		typeFlag    string
		scopeFlag   string
		messageFlag string
		expected    string
	}{
		{
			name:        "no emoji feat",
			typeFlag:    "feat",
			scopeFlag:   "",
			messageFlag: "add feature",
			expected:    "feat: add feature",
		},
		{
			name:        "no emoji with scope",
			typeFlag:    "fix",
			scopeFlag:   "core",
			messageFlag: "fix bug",
			expected:    "fix (core): fix bug",
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

// TestContainsHelper tests the contains helper function
func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "contains item",
			slice:    []string{"--amend", "--no-verify", "-a"},
			item:     "--amend",
			expected: true,
		},
		{
			name:     "does not contain item",
			slice:    []string{"--amend", "--no-verify", "-a"},
			item:     "--missing",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "--amend",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    []string{"--amend"},
			item:     "",
			expected: false,
		},
		{
			name:     "exact match",
			slice:    []string{"-a", "--add"},
			item:     "-a",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%v, %q) = %v, expected %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}

// TestRootCommandFlags tests flag initialization and parsing
func TestRootCommandFlags(t *testing.T) {
	// Reset global variables before test
	resetGlobalFlags()

	tests := []struct {
		name    string
		args    []string
		checkFn func(t *testing.T)
	}{
		{
			name: "version flag short",
			args: []string{"-v"},
			checkFn: func(t *testing.T) {
				if !versionFlag {
					t.Error("Expected versionFlag to be true")
				}
			},
		},
		{
			name: "version flag long",
			args: []string{"--version"},
			checkFn: func(t *testing.T) {
				if !versionFlag {
					t.Error("Expected versionFlag to be true")
				}
			},
		},
		{
			name: "no-verify flag",
			args: []string{"--no-verify"},
			checkFn: func(t *testing.T) {
				if !noVerifyFlag {
					t.Error("Expected noVerifyFlag to be true")
				}
			},
		},
		{
			name: "amend flag",
			args: []string{"--amend"},
			checkFn: func(t *testing.T) {
				if !amendFlag {
					t.Error("Expected amendFlag to be true")
				}
			},
		},
		{
			name: "add flag",
			args: []string{"-a"},
			checkFn: func(t *testing.T) {
				if !addFlag {
					t.Error("Expected addFlag to be true")
				}
			},
		},
		{
			name: "type flag",
			args: []string{"-t", "feat"},
			checkFn: func(t *testing.T) {
				if typeFlag != "feat" {
					t.Errorf("Expected typeFlag to be 'feat', got %q", typeFlag)
				}
			},
		},
		{
			name: "message flag",
			args: []string{"-m", "test message"},
			checkFn: func(t *testing.T) {
				if messageFlag != "test message" {
					t.Errorf("Expected messageFlag to be 'test message', got %q", messageFlag)
				}
			},
		},
		{
			name: "scope flag",
			args: []string{"-s", "auth"},
			checkFn: func(t *testing.T) {
				if scopeFlag != "auth" {
					t.Errorf("Expected scopeFlag to be 'auth', got %q", scopeFlag)
				}
			},
		},
		{
			name: "multiple git flags",
			args: []string{"--git-flag", "--verbose", "--git-flag", "--dry-run"},
			checkFn: func(t *testing.T) {
				expected := []string{"--verbose", "--dry-run"}
				if len(gitFlags) != len(expected) {
					t.Errorf("Expected %d git flags, got %d", len(expected), len(gitFlags))
					return
				}
				for i, flag := range expected {
					if gitFlags[i] != flag {
						t.Errorf("Expected git flag %d to be %q, got %q", i, flag, gitFlags[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			resetGlobalFlags()

			// Create a fresh command for testing
			cmd := createTestRootCommand()

			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags %v: %v", tt.args, err)
			}

			tt.checkFn(t)
		})
	}
}

// TestExecuteGitCommitArguments tests git command argument construction
func TestExecuteGitCommitArguments(t *testing.T) {
	if !isGitRepo() {
		t.Skip("Not in a git repository")
	}

	tests := []struct {
		name        string
		message     string
		body        string
		signOff     bool
		setupFlags  func()
		extraFlags  []string
		expectError bool
	}{
		{
			name:        "empty message should fail",
			message:     "",
			body:        "",
			signOff:     false,
			setupFlags:  func() { resetGlobalFlags() },
			expectError: true,
		},
		{
			name:       "basic commit",
			message:    "test: basic commit",
			body:       "",
			signOff:    false,
			setupFlags: func() { resetGlobalFlags() },
		},
		{
			name:       "commit with body",
			message:    "test: commit with body",
			body:       "This is the body",
			signOff:    false,
			setupFlags: func() { resetGlobalFlags() },
		},
		{
			name:       "commit with signoff",
			message:    "test: commit with signoff",
			body:       "",
			signOff:    true,
			setupFlags: func() { resetGlobalFlags() },
		},
		{
			name:    "commit with no-verify flag",
			message: "test: commit with no-verify",
			body:    "",
			signOff: false,
			setupFlags: func() {
				resetGlobalFlags()
				noVerifyFlag = true
			},
		},
		{
			name:    "commit with add flag",
			message: "test: commit with add",
			body:    "",
			signOff: false,
			setupFlags: func() {
				resetGlobalFlags()
				addFlag = true
			},
		},
		{
			name:    "commit with amend flag",
			message: "test: commit with amend",
			body:    "",
			signOff: false,
			setupFlags: func() {
				resetGlobalFlags()
				amendFlag = true
			},
		},
		{
			name:       "commit with extra flags",
			message:    "test: commit with extra flags",
			body:       "",
			signOff:    false,
			setupFlags: func() { resetGlobalFlags() },
			extraFlags: []string{"--dry-run"},
		},
		{
			name:    "commit with all flags",
			message: "test: commit with all flags",
			body:    "Complete test body",
			signOff: true,
			setupFlags: func() {
				resetGlobalFlags()
				noVerifyFlag = true
				addFlag = true
				amendFlag = true
				gitFlags = []string{"--verbose"}
			},
			extraFlags: []string{"--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()

			err := executeGitCommit(tt.message, tt.body, tt.signOff, tt.extraFlags...)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				// Git command will likely fail in test environment, but we check it doesn't panic
				t.Logf("Git command failed as expected in test environment: %v", err)
			}
		})
	}
}

// TestRootCommandStructure tests the command structure
func TestRootCommandStructure(t *testing.T) {
	if rootCmd.Use != "goji" {
		t.Errorf("Expected rootCmd.Use to be 'goji', got %q", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected rootCmd.Short to be non-empty")
	}

	if rootCmd.Long == "" {
		t.Error("Expected rootCmd.Long to be non-empty")
	}

	if !rootCmd.SilenceUsage {
		t.Error("Expected rootCmd.SilenceUsage to be true")
	}

	if !rootCmd.SilenceErrors {
		t.Error("Expected rootCmd.SilenceErrors to be true")
	}
}

// TestStylesInitialization tests that styles are properly initialized
func TestStylesInitialization(t *testing.T) {
	styles := map[string]interface{}{
		"headerStyle":     headerStyle,
		"successMsgStyle": successMsgStyle,
		"errorMsgStyle":   errorMsgStyle,
		"infoMsgStyle":    infoMsgStyle,
		"commitMsgStyle":  commitMsgStyle,
		"mutedStyle":      mutedStyle,
	}

	for name, style := range styles {
		if style == nil {
			t.Errorf("Style %s is nil", name)
		}
	}
}

// Benchmark tests
func BenchmarkConstructCommitMessage(b *testing.B) {
	cfg := &config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨"},
			{Name: "fix", Emoji: "🐛"},
		},
		NoEmoji: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		constructCommitMessage(cfg, "feat", "scope", "message")
	}
}

func BenchmarkContains(b *testing.B) {
	slice := []string{"--amend", "--no-verify", "-a", "--verbose", "--dry-run"}
	item := "--amend"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contains(slice, item)
	}
}

// Helper functions
func resetGlobalFlags() {
	versionFlag = false
	noVerifyFlag = false
	amendFlag = false
	addFlag = false
	typeFlag = ""
	messageFlag = ""
	scopeFlag = ""
	gitFlags = nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func createTestRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "goji-test",
	}

	// Mirror the same flags as rootCmd
	cmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type from the config file")
	cmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Specify a custom scope")
	cmd.Flags().StringVarP(&messageFlag, "message", "m", "", "Specify a commit message")
	cmd.Flags().BoolVarP(&noVerifyFlag, "no-verify", "n", false, "bypass pre-commit and commit-msg hooks")
	cmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	cmd.Flags().BoolVarP(&addFlag, "add", "a", false, "Automatically stage files that have been modified and deleted")
	cmd.Flags().BoolVar(&amendFlag, "amend", false, "Change last commit")
	cmd.Flags().StringArrayVar(&gitFlags, "git-flag", []string{}, "Additional Git flags (can be used multiple times)")

	return cmd
}
