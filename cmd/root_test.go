package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/fatih/color"
	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
	"github.com/muandane/goji/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	testifyAssert "github.com/stretchr/testify/assert"
)

func TestRootCmd_VersionFlag(t *testing.T) {
	var versionFlagTest bool
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{
		Use:   "goji",
		Short: "Goji CLI",
		Long:  `Goji is a cli tool to generate conventional commits with emojis`,
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlagTest {
				color.Set(color.FgGreen)

				fmt.Printf("goji version: v%s\n", "test-version")
				color.Unset()
				return
			}
		},
	}

	cmd.Flags().BoolVarP(&versionFlagTest, "version", "v", false, "Display version information")
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	w.Close()
	os.Stdout = originalStdout
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Error reading captured output: %v", err)
	}
	output := buf.String()
	assert.Contains(t, output, "goji version: vtest-version")
}

func setupRootTestConfig(t *testing.T, dir string, noEmoji bool, signOff bool) string {
	t.Helper()
	cfg := config.Config{
		Types: []models.CommitType{
			{Name: "feat", Emoji: "✨", Description: "Features"},
			{Name: "fix", Emoji: "🐛", Description: "Bug fixes"},
		},
		NoEmoji: noEmoji,
		SignOff: signOff,
	}
	content, err := json.Marshal(cfg)
	require.NoError(t, err)

	configFile := filepath.Join(dir, ".goji.json")
	err = os.WriteFile(configFile, content, 0644)
	require.NoError(t, err)
	return configFile
}

func TestConstructCommitMessage(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		typeFlag      string
		scopeFlag     string
		messageFlag   string
		expected      string
		noEmojiConfig bool
	}{
		{
			name: "with emoji, no scope",
			cfg: &config.Config{
				Types:   []models.CommitType{{Name: "feat", Emoji: "✨"}},
				NoEmoji: false,
			},
			typeFlag:    "feat",
			messageFlag: "new feature",
			expected:    "feat ✨: new feature",
		},
		{
			name: "with emoji, with scope",
			cfg: &config.Config{
				Types:   []models.CommitType{{Name: "fix", Emoji: "🐛"}},
				NoEmoji: false,
			},
			typeFlag:    "fix",
			scopeFlag:   "api",
			messageFlag: "bug fix",
			expected:    "fix 🐛 (api): bug fix",
		},
		{
			name: "no emoji (config), no scope",
			cfg: &config.Config{
				Types:   []models.CommitType{{Name: "feat", Emoji: "✨"}},
				NoEmoji: true,
			},
			typeFlag:    "feat",
			messageFlag: "new feature",
			expected:    "feat: new feature",
		},
		{
			name: "no emoji (config), with scope",
			cfg: &config.Config{
				Types:   []models.CommitType{{Name: "docs", Emoji: "📚"}},
				NoEmoji: true,
			},
			typeFlag:    "docs",
			scopeFlag:   "readme",
			messageFlag: "update",
			expected:    "docs (readme): update",
		},
		{
			name: "type not in config (should use raw type)",
			cfg: &config.Config{
				Types:   []models.CommitType{{Name: "feat", Emoji: "✨"}},
				NoEmoji: false,
			},
			typeFlag:    "chore",
			messageFlag: "maintenance",
			expected:    "chore: maintenance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructCommitMessage(tt.cfg, tt.typeFlag, tt.scopeFlag, tt.messageFlag)
			testifyAssert.Equal(t, tt.expected, result)
		})
	}
}

var lastExecutedCmdArgs []string

var realExecCommand = exec.Command

