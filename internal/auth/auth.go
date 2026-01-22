package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"HyPrism/internal/env"
)

const (
	AuthURL         = "https://oauth.accounts.hytale.com/oauth2/auth"
	TokenURL        = "https://oauth.accounts.hytale.com/oauth2/token"
	LauncherDataURL = "https://account-data.hytale.com/my-account/get-launcher-data"
	SessionURL      = "https://sessions.hytale.com/game-session/new"
	
	ClientID = "hytale-launcher"
)

type AuthSession struct {
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token"`
	ExpiresAt      time.Time `json:"expires_at"`
	SessionToken   string    `json:"session_token"`
	IdentityToken  string    `json:"identity_token"`
	Username       string    `json:"username"`
	UUID           string    `json:"uuid"`
	AccountOwnerID string    `json:"account_owner_id"`
}

type Profile struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type ProfilesResponse struct {
	Owner    string    `json:"owner"`
	Profiles []Profile `json:"profiles"`
}

type GameSessionResponse struct {
	SessionToken  string `json:"sessionToken"`
	IdentityToken string `json:"identityToken"`
	ExpiresAt     string `json:"expiresAt"`
}

func (s *AuthSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func GetSessionPath() string {
	return filepath.Join(env.GetDefaultAppDir(), "session.json")
}

func SaveSession(session *AuthSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionPath := GetSessionPath()
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	if err := os.WriteFile(sessionPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	fmt.Printf("Session saved for user: %s (UUID: %s)\n", session.Username, session.UUID)
	return nil
}

func LoadSession() (*AuthSession, error) {
	sessionPath := GetSessionPath()
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session AuthSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	return &session, nil
}

func ClearSession() error {
	sessionPath := GetSessionPath()
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}
	fmt.Println("Session cleared")
	return nil
}

// RefreshSession refreshes an expired session using the refresh token
func RefreshSession(session *AuthSession) (*AuthSession, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", session.RefreshToken)
	data.Set("client_id", ClientID)

	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "HyPrism/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	newSession := &AuthSession{
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Username:       session.Username,
		UUID:           session.UUID,
		AccountOwnerID: session.AccountOwnerID,
	}

	if err := createGameSession(newSession); err != nil {
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	return newSession, nil
}

// GetValidSession returns a valid session, refreshing if necessary
func GetValidSession() (*AuthSession, error) {
	session, err := LoadSession()
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, fmt.Errorf("no session found - please login")
	}

	// If session is expired, try to refresh
	if session.IsExpired() {
		fmt.Println("Session expired, attempting to refresh...")
		newSession, err := RefreshSession(session)
		if err != nil {
			// Refresh failed, user needs to login again
			ClearSession()
			return nil, fmt.Errorf("session expired and refresh failed - please login again")
		}

		// Save the new session
		if err := SaveSession(newSession); err != nil {
			return nil, err
		}

		return newSession, nil
	}

	return session, nil
}

// IsLoggedIn checks if there's a valid session
func IsLoggedIn() bool {
	session, err := GetValidSession()
	return err == nil && session != nil
}

// fetchProfiles fetches the user's launcher data (profiles, patchlines, etc.)
func fetchProfiles(accessToken string) (*ProfilesResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	req, err := http.NewRequest("GET", LauncherDataURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "HyPrism/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("launcher data request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var profiles ProfilesResponse
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, err
	}

	return &profiles, nil
}

// createGameSession creates a new game session with session and identity tokens
func createGameSession(session *AuthSession) error {
	client := &http.Client{Timeout: 30 * time.Second}

	payload := map[string]string{
		"uuid": session.UUID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", SessionURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HyPrism/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("game session creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var gameSession GameSessionResponse
	if err := json.Unmarshal(body, &gameSession); err != nil {
		return err
	}

	session.SessionToken = gameSession.SessionToken
	session.IdentityToken = gameSession.IdentityToken

	// Parse expiry time
	expiresAt, err := time.Parse(time.RFC3339, gameSession.ExpiresAt)
	if err != nil {
		fmt.Printf("Warning: failed to parse game session expiry: %v\n", err)
	} else {
		// Use the earlier of OAuth token expiry or game session expiry
		if expiresAt.Before(session.ExpiresAt) {
			session.ExpiresAt = expiresAt
		}
	}

	return nil
}

// generateRandomState generates a random state string for OAuth
// Matches the official Hytale launcher's crypto/rand.Text implementation:
// - 26 random bytes
// - Each byte masked with 0x1F (giving values 0-31)
// - Used as index into "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
func generateRandomState() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	b := make([]byte, 26)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]&0x1F]
	}
	return string(b)
}

