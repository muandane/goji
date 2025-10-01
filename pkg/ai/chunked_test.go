package ai

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIProvider for testing
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) GenerateCommitMessage(diff, commitTypes, extraContext string) (string, error) {
	args := m.Called(diff, commitTypes, extraContext)
	return args.String(0), args.Error(1)
}

func (m *MockAIProvider) GenerateDetailedCommit(diff, commitTypes, extraContext string) (*CommitResult, error) {
	args := m.Called(diff, commitTypes, extraContext)
	return args.Get(0).(*CommitResult), args.Error(1)
}

func (m *MockAIProvider) GetModel() string {
	args := m.Called()
	return args.String(0)
}

func TestChunkedDiffProcessor_CreateAggressiveSummary(t *testing.T) {
	processor := &ChunkedDiffProcessor{}

	t.Run("creates aggressive summary", func(t *testing.T) {
		diff := `diff --git a/src/main.go b/src/main.go
+func main() {
+    fmt.Println("hello")
+}
diff --git a/README.md b/README.md
+# Project`

		summary := processor.createAggressiveSummary(diff)

		assert.Contains(t, summary, "Summary of changes:")
		assert.Contains(t, summary, "src/main.go: +3/-0")
		assert.Contains(t, summary, "README.md: +1/-0")
		assert.Contains(t, summary, "+func main() {")
	})

	t.Run("handles empty diff", func(t *testing.T) {
		summary := processor.createAggressiveSummary("")

		assert.Contains(t, summary, "Summary of changes:")
	})
}

func TestChunkedDiffProcessor_ProcessChunkedDiff(t *testing.T) {
	t.Run("small diff processes normally", func(t *testing.T) {
		mockProvider := &MockAIProvider{}
		processor := NewChunkedDiffProcessor(mockProvider)

		smallDiff := "diff --git a/file.txt b/file.txt\n+line1"
		expectedResult := "feat: add line1"

		mockProvider.On("GenerateCommitMessage", smallDiff, "types", "context").Return(expectedResult, nil)

		result, err := processor.ProcessChunkedDiff(smallDiff, "types", "context")

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockProvider.AssertExpectations(t)
	})

	t.Run("large diff processes in chunks", func(t *testing.T) {
		mockProvider := &MockAIProvider{}
		processor := NewChunkedDiffProcessor(mockProvider)

		// Create a large diff that will definitely be chunked
		var largeDiff strings.Builder
		largeDiff.WriteString("diff --git a/file.txt b/file.txt\n")
		for i := 0; i < 50000; i++ { // Much larger to ensure chunking
			largeDiff.WriteString(fmt.Sprintf("+line %d with some additional content to make it longer\n", i))
		}

		// Mock chunk processing - need to match the exact number of chunks
		mockProvider.On("GenerateCommitMessage", mock.Anything, "types", mock.Anything).Return("chunk result", nil).Maybe()
		// Mock merge processing
		mockProvider.On("GenerateCommitMessage", mock.Anything, "types", mock.Anything).Return("merged result", nil).Maybe()

		result, err := processor.ProcessChunkedDiff(largeDiff.String(), "types", "context")

		assert.NoError(t, err)
		// The result should be either "chunk result" (fallback) or "merged result" (successful merge)
		assert.True(t, result == "chunk result" || result == "merged result", "Unexpected result: %s", result)
		mockProvider.AssertExpectations(t)
	})

	t.Run("handles processing errors with large diff", func(t *testing.T) {
		mockProvider := &MockAIProvider{}
		processor := NewChunkedDiffProcessor(mockProvider)

		// Create a large diff
		var largeDiff strings.Builder
		largeDiff.WriteString("diff --git a/file.txt b/file.txt\n")
		for i := 0; i < 50000; i++ { // Much larger to ensure summarization
			largeDiff.WriteString(fmt.Sprintf("+line %d with some additional content to make it longer\n", i))
		}

		// Mock processing with error
		mockProvider.On("GenerateCommitMessage", mock.Anything, "types", mock.Anything).Return("", assert.AnError).Maybe()

		result, err := processor.ProcessChunkedDiff(largeDiff.String(), "types", "context")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "general error for testing")
		assert.Empty(t, result)
	})
}

func TestChunkedDiffProcessor_ProcessChunkedDetailedCommit(t *testing.T) {
	t.Run("processes detailed commit in chunks", func(t *testing.T) {
		mockProvider := &MockAIProvider{}
		processor := NewChunkedDiffProcessor(mockProvider)

		// Create a large diff
		var largeDiff strings.Builder
		largeDiff.WriteString("diff --git a/file.txt b/file.txt\n")
		for i := 0; i < 2000; i++ {
			largeDiff.WriteString(fmt.Sprintf("+line %d\n", i))
		}

		// Mock chunk processing
		mockProvider.On("GenerateCommitMessage", mock.Anything, "types", mock.Anything).Return("chunk result", nil).Maybe()
		// Mock merge processing
		expectedResult := &CommitResult{
			Message: "feat: merged changes",
			Body:    "• Multiple changes merged",
		}
		mockProvider.On("GenerateDetailedCommit", mock.Anything, "types", mock.Anything).Return(expectedResult, nil).Maybe()

		result, err := processor.ProcessChunkedDetailedCommit(largeDiff.String(), "types", "context")

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockProvider.AssertExpectations(t)
	})
}
