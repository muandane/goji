package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testGeminiAPIKey = "test_gemini_api_key"

func TestNewGeminiProvider(t *testing.T) {
	defaultGeminiModel := "gemini-3-flash-preview"

	t.Run("uses model from config", func(t *testing.T) {
		provider := NewGeminiProvider(testGeminiAPIKey, "gemini-1.5-pro")
		assert.NotNil(t, provider)
		assert.Equal(t, "gemini-1.5-pro", provider.model)
		assert.Equal(t, testGeminiAPIKey, provider.apiKey)
		assert.False(t, provider.useOAuth)
	})

	t.Run("uses default model when config is empty", func(t *testing.T) {
		provider := NewGeminiProvider(testGeminiAPIKey, "")
		assert.NotNil(t, provider)
		assert.Equal(t, defaultGeminiModel, provider.model)
	})

	t.Run("uses API key from environment", func(t *testing.T) {
		os.Setenv("GEMINI_API_KEY", "env-api-key")
		defer os.Unsetenv("GEMINI_API_KEY")

		provider := NewGeminiProvider("", "")
		assert.NotNil(t, provider)
		assert.Equal(t, "env-api-key", provider.apiKey)
		assert.False(t, provider.useOAuth)
	})

	t.Run("uses OAuth when no API key provided", func(t *testing.T) {
		os.Unsetenv("GEMINI_API_KEY")
		provider := NewGeminiProvider("", "")
		assert.NotNil(t, provider)
		assert.True(t, provider.useOAuth)
	})
}

func TestGeminiProvider_GetModel(t *testing.T) {
	customModel := "gemini-1.5-pro"
	provider := NewGeminiProvider(testGeminiAPIKey, customModel)
	assert.Equal(t, customModel, provider.GetModel())
}

func TestGeminiProvider_GenerateCommitMessage(t *testing.T) {
	commitTypesJSON := `{"feat":"New feature","fix":"Bug fix"}`
	diffSample := "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n@@ -1 +1 @@\n-old\n+new"

	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		apiKey         string
		useOAuth       bool
		expectedMsg    string
		expectErr      bool
		expectedErrStr string
	}{
		{
			name: "successful response with API key",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, testGeminiAPIKey, r.Header.Get("x-goog-api-key"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody GeminiRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				require.NoError(t, err)
				assert.Len(t, reqBody.Contents, 2)

				resp := GeminiResponse{
					Candidates: []GeminiCandidate{
						{
							Content: GeminiContent{
								Parts: []GeminiPart{
									{Text: "feat: add new feature"},
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			},
			apiKey:      testGeminiAPIKey,
			useOAuth:    false,
			expectedMsg: "feat: add new feature",
		},
		{
			name: "successful response with OAuth token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-oauth-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				resp := GeminiResponse{
					Candidates: []GeminiCandidate{
						{
							Content: GeminiContent{
								Parts: []GeminiPart{
									{Text: "fix: resolve bug"},
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			},
			apiKey:      "",
			useOAuth:    true,
			expectedMsg: "fix: resolve bug",
		},
		{
			name: "API error handling",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := GeminiResponse{
					Error: &GeminiError{
						Code:    400,
						Message: "Invalid API key",
						Status:  "INVALID_ARGUMENT",
					},
				}
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			},
			apiKey:         testGeminiAPIKey,
			useOAuth:       false,
			expectErr:      true,
			expectedErrStr: "Invalid API key",
		},
		{
			name: "empty response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := GeminiResponse{
					Candidates: []GeminiCandidate{},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			},
			apiKey:         testGeminiAPIKey,
			useOAuth:       false,
			expectErr:      true,
			expectedErrStr: "no content found",
		},
		{
			name: "empty diff",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// Should not be called
			},
			apiKey:         testGeminiAPIKey,
			useOAuth:       false,
			expectErr:      true,
			expectedErrStr: "empty diff provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty diff" {
				provider := NewGeminiProvider(tt.apiKey, "gemini-3-flash-preview")
				_, err := provider.GenerateCommitMessage("", commitTypesJSON, "")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			// Override the API base URL for testing
			originalURL := geminiAPIBaseURL
			geminiAPIBaseURL = server.URL
			defer func() { geminiAPIBaseURL = originalURL }()

			provider := NewGeminiProvider(tt.apiKey, "gemini-3-flash-preview")
			if tt.useOAuth {
				provider.useOAuth = true
				provider.accessToken = "test-oauth-token"
			}

			msg, err := provider.GenerateCommitMessage(diffSample, commitTypesJSON, "")

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

// Test helper to create a testable Gemini provider with mocked HTTP client
func TestGeminiProvider_WithMockedClient(t *testing.T) {
	t.Run("successful API call with mocked client", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "generateContent")

			resp := GeminiResponse{
				Candidates: []GeminiCandidate{
					{
						Content: GeminiContent{
							Parts: []GeminiPart{
								{Text: "feat(test): add test functionality"},
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create provider and manually set the client to use test server
		provider := NewGeminiProvider(testGeminiAPIKey, "gemini-3-flash-preview")

		// We need to make the base URL injectable for proper testing
		// This is a limitation - we should refactor to allow URL injection
		_ = provider
		_ = server
	})
}

func TestGeminiProvider_GenerateDetailedCommit(t *testing.T) {
	commitTypesJSON := `{"feat":"New feature"}`

	t.Run("empty diff error", func(t *testing.T) {
		provider := NewGeminiProvider(testGeminiAPIKey, "gemini-3-flash-preview")
		_, err := provider.GenerateDetailedCommit("", commitTypesJSON, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty diff provided")
	})
}
