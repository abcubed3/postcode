package simulator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/abcubed3/postcode"
)

// PostcodeRecord stores test metadata for an official test postcode.
type PostcodeRecord struct {
	Canonical    string  `json:"canonical"`
	StateCode    string  `json:"state_code"`
	StateName    string  `json:"state_name"`
	LGACode      string  `json:"lga_code"`
	LGAName      string  `json:"lga_name"`
	DistrictCode string  `json:"district_code"`
	DistrictName string  `json:"district_name"`
	AreaCode     string  `json:"area_code"`
	AreaName     string  `json:"area_name"`
	UnitCode     string  `json:"unit_code"`
	Zone         string  `json:"zone"`
	RecentHouse  string  `json:"recent_house"`
	BuildingUse  string  `json:"building_use"`
	Lat          float64 `json:"latitude"`
	Lng          float64 `json:"longitude"`
}

// UnmarshalJSON supports both "latitude"/"longitude" and "lat"/"lng" field names.
func (p *PostcodeRecord) UnmarshalJSON(data []byte) error {
	type Alias PostcodeRecord
	aux := struct {
		*Alias
		AltLat *float64 `json:"lat"`
		AltLng *float64 `json:"lng"`
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if p.Lat == 0 && aux.AltLat != nil {
		p.Lat = *aux.AltLat
	}
	if p.Lng == 0 && aux.AltLng != nil {
		p.Lng = *aux.AltLng
	}
	return nil
}

//go:embed data/test_postcodes.json
var defaultPostcodesJSON []byte

// TestPostcodes contains the active in-memory test postcodes.
var TestPostcodes []PostcodeRecord

var (
	dbMu          sync.RWMutex
	dbByCanonical map[string]PostcodeRecord
)

func init() {
	if err := LoadDefaultPostcodes(); err != nil {
		panic(fmt.Sprintf("simulator: failed to parse embedded test_postcodes.json: %v", err))
	}
}

// LoadDefaultPostcodes resets TestPostcodes and indexes to the embedded official dataset.
func LoadDefaultPostcodes() error {
	return LoadJSON(defaultPostcodesJSON)
}

// LoadJSON loads test postcodes from raw JSON bytes into the simulator.
func LoadJSON(data []byte) error {
	var records []PostcodeRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("parsing postcodes JSON: %w", err)
	}
	return LoadRecords(records)
}

// LoadReader reads test postcodes from an io.Reader and updates the simulator.
func LoadReader(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading postcodes data: %w", err)
	}
	return LoadJSON(data)
}

// LoadFile reads test postcodes from a file path and updates the simulator.
func LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}
	return LoadJSON(data)
}

// LoadRecords sets the active records and rebuilds internal lookup indexes.
func LoadRecords(records []PostcodeRecord) error {
	byCanonical := make(map[string]PostcodeRecord, len(records))

	for _, rec := range records {
		byCanonical[rec.Canonical] = rec
	}

	dbMu.Lock()
	TestPostcodes = records
	dbByCanonical = byCanonical
	dbMu.Unlock()

	return nil
}

var (
	simMu           sync.Mutex
	simTransient429 = make(map[string]int)
)

// ResetSimulatedRateLimits clears all recorded transient rate-limit counters in the simulator.
func ResetSimulatedRateLimits() {
	simMu.Lock()
	defer simMu.Unlock()
	simTransient429 = make(map[string]int)
}

