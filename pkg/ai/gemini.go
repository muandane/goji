package ai

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	geminiOAuthScope = "https://www.googleapis.com/auth/generative-language.retriever"

	// Default OAuth client ID for goji (desktop app)
	// This should be created in Google Cloud Console for the goji project
	// Users can override with GOOGLE_CLIENT_ID environment variable
	// Note: Using PKCE flow, so no client secret is required
	defaultGeminiClientID = "YOUR_GOOGLE_CLIENT_ID_HERE.apps.googleusercontent.com"
)

func getGeminiOAuthConfig() (*oauth2.Config, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	// Use default if not provided (for project-wide client ID)
	if clientID == "" {
		// TODO: Replace with actual client ID created for goji project
		// For now, users need to set GOOGLE_CLIENT_ID or use API key
		// Once a client ID is created for goji, it can be embedded here
		return nil, fmt.Errorf("OAuth client ID not configured")
	}

	// PKCE flow doesn't require client secret for desktop apps
	// Google allows this for "Desktop app" type OAuth clients
	return &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: "http://localhost:8080/oauth2callback",
		Scopes:      []string{geminiOAuthScope},
		Endpoint:    google.Endpoint,
	}, nil
}

// generatePKCEVerifier generates a code verifier for PKCE flow
func generatePKCEVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type GeminiProvider struct {
	accessToken string
	apiKey      string
	model       string
	client      *http.Client
	useOAuth    bool
}

func NewGeminiProvider(apiKey, modelFromConfig string) *GeminiProvider {
	chosenModel := modelFromConfig
	if chosenModel == "" {
		envModel := os.Getenv("GEMINI_MODEL")
		if envModel != "" {
			chosenModel = envModel
		} else {
			chosenModel = "gemini-3-flash-preview"
		}
	}

	// If no API key provided, try to get from environment
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}

	useOAuth := apiKey == ""
	if useOAuth {
		// Try to load token from cache first (fastest path)
		token, err := loadGeminiToken()
		if err == nil && token.Valid() {
			return &GeminiProvider{
				accessToken: token.AccessToken,
				model:       chosenModel,
				client:      &http.Client{Timeout: 20 * time.Second}, // Reduced timeout for faster failure
				useOAuth:    true,
			}
		}
	}

	return &GeminiProvider{
		apiKey:   apiKey,
		model:    chosenModel,
		client:   &http.Client{Timeout: 20 * time.Second}, // Reduced timeout for faster failure
		useOAuth: useOAuth,
	}
}

func (p *GeminiProvider) ensureAuthenticated() error {
	if !p.useOAuth {
		if p.apiKey == "" {
			return fmt.Errorf(`GEMINI_API_KEY not set. Choose an authentication method:

1. Login with Google (OAuth - Recommended, like Gemini CLI):
   Set GOOGLE_CLIENT_ID environment variable (no secret needed).
   Get client ID from: https://console.cloud.google.com/apis/credentials
   Create OAuth 2.0 Client ID (Desktop app type)
   Then run goji draft again - a browser will open for login.

2. Use API Key:
   Set GEMINI_API_KEY environment variable.
   Get key from: https://makersuite.google.com/app/apikey`)
		}
		return nil // Using API key, no auth needed
	}

	if p.accessToken != "" {
		// Token exists, assume it's valid (will be validated on API call)
		return nil
	}

	// Need to authenticate via OAuth
	fmt.Println("\n🔐 Authentication required for Gemini")
	token, err := authenticateGemini()
	if err != nil {
		return fmt.Errorf("Gemini authentication failed: %w", err)
	}

	p.accessToken = token.AccessToken
	if err := saveGeminiToken(token); err != nil {
		// Non-fatal: token is still valid for this session
		fmt.Printf("⚠️  Warning: failed to save token: %v\n", err)
	}
	return nil
}

func authenticateGemini() (*oauth2.Token, error) {
	ctx := context.Background()

	// First, try to load cached token
	token, err := loadGeminiToken()
	if err == nil && token.Valid() {
		return token, nil
	}

	// Try OAuth flow with browser-based authentication (like Gemini CLI)
	config, err := getGeminiOAuthConfig()
	if err == nil {
		return authenticateGeminiOAuthFlowWithConfig(ctx, config)
	}

	// Return error with instructions
	return nil, fmt.Errorf(`Gemini authentication required. Choose one:

1. Login with Google (OAuth - Recommended, like Gemini CLI):
   • Get OAuth client ID from: https://console.cloud.google.com/apis/credentials
   • Create OAuth 2.0 Client ID (Desktop app type)
   • Set GOOGLE_CLIENT_ID environment variable (no secret needed with PKCE)
   • Run goji draft again - a browser will open for Google login
   • Note: Once goji has a project-wide client ID, this step won't be needed

2. Use API Key (Simpler):
   • Get API key from: https://makersuite.google.com/app/apikey
   • Set GEMINI_API_KEY environment variable`)
}

