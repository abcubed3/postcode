package postcode_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abcubed3/postcode"
	"github.com/abcubed3/postcode/simulator"
)

func TestResolveLocation_OfficialBuildings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedLat float64
		expectedLng float64
		expectedAdd string
		expectedSt  string
		expectedLGA string
	}{
		{
			name:        "Ekiti Fabian Hotel Axis (Hyphenated)",
			input:       "EK-01-A03-FK-01",
			expectedLat: 7.6211,
			expectedLng: 5.2215,
			expectedAdd: "NTA Road, Back of Fabian Hotel, Ado Ekiti",
			expectedSt:  "Ekiti",
			expectedLGA: "Ado Ekiti",
		},
		{
			name:        "Lagos Alausa Town Centre (Compact)",
			input:       "LA11W06TC10",
			expectedLat: 6.6018,
			expectedLng: 3.3515,
			expectedAdd: "10 Obafemi Awolowo Way, Ikeja, Lagos",
			expectedSt:  "Lagos",
			expectedLGA: "Ikeja",
		},
		{
			name:        "Abuja Garki II Area 11 (Spaced)",
			input:       "FC 03 B06 AG 12",
			expectedLat: 9.0579,
			expectedLng: 7.4951,
			expectedAdd: "12 Shehu Shagari Way, Garki, Abuja",
			expectedSt:  "Federal Capital Territory",
			expectedLGA: "Abuja Municipal",
		},
		{
			name:        "Kano Bompai West Road (Lowercase Compact)",
			input:       "kn31f82wj80",
			expectedLat: 12.0022,
			expectedLng: 8.5920,
			expectedAdd: "80 Badu Road, Bompai, Kano",
			expectedSt:  "Kano",
			expectedLGA: "Kano Municipal",
		},
		{
			name:        "Ogun Abeokuta Banking Axis",
			input:       "OG-14-T18-BN-16",
			expectedLat: 7.1475,
			expectedLng: 3.3619,
			expectedAdd: "16 Lalubu Street, Oke-Ilewo, Abeokuta",
			expectedSt:  "Ogun",
			expectedLGA: "Abeokuta South",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := postcode.ResolveLocation(tc.input)
			if err != nil {
				t.Fatalf("unexpected error resolving %q: %v", tc.input, err)
			}

			if loc.Precision != postcode.PrecisionBuilding {
				t.Errorf("expected PrecisionBuilding (%v), got %v", postcode.PrecisionBuilding, loc.Precision)
			}
			if loc.Latitude != tc.expectedLat || loc.Longitude != tc.expectedLng {
				t.Errorf("coordinates mismatch: expected (%.4f, %.4f), got (%.4f, %.4f)",
					tc.expectedLat, tc.expectedLng, loc.Latitude, loc.Longitude)
			}
			if loc.StateName != tc.expectedSt {
				t.Errorf("expected state %q, got %q", tc.expectedSt, loc.StateName)
			}
			if loc.LGAName != tc.expectedLGA {
				t.Errorf("expected LGA %q, got %q", tc.expectedLGA, loc.LGAName)
			}
			if loc.Address != tc.expectedAdd {
				t.Errorf("expected address %q, got %q", tc.expectedAdd, loc.Address)
			}

			// Verify Google Maps URL format
			gmaps := loc.GoogleMapsURL()
			if !strings.HasPrefix(gmaps, "https://www.google.com/maps/search/?api=1&query=") {
				t.Errorf("invalid Google Maps URL: %s", gmaps)
			}
			if !strings.Contains(gmaps, "7.621100") && !strings.Contains(gmaps, "6.601800") &&
				!strings.Contains(gmaps, "9.057900") && !strings.Contains(gmaps, "12.002200") &&
				!strings.Contains(gmaps, "7.147500") {
				t.Errorf("Google Maps URL does not contain formatted coordinates: %s", gmaps)
			}

			// Verify direct helper
			directURL, err := postcode.GoogleMapsURL(tc.input)
			if err != nil || directURL != gmaps {
				t.Errorf("postcode.GoogleMapsURL() = %q, want %q (err: %v)", directURL, gmaps, err)
			}

			// Verify coordinates helper
			lat, lng, err := postcode.Coordinates(tc.input)
			if err != nil || lat != tc.expectedLat || lng != tc.expectedLng {
				t.Errorf("postcode.Coordinates() = (%.4f, %.4f), want (%.4f, %.4f)", lat, lng, tc.expectedLat, tc.expectedLng)
			}
		})
	}
}

