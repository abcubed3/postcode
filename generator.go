package postcode

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"slices"
	"strings"
	"time"
)

// SyntheticAddress represents a synthesized Nigerian address with ground-truth labels.
// Used to benchmark and evaluate AI agent entity extraction, validation, and geocoding accuracy.
type SyntheticAddress struct {
	RawText           string  `json:"raw_text"`
	ExpectedState     string  `json:"expected_state"`
	ExpectedStateCode string  `json:"expected_state_code"`
	ExpectedLGA       string  `json:"expected_lga,omitempty"`
	ExpectedPostcode  string  `json:"expected_postcode,omitempty"`
	ExpectedLat       float64 `json:"expected_lat,omitempty"`
	ExpectedLng       float64 `json:"expected_lng,omitempty"`
	NoiseType         string  `json:"noise_type,omitempty"`
}

// GeneratorOptions configures the synthetic Nigerian address generator.
type GeneratorOptions struct {
	Count     int     `json:"count"`
	Seed      uint64  `json:"seed,omitempty"`
	NoiseRate float64 `json:"noise_rate,omitempty"` // 0.0 to 1.0 (default 0.3)
}

// EvalResult contains accuracy and performance scores from evaluating an agent on synthetic data.
type EvalResult struct {
	TotalSamples     int     `json:"total_samples"`
	StateAccuracy    float64 `json:"state_accuracy"`
	PostcodeAccuracy float64 `json:"postcode_accuracy"`
	ValidRate        float64 `json:"valid_rate"`
	AvgFormatScore   float64 `json:"avg_format_score"`
	DurationMs       int64   `json:"duration_ms"`
}

var (
	sampleStreets = []string{
		"Adetokunbo Ademola Street",
		"Ahmadu Bello Way",
		"Awolowo Road",
		"Broad Street",
		"Nnamdi Azikiwe Avenue",
		"Herbert Macaulay Way",
		"Yakubu Gowon Crescent",
		"Ozumba Mbadiwe Avenue",
		"Bourdillon Road",
		"Ikorodu Road",
		"Airport Road",
		"Kofo Abayomi Street",
		"Constitution Avenue",
		"Ring Road",
	}

	sampleLandmarks = []string{
		"Opposite Central Mosque",
		"Beside First Bank",
		"Near Shoprite Mall",
		"Behind Total Filling Station",
		"Close to General Hospital",
		"After the Police Barracks",
		"Adjacent Zenith Bank Branch",
		"Under the Flyover Bridge",
		"Near Market Square",
	}

	stateLocalities = map[string][]string{
		"LA": {"Victoria Island", "Ikeja", "Lekki Phase 1", "Yaba", "Surulere", "Ikoyi", "Agege", "Festac Town"},
		"FC": {"Wuse II", "Garki", "Maitama", "Asokoro", "Gwarinpa", "Jabi", "Utako", "Kubwa"},
		"RI": {"Port Harcourt GRA", "Trans-Amadi", "D/Line", "Diobu", "Rumuokoro", "Woji"},
		"KN": {"Sabon Gari", "Nassarawa", "Fagge", "Tarauni", "Dala", "Kano City"},
		"OY": {"Bodija", "Dugbe", "Agodi GRA", "Mokola", "Ring Road", "Iwo Road"},
		"KD": {"Kaduna North", "Barnawa", "Kakuri", "Sabon Tasha", "Tudun Wada"},
		"EN": {"Independence Layout", "New Haven", "Ogui", "Achara Layout", "GRA"},
		"AN": {"Awka GRA", "Onitsha Main Market", "Nnewi Industrial Area"},
		"ED": {"GRA Benin", "Ugbowo", "Ikpoba Hill", "Airport Road"},
		"OG": {"Abeokuta GRA", "Ijebu Ode", "Sagamu", "Ota Industrial Zone"},
	}

	legacyPostcodes = map[string]string{
		"LA": "100001",
		"FC": "900001",
		"RI": "500001",
		"KN": "700001",
		"OY": "200001",
		"KD": "800001",
		"EN": "400001",
		"AN": "420001",
		"ED": "300001",
		"OG": "110001",
	}

	stateCodeList = []string{"LA", "FC", "RI", "KN", "OY", "KD", "EN", "AN", "ED", "OG"}
)

