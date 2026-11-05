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

func TestAgentGuardExtendedAndCacheIsolation(t *testing.T) {
	ctx := context.Background()

	t.Run("Non-Lookup Gateway Endpoints Enforced by MaxTotalCallsPerRun", func(t *testing.T) {
		nonLookupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"segment":"district","suggestions":[]}}`))
		})

		client, err := NewClient(
			WithHandler(nonLookupHandler),
			WithAgentGuard(AgentGuardConfig{
				MaxTotalCallsPerRun: 1,
			}),
		)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		if _, err := client.Autocomplete(ctx, "EK 01"); err != nil {
			t.Fatalf("first Autocomplete call should succeed: %v", err)
		}
		if _, err := client.Autocomplete(ctx, "EK 01 A"); err != ErrAgentBudgetExceeded {
			t.Fatalf("second Autocomplete call should fail with ErrAgentBudgetExceeded, got %v", err)
		}
	})

	t.Run("AutoDowngrade on 402 Does Not Double-Count TotalCalls or CommercialCalls", func(t *testing.T) {
		downgradeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("level") != "1" {
				w.WriteHeader(http.StatusPaymentRequired)
				_, _ = w.Write([]byte(`{"error":{"code":"insufficient_credits","message":"no credits"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"postcode":"EK-01-A03-FK-01","valid":true}}`))
		})

		client, err := NewClient(
			WithHandler(downgradeHandler),
			WithAgentGuard(AgentGuardConfig{
				MaxTotalCallsPerRun:   1,
				AutoDowngradeToLevel1: true,
			}),
		)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		res, err := client.Lookup(ctx, "EK-01-A03-FK-01", Level2)
		if err != nil {
			t.Fatalf("expected auto-downgraded lookup to succeed within 1 logical call budget, got %v", err)
		}
		if !res.Valid {
			t.Errorf("expected res.Valid=true")
		}
		m := client.AgentMetrics()
		if m.TotalCalls != 1 {
			t.Errorf("expected TotalCalls=1, got %d", m.TotalCalls)
		}
		if m.CommercialCalls != 0 {
			t.Errorf("expected CommercialCalls=0 after 402 rollback, got %d", m.CommercialCalls)
		}
		if m.DowngradedCalls != 1 {
			t.Errorf("expected DowngradedCalls=1, got %d", m.DowngradedCalls)
		}
	})

	t.Run("Cache Deep-Clone Isolation", func(t *testing.T) {
		cloneHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"postcode":"EK-01-A03-FK-01","valid":true,"administrative_address":{"state_name":"Ekiti"},"other_building_info":{"type":"detached"},"point_geometry":{"type":"Point","coordinates":[5.2215,7.6211]}}}`))
		})

		client, err := NewClient(
			WithHandler(cloneHandler),
			WithDefaultCache(),
		)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		res1, err := client.Lookup(ctx, "EK-01-A03-FK-01", Level5)
		if err != nil {
			t.Fatalf("first lookup failed: %v", err)
		}
		// Mutate returned struct fields
		res1.AdministrativeAddress.StateName = "MUTATED"
		res1.OtherBuildingInfo["type"] = "MUTATED"
		res1.PointGeometry.Coordinates[0] = 999.0

		res2, err := client.Lookup(ctx, "EK-01-A03-FK-01", Level5)
		if err != nil {
			t.Fatalf("second lookup failed: %v", err)
		}
		if res2.AdministrativeAddress.StateName != "Ekiti" {
			t.Errorf("cache was corrupted by caller mutation: StateName=%q, want Ekiti", res2.AdministrativeAddress.StateName)
		}
		if res2.OtherBuildingInfo["type"] != "detached" {
			t.Errorf("cache map was corrupted: type=%v, want detached", res2.OtherBuildingInfo["type"])
		}
		if res2.PointGeometry.Coordinates[0] != 5.2215 {
			t.Errorf("cache coordinates were corrupted: got %v, want 5.2215", res2.PointGeometry.Coordinates[0])
		}
	})
}
