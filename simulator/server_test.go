package simulator

import (
	"context"
	"errors"
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
