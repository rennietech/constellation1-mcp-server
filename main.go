package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rennietech/constellation1-mcp-server/api"
	"github.com/rennietech/constellation1-mcp-server/auth"
	"github.com/rennietech/constellation1-mcp-server/config"
	"github.com/rennietech/constellation1-mcp-server/tools"
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

// MCPResource represents an MCP resource
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ListResourcesResult represents the result of the resources/list method
type ListResourcesResult struct {
	Resources []MCPResource `json:"resources"`
}

// ReadResourceParams represents the parameters for the resources/read method
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// ReadResourceResult represents the result of the resources/read method
type ReadResourceResult struct {
	Contents []MCPResourceContent `json:"contents"`
}

// MCPResourceContent represents content in a resource
type MCPResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

// MCPServer represents the MCP server
type MCPServer struct {
	config          *config.Config
	apiClient       *api.Client
	resoTool        *tools.ResoQueryTool
	helpTool        *tools.ResoHelpTool
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
	s.helpTool = tools.NewResoHelpToolWithAPI(s.apiClient)

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
	case "resources/list":
		return s.handleResourcesList(msg)
	case "resources/read":
		return s.handleResourcesRead(msg)
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
			"resources": map[string]interface{}{
				"subscribe":   false,
				"listChanged": false,
			},
		},
		ServerInfo: map[string]interface{}{
			"name":        "constellation1-mcp-server",
			"version":     "1.0.0",
			"description": "RESO (Real Estate Standards Organization) MCP Server providing comprehensive access to MLS data through the Constellation1 API. Features include property listings, agent information, office details, media files, and market analytics with advanced filtering, entity expansion, and privacy controls.",
			"author":      "Rennie Technologies",
			"homepage":    "https://github.com/rennietech/constellation1-mcp-server",
			"license":     "MIT",
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
	if s.resoTool == nil || s.helpTool == nil {
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
			s.helpTool.GetToolDefinition(),
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
	case "reso_help":
		result := s.helpTool.Execute(params.Arguments)
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

// handleResourcesList handles the resources/list method
func (s *MCPServer) handleResourcesList(msg MCPMessage) MCPMessage {
	resources := []MCPResource{
		{
			URI:         "reso://field-reference",
			Name:        "RESO Field Reference Guide",
			Description: "Comprehensive guide to RESO fields, entities, enums, filter patterns, and best practices for AI-friendly real estate data queries",
			MimeType:    "text/markdown",
		},
		{
			URI:         "reso://quick-start",
			Name:        "RESO Query Quick Start",
			Description: "Quick reference for common RESO query patterns and examples organized by use case",
			MimeType:    "text/markdown",
		},
	}

	result := ListResourcesResult{
		Resources: resources,
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}
}

// handleResourcesRead handles the resources/read method
func (s *MCPServer) handleResourcesRead(msg MCPMessage) MCPMessage {
	var params ReadResourceParams
	if msg.Params != nil {
		if paramsBytes, err := json.Marshal(msg.Params); err == nil {
			json.Unmarshal(paramsBytes, &params)
		}
	}

	var content string
	var mimeType string

	switch params.URI {
	case "reso://field-reference":
		content = s.getFieldReferenceContent()
		mimeType = "text/markdown"
	case "reso://quick-start":
		content = s.getQuickStartContent()
		mimeType = "text/markdown"
	default:
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: fmt.Sprintf("Resource not found: %s", params.URI),
			},
		}
	}

	result := ReadResourceResult{
		Contents: []MCPResourceContent{
			{
				URI:      params.URI,
				MimeType: mimeType,
				Text:     content,
			},
		},
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}
}

