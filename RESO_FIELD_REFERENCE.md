# RESO Field Reference Guide

This guide provides AI-friendly reference information for commonly used RESO fields and their valid values, extracted from the constellation1_metadata.xml.

## Core Entity Relationships

- **Property.ListingKey** ↔ **Media.ResourceRecordKey** (Property photos/media)
- **Property.ListingKey** ↔ **OpenHouse.ListingKey** (Open house events)
- **Property.ListingKey** ↔ **Dom.ListingId** (Days on market data)
- **Property.ListingKey** ↔ **PropertyRooms.ListingKey** (Room details)
- **Property.ListingKey** ↔ **PropertyUnitTypes.ListingKey** (Unit types)
- **Member.MemberMlsId** ↔ **Property.ListAgentMlsId** (Agent-Property relationship)
- **Office.OfficeMlsId** ↔ **Member.OfficeMlsId** (Agent-Office relationship)

## StandardStatus (Property Status)

Use these exact values in filters:

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

**Example**: `"StandardStatus eq 'Active'"`

## PropertyType (Primary Property Categories)

- **Residential** - Single-family homes, condos, townhouses
- **ResidentialIncome** - Multi-family investment properties
- **ResidentialLease** - Rental properties
- **CommercialSale** - Commercial properties for sale
- **CommercialLease** - Commercial properties for lease
- **BusinessOpportunity** - Business sales
- **Farm** - Farm and agricultural properties
- **Land** - Vacant land and lots
- **ManufacturedInPark** - Mobile homes in parks

## PropertySubType (Detailed Property Types)

**Residential Subtypes**:
- **SingleFamilyResidence** - Detached single-family homes
- **Condominium** - Condo units
- **Townhouse** - Townhouse units
- **Duplex** - Two-unit properties
- **Triplex** - Three-unit properties
- **Quadruplex** - Four-unit properties
- **ManufacturedHome** - Manufactured/mobile homes
- **Cabin** - Cabin properties

**Commercial/Other**:
- **Office** - Office buildings
- **Retail** - Retail spaces
- **Industrial** - Industrial properties
- **Warehouse** - Warehouse facilities
- **Farm** - Farm properties
- **UnimprovedLand** - Raw land

## AreaUnits

- **SquareFeet** - US standard
- **SquareMeters** - Metric

## StateOrProvince (US States and Canadian Provinces)

US States: AL, AK, AZ, AR, CA, CO, CT, DE, FL, GA, HI, ID, IL, IN, IA, KS, KY, LA, ME, MD, MA, MI, MN, MS, MO, MT, NE, NV, NH, NJ, NM, NY, NC, ND, OH, OK, OR, PA, RI, SC, SD, TN, TX, UT, VT, VA, WA, WV, WI, WY, DC

Canadian Provinces: AB, BC, MB, NB, NF, NS, NT, NU, ON, PE, QC, SK, YT

## MediaCategory (Media Types)

- **Photo** - Property photos
- **Video** - Property videos
- **BrandedVideo** - Branded property videos
- **UnbrandedVideo** - Unbranded videos
- **BrandedVirtualTour** - Branded virtual tours
- **UnbrandedVirtualTour** - Unbranded virtual tours
- **FloorPlan** - Floor plan images
- **Document** - Property documents
- **AgentPhoto** - Agent photos
- **OfficePhoto** - Office photos
- **OfficeLogo** - Office logos

## Permission Field (Media Privacy)

Controls access to media content based on MLS privacy requirements:

- **Public** - Media is publicly accessible, MediaURL field available
- **Private** - Media is private/restricted, MediaURL field will be empty/null

**Important**: Private images are still counted in PhotosCount but cannot be accessed directly.

**Filter Examples**:
- Public images only: `"Permission ne 'Private'"`
- Private images only: `"Permission eq 'Private'"`
- All images with permission info: No filter needed, check Permission field in results

## Image URL Formatting and Dynamic Sizing

When MediaURL is available (public images), you can request different image sizes:

**Predefined Sizes**:
- Thumbnail (150px width): Add `?d=t` to URL
- Small (480px width): Add `?d=s` to URL
- Large (1024px width): Add `?d=l` to URL

**Dynamic Sizing**:
- By width only: Add `?d={width}` (e.g., `?d=600`)
- By width and height: Add `?d={width}x{height}` (e.g., `?d=500x320`)

**Example**:
```
Original: https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg
Thumbnail: https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=t
Custom: https://photos.prod.cirrussystem.net/1307/abc123/image.jpeg?d=800x600
```

