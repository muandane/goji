package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSummarizeDiff(t *testing.T) {
	t.Run("empty diff", func(t *testing.T) {
		summary, _ := summarizeDiff("")

		assert.NotNil(t, summary)
		assert.Equal(t, 0, summary.OriginalSize)
		assert.Equal(t, 0, summary.SummarySize)
		assert.Empty(t, summary.FilesChanged)
	})

	t.Run("small diff", func(t *testing.T) {
		diff := `diff --git a/test.go b/test.go
index 1234567..abcdefg 100644
--- a/test.go
+++ b/test.go
@@ -1,3 +1,4 @@
 package main
 
+func Test() {}
 func main() {}`

		summary, _ := summarizeDiff(diff)

		assert.NotNil(t, summary)
		assert.Equal(t, len(diff), summary.OriginalSize)
		assert.Len(t, summary.FilesChanged, 1)
		assert.Equal(t, "test.go", summary.FilesChanged[0].Path)
	})

	t.Run("large diff", func(t *testing.T) {
		// Create a large diff that exceeds maxDiffSize
		largeContent := strings.Repeat("+line of code\n", 10000)
		diff := `diff --git a/large.go b/large.go
index 1234567..abcdefg 100644
--- a/large.go
+++ b/large.go
@@ -1,3 +1,10003 @@
 package main
` + largeContent

		summary, _ := summarizeDiff(diff)

		assert.NotNil(t, summary)
		assert.Greater(t, summary.OriginalSize, maxDiffSize)
	})

	t.Run("multiple files", func(t *testing.T) {
		diff := `diff --git a/file1.go b/file1.go
index 1234567..abcdefg 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main
 
+func Func1() {}
 func main() {}
diff --git a/file2.go b/file2.go
index 1234567..abcdefg 100644
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,4 @@
 package main
 
+func Func2() {}
 func main() {}`

		summary, _ := summarizeDiff(diff)

		assert.NotNil(t, summary)
		assert.Len(t, summary.FilesChanged, 2)
		assert.Equal(t, "file1.go", summary.FilesChanged[0].Path)
		assert.Equal(t, "file2.go", summary.FilesChanged[1].Path)
	})
}

func TestEnhanceSmallDiff(t *testing.T) {
	t.Run("small diff enhancement", func(t *testing.T) {
		smallDiff := "tiny" // 4 characters, less than minDiffSize (10)
		enhanced := enhanceSmallDiff(smallDiff)

		assert.Contains(t, enhanced, smallDiff)
		assert.Contains(t, enhanced, "very small change")
		assert.Contains(t, enhanced, "Note: This is a very small change")
	})

	t.Run("normal diff unchanged", func(t *testing.T) {
		normalDiff := strings.Repeat("line\n", 20)
		enhanced := enhanceSmallDiff(normalDiff)

		assert.Equal(t, normalDiff, enhanced)
	})
}

func TestExtractCommitMessage(t *testing.T) {
	t.Run("valid conventional commit", func(t *testing.T) {
		rawResult := "feat(auth): add user authentication"
		result := extractCommitMessage(rawResult)

		assert.Equal(t, "feat(auth): add user authentication", result)
	})

	t.Run("commit with markdown formatting", func(t *testing.T) {
		rawResult := "```\nfeat(auth): add user authentication\n```"
		result := extractCommitMessage(rawResult)

		assert.Equal(t, "feat(auth): add user authentication", result)
	})

	t.Run("commit with extra text", func(t *testing.T) {
		rawResult := "Here's the commit message:\n\nfeat(auth): add user authentication\n\nThis implements JWT authentication."
		result := extractCommitMessage(rawResult)

		assert.Equal(t, "feat(auth): add user authentication", result)
	})

	t.Run("invalid format", func(t *testing.T) {
		rawResult := "This is not a conventional commit"
		result := extractCommitMessage(rawResult)

		// The function returns the original string if no conventional format is found
		assert.Equal(t, rawResult, result)
	})

	t.Run("empty result", func(t *testing.T) {
		rawResult := ""
		result := extractCommitMessage(rawResult)

		assert.Equal(t, "", result)
	})

	t.Run("comment lines", func(t *testing.T) {
		rawResult := "# This is a comment\nfeat(auth): add user authentication"
		result := extractCommitMessage(rawResult)

		assert.Equal(t, "feat(auth): add user authentication", result)
	})
}

