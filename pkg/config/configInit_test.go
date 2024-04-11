package config

import (
	"reflect"
	"testing"
)

func TestAddCustomCommitTypes(t *testing.T) {
	gitmojis := []Gitmoji{
		{Emoji: "🚀", Code: ":rocket:", Description: "Deploy stuff.", Name: "deploy"},
		{Emoji: "🐞", Code: ":bug:", Description: "Fix a bug.", Name: "fix"},
	}

	expected := []Gitmoji{
		{Emoji: "🚀", Code: ":rocket:", Description: "Deploy stuff.", Name: "deploy"},
		{Emoji: "🐞", Code: ":bug:", Description: "Fix a bug.", Name: "fix"},
		{Emoji: "✨", Code: ":sparkles:", Description: "Introduce new features.", Name: "feat"},
		{Emoji: "🐛", Code: ":bug:", Description: "Fix a bug.", Name: "fix"},
		{Emoji: "📚", Code: ":books:", Description: "Documentation change.", Name: "docs"},
		{Emoji: "🎨", Code: ":art:", Description: "Improve structure/format of the code.", Name: "refactor"},
		{Emoji: "🧹", Code: ":broom:", Description: "A chore change.", Name: "chore"},
		{Emoji: "🧪", Code: ":test_tube:", Description: "Add a test.", Name: "test"},
		{Emoji: "🚑️", Code: ":ambulance:", Description: "Critical hotfix.", Name: "hotfix"},
		{Emoji: "⚰️", Code: ":coffin:", Description: "Remove dead code.", Name: "deprecate"},
		{Emoji: "⚡️", Code: ":zap:", Description: "Improve performance.", Name: "perf"},
		{Emoji: "🚧", Code: ":construction:", Description: "Work in progress.", Name: "wip"},
		{Emoji: "📦", Code: ":package:", Description: "Add or update compiled files or packages.", Name: "package"},
	}

	result := AddCustomCommitTypes(gitmojis)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestGetGitRootDir(t *testing.T) {
	_, err := GetGitRootDir()
	if err != nil {
		t.Errorf("GetGitRootDir() error = %v", err)
	}
}

func TestInitRepoConfig(t *testing.T) {
	err := InitRepoConfig(false, true)
	if err != nil {
		t.Errorf("InitRepoConfig() error = %v", err)
	}
}
