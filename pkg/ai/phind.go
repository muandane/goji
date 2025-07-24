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
	defer resp.Body.Close()

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

func (p *PhindProvider) GetModel() string {
	return p.config.Model
}
