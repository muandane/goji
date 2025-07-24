package ai

import (
	"strings"
)

const maxDiffSize = 50000 // 50KB limit for diffs
const minDiffSize = 10    // Minimum diff size to be meaningful

// truncateDiff truncates a diff to fit within API limits while preserving context
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
