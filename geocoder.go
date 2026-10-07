package postcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Geocoder resolves a geographic textual address or landmark query to latitude and longitude coordinates.
type Geocoder interface {
	Geocode(ctx context.Context, query string) (lat, lng float64, err error)
}

// GoogleMapsGeocoder resolves coordinates using the modern Google Maps Geocoding API v4
// (https://geocode.googleapis.com/v4/geocode/address), with automatic fallback to v3.
type GoogleMapsGeocoder struct {
	APIKey     string
	BaseURL    string // Defaults to "https://geocode.googleapis.com"
	HTTPClient *http.Client
}

// NewGoogleMapsGeocoder creates a Geocoder backed by Google Maps Geocoding API v4.
func NewGoogleMapsGeocoder(apiKey string, client *http.Client) *GoogleMapsGeocoder {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &GoogleMapsGeocoder{
		APIKey:     apiKey,
		HTTPClient: client,
	}
}

type googleGeocodeV4Response struct {
	Results []struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
		Geometry struct { // Backward-compatibility fallback for v3 responses
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		FormattedAddress string `json:"formattedAddress"`
		PlaceID          string `json:"placeId"`
	} `json:"results"`
	Status string `json:"status,omitempty"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// Geocode queries the Google Maps Geocoding API v4 for the provided Nigerian address query.
func (g *GoogleMapsGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	if g.APIKey == "" {
		return 0, 0, fmt.Errorf("google geocoder: API key is empty")
	}
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return 0, 0, fmt.Errorf("google geocoder: query cannot be empty")
	}

	baseURL := g.BaseURL
	if baseURL == "" {
		baseURL = "https://geocode.googleapis.com"
	}

	// Geocoding API v4 address endpoint:
	// GET https://geocode.googleapis.com/v4/geocode/address/{ADDRESS}?regionCode=NG
	endpoint := fmt.Sprintf("%s/v4/geocode/address/%s?regionCode=NG", baseURL, url.PathEscape(cleanQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("google geocoder v4 request creation failed: %w", err)
	}

	// Geocoding v4 requires authentication via X-Goog-Api-Key header and supports field masking
	req.Header.Set("X-Goog-Api-Key", g.APIKey)
	req.Header.Set("X-Goog-FieldMask", "results.location,results.formattedAddress,results.placeId")

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("google geocoder v4 http call failed: %w", err)
	}
	defer resp.Body.Close()

	// If v4 returns 404 Not Found (e.g., custom mock or legacy endpoint), fall back gracefully to v3
	if resp.StatusCode == http.StatusNotFound && g.BaseURL == "" {
		return g.geocodeV3(ctx, cleanQuery)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		var errRes googleGeocodeV4Response
		if json.Unmarshal(body, &errRes) == nil && errRes.Error != nil {
			return 0, 0, fmt.Errorf("google geocoder v4 returned %d (%s): %s", resp.StatusCode, errRes.Error.Status, errRes.Error.Message)
		}
		return 0, 0, fmt.Errorf("google geocoder v4 returned http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return 0, 0, fmt.Errorf("google geocoder v4 reading body failed: %w", err)
	}

	var res googleGeocodeV4Response
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, 0, fmt.Errorf("google geocoder v4 decoding json failed: %w", err)
	}

	if len(res.Results) == 0 {
		return 0, 0, fmt.Errorf("google geocoder v4: no results found for %q", cleanQuery)
	}

	// In v4, location coordinates are at results[0].location (lat, lng)
	loc := res.Results[0].Location
	if loc.Lat == 0 && loc.Lng == 0 {
		// Fallback to geometry.location if populated
		loc = res.Results[0].Geometry.Location
	}

	if loc.Lat == 0 && loc.Lng == 0 {
		return 0, 0, fmt.Errorf("google geocoder v4: zero coordinates returned for %q", cleanQuery)
	}

	return loc.Lat, loc.Lng, nil
}

// geocodeV3 acts as a fallback to the legacy Geocoding API v3 if the v4 endpoint is unavailable.
func (g *GoogleMapsGeocoder) geocodeV3(ctx context.Context, cleanQuery string) (float64, float64, error) {
	endpoint := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&components=country:NG&key=%s",
		url.QueryEscape(cleanQuery), url.QueryEscape(g.APIKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("google geocoder v3 request creation failed: %w", err)
	}

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("google geocoder v3 http call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("google geocoder v3 returned http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return 0, 0, fmt.Errorf("google geocoder v3 reading body failed: %w", err)
	}

	var res googleGeocodeV4Response
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, 0, fmt.Errorf("google geocoder v3 decoding json failed: %w", err)
	}

	if len(res.Results) == 0 {
		return 0, 0, fmt.Errorf("google geocoder v3: no results found for %q", cleanQuery)
	}

	loc := res.Results[0].Geometry.Location
	if loc.Lat == 0 && loc.Lng == 0 {
		loc = res.Results[0].Location
	}
	return loc.Lat, loc.Lng, nil
}

// NominatimGeocoder resolves coordinates using OpenStreetMap Nominatim with country restricted to Nigeria.
type NominatimGeocoder struct {
	UserAgent  string
	BaseURL    string // Defaults to "https://nominatim.openstreetmap.org"
	HTTPClient *http.Client
}

// NewNominatimGeocoder creates a Geocoder backed by OpenStreetMap Nominatim.
func NewNominatimGeocoder(userAgent string, client *http.Client) *NominatimGeocoder {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	if userAgent == "" {
		userAgent = "PostcodeNigeria-SDK/1.0"
	}
	return &NominatimGeocoder{
		UserAgent:  userAgent,
		HTTPClient: client,
	}
}

type nominatimResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// Geocode queries the OSM Nominatim API for the address query in Nigeria.
func (n *NominatimGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return 0, 0, fmt.Errorf("nominatim geocoder: query cannot be empty")
	}

	baseURL := n.BaseURL
	if baseURL == "" {
		baseURL = "https://nominatim.openstreetmap.org"
	}

	endpoint := fmt.Sprintf("%s/search?format=json&countrycodes=ng&limit=1&q=%s",
		baseURL, url.QueryEscape(cleanQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("nominatim geocoder request creation failed: %w", err)
	}
	req.Header.Set("User-Agent", n.UserAgent)

	resp, err := n.HTTPClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("nominatim geocoder http call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("nominatim geocoder returned http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return 0, 0, fmt.Errorf("nominatim geocoder reading body failed: %w", err)
	}

	var results []nominatimResult
	if err := json.Unmarshal(body, &results); err != nil {
		return 0, 0, fmt.Errorf("nominatim geocoder decoding json failed: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("nominatim geocoder: no results found for %q", cleanQuery)
	}

	lat, errLat := strconv.ParseFloat(results[0].Lat, 64)
	lng, errLng := strconv.ParseFloat(results[0].Lon, 64)
	if errLat != nil || errLng != nil {
		return 0, 0, fmt.Errorf("nominatim geocoder parsing coordinates failed: %v, %v", errLat, errLng)
	}

	return lat, lng, nil
}

// ChainGeocoder attempts each provided Geocoder in sequence until one succeeds.
type ChainGeocoder struct {
	Geocoders []Geocoder
}

// NewChainGeocoder creates a composite geocoder.
func NewChainGeocoder(geocoders ...Geocoder) *ChainGeocoder {
	valid := make([]Geocoder, 0, len(geocoders))
	for _, g := range geocoders {
		if g != nil {
			valid = append(valid, g)
		}
	}
	return &ChainGeocoder{Geocoders: valid}
}

// Geocode executes the geocoder chain in priority order.
func (c *ChainGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	var lastErr error
	for _, g := range c.Geocoders {
		lat, lng, err := g.Geocode(ctx, query)
		if err == nil && (lat != 0 || lng != 0) {
			return lat, lng, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return 0, 0, lastErr
	}
	return 0, 0, fmt.Errorf("chain geocoder: no geocoder succeeded")
}

// DefaultGeocoder returns a multi-tier geocoder incorporating Google Maps v4 (if key provided) and OSM Nominatim.
func DefaultGeocoder(googleMapsKey string, httpClient *http.Client) Geocoder {
	geocoders := make([]Geocoder, 0, 2)
	if googleMapsKey != "" {
		geocoders = append(geocoders, NewGoogleMapsGeocoder(googleMapsKey, httpClient))
	}
	geocoders = append(geocoders, NewNominatimGeocoder("PostcodeNigeria-SDK/1.0", httpClient))
	return NewChainGeocoder(geocoders...)
}