func TestResolveLocation_LGAFallback(t *testing.T) {
	t.Parallel()

	// Postcode in Ikeja (LA-11) with non-registered building unit
	input := "LA-11-Z99-XX-99"
	loc, err := postcode.ResolveLocation(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loc.Precision != postcode.PrecisionLGA {
		t.Errorf("expected PrecisionLGA (%v), got %v", postcode.PrecisionLGA, loc.Precision)
	}
	if loc.StateName != "Lagos" {
		t.Errorf("expected state Lagos, got %q", loc.StateName)
	}
	if loc.LGAName != "Ikeja" {
		t.Errorf("expected LGA Ikeja, got %q", loc.LGAName)
	}
	if loc.Latitude == 0 || loc.Longitude == 0 {
		t.Errorf("expected non-zero coordinates for Ikeja centroid, got (%.4f, %.4f)", loc.Latitude, loc.Longitude)
	}
}

func TestResolveLocation_StateFallback(t *testing.T) {
	t.Parallel()

	// Postcode in Zamfara (ZA) with arbitrary LGA 55
	input := "ZA-55-A01-AA-01"
	loc, err := postcode.ResolveLocation(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loc.Precision != postcode.PrecisionState {
		t.Errorf("expected PrecisionState (%v), got %v", postcode.PrecisionState, loc.Precision)
	}
	if loc.StateName != "Zamfara" {
		t.Errorf("expected state Zamfara, got %q", loc.StateName)
	}
	if loc.Latitude == 0 || loc.Longitude == 0 {
		t.Errorf("expected non-zero state coordinates, got (%.4f, %.4f)", loc.Latitude, loc.Longitude)
	}
}

func TestAll37StatesCoverage(t *testing.T) {
	t.Parallel()

	if len(postcode.NigerianStates) != 37 {
		t.Fatalf("expected 37 Nigerian states/FCT, got %d", len(postcode.NigerianStates))
	}

	for code, st := range postcode.NigerianStates {
		code := code
		st := st
		t.Run(code+"_"+st.Name, func(t *testing.T) {
			raw := code + "01A01AA01"
			loc, err := postcode.ResolveLocation(raw)
			if err != nil {
				t.Fatalf("failed to resolve valid code %q for %s: %v", raw, st.Name, err)
			}

			if loc.StateCode != code {
				t.Errorf("expected StateCode %q, got %q", code, loc.StateCode)
			}
			if loc.StateName != st.Name {
				t.Errorf("expected StateName %q, got %q", st.Name, loc.StateName)
			}
			// Nigeria latitude spans ~4.0 to 14.0 N, longitude ~2.5 to 15.0 E
			if loc.Latitude < 4.0 || loc.Latitude > 14.0 {
				t.Errorf("latitude %.4f for %s out of realistic Nigeria bounds (4.0 - 14.0)", loc.Latitude, st.Name)
			}
			if loc.Longitude < 2.5 || loc.Longitude > 15.0 {
				t.Errorf("longitude %.4f for %s out of realistic Nigeria bounds (2.5 - 15.0)", loc.Longitude, st.Name)
			}

			gmapsURL := loc.GoogleMapsURL()
			if !strings.HasPrefix(gmapsURL, "https://www.google.com/maps/search/?api=1&query=") {
				t.Errorf("invalid Google Maps URL: %s", gmapsURL)
			}
		})
	}
}

func TestLocation_URLGenerators(t *testing.T) {
	t.Parallel()

	loc := postcode.Location{
		Latitude:  7.6211,
		Longitude: 5.2215,
		Address:   "NTA Road, Back of Fabian Hotel, Ado Ekiti",
		StateName: "Ekiti",
		LGAName:   "Ado Ekiti",
	}

	// 1. Google Maps Search
	gmaps := loc.GoogleMapsURL()
	expectedGMaps := "https://www.google.com/maps/search/?api=1&query=7.621100,5.221500"
	if gmaps != expectedGMaps {
		t.Errorf("GoogleMapsURL() = %q, want %q", gmaps, expectedGMaps)
	}

	// 2. Google Maps Directions
	dir := loc.GoogleMapsDirectionsURL()
	expectedDir := "https://www.google.com/maps/dir/?api=1&destination=7.621100,5.221500"
	if dir != expectedDir {
		t.Errorf("GoogleMapsDirectionsURL() = %q, want %q", dir, expectedDir)
	}

	// 3. Apple Maps
	apple := loc.AppleMapsURL()
	if !strings.HasPrefix(apple, "https://maps.apple.com/?ll=7.621100,5.221500") {
		t.Errorf("AppleMapsURL() = %q, expected prefix", apple)
	}

	// 4. OpenStreetMap
	osm := loc.OpenStreetMapURL()
	if !strings.HasPrefix(osm, "https://www.openstreetmap.org/?mlat=7.621100&mlon=5.221500") {
		t.Errorf("OpenStreetMapURL() = %q, expected prefix", osm)
	}

	// 5. Zero value Location
	var zeroLoc postcode.Location
	if zeroLoc.GoogleMapsURL() != "" {
		t.Errorf("zeroLoc.GoogleMapsURL() = %q, want empty", zeroLoc.GoogleMapsURL())
	}
	if zeroLoc.GoogleMapsDirectionsURL() != "" {
		t.Errorf("zeroLoc.GoogleMapsDirectionsURL() = %q, want empty", zeroLoc.GoogleMapsDirectionsURL())
	}
}

func TestPostcode_DirectMethods(t *testing.T) {
	t.Parallel()

	p, err := postcode.Parse("EK-01-A03-FK-01")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	loc := p.Location()
	if loc.Precision != postcode.PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding, got %v", loc.Precision)
	}
	if loc.Latitude != 7.6211 || loc.Longitude != 5.2215 {
		t.Errorf("coordinates mismatch: (%.4f, %.4f)", loc.Latitude, loc.Longitude)
	}

	lat, lng := p.Coordinates()
	if lat != 7.6211 || lng != 5.2215 {
		t.Errorf("p.Coordinates() = (%.4f, %.4f)", lat, lng)
	}

	gmaps := p.GoogleMapsURL()
	if !strings.Contains(gmaps, "7.621100,5.221500") {
		t.Errorf("p.GoogleMapsURL() = %q", gmaps)
	}

	// Zero Postcode
	var zeroP postcode.Postcode
	zeroLoc := zeroP.Location()
	if !zeroP.IsZero() || zeroLoc.Latitude != 0 {
		t.Errorf("expected zero location for zero postcode")
	}
}

func TestRegisterCustomBuilding(t *testing.T) {
	// Register a new custom building unit
	customCode := "LA-11-C99-XX-88"
	postcode.RegisterKnownBuilding(postcode.BuildingRecord{
		Postcode:  customCode,
		Latitude:  6.5920,
		Longitude: 3.3550,
		Address:   "Test Custom Headquarters, Ikeja, Lagos",
		StateCode: "LA",
		StateName: "Lagos",
		LGACode:   "11",
		LGAName:   "Ikeja",
		Zone:      "SOUTH WEST",
	})

	loc, err := postcode.ResolveLocation(customCode)
	if err != nil {
		t.Fatalf("failed to resolve custom building: %v", err)
	}

	if loc.Precision != postcode.PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding, got %v", loc.Precision)
	}
	if loc.Latitude != 6.5920 || loc.Longitude != 3.3550 {
		t.Errorf("coordinates mismatch: (%.4f, %.4f)", loc.Latitude, loc.Longitude)
	}
	if loc.Address != "Test Custom Headquarters, Ikeja, Lagos" {
		t.Errorf("unexpected address: %q", loc.Address)
	}
}

