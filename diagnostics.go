package postcode

import (
	"fmt"
	"sort"
	"strings"
)

// SegmentDiagnosis describes an issue identified in a specific segment of a postcode.
type SegmentDiagnosis struct {
	Segment     string   `json:"segment"`               // "Length", "LegacyFormat", "State", "LGA", "District", "Area", "BuildingUnit"
	Input       string   `json:"input"`                 // The actual segment characters received
	Expected    string   `json:"expected"`              // What the NIPOST grammar requires
	Message     string   `json:"message"`               // Human & agent readable explanation
	Suggestions []string `json:"suggestions,omitempty"` // Actionable candidates for agent self-correction
}

// DiagnosticReport provides deep structural and administrative analysis of a postcode string.
// Designed for humans and autonomous AI agents to self-correct invalid inputs.
type DiagnosticReport struct {
	Input         string             `json:"input"`
	Valid         bool               `json:"valid"`
	IsTemporary   bool               `json:"is_temporary,omitempty"`  // True if input is a temporary NIPOST building code (TP-...)
	Normalized    string             `json:"normalized,omitempty"`    // Hyphenated canonical form if valid or recoverable
	SuggestedFix  string             `json:"suggested_fix,omitempty"` // Auto-repaired canonical postcode when deterministically recoverable
	CleanLength   int                `json:"clean_length"`            // Number of alphanumeric characters found
	Diagnoses     []SegmentDiagnosis `json:"diagnoses,omitempty"`
	ActionableTip string             `json:"actionable_tip,omitempty"` // Summary recommendation for the caller/agent
	FormatScore   float64            `json:"format_score"`             // Structural conformance score (0.0 to 100.0)
}

// ValidationError represents a validation failure enriched with an actionable DiagnosticReport.
type ValidationError struct {
	Report DiagnosticReport
}

func (e *ValidationError) Error() string {
	if len(e.Report.Diagnoses) > 0 {
		return fmt.Sprintf("postcode: invalid %q: %s", e.Report.Input, e.Report.Diagnoses[0].Message)
	}
	return fmt.Sprintf("postcode: invalid postcode %q", e.Report.Input)
}

