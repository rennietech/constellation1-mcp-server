package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"reso-mcp-server/api"
	"reso-mcp-server/auth"
	"reso-mcp-server/config"
	"reso-mcp-server/tools"
)

// MCPMessage represents a message in the MCP protocol
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in the MCP protocol
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams represents the parameters for the initialize method
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]interface{} `json:"clientInfo"`
}

// InitializeResult represents the result of the initialize method
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]interface{} `json:"serverInfo"`
}

// ListToolsResult represents the result of the tools/list method
type ListToolsResult struct {
	Tools []tools.MCPTool `json:"tools"`
}

// CallToolParams represents the parameters for the tools/call method
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// MCPServer represents the MCP server
type MCPServer struct {
	config          *config.Config
	apiClient       *api.Client
	resoTool        *tools.ResoQueryTool
	pendingSettings map[string]interface{}
}

// NewMCPServer creates a new MCP server
func NewMCPServer() *MCPServer {
	return &MCPServer{
		config: config.DefaultConfig(),
	}
}

// Initialize initializes the MCP server with configuration
func (s *MCPServer) Initialize(settings map[string]interface{}) error {
	// Load configuration from settings
	if err := s.config.LoadFromMCPSettings(settings); err != nil {
		// Try loading from environment variables as fallback
		s.config.LoadFromEnv()
	}

	// Create OAuth client (even if credentials are not yet provided)
	oauthClient := auth.NewOAuthClient(s.config.ClientID, s.config.ClientSecret, s.config.AuthURL)

	// Create API client
	s.apiClient = api.NewClient(s.config.BaseURL, oauthClient)

	// Create tools
	s.resoTool = tools.NewResoQueryTool(s.apiClient, s.config)

	// Don't test connection during initialization - defer until first tool call
	// This allows the MCP server to start even if RESO API is temporarily unavailable

	return nil
}

// HandleMessage handles an incoming MCP message
func (s *MCPServer) HandleMessage(msg MCPMessage) MCPMessage {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return s.handleInitialized(msg)
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolsCall(msg)
	default:
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", msg.Method),
			},
		}
	}
}

// handleInitialize handles the initialize method
func (s *MCPServer) handleInitialize(msg MCPMessage) MCPMessage {
	var params InitializeParams
	if msg.Params != nil {
		if paramsBytes, err := json.Marshal(msg.Params); err == nil {
			json.Unmarshal(paramsBytes, &params)
		}
	}

	// Start with pending settings from command line/environment
	var settings map[string]interface{}
	if s.pendingSettings != nil {
		settings = make(map[string]interface{})
		for k, v := range s.pendingSettings {
			settings[k] = v
		}
	}
	
	// Extract settings - MCP clients may pass settings in different ways
	// Method 1: Check if settings are in capabilities
	if clientCaps, ok := params.Capabilities["settings"].(map[string]interface{}); ok {
		if settings == nil {
			settings = make(map[string]interface{})
		}
		for k, v := range clientCaps {
			settings[k] = v // JSON-RPC settings override env/args
		}
	}
	
	// Method 2: Check if settings are directly in params (common for some MCP clients)
	if rawParams, ok := msg.Params.(map[string]interface{}); ok {
		if settingsData, exists := rawParams["settings"]; exists {
			if settingsMap, ok := settingsData.(map[string]interface{}); ok {
				if settings == nil {
					settings = make(map[string]interface{})
				}
				for k, v := range settingsMap {
					settings[k] = v // JSON-RPC settings override env/args
				}
			}
		}
	}
	
	// Method 3: Check if the entire params contains client_id/client_secret directly
	if rawParams, ok := msg.Params.(map[string]interface{}); ok {
		if settings == nil {
			settings = make(map[string]interface{})
		}
		if clientID, exists := rawParams["client_id"]; exists {
			settings["client_id"] = clientID
		}
		if clientSecret, exists := rawParams["client_secret"]; exists {
			settings["client_secret"] = clientSecret
		}
	}

	// Initialize server with settings
	if err := s.Initialize(settings); err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: fmt.Sprintf("Initialization failed: %s", err.Error()),
			},
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		ServerInfo: map[string]interface{}{
			"name":    "reso-mcp-server",
			"version": "1.0.0",
		},
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}
}

