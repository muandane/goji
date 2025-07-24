package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected string
	}{
		{
			name:     "small diff unchanged",
			diff:     "diff --git a/file.txt b/file.txt\nindex 123..456 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old\n+new",
			expected: "diff --git a/file.txt b/file.txt\nindex 123..456 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old\n+new",
		},
		{
			name: "large diff truncated",
			diff: strings.Repeat("line with content\n", 10000), // ~180KB
			expected: func() string {
				// Should be truncated to maxDiffSize
				lines := strings.Split(strings.Repeat("line with content\n", 10000), "\n")
				headerLines := 10
				truncatedLines := lines[:headerLines]
				truncatedLines = append(truncatedLines, "... (diff truncated)")
				// Add some lines from the end
				remainingSpace := maxDiffSize - len(strings.Join(truncatedLines, "\n"))
				linesFromEnd := remainingSpace / 100
				if linesFromEnd > 0 {
					truncatedLines = append(truncatedLines, lines[len(lines)-linesFromEnd:]...)
				}
				return strings.Join(truncatedLines, "\n")
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateDiff(tt.diff)
			assert.LessOrEqual(t, len(result), maxDiffSize)
			if len(tt.diff) <= maxDiffSize {
				assert.Equal(t, tt.diff, result)
			} else {
				assert.Contains(t, result, "... (diff truncated)")
			}
		})
	}
}

func TestEnhanceSmallDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected string
	}{
		{
			name:     "small diff enhanced",
			diff:     "a",
			expected: "a\n\n# Note: This is a very small change. Please focus on the specific modification shown above.",
		},
		{
			name:     "large diff unchanged",
			diff:     strings.Repeat("line with content\n", 100),
			expected: strings.Repeat("line with content\n", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enhanceSmallDiff(tt.diff)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
	}{
		{
			name:     "conventional commit format",
			response: "feat(api): add new endpoint\n# comment\n```code```",
			expected: "feat(api): add new endpoint",
		},
		{
			name:     "with markdown formatting",
			response: "```\nfix(parser): resolve issue\n```",
			expected: "fix(parser): resolve issue",
		},
		{
			name:     "first non-empty line",
			response: "\n\n\nsimple message\n# comment",
			expected: "simple message",
		},
		{
			name:     "short response",
			response: "short",
			expected: "short",
		},
		{
			name:     "only comments and whitespace",
			response: "# comment\n\n   \n```\n```",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCommitMessage(tt.response)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "valid conventional commit",
			message:  "feat(api): add new endpoint",
			expected: true,
		},
		{
			name:     "valid without scope",
			message:  "fix: resolve bug",
			expected: true,
		},
		{
			name:     "too long",
			message:  strings.Repeat("a", 100),
			expected: false,
		},
		{
			name:     "missing colon",
			message:  "feat add feature",
			expected: false,
		},
		{
			name:     "empty description",
			message:  "feat: ",
			expected: false,
		},
		{
			name:     "empty type",
			message:  ": description",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCommitMessage(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}
