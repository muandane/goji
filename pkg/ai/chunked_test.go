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

func TestChunkedDiffProcessor_SplitDiffIntoChunks(t *testing.T) {
	processor := &ChunkedDiffProcessor{}

	t.Run("small diff returns single chunk", func(t *testing.T) {
		smallDiff := "diff --git a/file.txt b/file.txt\n+line1\n-line2"
		chunks := processor.splitDiffIntoChunks(smallDiff)

		assert.Len(t, chunks, 1)
		assert.Equal(t, smallDiff+"\n", chunks[0])
	})

	t.Run("large diff splits into multiple chunks", func(t *testing.T) {
		// Create a large diff that will definitely exceed maxChunkSize
		var largeDiff strings.Builder
		for i := 0; i < 10000; i++ {
			largeDiff.WriteString(fmt.Sprintf("+line %d with some additional content to make it longer\n", i))
		}

		chunks := processor.splitDiffIntoChunks(largeDiff.String())

		assert.Greater(t, len(chunks), 1)

		// Verify each chunk is within size limits
		for i, chunk := range chunks {
			assert.LessOrEqual(t, len(chunk), maxChunkSize+100, "chunk %d exceeds max size", i)
		}
	})

	t.Run("empty diff returns empty chunks", func(t *testing.T) {
		chunks := processor.splitDiffIntoChunks("")
		assert.Len(t, chunks, 1) // Empty string still creates one chunk with newline
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

	t.Run("handles chunk processing errors", func(t *testing.T) {
		mockProvider := &MockAIProvider{}
		processor := NewChunkedDiffProcessor(mockProvider)

		// Create a large diff
		var largeDiff strings.Builder
		largeDiff.WriteString("diff --git a/file.txt b/file.txt\n")
		for i := 0; i < 50000; i++ { // Much larger to ensure chunking
			largeDiff.WriteString(fmt.Sprintf("+line %d with some additional content to make it longer\n", i))
		}

		// Mock chunk processing with error
		mockProvider.On("GenerateCommitMessage", mock.Anything, "types", mock.Anything).Return("", assert.AnError).Maybe()

		result, err := processor.ProcessChunkedDiff(largeDiff.String(), "types", "context")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chunk processing errors")
		assert.Empty(t, result)
	})
}

func TestChunkedDiffProcessor_ExtractFilesFromChunk(t *testing.T) {
	processor := &ChunkedDiffProcessor{}

	t.Run("extracts files from diff chunk", func(t *testing.T) {
		chunk := `diff --git a/src/main.go b/src/main.go
+func main() {
+    fmt.Println("hello")
+}
diff --git a/README.md b/README.md
+# Project`

		files := processor.extractFilesFromChunk(chunk)

		assert.Len(t, files, 2)
		assert.Contains(t, files, "src/main.go")
		assert.Contains(t, files, "README.md")
	})

	t.Run("handles chunk with no files", func(t *testing.T) {
		chunk := "+line1\n-line2"

		files := processor.extractFilesFromChunk(chunk)

		assert.Len(t, files, 0)
	})
}

func TestChunkedDiffProcessor_CreateMergePrompt(t *testing.T) {
	processor := &ChunkedDiffProcessor{}

	t.Run("creates merge prompt", func(t *testing.T) {
		summaries := []string{"feat: add feature", "fix: fix bug"}
		files := []string{"file1.go", "file2.go"}

		prompt := processor.createMergePrompt(summaries, files, "types", "context")

		assert.Contains(t, prompt, "Merge the following commit message summaries")
		assert.Contains(t, prompt, "Summary 1: feat: add feature")
		assert.Contains(t, prompt, "Summary 2: fix: fix bug")
		assert.Contains(t, prompt, "Files affected: file1.go, file2.go")
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
