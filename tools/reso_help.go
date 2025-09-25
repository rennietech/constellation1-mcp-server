package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/rennietech/constellation1-mcp-server/metadata"
)

// ResoHelpTool implements the reso_help MCP tool for accessing RESO field reference and documentation
type ResoHelpTool struct {
	metadataParser *metadata.MetadataParser
	apiClient      APIClientInterface
}

// APIClientInterface defines the interface for API metadata access
type APIClientInterface interface {
	GetMetadata() (string, error)
}

// NewResoHelpTool creates a new RESO help tool
func NewResoHelpTool() *ResoHelpTool {
	return NewResoHelpToolWithAPI(nil)
}

// NewResoHelpToolWithAPI creates a help tool with optional API client for live metadata fetching
func NewResoHelpToolWithAPI(apiClient APIClientInterface) *ResoHelpTool {
	tool := &ResoHelpTool{
		apiClient: apiClient,
	}

	parser := metadata.NewMetadataParser()
	cacheFile := "/tmp/constellation1_metadata.xml"

	// First priority: Check cache file (avoid re-downloading)
	if _, err := os.Stat(cacheFile); err == nil {
		if err := parser.ParseFromFile(cacheFile); err == nil {
			tool.metadataParser = parser
			return tool
		}
	}

	// Second priority: Fetch from API if client is available
	if apiClient != nil {
		if metadataXML, err := apiClient.GetMetadata(); err == nil {
			// Parse the metadata
			if err := parser.ParseFromReader(strings.NewReader(metadataXML)); err == nil {
				tool.metadataParser = parser
				// Cache the metadata for future use
				if err := os.WriteFile(cacheFile, []byte(metadataXML), 0644); err == nil {
					// Successfully cached metadata
				}
				return tool
			}
		}
	}

	// Third priority: Try local files as fallback
	metadataLocations := []string{
		"constellation1_metadata.xml",
		"../constellation1_metadata.xml",
		"../../constellation1_metadata.xml",
	}

	for _, location := range metadataLocations {
		if _, err := os.Stat(location); err == nil {
			if err := parser.ParseFromFile(location); err == nil {
				tool.metadataParser = parser
				return tool
			}
		}
	}

	// If no metadata available, metadataParser will be nil and we'll use fallback content
	return tool
}

// NewResoHelpToolWithMetadata creates a help tool with specific metadata file
func NewResoHelpToolWithMetadata(metadataPath string) *ResoHelpTool {
	tool := &ResoHelpTool{}
	parser := metadata.NewMetadataParser()

	if err := parser.ParseFromFile(metadataPath); err == nil {
		tool.metadataParser = parser
	}

	return tool
}

// HasMetadata returns true if metadata parser is available
func (t *ResoHelpTool) HasMetadata() bool {
	return t.metadataParser != nil
}

// GetEntityGuide returns the dynamic entity guide if metadata is available
func (t *ResoHelpTool) GetEntityGuide() string {
	if t.metadataParser != nil {
		return t.metadataParser.GenerateEntityGuide()
	}
	return ""
}

// GetEnumsGuide returns the dynamic enums guide if metadata is available
func (t *ResoHelpTool) GetEnumsGuide() string {
	if t.metadataParser != nil {
		return t.metadataParser.GenerateEnumsGuide()
	}
	return ""
}

// GetToolDefinition returns the MCP tool definition for the help tool
func (t *ResoHelpTool) GetToolDefinition() MCPTool {
	return MCPTool{
		Name:        "reso_help",
		Description: "Get comprehensive RESO field reference documentation, query examples, and best practices. This tool provides instant access to field guides, entity descriptions, filter patterns, and common use cases for effective RESO API usage.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"topic": map[string]interface{}{
					"type":        "string",
					"description": "Help topic to retrieve. Choose from:\n\n• **entities** - Complete guide to all RESO entities with use cases and key fields (dynamic from metadata when available)\n• **fields** - Field reference organized by category (dynamic from metadata when available)\n• **filters** - Filter pattern examples for all common search scenarios\n• **enums** - Valid enum values for StandardStatus, PropertyType, etc. (dynamic from metadata when available)\n• **expand** - Entity expansion examples for fetching related data\n• **examples** - Complete query examples for common real estate use cases\n• **performance** - Best practices for optimal API performance and response times\n• **images** - Image handling, sizing, and privacy controls for Media entities\n• **metadata** - Shows metadata parsing status and available dynamic content\n• **overview** - Complete overview of all available help topics",
					"enum": []string{
						"entities", "fields", "filters", "enums", "expand",
						"examples", "performance", "images", "metadata", "overview",
					},
				},
			},
			"required": []string{"topic"},
		},
	}
}