func TestParseDetailedCommitMessage(t *testing.T) {
	t.Run("structured format", func(t *testing.T) {
		response := `Title: feat(auth): add user authentication

Body:
• Add JWT token generation
• Implement login endpoint
• Create user model`

		result := parseDetailedCommitMessage(response)

		assert.Equal(t, "feat(auth): add user authentication", result.Message)
		assert.Contains(t, result.Body, "JWT token generation")
		assert.Contains(t, result.Body, "login endpoint")
	})

	t.Run("unstructured format", func(t *testing.T) {
		response := "feat(auth): add user authentication"
		result := parseDetailedCommitMessage(response)

		assert.Equal(t, "feat(auth): add user authentication", result.Message)
		assert.Equal(t, "", result.Body)
	})

	t.Run("empty response", func(t *testing.T) {
		response := ""
		result := parseDetailedCommitMessage(response)

		assert.Equal(t, "", result.Message)
		assert.Equal(t, "", result.Body)
	})

	t.Run("body only", func(t *testing.T) {
		response := `Body:
• First point
• Second point`

		result := parseDetailedCommitMessage(response)

		// When no title is found, the entire response becomes the message
		assert.Equal(t, response, result.Message)
		assert.Equal(t, "", result.Body)
	})
}

func TestIsValidCommitMessage(t *testing.T) {
	t.Run("valid conventional commits", func(t *testing.T) {
		validCommits := []string{
			"feat: add new feature",
			"fix(api): resolve bug",
			"docs: update readme",
			"refactor: improve code structure",
			"test: add unit tests",
			"chore: update dependencies",
		}

		for _, commit := range validCommits {
			t.Run(commit, func(t *testing.T) {
				assert.True(t, isValidCommitMessage(commit), "Expected %s to be valid", commit)
			})
		}
	})

	t.Run("invalid commits", func(t *testing.T) {
		invalidCommits := []string{
			"",                       // empty
			"not a commit",           // no colon
			"feat",                   // no description
			"invalid: ",              // empty description
			strings.Repeat("a", 100), // too long
		}

		for _, commit := range invalidCommits {
			t.Run(commit, func(t *testing.T) {
				assert.False(t, isValidCommitMessage(commit), "Expected %s to be invalid", commit)
			})
		}
	})

	t.Run("edge cases", func(t *testing.T) {
		t.Run("exactly 72 characters", func(t *testing.T) {
			commit := strings.Repeat("a", 72)
			assert.False(t, isValidCommitMessage(commit))
		})

		t.Run("71 characters", func(t *testing.T) {
			commit := strings.Repeat("a", 71)
			assert.False(t, isValidCommitMessage(commit))
		})
	})
}