func TestRootCmd_RunE_WithFlags(t *testing.T) {
	tempDir := t.TempDir()
	_ = setupRootTestConfig(t, tempDir, false, false)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	tests := []struct {
		nameArgs      []string
		typeF         string
		messageF      string
		scopeF        string
		expectedError string
		expectGitArgs []string
	}{
		{
			nameArgs: []string{"--type", "feat", "--message", "add new thing", "--scope", "app"},
			typeF:    "feat", messageF: "add new thing", scopeF: "app",
			expectGitArgs: []string{"commit", "-m", "feat ✨ (app): add new thing"},
		},
		{
			nameArgs: []string{"-t", "fix", "-m", "resolve bug"},
			typeF:    "fix", messageF: "resolve bug",
			expectGitArgs: []string{"commit", "-m", "fix 🐛: resolve bug"},
		},
		{
			nameArgs:      []string{"-t", "feat"},
			typeF:         "feat",
			expectedError: "commit message cannot be empty",
		},
	}

	originalExec := execCommand
	defer func() { execCommand = originalExec }()

	for _, tt := range tests {
		t.Run(strings.Join(tt.nameArgs, "_"), func(t *testing.T) {

			testRootCmd, _ := createTestRootCmd()

			execCommand = func(name string, arg ...string) *exec.Cmd {
				if name == "git" {
					lastExecutedCmdArgs = append([]string{name}, arg...)

					return exec.Command("true")
				}
				return realExecCommand(name, arg...)
			}
			lastExecutedCmdArgs = nil

			testRootCmd.Flags().Set("type", tt.typeF)
			testRootCmd.Flags().Set("message", tt.messageF)
			if tt.scopeF != "" {
				testRootCmd.Flags().Set("scope", tt.scopeF)
			}

			err := testRootCmd.RunE(testRootCmd, tt.nameArgs)

			if tt.expectedError != "" {
				require.Error(t, err)
				testifyAssert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.expectGitArgs != nil {
					testifyAssert.Equal(t, tt.expectGitArgs, lastExecutedCmdArgs, "Git command args mismatch")
				}
			}
		})
	}
}

func TestRootCmd_RunE_Interactive(t *testing.T) {
	tempDir := t.TempDir()
	_ = setupRootTestConfig(t, tempDir, false, true)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	originalAskQuestions := utilsAskQuestions
	defer func() { utilsAskQuestions = originalAskQuestions }()

	originalExec := execCommand
	defer func() { execCommand = originalExec }()

	tests := []struct {
		name            string
		askReturnMsg    string
		askReturnBody   string
		askReturnErr    error
		expectedError   string
		expectGitArgs   []string
		signOffExpected bool
	}{
		{
			name:            "successful interactive commit",
			askReturnMsg:    "feat ✨ (ui): interactive commit",
			askReturnBody:   "This is the body.",
			expectGitArgs:   []string{"commit", "-m", "feat ✨ (ui): interactive commit", "-m", "This is the body.", "--signoff"},
			signOffExpected: true,
		},
		{
			name:          "AskQuestions returns error",
			askReturnErr:  fmt.Errorf("user cancelled"),
			expectedError: "failed to get commit details: user cancelled",
		},
		{
			name:          "AskQuestions returns empty message",
			askReturnMsg:  "",
			expectedError: "commit message cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRootCmd, _ := createTestRootCmd()

			utilsAskQuestions = func(cfg *config.Config, presetType, presetMessage string) ([]string, error) {
				return []string{tt.askReturnMsg, tt.askReturnBody}, tt.askReturnErr
			}

			execCommand = func(name string, arg ...string) *exec.Cmd {
				if name == "git" {
					lastExecutedCmdArgs = append([]string{name}, arg...)
					return exec.Command("true")
				}
				return realExecCommand(name, arg...)
			}
			lastExecutedCmdArgs = nil

			err := testRootCmd.RunE(testRootCmd, []string{})

			if tt.expectedError != "" {
				require.Error(t, err)
				testifyAssert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.expectGitArgs != nil {
					testifyAssert.Equal(t, tt.expectGitArgs, lastExecutedCmdArgs)
				}
			}
		})
	}
}

