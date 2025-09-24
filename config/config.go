package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the configuration for the RESO MCP server
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthURL      string `json:"auth_url"`
	BaseURL      string `json:"base_url"`
}

// MCPSettings represents the MCP server settings format
type MCPSettings struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		AuthURL: "https://authenticate.constellation1apis.com/oauth2/token",
		BaseURL: "https://listings.cdatalabs.com/odata",
	}
}

// LoadFromMCPSettings loads configuration from MCP settings
func (c *Config) LoadFromMCPSettings(settings map[string]interface{}) error {
	if settings == nil {
		return fmt.Errorf("no settings provided")
	}

	if clientID, ok := settings["client_id"].(string); ok && clientID != "" {
		c.ClientID = clientID
	}

	if clientSecret, ok := settings["client_secret"].(string); ok && clientSecret != "" {
		c.ClientSecret = clientSecret
	}

	// Don't require credentials during MCP initialization
	// They will be validated when actually needed
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if clientID := os.Getenv("RESO_CLIENT_ID"); clientID != "" {
		c.ClientID = clientID
	}
	if clientSecret := os.Getenv("RESO_CLIENT_SECRET"); clientSecret != "" {
		c.ClientSecret = clientSecret
	}
	if authURL := os.Getenv("RESO_AUTH_URL"); authURL != "" {
		c.AuthURL = authURL
	}
	if baseURL := os.Getenv("RESO_BASE_URL"); baseURL != "" {
		c.BaseURL = baseURL
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	if c.AuthURL == "" {
		return fmt.Errorf("auth_url is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	return nil
}

// ValidateCredentials checks if credentials are provided (for API calls)
func (c *Config) ValidateCredentials() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required - please configure in MCP settings")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("client_secret is required - please configure in MCP settings")
	}
	return nil
}

// ToJSON converts the config to JSON string
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