// GenerateSyntheticDataset produces a slice of realistic synthetic Nigerian addresses.
func GenerateSyntheticDataset(opts GeneratorOptions) []SyntheticAddress {
	if opts.Count <= 0 {
		opts.Count = 50
	}
	if opts.NoiseRate < 0 {
		opts.NoiseRate = 0
	} else if opts.NoiseRate > 1.0 {
		opts.NoiseRate = 1.0
	}

	var rng *rand.Rand
	if opts.Seed != 0 {
		rng = rand.New(rand.NewPCG(opts.Seed, opts.Seed^0x5DEECE66D))
	} else {
		rng = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
	}

	dataset := make([]SyntheticAddress, opts.Count)
	for i := 0; i < opts.Count; i++ {
		dataset[i] = generateOneAddress(rng, opts.NoiseRate)
	}
	return dataset
}

func generateOneAddress(rng *rand.Rand, noiseRate float64) SyntheticAddress {
	stateCode := stateCodeList[rng.IntN(len(stateCodeList))]
	stateInfo := NigerianStates[stateCode]

	localities := stateLocalities[stateCode]
	locality := localities[rng.IntN(len(localities))]
	street := sampleStreets[rng.IntN(len(sampleStreets))]
	houseNum := rng.IntN(120) + 1

	lgaNum := rng.IntN(15) + 1
	districtCode := fmt.Sprintf("A%02d", rng.IntN(10)+1)
	areaCode := "AA"
	unitNum := rng.IntN(50) + 1
	postcodeCode := fmt.Sprintf("%s %02d %s %s %02d", stateCode, lgaNum, districtCode, areaCode, unitNum)

	isNoisy := rng.Float64() < noiseRate
	noiseType := "clean"
	var rawParts []string

	prefix := fmt.Sprintf("No. %d, %s", houseNum, street)
	rawParts = append(rawParts, prefix)

	if isNoisy {
		roll := rng.IntN(4)
		switch roll {
		case 0:
			// Landmark insertion
			noiseType = "noisy_landmark"
			lm := sampleLandmarks[rng.IntN(len(sampleLandmarks))]
			rawParts = append(rawParts, lm)
			rawParts = append(rawParts, locality)
			rawParts = append(rawParts, stateInfo.Name+" State")
			rawParts = append(rawParts, postcodeCode)

		case 1:
			// Typo injection into state name
			noiseType = "typo"
			noisyState := stateInfo.Name + "s"
			if stateCode == "FC" {
				noisyState = "Abuja FCT"
			}
			rawParts = append(rawParts, locality)
			rawParts = append(rawParts, noisyState)
			rawParts = append(rawParts, postcodeCode)

		case 2:
			// Legacy 6-digit postcode alongside or instead
			noiseType = "old_format"
			rawParts = append(rawParts, locality)
			rawParts = append(rawParts, stateInfo.Name)
			legacy := legacyPostcodes[stateCode]
			if legacy == "" {
				legacy = "100001"
			}
			rawParts = append(rawParts, legacy)

		default:
			// Unstructured / no commas
			noiseType = "unstructured"
			rawParts = append(rawParts, locality)
			rawParts = append(rawParts, stateInfo.Name)
			rawParts = append(rawParts, postcodeCode)
		}
	} else {
		rawParts = append(rawParts, locality)
		rawParts = append(rawParts, stateInfo.Name+" State")
		rawParts = append(rawParts, postcodeCode)
	}

	var rawText string
	if noiseType == "unstructured" {
		rawText = strings.Join(rawParts, " ")
	} else {
		rawText = strings.Join(rawParts, ", ")
	}

	return SyntheticAddress{
		RawText:           rawText,
		ExpectedState:     stateInfo.Name,
		ExpectedStateCode: stateCode,
		ExpectedLGA:       fmt.Sprintf("%02d", lgaNum),
		ExpectedPostcode:  postcodeCode,
		ExpectedLat:       stateInfo.Latitude,
		ExpectedLng:       stateInfo.Longitude,
		NoiseType:         noiseType,
	}
}