// Execute executes the RESO help tool
func (t *ResoHelpTool) Execute(args map[string]interface{}) MCPToolResult {
	// Parse arguments
	topic, ok := args["topic"].(string)
	if !ok {
		return MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "Error: topic parameter is required",
			}},
			IsError: true,
		}
	}

	// Get help content based on topic
	content := t.getHelpContent(topic)
	if content == "" {
		return MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error: Unknown help topic '%s'. Use 'overview' to see all available topics.", topic),
			}},
			IsError: true,
		}
	}

	return MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: content,
		}},
	}
}

// getHelpContent returns help content for the specified topic
func (t *ResoHelpTool) getHelpContent(topic string) string {
	switch strings.ToLower(topic) {
	case "overview":
		return t.getOverviewContent()
	case "entities":
		return t.getEntitiesContent()
	case "fields":
		return t.getFieldsContent()
	case "filters":
		return t.getFiltersContent()
	case "enums":
		return t.getEnumsContent()
	case "expand":
		return t.getExpandContent()
	case "examples":
		return t.getExamplesContent()
	case "performance":
		return t.getPerformanceContent()
	case "images":
		return t.getImagesContent()
	case "metadata":
		return t.getMetadataContent()
	default:
		return ""
	}
}

// getOverviewContent returns overview of all help topics
func (t *ResoHelpTool) getOverviewContent() string {
	return `# RESO Help System Overview

## Available Help Topics

Use reso_help with any of these topics to get detailed information:

### 📋 **entities** - Entity Guide
Complete reference for all RESO entities (Property, Member, Office, Media, OpenHouse, Dom, PropertyUnitTypes, PropertyRooms, RawMlsProperty) with use cases and key fields.

### 🏷️ **fields** - Field Reference
Common RESO fields organized by category: identification, address, pricing, property details, agent info, status & dates, features.

### 🔍 **filters** - Filter Patterns
Comprehensive filter examples for all search scenarios: status, price ranges, location, property features, dates, and complex combinations.

### 📊 **enums** - Enum Values
Complete list of valid values for StandardStatus, PropertyType, PropertySubType, MediaCategory, Permission, StateOrProvince, and other enumerated fields.

### 🔗 **expand** - Entity Expansion
Advanced examples for fetching related entities in single queries (Property+Media, Property+OpenHouse, filtered expansions).

### 💡 **examples** - Query Examples
Ready-to-use query examples for common real estate scenarios: property searches, agent lookup, market analysis, media retrieval.

### ⚡ **performance** - Performance Tips
Best practices for optimal API usage: field selection, filtering strategies, pagination, compression, and payload optimization.

### 🖼️ **images** - Image Handling
Media entity usage, image URL manipulation, dynamic sizing, privacy controls, and thumbnail generation.

### 📊 **metadata** - Metadata Status
Shows metadata parsing status, available dynamic content, and metadata file information.

## Quick Start

For immediate help with common queries, try:
- ` + "`reso_help('examples')`" + ` - Get ready-to-use query patterns
- ` + "`reso_help('filters')`" + ` - Learn filter syntax and patterns
- ` + "`reso_help('entities')`" + ` - Understand which entity to use when

## Integration Note

This help system is built into the MCP server and provides the same information available in the external RESO_FIELD_REFERENCE.md documentation file, but accessible directly through the MCP protocol.`
}

