package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	httpReferer                       = "https://github.com/muandane/goji"
	xTitle                            = "Goji CLI"
	openRouterDefaultModelForProvider = "mistralai/devstral-small:free"
)

var (
	openRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
)

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
			chosenModel = openRouterDefaultModelForProvider
		}
	}
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  chosenModel,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *OpenRouterProvider) GenerateCommitMessage(diff string, commitTypesJSON string, extraContext string) (string, error) {
	systemPrompt := `You are a commit message generator that follows these rules:
		1. Write in present tense
		2. Be concise and direct
		3. Output only the commit message without any explanations
		4. Follow the format: <type>(<optional scope>): <commit message>`

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

	for _, line := range strings.Split(rawResult, "\n") {
		trimmedLine := strings.TrimSpace(line)

		trimmedLine = strings.TrimPrefix(trimmedLine, "```")
		trimmedLine = strings.TrimSuffix(trimmedLine, "```")
		trimmedLine = strings.TrimPrefix(trimmedLine, "`")
		trimmedLine = strings.TrimSuffix(trimmedLine, "`")
		trimmedLine = strings.TrimSpace(trimmedLine)

		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
			return trimmedLine, nil
		}
	}

	if strings.TrimSpace(rawResult) != "" {
		return strings.TrimSpace(rawResult), nil
	}

	return "", fmt.Errorf("no usable commit message content found in OpenRouter response after cleanup. Original: %s", rawResult)
}

func (p *OpenRouterProvider) GetModel() string {
	return p.model
}
