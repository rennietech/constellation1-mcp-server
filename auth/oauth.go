package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// OAuthClient handles OAuth2 authentication for RESO API
type OAuthClient struct {
	clientID     string
	clientSecret string
	authURL      string
	token        *TokenResponse
	tokenExpiry  time.Time
	mutex        sync.RWMutex
	httpClient   *http.Client
}

// NewOAuthClient creates a new OAuth client
func NewOAuthClient(clientID, clientSecret, authURL string) *OAuthClient {
	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		authURL:      authURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken returns a valid access token, refreshing if necessary
func (c *OAuthClient) GetToken() (string, error) {
	c.mutex.RLock()
	if c.token != nil && time.Now().Before(c.tokenExpiry) {
		token := c.token.AccessToken
		c.mutex.RUnlock()
		return token, nil
	}
	c.mutex.RUnlock()

	return c.refreshToken()
}

// refreshToken obtains a new access token
func (c *OAuthClient) refreshToken() (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Double-check pattern
	if c.token != nil && time.Now().Before(c.tokenExpiry) {
		return c.token.AccessToken, nil
	}

	// Encode credentials in Base64
	credentials := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)

	// Create request
	req, err := http.NewRequest("POST", c.authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Host", "authenticate.constellation1apis.com")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// Store token with buffer time (subtract 60 seconds for safety)
	c.token = &tokenResp
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return tokenResp.AccessToken, nil
}

// IsTokenValid checks if the current token is valid
func (c *OAuthClient) IsTokenValid() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.token != nil && time.Now().Before(c.tokenExpiry)
}

// ClearToken clears the stored token (useful for testing or forced refresh)
func (c *OAuthClient) ClearToken() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.token = nil
	c.tokenExpiry = time.Time{}
}

