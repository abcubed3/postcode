package postcode

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_Lookup(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/lookup" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if gotKey := r.Header.Get("X-API-Key"); gotKey != "test_secret_key" {
			t.Errorf("X-API-Key = %q, want test_secret_key", gotKey)
		}
		if code := r.URL.Query().Get("code"); code != "EK 01 A03 FK 01" {
			t.Errorf("query param code = %q, want 'EK 01 A03 FK 01'", code)
		}
		if level := r.URL.Query().Get("level"); level != "2" {
			t.Errorf("query param level = %q, want '2'", level)
		}

		resp := map[string]any{
			"data": map[string]any{
				"postcode": "EK-01-A03-FK-01",
				"valid":    true,
				"administrative_address": map[string]any{
					"state":      "EK",
					"state_name": "Ekiti",
					"lga":        "01",
					"lga_name":   "Ado-Ekiti",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("test_secret_key"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := client.Lookup(context.Background(), "EK 01 A03 FK 01", Level2)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	if !res.Valid {
		t.Errorf("res.Valid = false, want true")
	}
	if res.Postcode != "EK-01-A03-FK-01" {
		t.Errorf("res.Postcode = %q, want EK-01-A03-FK-01", res.Postcode)
	}
	if res.AdministrativeAddress == nil || res.AdministrativeAddress.State != "EK" {
		t.Errorf("res.AdministrativeAddress = %+v, want State=EK", res.AdministrativeAddress)
	}
	if res.AdministrativeAddress.StateName != "Ekiti" {
		t.Errorf("res.AdministrativeAddress.StateName = %q, want Ekiti", res.AdministrativeAddress.StateName)
	}
}

func TestClient_Autocomplete(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search/autocomplete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if q := r.URL.Query().Get("q"); q != "EK 01 A" {
			t.Errorf("q = %q, want 'EK 01 A'", q)
		}

		resp := map[string]any{
			"data": map[string]any{
				"segment": "district",
				"suggestions": []map[string]any{
					{"code": "A03", "label": "Ado District 03"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := client.Autocomplete(context.Background(), "EK 01 A")
	if err != nil {
		t.Fatalf("Autocomplete failed: %v", err)
	}

	if res.Segment != "district" {
		t.Errorf("res.Segment = %q, want district", res.Segment)
	}
	if len(res.Suggestions) != 1 || res.Suggestions[0].Code != "A03" {
		t.Errorf("res.Suggestions = %+v, want 1 item with code A03", res.Suggestions)
	}
}

func TestClient_Assemble(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/assembly/assemble" {
			t.Errorf("path = %s, want /v1/assembly/assemble", r.URL.Path)
		}

		var segs Segments
		if err := json.NewDecoder(r.Body).Decode(&segs); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if segs.State != "EK" || segs.Unit != "01" {
			t.Errorf("unexpected segments: %+v", segs)
		}

		resp := map[string]any{
			"data": map[string]any{
				"postcode": "EK-01-A03-FK-01",
				"display":  "EK 01 A03 FK 01",
				"compact":  "EK01A03FK01",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := client.Assemble(context.Background(), Segments{
		State:    "EK",
		LGA:      "01",
		District: "A03",
		Area:     "FK",
		Unit:     "01",
	})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if res.Postcode != "EK-01-A03-FK-01" {
		t.Errorf("res.Postcode = %q, want EK-01-A03-FK-01", res.Postcode)
	}
}

func TestClient_Nearby(t *testing.T) {
	// Test both raw array payload (live NIPOST) and reference postcode resolution
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/lookup" {
			_, _ = w.Write([]byte(`{"data":{"postcode":"LA-08-A86-RG-01","valid":true,"point_geometry":{"type":"Point","coordinates":[3.633990,6.476111]}}}`))
			return
		}
		if r.URL.Path != "/v1/search/nearby" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("lat") == "" || r.URL.Query().Get("lng") == "" {
			t.Errorf("missing lat or lng query param")
		}

		// Return live NIPOST format: {"data": [ ... ]}
		_, _ = w.Write([]byte(`{"data":[{"postcode":"LA-11-A12-GN-01","display":"LA 11 A12 GN 01","distance_m":9.2}]}`))
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// 1. By coordinates
	res, err := client.Nearby(context.Background(), NearbyParams{
		Latitude:  6.6018,
		Longitude: 3.3515,
		RadiusM:   300,
	})
	if err != nil {
		t.Fatalf("Nearby failed: %v", err)
	}
	if len(res.Results) != 1 || res.Results[0].Postcode != "LA-11-A12-GN-01" {
		t.Errorf("unexpected nearby results: %+v", res)
	}

	// 2. By reference postcode (NearbyPostcode)
	resCode, err := client.NearbyPostcode(context.Background(), "LA-08-A86-RG-01", 250)
	if err != nil {
		t.Fatalf("NearbyPostcode failed: %v", err)
	}
	if len(resCode.Results) != 1 {
		t.Errorf("unexpected nearby postcode results count: %d", len(resCode.Results))
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		resp := map[string]any{
			"error": map[string]any{
				"code":    "insufficient_credits",
				"message": "not enough credits; top up to continue",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	_, err = client.Lookup(context.Background(), "EK 01 A03 FK 01", Level3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != http.StatusPaymentRequired {
		t.Errorf("StatusCode = %d, want 402", apiErr.StatusCode)
	}
	if apiErr.Code != "insufficient_credits" {
		t.Errorf("Code = %q, want insufficient_credits", apiErr.Code)
	}
	if !apiErr.IsInsufficientCredits() {
		t.Errorf("IsInsufficientCredits() = false, want true")
	}
}

func TestClient_BaseTransportAndRetryConfig(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := attempts.Add(1)
		if att < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"postcode": "EK-01-A03-FK-01", "valid": true},
		})
	}))
	defer server.Close()

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithBaseTransport(http.DefaultTransport),
		WithRetryConfig(RetryConfig{
			MaxRetries: 2,
			BaseDelay:  5 * time.Millisecond,
			MaxDelay:   20 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := client.Lookup(context.Background(), "EK 01 A03 FK 01", Level1)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if !res.Valid {
		t.Errorf("res.Valid = false, want true")
	}
	if attempts.Load() != 2 {
		t.Errorf("attempts = %d, want 2", attempts.Load())
	}
}

func TestClient_LimitReaderDefense(t *testing.T) {
	t.Parallel()
	// Server returns a response exceeding maxResponseBytes (4MB)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Write 5MB of garbage data
		chunk := strings.Repeat("A", 1024*1024)
		for range 5 {
			_, _ = w.Write([]byte(chunk))
		}
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Should safely fail JSON decoding of truncated stream rather than OOMing
	_, err = client.Lookup(context.Background(), "EK 01 A03 FK 01", Level1)
	if err == nil {
		t.Fatal("expected error on oversized stream, got nil")
	}
}

func TestClient_RateLimitHandling(t *testing.T) {
	t.Parallel()

	// 1. Successful request records rate limit info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "597")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"postcode":"EK-01-A03-FK-01","valid":true}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := client.Lookup(context.Background(), "EK-01-A03-FK-01", Level1)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if !res.Valid {
		t.Errorf("res.Valid = false, want true")
	}

	rl := client.RateLimit()
	if rl == nil {
		t.Fatal("expected client.RateLimit() to be non-nil after request")
	}
	if rl.Limit != 600 || rl.Remaining != 597 {
		t.Errorf("RateLimit = %+v, want Limit=600, Remaining=597", rl)
	}
}

func TestClient_RateLimit429_ErrorDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("Retry-After", "15")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"too many requests"}}`))
	}))
	defer server.Close()

	// Disable retries so we see the 429 APIError immediately
	client, err := NewClient(
		WithBaseURL(server.URL),
		WithRateLimitRetry(false, 0),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	_, err = client.Lookup(context.Background(), "EK-01-A03-FK-01", Level1)
	if err == nil {
		t.Fatal("expected error on 429, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if !apiErr.IsRateLimit() {
		t.Errorf("apiErr.IsRateLimit() = false, want true")
	}
	if apiErr.RetryAfter != 15*time.Second {
		t.Errorf("apiErr.RetryAfter = %v, want 15s", apiErr.RetryAfter)
	}
	if apiErr.RateLimit == nil || apiErr.RateLimit.Remaining != 0 {
		t.Errorf("apiErr.RateLimit = %+v, want Remaining=0", apiErr.RateLimit)
	}
}

func TestClient_Health(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("unexpected path: %s, want /healthz", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("client.Health() failed: %v", err)
	}
}

func TestNormalizeAndAssembleSegments(t *testing.T) {
	t.Parallel()

	// Official documentation example: single-digit lga "1", unit "1", lowercase "ek", "a03", "fk"
	raw := Segments{
		State:    "ek",
		LGA:      "1",
		District: "a03",
		Area:     "fk",
		Unit:     "1",
	}

	norm := NormalizeSegments(raw)
	if norm.State != "EK" || norm.LGA != "01" || norm.District != "A03" || norm.Area != "FK" || norm.Unit != "01" {
		t.Fatalf("NormalizeSegments failed, got: %+v", norm)
	}

	assembled, err := AssembleSegments(raw)
	if err != nil {
		t.Fatalf("AssembleSegments failed: %v", err)
	}
	if assembled.Postcode != "EK-01-A03-FK-01" {
		t.Errorf("assembled.Postcode = %q, want EK-01-A03-FK-01", assembled.Postcode)
	}
	if assembled.Display != "EK 01 A03 FK 01" {
		t.Errorf("assembled.Display = %q, want EK 01 A03 FK 01", assembled.Display)
	}
	if assembled.Compact != "EK01A03FK01" {
		t.Errorf("assembled.Compact = %q, want EK01A03FK01", assembled.Compact)
	}
}

func TestClient_ReferenceCatalog(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/reference/states":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"states": []map[string]string{
						{"code": "FC", "name": "Federal Capital Territory"},
						{"code": "EK", "name": "Ekiti"},
					},
				},
			})
		case "/v1/reference/lgas":
			if state := r.URL.Query().Get("state"); state != "FC" {
				t.Errorf("expected state=FC, got %s", state)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"lgas": []map[string]string{
						{"code": "01", "name": "Abaji"},
					},
				},
			})
		case "/v1/reference/districts":
			if state := r.URL.Query().Get("state"); state != "FC" {
				t.Errorf("expected state=FC, got %s", state)
			}
			if lga := r.URL.Query().Get("lga"); lga != "01" {
				t.Errorf("expected lga=01, got %s", lga)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"districts": []map[string]string{
						{"code": "A01"},
					},
				},
			})
		case "/v1/reference/areas":
			if state := r.URL.Query().Get("state"); state != "FC" {
				t.Errorf("expected state=FC, got %s", state)
			}
			if lga := r.URL.Query().Get("lga"); lga != "01" {
				t.Errorf("expected lga=01, got %s", lga)
			}
			if dist := r.URL.Query().Get("district"); dist != "A01" {
				t.Errorf("expected district=A01, got %s", dist)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"areas": []map[string]string{
						{"code": "KP"},
					},
				},
			})
		default:
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	states, err := client.ReferenceStates(context.Background())
	if err != nil {
		t.Fatalf("ReferenceStates failed: %v", err)
	}
	if len(states) != 2 || states[0].Code != "FC" {
		t.Errorf("unexpected states: %+v", states)
	}

	lgas, err := client.ReferenceLGAs(context.Background(), "fc")
	if err != nil {
		t.Fatalf("ReferenceLGAs failed: %v", err)
	}
	if len(lgas) != 1 || lgas[0].Code != "01" || lgas[0].Name != "Abaji" {
		t.Errorf("unexpected LGAs: %+v", lgas)
	}

	districts, err := client.ReferenceDistricts(context.Background(), "fc", "1")
	if err != nil {
		t.Fatalf("ReferenceDistricts failed: %v", err)
	}
	if len(districts) != 1 || districts[0].Code != "A01" {
		t.Errorf("unexpected districts: %+v", districts)
	}

	areas, err := client.ReferenceAreas(context.Background(), "fc", "1", "a01")
	if err != nil {
		t.Fatalf("ReferenceAreas failed: %v", err)
	}
	if len(areas) != 1 || areas[0].Code != "KP" {
		t.Errorf("unexpected areas: %+v", areas)
	}
}

