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

// ResoQueryTool implements the reso_query MCP tool for querying RESO standard real estate data
//
// Common Use Cases and Examples:
//
// 1. Property Search Examples:
//   - Active homes in Seattle: entity="Property", filter="StandardStatus eq 'Active' and City eq 'Seattle'"
//   - Luxury condos: entity="Property", filter="PropertySubType eq 'Condominium' and ListPrice gt 1000000"
//   - Recently sold homes: entity="Property", filter="StandardStatus eq 'Closed' and CloseDate ge 2024-01-01"
//   - Homes with specific features: entity="Property", filter="BedroomsTotal ge 3 and BathroomsTotal ge 2 and LivingArea gt 2000"
//
// 2. Agent Research Examples:
//   - Find agent by name: entity="Member", filter="MemberFullName eq 'John Smith'"
//   - Agents in specific office: entity="Member", filter="OfficeName eq 'Keller Williams'"
//   - Agents with designations: entity="Member", filter="MemberDesignation has 'GRI'"
//
// 3. Market Analysis Examples:
//   - Price trends: entity="Property", select="ListPrice,CloseDate,City", filter="StandardStatus eq 'Closed'", orderby="CloseDate desc"
//   - Days on market analysis: entity="Dom", select="DaysOnMarket,CumulativeDaysOnMarket", orderby="DaysOnMarket desc"
//
// 4. Media and Marketing:
//   - Property photos: entity="Media", filter="ResourceRecordKey eq 'LISTING_KEY' and MediaCategory eq 'Photo'"
//   - Virtual tours: entity="Media", filter="MediaCategory eq 'BrandedVirtualTour'"
//   - Property with photos: entity="Property", expand="Media($filter=Permission ne 'Private')"
//
// 5. Advanced Expand Queries:
//   - Property with public photos: entity="Property", expand="Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc)"
//   - Property with open houses: entity="Property", expand="OpenHouse($filter=OpenHouseStartTime gt now())"
//   - Complete property package: entity="Property", expand="Media($filter=Permission ne 'Private'),OpenHouse,Dom"
//
// StandardStatus Values: Active, ActiveUnderContract, Canceled, Closed, ComingSoon, Delete, Expired, Hold, Incomplete, Pending, Withdrawn, OffMarket
// PropertyType Values: Residential, ResidentialIncome, ResidentialLease, CommercialSale, CommercialLease, BusinessOpportunity, Farm, Land, ManufacturedInPark
// PropertySubType Values: SingleFamilyResidence, Condominium, Townhouse, Duplex, Triplex, Quadruplex, ManufacturedHome, Farm, UnimprovedLand, etc.
// MediaCategory Values: Photo, Video, BrandedVideo, UnbrandedVideo, BrandedVirtualTour, UnbrandedVirtualTour, FloorPlan, Document
// Permission Values: Public (MediaURL available), Private (MediaURL not available)
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
		Description: "Query the RESO (Real Estate Standards Organization) API for comprehensive real estate data. This tool provides access to MLS (Multiple Listing Service) data including property listings, agent information, office details, media files, and market analytics. Perfect for real estate research, market analysis, property searches, and lead generation. Supports advanced filtering, sorting, and field selection with standardized RESO field names for consistent data access across different MLS systems.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"entity": map[string]interface{}{
					"type":        "string",
					"description": "RESO Entity to query. Choose based on your data needs:\n\n• **Property** - Primary real estate listings with comprehensive property details (address, price, features, status, agent info, etc.). Use for: searching homes, analyzing market data, getting listing details. Key fields: ListingKey, StandardStatus, ListPrice, PropertyType, PropertySubType, StreetNumber, City, StateOrProvince, PostalCode, BedroomsTotal, BathroomsTotal, LivingArea, YearBuilt, ListAgentFullName, PublicRemarks.\n\n• **Member** - MLS agents/members with contact information and credentials. Use for: finding agent details, contact information, professional designations. Key fields: MemberMlsId, MemberFullName, MemberEmail, MemberDirectPhone, OfficeKey, MemberDesignation.\n\n• **Office** - Real estate offices/brokerages. Use for: finding office information, brokerage details. Key fields: OfficeMlsId, OfficeName, OfficePhone, OfficeEmail, OfficeAddress1, OfficeCity.\n\n• **Media** - Photos, videos, virtual tours, and documents associated with listings. Use for: getting listing media, photos, virtual tours. Key fields: MediaKey, ResourceRecordKey (links to ListingKey), MediaType, MediaCategory, MediaURL, MediaStatus.\n\n• **OpenHouse** - Scheduled open house events. Use for: finding open houses, event scheduling. Key fields: OpenHouseKey, ListingKey, OpenHouseStartTime, OpenHouseEndTime, OpenHouseRemarks.\n\n• **Dom** - Days on Market tracking data. Use for: market timing analysis, DOM calculations. Key fields: ListingId, DaysOnMarket, CumulativeDaysOnMarket.\n\n• **PropertyUnitTypes** - Unit type details for multi-unit properties (apartments, condos). Use for: rental properties, multi-family analysis. Key fields: ListingKey, UnitTypeDescription, UnitTypeBedsTotal, UnitTypeBathsTotal, UnitTypeActualRent.\n\n• **PropertyRooms** - Detailed room-by-room information. Use for: detailed property layouts, room specifications. Key fields: ListingKey, RoomType, RoomDimensions, RoomFeatures, RoomLevel.\n\n• **RawMlsProperty** - Raw MLS data fields (original unprocessed data). Use for: accessing MLS-specific fields not in standardized Property entity.",
					"enum": []string{
						"Property", "Member", "Office", "Media", "OpenHouse",
						"Dom", "PropertyUnitTypes", "PropertyRooms", "RawMlsProperty",
					},
				},
				"select": map[string]interface{}{
					"type":        "string",
					"description": "Comma-separated list of fields to return. Leave empty to get all available fields. For Property entity, common fields include:\n• **Identifiers**: ListingKey, ListingId, MlsStatus\n• **Address**: StreetNumber, StreetName, City, StateOrProvince, PostalCode, UnparsedAddress\n• **Pricing**: ListPrice, ClosePrice, OriginalListPrice, PreviousListPrice\n• **Property Details**: PropertyType, PropertySubType, BedroomsTotal, BathroomsTotal, LivingArea, YearBuilt, LotSizeSquareFeet\n• **Status & Dates**: StandardStatus, OnMarketTimestamp, ModificationTimestamp, DaysOnMarket\n• **Agent Info**: ListAgentFullName, ListAgentEmail, ListAgentDirectPhone, ListOfficeName\n• **Features**: PublicRemarks, Appliances, Heating, Cooling, ParkingFeatures, ExteriorFeatures\n• **Location**: Latitude, Longitude, MLSAreaMajor, MLSAreaMinor, SchoolDistrict\nExample: 'ListingKey,StandardStatus,ListPrice,BedroomsTotal,City,PublicRemarks'",
				},
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "OData filter expression for querying data. Supports comparison operators (eq, ne, gt, ge, lt, le), collection operators (has, in), and logical operators (and, or, not). Common Property filters:\n\n**Status Filters**:\n• Active listings: \"StandardStatus eq 'Active'\"\n• Recently sold: \"StandardStatus eq 'Closed' and CloseDate ge 2024-01-01\"\n• Under contract: \"StandardStatus eq 'Pending'\"\n\n**Price Filters**:\n• Price range: \"ListPrice ge 200000 and ListPrice le 500000\"\n• Luxury properties: \"ListPrice gt 1000000\"\n\n**Property Features**:\n• Bedrooms: \"BedroomsTotal ge 3\"\n• Bathrooms: \"BathroomsTotal ge 2\"\n• Square footage: \"LivingArea gt 2000\"\n• Year built: \"YearBuilt ge 2000\"\n\n**Location Filters**:\n• By city: \"City eq 'Seattle'\"\n• By state: \"StateOrProvince eq 'WA'\"\n• By zip: \"PostalCode eq '98101'\"\n• By area: \"MLSAreaMajor eq 'Downtown'\"\n\n**Property Type**:\n• Single family: \"PropertySubType eq 'SingleFamilyResidence'\"\n• Condos: \"PropertySubType eq 'Condominium'\"\n• Multi-family: \"PropertyType eq 'ResidentialIncome'\"\n\n**Complex Examples**:\n• \"StandardStatus eq 'Active' and PropertySubType eq 'Condominium' and ListPrice le 400000 and City eq 'Bellevue'\"\n• \"StandardStatus eq 'Closed' and CloseDate ge 2024-01-01 and PropertyType eq 'Residential'\"\n\nNote: Use single quotes for string values, proper date formats (YYYY-MM-DD), and combine with 'and'/'or' operators.",
				},
				"top": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of records to return in this request. Use smaller values (10-50) for quick searches, larger values (100-1000) for comprehensive data analysis. Default: 10, Maximum: 1000. For large datasets, use pagination with 'skip' parameter.",
					"minimum":     1,
					"maximum":     1000,
				},
				"skip": map[string]interface{}{
					"type":        "integer",
					"description": "Number of records to skip for pagination. Used with 'top' to implement paging through large result sets. Skip limits vary by entity: Property (1M), Office/Member (500K), Media/Rooms (50K). Example: skip=0&top=100 for first page, skip=100&top=100 for second page.",
					"minimum":     0,
				},
				"orderby": map[string]interface{}{
					"type":        "string",
					"description": "Sort order for results. Format: 'FieldName [asc|desc]'. Multiple fields supported with comma separation. Common patterns:\n• **Price sorting**: 'ListPrice desc' (high to low), 'ListPrice asc' (low to high)\n• **Date sorting**: 'ModificationTimestamp desc' (newest first), 'OnMarketTimestamp desc'\n• **Location sorting**: 'City asc, ListPrice desc'\n• **Size sorting**: 'LivingArea desc, BedroomsTotal desc'\nDefault direction is ascending if not specified. Examples: 'ListPrice desc', 'City asc, ModificationTimestamp desc'",
				},
				"expand": map[string]interface{}{
					"type":        "string",
					"description": "OData expand clause to include related entities in the response. This powerful feature allows fetching related data in a single query instead of multiple API calls. Common expansions:\n\n**Property Entity Expansions**:\n• **Media**: 'Media' - Include all photos/videos/virtual tours\n• **Media (public only)**: 'Media($filter=Permission ne \\'Private\\')' - Exclude private images\n• **Media (photos only)**: 'Media($filter=MediaCategory eq \\'Photo\\')' - Only photos\n• **OpenHouse**: 'OpenHouse' - Include open house events\n• **Dom**: 'Dom' - Include days on market data\n• **PropertyRooms**: 'PropertyRooms' - Include room details\n• **PropertyUnitTypes**: 'PropertyUnitTypes' - Include unit type data\n\n**Multiple Expansions**: Use comma separation: 'Media,OpenHouse,Dom'\n\n**Filtered Expansions**: Apply filters to expanded entities:\n• 'Media($filter=MediaCategory eq \\'Photo\\' and Permission ne \\'Private\\';$orderby=Order asc)'\n• 'OpenHouse($filter=OpenHouseStartTime gt now())'\n\n**Performance Note**: Expanding large related datasets (like Media) may impact response time. Use filters and selection within expansions to optimize performance.\n\nExample: 'Media($select=MediaURL,MediaCategory,Order;$filter=Permission ne \\'Private\\';$orderby=Order asc)'",
				},
				"ignorenulls": map[string]interface{}{
					"type":        "boolean",
					"description": "When true, excludes fields with null/empty values from the response to reduce payload size and improve readability. Recommended for most queries unless you specifically need to see which fields are empty. Default: true.",
					"default":     true,
				},
				"ignorecase": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable case-insensitive text matching for string comparisons in filters. Useful when searching for cities, agent names, or other text fields where case might vary. Example: with ignorecase=true, \"City eq 'seattle'\" will match 'Seattle', 'SEATTLE', etc. Default: false.",
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

	// Optional: expand
	if expand, ok := args["expand"].(string); ok {
		params.Expand = strings.TrimSpace(expand)
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
