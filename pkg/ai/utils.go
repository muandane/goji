package ai

import (
	"fmt"
	"strings"
)

const maxDiffSize = 50000 // 50KB limit for diffs
const minDiffSize = 10    // Minimum diff size to be meaningful

// DiffSummary represents a summarized version of a git diff
type DiffSummary struct {
	FilesChanged []FileChange
	Summary      string
	OriginalSize int
	SummarySize  int
}

// FileChange represents changes in a single file
type FileChange struct {
	Path      string
	Delta     string // "+" or "-" for added/removed lines count
	Changes   []ChangeLine
	IsNewFile bool
	IsDeleted bool
	IsBinary  bool
}

// ChangeLine represents a specific line change
type ChangeLine struct {
	LineNum int
	Content string
	Type    string // "+", "-", or "=" for unchanged (kept for context)
	Context string // surrounding context if needed
}

// summarizeDiff creates a smart summary of the diff for the AI model
func summarizeDiff(diff string) (*DiffSummary, string) {
	if diff == "" {
		return &DiffSummary{}, ""
	}

	lines := strings.Split(diff, "\n")
	var fileChanges []FileChange
	var currentFile string
	var currentLines []ChangeLine

	// Parse the diff to extract meaningful changes
	for i, line := range lines {
		// Track file boundaries
		if strings.HasPrefix(line, "diff --git") {
			if currentFile != "" && len(currentLines) > 0 {
				fileChanges = append(fileChanges, createFileChange(currentFile, currentLines))
				currentLines = nil
			}
			// Extract filename from diff --git a/path b/path
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				aFile := strings.TrimPrefix(parts[2], "a/")
				bFile := strings.TrimPrefix(parts[3], "b/")
				if aFile == bFile {
					currentFile = aFile
				}
			}
		} else if strings.HasPrefix(line, "new file mode") {
			// File is new
		} else if strings.HasPrefix(line, "deleted file mode") {
			// File is deleted
		} else if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			// Track the actual changes
			changeLine := ChangeLine{
				LineNum: i,
				Content: line,
				Type:    string(line[0]),
			}
			currentLines = append(currentLines, changeLine)
		}
	}

	// Add the last file if any
	if currentFile != "" && len(currentLines) > 0 {
		fileChanges = append(fileChanges, createFileChange(currentFile, currentLines))
	}

	// Generate a smart summary
	summary := generateDiffSummary(fileChanges, len(diff))

	// If the original diff is small enough, use it directly
	if len(diff) <= maxDiffSize {
		return summary, diff
	}

	// Create an optimized diff that keeps only relevant changes
	optimizedDiff := createOptimizedDiff(diff, fileChanges)
	return summary, optimizedDiff
}

// createOptimizedDiff creates a smaller version of the diff with only relevant changes
func createOptimizedDiff(originalDiff string, fileChanges []FileChange) string {
	if len(originalDiff) <= maxDiffSize {
		return originalDiff
	}

	var optimizedLines []string
	lines := strings.Split(originalDiff, "\n")

	// Always include git diff headers
	for i := 0; i < len(lines) && i < 10; i++ {
		if strings.HasPrefix(lines[i], "diff --git") ||
			strings.HasPrefix(lines[i], "index ") ||
			strings.HasPrefix(lines[i], "---") ||
			strings.HasPrefix(lines[i], "+++") ||
			strings.HasPrefix(lines[i], "@@") {
			optimizedLines = append(optimizedLines, lines[i])
		}
	}

	// Add key changes for each file - focus on the most impactful ones
	for _, fileChange := range fileChanges {
		// Limit changes per file to keep under size limit
		maxChangesPerFile := maxDiffSize / (len(fileChanges) + 1)
		changesToKeep := fileChange.Changes
		if len(changesToKeep) > maxChangesPerFile {
			changesToKeep = changesToKeep[:maxChangesPerFile]
		}

		if len(changesToKeep) > 0 {
			optimizedLines = append(optimizedLines, "")
			optimizedLines = append(optimizedLines, "# Key changes in "+fileChange.Path+":")
			for _, change := range changesToKeep {
				optimizedLines = append(optimizedLines, change.Content)
			}
		}

		if len(optimizedLines) > maxDiffSize {
			break
		}
	}

	result := strings.Join(optimizedLines, "\n")
	if len(result) > maxDiffSize {
		result = result[:maxDiffSize] + "\n... (diff optimized for size)"
	}

	return result
}

func createFileChange(filepath string, changes []ChangeLine) FileChange {
	var delta string
	addCount := 0
	delCount := 0

	for _, change := range changes {
		if change.Type == "+" {
			addCount++
		} else if change.Type == "-" {
			delCount++
		}
	}

	if addCount > delCount {
		delta = fmt.Sprintf("+%d", addCount-delCount)
	} else if delCount > addCount {
		delta = fmt.Sprintf("-%d", delCount-addCount)
	} else {
		delta = "="
	}

	return FileChange{
		Path:    filepath,
		Delta:   delta,
		Changes: changes,
	}
}

func generateDiffSummary(fileChanges []FileChange, originalSize int) *DiffSummary {
	summary := &DiffSummary{
		FilesChanged: fileChanges,
		OriginalSize: originalSize,
		SummarySize:  len(fileChanges) * 50, // Rough estimate
	}

	// Create a human-readable summary
	var summaryParts []string
	if len(fileChanges) == 1 {
		change := fileChanges[0]
		summaryParts = append(summaryParts, fmt.Sprintf("Modified: %s (%s)", change.Path, change.Delta))
	} else {
		summaryParts = append(summaryParts, fmt.Sprintf("%d files changed", len(fileChanges)))

		// List changed files
		for _, change := range fileChanges {
			summaryParts = append(summaryParts, fmt.Sprintf("  %s (%s)", change.Path, change.Delta))
		}
	}

	summary.Summary = strings.Join(summaryParts, "\n")
	return summary
}