// getEntitiesContent returns entity-specific help content
func (t *ResoHelpTool) getEntitiesContent() string {
	// Use dynamic content if metadata parser is available
	if t.metadataParser != nil {
		return t.metadataParser.GenerateEntityGuide()
	}

	// Fallback to static content if metadata not available
	return `# RESO Entities Guide (Static Fallback)

## Property Entity 🏠
**Purpose**: Primary real estate listings with comprehensive property information
**Use for**: Property searches, market analysis, listing details
**Key Fields**: ListingKey, StandardStatus, ListPrice, PropertyType, PropertySubType, BedroomsTotal, BathroomsTotal, LivingArea, UnparsedAddress, City, StateOrProvince, PublicRemarks
**Relationships**: Links to Media, OpenHouse, Dom, PropertyRooms, PropertyUnitTypes

## Member Entity 👤  
**Purpose**: MLS agents/members with contact and credential information
**Use for**: Agent research, contact discovery, professional verification
**Key Fields**: MemberMlsId, MemberFullName, MemberEmail, MemberDirectPhone, OfficeKey, MemberDesignation, MemberStatus
**Relationships**: Links to Office, associated Properties via ListAgentMlsId

## Media Entity 📸
**Purpose**: Photos, videos, virtual tours, documents for properties
**Use for**: Property imagery, marketing materials, virtual tours
**Key Fields**: MediaKey, ResourceRecordKey (links to Property), MediaType, MediaCategory, MediaURL, Permission, Order
**Special**: Handles privacy controls, supports dynamic image sizing

## OpenHouse Entity 🏡
**Purpose**: Scheduled open house events
**Use for**: Event scheduling, open house discovery
**Key Fields**: OpenHouseKey, ListingKey, OpenHouseStartTime, OpenHouseEndTime, OpenHouseRemarks

## Dom Entity 📈
**Purpose**: Days on Market tracking and calculations
**Use for**: Market timing analysis, pricing strategy insights
**Key Fields**: ListingId, DaysOnMarket, CumulativeDaysOnMarket

*Note: For complete entity information with all fields from metadata, ensure constellation1_metadata.xml is available during server startup.*`
}

// getFieldsContent returns field reference content
func (t *ResoHelpTool) getFieldsContent() string {
	// Use dynamic content if metadata parser is available
	if t.metadataParser != nil {
		return t.metadataParser.GenerateFieldsGuide("Property")
	}

	// Fallback to static content
	return `# RESO Fields by Category (Static Fallback)

*Note: Dynamic field information from metadata not available. Ensure constellation1_metadata.xml is accessible for complete field listings.*

## Common Property Fields

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

### Boolean Features (use true/false)
PoolPrivateYN, GarageYN, BasementYN, WaterfrontYN, ViewYN, FireplaceYN, NewConstructionYN

### Collection Fields (use 'has' operator)
Appliances, Heating, Cooling, ParkingFeatures, ExteriorFeatures, InteriorFeatures`
}

