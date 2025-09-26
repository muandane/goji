package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var openRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
var httpReferer = "https://github.com/muandane/goji"
var xTitle = "Goji"

type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterRequest struct {
	Model    string              `json:"model"`
	Messages []OpenRouterMessage `json:"messages"`
}

type OpenRouterResponseChoice struct {
	Index   int               `json:"index"`
	Message OpenRouterMessage `json:"message"`
}

type OpenRouterResponse struct {
	ID      string                     `json:"id,omitempty"`
	Model   string                     `json:"model,omitempty"`
	Choices []OpenRouterResponseChoice `json:"choices"`
}

type OpenRouterErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

type OpenRouterErrorResponse struct {
	Error OpenRouterErrorDetail `json:"error"`
}

type OpenRouterProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenRouterProvider(apiKey, modelFromConfig string) *OpenRouterProvider {
	chosenModel := modelFromConfig

	if chosenModel == "" {
		envModel := os.Getenv("OPENROUTER_MODEL")
		if envModel != "" {
			chosenModel = envModel
		} else {
			chosenModel = "meta-llama/llama-4-maverick:free"
		}
	}
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  chosenModel,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *OpenRouterProvider) GenerateCommitMessage(diff string, commitTypesJSON string, extraContext string) (string, error) {
	// Validate and enhance the diff
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

	systemPrompt := `You are a commit message generator that follows these rules:
		1. Write in present tense
		2. Be concise and direct
		3. Output only the commit message without any explanations
		4. Follow the format: <type>(<optional scope>): <commit message>
		5. Keep the message under 72 characters
		6. Focus on what changed, not how it changed`

	userPrompt := fmt.Sprintf(`Generate a concise git commit message written in present tense for the following code diff with the given specifications below:

The output response must be in format:
<type>(<optional scope>): <commit message>

Choose a type from the type-to-description JSON below that best describes the git diff:
%s
`, commitTypesJSON)

	if extraContext != "" {
		userPrompt += fmt.Sprintf("\nAdditional context: %s\n", extraContext)
	}

	userPrompt += `Focus on being accurate and concise.
Commit message must be a maximum of 72 characters.
Exclude anything unnecessary such as translation.
Your entire response will be passed directly into git commit.
Code diff:`
	userPrompt += fmt.Sprintf("\n```diff\n%s\n```", diff)

	requestPayload := OpenRouterRequest{
		Model: p.model,
		Messages: []OpenRouterMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenRouter request payload: %w", err)
	}

	req, err := http.NewRequest("POST", openRouterAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create OpenRouter request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", httpReferer)
	req.Header.Set("X-Title", xTitle)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read OpenRouter response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse OpenRouterErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return "", fmt.Errorf("openrouter API error (status %d): %s", resp.StatusCode, errorResponse.Error.Message)
		}
		return "", fmt.Errorf("openrouter API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var successResponse OpenRouterResponse
	if err := json.Unmarshal(bodyBytes, &successResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal OpenRouter success response: %w. Body: %s", err, string(bodyBytes))
	}

	if len(successResponse.Choices) == 0 || successResponse.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no content found in OpenRouter response or choices array is empty. Body: %s", string(bodyBytes))
	}

	rawResult := successResponse.Choices[0].Message.Content

	// Improved response parsing - try multiple strategies
	commitMessage := extractCommitMessage(rawResult)
	if commitMessage != "" {
		return commitMessage, nil
	}

	// If no valid commit message found, return error with context
	return "", fmt.Errorf("no valid commit message found in response. Raw response: %s", rawResult)
}

func (p *OpenRouterProvider) GenerateDetailedCommit(diff string, commitTypesJSON string, extraContext string) (*CommitResult, error) {
	// Validate and enhance the diff
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

	userPrompt := fmt.Sprintf(`Generate a commit message with detailed body for the following code diff:

Follow the exact format specified. Include both a concise title and detailed bullet points in the body.

Choose a type from the type-to-description JSON below that best describes the git diff:
%s
`, commitTypesJSON)

	if extraContext != "" {
		userPrompt += fmt.Sprintf("\nAdditional context: %s\n", extraContext)
	}

	userPrompt += `Focus on being accurate and comprehensive.
Code diff:`
	userPrompt += fmt.Sprintf("\n```diff\n%s\n```", diff)

	requestPayload := OpenRouterRequest{
		Model: p.model,
		Messages: []OpenRouterMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenRouter request payload: %w", err)
	}

	req, err := http.NewRequest("POST", openRouterAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", httpReferer)
	req.Header.Set("X-Title", xTitle)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenRouter response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse OpenRouterErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return nil, fmt.Errorf("openrouter API error (status %d): %s", resp.StatusCode, errorResponse.Error.Message)
		}
		return nil, fmt.Errorf("openrouter API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var successResponse OpenRouterResponse
	if err := json.Unmarshal(bodyBytes, &successResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenRouter success response: %w. Body: %s", err, string(bodyBytes))
	}

	if len(successResponse.Choices) == 0 || successResponse.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("no content found in OpenRouter response or choices array is empty. Body: %s", string(bodyBytes))
	}

	rawResult := successResponse.Choices[0].Message.Content

	// Parse the detailed commit message to extract title and body
	result := parseDetailedCommitMessage(rawResult)
	if result.Message == "" {
		return nil, fmt.Errorf("no valid commit message found in response. Raw response: %s", rawResult)
	}

	return result, nil
}

func (p *OpenRouterProvider) GetModel() string {
	return p.model
}
