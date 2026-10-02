package postcode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// LookupLevel specifies the depth of attributes returned by a lookup request.
// Levels 1 is free; Levels 2-4 require commercial credits; Level 5 is restricted.
type LookupLevel int

const (
	LevelDefault LookupLevel = 0 // Default to level 1
	Level1       LookupLevel = 1 // Free: Administrative boundary
	Level2       LookupLevel = 2 // Commercial: Street and locality names
	Level3       LookupLevel = 3 // Commercial: Building unit metadata
	Level4       LookupLevel = 4 // Commercial: Additional building use info
	Level5       LookupLevel = 5 // Restricted: High-precision point geometry
)

// AdministrativeAddress details administrative boundaries up to area level (L2+).
type AdministrativeAddress struct {
	State        string `json:"state,omitempty"`
	StateName    string `json:"state_name,omitempty"`
	LGA          string `json:"lga,omitempty"`
	LGAName      string `json:"lga_name,omitempty"`
	District     string `json:"district,omitempty"`
	DistrictName string `json:"district_name,omitempty"`
	Area         string `json:"area,omitempty"`
	AreaName     string `json:"area_name,omitempty"`
	Unit         string `json:"unit,omitempty"`
	Zone         string `json:"zone,omitempty"`
	LocalityName string `json:"locality_name,omitempty"`
}

// RecentHouseAddress contains residential and street metadata (L2+).
type RecentHouseAddress struct {
	Address      string `json:"address,omitempty"`
	Recent       string `json:"recent,omitempty"`
	HouseNumber  string `json:"house_number,omitempty"`
	StreetName   string `json:"street_name,omitempty"`
	LocalityName string `json:"locality_name,omitempty"`
}

// PointGeometry represents GeoJSON point coordinates (L5).
type PointGeometry struct {
	Type        string    `json:"type,omitempty"`
	Coordinates []float64 `json:"coordinates,omitempty"` // [lng, lat]
}

// LookupResponse represents the graded attributes returned by GET /v1/lookup.
type LookupResponse struct {
	Postcode              string                 `json:"postcode"`
	Valid                 bool                   `json:"valid"`
	AdministrativeAddress *AdministrativeAddress `json:"administrative_address,omitempty"`
	RecentHouseAddress    *RecentHouseAddress    `json:"recent_house_address,omitempty"`
	BuildingUseStatus     string                 `json:"building_use_status,omitempty"`
	OtherBuildingInfo     map[string]any         `json:"other_building_info,omitempty"`
	PointGeometry         *PointGeometry         `json:"point_geometry,omitempty"`
}