// getFiltersContent returns filter pattern examples
func (t *ResoHelpTool) getFiltersContent() string {
	return `# RESO Filter Patterns

## Basic Syntax
- **Equals**: ` + "`field eq 'value'`" + `
- **Not Equals**: ` + "`field ne 'value'`" + `
- **Greater Than**: ` + "`field gt 100000`" + `
- **Greater/Equal**: ` + "`field ge 100000`" + `
- **Less Than**: ` + "`field lt 500000`" + `
- **Less/Equal**: ` + "`field le 500000`" + `
- **Has (collections)**: ` + "`field has 'value'`" + `
- **In (list)**: ` + "`field in ('val1','val2')`" + `

## Status Filters
` + "```" + `
StandardStatus eq 'Active'
StandardStatus eq 'Closed' and CloseDate ge 2024-01-01
StandardStatus in ('Active','Pending')
StandardStatus ne 'Withdrawn'
` + "```" + `

## Price Range Filters
` + "```" + `
ListPrice ge 200000 and ListPrice le 500000
ListPrice gt 1000000
ListPrice le 300000 and PropertySubType eq 'Condominium'
` + "```" + `

## Location Filters
` + "```" + `
City eq 'Seattle'
StateOrProvince eq 'WA'
PostalCode eq '98101'
City eq 'Bellevue' and StateOrProvince eq 'WA'
MLSAreaMajor eq 'Downtown'
` + "```" + `

## Property Feature Filters
` + "```" + `
BedroomsTotal ge 3
BathroomsTotal ge 2
LivingArea gt 2000
YearBuilt ge 2000
BedroomsTotal ge 3 and BathroomsTotal ge 2 and LivingArea gt 1500
` + "```" + `

## Property Type Filters
` + "```" + `
PropertySubType eq 'SingleFamilyResidence'
PropertySubType eq 'Condominium'
PropertyType eq 'ResidentialIncome'
PropertyType eq 'Residential' and PropertySubType ne 'ManufacturedHome'
` + "```" + `

## Boolean Feature Filters
` + "```" + `
PoolPrivateYN eq true
GarageYN eq true and BasementYN eq true
WaterfrontYN eq true
ViewYN eq true and PoolPrivateYN eq true
` + "```" + `

## Collection Filters (use 'has')
` + "```" + `
Appliances has 'Dishwasher'
Heating has 'CentralAir'
ParkingFeatures has 'Garage'
` + "```" + `

## Date/Time Filters
` + "```" + `
OnMarketTimestamp ge 2024-01-01T00:00:00Z
ModificationTimestamp ge 2024-06-01T00:00:00Z
CloseDate ge 2024-01-01
DaysOnMarket le 30
` + "```" + `

## Complex Combined Filters
` + "```" + `
StandardStatus eq 'Active' and PropertySubType eq 'Condominium' and ListPrice le 400000 and City eq 'Bellevue'

StandardStatus eq 'Closed' and CloseDate ge 2024-01-01 and PropertyType eq 'Residential' and ListPrice gt 500000

StandardStatus eq 'Active' and BedroomsTotal ge 3 and BathroomsTotal ge 2 and LivingArea gt 2000 and PoolPrivateYN eq true
` + "```" + `

## Tips
- Use single quotes for string values
- Use proper date formats (ISO 8601)
- Combine with 'and'/'or' operators
- Case matters unless ignorecase=true is set`
}

// getEnumsContent returns enum values content
func (t *ResoHelpTool) getEnumsContent() string {
	// Use dynamic content if metadata parser is available
	if t.metadataParser != nil {
		return t.metadataParser.GenerateEnumsGuide()
	}

	// Fallback to static content
	return `# RESO Enum Values (Static Fallback)

*Note: Dynamic enum information from metadata not available. Ensure constellation1_metadata.xml is accessible for complete enum listings.*

## StandardStatus Values
Active, ActiveUnderContract, Pending, Closed, Canceled, Expired, Withdrawn, Hold, ComingSoon, OffMarket

## PropertyType Values  
Residential, ResidentialIncome, ResidentialLease, CommercialSale, CommercialLease, BusinessOpportunity, Farm, Land, ManufacturedInPark

## PropertySubType Values
SingleFamilyResidence, Condominium, Townhouse, Duplex, Triplex, Quadruplex, ManufacturedHome, Farm, UnimprovedLand, Office, Retail, Industrial, Warehouse

## MediaCategory Values
Photo, Video, BrandedVideo, UnbrandedVideo, BrandedVirtualTour, UnbrandedVirtualTour, FloorPlan, Document, AgentPhoto, OfficePhoto, OfficeLogo

## Permission Values
Public, Private

## Common State Codes
WA, CA, NY, TX, FL, IL, PA, OH, GA, NC, MI, NJ, VA, WI, AZ, MA, TN, IN, MD, MO, MN, CO, AL, LA, KY, OR, OK, CT, UT, NV, AR, MS, KS, NM, NE, ID, WV, HI, ME, MT, ND, SD, DE, VT, WY, AK, RI, DC

*For complete enum listings with descriptions, ensure metadata file is available.*`
}

