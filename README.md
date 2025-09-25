# Constellation1 MCP Server

A Model Context Protocol (MCP) server for accessing the Constellation1 RESO (Real Estate Standards Organization) API. This server provides a single tool for querying real estate data with comprehensive filtering and selection options.

## Features

- **OAuth2 Authentication**: Secure client credentials flow
- **Comprehensive Querying**: Support for all OData operators and boolean logic
- **Multiple Entities**: Access to Property, Member, Office, Media, OpenHouse, Dom, PropertyUnitTypes, PropertyRooms, and RawMlsProperty
- **Pagination Support**: Both client-side and server-side pagination
- **Case-Insensitive Search**: For supported text fields
- **Response Optimization**: Null field exclusion to reduce payload size

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
    "reso": {
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

The server provides a single tool: `reso_query`

### Tool Parameters

- **entity** (required): Entity type to query
  - `Property` - Real Estate Listing
  - `Member` - MLS Agent
  - `Office` - MLS Office
  - `Media` - Photos, Virtual Tours, etc.
  - `OpenHouse` - Open House Events
  - `Dom` - Days on Market
  - `PropertyUnitTypes` - Property Units
  - `PropertyRooms` - Property Rooms
  - `RawMlsProperty` - Raw Mls Data Fields

- **select** (optional): Comma-separated list of fields to return
  - Example: `"ListingKey,StreetNumberNumeric,StandardStatus"`

- **filter** (optional): OData filter expression
  - Comparison operators: `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `has`, `in`
  - Boolean operators: `and`, `or`
  - Example: `"StandardStatus eq 'Active' and ListPrice gt 100000"`

- **top** (optional): Number of records to return (default: 10, max: 1000)

- **skip** (optional): Number of records to skip for pagination

- **orderby** (optional): Field(s) to order results by
  - Example: `"ListPrice desc"` or `"ModificationTimestamp"`

- **ignorenulls** (optional): Exclude null fields (default: true)

- **ignorecase** (optional): Case-insensitive text searches (default: false)

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

