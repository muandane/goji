package config

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestLoadConfig(t *testing.T) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	rootDirBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error finding git root directory: %v", err)
	}
	rootDir := string(rootDirBytes)
	rootDir = strings.TrimSpace(rootDir) // Remove newline character at the end

	// Prepare a temporary configuration file for testing
	filename := "test_config.json"
	content := `{
		"Types": [
			{
				"Emoji": "✨",
				"Code": ":sparkles:",
				"Description": "Introducing new features.",
				"Name": "feat"
			},
			{
				"Emoji": "🐛",
				"Code": ":bug:",
				"Description": "Fixing a bug.",
				"Name": "fix"
			},
			{
				"Emoji": "🧹",
				"Code": ":broom:",
				"Description": "A chore change.",
				"Name": "chore"
			}
		],
		"Scopes": ["home", "accounts", "ci"],
		"Symbol": true,
		"SkipQuestions": [],
		"SubjectMaxLength": 50
	}
	`

	tmpDir, err := os.MkdirTemp("", "goji-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	// Set the temporary directory as the home directory for the test
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	// Move the configuration file from the current directory to the temporary home directory

	testConfigPath := filepath.Join(rootDir, filename)
	err = os.WriteFile(testConfigPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(testConfigPath)
	err = os.Rename(testConfigPath, filepath.Join(tmpDir, filename))
	if err != nil {
		t.Fatalf("Failed to move test config file to the temporary home directory: %v", err)
	}

	// Test the LoadConfig function
	config, err := LoadConfig(filename)
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("config is nil")
	}

	expectedTypeName := "feat"
	if config.Types[0].Name != expectedTypeName {
		t.Errorf("Expected type name %s, got %s", expectedTypeName, config.Types[0].Name)
	}
	expectedEmoji := "✨"
	if config.Types[0].Emoji != expectedEmoji {
		t.Errorf("Expected emoji %s, got %s", expectedEmoji, config.Types[0].Emoji)
	}

	expectedCode := ":sparkles:"
	if config.Types[0].Code != expectedCode {
		t.Errorf("Expected code %s, got %s", expectedCode, config.Types[0].Code)
	}

	expectedDescription := "Introducing new features."
	if config.Types[0].Description != expectedDescription {
		t.Errorf("Expected description %s, got %s", expectedDescription, config.Types[0].Description)
	}
}

func TestGitRootDirectory(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a Git repository in the temporary directory
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize Git repository: %v", err)
	}

	// Create a subdirectory in the Git repository
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Execute the command in the subdirectory
	cmd = exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = subDir
	rootDirBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error finding git root directory: %v", err)
	}
	rootDir := string(rootDirBytes)
	rootDir = strings.TrimSpace(rootDir) // Remove newline character at the end

	// Resolve symlinks in the paths
	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks in tmpDir: %v", err)
	}
	resolvedRootDir, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks in rootDir: %v", err)
	}

	// Check if the command returned the correct repository root directory
	if resolvedTmpDir != resolvedRootDir {
		t.Errorf("Expected root directory: %s, got: %s", resolvedTmpDir, resolvedRootDir)
	}
}
func TestLoadConfigInvalidJSON(t *testing.T) {
	// Prepare a temporary configuration file with invalid JSON content
	filename := "test_invalid_config.json"
	content := `{
		"Types": [
			{
				"Emoji": "✨",
				"Code": ":sparkles:",
				"Description": "Introducing new features.",
				"Name": "feat",
			}, // Extra comma here makes the JSON invalid
		}
	}`

	tmpDir, err := os.MkdirTemp("", "goji-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	testConfigPath := filepath.Join(tmpDir, filename)
	err = os.WriteFile(testConfigPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test the LoadConfig function with the invalid JSON file
	config, err := LoadConfig(filename)
	if err == nil {
		t.Error("Expected an error for invalid JSON, but got no error")
	}

	if config != nil {
		t.Error("Expected config to be nil for invalid JSON")
	}
}

var mockedExitStatus = 1
var mockedStdout = ""

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Print the mocked stdout.
	os.Stdout.WriteString(os.Getenv("STDOUT"))

	// Exit with the mocked exit status.
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		"EXIT_STATUS=" + strconv.Itoa(mockedExitStatus)}
	return cmd
}

var execCommand = exec.Command

func TestLoadConfig_Failure(t *testing.T) {
	// Replace the exec.Command function with the mock version.
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	_, err := LoadConfig("config.json")
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}

// Define interfaces for os and exec
type OS interface {
	ReadFile(filename string) ([]byte, error)
}

type Exec interface {
	Command(name string, arg ...string) *exec.Cmd
}

// Create mock types for the interfaces
type MockOS struct {
	mock.Mock
}

func (m *MockOS) ReadFile(filename string) ([]byte, error) {
	args := m.Called(filename)
	return args.Get(0).([]byte), args.Error(1)
}

type MockExec struct {
	mock.Mock
}

func (m *MockExec) Command(name string, arg ...string) *exec.Cmd {
	args := m.Called(name, arg)
	return args.Get(0).(*exec.Cmd)
}
func TestLoadConfig_Failure1(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}

	// Delete the temporary directory after the test
	defer os.RemoveAll(tmpDir)

	// Change the working directory to the temporary directory
	os.Chdir(tmpDir)

	// Try to load a non-existent config file
	_, err = LoadConfig("non_existent_file.json")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

}