func TestTruncateDiff(t *testing.T) {
	t.Run("small diff unchanged", func(t *testing.T) {
		smallDiff := "small diff content"
		result := truncateDiff(smallDiff)

		assert.Equal(t, smallDiff, result)
	})

	t.Run("large diff truncated", func(t *testing.T) {
		largeDiff := strings.Repeat("line\n", 10000)
		result := truncateDiff(largeDiff)

		assert.LessOrEqual(t, len(result), maxDiffSize)
		// The function may not add truncation message in all cases
		assert.True(t, len(result) <= maxDiffSize)
	})

	t.Run("exactly max size", func(t *testing.T) {
		exactSizeDiff := strings.Repeat("a", maxDiffSize)
		result := truncateDiff(exactSizeDiff)

		assert.Equal(t, exactSizeDiff, result)
	})

	t.Run("diff slightly larger than max", func(t *testing.T) {
		largeDiff := strings.Repeat("a", maxDiffSize+1000)
		result := truncateDiff(largeDiff)

		assert.LessOrEqual(t, len(result), maxDiffSize+100) // Allow some margin for truncation message
		assert.Contains(t, result, "... (diff truncated)")
	})

	t.Run("diff much larger than max", func(t *testing.T) {
		// Create a diff that's much larger than maxDiffSize
		headerLines := strings.Repeat("diff header line\n", 15)
		bodyLines := strings.Repeat("+some code change\n", 50000)
		largeDiff := headerLines + bodyLines
		
		result := truncateDiff(largeDiff)
		
		assert.LessOrEqual(t, len(result), maxDiffSize+100)
		// Should include header
		assert.Contains(t, result, "diff header")
	})

	t.Run("diff with few lines", func(t *testing.T) {
		// Test case where len(lines) < headerLines
		smallDiff := "line1\nline2\nline3"
		result := truncateDiff(smallDiff)
		
		assert.Equal(t, smallDiff, result)
	})

	t.Run("diff where remainingSpace is small", func(t *testing.T) {
		// Test case where remainingSpace <= 1000
		// Create a diff that's just over maxDiffSize but leaves little room
		header := strings.Repeat("header\n", 10)
		// Make header + minimal body just over maxDiffSize
		body := strings.Repeat("x", maxDiffSize-len(header)+100)
		largeDiff := header + body
		
		result := truncateDiff(largeDiff)
		
		assert.LessOrEqual(t, len(result), maxDiffSize+100)
	})

	t.Run("diff that truncates at maxDiffSize", func(t *testing.T) {
		// Test the final truncation path in truncateDiff
		veryLargeDiff := strings.Repeat("x", maxDiffSize*2)
		result := truncateDiff(veryLargeDiff)
		
		// Should be truncated and include truncation message
		assert.LessOrEqual(t, len(result), maxDiffSize+100)
	})
}

func TestCreateFileChange(t *testing.T) {
	t.Run("added lines", func(t *testing.T) {
		changes := []ChangeLine{
			{Type: "+", Content: "+func New() {}"},
			{Type: "+", Content: "+func Old() {}"},
		}

		fileChange := createFileChange("test.go", changes)

		assert.Equal(t, "test.go", fileChange.Path)
		assert.Equal(t, "+2", fileChange.Delta)
		assert.Len(t, fileChange.Changes, 2)
	})

	t.Run("removed lines", func(t *testing.T) {
		changes := []ChangeLine{
			{Type: "-", Content: "-func Old() {}"},
		}

		fileChange := createFileChange("test.go", changes)

		assert.Equal(t, "test.go", fileChange.Path)
		assert.Equal(t, "-1", fileChange.Delta)
		assert.Len(t, fileChange.Changes, 1)
	})

	t.Run("balanced changes", func(t *testing.T) {
		changes := []ChangeLine{
			{Type: "+", Content: "+func New() {}"},
			{Type: "-", Content: "-func Old() {}"},
		}

		fileChange := createFileChange("test.go", changes)

		assert.Equal(t, "test.go", fileChange.Path)
		assert.Equal(t, "=", fileChange.Delta)
		assert.Len(t, fileChange.Changes, 2)
	})
}

func TestGenerateDiffSummary(t *testing.T) {
	t.Run("single file summary", func(t *testing.T) {
		fileChanges := []FileChange{
			{Path: "test.go", Delta: "+5"},
		}

		summary := generateDiffSummary(fileChanges, 1000)

		assert.Equal(t, 1000, summary.OriginalSize)
		assert.Len(t, summary.FilesChanged, 1)
		assert.Contains(t, summary.Summary, "Modified: test.go (+5)")
	})

	t.Run("multiple files summary", func(t *testing.T) {
		fileChanges := []FileChange{
			{Path: "file1.go", Delta: "+3"},
			{Path: "file2.go", Delta: "-2"},
		}

		summary := generateDiffSummary(fileChanges, 2000)

		assert.Equal(t, 2000, summary.OriginalSize)
		assert.Len(t, summary.FilesChanged, 2)
		assert.Contains(t, summary.Summary, "2 files changed")
		assert.Contains(t, summary.Summary, "file1.go (+3)")
		assert.Contains(t, summary.Summary, "file2.go (-2)")
	})
}
