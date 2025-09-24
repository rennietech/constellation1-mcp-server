package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"reso-mcp-server/auth"
)

// Client represents the RESO API client
type Client struct {
	baseURL     string
	oauthClient *auth.OAuthClient
	httpClient  *http.Client
}

// NewClient creates a new RESO API client
func NewClient(baseURL string, oauthClient *auth.OAuthClient) *Client {
	return &Client{
		baseURL:     baseURL,
		oauthClient: oauthClient,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Query executes a query against the RESO API
func (c *Client) Query(params QueryParams) (*APIResponse, error) {
	startTime := time.Now()

	// Validate entity
	if !IsValidEntity(params.Entity) {
		return nil, fmt.Errorf("unsupported entity: %s", params.Entity)
	}

	// Validate skip limit
	if params.Skip > 0 {
		limit := GetEntitySkipLimit(params.Entity)
		if params.Skip > limit {
			return nil, fmt.Errorf("skip value %d exceeds limit %d for entity %s", params.Skip, limit, params.Entity)
		}
	}

	// Build URL
	apiURL := fmt.Sprintf("%s/%s", c.baseURL, params.Entity)

	// Build query parameters
	queryParams := url.Values{}

	if params.Select != "" {
		queryParams.Set("$select", params.Select)
	}

	if params.Filter != "" {
		queryParams.Set("$filter", params.Filter)
	}

	if params.Top > 0 {
		queryParams.Set("$top", strconv.Itoa(params.Top))
	}

	if params.Skip > 0 {
		queryParams.Set("$skip", strconv.Itoa(params.Skip))
	}

	if params.OrderBy != "" {
		queryParams.Set("$orderby", params.OrderBy)
	}

	if params.IgnoreNulls {
		queryParams.Set("$ignorenulls", "true")
	}

	if params.IgnoreCase {
		queryParams.Set("$ignorecase", "true")
	}

	// Add query parameters to URL
	if len(queryParams) > 0 {
		apiURL += "?" + queryParams.Encode()
	}

	// Get access token
	token, err := c.oauthClient.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Host", "listings.cdatalabs.com")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("User-Agent", "RESO-MCP-Server/1.0")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response with decompression support
	var reader io.Reader = resp.Body

	// Check if response is gzip-compressed
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, errorResp.Error.Code, errorResp.Error.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Add metadata
	apiResp.RequestTime = startTime
	apiResp.ResponseTime = time.Since(startTime)
	apiResp.RequestParams = params

	return &apiResp, nil
}

// GetMetadata retrieves the metadata for the RESO API
func (c *Client) GetMetadata() (string, error) {
	// Get access token
	token, err := c.oauthClient.GetToken()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Create request
	metadataURL := c.baseURL + "/$metadata"
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Host", "listings.cdatalabs.com")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("User-Agent", "RESO-MCP-Server/1.0")

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
		return "", fmt.Errorf("metadata request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// TestConnection tests the connection to the RESO API
func (c *Client) TestConnection() error {
	// Try to get a token first
	_, err := c.oauthClient.GetToken()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Try a simple query
	params := QueryParams{
		Entity:      "Property",
		Top:         1,
		IgnoreNulls: true,
	}

	_, err = c.Query(params)
	if err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	return nil
}
