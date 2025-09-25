package api

import (
	"encoding/json"
	"time"
)

// QueryParams represents the parameters for a RESO API query
type QueryParams struct {
	Entity      string `json:"entity"`
	Select      string `json:"select,omitempty"`
	Filter      string `json:"filter,omitempty"`
	Top         int    `json:"top,omitempty"`
	Skip        int    `json:"skip,omitempty"`
	OrderBy     string `json:"orderby,omitempty"`
	Expand      string `json:"expand,omitempty"`
	IgnoreNulls bool   `json:"ignorenulls,omitempty"`
	IgnoreCase  bool   `json:"ignorecase,omitempty"`
}

// APIResponse represents the standard RESO API response structure
type APIResponse struct {
	Context       string                   `json:"@odata.context"`
	Count         int                      `json:"@odata.count"`
	TotalCount    int                      `json:"@odata.totalCount"`
	Value         []map[string]interface{} `json:"value"`
	Group         []map[string]interface{} `json:"group,omitempty"`
	NextLink      string                   `json:"@odata.nextLink,omitempty"`
	Debug         map[string]interface{}   `json:"debug,omitempty"`
	RequestTime   time.Time                `json:"request_time"`
	ResponseTime  time.Duration            `json:"response_time"`
	RequestParams QueryParams              `json:"request_params"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Target  string `json:"target,omitempty"`
		} `json:"details,omitempty"`
	} `json:"error"`
}

// SupportedEntity represents a supported RESO entity
type SupportedEntity struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// GetSupportedEntities returns the list of supported RESO entities
func GetSupportedEntities() []SupportedEntity {
	return []SupportedEntity{
		{
			Name:        "Property",
			Description: "Primary real estate listings entity containing comprehensive property information including address, pricing, features, status, agent details, and property characteristics. The main entity for property searches and market analysis.",
			URL:         "/odata/Property",
		},
		{
			Name:        "Member",
			Description: "MLS members/agents with contact information, credentials, professional designations, and office affiliations. Use for agent research, contact discovery, and professional verification.",
			URL:         "/odata/Member",
		},
		{
			Name:        "Office",
			Description: "Real estate offices and brokerages with contact details, addresses, and organizational information. Use for finding brokerage information and office affiliations.",
			URL:         "/odata/Office",
		},
		{
			Name:        "Media",
			Description: "Property media including photos, videos, virtual tours, floor plans, and documents. Links to Property entity via ResourceRecordKey. Use for accessing listing imagery and marketing materials.",
			URL:         "/odata/Media",
		},
		{
			Name:        "OpenHouse",
			Description: "Scheduled open house events with dates, times, and showing details. Links to Property entity. Use for finding upcoming open houses and event scheduling.",
			URL:         "/odata/OpenHouse",
		},
		{
			Name:        "Dom",
			Description: "Days on Market tracking for properties including cumulative and current DOM calculations. Use for market timing analysis and pricing strategy insights.",
			URL:         "/odata/Dom",
		},
		{
			Name:        "PropertyUnitTypes",
			Description: "Individual unit details for multi-unit properties including rent, bedrooms, bathrooms, and square footage per unit type. Use for rental property analysis and multi-family investments.",
			URL:         "/odata/PropertyUnitTypes",
		},
		{
			Name:        "PropertyRooms",
			Description: "Detailed room-by-room information including dimensions, features, and level location. Use for space planning, detailed property layouts, and room-specific searches.",
			URL:         "/odata/PropertyRooms",
		},
		{
			Name:        "RawMlsProperty",
			Description: "Original, unprocessed MLS data fields as provided by the source MLS system. Use for accessing MLS-specific fields not available in the standardized Property entity.",
			URL:         "/odata/RawMlsProperty",
		},
	}
}

// IsValidEntity checks if the given entity name is supported
func IsValidEntity(entity string) bool {
	entities := GetSupportedEntities()
	for _, e := range entities {
		if e.Name == entity {
			return true
		}
	}
	return false
}

// GetEntitySkipLimit returns the skip limit for a given entity
func GetEntitySkipLimit(entity string) int {
	limits := map[string]int{
		"Property":          1000000,
		"Office":            500000,
		"Media":             50000,
		"OpenHouse":         500000,
		"PropertyRooms":     50000,
		"Dom":               50000,
		"Member":            500000,
		"RawMlsProperty":    50000,
		"PropertyUnitTypes": 50000, // Default assumption
	}

	if limit, exists := limits[entity]; exists {
		return limit
	}
	return 50000 // Default conservative limit
}

// ToJSON converts the response to JSON string
func (r *APIResponse) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