// legacyPostalPrefixes maps legacy 3-digit NIPOST postal prefixes (first 3 digits of 6-digit codes)
// to their corresponding modern 11-character State-LGA prefix and human label.
var legacyPostalPrefixes = map[string]struct {
	Prefix string
	Label  string
}{
	"100": {Prefix: "LA-11", Label: "Lagos (Ikeja / Mainland)"},
	"101": {Prefix: "LA-14", Label: "Lagos (Lagos Island / Eti-Osa)"},
	"102": {Prefix: "LA-03", Label: "Lagos (Alimosho / Amuwo-Odofin)"},
	"103": {Prefix: "LA-06", Label: "Lagos (Badagry / Ojo)"},
	"104": {Prefix: "LA-12", Label: "Lagos (Ikorodu)"},
	"105": {Prefix: "LA-09", Label: "Lagos (Ibeju-Lekki)"},
	"106": {Prefix: "LA-07", Label: "Lagos (Epe)"},
	"110": {Prefix: "OG-14", Label: "Ogun (Abeokuta)"},
	"200": {Prefix: "OY-01", Label: "Oyo (Ibadan)"},
	"230": {Prefix: "OS-01", Label: "Osun (Osogbo)"},
	"240": {Prefix: "KW-01", Label: "Kwara (Ilorin)"},
	"300": {Prefix: "ED-01", Label: "Edo (Benin City)"},
	"310": {Prefix: "DE-02", Label: "Delta (Warri)"},
	"320": {Prefix: "DE-01", Label: "Delta (Asaba)"},
	"340": {Prefix: "ON-01", Label: "Ondo (Akure)"},
	"360": {Prefix: "EK-01", Label: "Ekiti (Ado Ekiti)"},
	"400": {Prefix: "EN-05", Label: "Enugu (Enugu)"},
	"420": {Prefix: "AN-01", Label: "Anambra (Awka)"},
	"430": {Prefix: "AN-02", Label: "Anambra (Onitsha)"},
	"440": {Prefix: "AB-15", Label: "Abia (Umuahia)"},
	"450": {Prefix: "AB-01", Label: "Abia (Aba)"},
	"460": {Prefix: "IM-01", Label: "Imo (Owerri)"},
	"480": {Prefix: "EB-13", Label: "Ebonyi (Abakaliki)"},
	"500": {Prefix: "RI-01", Label: "Rivers (Port Harcourt)"},
	"520": {Prefix: "AK-11", Label: "Akwa Ibom (Uyo)"},
	"540": {Prefix: "CR-01", Label: "Cross River (Calabar)"},
	"560": {Prefix: "BY-01", Label: "Bayelsa (Yenagoa)"},
	"600": {Prefix: "BO-01", Label: "Borno (Maiduguri)"},
	"620": {Prefix: "YO-01", Label: "Yobe (Damaturu)"},
	"640": {Prefix: "AD-01", Label: "Adamawa (Yola)"},
	"660": {Prefix: "TA-01", Label: "Taraba (Jalingo)"},
	"700": {Prefix: "KN-31", Label: "Kano (Kano Municipal)"},
	"720": {Prefix: "JI-24", Label: "Jigawa (Dutse)"},
	"740": {Prefix: "BA-02", Label: "Bauchi (Bauchi)"},
	"760": {Prefix: "GO-01", Label: "Gombe (Gombe)"},
	"800": {Prefix: "KD-01", Label: "Kaduna (Kaduna)"},
	"820": {Prefix: "KT-01", Label: "Katsina (Katsina)"},
	"840": {Prefix: "SO-01", Label: "Sokoto (Sokoto)"},
	"860": {Prefix: "KE-01", Label: "Kebbi (Birnin Kebbi)"},
	"880": {Prefix: "ZA-01", Label: "Zamfara (Gusau)"},
	"900": {Prefix: "FC-03", Label: "FCT Abuja (Abuja Municipal)"},
	"901": {Prefix: "FC-02", Label: "FCT Abuja (Bwari / Kubwa)"},
	"910": {Prefix: "NI-09", Label: "Niger (Minna)"},
	"930": {Prefix: "PL-01", Label: "Plateau (Jos)"},
	"960": {Prefix: "KO-01", Label: "Kogi (Lokoja)"},
	"970": {Prefix: "BN-01", Label: "Benue (Makurdi)"},
	"980": {Prefix: "NA-01", Label: "Nasarawa (Lafia)"},
}

// LookupLegacyPrefix maps a 6-digit legacy Nigerian postal code (e.g. "100001", "900001")
// to its corresponding modern NIPOST State-LGA prefix (e.g. "LA-11", "FC-03") and description.
func LookupLegacyPrefix(sixDigit string) (prefix string, label string, ok bool) {
	s := strings.TrimSpace(sixDigit)
	if len(s) != 6 || !isAllDigitsStr(s) {
		return "", "", false
	}
	if m, found := legacyPostalPrefixes[s[:3]]; found {
		return m.Prefix, m.Label, true
	}
	// Fallback to 2-digit regional zone prefix (e.g. "10" -> LA, "90" -> FC)
	for k, m := range legacyPostalPrefixes {
		if k[:2] == s[:2]+"0"[:0] && strings.HasPrefix(k, s[:2]) {
			return m.Prefix, m.Label, true
		}
	}
	return "", "", false
}