func TestResolveLocation_InvalidInputs(t *testing.T) {
	t.Parallel()

	invalids := []string{
		"",
		"INVALID",
		"EK-01",
		"EK-00-A03-FK-01", // LGA 00 invalid
		"ZZ-01-A03-FK-01", // unknown state code
	}

	for _, inv := range invalids {
		_, err := postcode.ResolveLocation(inv)
		if err == nil {
			t.Errorf("expected error for invalid input %q, got nil", inv)
		}
		_, errURL := postcode.GoogleMapsURL(inv)
		if errURL == nil {
			t.Errorf("expected error for GoogleMapsURL with invalid input %q, got nil", inv)
		}
		_, _, errCoords := postcode.Coordinates(inv)
		if errCoords == nil {
			t.Errorf("expected error for Coordinates with invalid input %q, got nil", inv)
		}
	}
}

func TestClient_ResolveLocation_LiveSimulator(t *testing.T) {
	srv := simulator.NewServer()
	defer srv.Close()

	client, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithAPIKey("sim_test_key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// 1. Resolve official test postcode
	loc, err := client.ResolveLocation(ctx, "LA-11-W06-TC-10")
	if err != nil {
		t.Fatalf("client.ResolveLocation failed: %v", err)
	}

	if loc.Precision != postcode.PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding, got %v", loc.Precision)
	}
	if loc.Latitude != 6.6018 || loc.Longitude != 3.3515 {
		t.Errorf("coordinates mismatch: (%.4f, %.4f)", loc.Latitude, loc.Longitude)
	}
	if !strings.EqualFold(loc.StateName, "Lagos") || !strings.EqualFold(loc.LGAName, "Ikeja") {
		t.Errorf("admin names mismatch: State=%q, LGA=%q", loc.StateName, loc.LGAName)
	}

	gmapsURL := loc.GoogleMapsURL()
	if !strings.Contains(gmapsURL, "6.601800,3.351500") {
		t.Errorf("unexpected Google Maps URL: %s", gmapsURL)
	}
}

