package ai

import (
	"fmt"
	"strings"
)

const (
	// Token estimation: ~4 characters per token for English text
	// Use summarization for diffs larger than 20k characters (≈5k tokens)
	maxChunkSize = 20000
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

// ProcessChunkedDiff processes a large diff using smart summarization instead of chunking
func (c *ChunkedDiffProcessor) ProcessChunkedDiff(diff, commitTypes, extraContext string) (string, error) {
	if diff == "" {
		return "", fmt.Errorf("empty diff provided")
	}

	// If diff is small enough, process normally
	if len(diff) <= maxChunkSize {
		return c.provider.GenerateCommitMessage(diff, commitTypes, extraContext)
	}

	// For large diffs, use aggressive summarization
	fmt.Printf("🔍 Large diff detected: %d chars, using aggressive summarization\n", len(diff))

	// Create a very aggressive summary
	aggressiveSummary := c.createAggressiveSummary(diff)
	fmt.Printf("📊 Summarized to %d chars (%.1f%% reduction)\n", len(aggressiveSummary),
		float64(len(diff)-len(aggressiveSummary))/float64(len(diff))*100)

	// Add summary info to context
	summaryContext := "(Large diff summarized to key changes only)"
	if extraContext != "" {
		extraContext += " " + summaryContext
	} else {
		extraContext = summaryContext
	}

	// Process the aggressively summarized diff
	return c.provider.GenerateCommitMessage(aggressiveSummary, commitTypes, extraContext)
}

// ProcessChunkedDetailedCommit processes a large diff using smart summarization instead of chunking
func (c *ChunkedDiffProcessor) ProcessChunkedDetailedCommit(diff, commitTypes, extraContext string) (*CommitResult, error) {
	if diff == "" {
		return nil, fmt.Errorf("empty diff provided")
	}

	// If diff is small enough, process normally
	if len(diff) <= maxChunkSize {
		return c.provider.GenerateDetailedCommit(diff, commitTypes, extraContext)
	}

	// For large diffs, use aggressive summarization
	fmt.Printf("🔍 Large diff detected: %d chars, using aggressive summarization\n", len(diff))

	// Create a very aggressive summary
	aggressiveSummary := c.createAggressiveSummary(diff)
	fmt.Printf("📊 Summarized to %d chars (%.1f%% reduction)\n", len(aggressiveSummary),
		float64(len(diff)-len(aggressiveSummary))/float64(len(diff))*100)

	// Add summary info to context
	summaryContext := "(Large diff summarized to key changes only)"
	if extraContext != "" {
		extraContext += " " + summaryContext
	} else {
		extraContext = summaryContext
	}

	// Process the aggressively summarized diff
	return c.provider.GenerateDetailedCommit(aggressiveSummary, commitTypes, extraContext)
}

// createAggressiveSummary creates a very aggressive summary of the diff
func (c *ChunkedDiffProcessor) createAggressiveSummary(diff string) string {
	lines := strings.Split(diff, "\n")
	var summaryLines []string
	var currentFile string
	var addedLines, removedLines int
	var fileChanges []string

	for _, line := range lines {
		// Track file boundaries
		if strings.HasPrefix(line, "diff --git") {
			// Save previous file summary
			if currentFile != "" && (addedLines > 0 || removedLines > 0) {
				change := fmt.Sprintf("%s: +%d/-%d", currentFile, addedLines, removedLines)
				fileChanges = append(fileChanges, change)
			}

			// Extract new filename
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentFile = strings.TrimPrefix(parts[2], "a/")
				addedLines, removedLines = 0, 0
			}
			continue
		}

		// Count changes
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removedLines++
		}
	}

	// Add the last file
	if currentFile != "" && (addedLines > 0 || removedLines > 0) {
		change := fmt.Sprintf("%s: +%d/-%d", currentFile, addedLines, removedLines)
		fileChanges = append(fileChanges, change)
	}

	// Create a very concise summary
	summaryLines = append(summaryLines, "Summary of changes:")
	for _, change := range fileChanges {
		summaryLines = append(summaryLines, "  "+change)
	}

	// Add a few key lines from the actual diff (first 10 lines of changes)
	changeCount := 0
	for _, line := range lines {
		if (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")) &&
			!strings.HasPrefix(line, "+++") && !strings.HasPrefix(line, "---") {
			summaryLines = append(summaryLines, line)
			changeCount++
			if changeCount >= 10 { // Limit to 10 key changes
				break
			}
		}
	}

	return strings.Join(summaryLines, "\n")
}