// getExpandContent returns expand functionality examples
func (t *ResoHelpTool) getExpandContent() string {
	return `# Entity Expansion Guide

## What is Expand?
The expand parameter allows fetching related entities in a single API call, reducing the need for multiple requests and improving performance.

## Basic Property Expansions

### Include All Media
` + "```" + `
expand: "Media"
` + "```" + `

### Include Public Media Only (Recommended)
` + "```" + `
expand: "Media($filter=Permission ne 'Private')"
` + "```" + `

### Include Photos Only
` + "```" + `
expand: "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc)"
` + "```" + `

### Include Open Houses
` + "```" + `
expand: "OpenHouse"
` + "```" + `

### Include Future Open Houses Only
` + "```" + `
expand: "OpenHouse($filter=OpenHouseStartTime gt now())"
` + "```" + `

## Advanced Expansions

### Multiple Related Entities
` + "```" + `
expand: "Media($filter=Permission ne 'Private'),OpenHouse,Dom"
` + "```" + `

### Optimized Media with Selection
` + "```" + `
expand: "Media($select=MediaURL,MediaCategory,Order;$filter=Permission ne 'Private';$orderby=Order asc;$top=5)"
` + "```" + `

### Complete Property Marketing Package
` + "```" + `
expand: "Media($filter=Permission ne 'Private'),OpenHouse($select=OpenHouseStartTime,OpenHouseEndTime,OpenHouseRemarks),Dom($select=DaysOnMarket)"
` + "```" + `

## Real-World Examples

### Property Listing with Photos
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and City eq 'Seattle'",
  "expand": "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc;$top=8)",
  "select": "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,UnparsedAddress,PublicRemarks,PhotosCount",
  "top": 10
}
` + "```" + `

### Property with Upcoming Open Houses
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active'",
  "expand": "OpenHouse($filter=OpenHouseStartTime gt now();$select=OpenHouseStartTime,OpenHouseEndTime)",
  "select": "ListingKey,UnparsedAddress,ListPrice",
  "top": 20
}
` + "```" + `

## Performance Tips
- **Filter expansions** to avoid large payloads
- **Select specific fields** in expansions
- **Limit expansion results** with $top
- **Order expansion results** for consistency
- **Avoid expanding large datasets** without filters`
}

// getExamplesContent returns comprehensive query examples
func (t *ResoHelpTool) getExamplesContent() string {
	return `# RESO Query Examples

## Property Search Examples

### 1. Active Homes in Seattle
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and City eq 'Seattle'",
  "select": "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,LivingArea,UnparsedAddress,PublicRemarks",
  "orderby": "ListPrice asc",
  "top": 25
}
` + "```" + `

### 2. Luxury Condominiums
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and PropertySubType eq 'Condominium' and ListPrice gt 1000000",
  "select": "ListingKey,ListPrice,BedroomsTotal,LivingArea,City,UnparsedAddress,ListAgentFullName",
  "orderby": "ListPrice desc",
  "top": 50
}
` + "```" + `

### 3. Family Homes with Features
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and BedroomsTotal ge 4 and BathroomsTotal ge 2 and GarageYN eq true",
  "select": "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,LivingArea,UnparsedAddress,YearBuilt",
  "orderby": "ListPrice asc",
  "top": 30
}
` + "```" + `

### 4. Recent Sales Analysis
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Closed' and CloseDate ge 2024-01-01 and PropertyType eq 'Residential'",
  "select": "ListingKey,ClosePrice,CloseDate,BedroomsTotal,LivingArea,City,DaysOnMarket",
  "expand": "Dom($select=DaysOnMarket,CumulativeDaysOnMarket)",
  "orderby": "CloseDate desc",
  "top": 200
}
` + "```" + `

### 5. Property with Marketing Media
` + "```json" + `
{
  "entity": "Property",
  "filter": "ListingKey eq 'SPECIFIC_LISTING_KEY'",
  "expand": "Media($filter=Permission ne 'Private';$orderby=Order asc)",
  "select": "ListingKey,ListPrice,UnparsedAddress,PublicRemarks,PhotosCount"
}
` + "```" + `

## Agent & Office Examples

### 6. Find Agent by Name
` + "```json" + `
{
  "entity": "Member",
  "filter": "MemberFullName eq 'John Smith'",
  "select": "MemberMlsId,MemberFullName,MemberEmail,MemberDirectPhone,OfficeName,MemberDesignation"
}
` + "```" + `

### 7. Agents in Specific Office
` + "```json" + `
{
  "entity": "Member",
  "filter": "OfficeName eq 'Keller Williams'",
  "select": "MemberMlsId,MemberFullName,MemberEmail,MemberDirectPhone",
  "orderby": "MemberLastName asc",
  "top": 50
}
` + "```" + `

### 8. Office Information
` + "```json" + `
{
  "entity": "Office",
  "filter": "OfficeName eq 'Keller Williams'",
  "select": "OfficeMlsId,OfficeName,OfficePhone,OfficeEmail,OfficeAddress1,OfficeCity"
}
` + "```" + `

## Media & Open House Examples

### 9. Property Photos Only
` + "```json" + `
{
  "entity": "Media",
  "filter": "ResourceRecordKey eq 'LISTING_KEY' and MediaCategory eq 'Photo' and Permission ne 'Private'",
  "select": "MediaKey,MediaURL,Order,MediaCategory",
  "orderby": "Order asc",
  "top": 20
}
` + "```" + `

### 10. Upcoming Open Houses
` + "```json" + `
{
  "entity": "OpenHouse",
  "filter": "OpenHouseStartTime gt now()",
  "select": "OpenHouseKey,ListingKey,OpenHouseStartTime,OpenHouseEndTime,OpenHouseRemarks",
  "orderby": "OpenHouseStartTime asc",
  "top": 50
}
` + "```" + `

## Market Analysis Examples

### 11. Price Trends by Area
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Closed' and CloseDate ge 2024-01-01 and MLSAreaMajor eq 'Downtown'",
  "select": "ListingKey,ClosePrice,CloseDate,BedroomsTotal,LivingArea,PropertySubType",
  "orderby": "CloseDate desc",
  "top": 500
}
` + "```" + `

### 12. Days on Market Analysis
` + "```json" + `
{
  "entity": "Dom",
  "filter": "DaysOnMarket gt 0",
  "select": "ListingId,DaysOnMarket,CumulativeDaysOnMarket",
  "orderby": "DaysOnMarket desc",
  "top": 100
}
` + "```" + `

## Performance Optimized Examples

### 13. High-Performance Property Search
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and PropertySubType eq 'Condominium' and ListPrice le 750000 and City eq 'Seattle'",
  "select": "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,LivingArea,UnparsedAddress,PublicRemarks,PhotosCount",
  "expand": "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$select=MediaURL,Order;$orderby=Order asc;$top=3)",
  "orderby": "ListPrice asc",
  "top": 25,
  "ignorenulls": true
}
` + "```" + ``
}