func TestClient_ResolveLocation_FallbackOnUnauthorized(t *testing.T) {
	// Simulator without API key causes commercial Levels 2 & 3 to return 401 Unauthorized
	srv := httptest.NewServer(simulator.NewHandler())
	defer srv.Close()

	client, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		// No API Key! Level 2 and Level 3 will return 401 Unauthorized
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Should gracefully fall back to local offline resolution
	loc, err := client.ResolveLocation(ctx, "EK-01-A03-FK-01")
	if err != nil {
		t.Fatalf("client.ResolveLocation should not fail on 401, got err: %v", err)
	}

	if loc.Latitude != 7.6211 || loc.Longitude != 5.2215 {
		t.Errorf("expected fallback coordinates (7.6211, 5.2215), got (%.4f, %.4f)", loc.Latitude, loc.Longitude)
	}
	if loc.StateName != "Ekiti" {
		t.Errorf("expected state Ekiti, got %q", loc.StateName)
	}
}

func TestLookupLevels_OfficialThreeLevels(t *testing.T) {
	if postcode.LevelMax != postcode.Level3 {
		t.Errorf("expected LevelMax to equal Level3 (3), got %v", postcode.LevelMax)
	}
	if postcode.Level1 != 1 || postcode.Level2 != 2 || postcode.Level3 != 3 {
		t.Errorf("unexpected level constants: L1=%d, L2=%d, L3=%d", postcode.Level1, postcode.Level2, postcode.Level3)
	}
}

