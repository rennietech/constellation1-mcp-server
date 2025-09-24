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
			Description: "Real Estate Listing",
			URL:         "/odata/Property",
		},
		{
			Name:        "Member",
			Description: "MLS Agent",
			URL:         "/odata/Member",
		},
		{
			Name:        "Office",
			Description: "MLS Office",
			URL:         "/odata/Office",
		},
		{
			Name:        "Media",
			Description: "Photos, Virtual Tours, etc.",
			URL:         "/odata/Media",
		},
		{
			Name:        "OpenHouse",
			Description: "Open House Events",
			URL:         "/odata/OpenHouse",
		},
		{
			Name:        "Dom",
			Description: "Days on Market",
			URL:         "/odata/Dom",
		},
		{
			Name:        "PropertyUnitTypes",
			Description: "Property Units",
			URL:         "/odata/PropertyUnitTypes",
		},
		{
			Name:        "PropertyRooms",
			Description: "Property Rooms",
			URL:         "/odata/PropertyRooms",
		},
		{
			Name:        "RawMlsProperty",
			Description: "Raw Mls Data Fields",
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