// Diagnose inspects a candidate postcode string, performing exhaustive grammar,
// structural, and administrative checks. Returns an actionable DiagnosticReport
// that AI agents can use to understand the error and formulate corrections.
func Diagnose(raw string) DiagnosticReport {
	report := DiagnosticReport{
		Input: strings.TrimSpace(raw),
	}

	// 1. Clean and count alphanumeric characters
	var clean []byte
	for i := 0; i < len(report.Input); i++ {
		b := report.Input[i]
		if b == ' ' || b == '-' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		clean = append(clean, b)
	}
	report.CleanLength = len(clean)

	// Special Case A: Temporary NIPOST building code (e.g. "TP-...", "TP_...")
	if IsTemporary(report.Input) {
		report.IsTemporary = true
		report.Valid = false
		formatted := FormatTemporary(report.Input)
		report.Normalized = formatted
		report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
			Segment:     "TemporaryCode",
			Input:       report.Input,
			Expected:    "Official 11-character NIPOST postcode (SS-LL-DDD-AA-UU)",
			Message:     fmt.Sprintf("Temporary NIPOST building code %q detected; minted for unsurveyed property awaiting cadastral indexing.", report.Input),
			Suggestions: []string{"Resolve via NIPOST gateway: GET /v1/lookup?code=" + formatted},
		})
		report.ActionableTip = fmt.Sprintf("Temporary code %q detected (pending survey). Query live gateway (postcode lookup %s or GET /v1/lookup) to check if an official 11-character postcode has been assigned.", report.Input, formatted)
		report.FormatScore = 50.0
		return report
	}

	// Special Case B: Legacy 6-digit numeric NIPOST postal code (e.g. "100001", "900001")
	if len(clean) == 6 && isAllDigits(clean) {
		cleanStr := string(clean)
		var suggestions []string
		msg := fmt.Sprintf("Deprecated 6-digit numeric NIPOST code %q detected; replaced by 11-character alphanumeric grammar (SS-LL-DDD-AA-UU).", cleanStr)
		if prefix, label, ok := LookupLegacyPrefix(cleanStr); ok {
			msg = fmt.Sprintf("Deprecated 6-digit numeric NIPOST code %q detected for %s; replaced by 11-character alphanumeric grammar starting with %s.", cleanStr, label, prefix)
			suggestions = append(suggestions, fmt.Sprintf("%s (modern prefix for %s)", prefix, label))
		}
		report.Diagnoses = append(report.Diagnoses,
			SegmentDiagnosis{
				Segment:     "LegacyFormat",
				Input:       cleanStr,
				Expected:    "11-character alphanumeric NIPOST postcode (e.g. 'LA-11-W06-TC-10')",
				Message:     msg,
				Suggestions: suggestions,
			},
			SegmentDiagnosis{
				Segment:  "Length",
				Input:    cleanStr,
				Expected: "Exactly 11 alphanumeric characters (e.g. 'EK 01 A03 FK 01')",
				Message:  "Input has 6 characters (legacy numeric format); missing 5 characters required for 11-character digital postcode.",
			},
		)
		report.Valid = false
		if len(suggestions) > 0 {
			report.ActionableTip = fmt.Sprintf("Legacy 6-digit postal code %q is deprecated. Use search/reverse geocoding or start with modern prefix %s.", cleanStr, suggestions[0])
		} else {
			report.ActionableTip = fmt.Sprintf("Legacy 6-digit postal code %q is deprecated. Upgrade to the 11-character NIPOST format (SS-LL-DDD-AA-UU).", cleanStr)
		}
		report.FormatScore = 25.0
		return report
	}

	// Special Case C: 5 delimited tokens where LGA or BuildingUnit is an unpadded single digit (e.g. "EK-1-A03-FK-1")
	if tokens := splitDelimitedTokens(report.Input); len(tokens) == 5 && (len(tokens[1]) == 1 || len(tokens[4]) == 1) &&
		len(tokens[0]) == 2 && len(tokens[2]) == 3 && len(tokens[3]) == 2 {
		if len(tokens[1]) == 1 && isDigit(tokens[1][0]) {
			padded := "0" + tokens[1]
			if tokens[1] == "0" {
				padded = "01"
			}
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "LGA",
				Input:       tokens[1],
				Expected:    "2 zero-padded digits in range 01-99",
				Message:     fmt.Sprintf("LGA segment %q is a single digit; must be zero-padded to 2 digits (%q)", tokens[1], padded),
				Suggestions: []string{padded},
			})
			tokens[1] = padded
		}
		if len(tokens[4]) == 1 && isDigit(tokens[4][0]) {
			padded := "0" + tokens[4]
			if tokens[4] == "0" {
				padded = "01"
			}
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "BuildingUnit",
				Input:       tokens[4],
				Expected:    "2 zero-padded digits in range 01-99",
				Message:     fmt.Sprintf("Building unit segment %q is a single digit; must be zero-padded to 2 digits (%q)", tokens[4], padded),
				Suggestions: []string{padded},
			})
			tokens[4] = padded
		}
		// Now let clean be the reconstructed 11-char candidate so remaining segments are checked accurately
		reconstructed := strings.Join(tokens, "")
		if len(reconstructed) == 11 {
			clean = []byte(reconstructed)
		}
	}

	// Length checks
	if report.CleanLength != 11 {
		var missingDesc string
		switch {
		case report.CleanLength == 0:
			missingDesc = "Input is completely empty."
		case report.CleanLength < 2:
			missingDesc = "Input has fewer than 2 characters (need 2-letter State, 2-digit LGA, 3-char District, 2-letter Area, 2-digit Unit)."
		case report.CleanLength < 4:
			missingDesc = "Input has State segment only; missing LGA, District, Area, and Building Unit segments."
		case report.CleanLength < 7:
			missingDesc = "Input has State and LGA; missing District, Area, and Building Unit segments."
		case report.CleanLength < 9:
			missingDesc = "Input is missing Area and Building Unit segments."
		case report.CleanLength < 11:
			missingDesc = fmt.Sprintf("Input has %d characters; missing final %d character(s) for the Building Unit segment.", report.CleanLength, 11-report.CleanLength)
		default:
			missingDesc = fmt.Sprintf("Input has %d characters (exceeds required 11). Check for extraneous digits or delimiters.", report.CleanLength)
		}

		report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
			Segment:  "Length",
			Input:    string(clean),
			Expected: "Exactly 11 alphanumeric characters (e.g. 'EK 01 A03 FK 01')",
			Message:  missingDesc,
		})
	}

	// 2. Segment-by-segment grammar checks (only for characters present)
	// Segment 1: State (positions 0..1)
	if len(clean) >= 2 {
		stateStr := string(clean[0:2])
		if !isAlpha(clean[0]) || !isAlpha(clean[1]) {
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "State",
				Input:       stateStr,
				Expected:    "2 uppercase letters (A-Z) representing a Nigerian State or FCT",
				Message:     fmt.Sprintf("State segment must be 2 letters, but got %q", stateStr),
				Suggestions: suggestValidStates(stateStr),
			})
		} else {
			// Check if state code is recognized in official 36 States + FCT registry
			if stateRec, ok := NigerianStates[stateStr]; !ok {
				report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
					Segment:     "State",
					Input:       stateStr,
					Expected:    "Valid 2-letter Nigerian state code (one of 36 States or FC)",
					Message:     fmt.Sprintf("State code %q is not a recognized Nigerian state code", stateStr),
					Suggestions: suggestValidStates(stateStr),
				})
			} else {
				_ = stateRec // valid
			}
		}
	}

	// Segment 2: LGA (positions 2..3)
	if len(clean) >= 4 {
		lgaStr := string(clean[2:4])
		if !isDigit(clean[2]) || !isDigit(clean[3]) {
			var sug []string
			if fixed, ok := fixNumericLookalikes(lgaStr); ok {
				sug = []string{fixed}
			}
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "LGA",
				Input:       lgaStr,
				Expected:    "2 digits in range 01-99",
				Message:     fmt.Sprintf("LGA segment must be 2 numeric digits, but got %q", lgaStr),
				Suggestions: sug,
			})
		} else if clean[2] == '0' && clean[3] == '0' {
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "LGA",
				Input:       lgaStr,
				Expected:    "2 digits in range 01-99 (00 is prohibited per NIPOST standard)",
				Message:     "LGA code '00' is invalid; valid LGA codes range from 01 to 99",
				Suggestions: []string{"01", "02"},
			})
		}
	}

	// Segment 3: District (positions 4..6)
	if len(clean) >= 7 {
		distStr := string(clean[4:7])
		if !isAlnum(clean[4]) || !isAlnum(clean[5]) || !isAlnum(clean[6]) {
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:  "District",
				Input:    distStr,
				Expected: "3 alphanumeric characters (e.g. 'A03', 'W06')",
				Message:  fmt.Sprintf("District segment contains non-alphanumeric characters: %q", distStr),
			})
		}
	}

	// Segment 4: Area (positions 7..8)
	if len(clean) >= 9 {
		areaStr := string(clean[7:9])
		if !isAlpha(clean[7]) || !isAlpha(clean[8]) {
			var sug []string
			if fixed, ok := fixAlphaLookalikes(areaStr); ok {
				sug = []string{fixed}
			}
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "Area",
				Input:       areaStr,
				Expected:    "2 uppercase letters (A-Z, e.g. 'FK', 'TC')",
				Message:     fmt.Sprintf("Area segment must be 2 letters, but got %q", areaStr),
				Suggestions: sug,
			})
		}
	}

	// Segment 5: Building Unit (positions 9..10)
	if len(clean) >= 11 {
		unitStr := string(clean[9:11])
		if !isDigit(clean[9]) || !isDigit(clean[10]) {
			var sug []string
			if fixed, ok := fixNumericLookalikes(unitStr); ok {
				sug = []string{fixed}
			}
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "BuildingUnit",
				Input:       unitStr,
				Expected:    "2 digits in range 01-99",
				Message:     fmt.Sprintf("Building unit segment must be 2 numeric digits, but got %q", unitStr),
				Suggestions: sug,
			})
		} else if clean[9] == '0' && clean[10] == '0' {
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:     "BuildingUnit",
				Input:       unitStr,
				Expected:    "2 digits in range 01-99 (00 is prohibited per NIPOST standard)",
				Message:     "Building unit code '00' is invalid; valid unit codes range from 01 to 99",
				Suggestions: []string{"01", "02"},
			})
		}
	}

	// 3. Overall evaluation
	if len(report.Diagnoses) == 0 && report.CleanLength == 11 {
		p, err := Parse(raw)
		if err == nil {
			report.Valid = true
			report.Normalized = p.Formatted()
			report.ActionableTip = "Postcode is fully valid and conforms to NIPOST standard."
			report.FormatScore = 100.0
			return report
		}
	}

	// Attempt deterministic auto-repair to populate SuggestedFix
	if repaired, ok := TryRepair(raw); ok {
		report.SuggestedFix = repaired.Formatted()
	}

	// Build actionable tip for agent
	report.Valid = false
	var tips []string
	for _, d := range report.Diagnoses {
		if len(d.Suggestions) > 0 {
			tips = append(tips, fmt.Sprintf("Fix %s: consider %s", d.Segment, strings.Join(d.Suggestions, ", ")))
		} else {
			tips = append(tips, fmt.Sprintf("Fix %s: %s", d.Segment, d.Expected))
		}
	}
	if report.SuggestedFix != "" {
		tips = append(tips, fmt.Sprintf("Auto-repair candidate: %s", report.SuggestedFix))
	}
	if len(tips) > 0 {
		report.ActionableTip = strings.Join(tips, "; ")
	} else {
		report.ActionableTip = "Check standard NIPOST format: State(2A) LGA(2D) District(3AN) Area(2A) Unit(2D), e.g. 'EK-01-A03-FK-01'."
	}

	// Calculate format score (0.0 to 100.0) based on severity of issues
	deduction := float64(len(report.Diagnoses)) * 15.0
	diff := float64(11 - report.CleanLength)
	if diff < 0 {
		diff = -diff
	}
	deduction += diff * 8.0
	score := 100.0 - deduction
	if score < 0.0 {
		score = 0.0
	}
	report.FormatScore = score

	return report
}

