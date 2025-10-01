package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PhindConfig struct {
	Model      string
	APIBaseURL string
}

func NewPhindConfig(model string) PhindConfig {
	if model == "" {
		model = "Phind-70B"
	}
	return PhindConfig{
		Model:      model,
		APIBaseURL: "https://https.extension.phind.com/agent/",
	}
}

type PhindProvider struct {
	client *http.Client
	config PhindConfig
}

func NewPhindProvider(model string) *PhindProvider {
	return &PhindProvider{
		client: &http.Client{Timeout: 30 * time.Second},
		config: NewPhindConfig(model),
	}
}

func createPhindHeaders() http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("User-Agent", "")
	headers.Set("Accept", "*/*")
	headers.Set("Accept-Encoding", "Identity")
	return headers
}

func parsePhindLine(line string) (string, bool) {
	const prefix = "data: "
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	data := strings.TrimPrefix(line, prefix)
	var jsonValue map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonValue); err != nil {
		return "", false
	}
	choices, ok := jsonValue["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", false
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", false
	}
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return "", false
	}
	content, ok := delta["content"].(string)
	return content, ok
}

func parsePhindStreamResponse(responseText string) string {
	var builder strings.Builder
	for _, line := range strings.Split(responseText, "\n") {
		if content, ok := parsePhindLine(line); ok {
			builder.WriteString(content)
		}
	}
	return builder.String()
}

