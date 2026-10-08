package postcode

import (
	"cmp"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var (
	diskCacheMu   sync.Mutex
	diskCachePath string
	cacheOnce     sync.Once
)

// DefaultCachePath returns the standard user-level cache file path: ~/.postcode/cache.json
func DefaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".postcode_cache.json"
	}
	return filepath.Join(home, ".postcode", "cache.json")
}

func getActiveCachePathLocked() string {
	if diskCachePath != "" {
		return diskCachePath
	}
	diskCachePath = DefaultCachePath()
	return diskCachePath
}

// SetCachePath overrides the default cache file path and loads existing entries.
func SetCachePath(path string) error {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()
	diskCachePath = path
	return loadDiskCacheLocked()
}

// GetCachePath returns the active cache file path.
func GetCachePath() string {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()
	return getActiveCachePathLocked()
}

func ensureCacheLoaded() {
	cacheOnce.Do(func() {
		_ = LoadDiskCache()
	})
}

// LoadDiskCache ensures disk cache entries from ~/.postcode/cache.json (or custom path)
// are loaded into the offline building registry.
func LoadDiskCache() error {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()
	return loadDiskCacheLocked()
}

func loadDiskCacheLocked() error {
	path := getActiveCachePathLocked()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var records []BuildingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		var m map[string]BuildingRecord
		if mErr := json.Unmarshal(data, &m); mErr == nil {
			for _, rec := range m {
				records = append(records, rec)
			}
		} else {
			return fmt.Errorf("parsing cache %s: %w", path, err)
		}
	}

	customMu.Lock()
	for _, rec := range records {
		if parsed, err := Parse(rec.Postcode); err == nil {
			customBuildings[parsed.Raw()] = rec
		} else {
			customBuildings[rec.Postcode] = rec
		}
	}
	customMu.Unlock()

	return nil
}

// SaveDiskCache persists all registered custom building records to the cache file.
func SaveDiskCache() error {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()

	path := getActiveCachePathLocked()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	customMu.RLock()
	records := make([]BuildingRecord, 0, len(customBuildings))
	for _, r := range customBuildings {
		records = append(records, r)
	}
	customMu.RUnlock()

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache records: %w", err)
	}

	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("writing cache tempfile: %w", err)
	}
	if err := os.Rename(tmpFile, path); err != nil {
		return fmt.Errorf("finalizing cache file: %w", err)
	}

	return nil
}

// ClearDiskCache removes all custom in-memory buildings and deletes the persistent cache file.
func ClearDiskCache() error {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()

	customMu.Lock()
	customBuildings = make(map[string]BuildingRecord)
	customMu.Unlock()

	path := getActiveCachePathLocked()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// GetCacheStats returns the number of custom cached building records and the file location.
func GetCacheStats() (int, string) {
	ensureCacheLoaded()
	diskCacheMu.Lock()
	path := getActiveCachePathLocked()
	diskCacheMu.Unlock()

	customMu.RLock()
	count := len(customBuildings)
	customMu.RUnlock()

	return count, path
}

// LoadBuildingsJSON reads building records from a JSON array or object reader and stores them in cache.
func LoadBuildingsJSON(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("reading json: %w", err)
	}

	var records []BuildingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		var m map[string]BuildingRecord
		if mErr := json.Unmarshal(data, &m); mErr != nil {
			return 0, fmt.Errorf("invalid json format: %w", err)
		}
		for _, rec := range m {
			records = append(records, rec)
		}
	}

	count := 0
	for _, rec := range records {
		if rec.Postcode == "" {
			continue
		}
		RegisterKnownBuilding(rec)
		count++
	}

	if count > 0 {
		_ = SaveDiskCache()
	}

	return count, nil
}