// getPerformanceContent returns performance optimization tips
func (t *ResoHelpTool) getPerformanceContent() string {
	return `# RESO API Performance Best Practices

## Query Optimization

### 1. Field Selection
**Do**: Select only needed fields
` + "```" + `
select: "ListingKey,ListPrice,City,BedroomsTotal"
` + "```" + `

**Don't**: Return all fields unnecessarily
` + "```" + `
// Avoid: no select parameter = all fields returned
` + "```" + `

### 2. Effective Filtering
**Do**: Filter at the server level
` + "```" + `
filter: "StandardStatus eq 'Active' and City eq 'Seattle'"
` + "```" + `

**Don't**: Fetch all data and filter client-side
` + "```" + `
// Avoid: no filter + client-side filtering
` + "```" + `

### 3. Smart Pagination
**Do**: Use appropriate batch sizes
` + "```" + `
top: 25-100  // For interactive browsing
top: 500-1000  // For data analysis
` + "```" + `

**Don't**: Request enormous datasets
` + "```" + `
// Avoid: top: 10000 without good reason
` + "```" + `

## Expand Optimization

### 4. Filtered Expansions
**Do**: Filter expanded entities
` + "```" + `
expand: "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$top=5)"
` + "```" + `

**Don't**: Expand without filters
` + "```" + `
// Avoid: expand: "Media" (could return 50+ images per property)
` + "```" + `

### 5. Selective Expansion Fields
**Do**: Select only needed expansion fields
` + "```" + `
expand: "Media($select=MediaURL,Order,MediaCategory;$filter=Permission ne 'Private')"
` + "```" + `

**Don't**: Return all expansion fields
` + "```" + `
// Avoid: expand: "Media" (returns all media metadata)
` + "```" + `

## Response Optimization

### 6. Null Field Exclusion
**Do**: Use ignorenulls=true (default)
` + "```" + `
ignorenulls: true  // Reduces payload size significantly
` + "```" + `

### 7. Compression
**Automatic**: Server automatically requests compression
- Headers: Accept-Encoding: gzip, deflate, br
- Reduces transfer time by 60-80%

### 8. Optimal Result Ordering
**Do**: Order by relevant fields
` + "```" + `
orderby: "ListPrice asc"  // Price-focused searches
orderby: "ModificationTimestamp desc"  // Recent changes first
` + "```" + `

## Real-World Optimized Queries

### High-Performance Property Search
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and PropertySubType eq 'Condominium' and ListPrice le 750000 and City eq 'Seattle'",
  "select": "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,LivingArea,UnparsedAddress,PublicRemarks,PhotosCount",
  "expand": "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$select=MediaURL,Order;$orderby=Order asc;$top=3)",
  "orderby": "ListPrice asc",
  "top": 25,
  "ignorenulls": true
}
` + "```" + `

### Efficient Market Analysis
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Closed' and CloseDate ge 2024-01-01 and PropertyType eq 'Residential'",
  "select": "ListingKey,ClosePrice,CloseDate,BedroomsTotal,LivingArea,City",
  "expand": "Dom($select=DaysOnMarket,CumulativeDaysOnMarket)",
  "orderby": "CloseDate desc",
  "top": 1000,
  "ignorenulls": true
}
` + "```" + `

## Performance Monitoring

**Response Fields to Watch**:
- ` + "`@odata.count`" + ` - Records returned
- ` + "`@odata.totalCount`" + ` - Total available records
- Response time metadata in tool results

**Pagination Indicators**:
- ` + "`@odata.nextLink`" + ` - More data available
- Use for server-side pagination when client-side skip limits reached`
}