// TryRepair attempts to deterministically self-heal a malformed or noisy Nigerian postcode string.
// It handles:
//   - Single-digit unpadded LGA or BuildingUnit segments in delimited strings (e.g. "EK-1-A03-FK-1" -> "EK-01-A03-FK-01")
//   - Zero-floor violations ("00" -> "01" in LGA or BuildingUnit)
//   - Common state code aliases ("BE" -> "BN" for Benue)
//   - OCR lookalikes in digit/alpha segments ('O' <-> '0', 'I'/'L' <-> '1', 'S' <-> '5', 'B' <-> '8')
//   - Surrounding noise or punctuation (e.g. "(LA-11-W06-TC-10)")
//
// Returns the repaired Postcode and true only if the resulting candidate passes full Validate().
func TryRepair(raw string) (Postcode, bool) {
	if Validate(raw) == nil {
		p, err := Parse(raw)
		if err == nil {
			// Check if District[1:3] has an OCR 'O' instead of '0' (e.g. "EK-01-AO3-FK-01")
			dist := p.District()
			if len(dist) == 3 && isAlpha(dist[0]) && (dist[1] == 'O' || dist[2] == 'O') {
				if fixedDigits, ok := fixNumericLookalikes(dist[1:3]); ok {
					if pFixed, errFixed := Parse(p.State() + p.LGA() + dist[:1] + fixedDigits + p.Area() + p.BuildingUnit()); errFixed == nil {
						return pFixed, true
					}
				}
			}
			return p, true
		}
	}

	s := strings.TrimSpace(raw)
	if s == "" {
		return Postcode{}, false
	}

	// 1. Check 5-part delimited input (handles single-digit LGA/Unit and 3-letter state aliases like "FCT"/"ABJ")
	if tokens := splitDelimitedTokens(s); len(tokens) == 5 {
		st, lga, dist, area, unit := tokens[0], tokens[1], tokens[2], tokens[3], tokens[4]
		if alias, ok := stateAliases[st]; ok {
			st = alias
		} else if len(st) == 2 {
			if fixed, ok := fixAlphaLookalikes(st); ok {
				if alias, ok2 := stateAliases[fixed]; ok2 {
					st = alias
				} else {
					st = fixed
				}
			}
		}
		if len(lga) == 1 {
			if fixed, ok := fixNumericLookalikes("0" + lga); ok {
				lga = fixed
			}
		} else if len(lga) == 2 {
			if fixed, ok := fixNumericLookalikes(lga); ok {
				lga = fixed
			}
		}
		if lga == "00" {
			lga = "01"
		}
		if len(dist) == 3 {
			if d0, ok := fixAlphaLookalikes(dist[:1]); ok {
				dist = d0 + dist[1:]
			}
			if d12, ok := fixNumericLookalikes(dist[1:3]); ok {
				dist = dist[:1] + d12
			}
		}
		if len(area) == 2 {
			if fixed, ok := fixAlphaLookalikes(area); ok {
				area = fixed
			}
		}
		if len(unit) == 1 {
			if fixed, ok := fixNumericLookalikes("0" + unit); ok {
				unit = fixed
			}
		} else if len(unit) == 2 {
			if fixed, ok := fixNumericLookalikes(unit); ok {
				unit = fixed
			}
		}
		if unit == "00" {
			unit = "01"
		}

		candidate := st + lga + dist + area + unit
		if Validate(candidate) == nil {
			if p, err := Parse(candidate); err == nil {
				return p, true
			}
		}
	}

	// 2. Strip non-alphanumeric characters and attempt 11-char positional repair
	var alnum []byte
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		if isAlnum(b) {
			alnum = append(alnum, b)
		}
	}

	if len(alnum) == 11 {
		if p, ok := repair11Bytes(alnum); ok {
			return p, true
		}
	}

	return Postcode{}, false
}

