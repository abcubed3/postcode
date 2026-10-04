package postcode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// LookupLevel specifies the depth of attributes returned by a lookup request.
// Per official NIPOST documentation (docs.postcode.gov.ng/concepts/lookup-levels):
// Lookup responses are graded into 3 cumulative levels (L1–L3):
//   - Level 1 (Free, Public): Validity status (valid: true/false) and canonical postcode.
//   - Level 2 (Commercial): Administrative address (state, LGA, district, area) and street address.
//   - Level 3 (Commercial): Full building use status (residential, commercial, mixed, etc.) and all lower levels.
//
// Standard commercial API keys are strictly capped at Level 3 (LevelMax).
// Attempting to request a level higher than your organization's granted quota yields a
// 403 Forbidden ("key lacks scope or level").
//
// Note on Level 4 & Level 5:
// In raw backend OpenAPI schemas (docs.postcode.gov.ng/api-reference/graded-postcode-lookup-levels-1–5-cumulative),
// levels 4 and 5 represent unreleased or restricted government tiers:
//   - Level 4: other_building_info (unreleased building metadata)
//   - Level 5: point_geometry (restricted government GIS coordinates)
// Standard commercial keys do NOT have access to L4/L5. The SDK defines Level4 and Level5
// for backward compatibility and private enterprise mock environments.
type LookupLevel int

const (
	LevelDefault LookupLevel = 0 // Default to Level 1
	Level1       LookupLevel = 1 // Free: Validity check and canonical postcode
	Level2       LookupLevel = 2 // Commercial: Administrative boundaries and street address
	Level3       LookupLevel = 3 // Commercial: Building use status and unit metadata (Official Maximum)
	LevelMax     LookupLevel = 3 // Standard commercial maximum level (L3)

	// Restricted / Internal tiers (not accessible with standard API keys):
	Level4 LookupLevel = 4 // Internal/Enterprise: Additional building info
	Level5 LookupLevel = 5 // Restricted: High-precision point geometry (requires special government grant)
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

	requestedLevel := level
	if requestedLevel == 0 {
		requestedLevel = Level1
	}

	cacheKey := fmt.Sprintf("lookup:%s:%d", code, requestedLevel)
	if c.cache != nil {
		if val, ok := c.cache.Get(cacheKey); ok {
			if cached, ok := val.(*LookupResponse); ok {
				return cached, nil
			}
		}
	}

	effectiveLevel := requestedLevel
	if c.agentGuard != nil {
		var err error
		effectiveLevel, err = c.agentGuard.checkAndRecord(requestedLevel)
		if err != nil {
			return nil, err
		}
	}

	effectiveCacheKey := fmt.Sprintf("lookup:%s:%d", code, effectiveLevel)
	if effectiveLevel != requestedLevel && c.cache != nil {
		if val, ok := c.cache.Get(effectiveCacheKey); ok {
			if cached, ok := val.(*LookupResponse); ok {
				return cached, nil
			}
		}
	}

	q := make(url.Values, 2)
	q.Set("code", code)
	if effectiveLevel > 0 {
		q.Set("level", strconv.Itoa(int(effectiveLevel)))
	}

	var res LookupResponse
	endpoint := []string{"v1", "lookup"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.StatusCode == 402 || apiErr.StatusCode == 403) && effectiveLevel > Level1 && c.agentGuard != nil && c.agentGuard.cfg.AutoDowngradeToLevel1 {
			c.agentGuard.recordDowngrade()
			return c.Lookup(ctx, code, Level1)
		}
		return nil, err
	}

	if c.cache != nil {
		c.cache.Set(effectiveCacheKey, &res, 0)
	}

	return &res, nil
}

// ResolveLocation resolves an 11-character Nigerian postcode to its geographic Location
// and Google Maps URL. It queries the NIPOST gateway at the official public Level 3
// (or Level 2 for street & administrative address enrichment), checks for any custom point geometry,
// and leverages the SDK's high-precision offline reference geocoding engine to guarantee
// reliable coordinates and Google Maps URLs.
func (c *Client) ResolveLocation(ctx context.Context, code string) (*Location, error) {
	p, err := Parse(code)
	if err != nil {
		return nil, fmt.Errorf("postcode: invalid code %q: %w", code, err)
	}

	loc := p.Location()

	// 1. Query gateway at official maximum commercial Level 3 (or fallback to Level 2)
	res, errLookup := c.Lookup(ctx, p.Formatted(), Level3)
	if errLookup != nil {
		res, errLookup = c.Lookup(ctx, p.Formatted(), Level2)
	}

	if errLookup != nil && c.agentGuard != nil && c.agentGuard.cfg.AutoFallbackToOffline {
		c.agentGuard.recordFallback()
	}

	if res != nil && res.Valid {
		// If gateway provides explicit point geometry (enterprise/mock extension), adopt it
		if res.PointGeometry != nil && len(res.PointGeometry.Coordinates) >= 2 {
			loc.Longitude = res.PointGeometry.Coordinates[0]
			loc.Latitude = res.PointGeometry.Coordinates[1]
			loc.Precision = PrecisionBuilding
		}
		if res.RecentHouseAddress != nil && res.RecentHouseAddress.Address != "" {
			loc.Address = res.RecentHouseAddress.Address
		}
		if res.AdministrativeAddress != nil {
			if res.AdministrativeAddress.StateName != "" {
				loc.StateName = res.AdministrativeAddress.StateName
			}
			if res.AdministrativeAddress.LGAName != "" {
				loc.LGAName = res.AdministrativeAddress.LGAName
			}
			if res.AdministrativeAddress.Zone != "" {
				loc.Zone = res.AdministrativeAddress.Zone
			}
		}
	}

	return &loc, nil
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