// authenticateGeminiOAuthFlowWithConfig performs browser-based OAuth flow
// Similar to Gemini CLI's "Login with Google" feature

func authenticateGeminiOAuthFlowWithConfig(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	fmt.Println("🔐 Starting Google authentication...")
	fmt.Println("   Opening browser for login...")

	// Generate PKCE verifier (challenge is computed automatically by oauth2.S256ChallengeOption)
	verifier, err := generatePKCEVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}

	// Start local server for OAuth callback
	server := startLocalOAuthServer(verifier)
	defer server.Close()

	// Generate auth URL with state for CSRF protection and PKCE
	state := generateState()
	authURL := config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
		oauth2.S256ChallengeOption(verifier),
	)

	// Open browser automatically
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("\n⚠️  Could not open browser automatically.\n")
		fmt.Printf("   Please visit this URL in your browser:\n\n")
		fmt.Printf("   %s\n\n", authURL)
		fmt.Println("   Waiting for authentication...")
	} else {
		fmt.Println("   Browser opened. Please complete the login...")
	}

	// Wait for callback with timeout
	select {
	case code := <-server.codeChan:
		if code == "" {
			return nil, fmt.Errorf("authentication cancelled or failed")
		}

		// Verify state matches (CSRF protection)
		receivedState := <-server.stateChan
		if receivedState != state {
			return nil, fmt.Errorf("state mismatch - possible CSRF attack")
		}

		// Exchange code for token using PKCE verifier
		token, err := config.Exchange(
			ctx,
			code,
			oauth2.VerifierOption(verifier),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to exchange code for token: %w", err)
		}

		// Save token
		if err := saveGeminiToken(token); err != nil {
			fmt.Printf("⚠️  Warning: failed to save token: %v\n", err)
		} else {
			fmt.Println("✅ Authentication successful! Token saved.")
		}

		return token, nil
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timeout - please try again")
	}
}

func (p *GeminiProvider) GenerateCommitMessage(diff, commitTypes, extraContext string) (string, error) {
	if err := p.ensureAuthenticated(); err != nil {
		return "", err
	}

	if diff == "" {
		return "", fmt.Errorf("empty diff provided")
	}

	diffSummary, optimizedDiff := summarizeDiff(diff)
	if diffSummary.SummarySize < diffSummary.OriginalSize {
		if extraContext != "" {
			extraContext += " "
		}
		extraContext += fmt.Sprintf("(Diff summarized: %d files changed. %s)",
			len(diffSummary.FilesChanged),
			diffSummary.Summary)
	}

	diff = optimizedDiff
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

	return p.callGeminiAPI(systemPrompt, userPrompt)
}

func (p *GeminiProvider) GenerateDetailedCommit(diff, commitTypes, extraContext string) (*CommitResult, error) {
	if err := p.ensureAuthenticated(); err != nil {
		return nil, err
	}

	if diff == "" {
		return nil, fmt.Errorf("empty diff provided")
	}

	diffSummary, optimizedDiff := summarizeDiff(diff)
	if diffSummary.SummarySize < diffSummary.OriginalSize {
		if extraContext != "" {
			extraContext += " "
		}
		extraContext += fmt.Sprintf("(Diff summarized: %d files changed. %s)",
			len(diffSummary.FilesChanged),
			diffSummary.Summary)
	}

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

	response, err := p.callGeminiAPI(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	result := parseDetailedCommitMessage(response)
	return result, nil
}

func (p *GeminiProvider) callGeminiAPI(systemPrompt, userPrompt string) (string, error) {
	// Construct URL - handle both test server URLs and production URLs
	url := fmt.Sprintf("%s/models/%s:generateContent", geminiAPIBaseURL, p.model)

	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{{Text: systemPrompt}},
				Role:  "user",
			},
			{
				Parts: []GeminiPart{{Text: userPrompt}},
				Role:  "user",
			},
		},
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     0.1, // Low temperature for deterministic, consistent commit messages
			MaxOutputTokens: 500, // Limit output length for faster responses (commit messages are short)
		},
	}

	jsonPayload, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Gemini request: %w", err)
	}

	// Retry logic for transient errors (503, 429, network errors)
	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			fmt.Printf("⏳ Retrying in %v... (attempt %d/%d)\n", backoff, attempt+1, maxRetries)
			time.Sleep(backoff)
		}

		result, err := p.executeGeminiRequest(url, jsonPayload)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return "", err // Don't retry non-retryable errors
		}
	}

	return "", fmt.Errorf("gemini API failed after %d attempts: %w", maxRetries, lastErr)
}

