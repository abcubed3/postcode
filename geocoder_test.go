package postcode_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abcubed3/postcode"
	"github.com/abcubed3/postcode/simulator"
)

func TestGoogleMapsGeocoder_V4(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Geocoding API v4 path specification
		if !strings.HasPrefix(r.URL.Path, "/v4/geocode/address/") {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}

		// Verify v4 headers
		if r.Header.Get("X-Goog-Api-Key") != "test-v4-key" {
			http.Error(w, "missing or invalid X-Goog-Api-Key header", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Goog-FieldMask") != "results.location,results.formattedAddress,results.placeId" {
			http.Error(w, "missing or invalid X-Goog-FieldMask header", http.StatusBadRequest)
			return
		}

		// Verify query parameter regionCode=NG
		if r.URL.Query().Get("regionCode") != "NG" {
			http.Error(w, "missing regionCode=NG", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "Ado") {
			fmt.Fprint(w, `{
				"results": [{
					"location": {
						"lat": 7.6211,
						"lng": 5.2215
					},
					"formattedAddress": "NTA Road, Ado Ekiti, Ekiti, Nigeria",
					"placeId": "ChIJ12345"
				}]
			}`)
			return
		}

		fmt.Fprint(w, `{"results": []}`)
	}))
	defer ts.Close()

	g := &postcode.GoogleMapsGeocoder{
		APIKey:     "test-v4-key",
		BaseURL:    ts.URL,
		HTTPClient: ts.Client(),
	}

	ctx := context.Background()

	t.Run("successful v4 geocode", func(t *testing.T) {
		lat, lng, err := g.Geocode(ctx, "Ado Ekiti, Ekiti, Nigeria")
		if err != nil {
			t.Fatalf("expected successful geocode, got: %v", err)
		}
		if lat != 7.6211 || lng != 5.2215 {
			t.Errorf("expected (7.6211, 5.2215), got (%f, %f)", lat, lng)
		}
	})

	t.Run("empty results error", func(t *testing.T) {
		_, _, err := g.Geocode(ctx, "Unknown Location")
		if err == nil {
			t.Error("expected error for empty results")
		}
	})

	t.Run("empty API key error", func(t *testing.T) {
		emptyG := &postcode.GoogleMapsGeocoder{}
		_, _, err := emptyG.Geocode(ctx, "Lagos")
		if err == nil {
			t.Error("expected error for empty API key")
		}
	})
}

func TestNominatimGeocoder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if r.URL.Query().Get("countrycodes") != "ng" {
			http.Error(w, "missing countrycodes=ng", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(q, "Ikeja") {
			fmt.Fprint(w, `[{"lat": "6.6018", "lon": "3.3515"}]`)
			return
		}
		fmt.Fprint(w, `[]`)
	}))
	defer ts.Close()

	n := &postcode.NominatimGeocoder{
		UserAgent:  "TestAgent",
		BaseURL:    ts.URL,
		HTTPClient: ts.Client(),
	}

	lat, lng, err := n.Geocode(context.Background(), "Ikeja, Lagos, Nigeria")
	if err != nil {
		t.Fatalf("expected successful geocode, got: %v", err)
	}
	if lat != 6.6018 || lng != 3.3515 {
		t.Errorf("expected (6.6018, 3.3515), got (%f, %f)", lat, lng)
	}
}

func TestChainGeocoder(t *testing.T) {
	mockFail := &mockGeocoder{err: fmt.Errorf("provider down")}
	mockSuccess := &mockGeocoder{lat: 9.0579, lng: 7.4951}

	chain := postcode.NewChainGeocoder(mockFail, mockSuccess)
	lat, lng, err := chain.Geocode(context.Background(), "Abuja, Nigeria")
	if err != nil {
		t.Fatalf("chain geocoder failed: %v", err)
	}
	if lat != 9.0579 || lng != 7.4951 {
		t.Errorf("expected (9.0579, 7.4951), got (%f, %f)", lat, lng)
	}
}

func TestResolveLocation_MultiStageGeocodingPipeline(t *testing.T) {
	srv := httptest.NewServer(simulator.NewHandler())
	defer srv.Close()

	mockGeo := &mockGeocoder{lat: 7.6255, lng: 5.2288}

	client, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithAPIKey("test-key"),
		postcode.WithGeocoder(mockGeo),
	)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	loc, err := client.ResolveLocation(context.Background(), "EK-01-A03-FK-01")
	if err != nil {
		t.Fatalf("ResolveLocation failed: %v", err)
	}

	if loc.Precision != postcode.PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding, got %v", loc.Precision)
	}
	if loc.Latitude == 0 || loc.Longitude == 0 {
		t.Error("expected non-zero coordinates")
	}
}

func TestLocation_SmartGoogleMapsURL(t *testing.T) {
	// 1. Building precision uses coordinates
	buildingLoc := postcode.Location{
		Latitude:  7.6211,
		Longitude: 5.2215,
		Precision: postcode.PrecisionBuilding,
		Address:   "12 NTA Road, Ado Ekiti",
	}
	bURL := buildingLoc.GoogleMapsURL()
	if !strings.Contains(bURL, "7.621100,5.221500") {
		t.Errorf("expected coordinates in building URL, got: %s", bURL)
	}

	// 2. Coarse LGA precision uses rich text query
	lgaLoc := postcode.Location{
		Postcode:  "EK-01-A03-FK-01",
		Latitude:  7.6211,
		Longitude: 5.2215,
		Precision: postcode.PrecisionLGA,
		LGAName:   "Ado Ekiti",
		StateName: "Ekiti",
	}
	lgaURL := lgaLoc.GoogleMapsURL()
	if !strings.Contains(lgaURL, "query=") || strings.Contains(lgaURL, "7.621100,5.221500") {
		t.Errorf("expected query search in coarse LGA URL, got: %s", lgaURL)
	}
	if !strings.Contains(lgaURL, "Ado+Ekiti") && !strings.Contains(lgaURL, "Ado%20Ekiti") {
		t.Errorf("expected Ado Ekiti in search query, got: %s", lgaURL)
	}
}

type mockGeocoder struct {
	lat, lng float64
	err      error
}

func (m *mockGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	if m.err != nil {
		return 0, 0, m.err
	}
	return m.lat, m.lng, nil
}