## Expand Functionality (Related Entity Fetching)

The `expand` parameter allows fetching related entities in a single API call, reducing the need for multiple requests. This is particularly powerful for Property queries.

### Property Entity Expansions

**Basic Expansions**:
- `"Media"` - All associated photos, videos, documents
- `"OpenHouse"` - All scheduled open house events
- `"Dom"` - Days on market tracking data
- `"PropertyRooms"` - Detailed room information
- `"PropertyUnitTypes"` - Unit type details for multi-family properties

**Filtered Expansions** (Recommended):
- `"Media($filter=Permission ne 'Private')"` - Only public media
- `"Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private')"` - Public photos only
- `"Media($filter=MediaCategory eq 'Photo';$orderby=Order asc)"` - Photos in order
- `"OpenHouse($filter=OpenHouseStartTime gt now())"` - Future open houses only

**Advanced Expansions with Selection**:
- `"Media($select=MediaURL,MediaCategory,Order;$filter=Permission ne 'Private';$orderby=Order asc)"` - Optimized media query
- `"OpenHouse($select=OpenHouseStartTime,OpenHouseEndTime,OpenHouseRemarks)"` - Essential open house info

**Multiple Expansions**:
- `"Media,OpenHouse"` - Both media and open house data
- `"Media($filter=Permission ne 'Private'),Dom"` - Public media and DOM data

### Expand Examples by Use Case

**Property with Photos**:
```
entity: "Property"
filter: "StandardStatus eq 'Active' and City eq 'Seattle'"
expand: "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$orderby=Order asc)"
select: "ListingKey,ListPrice,PublicRemarks,PhotosCount"
```

**Property with Open Houses**:
```
entity: "Property"
filter: "StandardStatus eq 'Active'"
expand: "OpenHouse($filter=OpenHouseStartTime gt now())"
select: "ListingKey,UnparsedAddress,ListPrice"
```

**Complete Property Marketing Package**:
```
entity: "Property"
filter: "ListingKey eq 'SPECIFIC_LISTING_KEY'"
expand: "Media($filter=Permission ne 'Private'),OpenHouse"
```

## Common Filter Patterns

### Price Range Queries
```
"ListPrice ge 200000 and ListPrice le 500000"
"ListPrice gt 1000000"  // Luxury properties
```

### Property Feature Filters
```
"BedroomsTotal ge 3"
"BathroomsTotal ge 2"
"LivingArea gt 2000"
"YearBuilt ge 2000"
"PoolPrivateYN eq true"
"GarageYN eq true"
```

### Location-Based Filters
```
"City eq 'Seattle'"
"StateOrProvince eq 'WA'"
"PostalCode eq '98101'"
"MLSAreaMajor eq 'Downtown'"
```

### Status and Date Filters
```
"StandardStatus eq 'Active'"
"ModificationTimestamp ge 2024-01-01T00:00:00Z"
"OnMarketTimestamp ge 2024-01-01T00:00:00Z"
"DaysOnMarket le 30"
```

### Property Type Combinations
```
"PropertyType eq 'Residential' and PropertySubType eq 'Condominium'"
"PropertyType eq 'ResidentialIncome'"  // Multi-family
"PropertySubType eq 'SingleFamilyResidence'"
```

## Key Property Fields by Category

### Identification
- ListingKey, ListingId, MlsStatus, UniversalPropertyId

### Address & Location
- StreetNumber, StreetName, City, StateOrProvince, PostalCode, UnparsedAddress
- Latitude, Longitude, MLSAreaMajor, MLSAreaMinor, CountyOrParish

### Pricing
- ListPrice, ClosePrice, OriginalListPrice, PreviousListPrice
- TaxAnnualAmount, TaxAssessedValue

### Property Characteristics
- BedroomsTotal, BathroomsTotal, RoomsTotal
- LivingArea, BuildingAreaTotal, LotSizeSquareFeet, LotSizeAcres
- YearBuilt, YearBuiltEffective, Stories, StoriesTotal

### Features & Amenities
- Appliances, Heating, Cooling, Flooring, Roof
- ParkingFeatures, ExteriorFeatures, InteriorFeatures
- SecurityFeatures, PoolFeatures, SpaFeatures

### Agent & Office Information
- ListAgentFullName, ListAgentEmail, ListAgentDirectPhone, ListAgentMlsId
- ListOfficeName, ListOfficePhone, ListOfficeEmail