// getImagesContent returns image handling documentation
func (t *ResoHelpTool) getImagesContent() string {
	return `# Image and Media Handling

## Media Entity Structure

**Key Fields**:
- **MediaKey** - Unique media identifier
- **ResourceRecordKey** - Links to Property.ListingKey
- **MediaURL** - Image/video URL (if Permission = 'Public')
- **MediaCategory** - Type of media (Photo, Video, etc.)
- **Permission** - Access level (Public/Private)
- **Order** - Display order for media
- **MediaType** - File format (jpeg, mp4, etc.)

## Privacy Controls

### Permission Field Handling
` + "```" + `
// Public images only (recommended)
filter: "Permission ne 'Private'"

// Private images only (for counting)
filter: "Permission eq 'Private'"

// All images (check Permission field in results)
// No Permission filter needed
` + "```" + `

### PhotosCount vs Available Images
- **PhotosCount** includes both public and private images
- **Actual accessible images** = images with Permission = 'Public'
- Always filter by Permission when fetching image URLs

## Image URL Manipulation

### Original Image
` + "```" + `
https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg
` + "```" + `

### Predefined Sizes
` + "```" + `
// Thumbnail (150px width)
https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=t

// Small (480px width)
https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=s

// Large (1024px width)
https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=l
` + "```" + `

### Dynamic Sizing
` + "```" + `
// Custom width
https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=600

// Custom width x height
https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=800x600
` + "```" + `

## Optimized Media Queries

### Get Property Photos Efficiently
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and PhotosCount gt 0",
  "expand": "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$select=MediaURL,Order;$orderby=Order asc;$top=5)",
  "select": "ListingKey,ListPrice,UnparsedAddress,PhotosCount",
  "top": 10
}
` + "```" + `

### Get All Property Media Types
` + "```json" + `
{
  "entity": "Property",
  "filter": "ListingKey eq 'SPECIFIC_LISTING_KEY'",
  "expand": "Media($filter=Permission ne 'Private';$select=MediaURL,MediaCategory,Order;$orderby=MediaCategory asc,Order asc)"
}
` + "```" + `

### Direct Media Query
` + "```json" + `
{
  "entity": "Media",
  "filter": "ResourceRecordKey eq 'LISTING_KEY' and MediaCategory eq 'Photo' and Permission ne 'Private'",
  "select": "MediaKey,MediaURL,Order,MediaType",
  "orderby": "Order asc",
  "top": 25
}
` + "```" + `

## Virtual Tours and Videos

### Find Virtual Tours
` + "```json" + `
{
  "entity": "Media",
  "filter": "MediaCategory in ('BrandedVirtualTour','UnbrandedVirtualTour') and Permission ne 'Private'",
  "select": "MediaKey,ResourceRecordKey,MediaURL,MediaCategory"
}
` + "```" + `

### Property Marketing Media Package
` + "```json" + `
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and ListingKey eq 'SPECIFIC_KEY'",
  "expand": "Media($filter=Permission ne 'Private';$orderby=MediaCategory asc,Order asc)",
  "select": "ListingKey,ListPrice,UnparsedAddress,PublicRemarks,PhotosCount"
}
` + "```" + `

## Best Practices

1. **Always check Permission** before accessing MediaURL
2. **Use Order field** to display images in correct sequence
3. **Filter by MediaCategory** to get specific media types
4. **Limit expanded media** with $top to control payload size
5. **Request appropriate image sizes** for your use case
6. **Cache image URLs** - they're content-based and stable`
}

