package tools

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rennietech/constellation1-mcp-server/api"
	"github.com/rennietech/constellation1-mcp-server/config"
)

// MCPTool represents an MCP tool
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolResult represents the result of an MCP tool execution
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in an MCP tool result
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ResoQueryTool implements the reso_query MCP tool
type ResoQueryTool struct {
	client *api.Client
	config *config.Config
}

// NewResoQueryTool creates a new RESO query tool
func NewResoQueryTool(client *api.Client, cfg *config.Config) *ResoQueryTool {
	return &ResoQueryTool{
		client: client,
		config: cfg,
	}
}

// GetToolDefinition returns the MCP tool definition
func (t *ResoQueryTool) GetToolDefinition() MCPTool {
	return MCPTool{
		Name:        "reso_query",
		Description: "Query the RESO standard API for real estate data with comprehensive filtering and selection options. Supports all major entities including Property, Member, Office, Media, OpenHouse, Dom, PropertyUnitTypes, PropertyRooms, and RawMlsProperty.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"entity": map[string]interface{}{
					"type":        "string",
					"description": "Entity type to query",
					"enum": []string{
						"Property", "Member", "Office", "Media", "OpenHouse",
						"Dom", "PropertyUnitTypes", "PropertyRooms", "RawMlsProperty",
					},
				},
				"select": map[string]interface{}{
					"type":        "string",
					"description": "Comma-separated list of fields to return (e.g., 'ListingKey,StreetNumberNumeric,StandardStatus')",
				},
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "OData filter expression using supported operators (eq, ne, gt, ge, lt, le, has, in) and boolean operators (and, or). Example: \"StandardStatus eq 'Active' and ListPrice gt 100000\"",
				},
				"top": map[string]interface{}{
					"type":        "integer",
					"description": "Number of records to return (default: 10, max recommended: 1000)",
					"minimum":     1,
					"maximum":     1000,
				},
				"skip": map[string]interface{}{
					"type":        "integer",
					"description": "Number of records to skip for pagination (limits vary by entity)",
					"minimum":     0,
				},
				"orderby": map[string]interface{}{
					"type":        "string",
					"description": "Field(s) to order results by (e.g., 'ListPrice desc' or 'ModificationTimestamp')",
				},
				"ignorenulls": map[string]interface{}{
					"type":        "boolean",
					"description": "Exclude null fields to reduce payload size (default: true)",
					"default":     true,
				},
				"ignorecase": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable case-insensitive text searches for supported fields (default: false)",
					"default":     false,
				},
			},
			"required": []string{"entity"},
		},
	}
}

// Execute executes the RESO query tool
func (t *ResoQueryTool) Execute(args map[string]interface{}) MCPToolResult {
	// Validate credentials before proceeding
	if err := t.config.ValidateCredentials(); err != nil {
		return MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Configuration error: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	// Parse arguments
	params, err := t.parseArguments(args)
	if err != nil {
		return MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error parsing arguments: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	// Execute query
	response, err := t.client.Query(*params)
	if err != nil {
		return MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error executing query: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	// Format response
	responseJSON, err := response.ToJSON()
	if err != nil {
		return MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error formatting response: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	// Create summary
	summary := t.createSummary(response)

	return MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: summary,
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Full Response:\n```json\n%s\n```", responseJSON),
			},
		},
	}
}

// parseArguments parses the tool arguments into QueryParams
func (t *ResoQueryTool) parseArguments(args map[string]interface{}) (*api.QueryParams, error) {
	params := &api.QueryParams{
		IgnoreNulls: true, // Default to true
	}

	// Required: entity
	if entity, ok := args["entity"].(string); ok {
		params.Entity = entity
	} else {
		return nil, fmt.Errorf("entity is required")
	}

	// Optional: select
	if selectFields, ok := args["select"].(string); ok {
		params.Select = strings.TrimSpace(selectFields)
	}

	// Optional: filter
	if filter, ok := args["filter"].(string); ok {
		params.Filter = strings.TrimSpace(filter)
	}

	// Optional: top
	if top, ok := args["top"]; ok {
		switch v := top.(type) {
		case float64:
			params.Top = int(v)
		case int:
			params.Top = v
		case string:
			if topInt, err := strconv.Atoi(v); err == nil {
				params.Top = topInt
			}
		}
	}

	// Optional: skip
	if skip, ok := args["skip"]; ok {
		switch v := skip.(type) {
		case float64:
			params.Skip = int(v)
		case int:
			params.Skip = v
		case string:
			if skipInt, err := strconv.Atoi(v); err == nil {
				params.Skip = skipInt
			}
		}
	}

	// Optional: orderby
	if orderby, ok := args["orderby"].(string); ok {
		params.OrderBy = strings.TrimSpace(orderby)
	}

	// Optional: ignorenulls
	if ignorenulls, ok := args["ignorenulls"].(bool); ok {
		params.IgnoreNulls = ignorenulls
	}

	// Optional: ignorecase
	if ignorecase, ok := args["ignorecase"].(bool); ok {
		params.IgnoreCase = ignorecase
	}

	return params, nil
}

// createSummary creates a human-readable summary of the response
func (t *ResoQueryTool) createSummary(response *api.APIResponse) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("RESO API Query Results\n"))
	summary.WriteString(fmt.Sprintf("======================\n\n"))

	summary.WriteString(fmt.Sprintf("Entity: %s\n", response.RequestParams.Entity))
	summary.WriteString(fmt.Sprintf("Records Returned: %d\n", response.Count))
	summary.WriteString(fmt.Sprintf("Total Records Available: %d\n", response.TotalCount))
	summary.WriteString(fmt.Sprintf("Request Time: %s\n", response.RequestTime.Format("2006-01-02 15:04:05 UTC")))
	summary.WriteString(fmt.Sprintf("Response Time: %s\n\n", response.ResponseTime))

	// Query parameters
	if response.RequestParams.Select != "" {
		summary.WriteString(fmt.Sprintf("Selected Fields: %s\n", response.RequestParams.Select))
	}
	if response.RequestParams.Filter != "" {
		summary.WriteString(fmt.Sprintf("Filter: %s\n", response.RequestParams.Filter))
	}
	if response.RequestParams.Top > 0 {
		summary.WriteString(fmt.Sprintf("Top: %d\n", response.RequestParams.Top))
	}
	if response.RequestParams.Skip > 0 {
		summary.WriteString(fmt.Sprintf("Skip: %d\n", response.RequestParams.Skip))
	}
	if response.RequestParams.OrderBy != "" {
		summary.WriteString(fmt.Sprintf("Order By: %s\n", response.RequestParams.OrderBy))
	}
	summary.WriteString(fmt.Sprintf("Ignore Nulls: %t\n", response.RequestParams.IgnoreNulls))
	summary.WriteString(fmt.Sprintf("Ignore Case: %t\n", response.RequestParams.IgnoreCase))

	// Pagination info
	if response.NextLink != "" {
		summary.WriteString(fmt.Sprintf("\nNext Page Available: %s\n", response.NextLink))
	}

	// Sample data preview
	if len(response.Value) > 0 {
		summary.WriteString(fmt.Sprintf("\nSample Record Fields:\n"))
		sampleRecord := response.Value[0]
		fieldCount := 0
		for key := range sampleRecord {
			if fieldCount >= 10 { // Limit to first 10 fields
				summary.WriteString("... (and more fields)\n")
				break
			}
			summary.WriteString(fmt.Sprintf("- %s\n", key))
			fieldCount++
		}
	}

	return summary.String()
}