// Repair attempts to deterministically self-heal a postcode string, returning a ValidationError if unrecoverable.
func Repair(raw string) (Postcode, error) {
	if p, ok := TryRepair(raw); ok {
		return p, nil
	}
	return Postcode{}, &ValidationError{Report: Diagnose(raw)}
}

func repair11Bytes(buf []byte) (Postcode, bool) {
	b := make([]byte, 11)
	copy(b, buf)

	// State (0..1)
	if st, ok := fixAlphaLookalikes(string(b[0:2])); ok {
		b[0], b[1] = st[0], st[1]
	}
	if alias, ok := stateAliases[string(b[0:2])]; ok && len(alias) == 2 {
		b[0], b[1] = alias[0], alias[1]
	}

	// LGA (2..3)
	if lga, ok := fixNumericLookalikes(string(b[2:4])); ok {
		b[2], b[3] = lga[0], lga[1]
	}
	if b[2] == '0' && b[3] == '0' {
		b[3] = '1'
	}

	// District (4..6): standard NIPOST district is 1 letter + 2 digits (e.g. A03, W06)
	if d0, ok := fixAlphaLookalikes(string(b[4:5])); ok {
		b[4] = d0[0]
	}
	if d12, ok := fixNumericLookalikes(string(b[5:7])); ok {
		b[5], b[6] = d12[0], d12[1]
	}

	// Area (7..8)
	if area, ok := fixAlphaLookalikes(string(b[7:9])); ok {
		b[7], b[8] = area[0], area[1]
	}

	// BuildingUnit (9..10)
	if unit, ok := fixNumericLookalikes(string(b[9:11])); ok {
		b[9], b[10] = unit[0], unit[1]
	}
	if b[9] == '0' && b[10] == '0' {
		b[10] = '1'
	}

	candidate := string(b)
	if Validate(candidate) == nil {
		if p, err := Parse(candidate); err == nil {
			return p, true
		}
	}
	return Postcode{}, false
}

