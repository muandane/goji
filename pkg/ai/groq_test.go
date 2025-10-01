package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGroqProvider(t *testing.T) {
	t.Run("uses model from config", func(t *testing.T) {
		provider := NewGroqProvider("test-api-key", "llama-3-70b-8192")
		if provider.model != "llama-3-70b-8192" {
			t.Errorf("expected model llama-3-70b-8192, got %s", provider.model)
		}
	})

	t.Run("uses default model when config is empty", func(t *testing.T) {
		provider := NewGroqProvider("test-api-key", "")
		if provider.model != "mixtral-8x7b-32768" {
			t.Errorf("expected default model mixtral-8x7b-32768, got %s", provider.model)
		}
	})
}

func TestGroqProvider_GenerateCommitMessage(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		mockResponse := GroqResponse{
			ID:    "test-id",
			Model: "mixtral-8x7b-32768",
			Choices: []GroqResponseChoice{
				{
					Index: 0,
					Message: GroqMessage{
						Role:    "assistant",
						Content: "feat(auth): add user authentication",
					},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer test-api-key" {
				t.Errorf("expected Authorization header 'Bearer test-api-key', got %s", authHeader)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		groqAPIURL = server.URL
		provider := NewGroqProvider("test-api-key", "mixtral-8x7b-32768")

		diff := `diff --git a/auth.go b/auth.go
+func Authenticate() {}
`
		commitTypes := `{"feat": "A new feature"}`

		result, err := provider.GenerateCommitMessage(diff, commitTypes, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result != "feat(auth): add user authentication" {
			t.Errorf("expected 'feat(auth): add user authentication', got %s", result)
		}
	})

	t.Run("API error handling", func(t *testing.T) {
		errorResponse := GroqErrorResponse{
			Error: GroqErrorDetail{
				Message: "Invalid API key",
				Type:    "invalid_request_error",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(errorResponse)
		}))
		defer server.Close()

		groqAPIURL = server.URL
		provider := NewGroqProvider("invalid-key", "mixtral-8x7b-32768")

		diff := `diff --git a/test.go b/test.go
+func Test() {}
`
		commitTypes := `{"feat": "A new feature"}`

		_, err := provider.GenerateCommitMessage(diff, commitTypes, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		expectedError := "groq API error (status 401): Invalid API key"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("empty diff error", func(t *testing.T) {
		provider := NewGroqProvider("test-api-key", "mixtral-8x7b-32768")

		_, err := provider.GenerateCommitMessage("", "{}", "")
		if err == nil {
			t.Fatal("expected error for empty diff, got nil")
		}

		expectedError := "empty diff provided"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestGroqProvider_GenerateDetailedCommit(t *testing.T) {
	t.Run("successful detailed generation", func(t *testing.T) {
		mockResponse := GroqResponse{
			ID:    "test-id",
			Model: "mixtral-8x7b-32768",
			Choices: []GroqResponseChoice{
				{
					Index: 0,
					Message: GroqMessage{
						Role: "assistant",
						Content: `Title: feat(auth): add user authentication

Body:
• Implement JWT token generation
• Add login endpoint with validation
• Create user authentication middleware`,
					},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		groqAPIURL = server.URL
		provider := NewGroqProvider("test-api-key", "mixtral-8x7b-32768")

		diff := `diff --git a/auth.go b/auth.go
+func Authenticate() {}
+func GenerateToken() {}
`
		commitTypes := `{"feat": "A new feature"}`

		result, err := provider.GenerateDetailedCommit(diff, commitTypes, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Message != "feat(auth): add user authentication" {
			t.Errorf("expected message 'feat(auth): add user authentication', got '%s'", result.Message)
		}

		if result.Body == "" {
			t.Error("expected non-empty body")
		}
	})
}

func TestGroqProvider_GetModel(t *testing.T) {
	provider := NewGroqProvider("test-api-key", "llama-3-70b-8192")
	if provider.GetModel() != "llama-3-70b-8192" {
		t.Errorf("expected model llama-3-70b-8192, got %s", provider.GetModel())
	}
}