// getFieldReferenceContent returns the complete RESO field reference guide
func (s *MCPServer) getFieldReferenceContent() string {
	// Use dynamic content from help tool if available
	if s.helpTool != nil && s.helpTool.HasMetadata() {
		entityGuide := s.helpTool.GetEntityGuide()
		enumsGuide := s.helpTool.GetEnumsGuide()
		if entityGuide != "" && enumsGuide != "" {
			return entityGuide + "\n\n" + enumsGuide
		}
	}

	// Fallback to static content
	return `# RESO Field Reference Guide

This guide provides AI-friendly reference information for commonly used RESO fields and their valid values.

## Core Entity Relationships

- **Property.ListingKey** ↔ **Media.ResourceRecordKey** (Property photos/media)
- **Property.ListingKey** ↔ **OpenHouse.ListingKey** (Open house events)
- **Property.ListingKey** ↔ **Dom.ListingId** (Days on market data)
- **Property.ListingKey** ↔ **PropertyRooms.ListingKey** (Room details)
- **Property.ListingKey** ↔ **PropertyUnitTypes.ListingKey** (Unit types)
- **Member.MemberMlsId** ↔ **Property.ListAgentMlsId** (Agent-Property relationship)
- **Office.OfficeMlsId** ↔ **Member.OfficeMlsId** (Agent-Office relationship)

## StandardStatus (Property Status)

- **Active** - Currently available for sale
- **ActiveUnderContract** - Under contract but still active
- **Pending** - Sale pending, not available
- **Closed** - Sale completed
- **Canceled** - Listing canceled
- **Expired** - Listing expired
- **Withdrawn** - Withdrawn from market
- **Hold** - Temporarily held
- **ComingSoon** - Coming soon to market
- **OffMarket** - Off market

## PropertyType (Primary Categories)

- **Residential** - Single-family homes, condos, townhouses
- **ResidentialIncome** - Multi-family investment properties
- **ResidentialLease** - Rental properties
- **CommercialSale** - Commercial properties for sale
- **CommercialLease** - Commercial properties for lease
- **BusinessOpportunity** - Business sales
- **Farm** - Farm and agricultural properties
- **Land** - Vacant land and lots
- **ManufacturedInPark** - Mobile homes in parks

## PropertySubType (Detailed Types)

**Residential**: SingleFamilyResidence, Condominium, Townhouse, Duplex, Triplex, Quadruplex, ManufacturedHome, Cabin
**Commercial**: Office, Retail, Industrial, Warehouse
**Other**: Farm, UnimprovedLand

## MediaCategory & Permission

**Categories**: Photo, Video, BrandedVideo, UnbrandedVideo, BrandedVirtualTour, UnbrandedVirtualTour, FloorPlan, Document, AgentPhoto, OfficePhoto, OfficeLogo

**Permission**: Public (MediaURL available), Private (MediaURL not available)

## Common Property Fields by Category

### Identification
ListingKey, ListingId, MlsStatus, UniversalPropertyId

### Address & Location
StreetNumber, StreetName, City, StateOrProvince, PostalCode, UnparsedAddress, Latitude, Longitude, MLSAreaMajor, MLSAreaMinor

### Pricing
ListPrice, ClosePrice, OriginalListPrice, PreviousListPrice, TaxAnnualAmount

### Property Details
BedroomsTotal, BathroomsTotal, LivingArea, YearBuilt, LotSizeSquareFeet, Stories, PropertyType, PropertySubType

### Agent Info
ListAgentFullName, ListAgentEmail, ListAgentDirectPhone, ListOfficeName

### Status & Dates
StandardStatus, OnMarketTimestamp, ModificationTimestamp, DaysOnMarket

## Essential Filter Patterns

**Active Listings**: ` + "`" + `StandardStatus eq 'Active'` + "`" + `
**Price Range**: ` + "`" + `ListPrice ge 200000 and ListPrice le 500000` + "`" + `
**Location**: ` + "`" + `City eq 'Seattle' and StateOrProvince eq 'WA'` + "`" + `
**Property Features**: ` + "`" + `BedroomsTotal ge 3 and BathroomsTotal ge 2` + "`" + `
**Recent Sales**: ` + "`" + `StandardStatus eq 'Closed' and CloseDate ge 2024-01-01` + "`" + `
**Property Types**: ` + "`" + `PropertySubType eq 'Condominium'` + "`" + `

## Expand Examples

**Property with Public Photos**: ` + "`" + `Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc)` + "`" + `
**Property with Open Houses**: ` + "`" + `OpenHouse($filter=OpenHouseStartTime gt now())` + "`" + `
**Multiple Expansions**: ` + "`" + `Media($filter=Permission ne 'Private'),OpenHouse,Dom` + "`" + `

## Image URL Sizing

**Predefined**: Add ?d=t (thumbnail), ?d=s (small), ?d=l (large)
**Custom**: Add ?d=600 (width) or ?d=500x320 (width x height)

For complete details, see the full RESO Field Reference documentation.`
}

// getQuickStartContent returns quick start examples and patterns
func (s *MCPServer) getQuickStartContent() string {
	return `# RESO Query Quick Start

## Most Common Query Patterns

### 1. Active Properties in a City
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and City eq 'Seattle'",
  "select": "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,UnparsedAddress,PublicRemarks",
  "orderby": "ListPrice asc",
  "top": 25
}
` + "```" + `

### 2. Recent Sales Analysis
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Closed' and CloseDate ge 2024-01-01",
  "select": "ListingKey,ClosePrice,CloseDate,BedroomsTotal,LivingArea,City,DaysOnMarket",
  "orderby": "CloseDate desc",
  "top": 100
}
` + "```" + `

### 3. Property with Photos
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and PhotosCount gt 0",
  "expand": "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc;$top=5)",
  "select": "ListingKey,ListPrice,UnparsedAddress,PhotosCount",
  "top": 10
}
` + "```" + `

### 4. Agent Information
` + "```json" + `
{
  "entity": "Member",
  "filter": "MemberFullName eq 'John Smith'",
  "select": "MemberMlsId,MemberFullName,MemberEmail,MemberDirectPhone,OfficeName"
}
` + "```" + `

### 5. Luxury Properties
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and ListPrice gt 1000000",
  "select": "ListingKey,ListPrice,BedroomsTotal,LivingArea,City,UnparsedAddress",
  "orderby": "ListPrice desc",
  "top": 50
}
` + "```" + `

## Key Tips

- **Always filter by StandardStatus** for current data
- **Use expand for related data** instead of separate queries
- **Filter out private media** with ` + "`Permission ne 'Private'`" + `
- **Limit results with top** for better performance
- **Use ignorenulls=true** to reduce payload size
- **Order by relevant fields** for better data organization

## Essential Fields to Include

**Property Searches**: ListingKey, StandardStatus, ListPrice, BedroomsTotal, BathroomsTotal, LivingArea, UnparsedAddress, City, StateOrProvince, PublicRemarks

**Agent Searches**: MemberMlsId, MemberFullName, MemberEmail, MemberDirectPhone, OfficeName

**Media Searches**: MediaKey, ResourceRecordKey, MediaCategory, MediaURL, Permission, Order`
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
