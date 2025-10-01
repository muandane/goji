package ai

import (
	"fmt"
	"strings"
)

const (
	// Token estimation: ~4 characters per token for English text
	// 6k tokens ≈ 24k characters
	maxChunkSize = 24000
	// Minimum chunk size to avoid too many small requests
	minChunkSize = 1000
)

// ChunkedDiffProcessor handles chunked diff processing for large diffs
type ChunkedDiffProcessor struct {
	provider AIProvider
}

// NewChunkedDiffProcessor creates a new chunked diff processor
func NewChunkedDiffProcessor(provider AIProvider) *ChunkedDiffProcessor {
	return &ChunkedDiffProcessor{
		provider: provider,
	}
}

// ChunkResult represents the result of processing a single chunk
type ChunkResult struct {
	ChunkIndex int
	Summary    string
	Files      []string
	Error      error
}

// ProcessChunkedDiff processes a large diff in chunks and merges results
func (c *ChunkedDiffProcessor) ProcessChunkedDiff(diff, commitTypes, extraContext string) (string, error) {
	if diff == "" {
		return "", fmt.Errorf("empty diff provided")
	}

	// If diff is small enough, process normally
	if len(diff) <= maxChunkSize {
		return c.provider.GenerateCommitMessage(diff, commitTypes, extraContext)
	}

	// Split diff into chunks
	chunks := c.splitDiffIntoChunks(diff)
	if len(chunks) == 0 {
		return "", fmt.Errorf("failed to create diff chunks")
	}

	// Process each chunk
	chunkResults := make([]ChunkResult, len(chunks))
	for i, chunk := range chunks {
		chunkResults[i] = c.processChunk(i, chunk, commitTypes, extraContext)
	}

	// Merge results
	return c.mergeChunkResults(chunkResults, commitTypes, extraContext)
}

// ProcessChunkedDetailedCommit processes a large diff in chunks for detailed commits
func (c *ChunkedDiffProcessor) ProcessChunkedDetailedCommit(diff, commitTypes, extraContext string) (*CommitResult, error) {
	if diff == "" {
		return nil, fmt.Errorf("empty diff provided")
	}

	// If diff is small enough, process normally
	if len(diff) <= maxChunkSize {
		return c.provider.GenerateDetailedCommit(diff, commitTypes, extraContext)
	}

	// Split diff into chunks
	chunks := c.splitDiffIntoChunks(diff)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("failed to create diff chunks")
	}

	// Process each chunk
	chunkResults := make([]ChunkResult, len(chunks))
	for i, chunk := range chunks {
		chunkResults[i] = c.processChunk(i, chunk, commitTypes, extraContext)
	}

	// Merge results for detailed commit
	return c.mergeDetailedChunkResults(chunkResults, commitTypes, extraContext)
}

// splitDiffIntoChunks splits a diff into manageable chunks
func (c *ChunkedDiffProcessor) splitDiffIntoChunks(diff string) []string {
	lines := strings.Split(diff, "\n")
	var chunks []string
	var currentChunk strings.Builder
	var currentSize int

	for _, line := range lines {
		lineSize := len(line) + 1 // +1 for newline

		// If adding this line would exceed max size, start a new chunk
		if currentSize+lineSize > maxChunkSize && currentSize > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentSize = 0
		}

		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
		currentSize += lineSize
	}

	// Add the last chunk if it has content
	if currentSize > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// processChunk processes a single chunk
func (c *ChunkedDiffProcessor) processChunk(index int, chunk, commitTypes, extraContext string) ChunkResult {
	// Add chunk context
	chunkContext := fmt.Sprintf("(Processing chunk %d of multiple chunks)", index+1)
	if extraContext != "" {
		extraContext += " " + chunkContext
	} else {
		extraContext = chunkContext
	}

	// Process the chunk
	summary, err := c.provider.GenerateCommitMessage(chunk, commitTypes, extraContext)
	if err != nil {
		return ChunkResult{
			ChunkIndex: index,
			Error:      err,
		}
	}

	// Extract files from chunk
	files := c.extractFilesFromChunk(chunk)

	return ChunkResult{
		ChunkIndex: index,
		Summary:    summary,
		Files:      files,
		Error:      nil,
	}
}

