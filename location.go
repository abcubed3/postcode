package postcode

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrUnknownState indicates that a postcode has a state prefix not recognized in the Nigerian state registry.
var ErrUnknownState = errors.New("postcode: unknown nigerian state code")

// Precision indicates the resolution accuracy of the resolved coordinates.
type Precision uint8

const (
	// PrecisionBuilding indicates exact building-level point geometry.
	PrecisionBuilding Precision = iota
	// PrecisionArea indicates area or street axis level resolution.
	PrecisionArea
	// PrecisionDistrict indicates district-level centroid resolution.
	PrecisionDistrict
	// PrecisionLGA indicates Local Government Area centroid resolution.
	PrecisionLGA
	// PrecisionState indicates State-level centroid resolution.
	PrecisionState
)

var (
	_ encoding.TextMarshaler   = Precision(0)
	_ encoding.TextUnmarshaler = (*Precision)(nil)
	_ json.Marshaler           = Precision(0)
	_ json.Unmarshaler         = (*Precision)(nil)
)

// String returns the human-readable description of the precision level.
func (p Precision) String() string {
	switch p {
	case PrecisionBuilding:
		return "building"
	case PrecisionArea:
		return "area"
	case PrecisionDistrict:
		return "district"
	case PrecisionLGA:
		return "lga"
	case PrecisionState:
		return "state"
	default:
		return "unknown"
	}
}

// MarshalText implements encoding.TextMarshaler.
func (p Precision) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *Precision) UnmarshalText(text []byte) error {
	switch string(text) {
	case "building":
		*p = PrecisionBuilding
	case "area":
		*p = PrecisionArea
	case "district":
		*p = PrecisionDistrict
	case "lga":
		*p = PrecisionLGA
	case "state":
		*p = PrecisionState
	default:
		return fmt.Errorf("postcode: invalid precision %q", string(text))
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p Precision) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler to accept both string ("building") and numeric (0) forms.
func (p *Precision) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		return p.UnmarshalText(data[1 : len(data)-1])
	}
	var n uint8
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	if n > uint8(PrecisionState) {
		return fmt.Errorf("postcode: invalid precision number %d", n)
	}
	*p = Precision(n)
	return nil
}

// Location encapsulates geographic coordinates, address descriptors, and mapping URLs.
type Location struct {
	Postcode  string    `json:"postcode"`           // Canonical hyphenated: "EK-01-A03-FK-01"
	Compact   string    `json:"compact"`            // Compact 11-char: "EK01A03FK01"
	Latitude  float64   `json:"latitude"`           // Geographic latitude (-90 to +90)
	Longitude float64   `json:"longitude"`          // Geographic longitude (-180 to +180)
	Address   string    `json:"address,omitempty"`  // Known street/building address if available
	StateCode string    `json:"state_code"`         // 2-letter state code (e.g. "EK")
	StateName string    `json:"state_name"`         // Full state name (e.g. "Ekiti")
	LGACode   string    `json:"lga_code,omitempty"` // 2-digit LGA code (e.g. "01")
	LGAName   string    `json:"lga_name,omitempty"` // Full LGA name (e.g. "Ado Ekiti")
	Zone      string    `json:"zone,omitempty"`     // Geopolitical zone (e.g. "SOUTH WEST")
	Precision Precision `json:"precision"`          // Coordinate precision tier (placed at end to minimize struct alignment padding)
}

// IsZero reports whether the location represents a zero/uninitialized value.
func (l Location) IsZero() bool {
	return l == Location{}
}

// GoogleMapsURL returns a standard, universal Google Maps search URL with exact coordinates
// when Precision is Building, or a rich search query for coarse (LGA/State) locations.
// Format: https://www.google.com/maps/search/?api=1&query=...
func (l Location) GoogleMapsURL() string {
	if l.Precision == PrecisionBuilding && (l.Latitude != 0 || l.Longitude != 0) {
		return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%.6f,%.6f", l.Latitude, l.Longitude)
	}
	if l.Latitude != 0 || l.Longitude != 0 {
		query := l.SearchQuery()
		if query != "" && query != "Nigeria" {
			return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(query))
		}
		return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%.6f,%.6f", l.Latitude, l.Longitude)
	}
	if l.Address != "" {
		return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(l.Address))
	}
	if l.StateName != "" {
		return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(l.SearchQuery()))
	}
	return ""
}


