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

var phindAPIURL = "https://https.extension.phind.com/agent/"

const defaultPhindModel = "Phind-70B"

func SetPhindAPIURL(url string) {
	phindAPIURL = url
}

type PhindConfig struct {
	Model string
}

type PhindProvider struct {
	config PhindConfig
	client *http.Client
}

func NewPhindProvider(model string) *PhindProvider {
	if model == "" {
		model = defaultPhindModel
	}
	return &PhindProvider{
		config: PhindConfig{Model: model},
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *PhindProvider) GenerateCommitMessage(diff string, commitTypes string, extraContext string) (string, error) {
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
`, commitTypes)

	if extraContext != "" {
		userPrompt += fmt.Sprintf("\nAdditional context: %s\n", extraContext)
	}

	userPrompt += `Focus on being accurate and concise.
Commit message must be a maximum of 72 characters.
Exclude anything unnecessary such as translation.
Your entire response will be passed directly into git commit.
Code diff:`

	userPrompt += fmt.Sprintf("\n```diff\n%s\n```", diff)

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

	req, err := http.NewRequest("POST", phindAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "Identity")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Phind: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("phind API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var fullContent strings.Builder
	lines := strings.Split(string(bodyBytes), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var responseJson map[string]interface{}
			if err := json.Unmarshal([]byte(data), &responseJson); err != nil {
				continue
			}
			if choices, ok := responseJson["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fullContent.WriteString(content)
						}
					}
				}
			}
		}
	}
	if fullContent.Len() == 0 {
		return "", fmt.Errorf("no content found in Phind response")
	}

	rawResult := fullContent.String() // Get the full content before line-by-line processing

	// Iterate through lines to find the *first* valid commit message line.
	// If no such line is found, we should return an empty string or an appropriate error/message.
	for _, line := range strings.Split(rawResult, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "```") && !strings.HasPrefix(trimmedLine, "#") {
			return trimmedLine, nil // Found a valid line, return it
		}
	}

	// If the loop finishes and no valid line was found, return an empty string
	// because the content was only whitespace, comments, or code blocks.
	return "", nil
}

func (p *PhindProvider) GetModel() string {
	return p.config.Model
}