// truncateDiff truncates a diff to fit within API limits while preserving context
// DEPRECATED: Use summarizeDiff instead for better intelligence
func truncateDiff(diff string) string {
	if len(diff) <= maxDiffSize {
		return diff
	}

	// Try to keep the most relevant parts: file headers and recent changes
	lines := strings.Split(diff, "\n")
	var truncatedLines []string

	// Always include the first few lines (git diff header)
	headerLines := 10
	if len(lines) < headerLines {
		headerLines = len(lines)
	}
	truncatedLines = append(truncatedLines, lines[:headerLines]...)

	// Include the last portion of the diff (most recent changes)
	remainingSpace := maxDiffSize - len(strings.Join(truncatedLines, "\n"))
	if remainingSpace > 1000 { // Ensure we have enough space for meaningful content
		// Calculate how many lines from the end we can include
		linesFromEnd := remainingSpace / 100 // Rough estimate of chars per line
		if linesFromEnd > len(lines)-headerLines {
			linesFromEnd = len(lines) - headerLines
		}
		if linesFromEnd > 0 {
			truncatedLines = append(truncatedLines, "... (diff truncated)")
			truncatedLines = append(truncatedLines, lines[len(lines)-linesFromEnd:]...)
		}
	}

	truncatedDiff := strings.Join(truncatedLines, "\n")
	if len(truncatedDiff) > maxDiffSize {
		// If still too large, just truncate at maxDiffSize
		return truncatedDiff[:maxDiffSize] + "\n... (diff truncated)"
	}

	return truncatedDiff
}

// enhanceSmallDiff adds context to very small diffs to make them more meaningful
func enhanceSmallDiff(diff string) string {
	if len(diff) >= minDiffSize {
		return diff
	}

	// For very small diffs, add some context about what this might be
	enhancedDiff := diff + "\n\n# Note: This is a very small change. Please focus on the specific modification shown above."
	return enhancedDiff
}

// extractCommitMessage tries multiple strategies to extract a valid commit message
func extractCommitMessage(rawResult string) string {
	lines := strings.Split(rawResult, "\n")

	// Strategy 1: Look for conventional commit format
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// Remove markdown formatting
		trimmedLine = strings.TrimPrefix(trimmedLine, "```")
		trimmedLine = strings.TrimSuffix(trimmedLine, "```")
		trimmedLine = strings.TrimPrefix(trimmedLine, "`")
		trimmedLine = strings.TrimSuffix(trimmedLine, "`")
		trimmedLine = strings.TrimSpace(trimmedLine)

		// Skip comments and empty lines
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Check if it looks like a conventional commit
		if isValidCommitMessage(trimmedLine) {
			return trimmedLine
		}
	}

	// Strategy 2: If no conventional format found, return the first non-empty line
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "```") {
			return trimmedLine
		}
	}

	// Strategy 3: Return the entire response if it's not too long and contains valid content
	trimmedResult := strings.TrimSpace(rawResult)
	if len(trimmedResult) <= 72 && trimmedResult != "" {
		// Check if it contains any non-comment, non-whitespace content
		lines := strings.Split(trimmedResult, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "```") {
				return trimmedResult
			}
		}
	}

	return ""
}

// parseDetailedCommitMessage parses a response that contains both title and body
func parseDetailedCommitMessage(response string) *CommitResult {
	lines := strings.Split(response, "\n")
	var title, body string
	var inBody bool
	var bodyLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for title line
		if strings.HasPrefix(trimmedLine, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Title:"))
			continue
		}

		// Check for body section start
		if trimmedLine == "Body:" || trimmedLine == "body:" {
			inBody = true
			continue
		}

		// Collect body lines
		if inBody && trimmedLine != "" {
			bodyLines = append(bodyLines, trimmedLine)
		}
	}

	if title == "" {
		// If we didn't find a structured format, try to parse from the original format
		// and just use the commit message as title, no body
		title = strings.TrimSpace(response)
		body = ""
	} else {
		body = strings.Join(bodyLines, "\n")
	}

	return &CommitResult{
		Message: title,
		Body:    body,
	}
}

// isValidCommitMessage checks if a string looks like a valid conventional commit message
func isValidCommitMessage(message string) bool {
	// Basic conventional commit pattern: type(scope): description
	// This is a simplified check - the actual regex could be more sophisticated
	if len(message) > 72 {
		return false
	}

	// Check for basic conventional commit structure
	parts := strings.SplitN(message, ":", 2)
	if len(parts) != 2 {
		return false
	}

	// Check the type/scope part
	typeScope := strings.TrimSpace(parts[0])
	if typeScope == "" {
		return false
	}

	// Check the description part
	description := strings.TrimSpace(parts[1])
	if description == "" {
		return false
	}

	// Basic type validation (could be expanded)
	validTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore", "perf", "ci", "build", "revert"}
	typePart := typeScope
	if strings.Contains(typeScope, "(") {
		typePart = strings.Split(typeScope, "(")[0]
	}

	for _, validType := range validTypes {
		if typePart == validType {
			return true
		}
	}

	// If it doesn't match known types but has the right structure, accept it
	return true
}