type stateMatcherEntry struct {
	name  string
	upper string
}

var (
	postcodeRegex  = regexp.MustCompile(`\b([A-Za-z]{2})[\s\-]?([0-9]{2})[\s\-]?([A-Za-z0-9]{3})[\s\-]?([A-Za-z]{2})[\s\-]?([0-9]{2})\b`)
	legacyRegex    = regexp.MustCompile(`\b([0-9]{6})\b`)
	stateCodeRegex = regexp.MustCompile(`\b(AB|AD|AK|AN|BA|BY|BN|BO|CR|DE|EB|ED|EK|EN|FC|GO|IM|JI|KD|KN|KT|KE|KO|KW|LA|NA|NI|OG|ON|OS|OY|PL|RI|SO|TA|YO|ZA)\b`)

	sortedStateEntries = func() []stateMatcherEntry {
		entries := make([]stateMatcherEntry, 0, len(NigerianStates))
		for _, info := range NigerianStates {
			entries = append(entries, stateMatcherEntry{
				name:  info.Name,
				upper: strings.ToUpper(info.Name),
			})
		}
		slices.SortFunc(entries, func(a, b stateMatcherEntry) int {
			return len(b.upper) - len(a.upper) // longest names first
		})
		return entries
	}()
)

// BaselineExtract provides an offline heuristic extractor for Nigerian addresses.
// It searches raw unstructured text for state names/codes and postcode patterns.
func BaselineExtract(raw string) (extractedState string, extractedPostcode string) {
	upper := strings.ToUpper(raw)

	// 1. Detect postcode
	if match := postcodeRegex.FindStringSubmatch(raw); len(match) == 6 {
		extractedPostcode = fmt.Sprintf("%s %s %s %s %s",
			strings.ToUpper(match[1]),
			match[2],
			strings.ToUpper(match[3]),
			strings.ToUpper(match[4]),
			match[5],
		)
	} else if match := legacyRegex.FindStringSubmatch(raw); len(match) == 2 {
		extractedPostcode = match[1]
	}

	// 2. Detect state (longest name match first)
	for _, entry := range sortedStateEntries {
		if strings.Contains(upper, entry.upper) {
			extractedState = entry.name
			break
		}
	}
	// Fallback to 2-letter state code
	if extractedState == "" {
		if match := stateCodeRegex.FindString(upper); match != "" {
			if info, ok := NigerianStates[match]; ok {
				extractedState = info.Name
			}
		}
	}

	return extractedState, extractedPostcode
}

// EvaluateAgent runs an extraction function against a synthetic dataset and tallies accuracy metrics.
func EvaluateAgent(dataset []SyntheticAddress, extractor func(raw string) (string, string)) EvalResult {
	if len(dataset) == 0 {
		return EvalResult{}
	}

	start := time.Now()
	correctState := 0
	correctPostcode := 0
	validPostcodeCount := 0
	totalScore := 0.0

	for _, item := range dataset {
		gotState, gotPostcode := extractor(item.RawText)

		if strings.EqualFold(gotState, item.ExpectedState) || strings.EqualFold(gotState, item.ExpectedStateCode) {
			correctState++
		}

		if gotPostcode != "" {
			if _, err := Parse(gotPostcode); err == nil {
				validPostcodeCount++
			}
			diag := Diagnose(gotPostcode)
			totalScore += diag.FormatScore

			// Clean compare
			cleanGot := strings.ReplaceAll(strings.ReplaceAll(gotPostcode, " ", ""), "-", "")
			cleanExp := strings.ReplaceAll(strings.ReplaceAll(item.ExpectedPostcode, " ", ""), "-", "")
			if strings.EqualFold(cleanGot, cleanExp) {
				correctPostcode++
			}
		}
	}

	n := float64(len(dataset))
	duration := time.Since(start)

	return EvalResult{
		TotalSamples:     len(dataset),
		StateAccuracy:    (float64(correctState) / n) * 100.0,
		PostcodeAccuracy: (float64(correctPostcode) / n) * 100.0,
		ValidRate:        (float64(validPostcodeCount) / n) * 100.0,
		AvgFormatScore:   totalScore / n,
		DurationMs:       duration.Milliseconds(),
	}
}
