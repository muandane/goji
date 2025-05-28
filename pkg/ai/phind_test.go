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
		assert.Equal(t, defaultPhindModel, provider.config.Model)
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
				assert.Equal(t, defaultPhindModel, payload["requested_model"])

				// This is the key: The server's response must match the expected output.
				// Change the mock response to match your expectedMsg from the test case.
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"feat(test): implement new API final bit."}}]}`,
				}
				for _, line := range responseLines {
					fmt.Fprintln(w, line)
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
				assert.Contains(t, userInput, "Additional context: testing context")

				// Change the mock response to match your expectedMsg from the test case.
				responseLines := []string{
					`data: {"choices":[{"delta":{"content":"fix(auth): resolve context issue"}}]}`,
				}
				for _, line := range responseLines {
					fmt.Fprintln(w, line)
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
				fmt.Fprintln(w, `data: {"choices":[{"delta":{}}]}`)
				fmt.Fprintln(w, `data: {}`)
			},
			expectErr:      true,
			expectedErrMsg: "no content found in Phind response",
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
					fmt.Fprintln(w, line)
				}
			},
			expectedMsg: "",
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
				w.Write([]byte("partial content"))
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
			// Save original API URL and restore it after the test
			originalPhindAPIURL := phindAPIURL
			defer func() { phindAPIURL = originalPhindAPIURL }()

			if tt.malformedURL {
				// Set a malformed URL to trigger the http.NewRequest error path
				SetPhindAPIURL("http://%ghjk")
			} else {
				// Create a test server and set the provider's client to use it
				server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
				// Defer closing the server before the GenerateCommitMessage call
				// but after the test's execution for the server handler.
				defer server.Close()

				// Set the PhindAPIURL to the test server's URL
				SetPhindAPIURL(server.URL + "/agent/") // Ensure it matches the constant's path
			}

			providerForTest := NewPhindProvider(defaultPhindModel)
			// Only set the client if not testing malformed URL, as client isn't used then
			if !tt.malformedURL {
				// Create a test server and set the provider's client to use it
				server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
				// Defer closing the server before the GenerateCommitMessage call
				// but after the test's execution for the server handler.
				defer server.Close()

				// Set the PhindAPIURL to the test server's URL
				SetPhindAPIURL(server.URL + "/agent/") // Ensure it matches the constant's path

				providerForTest := NewPhindProvider(defaultPhindModel)
				providerForTest.client = &http.Client{
					Timeout:   5 * time.Second, // Shorter timeout for tests
					Transport: server.Client().Transport,
				}
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