// handleInitialized handles the initialized notification
func (s *MCPServer) handleInitialized(msg MCPMessage) MCPMessage {
	// This is a notification, no response needed
	return MCPMessage{}
}

// handleToolsList handles the tools/list method
func (s *MCPServer) handleToolsList(msg MCPMessage) MCPMessage {
	if s.resoTool == nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Server not initialized",
			},
		}
	}

	result := ListToolsResult{
		Tools: []tools.MCPTool{
			s.resoTool.GetToolDefinition(),
		},
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}
}

// handleToolsCall handles the tools/call method
func (s *MCPServer) handleToolsCall(msg MCPMessage) MCPMessage {
	if s.resoTool == nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Server not initialized",
			},
		}
	}

	var params CallToolParams
	if msg.Params != nil {
		if paramsBytes, err := json.Marshal(msg.Params); err == nil {
			if err := json.Unmarshal(paramsBytes, &params); err != nil {
				return MCPMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error: &MCPError{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
			}
		}
	}

	switch params.Name {
	case "reso_query":
		result := s.resoTool.Execute(params.Arguments)
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  result,
		}
	default:
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Tool not found: %s", params.Name),
			},
		}
	}
}

func main() {
	// Configure logging to stderr to avoid interfering with MCP JSON-RPC on stdout
	log.SetOutput(os.Stderr)
	
	// Parse command line arguments
	var clientID = flag.String("client-id", "", "RESO API Client ID")
	var clientSecret = flag.String("client-secret", "", "RESO API Client Secret")
	flag.Parse()

	server := NewMCPServer()
	scanner := bufio.NewScanner(os.Stdin)

	log.Println("RESO MCP Server starting...")
	
	// Store settings for later use but don't pre-initialize
	// This avoids sending any messages before the MCP client is ready
	envSettings := make(map[string]interface{})
	
	// 1. Command line arguments (highest priority)
	if *clientID != "" {
		envSettings["client_id"] = *clientID
	}
	if *clientSecret != "" {
		envSettings["client_secret"] = *clientSecret
	}
	
	// 2. Standard environment variables
	if clientID := os.Getenv("CLIENT_ID"); clientID != "" && envSettings["client_id"] == nil {
		envSettings["client_id"] = clientID
	}
	if clientSecret := os.Getenv("CLIENT_SECRET"); clientSecret != "" && envSettings["client_secret"] == nil {
		envSettings["client_secret"] = clientSecret
	}
	
	// 3. RESO-specific environment variables
	if clientID := os.Getenv("RESO_CLIENT_ID"); clientID != "" && envSettings["client_id"] == nil {
		envSettings["client_id"] = clientID
	}
	if clientSecret := os.Getenv("RESO_CLIENT_SECRET"); clientSecret != "" && envSettings["client_secret"] == nil {
		envSettings["client_secret"] = clientSecret
	}
	
	// 4. MCP client may set these specific environment variables
	if clientID := os.Getenv("MCP_RESO_CLIENT_ID"); clientID != "" && envSettings["client_id"] == nil {
		envSettings["client_id"] = clientID
	}
	if clientSecret := os.Getenv("MCP_RESO_CLIENT_SECRET"); clientSecret != "" && envSettings["client_secret"] == nil {
		envSettings["client_secret"] = clientSecret
	}
	
	// Store settings in server for use during initialization
	if len(envSettings) > 0 {
		log.Printf("Found settings from environment/args, will use during initialization")
		// Store in server for later use
		server.pendingSettings = envSettings
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg MCPMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		response := server.HandleMessage(msg)
		
		// Only send response if it's not empty (for notifications)
		if response.JSONRPC != "" {
			responseBytes, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshaling response: %v", err)
				continue
			}
			fmt.Println(string(responseBytes))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