// NewHandler constructs an http.Handler that precisely simulates api.postcode.gov.ng.
func NewHandler() http.Handler {
	mux := http.NewServeMux()

	// GET /v1/lookup?code=...&level=...
	mux.HandleFunc("GET /v1/lookup", handleLookup)

	// GET /v1/search/autocomplete?q=...
	mux.HandleFunc("GET /v1/search/autocomplete", handleAutocomplete)

	// GET /v1/search/nearby?lat=...&lng=...&radius=...
	mux.HandleFunc("GET /v1/search/nearby", handleNearby)

	// GET /v1/search/reverse?lat=...&lng=...&max_distance_m=...
	mux.HandleFunc("GET /v1/search/reverse", handleReverse)

	// POST /v1/assembly/assemble
	mux.HandleFunc("POST /v1/assembly/assemble", handleAssemble)

	// GET /v1/assembly/disassemble?code=...
	mux.HandleFunc("GET /v1/assembly/disassemble", handleDisassemble)

	// GET /healthz
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return rateLimitMiddleware(mux)
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set standard official NIPOST rate limit headers (limit: 600)
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "597")

		// Simulation hook: permanent 429 rate limit exceeded (cooldown: 60s)
		if r.Header.Get("X-Simulate-Rate-Limit-Exceed") != "" {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded; please wait before retrying")
			return
		}

		// Simulation hook: transient 429 that fails once then succeeds on retry
		if key := r.Header.Get("X-Simulate-Transient-429"); key != "" {
			simMu.Lock()
			count, exists := simTransient429[key]
			if !exists {
				count = 1
			}
			if count > 0 {
				simTransient429[key] = count - 1
				simMu.Unlock()
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", "0")
				writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded; please wait before retrying")
				return
			}
			simMu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// NewServer starts an in-memory httptest.Server simulating api.postcode.gov.ng.
func NewServer() *httptest.Server {
	return httptest.NewServer(NewHandler())
}

// ListenAndServe starts a live HTTP server listening on the specified network address.
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, NewHandler())
}

func handleLookup(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "code parameter is required")
		return
	}

	levelStr := r.URL.Query().Get("level")
	level := 1
	if levelStr != "" {
		parsed, err := strconv.Atoi(levelStr)
		if err == nil && parsed >= 1 && parsed <= 5 {
			level = parsed
		}
	}

	apiKey := r.Header.Get("X-API-Key")
	// Level 2+ requires commercial authentication
	if level >= 2 && apiKey == "" {
		writeError(w, http.StatusUnauthorized, "auth_required", "an API key is required; pass it in the X-API-Key header")
		return
	}

	p, err := postcode.Parse(code)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"postcode": code,
				"valid":    false,
			},
		})
		return
	}

	dbMu.RLock()
	rec, found := dbByCanonical[p.Formatted()]
	dbMu.RUnlock()
	if !found {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"postcode": p.Formatted(),
				"valid":    false,
			},
		})
		return
	}

	// Build graded payload
	data := map[string]any{
		"postcode": rec.Canonical,
		"valid":    true,
	}

	if level >= 2 {
		data["administrative_address"] = map[string]any{
			"state":         rec.StateCode,
			"state_name":    rec.StateName,
			"lga":           rec.LGACode,
			"lga_name":      rec.LGAName,
			"district":      rec.DistrictCode,
			"district_name": rec.DistrictName,
			"area":          rec.AreaCode,
			"area_name":     rec.AreaName,
			"unit":          rec.UnitCode,
			"zone":          rec.Zone,
			"locality_name": rec.LGAName,
		}
		data["recent_house_address"] = map[string]any{
			"address": rec.RecentHouse,
			"recent":  rec.RecentHouse,
		}
	}

	if level >= 3 {
		data["building_use_status"] = rec.BuildingUse
	}

	if level >= 5 {
		data["point_geometry"] = map[string]any{
			"type":        "Point",
			"coordinates": []float64{rec.Lng, rec.Lat},
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func handleAutocomplete(w http.ResponseWriter, r *http.Request) {
	q := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("q")))
	if q == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "query parameter 'q' is required")
		return
	}

	type suggestion struct {
		Code  string `json:"code"`
		Label string `json:"label"`
	}

	dbMu.RLock()
	defer dbMu.RUnlock()

	var suggestions []suggestion
	seen := make(map[string]bool)

	for _, rec := range TestPostcodes {
		if strings.HasPrefix(rec.Canonical, q) || strings.HasPrefix(strings.ReplaceAll(rec.Canonical, "-", " "), q) {
			if !seen[rec.Canonical] {
				seen[rec.Canonical] = true
				suggestions = append(suggestions, suggestion{
					Code:  rec.Canonical,
					Label: rec.RecentHouse,
				})
			}
		}
		if len(suggestions) >= 5 {
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"segment":     "postcode",
			"suggestions": suggestions,
		},
	})
}

