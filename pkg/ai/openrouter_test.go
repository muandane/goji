package ai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testOpenRouterAPIKey = "test_openrouter_api_key"

func TestNewOpenRouterProvider(t *testing.T) {
	t.Run("default model", func(t *testing.T) {
		provider := NewOpenRouterProvider(testOpenRouterAPIKey, "")
		assert.NotNil(t, provider)
		assert.Equal(t, defaultOpenRouterModel, provider.model)
		assert.Equal(t, testOpenRouterAPIKey, provider.apiKey)
		assert.NotNil(t, provider.client)
	})

	t.Run("custom model", func(t *testing.T) {
		customModel := "custom/model"
		provider := NewOpenRouterProvider(testOpenRouterAPIKey, customModel)
		assert.NotNil(t, provider)
		assert.Equal(t, customModel, provider.model)
	})
}

func TestOpenRouterProvider_GetModel(t *testing.T) {
	customModel := "OpenRouter-Test-Model"
	provider := NewOpenRouterProvider(testOpenRouterAPIKey, customModel)
	assert.Equal(t, customModel, provider.GetModel())
}

func TestOpenRouterProvider_GenerateCommitMessage(t *testing.T) {
	commitTypesJSON := `{"feat":"New feature","fix":"Bug fix"}`
	diffSample := "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n@@ -1 +1 @@\n-old\n+new"

	// Store original and defer restoration if you plan to use a global test server URL variable
	// originalURL := openRouterAPIURL
	// defer func() { openRouterAPIURL = originalURL }()

	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		extraContext   string
		apiKeyEnv      string // To simulate API key presence
		expectedMsg    string
		expectErr      bool
		expectedErrStr string
	}{
		{
			name: "successful response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer "+testOpenRouterAPIKey, r.Header.Get("Authorization"))
				assert.Equal(t, httpReferer, r.Header.Get("HTTP-Referer"))
				assert.Equal(t, xTitle, r.Header.Get("X-Title"))

				var reqPayload OpenRouterRequest
				err := json.NewDecoder(r.Body).Decode(&reqPayload)
				require.NoError(t, err)
				assert.Contains(t, reqPayload.Messages[1].Content, diffSample)

				resp := OpenRouterResponse{
					Choices: []OpenRouterResponseChoice{
						{Message: OpenRouterMessage{Content: "feat: add new feature"}},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			apiKeyEnv:   testOpenRouterAPIKey,
			expectedMsg: "feat: add new feature",
		},
		{
			name: "successful response with context",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var reqPayload OpenRouterRequest
				err := json.NewDecoder(r.Body).Decode(&reqPayload)
				require.NoError(t, err)
				assert.Contains(t, reqPayload.Messages[1].Content, "Test Context")

				resp := OpenRouterResponse{
					Choices: []OpenRouterResponseChoice{
						{Message: OpenRouterMessage{Content: "docs(readme): update with context"}},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			apiKeyEnv:    testOpenRouterAPIKey,
			extraContext: "Test Context",
			expectedMsg:  "docs(readme): update with context",
		},
		{
			name: "API error 401 unauthorized",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				errResp := OpenRouterErrorResponse{
					Error: OpenRouterErrorDetail{Message: "Invalid API key"},
				}
				json.NewEncoder(w).Encode(errResp)
			},
			apiKeyEnv:      testOpenRouterAPIKey,
			expectErr:      true,
			expectedErrStr: "openrouter API error (status 401): Invalid API key",
		},
		{
			name: "API error 500 internal server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "Something went wrong")
			},
			apiKeyEnv:      testOpenRouterAPIKey,
			expectErr:      true,
			expectedErrStr: "openrouter API returned status 500: Something went wrong",
		},
		{
			name: "no choices in response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := OpenRouterResponse{Choices: []OpenRouterResponseChoice{}}
				json.NewEncoder(w).Encode(resp)
			},
			apiKeyEnv:      testOpenRouterAPIKey,
			expectErr:      true,
			expectedErrStr: "no content found in OpenRouter response or choices array is empty",
		},
		{
			name: "message content needs cleanup",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := OpenRouterResponse{
					Choices: []OpenRouterResponseChoice{
						{Message: OpenRouterMessage{Content: "  `fix(parser): resolve issue`  \n#comment\n"}},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			apiKeyEnv:   testOpenRouterAPIKey,
			expectedMsg: "fix(parser): resolve issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			// Temporarily override the global constant for this test
			originalURL := openRouterAPIURL
			openRouterAPIURL = server.URL // Point to test server
			defer func() { openRouterAPIURL = originalURL }()

			// Simulate API key presence (passed to constructor, not via env for NewOpenRouterProvider)
			provider := NewOpenRouterProvider(tt.apiKeyEnv, defaultOpenRouterModel)
			// The provider's client will hit the httptest.Server URL due to the override above.
			// For more robust testing, you might inject the server.Client() into the provider.
			// However, for this setup, changing the global openRouterAPIURL for the test scope works.

			msg, err := provider.GenerateCommitMessage(diffSample, commitTypesJSON, tt.extraContext)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedErrStr != "" {
					assert.Contains(t, err.Error(), tt.expectedErrStr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, msg)
			}
		})
	}
}
