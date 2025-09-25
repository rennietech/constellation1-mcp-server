# Constellation1 MCP Server

A Model Context Protocol (MCP) server for accessing the Constellation1 RESO (Real Estate Standards Organization) API. This server provides a single tool for querying real estate data with comprehensive filtering and selection options.

## Quick Install for Cursor

### Easy Setup (Copy-Paste Configuration)

**Step 1**: Download the binary for your platform:
- [macOS Apple Silicon](https://github.com/rennietech/constellation1-mcp-server/releases/latest/download/constellation1-mcp-server-darwin-arm64)
- [Linux AMD64](https://github.com/rennietech/constellation1-mcp-server/releases/latest/download/constellation1-mcp-server-linux-amd64)

**Step 2**: Make it executable and move to a permanent location:
```bash
# For macOS
chmod +x constellation1-mcp-server-darwin-arm64
mv constellation1-mcp-server-darwin-arm64 /usr/local/bin/

# For Linux
chmod +x constellation1-mcp-server-linux-amd64
mv constellation1-mcp-server-linux-amd64 /usr/local/bin/
```

**Step 3**: Add this configuration to your Cursor MCP settings (`Cursor Settings` → `Features` → `Model Context Protocol`):

**For macOS:**
```json
{
  "constellation1-reso": {
    "command": "/usr/local/bin/constellation1-mcp-server-darwin-arm64",
    "args": [
      "-client-id",
      "your_client_id_here",
      "-client-secret",
      "your_client_secret_here"
    ]
  }
}
```

**For Linux:**
```json
{
  "constellation1-reso": {
    "command": "/usr/local/bin/constellation1-mcp-server-linux-amd64",
    "args": [
      "-client-id",
      "your_client_id_here",
      "-client-secret",
      "your_client_secret_here"
    ]
  }
}
```

Replace `your_client_id_here` and `your_client_secret_here` with your actual RESO API credentials.

## Features

- **OAuth2 Authentication**: Secure client credentials flow with automatic token management
- **Comprehensive Querying**: Full OData support with filtering, sorting, and field selection
- **Entity Expansion**: Fetch related entities in a single query (Property + Media, OpenHouse, etc.)
- **Multiple Entities**: Access to Property, Member, Office, Media, OpenHouse, Dom, PropertyUnitTypes, PropertyRooms, and RawMlsProperty
- **Private Image Handling**: Respects MLS privacy controls with Permission field filtering
- **Dynamic Image Sizing**: Support for thumbnail, small, large, and custom image dimensions
- **Pagination Support**: Both client-side and server-side pagination with configurable limits
- **Case-Insensitive Search**: For supported text fields
- **Response Optimization**: Automatic compression (gzip/deflate/brotli) and null field exclusion
- **Performance Optimized**: Built-in best practices for efficient API usage
- **Dynamic Metadata Parsing**: Automatically loads constellation1_metadata.xml for accurate, up-to-date field information
- **Live Metadata Fallback**: Fetches metadata from API endpoint when local file unavailable  
- **Integrated Documentation**: Built-in help system and MCP resources for field reference and examples
- **MCP Resources Support**: Access documentation directly through MCP protocol
- **Self-Updating Field Reference**: Dynamic content generation from actual RESO metadata

## Installation

### Prerequisites

- Go 1.21 or later (for building from source)
- RESO API credentials (client_id and client_secret)

### Using Pre-built Binaries

Download the appropriate pre-built binary for your platform from the [releases page](https://github.com/rennietech/constellation1-mcp-server/releases):

- **Linux AMD64**: `constellation1-mcp-server-linux-amd64`
- **macOS Apple Silicon (ARM64)**: `constellation1-mcp-server-darwin-arm64`

Make it executable:

```bash
chmod +x constellation1-mcp-server-*
```

## Configuration

The server accepts configuration through command line arguments or environment variables:

### Command Line Arguments (Recommended)

Configure in your MCP client with command line arguments:

```json
{
  "mcpServers": {
    "constellation1-reso": {
      "command": "./constellation1-mcp-server-darwin-arm64",
      "args": [
        "-client-id", "your_client_id_here",
        "-client-secret", "your_client_secret_here"
      ]
    }
  }
}
```

### Environment Variables (Alternative)

```bash
export RESO_CLIENT_ID="your_client_id_here"
export RESO_CLIENT_SECRET="your_client_secret_here"
export RESO_AUTH_URL="https://authenticate.constellation1apis.com/oauth2/token"
export RESO_BASE_URL="https://listings.cdatalabs.com/odata"
```

## Usage

The server provides comprehensive tools and resources:

### 🔧 **Tools Available:**
- **`reso_query`** - Query RESO API for real estate data
- **`reso_help`** - Get field reference, examples, and best practices

### 📚 **Resources Available:**
- **RESO Field Reference Guide** - Comprehensive field and entity documentation
- **RESO Query Quick Start** - Common query patterns and examples

Access resources via MCP resources/list and resources/read methods.

> 📖 **For detailed field reference and examples, see [RESO_FIELD_REFERENCE.md](RESO_FIELD_REFERENCE.md)**

### Tool Parameters

- **entity** (required): RESO Entity type to query
  - `Property` - Primary real estate listings with comprehensive property details
  - `Member` - MLS agents/members with contact information and credentials
  - `Office` - Real estate offices and brokerages
  - `Media` - Property media (photos, videos, virtual tours, documents)
  - `OpenHouse` - Scheduled open house events
  - `Dom` - Days on Market tracking data
  - `PropertyUnitTypes` - Unit details for multi-unit properties
  - `PropertyRooms` - Detailed room-by-room information
  - `RawMlsProperty` - Original unprocessed MLS data fields

- **select** (optional): Comma-separated list of specific fields to return
  - Leave empty to get all available fields
  - Common Property fields: `ListingKey,StandardStatus,ListPrice,BedroomsTotal,City,PublicRemarks`
  - See [RESO_FIELD_REFERENCE.md](RESO_FIELD_REFERENCE.md) for complete field lists

- **filter** (optional): OData filter expression for data querying
  - Status: `"StandardStatus eq 'Active'"`
  - Price range: `"ListPrice ge 200000 and ListPrice le 500000"`
  - Location: `"City eq 'Seattle' and StateOrProvince eq 'WA'"`
  - Features: `"BedroomsTotal ge 3 and BathroomsTotal ge 2"`
  - See [RESO_FIELD_REFERENCE.md](RESO_FIELD_REFERENCE.md) for comprehensive filter examples

- **top** (optional): Maximum records to return (default: 10, max: 1000)
  - Use 10-50 for quick searches, 100-1000 for comprehensive analysis

- **skip** (optional): Records to skip for pagination
  - Limits vary by entity (Property: 1M, Office/Member: 500K, Media: 50K)

- **orderby** (optional): Sort order for results
  - Format: `"FieldName [asc|desc]"`
  - Examples: `"ListPrice desc"`, `"City asc, ModificationTimestamp desc"`

- **expand** (optional): Include related entities in the response
  - Property + Media: `"Media($filter=Permission ne 'Private')"`
  - Property + Open Houses: `"OpenHouse"`
  - Multiple entities: `"Media,OpenHouse,Dom"`
  - See [RESO_FIELD_REFERENCE.md](RESO_FIELD_REFERENCE.md) for comprehensive expand examples

- **ignorenulls** (optional): Exclude null/empty fields to reduce payload size (default: true)

- **ignorecase** (optional): Enable case-insensitive text matching (default: false)

## reso_help Tool

Get instant access to field reference documentation and query examples:

- **topic** (required): Help topic to retrieve
  - `entities` - Complete entity guide with use cases and key fields
  - `fields` - Field reference organized by category
  - `filters` - Filter pattern examples for all search scenarios
  - `enums` - Valid enum values (StandardStatus, PropertyType, etc.)
  - `expand` - Entity expansion examples and best practices
  - `examples` - Ready-to-use query examples
  - `performance` - API performance optimization tips
  - `images` - Image handling and dynamic sizing guide
  - `overview` - Complete overview of all help topics

**Example**: `{"topic": "examples"}` or `{"topic": "filters"}`

## Dynamic Metadata System

The server automatically loads RESO metadata to provide accurate, up-to-date field information:

### 📊 **Metadata Sources** (in priority order):
1. **Local File**: `constellation1_metadata.xml` in working directory
2. **API Endpoint**: Live fetch from `https://listings.constellation1apis.com/$metadata`  
3. **Static Fallback**: Hardcoded essential field information

### 🔄 **Dynamic Content Available:**
- **`reso_help('entities')`** - Generated from actual entity definitions (18 entities, 678+ Property fields)
- **`reso_help('fields')`** - Categorized field reference with types and descriptions
- **`reso_help('enums')`** - Complete enum values with standard names (200+ enums)
- **`reso_help('metadata')`** - Shows current metadata status and statistics

### 📁 **Setup for Full Dynamic Content:**
```bash
# Option 1: Place metadata file in working directory
cp constellation1_metadata.xml /path/to/mcp/server/

# Option 2: Download latest from API (requires credentials)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://listings.constellation1apis.com/\$metadata \
     > constellation1_metadata.xml
```

When metadata is available, help content includes:
- ✅ **All 18 entity types** with actual field counts
- ✅ **Complete enum definitions** with standard names and descriptions
- ✅ **Categorized field lists** (9 categories for Property entity)
- ✅ **Accurate type information** for all fields

### Example Queries

#### Basic Property Search
```json
{
  "entity": "Property",
  "top": 5,
  "select": "ListingKey,StandardStatus,ListPrice,StreetNumberNumeric"
}
```

#### Filtered Property Search
```json
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and ListPrice ge 200000 and ListPrice le 500000",
  "top": 10,
  "orderby": "ListPrice desc",
  "ignorenulls": true
}
```

#### Agent Search
```json
{
  "entity": "Member",
  "filter": "MemberFirstName eq 'John'",
  "ignorecase": true,
  "select": "MemberKey,MemberFirstName,MemberLastName,MemberEmail"
}
```

#### Media Search
```json
{
  "entity": "Media",
  "filter": "ResourceRecordKey eq 'PROPERTY_KEY_HERE'",
  "top": 20
}
```

#### Property with Photos (Using Expand)
```json
{
  "entity": "Property",
  "filter": "StandardStatus eq 'Active' and City eq 'Seattle'",
  "expand": "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc)",
  "select": "ListingKey,ListPrice,PublicRemarks,PhotosCount",
  "top": 5
}
```

#### Complete Property Package
```json
{
  "entity": "Property",
  "filter": "ListingKey eq 'SPECIFIC_LISTING_KEY'",
  "expand": "Media($filter=Permission ne 'Private'),OpenHouse,Dom",
  "ignorenulls": true
}
```

#### Get Help and Examples
```json
{
  "topic": "examples"
}
```

#### Field Reference
```json
{
  "topic": "fields"
}
```

## Supported Filter Operations

### Comparison Operators
- `eq` - Equal: `StandardStatus eq 'Active'`
- `ne` - Not Equal: `StandardStatus ne 'Sold'`
- `gt` - Greater Than: `ListPrice gt 100000`
- `ge` - Greater Than or Equal: `ListPrice ge 100000`
- `lt` - Less Than: `ListPrice lt 500000`
- `le` - Less Than or Equal: `ListPrice le 500000`
- `has` - Has enumerable: `Appliances has 'Dishwasher'`
- `in` - Is member of: `StandardStatus in ('Active','Pending')`

### Boolean Operators
- `and` - Logical AND: `StandardStatus eq 'Active' and ListPrice gt 100000`
- `or` - Logical OR: `PropertyType eq 'Residential' or PropertyType eq 'Condo'`

### String Handling
- Single quotes required for string values
- Case-sensitive by default (use `ignorecase: true` for case-insensitive searches)
- Collections require 'any' expression for filtering

## Pagination

### Client-Side Pagination
Use `top` and `skip` parameters:
```json
{
  "entity": "Property",
  "top": 50,
  "skip": 100
}
```

### Server-Side Pagination
The response includes `@odata.nextLink` for server-side pagination when available.

### Skip Limits by Entity
- Property: 1,000,000 records
- Office: 500,000 records
- Member: 500,000 records
- OpenHouse: 500,000 records
- Media: 50,000 records
- PropertyRooms: 50,000 records
- Dom: 50,000 records
- RawMlsProperty: 50,000 records
- PropertyUnitTypes: 50,000 records

## Response Format

The tool returns a structured response with:
- Summary of the query and results
- Full JSON response from the RESO API
- Metadata including request time and response time
- Pagination information when available

## Error Handling

The server handles various error conditions:
- Authentication failures
- Invalid query parameters
- API rate limiting
- Network connectivity issues
- Malformed responses

## Building from Source

```bash
# Clone or download the source code
cd constellation1-mcp-server

# Initialize Go modules
go mod tidy

# Build for macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o constellation1-mcp-server-darwin-arm64

# Build for Linux AMD64
GOOS=linux GOARCH=amd64 go build -o constellation1-mcp-server-linux-amd64

# Build for other platforms (if needed)
GOOS=windows GOARCH=amd64 go build -o constellation1-mcp-server-windows-amd64.exe
```

## License

This project is provided as-is for integration with RESO standard APIs.

## Support

For issues related to:
- RESO API access: Contact your RESO API provider
- MCP integration: Check MCP client documentation
- Server functionality: Review the error messages and logs