// generateCodeVerifier generates a random PKCE code verifier
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge generates the PKCE code challenge from verifier
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// LoginWithAuthCodeFlow initiates the OAuth authorization code flow with PKCE
func LoginWithAuthCodeFlow(ctx context.Context, progressCallback func(message string), openBrowser func(string) error) (*AuthSession, error) {
	// Generate PKCE verifier and challenge
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Start local callback server on random available port
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer listener.Close()

	// Get the actual port assigned
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("OAuth callback server listening on port %d\n", port)

	// HTTP server for OAuth callback
	mux := http.NewServeMux()
	mux.HandleFunc("/authorization-callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		error := r.URL.Query().Get("error")

		if error != "" {
			errChan <- fmt.Errorf("OAuth error: %s", error)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h1>Authentication failed</h1><p>Error: %s</p><p>You can close this window.</p></body></html>", error)
			return
		}

		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "<html><body><h1>Authentication failed</h1><p>No code received.</p><p>You can close this window.</p></body></html>")
			return
		}

		codeChan <- code
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html><body><h1>Authentication successful!</h1><p>You can close this window and return to HyPrism.</p></body></html>")
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Shutdown(context.Background())

	// Build authorization URL matching official launcher
	// State contains JSON with random state and port number, base64 encoded
	stateData := map[string]string{
		"state": generateRandomState(),
		"port":  fmt.Sprintf("%d", port),
	}
	stateJSON, _ := json.Marshal(stateData)
	stateParam := base64.StdEncoding.EncodeToString(stateJSON)

	// Build URL with parameters in exact order as official launcher
	// Must manually construct to preserve order (url.Values.Encode() sorts alphabetically)
	query := fmt.Sprintf(
		"access_type=offline&client_id=%s&code_challenge=%s&code_challenge_method=S256&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		url.QueryEscape(ClientID),
		url.QueryEscape(codeChallenge),
		url.QueryEscape("https://accounts.hytale.com/consent/client"),
		url.QueryEscape("openid offline auth:launcher"),
		url.QueryEscape(stateParam),
	)
	authURL := AuthURL + "?" + query

	if progressCallback != nil {
		progressCallback("Opening browser for authentication...")
	}

	// Open browser
	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	if progressCallback != nil {
		progressCallback("Waiting for authorization...")
	}

	// Wait for code or error
	var authCode string
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("login cancelled")
	case err := <-errChan:
		return nil, err
	case authCode = <-codeChan:
		// Success, continue
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timed out")
	}

	if progressCallback != nil {
		progressCallback("Exchanging authorization code for token...")
	}

	// Exchange authorization code for tokens
	// Must use the same redirect_uri as in the authorization request
	session, err := exchangeCodeForToken(authCode, codeVerifier, "https://accounts.hytale.com/consent/client")
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Fetch profile information
	if progressCallback != nil {
		progressCallback("Fetching profile information...")
	}

	profiles, err := fetchProfiles(session.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profiles: %w", err)
	}

	if len(profiles.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles found for this account")
	}

	// Use the first profile
	profile := profiles.Profiles[0]
	session.Username = profile.Username
	session.UUID = profile.UUID
	session.AccountOwnerID = profiles.Owner

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Logged in as: %s", session.Username))
	}

	// Create game session
	if err := createGameSession(session); err != nil {
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	return session, nil
}

// exchangeCodeForToken exchanges authorization code for access token using PKCE
func exchangeCodeForToken(code, verifier, redirectURL string) (*AuthSession, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)
	data.Set("client_id", ClientID)
	data.Set("code_verifier", verifier)

	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "HyPrism/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// Parse error response
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.Unmarshal(body, &errResp)

		if errResp.Error == "authorization_pending" {
			return nil, fmt.Errorf("authorization_pending")
		}

		return nil, fmt.Errorf("token request failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	session := &AuthSession{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return session, nil
}
