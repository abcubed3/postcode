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
