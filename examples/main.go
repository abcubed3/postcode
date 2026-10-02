package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/abcubed3/postcode"
	"github.com/abcubed3/postcode/otelpostcode"
	"github.com/abcubed3/postcode/simulator"
)

func main() {
	ctx := context.Background()

	fmt.Println("==============================================================")
	fmt.Println(" 1. Local Parsing & Validation")
	fmt.Println("==============================================================")

	sampleInputs := []string{
		"AK-11-I61-ZF-12", // Akwa Ibom
		"ba 02 m67 bl 69", // Bauchi (lowercase & spaced)
		"FC03B06AG12",     // Abuja FCT (compact)
		"INVALID_CODE_99", // Malformed
	}

	for p, err := range postcode.ParseSeq(sampleInputs) {
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			continue
		}
		fmt.Printf("✅ Parsed: %-15s (State: %s, LGA: %s, District: %s, Area: %s, Unit: %s)\n",
			p.Formatted(), p.State(), p.LGA(), p.District(), p.Area(), p.BuildingUnit())
	}

	fmt.Println("\n==============================================================")
	fmt.Println(" 2. Local Simulator (Official Test Codes)")
	fmt.Println("==============================================================")

	// Spin up local simulator populated with 20 official NIPOST test postcodes
	srv := simulator.NewServer()
	defer srv.Close()
	fmt.Printf("Simulator running at: %s\n", srv.URL)

	// Initialize OpenTelemetry
	tel, err := otelpostcode.NewTelemetry()
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}

	// 2a. Public Client (No API Key)
	publicClient, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithTelemetry(tel),
		postcode.WithTimeout(3*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create public client: %v", err)
	}

	fmt.Println("\n--- [A] Public Level 1 Lookup (Free, No API Key Required) ---")
	testCode := "AK-11-I61-ZF-12"
	l1Res, err := publicClient.Lookup(ctx, testCode, postcode.Level1)
	if err != nil {
		log.Fatalf("L1 Lookup failed: %v", err)
	}
	fmt.Printf("Code: %s | Valid: %t\n", l1Res.Postcode, l1Res.Valid)

	fmt.Println("\n--- [B] Level 2 Lookup Without API Key (Expecting 401 Auth Required) ---")
	_, err = publicClient.Lookup(ctx, testCode, postcode.Level2)
	var apiErr *postcode.APIError
	if errors.As(err, &apiErr) && apiErr.IsUnauthorized() {
		fmt.Printf("🔒 Caught expected auth error: HTTP %d [%s]: %s\n",
			apiErr.StatusCode, apiErr.Code, apiErr.Message)
	} else {
		log.Fatalf("Expected 401 unauthorized, got %v", err)
	}

	// 2b. Authenticated Client (With API Key)
	authClient, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithAPIKey("nipost_live_test_secret_key"),
		postcode.WithTelemetry(tel),
	)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	fmt.Println("\n--- [C] Authenticated Level 2 Lookup (Lagos: Street & Locality) ---")
	l2Res, err := authClient.Lookup(ctx, "LA-11-W06-TC-10", postcode.Level2)
	if err != nil {
		log.Fatalf("L2 Lookup failed: %v", err)
	}
	fmt.Printf("Postcode: %s\n", l2Res.Postcode)
	fmt.Printf("  State:    %s (%s)\n", l2Res.AdministrativeAddress.StateName, l2Res.AdministrativeAddress.State)
	fmt.Printf("  LGA:      %s (%s)\n", l2Res.AdministrativeAddress.LGAName, l2Res.AdministrativeAddress.LGA)
	fmt.Printf("  District: %s\n", l2Res.AdministrativeAddress.DistrictName)
	fmt.Printf("  Address:  %s\n", l2Res.RecentHouseAddress.Address)

	fmt.Println("\n--- [D] Authenticated Level 3 Lookup (Official Docs Example: EK-01-A03-FK-01) ---")
	l3Res, err := authClient.Lookup(ctx, "EK-01-A03-FK-01", postcode.Level3)
	if err != nil {
		log.Fatalf("L3 Lookup failed: %v", err)
	}
	fmt.Printf("Postcode: %s\n", l3Res.Postcode)
	fmt.Printf("  State:         %s\n", l3Res.AdministrativeAddress.StateName)
	fmt.Printf("  LGA:           %s\n", l3Res.AdministrativeAddress.LGAName)
	fmt.Printf("  Locality:      %s\n", l3Res.AdministrativeAddress.LocalityName)
	fmt.Printf("  Zone:          %s\n", l3Res.AdministrativeAddress.Zone)
	fmt.Printf("  Recent House:  %s\n", l3Res.RecentHouseAddress.Recent)
	fmt.Printf("  Building Use:  %s\n", l3Res.BuildingUseStatus)

	fmt.Println("\n==============================================================")
	fmt.Println(" 3. Autocomplete & Spatial Search")
	fmt.Println("==============================================================")

	// Autocomplete
	autoRes, err := authClient.Autocomplete(ctx, "OG-14")
	if err != nil {
		log.Fatalf("Autocomplete failed: %v", err)
	}
	fmt.Printf("Autocomplete for 'OG-14': %d suggestion(s) found:\n", len(autoRes.Suggestions))
	for _, s := range autoRes.Suggestions {
		fmt.Printf("  • %-15s - %s\n", s.Code, s.Label)
	}

	// Nearby search (Ikeja Town Centre area)
	nearbyRes, err := authClient.Nearby(ctx, postcode.NearbyParams{
		Latitude:  6.6018,
		Longitude: 3.3515,
		RadiusM:   200,
	})
	if err != nil {
		log.Fatalf("Nearby search failed: %v", err)
	}
	fmt.Printf("\nNearby search at (6.6018, 3.3515): %d unit(s) found within 200m:\n", len(nearbyRes.Results))
	for _, u := range nearbyRes.Results {
		fmt.Printf("  • %-15s (dist: %.1fm) - %s, %s\n", u.Postcode, u.DistanceM, u.Address, u.StateName)
	}

	// Reverse geocoding
	revRes, err := authClient.Reverse(ctx, postcode.ReverseParams{
		Latitude:     6.6018,
		Longitude:    3.3515,
		MaxDistanceM: 25,
	})
	if err != nil {
		log.Fatalf("Reverse geocoding failed: %v", err)
	}
	if revRes.Found && revRes.Unit != nil {
		fmt.Printf("\nReverse Geocoded (6.6018, 3.3515) -> %s (%s, %s)\n",
			revRes.Unit.Postcode, revRes.Unit.Address, revRes.Unit.StateName)
	}

	fmt.Println("\n==============================================================")
	fmt.Println(" 4. Assembly & Disassembly")
	fmt.Println("==============================================================")

	// Assembly
	assembled, err := authClient.Assemble(ctx, postcode.Segments{
		State:    "EB",
		LGA:      "13",
		District: "G95",
		Area:     "FR",
		Unit:     "90",
	})
	if err != nil {
		log.Fatalf("Assemble failed: %v", err)
	}
	fmt.Printf("Assembled Segments: Canonical=%s | Display=%s | Compact=%s\n",
		assembled.Postcode, assembled.Display, assembled.Compact)

	// Disassembly
	dis, err := authClient.Disassemble(ctx, assembled.Postcode)
	if err != nil {
		log.Fatalf("Disassemble failed: %v", err)
	}
	fmt.Printf("Disassembled back: State=%s, LGA=%s, District=%s, Area=%s, Unit=%s\n",
		dis.State, dis.LGA, dis.District, dis.Area, dis.Unit)

	fmt.Println("\n==============================================================")
	fmt.Println("⚡ 5. Gateway Rate-Limiting Metrics")
	fmt.Println("==============================================================")
	if rl := authClient.RateLimit(); rl != nil {
		fmt.Printf("Live Gateway Quota: %d/%d remaining in current window\n", rl.Remaining, rl.Limit)
	}

	fmt.Println("\n==============================================================")
	fmt.Println(" All tests and simulator scenarios passed successfully!")
	fmt.Println("==============================================================")
}