// extractFilesFromChunk extracts file names from a diff chunk
func (c *ChunkedDiffProcessor) extractFilesFromChunk(chunk string) []string {
	var files []string
	lines := strings.Split(chunk, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				file := strings.TrimPrefix(parts[2], "a/")
				files = append(files, file)
			}
		}
	}

	return files
}

// mergeChunkResults merges results from multiple chunks into a single commit message
func (c *ChunkedDiffProcessor) mergeChunkResults(results []ChunkResult, commitTypes, extraContext string) (string, error) {
	// Check for errors
	var errors []string
	var summaries []string
	var allFiles []string

	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("chunk %d: %v", result.ChunkIndex, result.Error))
		} else {
			summaries = append(summaries, result.Summary)
			allFiles = append(allFiles, result.Files...)
		}
	}

	if len(errors) > 0 {
		return "", fmt.Errorf("chunk processing errors: %s", strings.Join(errors, "; "))
	}

	if len(summaries) == 0 {
		return "", fmt.Errorf("no valid chunk results")
	}

	// If only one chunk succeeded, return its result
	if len(summaries) == 1 {
		return summaries[0], nil
	}

	// If we have multiple chunks, we need to merge them
	if len(summaries) > 1 {
		// Create a merge prompt
		mergePrompt := c.createMergePrompt(summaries, allFiles, commitTypes, extraContext)

		// Use the provider to merge the results
		mergedResult, err := c.provider.GenerateCommitMessage(mergePrompt, commitTypes, extraContext)
		if err != nil {
			// Fallback: return the first valid summary
			return summaries[0], nil
		}

		return mergedResult, nil
	}

	return "", fmt.Errorf("unexpected state: no valid summaries")
}

// mergeDetailedChunkResults merges results for detailed commits
func (c *ChunkedDiffProcessor) mergeDetailedChunkResults(results []ChunkResult, commitTypes, extraContext string) (*CommitResult, error) {
	// Check for errors
	var errors []string
	var summaries []string
	var allFiles []string

	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("chunk %d: %v", result.ChunkIndex, result.Error))
		} else {
			summaries = append(summaries, result.Summary)
			allFiles = append(allFiles, result.Files...)
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("chunk processing errors: %s", strings.Join(errors, "; "))
	}

	if len(summaries) == 0 {
		return nil, fmt.Errorf("no valid chunk results")
	}

	// If only one chunk succeeded, return its result
	if len(summaries) == 1 {
		// Parse the single result as detailed commit
		return parseDetailedCommitMessage(summaries[0]), nil
	}

	// Create a merge prompt for detailed commit
	mergePrompt := c.createDetailedMergePrompt(summaries, allFiles, commitTypes, extraContext)

	// Use the provider to merge the results
	mergedResult, err := c.provider.GenerateDetailedCommit(mergePrompt, commitTypes, extraContext)
	if err != nil {
		// Fallback: parse the first valid summary
		return parseDetailedCommitMessage(summaries[0]), nil
	}

	return mergedResult, nil
}

// createMergePrompt creates a prompt for merging chunk results
func (c *ChunkedDiffProcessor) createMergePrompt(summaries []string, files []string, commitTypes, extraContext string) string {
	var prompt strings.Builder

	prompt.WriteString("Merge the following commit message summaries into a single, coherent commit message:\n\n")

	for i, summary := range summaries {
		prompt.WriteString(fmt.Sprintf("Summary %d: %s\n", i+1, summary))
	}

	if len(files) > 0 {
		prompt.WriteString(fmt.Sprintf("\nFiles affected: %s\n", strings.Join(files, ", ")))
	}

	prompt.WriteString("\nGenerate a single commit message that captures the essence of all changes.")

	return prompt.String()
}

// createDetailedMergePrompt creates a prompt for merging detailed commit results
func (c *ChunkedDiffProcessor) createDetailedMergePrompt(summaries []string, files []string, commitTypes, extraContext string) string {
	var prompt strings.Builder

	prompt.WriteString("Merge the following detailed commit summaries into a single, coherent commit message with title and body:\n\n")

	for i, summary := range summaries {
		prompt.WriteString(fmt.Sprintf("Summary %d:\n%s\n", i+1, summary))
	}

	if len(files) > 0 {
		prompt.WriteString(fmt.Sprintf("\nFiles affected: %s\n", strings.Join(files, ", ")))
	}

	prompt.WriteString("\nGenerate a single commit message with title and body that captures the essence of all changes.")

	return prompt.String()
}