func TestPrecision_JSONAndTextMarshaling(t *testing.T) {
	t.Parallel()

	// 1. Text marshaling
	txt, err := postcode.PrecisionBuilding.MarshalText()
	if err != nil || string(txt) != "building" {
		t.Errorf("MarshalText() = %q, want %q (err: %v)", string(txt), "building", err)
	}

	// 2. Text unmarshaling
	var p postcode.Precision
	if err := p.UnmarshalText([]byte("lga")); err != nil || p != postcode.PrecisionLGA {
		t.Errorf("UnmarshalText('lga') = %v, want PrecisionLGA (err: %v)", p, err)
	}
	if err := p.UnmarshalText([]byte("invalid")); err == nil {
		t.Errorf("UnmarshalText('invalid') should error, got nil")
	}

	// 3. Location JSON serialization
	loc := postcode.Location{
		Postcode:  "EK-01-A03-FK-01",
		Precision: postcode.PrecisionBuilding,
	}
	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"precision":"building"`) {
		t.Errorf("json.Marshal(loc) does not contain string precision: %s", string(data))
	}

	// 4. JSON deserialization from string form
	var unmarshaled postcode.Location
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if unmarshaled.Precision != postcode.PrecisionBuilding {
		t.Errorf("unmarshaled.Precision = %v, want PrecisionBuilding", unmarshaled.Precision)
	}

	// 5. JSON deserialization from legacy numeric form (0..4)
	legacyJSON := `{"postcode":"LA-11-W06-TC-10","precision":3}` // 3 = PrecisionLGA
	var legacyLoc postcode.Location
	if err := json.Unmarshal([]byte(legacyJSON), &legacyLoc); err != nil {
		t.Fatalf("legacy json.Unmarshal failed: %v", err)
	}
	if legacyLoc.Precision != postcode.PrecisionLGA {
		t.Errorf("legacyLoc.Precision = %v, want PrecisionLGA", legacyLoc.Precision)
	}
}

func TestSearchQuery_AdministrativeNaming(t *testing.T) {
	t.Parallel()

	// FCT should not append "State"
	fctLoc := postcode.Location{
		StateCode: "FC",
		StateName: "Federal Capital Territory",
		LGAName:   "Abuja Municipal",
	}
	qFCT := fctLoc.SearchQuery()
	if strings.Contains(qFCT, "Federal Capital Territory State") {
		t.Errorf("SearchQuery() contained 'Territory State': %q", qFCT)
	}
	if qFCT != "Abuja Municipal, Federal Capital Territory, Nigeria" {
		t.Errorf("SearchQuery() = %q, want 'Abuja Municipal, Federal Capital Territory, Nigeria'", qFCT)
	}

	// Regular State should append "State"
	ekitiLoc := postcode.Location{
		StateCode: "EK",
		StateName: "Ekiti",
		LGAName:   "Ado Ekiti",
	}
	qEkiti := ekitiLoc.SearchQuery()
	if qEkiti != "Ado Ekiti, Ekiti State, Nigeria" {
		t.Errorf("SearchQuery() = %q, want 'Ado Ekiti, Ekiti State, Nigeria'", qEkiti)
	}
}

func TestLocation_IsZero(t *testing.T) {
	t.Parallel()

	var zeroLoc postcode.Location
	if !zeroLoc.IsZero() {
		t.Errorf("expected zeroLoc.IsZero() to be true")
	}

	nonZero := postcode.Location{StateCode: "EK"}
	if nonZero.IsZero() {
		t.Errorf("expected nonZero.IsZero() to be false")
	}
}

func TestLocation_DirectionsURL_Fallbacks(t *testing.T) {
	t.Parallel()

	// Zero lat/lng with address
	locWithAddr := postcode.Location{
		Address: "National Assembly Complex, Abuja",
	}
	gDir := locWithAddr.GoogleMapsDirectionsURL()
	if !strings.HasPrefix(gDir, "https://www.google.com/maps/dir/?api=1&destination=") {
		t.Errorf("GoogleMapsDirectionsURL() = %q, expected prefix", gDir)
	}
	appleDir := locWithAddr.AppleMapsURL()
	if !strings.HasPrefix(appleDir, "https://maps.apple.com/?daddr=") {
		t.Errorf("AppleMapsURL() = %q, expected prefix", appleDir)
	}

	// Truly zero location should return empty string
	var zeroLoc postcode.Location
	if zeroLoc.GoogleMapsDirectionsURL() != "" {
		t.Errorf("expected empty string for zeroLoc GoogleMapsDirectionsURL, got %q", zeroLoc.GoogleMapsDirectionsURL())
	}
	if zeroLoc.AppleMapsURL() != "" {
		t.Errorf("expected empty string for zeroLoc AppleMapsURL, got %q", zeroLoc.AppleMapsURL())
	}
}

func TestResolveLocation_ErrUnknownState(t *testing.T) {
	t.Parallel()

	// "ZZ" is not a valid Nigerian state
	_, err := postcode.ResolveLocation("ZZ-01-A01-AA-01")
	if err == nil {
		t.Fatalf("expected error for unknown state 'ZZ', got nil")
	}
	if !errors.Is(err, postcode.ErrUnknownState) {
		t.Errorf("expected errors.Is(err, ErrUnknownState), got %v", err)
	}
}

func BenchmarkResolveLocation(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = postcode.ResolveLocation("EK-01-A03-FK-01")
	}
}

func BenchmarkGoogleMapsURL(b *testing.B) {
	p, _ := postcode.Parse("EK-01-A03-FK-01")
	b.ReportAllocs()
	for b.Loop() {
		_ = p.GoogleMapsURL()
	}
}