func TestExecuteGitCommit(t *testing.T) {
	originalExec := execCommand
	defer func() { execCommand = originalExec }()

	tests := []struct {
		name         string
		message      string
		body         string
		signOff      bool
		noVerify     bool
		add          bool
		amend        bool
		gitFlags     []string
		expectedArgs []string
		cmdError     bool
	}{
		{"simple", "feat: msg", "", false, false, false, false, nil, []string{"git", "commit", "-m", "feat: msg"}, false},
		{"with body", "fix: msg", "body text", false, false, false, false, nil, []string{"git", "commit", "-m", "fix: msg", "-m", "body text"}, false},
		{"signoff", "docs: msg", "", true, false, false, false, nil, []string{"git", "commit", "-m", "docs: msg", "--signoff"}, false},
		{"no-verify", "test: msg", "", false, true, false, false, nil, []string{"git", "commit", "-m", "test: msg", "--no-verify"}, false},
		{"add", "ci: msg", "", false, false, true, false, nil, []string{"git", "commit", "-m", "ci: msg", "-a"}, false},
		{"amend", "refactor: msg", "", false, false, false, true, nil, []string{"git", "commit", "-m", "refactor: msg", "--amend"}, false},
		{"with git-flag", "style: msg", "", false, false, false, false, []string{"--no-edit"}, []string{"git", "commit", "-m", "style: msg", "--no-edit"}, false},
		{"all flags", "perf: msg", "more details", true, true, true, true, []string{"-S"}, []string{"git", "commit", "-m", "perf: msg", "-m", "more details", "--signoff", "--no-verify", "-a", "--amend", "-S"}, false},
		{"command error", "error: msg", "", false, false, false, false, nil, []string{"git", "commit", "-m", "error: msg"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			noVerifyFlag = tt.noVerify
			addFlag = tt.add
			amendFlag = tt.amend
			gitFlags = tt.gitFlags
			lastExecutedCmdArgs = nil

			execCommand = func(name string, arg ...string) *exec.Cmd {
				lastExecutedCmdArgs = append([]string{name}, arg...)
				if tt.cmdError {
					return exec.Command("false")
				}
				return exec.Command("true")
			}

			err := executeGitCommit(tt.message, tt.body, tt.signOff)

			if tt.cmdError {
				testifyAssert.Error(t, err)
				testifyAssert.Contains(t, err.Error(), "git command failed")
			} else {
				testifyAssert.NoError(t, err)
			}
			testifyAssert.Equal(t, tt.expectedArgs, lastExecutedCmdArgs)
		})
	}
}

func createTestRootCmd() (*cobra.Command, *config.Config) {

	cmd := &cobra.Command{
		Use:           "goji",
		Short:         "Goji CLI",
		Long:          `Goji is a CLI tool to generate conventional commits with emojis`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var (
		testTypeFlag, testMessageFlag, testScopeFlag string
		testNoVerifyFlag, testAddFlag, testAmendFlag bool
		testGitFlags                                 []string
	)

	cmd.Flags().StringVarP(&testTypeFlag, "type", "t", "", "Specify the type from the config file")
	cmd.Flags().StringVarP(&testScopeFlag, "scope", "s", "", "Specify a custom scope")
	cmd.Flags().StringVarP(&testMessageFlag, "message", "m", "", "Specify a commit message")
	cmd.Flags().BoolVarP(&testNoVerifyFlag, "no-verify", "n", false, "bypass pre-commit and commit-msg hooks")

	cmd.Flags().BoolVarP(&testAddFlag, "add", "a", false, "Automatically stage files that have been modified and deleted")
	cmd.Flags().BoolVar(&testAmendFlag, "amend", false, "Change last commit")
	cmd.Flags().StringArrayVar(&testGitFlags, "git-flag", []string{}, "Git flags (can be used multiple times)")

	cmd.RunE = func(c *cobra.Command, args []string) error {

		cfg, err := config.ViperConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		typeF, _ := c.Flags().GetString("type")
		messageF, _ := c.Flags().GetString("message")
		scopeF, _ := c.Flags().GetString("scope")

		noVerifyFlag, _ = c.Flags().GetBool("no-verify")
		addFlag, _ = c.Flags().GetBool("add")
		amendFlagGlobal, _ := c.Flags().GetBool("amend")
		amendFlag = amendFlagGlobal

		gitFlags, _ = c.Flags().GetStringArray("git-flag")

		var commitMessage, commitBody string
		if typeF != "" && messageF != "" {
			commitMessage = constructCommitMessage(cfg, typeF, scopeF, messageF)
		} else {

			messages, err := utilsAskQuestions(cfg, typeF, messageF)
			if err != nil {
				return fmt.Errorf("failed to get commit details: %w", err)
			}
			if len(messages) < 1 || messages[0] == "" {
				return fmt.Errorf("commit message cannot be empty")
			}
			commitMessage = messages[0]
			if len(messages) > 1 {
				commitBody = messages[1]
			}
		}

		if commitMessage == "" {
			return fmt.Errorf("commit message cannot be empty")
		}

		return executeGitCommit(commitMessage, commitBody, cfg.SignOff)
	}

	cfgForReturn, _ := config.ViperConfig()
	return cmd, cfgForReturn
}

var utilsAskQuestions = utils.AskQuestions
var execCommand = exec.Command
