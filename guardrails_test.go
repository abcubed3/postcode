package postcode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentGuardCeilings(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		level := r.URL.Query().Get("level")
		resp := LookupResponse{
			Postcode: "100001",
			Valid:    true,
		}
		if level == "2" || level == "3" {
			resp.AdministrativeAddress = &AdministrativeAddress{
				StateName: "Lagos",
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	ctx := context.Background()

	t.Run("Total Calls Ceiling", func(t *testing.T) {
		client, err := NewClient(
			WithBaseURL(srv.URL),
			WithAgentGuard(AgentGuardConfig{
				MaxTotalCallsPerRun: 2,
			}),
		)
		if err != nil {
			t.Fatalf("unexpected error creating client: %v", err)
		}

		// First call should succeed
		_, err = client.Lookup(ctx, "LA 01 A 01 01", Level1)
		if err != nil {
			t.Fatalf("call 1 failed: %v", err)
		}

		// Second call should succeed
		_, err = client.Lookup(ctx, "LA 01 A 01 02", Level1)
		if err != nil {
			t.Fatalf("call 2 failed: %v", err)
		}

		// Third call should fail with budget exceeded
		_, err = client.Lookup(ctx, "LA 01 A 01 03", Level1)
		if err != ErrAgentBudgetExceeded {
			t.Fatalf("expected ErrAgentBudgetExceeded, got %v", err)
		}

		metrics := client.AgentMetrics()
		if metrics.TotalCalls != 2 {
			t.Errorf("expected 2 total calls in metrics, got %d", metrics.TotalCalls)
		}
	})

	t.Run("Commercial Calls Ceiling with AutoDowngrade", func(t *testing.T) {
		client, err := NewClient(
			WithBaseURL(srv.URL),
			WithAgentGuard(AgentGuardConfig{
				MaxCommercialCallsPerRun: 1,
				AutoDowngradeToLevel1:    true,
			}),
		)
		if err != nil {
			t.Fatalf("unexpected error creating client: %v", err)
		}

		// First commercial call should succeed as Level2
		res1, err := client.Lookup(ctx, "LA 01 A 01 01", Level2)
		if err != nil {
			t.Fatalf("commercial call 1 failed: %v", err)
		}
		if res1.AdministrativeAddress == nil || res1.AdministrativeAddress.StateName != "Lagos" {
			t.Errorf("expected Level2 admin address")
		}

		// Second commercial call should automatically downgrade to Level1
		res2, err := client.Lookup(ctx, "LA 01 A 01 02", Level2)
		if err != nil {
			t.Fatalf("downgraded call failed: %v", err)
		}
		if res2.AdministrativeAddress != nil {
			t.Errorf("expected downgraded Level1 response without admin address")
		}

		metrics := client.AgentMetrics()
		if metrics.CommercialCalls != 1 {
			t.Errorf("expected 1 commercial call, got %d", metrics.CommercialCalls)
		}
		if metrics.DowngradedCalls != 1 {
			t.Errorf("expected 1 downgraded call, got %d", metrics.DowngradedCalls)
		}
	})

	t.Run("AutoFallbackToOffline on ResolveLocation", func(t *testing.T) {
		failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limit exceeded"}`))
		}))
		defer failSrv.Close()

		client, err := NewClient(
			WithBaseURL(failSrv.URL),
			WithAgentGuard(AgentGuardConfig{
				AutoFallbackToOffline: true,
			}),
		)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		loc, err := client.ResolveLocation(ctx, "EK 01 A03 FK 01")
		if err != nil {
			t.Fatalf("expected offline fallback without error, got %v", err)
		}
		if loc.StateCode != "EK" {
			t.Errorf("expected state code EK, got %s", loc.StateCode)
		}
		if loc.Latitude == 0 || loc.Longitude == 0 {
			t.Errorf("expected non-zero fallback coordinates, got lat: %f, lng: %f", loc.Latitude, loc.Longitude)
		}

		metrics := client.AgentMetrics()
		if metrics.OfflineFallbacks < 1 {
			t.Errorf("expected at least 1 offline fallback recorded, got %d", metrics.OfflineFallbacks)
		}
	})

	t.Run("Cache Idempotency", func(t *testing.T) {
		requestsReceived := 0
		cacheSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestsReceived++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(LookupResponse{
				Postcode: "100001",
				Valid:    true,
			})
		}))
		defer cacheSrv.Close()

		client, err := NewClient(
			WithBaseURL(cacheSrv.URL),
			WithDefaultCache(),
		)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		// First call should hit server
		_, err = client.Lookup(ctx, "LA 01 A 01 01", Level1)
		if err != nil {
			t.Fatalf("first lookup failed: %v", err)
		}
		if requestsReceived != 1 {
			t.Fatalf("expected 1 request to server, got %d", requestsReceived)
		}

		// Second call with same code & level should hit cache
		_, err = client.Lookup(ctx, "LA 01 A 01 01", Level1)
		if err != nil {
			t.Fatalf("second lookup failed: %v", err)
		}
		if requestsReceived != 1 {
			t.Fatalf("expected still 1 request to server due to cache hit, got %d", requestsReceived)
		}
	})
}