func (p *PhindProvider) GenerateCommitMessage(diff, commitTypes, extraContext string) (string, error) {
	if diff == "" {
		return "", fmt.Errorf("empty diff provided")
	}

	// Use smart diff summarization for better AI comprehension
	diffSummary, optimizedDiff := summarizeDiff(diff)

	// Add summary info to context if diff was significantly reduced
	if diffSummary.SummarySize < diffSummary.OriginalSize {
		if extraContext != "" {
			extraContext += " "
		}
		extraContext += fmt.Sprintf("(Diff summarized: %d files changed. %s)",
			len(diffSummary.FilesChanged),
			diffSummary.Summary)
	}

	// Use the optimized diff
	diff = optimizedDiff

	// Still apply small diff enhancement if needed
	diff = enhanceSmallDiff(diff)

	systemPrompt := `You are a commit message generator. Your ONLY job is to generate a conventional commit message.

RULES:
1. Write in present tense
2. Be concise and direct
3. Output ONLY the commit message without any explanations, quotes, or markdown
4. Follow the format: <type>(<optional scope>): <commit message>
5. Keep the message under 72 characters
6. Focus on what changed, not how it changed

VALID TYPES: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert

EXAMPLE OUTPUT: feat(auth): add user authentication system`

	context := ""
	if extraContext != "" {
		context = fmt.Sprintf(`
Use the following context to understand intent:
%s`, extraContext)
	}

	userPrompt := fmt.Sprintf("Generate a concise git commit message written in present tense for the following code diff with the given specifications below:\n\n"+
		"The output response must be in format:\n"+
		"<type>(<optional scope>): <commit message>\n\n"+
		"Choose a type from the type-to-description JSON below that best describes the git diff:\n"+
		"%s\n"+
		"Focus on being accurate and concise.%s\n"+
		"Commit message must be a maximum of 72 characters.\n"+
		"Exclude anything unnecessary such as translation. Your entire response will be passed directly into git commit.\n\n"+
		"Code diff:\n"+
		"```diff\n"+
		"%s\n"+
		"```", commitTypes, context, diff)

	payload := map[string]interface{}{
		"additional_extension_context": "",
		"allow_magic_buttons":          true,
		"is_vscode_extension":          true,
		"message_history": []map[string]interface{}{
			{
				"content": systemPrompt,
				"role":    "system",
			},
			{
				"content": userPrompt,
				"role":    "user",
			},
		},
		"requested_model": p.config.Model,
		"user_input":      userPrompt,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.APIBaseURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = createPhindHeaders()

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Phind: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error message from JSON
		var errorJSON struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(bodyBytes, &errorJSON); err == nil && errorJSON.Error.Message != "" {
			return "", fmt.Errorf("phind API error (status %d): %s", resp.StatusCode, errorJSON.Error.Message)
		}
		return "", fmt.Errorf("phind API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	fullText := parsePhindStreamResponse(string(bodyBytes))
	if strings.TrimSpace(fullText) == "" {
		return "", fmt.Errorf("no completion choice in Phind response")
	}

	return fullText, nil
}

func (p *PhindProvider) GenerateDetailedCommit(diff, commitTypes, extraContext string) (*CommitResult, error) {
	if diff == "" {
		return nil, fmt.Errorf("empty diff provided")
	}

	// Use smart diff summarization for better AI comprehension
	diffSummary, optimizedDiff := summarizeDiff(diff)

	// Add summary info to context if diff was significantly reduced
	if diffSummary.SummarySize < diffSummary.OriginalSize {
		if extraContext != "" {
			extraContext += " "
		}
		extraContext += fmt.Sprintf("(Diff summarized: %d files changed. %s)",
			len(diffSummary.FilesChanged),
			diffSummary.Summary)
	}

	// Use the optimized diff
	diff = optimizedDiff
	diff = enhanceSmallDiff(diff)

	systemPrompt := `You are a commit message generator that generates both a concise title and a detailed body.

RULES FOR TITLE:
1. Write in present tense
2. Be concise and direct
3. Follow the format: <type>(<optional scope>): <commit message>
4. Keep the title under 72 characters
5. Focus on what changed, not how it changed

RULES FOR BODY:
1. Write in present tense
2. Provide 2-4 bullet points that explain the changes in depth
3. Each bullet point should be concise but informative
4. Focus on the WHY and WHAT of the changes
5. Reference specific files or functionality when relevant

VALID TYPES: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert

OUTPUT FORMAT:
Title: <type>(<optional scope>): <commit message>

Body:
• First detailed point about the changes
• Second detailed point about the changes  
• Third detailed point about the changes (if applicable)

EXAMPLE OUTPUT:
Title: feat(auth): add user authentication system

Body:
• Add JWT token generation and validation middleware
• Implement login/logout endpoints with secure password hashing
• Create user model with role-based access control
• Add authentication middleware to protect API routes`

	context := ""
	if extraContext != "" {
		context = fmt.Sprintf(`
Use the following context to understand intent:
%s`, extraContext)
	}

	userPrompt := fmt.Sprintf("Generate a commit message with detailed body for the following code diff:\n\n"+
		"Follow the exact format specified. Include both a concise title and detailed bullet points in the body.\n\n"+
		"Choose a type from the type-to-description JSON below that best describes the git diff:\n"+
		"%s\n"+
		"Focus on being accurate and comprehensive.%s\n\n"+
		"Code diff:\n"+
		"```diff\n"+
		"%s\n"+
		"```", commitTypes, context, diff)

	payload := map[string]interface{}{
		"additional_extension_context": "",
		"allow_magic_buttons":          true,
		"is_vscode_extension":          true,
		"message_history": []map[string]interface{}{
			{
				"content": systemPrompt,
				"role":    "system",
			},
			{
				"content": userPrompt,
				"role":    "user",
			},
		},
		"requested_model": p.config.Model,
		"user_input":      userPrompt,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.APIBaseURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = createPhindHeaders()

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Phind: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorJSON struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(bodyBytes, &errorJSON); err == nil && errorJSON.Error.Message != "" {
			return nil, fmt.Errorf("phind API error (status %d): %s", resp.StatusCode, errorJSON.Error.Message)
		}
		return nil, fmt.Errorf("phind API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	fullText := parsePhindStreamResponse(string(bodyBytes))
	if strings.TrimSpace(fullText) == "" {
		return nil, fmt.Errorf("no completion choice in Phind response")
	}

	// Parse the detailed commit message to extract title and body
	result := parseDetailedCommitMessage(fullText)
	return result, nil
}

func (p *PhindProvider) GetModel() string {
	return p.config.Model
}