func splitDelimitedTokens(s string) []string {
	var tokens []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == ' ' || b == '-' || b == '\t' || b == '/' || b == '_' || b == '.' {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			continue
		}
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		cur.WriteByte(b)
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

func fixNumericLookalikes(s string) (string, bool) {
	b := []byte(s)
	for i := range b {
		switch b[i] {
		case 'O', 'Q', 'D':
			b[i] = '0'
		case 'I', 'L':
			b[i] = '1'
		case 'Z':
			b[i] = '2'
		case 'S':
			b[i] = '5'
		case 'G':
			b[i] = '6'
		case 'B':
			b[i] = '8'
		}
		if !isDigit(b[i]) {
			return "", false
		}
	}
	if len(b) == 2 && b[0] == '0' && b[1] == '0' {
		b[1] = '1'
	}
	return string(b), true
}

func fixAlphaLookalikes(s string) (string, bool) {
	b := []byte(s)
	for i := range b {
		switch b[i] {
		case '0':
			b[i] = 'O'
		case '1':
			b[i] = 'I'
		case '2':
			b[i] = 'Z'
		case '5':
			b[i] = 'S'
		case '8':
			b[i] = 'B'
		}
		if !isAlpha(b[i]) {
			return "", false
		}
	}
	return string(b), true
}

func isAllDigits(b []byte) bool {
	for _, ch := range b {
		if !isDigit(ch) {
			return false
		}
	}
	return len(b) > 0
}

func isAllDigitsStr(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return false
		}
	}
	return len(s) > 0
}

