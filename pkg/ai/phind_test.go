package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time" // Add time import for test server handler

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhindProvider(t *testing.T) {
	t.Run("default model", func(t *testing.T) {
		provider := NewPhindProvider("")
		assert.NotNil(t, provider)
		assert.Equal(t, "Phind-70B", provider.config.Model)
		assert.NotNil(t, provider.client)
	})

	t.Run("custom model", func(t *testing.T) {
		customModel := "Phind-Custom"
		provider := NewPhindProvider(customModel)
		assert.NotNil(t, provider)
		assert.Equal(t, customModel, provider.config.Model)
		assert.NotNil(t, provider.client)
	})
}

func TestPhindProvider_GetModel(t *testing.T) {
	customModel := "Phind-Test-Model"
	provider := NewPhindProvider(customModel)
	assert.Equal(t, customModel, provider.GetModel())
}

func TestPhindProvider_GenerateCommitMessage(t *testing.T) {
	commitTypesJSON := `{"feat":"Feature","fix":"Bug fix"}`
	diffSample := "diff --git a/main.go b/main.go\nindex 123..456 100644\n--- a/main.go\n+++ b/main.go\n@@ -1,1 +1,1 @@\n-fmt.Println(\"Hello\")\n+fmt.Println(\"Hi\")"

	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		extraContext   string
		expectedMsg    string
		expectErr      bool
		expectedErrMsg string
		// Add a flag to test `http.NewRequest` error directly, which is hard to mock
		// so we'll simulate by setting a malformed URL for the test.
		malformedURL bool
	}{
		{
			name: "successful response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				bodyBytes, _ := io.ReadAll(r.Body)
				var payload map[string]interface{}
				err := json.Unmarshal(bodyBytes, &payload)
				require.NoError(t, err)
				assert.Equal(t, "Phind-70B", payload["requested_model"])

				// This is the key: The server's response must match the expected output.
				// Change the mock response to match your expectedMsg from the test case.
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"feat(test): implement new API final bit."}}]}`,
				}
				for _, line := range responseLines {
					_, _ = fmt.Fprintln(w, line)
				}
			},
			expectedMsg: "feat(test): implement new API final bit.",
			expectErr:   false,
		},
		{
			name: "successful response with extra context in prompt",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, _ := io.ReadAll(r.Body)
				var payload map[string]interface{}
				err := json.Unmarshal(bodyBytes, &payload)
				require.NoError(t, err)

				userInput, ok := payload["user_input"].(string)
				require.True(t, ok)
				assert.Contains(t, userInput, "Use the following context to understand intent:\ntesting context")

				// Change the mock response to match your expectedMsg from the test case.
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"fix(auth): resolve context issue"}}]}`,
				}
				for _, line := range responseLines {
					_, _ = fmt.Fprintln(w, line)
				}
			},
			extraContext: "testing context",
			expectedMsg:  "fix(auth): resolve context issue",
			expectErr:    false,
		},
		{
			name: "API returns non-200 status",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
			expectErr:      true,
			expectedErrMsg: "phind API returned status 500: internal server error",
		},
		{
			name: "empty content in response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, `data: {"choices":[{"delta":{}}]}`)
				_, _ = fmt.Fprintln(w, `data: {}`)
			},
			expectErr:      true,
			expectedErrMsg: "no completion choice in Phind response",
		},
		{
			name: "response is just whitespace or comments",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"   "}}]}`,
					`data: {"choices":[{"delta":{"content":"\n# comment line"}}]}`,
					`data: {"choices":[{"delta":{"content":"\n"}}]}`,
				}
				for _, line := range responseLines {
					_, _ = fmt.Fprintln(w, line)
				}
			},
			expectedMsg: "   \n# comment line\n",
			expectErr:   false,
		},
		{
			name: "failed to read response body",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// This handler simulates a broken pipe or connection closure
				// to trigger an error during io.ReadAll(resp.Body).
				// We cannot use http.Hijacker directly with httptest.NewServer's default recorder.
				// Instead, we can write some data and then cause a delay/timeout on the server side
				// or write an incomplete response.
				w.Header().Set("Content-Length", "100") // Claim a length, but don't send all of it.
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("partial content"))
				// The client will try to read 100 bytes, but only gets "partial content".
				// This should cause an EOF or similar error on read.
			},
			expectErr:      true,
			expectedErrMsg: "failed to read response body", // The actual error might be EOF or unexpected EOF
		},
		{
			name: "failed to create request (malformed URL)",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// This handler won't even be called because http.NewRequest will fail first.
			},
			expectErr:      true,
			expectedErrMsg: "failed to create request",
			malformedURL:   true, // Trigger this scenario
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.malformedURL {
				// Create a provider with a malformed URL to trigger the http.NewRequest error path
				providerForTest := &PhindProvider{
					client: &http.Client{Timeout: 5 * time.Second},
					config: PhindConfig{
						Model:      "Phind-70B",
						APIBaseURL: "http://%ghjk", // Malformed URL
					},
				}
				_, err := providerForTest.GenerateCommitMessage(diffSample, commitTypesJSON, tt.extraContext)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}

			// Create a test server and set the provider's client to use it
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			providerForTest := &PhindProvider{
				client: &http.Client{
					Timeout:   5 * time.Second, // Shorter timeout for tests
					Transport: server.Client().Transport,
				},
				config: PhindConfig{
					Model:      "Phind-70B",
					APIBaseURL: server.URL + "/agent/",
				},
			}

			msg, err := providerForTest.GenerateCommitMessage(diffSample, commitTypesJSON, tt.extraContext)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, msg)
			}
		})
	}
}

func TestPhindProvider_GenerateDetailedCommit(t *testing.T) {
	commitTypesJSON := `{"feat":"Feature","fix":"Bug fix"}`
	diffSample := "diff --git a/main.go b/main.go\nindex 123..456 100644\n--- a/main.go\n+++ b/main.go\n@@ -1,1 +1,1 @@\n-fmt.Println(\"Hello\")\n+fmt.Println(\"Hi\")"

	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		extraContext   string
		diff           string
		expectedMsg    string
		expectedBody   string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name: "successful detailed commit response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				
				bodyBytes, _ := io.ReadAll(r.Body)
				var payload map[string]interface{}
				err := json.Unmarshal(bodyBytes, &payload)
				require.NoError(t, err)
				
				// Check that system prompt contains detailed commit instructions
				messageHistory := payload["message_history"].([]interface{})
				systemMsg := messageHistory[0].(map[string]interface{})
				assert.Contains(t, systemMsg["content"].(string), "detailed body")
				
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"Title: feat(test): implement new API"}}]}`,
					`data: {"choices":[{"delta":{"content":"\n\nBody:\n• First point\n• Second point"}}]}`,
				}
				for _, line := range responseLines {
					_, _ = fmt.Fprintln(w, line)
				}
			},
			diff:        diffSample,
			expectedMsg: "feat(test): implement new API",
			expectedBody: "• First point\n• Second point",
			expectErr:   false,
		},
		{
			name: "empty diff error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// Handler shouldn't be called
			},
			diff:           "",
			expectErr:      true,
			expectedErrMsg: "empty diff provided",
		},
		{
			name: "API returns non-200 status",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
			diff:           diffSample,
			expectErr:      true,
			expectedErrMsg: "phind API returned status 500",
		},
		{
			name: "empty content in response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, `data: {"choices":[{"delta":{}}]}`)
			},
			diff:           diffSample,
			expectErr:      true,
			expectedErrMsg: "no completion choice in Phind response",
		},
		{
			name: "detailed commit with extra context",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, _ := io.ReadAll(r.Body)
				var payload map[string]interface{}
				err := json.Unmarshal(bodyBytes, &payload)
				require.NoError(t, err)
				
				messageHistory := payload["message_history"].([]interface{})
				userMsg := messageHistory[1].(map[string]interface{})
				assert.Contains(t, userMsg["content"].(string), "testing context")
				
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"Title: fix(auth): resolve context issue\n\nBody:\n• Point about context"}}]}`,
				}
				for _, line := range responseLines {
					_, _ = fmt.Fprintln(w, line)
				}
			},
			diff:         diffSample,
			extraContext: "testing context",
			expectedMsg:  "fix(auth): resolve context issue",
			expectedBody: "• Point about context",
			expectErr:    false,
		},
		{
			name: "unstructured response format",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"feat(test): simple commit message"}}]}`,
				}
				for _, line := range responseLines {
					_, _ = fmt.Fprintln(w, line)
				}
			},
			diff:        diffSample,
			expectedMsg: "feat(test): simple commit message",
			expectedBody: "",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.diff == "" {
				// Test empty diff case
				provider := NewPhindProvider("")
				_, err := provider.GenerateDetailedCommit(tt.diff, commitTypesJSON, tt.extraContext)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			providerForTest := &PhindProvider{
				client: &http.Client{
					Timeout:   5 * time.Second,
					Transport: server.Client().Transport,
				},
				config: PhindConfig{
					Model:      "Phind-70B",
					APIBaseURL: server.URL + "/agent/",
				},
			}

			result, err := providerForTest.GenerateDetailedCommit(tt.diff, commitTypesJSON, tt.extraContext)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedMsg, result.Message)
				assert.Equal(t, tt.expectedBody, result.Body)
			}
		})
	}
}
