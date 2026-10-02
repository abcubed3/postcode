package simulator

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"

	"github.com/abcubed3/postcode"
)

// PostcodeRecord stores test metadata for an official test postcode.
type PostcodeRecord struct {
	Canonical    string
	StateCode    string
	StateName    string
	LGACode      string
	LGAName      string
	DistrictCode string
	DistrictName string
	AreaCode     string
	AreaName     string
	UnitCode     string
	Zone         string
	RecentHouse  string
	BuildingUse  string
	Lat          float64
	Lng          float64
}

// TestPostcodes contains the official test postcodes from docs.postcode.gov.ng.
var TestPostcodes = []PostcodeRecord{
	{
		Canonical:    "EK-01-A03-FK-01",
		StateCode:    "EK",
		StateName:    "EKITI",
		LGACode:      "01",
		LGAName:      "ADO EKITI",
		DistrictCode: "A03",
		DistrictName: "ADO DISTRICT 03",
		AreaCode:     "FK",
		AreaName:     "FABIAN HOTEL AXIS",
		UnitCode:     "01",
		Zone:         "SOUTH WEST",
		RecentHouse:  "NTA ROAD, BACK OF FABIAN HOTEL, ADO EKITI",
		BuildingUse:  "residential",
		Lat:          7.6211,
		Lng:          5.2215,
	},
	{
		Canonical:    "AK-11-I61-ZF-12",
		StateCode:    "AK",
		StateName:    "AKWA IBOM",
		LGACode:      "11",
		LGAName:      "UYO",
		DistrictCode: "I61",
		DistrictName: "UYO CENTRAL",
		AreaCode:     "ZF",
		AreaName:     "ORON ROAD",
		UnitCode:     "12",
		Zone:         "SOUTH SOUTH",
		RecentHouse:  "12 ORON ROAD, UYO",
		BuildingUse:  "commercial",
		Lat:          5.0377,
		Lng:          7.9128,
	},
	{
		Canonical:    "AK-11-H40-WD-11",
		StateCode:    "AK",
		StateName:    "AKWA IBOM",
		LGACode:      "11",
		LGAName:      "UYO",
		DistrictCode: "H40",
		DistrictName: "IBBESIKPO",
		AreaCode:     "WD",
		AreaName:     "WELLINGTON BASSEY WAY",
		UnitCode:     "11",
		Zone:         "SOUTH SOUTH",
		RecentHouse:  "11 WELLINGTON BASSEY WAY, UYO",
		BuildingUse:  "government",
		Lat:          5.0333,
		Lng:          7.9266,
	},
	{
		Canonical:    "BA-02-M67-BL-69",
		StateCode:    "BA",
		StateName:    "BAUCHI",
		LGACode:      "02",
		LGAName:      "BAUCHI",
		DistrictCode: "M67",
		DistrictName: "BAUCHI CENTRAL",
		AreaCode:     "BL",
		AreaName:     "BANK ROAD",
		UnitCode:     "69",
		Zone:         "NORTH EAST",
		RecentHouse:  "69 BANK ROAD, GRA, BAUCHI",
		BuildingUse:  "commercial",
		Lat:          10.3158,
		Lng:          9.8442,
	},
	{
		Canonical:    "BA-02-E99-NE-30",
		StateCode:    "BA",
		StateName:    "BAUCHI",
		LGACode:      "02",
		LGAName:      "BAUCHI",
		DistrictCode: "E99",
		DistrictName: "YELWA",
		AreaCode:     "NE",
		AreaName:     "NEW GRA",
		UnitCode:     "30",
		Zone:         "NORTH EAST",
		RecentHouse:  "30 AHMADU BELLO WAY, BAUCHI",
		BuildingUse:  "residential",
		Lat:          10.3012,
		Lng:          9.8234,
	},
	{
		Canonical:    "EB-13-G95-FR-90",
		StateCode:    "EB",
		StateName:    "EBONYI",
		LGACode:      "13",
		LGAName:      "ABAKALIKI",
		DistrictCode: "G95",
		DistrictName: "ABAKALIKI URBAN",
		AreaCode:     "FR",
		AreaName:     "FESTUS ROAD",
		UnitCode:     "90",
		Zone:         "SOUTH EAST",
		RecentHouse:  "90 OGOJA ROAD, ABAKALIKI",
		BuildingUse:  "commercial",
		Lat:          6.3249,
		Lng:          8.1137,
	},
	{
		Canonical:    "EB-13-I97-AB-30",
		StateCode:    "EB",
		StateName:    "EBONYI",
		LGACode:      "13",
		LGAName:      "ABAKALIKI",
		DistrictCode: "I97",
		DistrictName: "AZUIYIOKPA",
		AreaCode:     "AB",
		AreaName:     "AGBANI",
		UnitCode:     "30",
		Zone:         "SOUTH EAST",
		RecentHouse:  "30 WATER WORKS ROAD, ABAKALIKI",
		BuildingUse:  "residential",
		Lat:          6.3180,
		Lng:          8.1022,
	},
	{
		Canonical:    "EN-05-V19-CD-22",
		StateCode:    "EN",
		StateName:    "ENUGU",
		LGACode:      "05",
		LGAName:      "ENUGU NORTH",
		DistrictCode: "V19",
		DistrictName: "INDEPENDENCE LAYOUT",
		AreaCode:     "CD",
		AreaName:     "CHIME AVENUE",
		UnitCode:     "22",
		Zone:         "SOUTH EAST",
		RecentHouse:  "22 CHIME AVENUE, NEW HAVEN, ENUGU",
		BuildingUse:  "commercial",
		Lat:          6.4584,
		Lng:          7.5464,
	},
	{
		Canonical:    "EN-05-V19-FT-20",
		StateCode:    "EN",
		StateName:    "ENUGU",
		LGACode:      "05",
		LGAName:      "ENUGU NORTH",
		DistrictCode: "V19",
		DistrictName: "OGUI",
		AreaCode:     "FT",
		AreaName:     "FRANKLIN ROAD",
		UnitCode:     "20",
		Zone:         "SOUTH EAST",
		RecentHouse:  "20 OGUI ROAD, ENUGU",
		BuildingUse:  "residential",
		Lat:          6.4412,
		Lng:          7.5023,
	},
	{
		Canonical:    "FC-03-B06-AG-12",
		StateCode:    "FC",
		StateName:    "FCT",
		LGACode:      "03",
		LGAName:      "ABUJA MUNICIPAL",
		DistrictCode: "B06",
		DistrictName: "GARKI II",
		AreaCode:     "AG",
		AreaName:     "AREA 11",
		UnitCode:     "12",
		Zone:         "NORTH CENTRAL",
		RecentHouse:  "12 SHEHU SHAGARI WAY, GARKI, ABUJA",
		BuildingUse:  "government",
		Lat:          9.0579,
		Lng:          7.4951,
	},
	{
		Canonical:    "FC-02-B19-RT-30",
		StateCode:    "FC",
		StateName:    "FCT",
		LGACode:      "02",
		LGAName:      "BWARI",
		DistrictCode: "B19",
		DistrictName: "KUBWA",
		AreaCode:     "RT",
		AreaName:     "GADO NASKO ROAD",
		UnitCode:     "30",
		Zone:         "NORTH CENTRAL",
		RecentHouse:  "30 GADO NASKO WAY, PHASE 4, KUBWA, ABUJA",
		BuildingUse:  "residential",
		Lat:          9.1538,
		Lng:          7.3220,
	},
	{
		Canonical:    "JI-24-O18-JP-23",
		StateCode:    "JI",
		StateName:    "JIGAWA",
		LGACode:      "24",
		LGAName:      "DUTSE",
		DistrictCode: "O18",
		DistrictName: "DUTSE CENTRAL",
		AreaCode:     "JP",
		AreaName:     "JIGAWA POLY RD",
		UnitCode:     "23",
		Zone:         "NORTH WEST",
		RecentHouse:  "23 SANI ABACHA WAY, DUTSE",
		BuildingUse:  "educational",
		Lat:          11.7594,
		Lng:          9.3389,
	},
	{
		Canonical:    "JI-24-N11-VM-58",
		StateCode:    "JI",
		StateName:    "JIGAWA",
		LGACode:      "24",
		LGAName:      "DUTSE",
		DistrictCode: "N11",
		DistrictName: "TAKURA",
		AreaCode:     "VM",
		AreaName:     "VILLAGE MARKET",
		UnitCode:     "58",
		Zone:         "NORTH WEST",
		RecentHouse:  "58 KANO-DUTSE EXPRESSWAY, DUTSE",
		BuildingUse:  "commercial",
		Lat:          11.7231,
		Lng:          9.3102,
	},
	{
		Canonical:    "KN-31-F82-WJ-80",
		StateCode:    "KN",
		StateName:    "KANO",
		LGACode:      "31",
		LGAName:      "KANO MUNICIPAL",
		DistrictCode: "F82",
		DistrictName: "NASARAWA",
		AreaCode:     "WJ",
		AreaName:     "WEST ROAD",
		UnitCode:     "80",
		Zone:         "NORTH WEST",
		RecentHouse:  "80 BADU ROAD, BOMPAI, KANO",
		BuildingUse:  "industrial",
		Lat:          12.0022,
		Lng:          8.5920,
	},
	{
		Canonical:    "KN-31-D78-IQ-38",
		StateCode:    "KN",
		StateName:    "KANO",
		LGACode:      "31",
		LGAName:      "KANO MUNICIPAL",
		DistrictCode: "D78",
		DistrictName: "FAGGE",
		AreaCode:     "IQ",
		AreaName:     "IBRAHIM TAIWO",
		UnitCode:     "38",
		Zone:         "NORTH WEST",
		RecentHouse:  "38 IBRAHIM TAIWO ROAD, KANO",
		BuildingUse:  "commercial",
		Lat:          12.0150,
		Lng:          8.5201,
	},
	{
		Canonical:    "LA-11-W06-TC-10",
		StateCode:    "LA",
		StateName:    "LAGOS",
		LGACode:      "11",
		LGAName:      "IKEJA",
		DistrictCode: "W06",
		DistrictName: "ALAUSA",
		AreaCode:     "TC",
		AreaName:     "TOWN CENTRE",
		UnitCode:     "10",
		Zone:         "SOUTH WEST",
		RecentHouse:  "10 OBAFEMI AWOLOWO WAY, IKEJA, LAGOS",
		BuildingUse:  "commercial",
		Lat:          6.6018,
		Lng:          3.3515,
	},
	{
		Canonical:    "LA-11-U34-ZR-63",
		StateCode:    "LA",
		StateName:    "LAGOS",
		LGACode:      "11",
		LGAName:      "IKEJA",
		DistrictCode: "U34",
		DistrictName: "GRA IKEJA",
		AreaCode:     "ZR",
		AreaName:     "JOEL OGUNNAIKE",
		UnitCode:     "63",
		Zone:         "SOUTH WEST",
		RecentHouse:  "63 ISAAC JOHN STREET, GRA IKEJA, LAGOS",
		BuildingUse:  "hospitality",
		Lat:          6.5891,
		Lng:          3.3590,
	},
	{
		Canonical:    "NI-09-J67-QC-65",
		StateCode:    "NI",
		StateName:    "NIGER",
		LGACode:      "09",
		LGAName:      "CHANCHAGA",
		DistrictCode: "J67",
		DistrictName: "MINNA CENTRAL",
		AreaCode:     "QC",
		AreaName:     "QUEENS COURT",
		UnitCode:     "65",
		Zone:         "NORTH CENTRAL",
		RecentHouse:  "65 BOSSO ROAD, MINNA",
		BuildingUse:  "commercial",
		Lat:          9.6139,
		Lng:          6.5569,
	},
	{
		Canonical:    "NI-09-A75-DA-10",
		StateCode:    "NI",
		StateName:    "NIGER",
		LGACode:      "09",
		LGAName:      "CHANCHAGA",
		DistrictCode: "A75",
		DistrictName: "BOSSO",
		AreaCode:     "DA",
		AreaName:     "DUTSEN KURA",
		UnitCode:     "10",
		Zone:         "NORTH CENTRAL",
		RecentHouse:  "10 PAIDA ROAD, MINNA",
		BuildingUse:  "residential",
		Lat:          9.6288,
		Lng:          6.5412,
	},
	{
		Canonical:    "OG-14-T18-BN-16",
		StateCode:    "OG",
		StateName:    "OGUN",
		LGACode:      "14",
		LGAName:      "ABEOKUTA SOUTH",
		DistrictCode: "T18",
		DistrictName: "IBARA",
		AreaCode:     "BN",
		AreaName:     "BANKING AXIS",
		UnitCode:     "16",
		Zone:         "SOUTH WEST",
		RecentHouse:  "16 LALUBU STREET, OKE-ILEWO, ABEOKUTA",
		BuildingUse:  "commercial",
		Lat:          7.1475,
		Lng:          3.3619,
	},
	{
		Canonical:    "OG-14-M82-QA-09",
		StateCode:    "OG",
		StateName:    "OGUN",
		LGACode:      "14",
		LGAName:      "ABEOKUTA SOUTH",
		DistrictCode: "M82",
		DistrictName: "OKE-ITOKU",
		AreaCode:     "QA",
		AreaName:     "QUARRY ROAD",
		UnitCode:     "09",
		Zone:         "SOUTH WEST",
		RecentHouse:  "9 QUARRY ROAD, ABEOKUTA",
		BuildingUse:  "residential",
		Lat:          7.1550,
		Lng:          3.3480,
	},
}

var dbByCanonical map[string]PostcodeRecord
var dbByCompact map[string]PostcodeRecord

func init() {
	dbByCanonical = make(map[string]PostcodeRecord, len(TestPostcodes))
	dbByCompact = make(map[string]PostcodeRecord, len(TestPostcodes))
	for _, rec := range TestPostcodes {
		dbByCanonical[rec.Canonical] = rec
		p, err := postcode.Parse(rec.Canonical)
		if err == nil {
			dbByCompact[p.Raw()] = rec
		}
	}
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

	rec, found := dbByCanonical[p.Formatted()]
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