// suggestValidStates finds closest Nigerian state codes based on character overlap and similarity.
func suggestValidStates(input string) []string {
	up := strings.ToUpper(input)

	// Check known state aliases first (e.g. "BE" -> "BN (Benue)")
	if aliasCode, ok := stateAliases[up]; ok {
		if rec, ok2 := NigerianStates[aliasCode]; ok2 {
			return []string{fmt.Sprintf("%s (%s)", rec.Code, rec.Name)}
		}
	}

	type match struct {
		code  string
		name  string
		score int
	}

	var matches []match
	for code, rec := range NigerianStates {
		score := 0
		if len(up) == 2 {
			if up[0] == code[0] {
				score += 2
			}
			if up[1] == code[1] {
				score += 2
			}
			if up[0] == code[1] || up[1] == code[0] {
				score += 1
			}
		}
		// Also match state name prefix if input resembles name (e.g. "LA" -> Lagos)
		if strings.HasPrefix(strings.ToUpper(rec.Name), up) {
			score += 3
		}

		if score > 0 {
			matches = append(matches, match{code: code, name: rec.Name, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].code < matches[j].code
	})

	maxResults := 3
	if len(matches) < maxResults {
		maxResults = len(matches)
	}

	var suggestions []string
	for i := 0; i < maxResults; i++ {
		suggestions = append(suggestions, fmt.Sprintf("%s (%s)", matches[i].code, matches[i].name))
	}

	if len(suggestions) == 0 {
		// Provide common states as fallback
		return []string{"LA (Lagos)", "FC (Federal Capital Territory)", "KN (Kano)"}
	}

	return suggestions
}