### Status & Timestamps
- StandardStatus, MlsStatus, StatusChangeTimestamp
- OnMarketTimestamp, ModificationTimestamp, OriginalEntryTimestamp
- DaysOnMarket, CumulativeDaysOnMarket

### Marketing
- PublicRemarks, PrivateRemarks, SyndicationRemarks
- PhotosCount, VirtualTourURLBranded, VirtualTourURLUnbranded

## Boolean Fields (use true/false)

- NewConstructionYN, PoolPrivateYN, GarageYN, BasementYN
- WaterfrontYN, ViewYN, FireplaceYN, AssociationYN
- InternetEntireListingDisplayYN, VOWYN, IDXYN

## Date Format

Use ISO 8601 format for dates: `YYYY-MM-DDTHH:mm:ssZ` or `YYYY-MM-DD` for date-only fields.

Examples:
- `OnMarketTimestamp ge 2024-01-01T00:00:00Z`
- `CloseDate ge 2024-01-01`

## API Performance Best Practices

### Response Optimization
1. **Use ignorenulls=true** (default) - Reduces payload size by excluding empty fields
2. **Select specific fields** - Only request fields you need: `select="ListingKey,ListPrice,City"`
3. **Apply filters** - Narrow results at the server level: `filter="StandardStatus eq 'Active'"`
4. **Leverage compression** - Client automatically requests gzip/deflate compression (Accept-Encoding header)

### Efficient Media Queries
1. **Filter out private images**: `expand="Media($filter=Permission ne 'Private')"`
2. **Limit media types**: `expand="Media($filter=MediaCategory eq 'Photo')"`
3. **Order media properly**: `expand="Media($orderby=Order asc)"`
4. **Select essential media fields**: `expand="Media($select=MediaURL,MediaCategory,Order)"`

### Pagination Strategy
- **Small batches for browsing**: `top=10-50`
- **Large batches for analysis**: `top=100-1000`
- **Monitor skip limits**: Property (1M), Office/Member (500K), Media (50K)
- **Use server-side pagination**: Check for `@odata.nextLink` in responses

### Complex Query Examples

**High-Performance Property Search with Media**:
```
entity: "Property"
filter: "StandardStatus eq 'Active' and PropertySubType eq 'Condominium' and ListPrice le 750000 and City eq 'Seattle'"
select: "ListingKey,ListPrice,BedroomsTotal,BathroomsTotal,LivingArea,UnparsedAddress,PublicRemarks,PhotosCount"
expand: "Media($filter=MediaCategory eq 'Photo' and Permission ne 'Private';$select=MediaURL,Order;$orderby=Order asc;$top=5)"
top: 25
orderby: "ListPrice asc"
ignorenulls: true
```

**Market Analysis Query**:
```
entity: "Property"
filter: "StandardStatus eq 'Closed' and CloseDate ge 2024-01-01 and PropertyType eq 'Residential'"
select: "ListingKey,ClosePrice,CloseDate,BedroomsTotal,LivingArea,City,DaysOnMarket"
expand: "Dom($select=DaysOnMarket,CumulativeDaysOnMarket)"
top: 1000
orderby: "CloseDate desc"
```

## Advanced Features

### Metadata Access
The API provides detailed metadata about available entities and fields:
- **Endpoint**: `https://listings.cdatalabs.com/odata/$metadata`
- **Purpose**: Discover available fields, data types, and entity relationships
- **Authentication**: Required (uses same OAuth2 token)

### Debug Information
Responses can include debug information for troubleshooting:
- **Location**: `debug` field in API response
- **Content**: Underlying search criteria and performance metrics
- **Access**: Hidden by default, available in full API responses

### Response Structure Details
Standard API response includes:
- **@odata.context** - Request annotation and metadata reference
- **@odata.count** - Number of records in current response (affected by `top`)
- **@odata.totalCount** - Total matching records in entire result set
- **value** - Array of actual data records
- **group** - Used for aggregation queries (advanced feature)
- **@odata.nextLink** - Server-side pagination URL for next batch
- **debug** - Query troubleshooting information (when available)

### Navigation Properties
The metadata defines navigation properties for entity relationships:
- **Property → Media**: Photos, videos, documents for a listing
- **Property → OpenHouse**: Open house events for a listing
- **Property → Dom**: Days on market data
- **Property → PropertyRooms**: Room-by-room details
- **Property → PropertyUnitTypes**: Unit type information
- **Member → Office**: Agent's office affiliation
- **Office → Member**: All agents in an office

Use these relationships with the `expand` parameter to fetch related data efficiently.