// Lookup queries the gateway for detailed attributes of a postcode.
// Level defaults to Level1 (free) if 0 or unspecified.
func (c *Client) Lookup(ctx context.Context, code string, level LookupLevel) (*LookupResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("postcode: lookup code cannot be empty")
	}

	q := make(url.Values, 2)
	q.Set("code", code)
	if level > 0 {
		q.Set("level", strconv.Itoa(int(level)))
	}

	var res LookupResponse
	endpoint := []string{"v1", "lookup"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// AutocompleteSuggestion represents an individual suggestion returned by autocomplete.
type AutocompleteSuggestion struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// AutocompleteResponse contains suggestions for the currently active segment.
type AutocompleteResponse struct {
	Segment     string                   `json:"segment"`
	Suggestions []AutocompleteSuggestion `json:"suggestions"`
}

// Autocomplete retrieves segment-aware suggestions for a partial postcode input (e.g. "EK 01 A").
func (c *Client) Autocomplete(ctx context.Context, query string) (*AutocompleteResponse, error) {
	q := make(url.Values, 1)
	q.Set("q", query)

	var res AutocompleteResponse
	endpoint := []string{"v1", "search", "autocomplete"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// NearbyParams configures location search around a geographic coordinate.
type NearbyParams struct {
	Latitude  float64
	Longitude float64
	RadiusM   float64 // Search radius in meters (default 300, max 300)
}

// NearbyUnit represents a building unit found near a coordinate.
type NearbyUnit struct {
	Postcode   string  `json:"postcode"`
	Display    string  `json:"display"`
	DistanceM  float64 `json:"distance_m"`
	Confidence string  `json:"confidence"`
	StateName  string  `json:"state_name,omitempty"`
	LGAName    string  `json:"lga_name,omitempty"`
	Address    string  `json:"address,omitempty"`
}

// NearbyResponse contains location units within the requested radius.
type NearbyResponse struct {
	Results []NearbyUnit `json:"results"`
}

// Nearby searches for active postcode units within a radius (default 300m) of a coordinate.
func (c *Client) Nearby(ctx context.Context, params NearbyParams) (*NearbyResponse, error) {
	q := make(url.Values, 3)
	q.Set("lat", strconv.FormatFloat(params.Latitude, 'f', 6, 64))
	q.Set("lng", strconv.FormatFloat(params.Longitude, 'f', 6, 64))
	if params.RadiusM > 0 {
		q.Set("radius", strconv.FormatFloat(params.RadiusM, 'f', 1, 64))
	}

	var res NearbyResponse
	endpoint := []string{"v1", "search", "nearby"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ReverseParams configures reverse geocoding to snap coordinates to the nearest unit.
type ReverseParams struct {
	Latitude      float64
	Longitude     float64
	MaxDistanceM  float64 // Search radius in meters (default 25, max 250)
}

// ReverseResponse represents the result of reverse geocoding a coordinate.
type ReverseResponse struct {
	Found      bool        `json:"found"`
	Coordinate []float64   `json:"coordinate"` // [lng, lat]
	Unit       *NearbyUnit `json:"unit,omitempty"`
	Area       string      `json:"area,omitempty"`
	District   string      `json:"district,omitempty"`
	State      string      `json:"state,omitempty"`
	Message    string      `json:"message,omitempty"`
	RadiusM    float64     `json:"radius_m"`
}

// Reverse resolves a geographic coordinate to the nearest active postcode unit.
func (c *Client) Reverse(ctx context.Context, params ReverseParams) (*ReverseResponse, error) {
	q := make(url.Values, 3)
	q.Set("lat", strconv.FormatFloat(params.Latitude, 'f', 6, 64))
	q.Set("lng", strconv.FormatFloat(params.Longitude, 'f', 6, 64))
	if params.MaxDistanceM > 0 {
		q.Set("max_distance_m", strconv.FormatFloat(params.MaxDistanceM, 'f', 1, 64))
	}

	var res ReverseResponse
	endpoint := []string{"v1", "search", "reverse"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Segments represents the 5 administrative components of a Nigerian postcode.
type Segments struct {
	State    string `json:"state"`
	LGA      string `json:"lga"`
	District string `json:"district"`
	Area     string `json:"area"`
	Unit     string `json:"unit"`
}

// AssembledPostcode represents the output of POST /v1/assembly/assemble.
type AssembledPostcode struct {
	Postcode string `json:"postcode"` // Canonical: "EK-01-A03-FK-01"
	Display  string `json:"display"`  // Spaced: "EK 01 A03 FK 01"
	Compact  string `json:"compact"`  // Compact: "EK01A03FK01"
}

// Assemble builds canonical, display, and compact postcodes from individual segments.
func (c *Client) Assemble(ctx context.Context, segs Segments) (*AssembledPostcode, error) {
	var res AssembledPostcode
	endpoint := []string{"v1", "assembly", "assemble"}
	if err := c.execute(ctx, http.MethodPost, endpoint, nil, segs, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Disassemble decomposes a postcode string into its 5 structural segments.
func (c *Client) Disassemble(ctx context.Context, code string) (*Segments, error) {
	q := make(url.Values, 1)
	q.Set("code", code)

	var res Segments
	endpoint := []string{"v1", "assembly", "disassemble"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}