func (p *GeminiProvider) executeGeminiRequest(url string, jsonPayload []byte) (string, error) {
	// Use context with timeout for faster failure on slow responses
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.useOAuth {
		req.Header.Set("Authorization", "Bearer "+p.accessToken)
	} else {
		req.Header.Set("x-goog-api-key", p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse GeminiResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error != nil {
			return "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, errorResponse.Error.Message)
		}
		return "", fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var successResponse GeminiResponse
	if err := json.Unmarshal(bodyBytes, &successResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal Gemini response: %w. Body: %s", err, string(bodyBytes))
	}

	if len(successResponse.Candidates) == 0 || len(successResponse.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content found in Gemini response")
	}

	return successResponse.Candidates[0].Content.Parts[0].Text, nil
}

// isRetryableError checks if an error is transient and worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Retry on 503 (service unavailable) and 429 (rate limit)
	if strings.Contains(errStr, "status 503") || strings.Contains(errStr, "status 429") {
		return true
	}

	// Retry on network/timeout errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "temporary") {
		return true
	}

	// Don't retry on 4xx errors (except 429) or 5xx errors that aren't 503
	// These are typically permanent errors (auth, bad request, etc.)
	return false
}

func (p *GeminiProvider) GetModel() string {
	return p.model
}

// Helper functions for OAuth

type oauthServer struct {
	codeChan  chan string
	stateChan chan string
	verifier  string // Store verifier for PKCE
	server    *http.Server
}

func startLocalOAuthServer(verifier string) *oauthServer {
	codeChan := make(chan string, 1)
	stateChan := make(chan string, 1)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		if errorParam != "" {
			errorDesc := r.URL.Query().Get("error_description")
			codeChan <- ""
			stateChan <- ""
			w.WriteHeader(http.StatusBadRequest)
			html := fmt.Sprintf(`<html><body style="font-family: sans-serif; text-align: center; padding: 50px;">
				<h1 style="color: #d32f2f;">❌ Authentication Failed</h1>
				<p>Error: %s</p>
				<p>%s</p>
				<p>You can close this window.</p>
			</body></html>`, errorParam, errorDesc)
			w.Write([]byte(html))
			go func() {
				time.Sleep(2 * time.Second)
				server.Close()
			}()
			return
		}

		if code != "" && state != "" {
			codeChan <- code
			stateChan <- state
			w.WriteHeader(http.StatusOK)
			html := `<html><body style="font-family: sans-serif; text-align: center; padding: 50px;">
				<h1 style="color: #4caf50;">✅ Authentication Successful!</h1>
				<p>You have successfully authenticated with Google.</p>
				<p>You can close this window and return to the terminal.</p>
			</body></html>`
			w.Write([]byte(html))
			go func() {
				time.Sleep(2 * time.Second)
				server.Close()
			}()
		} else {
			codeChan <- ""
			stateChan <- ""
			w.WriteHeader(http.StatusBadRequest)
			html := `<html><body style="font-family: sans-serif; text-align: center; padding: 50px;">
				<h1 style="color: #d32f2f;">❌ Authentication Failed</h1>
				<p>Missing authorization code or state parameter.</p>
				<p>You can close this window.</p>
			</body></html>`
			w.Write([]byte(html))
			go func() {
				time.Sleep(2 * time.Second)
				server.Close()
			}()
		}
	})

	go func() {
		_ = server.ListenAndServe()
	}()

	return &oauthServer{
		codeChan:  codeChan,
		stateChan: stateChan,
		verifier:  verifier,
		server:    server,
	}
}

func (s *oauthServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}

func generateState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func getGeminiTokenPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".gemini_token.json"
	}
	return fmt.Sprintf("%s/.goji/gemini_token.json", homeDir)
}

func loadGeminiToken() (*oauth2.Token, error) {
	tokenPath := getGeminiTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func saveGeminiToken(token *oauth2.Token) error {
	tokenPath := getGeminiTokenPath()
	dir := strings.TrimSuffix(tokenPath, "/gemini_token.json")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600)
}