// LoadBuildingsCSV reads building records from a CSV reader.
// Expected header format: postcode,latitude,longitude,address,state_code,state_name,lga_code,lga_name,zone
func LoadBuildingsCSV(r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("reading csv headers: %w", err)
	}

	colIdx := make(map[string]int)
	for i, h := range headers {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	postcodeIdx, ok := colIdx["postcode"]
	if !ok {
		return 0, fmt.Errorf("csv missing required 'postcode' column")
	}
	latIdx, ok := colIdx["latitude"]
	if !ok {
		latIdx, ok = colIdx["lat"]
		if !ok {
			return 0, fmt.Errorf("csv missing required 'latitude' or 'lat' column")
		}
	}
	lngIdx, ok := colIdx["longitude"]
	if !ok {
		lngIdx, ok = colIdx["lng"]
		if !ok {
			lngIdx, ok = colIdx["long"]
			if !ok {
				return 0, fmt.Errorf("csv missing required 'longitude' or 'lng' column")
			}
		}
	}

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("reading csv row: %w", err)
		}

		if len(row) <= postcodeIdx || len(row) <= latIdx || len(row) <= lngIdx {
			continue
		}

		rawPC := strings.TrimSpace(row[postcodeIdx])
		if rawPC == "" {
			continue
		}

		lat, err := strconv.ParseFloat(strings.TrimSpace(row[latIdx]), 64)
		if err != nil {
			continue
		}
		lng, err := strconv.ParseFloat(strings.TrimSpace(row[lngIdx]), 64)
		if err != nil {
			continue
		}

		rec := BuildingRecord{
			Postcode:  rawPC,
			Latitude:  lat,
			Longitude: lng,
		}

		if idx, ok := colIdx["address"]; ok && idx < len(row) {
			rec.Address = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIdx["state_code"]; ok && idx < len(row) {
			rec.StateCode = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIdx["state_name"]; ok && idx < len(row) {
			rec.StateName = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIdx["lga_code"]; ok && idx < len(row) {
			rec.LGACode = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIdx["lga_name"]; ok && idx < len(row) {
			rec.LGAName = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIdx["zone"]; ok && idx < len(row) {
			rec.Zone = strings.TrimSpace(row[idx])
		}

		RegisterKnownBuilding(rec)
		count++
	}

	if count > 0 {
		_ = SaveDiskCache()
	}

	return count, nil
}

// ExportBuildingsJSON writes all cached custom building records to a JSON writer.
func ExportBuildingsJSON(w io.Writer) error {
	ensureCacheLoaded()
	customMu.RLock()
	records := make([]BuildingRecord, 0, len(customBuildings))
	for _, r := range customBuildings {
		records = append(records, r)
	}
	customMu.RUnlock()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

// AllBuildingRecords returns all known building benchmark records and loaded custom cached buildings.
func AllBuildingRecords() []BuildingRecord {
	ensureCacheLoaded()
	var res []BuildingRecord
	for _, b := range knownBuildings {
		res = append(res, b)
	}
	customMu.RLock()
	for _, b := range customBuildings {
		res = append(res, b)
	}
	customMu.RUnlock()
	return res
}

func haversineM(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000 // Earth's mean radius in meters
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

// SearchNearbyBuildingsOffline searches local embedded and cached buildings within radiusM meters of (lat, lng).
func SearchNearbyBuildingsOffline(lat, lng, radiusM float64) []NearbyUnit {
	all := AllBuildingRecords()
	if radiusM <= 0 {
		radiusM = 300.0
	}
	var results []NearbyUnit
	for _, b := range all {
		dist := haversineM(lat, lng, b.Latitude, b.Longitude)
		if dist <= radiusM {
			conf := "high"
			if dist > 50 {
				conf = "medium"
			}
			p, err := Parse(b.Postcode)
			canonical := b.Postcode
			display := b.Postcode
			if err == nil {
				canonical = p.Formatted()
				display = p.String()
			}
			results = append(results, NearbyUnit{
				Postcode:   canonical,
				Display:    display,
				DistanceM:  math.Round(dist*10) / 10,
				Confidence: conf,
				StateName:  b.StateName,
				LGAName:    b.LGAName,
				Address:    b.Address,
			})
		}
	}
	slices.SortFunc(results, func(a, b NearbyUnit) int {
		return cmp.Compare(a.DistanceM, b.DistanceM)
	})
	return results
}

// ReverseCoordinatesOffline resolves geographic coordinates to the closest building unit in the local database.
func ReverseCoordinatesOffline(lat, lng, maxDistanceM float64) *ReverseResponse {
	if maxDistanceM <= 0 {
		maxDistanceM = 25.0
	}
	if maxDistanceM > 250.0 {
		maxDistanceM = 250.0
	}
	units := SearchNearbyBuildingsOffline(lat, lng, maxDistanceM)
	res := &ReverseResponse{
		Coordinate: []float64{lng, lat},
		RadiusM:    maxDistanceM,
	}
	if len(units) > 0 {
		best := units[0]
		res.Found = true
		res.Unit = &best
		res.State = best.StateName
		if p, err := Parse(best.Postcode); err == nil {
			res.State = p.State()
			res.District = p.Formatted()[:8]
			res.Area = p.Formatted()[:11]
		}
	} else {
		res.Found = false
		res.Message = "no registered units found within search radius in offline database"
	}
	return res
}
