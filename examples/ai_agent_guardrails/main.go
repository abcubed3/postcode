package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/abcubed3/postcode"
)

// DeliveryOrder represents a customer order in an autonomous e-commerce dispatch system.
type DeliveryOrder struct {
	ID           string
	CustomerName string
	RawAddress   string
	Postcode     string
}

func main() {
	ctx := context.Background()

	fmt.Println("==================================================================")
	fmt.Println("  Autonomous Logistics Agent with Budget & Quota Guardrails")
	fmt.Println("==================================================================")

	// Spin up a mock NIPOST gateway server to simulate commercial & rate-limit behavior
	gatewayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		level := r.URL.Query().Get("level")
		code := r.URL.Query().Get("code")

		// Return simulated response
		w.Header().Set("Content-Type", "application/json")
		if level == "2" || level == "3" {
			_, _ = fmt.Fprintf(w, `{
				"postcode": %q,
				"valid": true,
				"administrative_address": {
					"state_name": "Lagos",
					"lga_name": "Ikeja",
					"district_name": "Central District"
				},
				"recent_house_address": {
					"address": "12, Commercial Avenue, Ikeja"
				}
			}`, code)
		} else {
			_, _ = fmt.Fprintf(w, `{
				"postcode": %q,
				"valid": true
			}`, code)
		}
	}))
	defer gatewayServer.Close()

	// 1. Initialize Client with Agent Guardrails and In-Memory Cache
	// Budget: Max 2 commercial calls (Level 2/3), max 6 total calls.
	// AutoDowngradeToLevel1: Once commercial budget is reached, automatically downgrade
	// to free Level 1 validation without throwing errors.
	// AutoFallbackToOffline: If gateway fails, seamlessly fall back to offline geocoding.
	guardCfg := postcode.AgentGuardConfig{
		MaxCommercialCallsPerRun: 2,
		MaxTotalCallsPerRun:      6,
		AutoDowngradeToLevel1:    true,
		AutoFallbackToOffline:    true,
	}

	client, err := postcode.NewClient(
		postcode.WithBaseURL(gatewayServer.URL),
		postcode.WithAPIKey("test_agent_api_key"),
		postcode.WithAgentGuard(guardCfg),
		postcode.WithDefaultCache(), // 1000 items, 30m TTL
	)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	// 2. Queue of orders to process
	orders := []DeliveryOrder{
		{ID: "ORD-001", CustomerName: "Amina Yusuf", RawAddress: "Victoria Island", Postcode: "LA 11 W06 TC 10"},
		{ID: "ORD-002", CustomerName: "Chidi Okafor", RawAddress: "Victoria Island (Repeat)", Postcode: "LA 11 W06 TC 10"}, // Same postcode -> cache hit!
		{ID: "ORD-003", CustomerName: "Tunde Bakare", RawAddress: "Ado Ekiti Center", Postcode: "EK 01 A03 FK 01"},      // Second commercial call
		{ID: "ORD-004", CustomerName: "Fatima Aliyu", RawAddress: "Abuja Central", Postcode: "FC 03 B06 AG 12"},         // Commercial budget exhausted -> auto-downgraded to Level 1!
	}

	for _, order := range orders {
		fmt.Printf("\n📦 Processing Order %s (%s)...\n", order.ID, order.CustomerName)

		// Request Level 2 commercial lookup for street address
		res, err := client.Lookup(ctx, order.Postcode, postcode.Level2)
		if err != nil {
			fmt.Printf("   ⚠️  Lookup error: %v\n", err)
			continue
		}

		if res.AdministrativeAddress != nil {
			fmt.Printf("   ✅ Commercial L2: State=%s, LGA=%s, Address=%s\n",
				res.AdministrativeAddress.StateName,
				res.AdministrativeAddress.LGAName,
				res.RecentHouseAddress.Address)
		} else {
			fmt.Printf("   ⚡ Auto-Downgraded to L1 (Free): Postcode=%s, Valid=%t (Administrative quota conserved)\n",
				res.Postcode, res.Valid)
		}
	}

	// 3. Test Offline Geocoding Fallback with simulated gateway failure
	fmt.Println("\n------------------------------------------------------------------")
	fmt.Println("  Simulating Gateway Outage / Rate Limit for ResolveLocation")
	fmt.Println("------------------------------------------------------------------")

	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message": "rate limit exceeded"}`))
	}))
	defer failingServer.Close()

	offlineClient, _ := postcode.NewClient(
		postcode.WithBaseURL(failingServer.URL),
		postcode.WithAgentGuard(postcode.AgentGuardConfig{
			AutoFallbackToOffline: true,
		}),
	)

	loc, err := offlineClient.ResolveLocation(ctx, "EK-01-A03-FK-01")
	if err != nil {
		log.Fatalf("Unexpected error during fallback: %v", err)
	}
	fmt.Printf("📍 Offline Fallback Success:\n")
	fmt.Printf("   State: %s (%s)\n", loc.StateName, loc.StateCode)
	fmt.Printf("   Coordinates: %.4f, %.4f\n", loc.Latitude, loc.Longitude)
	fmt.Printf("   Google Maps: %s\n", loc.GoogleMapsURL())

	// 4. Inspect Live Guardrails Telemetry
	fmt.Println("\n==================================================================")
	fmt.Println("  Session Guardrails Telemetry Metrics")
	fmt.Println("==================================================================")
	metrics := client.AgentMetrics()
	fmt.Printf("  • Total Calls Executed:     %d (Ceiling: %d)\n", metrics.TotalCalls, guardCfg.MaxTotalCallsPerRun)
	fmt.Printf("  • Commercial Calls Paid:    %d (Ceiling: %d)\n", metrics.CommercialCalls, guardCfg.MaxCommercialCallsPerRun)
	fmt.Printf("  • Automated Downgrades:     %d\n", metrics.DowngradedCalls)
	fmt.Printf("  • Offline Fallbacks:        %d\n", offlineClient.AgentMetrics().OfflineFallbacks)
	fmt.Println("==================================================================")
}