// getMetadataContent returns metadata parser status and information
func (t *ResoHelpTool) getMetadataContent() string {
	var content strings.Builder
	content.WriteString("# Metadata Parser Status\n\n")

	if t.metadataParser != nil {
		content.WriteString("✅ **Metadata Parser**: ACTIVE - Dynamic content available\n\n")

		entityNames := t.metadataParser.GetEntityNames()
		enumNames := t.metadataParser.GetEnumNames()

		content.WriteString(fmt.Sprintf("📊 **Entities Loaded**: %d\n", len(entityNames)))
		content.WriteString(fmt.Sprintf("📋 **Enums Loaded**: %d\n\n", len(enumNames)))

		content.WriteString("## Available Entities (from metadata)\n")
		for _, entityName := range entityNames {
			if entity, exists := t.metadataParser.GetEntityInfo(entityName); exists {
				content.WriteString(fmt.Sprintf("- **%s** (%d fields)\n", entityName, len(entity.Properties)))
			}
		}

		content.WriteString("\n## Sample Enum Types (from metadata)\n")
		priorityEnums := []string{"StandardStatus", "PropertyType", "PropertySubType", "MediaCategory", "StateOrProvince"}
		for _, enumName := range priorityEnums {
			if enumInfo, exists := t.metadataParser.GetEnumInfo(enumName); exists {
				content.WriteString(fmt.Sprintf("- **%s** (%d values)\n", enumName, len(enumInfo.Members)))
			}
		}

		content.WriteString("\n## Dynamic Content Available\n")
		content.WriteString("- ✅ `entities` - Generated from actual entity definitions\n")
		content.WriteString("- ✅ `fields` - Generated from actual field definitions with types\n")
		content.WriteString("- ✅ `enums` - Generated from actual enum definitions with standard names\n")
		content.WriteString("- ℹ️ `filters`, `expand`, `examples` - Static content with best practices\n")

	} else {
		content.WriteString("❌ **Metadata Parser**: NOT LOADED - Using static fallback content\n\n")
		content.WriteString("## Metadata File Search Locations\n")
		content.WriteString("The server searched for constellation1_metadata.xml in:\n")
		content.WriteString("- Current directory: `./constellation1_metadata.xml`\n")
		content.WriteString("- Parent directory: `../constellation1_metadata.xml`\n")
		content.WriteString("- Grandparent directory: `../../constellation1_metadata.xml`\n")
		content.WriteString("- Docker path: `/opt/metamcp/constellation1_metadata.xml`\n\n")

		content.WriteString("## Impact of Missing Metadata\n")
		content.WriteString("- ⚠️ `entities` - Using static fallback (may be incomplete)\n")
		content.WriteString("- ⚠️ `fields` - Using static fallback (limited field coverage)\n")
		content.WriteString("- ⚠️ `enums` - Using static fallback (may be outdated)\n")
		content.WriteString("- ✅ `filters`, `expand`, `examples` - Full static content available\n\n")

		content.WriteString("## How to Enable Dynamic Content\n")
		content.WriteString("1. **Ensure valid RESO API credentials** are configured (client_id and client_secret)\n")
		content.WriteString("2. **Restart the MCP server** - it will fetch and cache metadata automatically\n")
		content.WriteString("3. **Cache Management**: Metadata is cached at `/tmp/constellation1_metadata.xml`\n")
		content.WriteString("4. **Force Refresh**: Delete `/tmp/constellation1_metadata.xml` and restart to fetch fresh metadata\n")
	}

	return content.String()
}