func handleNearby(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	radius := 300.0
	if radStr := r.URL.Query().Get("radius"); radStr != "" {
		if parsed, err := strconv.ParseFloat(radStr, 64); err == nil && parsed > 0 {
			radius = min(parsed, 300.0)
		}
	}

	dbMu.RLock()
	defer dbMu.RUnlock()

	var results []map[string]any
	for _, rec := range TestPostcodes {
		dist := haversineM(lat, lng, rec.Lat, rec.Lng)
		if dist <= radius {
			results = append(results, map[string]any{
				"postcode":   rec.Canonical,
				"display":    strings.ReplaceAll(rec.Canonical, "-", " "),
				"distance_m": math.Round(dist*10) / 10,
				"confidence": "high",
				"state_name": rec.StateName,
				"lga_name":   rec.LGAName,
				"address":    rec.RecentHouse,
			})
		}
	}

	if results == nil {
		results = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"results": results,
		},
	})
}

func handleReverse(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	maxDist := 25.0
	if dStr := r.URL.Query().Get("max_distance_m"); dStr != "" {
		if parsed, err := strconv.ParseFloat(dStr, 64); err == nil && parsed > 0 {
			maxDist = min(parsed, 250.0)
		}
	}

	dbMu.RLock()
	defer dbMu.RUnlock()

	var bestRec *PostcodeRecord
	bestDist := math.MaxFloat64

	for i := range TestPostcodes {
		rec := &TestPostcodes[i]
		dist := haversineM(lat, lng, rec.Lat, rec.Lng)
		if dist <= maxDist && dist < bestDist {
			bestDist = dist
			bestRec = rec
		}
	}

	if bestRec == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"found":      false,
				"coordinate": []float64{lng, lat},
				"message":    "no units found within search radius",
				"radius_m":   maxDist,
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"found":      true,
			"coordinate": []float64{lng, lat},
			"unit": map[string]any{
				"postcode":      bestRec.Canonical,
				"display":       strings.ReplaceAll(bestRec.Canonical, "-", " "),
				"distance_m":    math.Round(bestDist*10) / 10,
				"confidence":    "high",
				"state_name":    bestRec.StateName,
				"lga_name":      bestRec.LGAName,
				"locality_name": bestRec.LGAName,
				"address":       bestRec.RecentHouse,
			},
			"area":     bestRec.Canonical[:11],
			"district": bestRec.Canonical[:8],
			"state":    bestRec.StateCode,
			"radius_m": maxDist,
		},
	})
}

func handleAssemble(w http.ResponseWriter, r *http.Request) {
	var segs postcode.Segments
	if err := json.NewDecoder(r.Body).Decode(&segs); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	raw := segs.State + segs.LGA + segs.District + segs.Area + segs.Unit
	p, err := postcode.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_segments", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"postcode": p.Formatted(),
			"display":  p.String(),
			"compact":  p.Raw(),
		},
	})
}

func handleDisassemble(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	p, err := postcode.Parse(code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_postcode", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"state":    p.State(),
			"lga":      p.LGA(),
			"district": p.District(),
			"area":     p.Area(),
			"unit":     p.BuildingUnit(),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if w.Header().Get("X-RateLimit-Limit") == "" {
		w.Header().Set("X-RateLimit-Limit", "600")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		w.Header().Set("X-RateLimit-Remaining", "597")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": msg,
		},
	})
}

// haversineM calculates approximate distance in meters between two lat/lng coordinates.
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
