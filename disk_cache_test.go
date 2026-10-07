package postcode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiskCache_Lifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test_cache.json")

	if err := SetCachePath(cacheFile); err != nil {
		t.Fatalf("SetCachePath failed: %v", err)
	}

	// Clean initial state
	if err := ClearDiskCache(); err != nil {
		t.Fatalf("ClearDiskCache failed: %v", err)
	}

	count, path := GetCacheStats()
	if count != 0 {
		t.Errorf("expected 0 cached items, got %d", count)
	}
	if path != cacheFile {
		t.Errorf("expected path %s, got %s", cacheFile, path)
	}

	// Register a test building
	RegisterKnownBuilding(BuildingRecord{
		Postcode:  "LA-08-A86-RG-99",
		Latitude:  6.4765,
		Longitude: 3.6335,
		Address:   "99 Lekki Test Way",
		StateCode: "LA",
		StateName: "Lagos",
		LGACode:   "08",
		LGAName:   "Eti-Osa",
		Zone:      "SOUTH WEST",
	})

	if err := SaveDiskCache(); err != nil {
		t.Fatalf("SaveDiskCache failed: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Fatalf("cache file %s was not created", cacheFile)
	}

	// Verify offline resolution picks up the cached record
	loc, err := ResolveLocation("LA-08-A86-RG-99")
	if err != nil {
		t.Fatalf("ResolveLocation failed: %v", err)
	}
	if loc.Precision != PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding, got %s", loc.Precision)
	}
	if loc.Latitude != 6.4765 || loc.Longitude != 3.6335 {
		t.Errorf("unexpected coordinates: %f, %f", loc.Latitude, loc.Longitude)
	}
	if loc.Address != "99 Lekki Test Way" {
		t.Errorf("unexpected address: %s", loc.Address)
	}
}

func TestDiskCache_LoadJSONAndCSV(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "import_cache.json")
	_ = SetCachePath(cacheFile)
	_ = ClearDiskCache()

	// JSON import
	jsonData := `[
		{
			"Postcode": "FC-03-B06-AG-99",
			"Latitude": 9.0580,
			"Longitude": 7.4955,
			"Address": "99 Shehu Shagari Way, Garki, Abuja"
		}
	]`
	count, err := LoadBuildingsJSON(strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("LoadBuildingsJSON failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 imported, got %d", count)
	}

	loc, err := ResolveLocation("FC-03-B06-AG-99")
	if err != nil || loc.Precision != PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding from JSON import, got %v (%s)", err, loc.Precision)
	}

	// CSV import
	csvData := `postcode,latitude,longitude,address
LA-11-W06-TC-99,6.6020,3.3520,"99 Obafemi Awolowo Way, Ikeja"`
	csvCount, err := LoadBuildingsCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("LoadBuildingsCSV failed: %v", err)
	}
	if csvCount != 1 {
		t.Errorf("expected 1 imported CSV, got %d", csvCount)
	}

	locCSV, err := ResolveLocation("LA-11-W06-TC-99")
	if err != nil || locCSV.Precision != PrecisionBuilding {
		t.Errorf("expected PrecisionBuilding from CSV import, got %v (%s)", err, locCSV.Precision)
	}
}
