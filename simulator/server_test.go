package simulator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/abcubed3/postcode"
)

func TestSimulator_LookupLevels(t *testing.T) {
	t.Parallel()
	srv := NewServer()
	defer srv.Close()

	// 1. Level 1 without API key (free public endpoint)
	publicClient, err := postcode.NewClient(postcode.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	l1Res, err := publicClient.Lookup(context.Background(), "EK-01-A03-FK-01", postcode.Level1)
	if err != nil {
		t.Fatalf("Level 1 lookup failed: %v", err)
	}
	if !l1Res.Valid {
		t.Errorf("l1Res.Valid = false, want true")
	}
	if l1Res.AdministrativeAddress != nil {
		t.Errorf("Level 1 should not have administrative address")
	}

	// 2. Level 2 without API key should fail with 401 auth_required
	_, err = publicClient.Lookup(context.Background(), "EK-01-A03-FK-01", postcode.Level2)
	if err == nil {
		t.Fatal("expected error on L2 without API key, got nil")
	}
	var apiErr *postcode.APIError
	if !errors.As(err, &apiErr) || !apiErr.IsUnauthorized() {
		t.Errorf("expected 401 unauthorized, got %v", err)
	}

	// 3. Level 2 and Level 3 with API key
	authClient, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithAPIKey("nipost_live_test_key"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	l2Res, err := authClient.Lookup(context.Background(), "LA-11-W06-TC-10", postcode.Level2)
	if err != nil {
		t.Fatalf("Level 2 lookup failed: %v", err)
	}
	if l2Res.AdministrativeAddress == nil || l2Res.AdministrativeAddress.StateName != "LAGOS" {
		t.Errorf("expected LAGOS state, got %+v", l2Res.AdministrativeAddress)
	}
	if l2Res.BuildingUseStatus != "" {
		t.Errorf("Level 2 should not include building use status")
	}

	l3Res, err := authClient.Lookup(context.Background(), "FC-03-B06-AG-12", postcode.Level3)
	if err != nil {
		t.Fatalf("Level 3 lookup failed: %v", err)
	}
	if l3Res.BuildingUseStatus != "government" {
		t.Errorf("expected government building use, got %q", l3Res.BuildingUseStatus)
	}

	// 4. Verify official documentation example EK-01-A03-FK-01
	ekRes, err := authClient.Lookup(context.Background(), "EK-01-A03-FK-01", postcode.Level3)
	if err != nil {
		t.Fatalf("Ekiti lookup failed: %v", err)
	}
	if ekRes.AdministrativeAddress.LocalityName != "ADO EKITI" || ekRes.AdministrativeAddress.Zone != "SOUTH WEST" {
		t.Errorf("unexpected Ekiti admin address: %+v", ekRes.AdministrativeAddress)
	}
	if ekRes.RecentHouseAddress.Recent == "" || ekRes.BuildingUseStatus != "residential" {
		t.Errorf("unexpected Ekiti recent/use: %+v, status=%s", ekRes.RecentHouseAddress, ekRes.BuildingUseStatus)
	}
}

func TestSimulator_AllTestPostcodesValid(t *testing.T) {
	t.Parallel()
	srv := NewServer()
	defer srv.Close()

	client, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithAPIKey("nipost_live_test"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	for _, rec := range TestPostcodes {
		res, err := client.Lookup(ctx, rec.Canonical, postcode.Level3)
		if err != nil {
			t.Errorf("Lookup(%s) failed: %v", rec.Canonical, err)
			continue
		}
		if !res.Valid {
			t.Errorf("Lookup(%s) returned valid=false", rec.Canonical)
		}
		if res.AdministrativeAddress == nil || res.AdministrativeAddress.StateName != rec.StateName {
			t.Errorf("Lookup(%s) state = %v, want %s", rec.Canonical, res.AdministrativeAddress, rec.StateName)
		}
	}
}

func TestSimulator_DynamicLoading(t *testing.T) {
	defer func() {
		_ = LoadDefaultPostcodes()
	}()

	customJSON := `[
		{
			"canonical": "LA-01-A01-AA-01",
			"state_code": "LA",
			"state_name": "LAGOS",
			"lga_code": "01",
			"lga_name": "AGEGE",
			"district_code": "A01",
			"district_name": "AGEGE CENTRAL",
			"area_code": "AA",
			"area_name": "STATION ROAD",
			"unit_code": "01",
			"zone": "SOUTH WEST",
			"recent_house": "1 STATION ROAD, AGEGE",
			"building_use": "commercial",
			"lat": 6.6180,
			"lng": 3.3209
		}
	]`

	tmpFile := filepath.Join(t.TempDir(), "custom.json")
	if err := os.WriteFile(tmpFile, []byte(customJSON), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	if err := LoadFile(tmpFile); err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if len(TestPostcodes) != 1 {
		t.Fatalf("expected 1 record after LoadFile, got %d", len(TestPostcodes))
	}
	if TestPostcodes[0].Canonical != "LA-01-A01-AA-01" {
		t.Errorf("expected canonical LA-01-A01-AA-01, got %s", TestPostcodes[0].Canonical)
	}
	if TestPostcodes[0].Lat != 6.6180 || TestPostcodes[0].Lng != 3.3209 {
		t.Errorf("expected lat 6.6180, lng 3.3209, got %f, %f", TestPostcodes[0].Lat, TestPostcodes[0].Lng)
	}

	srv := NewServer()
	defer srv.Close()

	client, err := postcode.NewClient(
		postcode.WithBaseURL(srv.URL),
		postcode.WithAPIKey("test_key"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	res, err := client.Lookup(context.Background(), "LA-01-A01-AA-01", postcode.Level2)
	if err != nil {
		t.Fatalf("lookup of dynamically loaded postcode failed: %v", err)
	}
	if !res.Valid || res.AdministrativeAddress == nil || res.AdministrativeAddress.LGAName != "AGEGE" {
		t.Errorf("unexpected lookup result: %+v", res)
	}

	// Test resetting to default embedded postcodes
	if err := LoadDefaultPostcodes(); err != nil {
		t.Fatalf("LoadDefaultPostcodes failed: %v", err)
	}
	if len(TestPostcodes) <= 1 {
		t.Errorf("expected multiple default records, got %d", len(TestPostcodes))
	}
}

func TestSimulator_ReferenceAndHealthAndAssemble(t *testing.T) {
	t.Parallel()
	srv := NewServer()
	defer srv.Close()

	client, err := postcode.NewClient(postcode.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// 1. Health
	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// 2. Assemble with single-digit zero filling (official docs example)
	assembled, err := client.Assemble(context.Background(), postcode.Segments{
		State:    "ek",
		LGA:      "1",
		District: "a03",
		Area:     "fk",
		Unit:     "1",
	})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if assembled.Postcode != "EK-01-A03-FK-01" {
		t.Errorf("assembled.Postcode = %q, want EK-01-A03-FK-01", assembled.Postcode)
	}

	// 3. Reference States
	states, err := client.ReferenceStates(context.Background())
	if err != nil {
		t.Fatalf("ReferenceStates failed: %v", err)
	}
	if len(states) != 37 {
		t.Errorf("expected 37 states, got %d", len(states))
	}

	// 4. Reference LGAs
	lgas, err := client.ReferenceLGAs(context.Background(), "LA")
	if err != nil {
		t.Fatalf("ReferenceLGAs failed: %v", err)
	}
	if len(lgas) == 0 {
		t.Errorf("expected LGAs for LA, got 0")
	}

	// 5. Reference Districts
	districts, err := client.ReferenceDistricts(context.Background(), "EK", "01")
	if err != nil {
		t.Fatalf("ReferenceDistricts failed: %v", err)
	}
	if len(districts) == 0 {
		t.Errorf("expected districts for EK 01, got 0")
	}

	// 6. Reference Areas
	areas, err := client.ReferenceAreas(context.Background(), "EK", "01", "A03")
	if err != nil {
		t.Fatalf("ReferenceAreas failed: %v", err)
	}
	if len(areas) == 0 {
		t.Errorf("expected areas for EK 01 A03, got 0")
	}
}

