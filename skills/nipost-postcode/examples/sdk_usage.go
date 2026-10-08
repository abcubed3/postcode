package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abcubed3/postcode"
)

func main() {
	ctx := context.Background()

	// 1. Zero-allocation offline parse (~28ns, 0 B/op)
	codeStr := "LA-11-W06-TC-10"
	p, err := postcode.Parse(codeStr)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("State: %s, LGA: %s, District: %s, Area: %s, Unit: %s\n",
		p.State(), p.LGA(), p.District(), p.Area(), p.BuildingUnit())
	fmt.Printf("Canonical formatted: %s\n", p.Formatted())

	// 2. Offline location & coordinate centroid resolution (~115ns, $0 cost)
	loc, err := postcode.ResolveLocation(codeStr)
	if err != nil {
		log.Fatalf("Location error: %v", err)
	}
	fmt.Printf("Coordinates: %.6f, %.6f (Precision: %s)\n",
		loc.Latitude, loc.Longitude, loc.Precision)
	fmt.Printf("Google Maps: %s\n", loc.GoogleMapsURL())

	// 3. High-throughput client with cache, timeout, and agent guardrails
	apiKey := os.Getenv("POSTCODE_API_KEY")
	client, err := postcode.NewClient(
		postcode.WithAPIKey(apiKey),
		postcode.WithTimeout(5*time.Second),
		postcode.WithDefaultCache(), // 1,000 entries, 30m TTL
		postcode.WithAgentGuard(postcode.AgentGuardConfig{
			MaxCommercialCallsPerRun: 10,   // Quota ceiling for Level 2/3
			MaxTotalCallsPerRun:      50,   // Hard ceiling for all requests
			AutoDowngradeToLevel1:    true, // Downgrade to free tier on budget depletion
			AutoFallbackToOffline:    true, // Offline centroid on 429/402
		}),
	)
	if err != nil {
		log.Fatalf("Client creation failed: %v", err)
	}

	// 4. Cost-aware lookup: default to Level 1 (Free / Public existence check)
	res, err := client.Lookup(ctx, codeStr, postcode.Level1)
	if err != nil {
		var apiErr *postcode.APIError
		if errors.As(err, &apiErr) {
			if apiErr.IsRateLimit() {
				log.Printf("Rate limit hit. Retry-After: %v", apiErr.RetryAfter)
			} else if apiErr.IsInsufficientCredits() {
				log.Printf("Insufficient commercial credits for paid lookup.")
			}
		}
		log.Printf("Lookup warning: %v", err)
	} else {
		fmt.Printf("Gateway verified: valid=%t, postcode=%s\n", res.Valid, res.Postcode)
	}

	// 5. Actionable self-healing diagnostic inspection
	invalidCode := "ZZ-00-A03-FK-00"
	report := postcode.Diagnose(invalidCode)
	fmt.Printf("Diagnosis for '%s': Format Score = %.1f/100\n", invalidCode, report.FormatScore)
	fmt.Printf("Actionable Tip: %s\n", report.ActionableTip)

	// 6. Range-over-func streaming iterator for batch slices
	samples := []string{"LA-11-W06-TC-10", "INVALID", "EK-01-A03-FK-01"}
	for parsed, pErr := range postcode.ParseSeq(samples) {
		if pErr != nil {
			fmt.Printf("Skipping invalid entry: %v\n", pErr)
			continue
		}
		fmt.Printf("Stream parsed valid code: %s (State: %s)\n", parsed.Formatted(), parsed.State())
	}
}
