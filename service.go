package postcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
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

// ResolveLocation resolves rich geographic metadata for a postcode.
// It enriches the location via the NIPOST gateway at the highest available commercial
// tier (Level 3 down to Level 1) if configured, adopting any explicit point geometry
// returned by the gateway (e.g. enterprise accounts or simulator).
// If point geometry is not present, it leverages the SDK's high-precision offline reference
// geocoding engine (building benchmarks -> LGA centroid -> State centroid) to guarantee
// reliable coordinates and mapping URLs.
func (c *Client) ResolveLocation(ctx context.Context, code string) (*Location, error) {
	p, err := Parse(code)
	if err != nil {
		return nil, fmt.Errorf("postcode: invalid code %q: %w", code, err)
	}

	loc := p.Location()

	// 1. Query gateway at available commercial tiers (L3 down to L1 if authenticated, L1 only if unauthenticated)
	var res *LookupResponse
	var errLookup error
	lookupTiers := []LookupLevel{Level3, Level2, Level1}
	if c.apiKey == "" {
		lookupTiers = []LookupLevel{Level1}
	}
	for _, lvl := range lookupTiers {
		res, errLookup = c.Lookup(ctx, p.Formatted(), lvl)
		if errLookup == nil && res != nil && res.Valid {
			break
		}
	}

	if errLookup != nil && c.agentGuard != nil && c.agentGuard.cfg.AutoFallbackToOffline {
		c.agentGuard.recordFallback()
	}

	if res != nil && res.Valid {
		// If gateway provides explicit point geometry (granted enterprise keys or simulator), adopt it
		if res.PointGeometry != nil && len(res.PointGeometry.Coordinates) >= 2 {
			loc.Longitude = res.PointGeometry.Coordinates[0]
			loc.Latitude = res.PointGeometry.Coordinates[1]
			loc.Precision = PrecisionBuilding
		}
		if res.RecentHouseAddress != nil && res.RecentHouseAddress.Address != "" {
			loc.Address = res.RecentHouseAddress.Address
		} else if res.RecentHouseAddress != nil && res.RecentHouseAddress.Recent != "" {
			loc.Address = res.RecentHouseAddress.Recent
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

	// 2. Multi-stage geocoding fallback: If gateway returned address metadata but no point coordinates,
	// invoke the configured external geocoder (Google Maps -> Nominatim) to resolve exact building coordinates.
	if loc.Precision != PrecisionBuilding && c.geocoder != nil {
		queryParts := make([]string, 0, 5)
		if loc.Address != "" {
			queryParts = append(queryParts, loc.Address)
		}
		if loc.LGAName != "" && loc.LGAName != loc.Address {
			queryParts = append(queryParts, loc.LGAName)
		}
		if loc.StateName != "" {
			queryParts = append(queryParts, loc.StateName)
		}
		queryParts = append(queryParts, "Nigeria")
		query := strings.Join(queryParts, ", ")

		if len(queryParts) > 1 { // Has at least LGA/State and Nigeria
			if geoLat, geoLng, geoErr := c.geocoder.Geocode(ctx, query); geoErr == nil && (geoLat != 0 || geoLng != 0) {
				loc.Latitude = geoLat
				loc.Longitude = geoLng
				loc.Precision = PrecisionBuilding
			}
		}
	}

	// 3. Cache building-level resolution locally if exact building coordinates are present
	if loc.Precision == PrecisionBuilding {
		RegisterKnownBuilding(BuildingRecord{
			Postcode:  loc.Postcode,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Address:   loc.Address,
			StateCode: loc.StateCode,
			StateName: loc.StateName,
			LGACode:   loc.LGACode,
			LGAName:   loc.LGAName,
			Zone:      loc.Zone,
		})
		_ = SaveDiskCache()
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

// NearbyParams configures location search around a geographic coordinate or reference postcode.
type NearbyParams struct {
	Postcode  string  `json:"postcode,omitempty"` // Optional reference postcode (auto-resolves centroid if coordinates are 0)
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusM   float64 `json:"radius_m"`           // Search radius in meters (default 300, max 300)
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

// UnmarshalJSON supports both a direct slice of units `[...]` (returned by live NIPOST Gateway)
// and an object payload `{"results": [...]}` (returned by simulators/proxies).
func (r *NearbyResponse) UnmarshalJSON(data []byte) error {
	var units []NearbyUnit
	if err := json.Unmarshal(data, &units); err == nil {
		r.Results = units
		return nil
	}

	type alias NearbyResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Results = a.Results
	return nil
}

// Nearby searches for active postcode units within a radius (default 300m) of a coordinate or reference postcode.
func (c *Client) Nearby(ctx context.Context, params NearbyParams) (*NearbyResponse, error) {
	if params.Latitude == 0 && params.Longitude == 0 && params.Postcode != "" {
		if resolved, err := c.ResolveLocation(ctx, params.Postcode); err == nil && (resolved.Latitude != 0 || resolved.Longitude != 0) {
			params.Latitude = resolved.Latitude
			params.Longitude = resolved.Longitude
		} else {
			p, err := Parse(params.Postcode)
			if err != nil {
				return nil, fmt.Errorf("postcode: nearby invalid reference code: %w", err)
			}
			loc := p.Location()
			params.Latitude = loc.Latitude
			params.Longitude = loc.Longitude
		}
	}

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

// NearbyPostcode searches for active units within a radius of a reference postcode.
func (c *Client) NearbyPostcode(ctx context.Context, code string, radiusM float64) (*NearbyResponse, error) {
	return c.Nearby(ctx, NearbyParams{
		Postcode: code,
		RadiusM:  radiusM,
	})
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

// ReverseCoordinates resolves geographic coordinates directly to the nearest active postcode unit.
func (c *Client) ReverseCoordinates(ctx context.Context, lat, lng, maxDistanceM float64) (*ReverseResponse, error) {
	return c.Reverse(ctx, ReverseParams{
		Latitude:     lat,
		Longitude:    lng,
		MaxDistanceM: maxDistanceM,
	})
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
// Numeric segments are zero-filled (e.g. "1" -> "01") and alpha segments are converted to uppercase.
func (c *Client) Assemble(ctx context.Context, segs Segments) (*AssembledPostcode, error) {
	normalized := NormalizeSegments(segs)
	var res AssembledPostcode
	endpoint := []string{"v1", "assembly", "assemble"}
	if err := c.execute(ctx, http.MethodPost, endpoint, nil, normalized, &res); err != nil {
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

// Health checks the operational status of the gateway via GET /healthz.
// Returns nil if the gateway responds with HTTP 200 OK.
func (c *Client) Health(ctx context.Context) error {
	endpoint := []string{"healthz"}
	return c.execute(ctx, http.MethodGet, endpoint, nil, nil, nil)
}

// NormalizeSegments trims whitespace, converts alpha segments to uppercase,
// and zero-fills single-digit numeric segments (e.g. "1" -> "01") according to official NIPOST rules.
func NormalizeSegments(segs Segments) Segments {
	s := strings.ToUpper(strings.TrimSpace(segs.State))
	lga := strings.TrimSpace(segs.LGA)
	if len(lga) == 1 && unicode.IsDigit(rune(lga[0])) {
		lga = "0" + lga
	}
	dist := strings.ToUpper(strings.TrimSpace(segs.District))
	area := strings.ToUpper(strings.TrimSpace(segs.Area))
	unit := strings.TrimSpace(segs.Unit)
	if len(unit) == 1 && unicode.IsDigit(rune(unit[0])) {
		unit = "0" + unit
	}
	return Segments{
		State:    s,
		LGA:      lga,
		District: dist,
		Area:     area,
		Unit:     unit,
	}
}

// AssembleSegments validates and combines segments into an AssembledPostcode locally without network calls.
func AssembleSegments(segs Segments) (AssembledPostcode, error) {
	norm := NormalizeSegments(segs)
	raw := norm.State + norm.LGA + norm.District + norm.Area + norm.Unit
	p, err := Parse(raw)
	if err != nil {
		return AssembledPostcode{}, err
	}
	return AssembledPostcode{
		Postcode: p.Formatted(),
		Display:  p.String(),
		Compact:  p.Raw(),
	}, nil
}

// NamedCode represents a geographic administrative unit code with an optional human name.
// For states and LGAs, Name is populated. For districts and areas, Name is omitted or empty.
type NamedCode struct {
	Code string `json:"code"`
	Name string `json:"name,omitempty"`
}

type referenceStatesResponse struct {
	States []NamedCode `json:"states"`
}

func (r *referenceStatesResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.States = list
		return nil
	}
	type alias referenceStatesResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.States = a.States
	return nil
}

type referenceLGAsResponse struct {
	LGAs []NamedCode `json:"lgas"`
}

func (r *referenceLGAsResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.LGAs = list
		return nil
	}
	type alias referenceLGAsResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.LGAs = a.LGAs
	return nil
}

type referenceDistrictsResponse struct {
	Districts []NamedCode `json:"districts"`
}

func (r *referenceDistrictsResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.Districts = list
		return nil
	}
	type alias referenceDistrictsResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Districts = a.Districts
	return nil
}

type referenceAreasResponse struct {
	Areas []NamedCode `json:"areas"`
}

func (r *referenceAreasResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.Areas = list
		return nil
	}
	type alias referenceAreasResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Areas = a.Areas
	return nil
}

// ReferenceStates queries the official precomputed catalog for all 37 Nigerian states (GET /v1/reference/states).
// This endpoint is free and consumes no credits.
func (c *Client) ReferenceStates(ctx context.Context) ([]NamedCode, error) {
	var res referenceStatesResponse
	endpoint := []string{"v1", "reference", "states"}
	if err := c.execute(ctx, http.MethodGet, endpoint, nil, nil, &res); err != nil {
		return nil, err
	}
	return res.States, nil
}

// ReferenceLGAs queries all Local Government Areas for a given state code (GET /v1/reference/lgas?state=...).
// This endpoint is free and consumes no credits.
func (c *Client) ReferenceLGAs(ctx context.Context, state string) ([]NamedCode, error) {
	if state == "" {
		return nil, fmt.Errorf("postcode: state code cannot be empty")
	}
	q := make(url.Values, 1)
	q.Set("state", strings.ToUpper(strings.TrimSpace(state)))
	var res referenceLGAsResponse
	endpoint := []string{"v1", "reference", "lgas"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}
	return res.LGAs, nil
}

// ReferenceDistricts queries all postal district codes under a state and LGA (GET /v1/reference/districts?state=...&lga=...).
// Districts have no human name (code only). Free, no credits consumed.
func (c *Client) ReferenceDistricts(ctx context.Context, state, lga string) ([]NamedCode, error) {
	if state == "" || lga == "" {
		return nil, fmt.Errorf("postcode: state and lga parameters are required")
	}
	lgaNorm := strings.TrimSpace(lga)
	if len(lgaNorm) == 1 && unicode.IsDigit(rune(lgaNorm[0])) {
		lgaNorm = "0" + lgaNorm
	}
	q := make(url.Values, 2)
	q.Set("state", strings.ToUpper(strings.TrimSpace(state)))
	q.Set("lga", lgaNorm)
	var res referenceDistrictsResponse
	endpoint := []string{"v1", "reference", "districts"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}
	return res.Districts, nil
}

// ReferenceAreas queries all postal area codes under a state, LGA, and district (GET /v1/reference/areas?state=...&lga=...&district=...).
// Areas have no human name (code only). Free, no credits consumed.
func (c *Client) ReferenceAreas(ctx context.Context, state, lga, district string) ([]NamedCode, error) {
	if state == "" || lga == "" || district == "" {
		return nil, fmt.Errorf("postcode: state, lga, and district parameters are required")
	}
	lgaNorm := strings.TrimSpace(lga)
	if len(lgaNorm) == 1 && unicode.IsDigit(rune(lgaNorm[0])) {
		lgaNorm = "0" + lgaNorm
	}
	q := make(url.Values, 3)
	q.Set("state", strings.ToUpper(strings.TrimSpace(state)))
	q.Set("lga", lgaNorm)
	q.Set("district", strings.ToUpper(strings.TrimSpace(district)))
	var res referenceAreasResponse
	endpoint := []string{"v1", "reference", "areas"}
	if err := c.execute(ctx, http.MethodGet, endpoint, q, nil, &res); err != nil {
		return nil, err
	}
	return res.Areas, nil
}