// GoogleMapsDirectionsURL returns a Google Maps navigation/directions URL to this location.
// Format: https://www.google.com/maps/dir/?api=1&destination=lat,lng
func (l Location) GoogleMapsDirectionsURL() string {
	if l.Latitude == 0 && l.Longitude == 0 {
		if l.Address != "" {
			return fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%s", url.QueryEscape(l.Address))
		}
		if l.StateName != "" {
			return fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%s", url.QueryEscape(l.SearchQuery()))
		}
		return ""
	}
	return fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%.6f,%.6f", l.Latitude, l.Longitude)
}

// AppleMapsURL returns an Apple Maps URL for iOS/macOS devices.
// Format: https://maps.apple.com/?ll=lat,lng&q=...
func (l Location) AppleMapsURL() string {
	if l.Latitude == 0 && l.Longitude == 0 {
		label := l.Address
		if label == "" {
			label = l.SearchQuery()
		}
		if label == "" || label == "Nigeria" {
			return ""
		}
		return fmt.Sprintf("https://maps.apple.com/?daddr=%s", url.QueryEscape(label))
	}
	label := l.Address
	if label == "" {
		label = l.SearchQuery()
	}
	return fmt.Sprintf("https://maps.apple.com/?ll=%.6f,%.6f&q=%s", l.Latitude, l.Longitude, url.QueryEscape(label))
}

// OpenStreetMapURL returns an OpenStreetMap pin URL.
func (l Location) OpenStreetMapURL() string {
	if l.Latitude == 0 && l.Longitude == 0 {
		return ""
	}
	return fmt.Sprintf("https://www.openstreetmap.org/?mlat=%.6f&mlon=%.6f#map=16/%.6f/%.6f",
		l.Latitude, l.Longitude, l.Latitude, l.Longitude)
}

// Coordinates returns the latitude and longitude as a float64 pair.
func (l Location) Coordinates() (lat, lng float64) {
	return l.Latitude, l.Longitude
}

// SearchQuery produces a descriptive location query string suitable for search engines
// (e.g. "Ado Ekiti, Ekiti State, Nigeria" or "Abuja Municipal, Federal Capital Territory, Nigeria").
func (l Location) SearchQuery() string {
	var parts []string
	if l.Address != "" {
		parts = append(parts, l.Address)
	} else {
		if l.Postcode != "" {
			parts = append(parts, l.Postcode)
		}
		if l.LGAName != "" {
			parts = append(parts, l.LGAName)
		}
		if l.StateName != "" {
			if l.StateCode == "FC" || strings.HasSuffix(l.StateName, "State") || strings.Contains(l.StateName, "Territory") {
				parts = append(parts, l.StateName)
			} else {
				parts = append(parts, l.StateName+" State")
			}
		}
	}
	parts = append(parts, "Nigeria")
	return strings.Join(parts, ", ")
}

// ResolveLocation parses any valid 11-character Nigerian postcode string (compact,
// hyphenated, or spaced) and returns its resolved geographic Location.
// Uses local administrative reference data without requiring network calls.
func ResolveLocation(raw string) (Location, error) {
	p, err := Parse(raw)
	if err != nil {
		return Location{}, err
	}
	loc := p.Location()
	if loc.StateName == "" && loc.Latitude == 0 && loc.Longitude == 0 {
		return Location{}, fmt.Errorf("%w: %q", ErrUnknownState, p.State())
	}
	return loc, nil
}

// GoogleMapsURL parses any valid 11-character Nigerian postcode string and directly
// generates its Google Maps location URL.
func GoogleMapsURL(raw string) (string, error) {
	loc, err := ResolveLocation(raw)
	if err != nil {
		return "", err
	}
	return loc.GoogleMapsURL(), nil
}

// Coordinates parses any valid 11-character Nigerian postcode string and returns
// its latitude and longitude coordinates.
func Coordinates(raw string) (lat, lng float64, err error) {
	loc, err := ResolveLocation(raw)
	if err != nil {
		return 0, 0, err
	}
	return loc.Latitude, loc.Longitude, nil
}
