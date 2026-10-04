package postcode

import (
	"fmt"
	"sort"
	"strings"
)

// SegmentDiagnosis describes an issue identified in a specific segment of a postcode.
type SegmentDiagnosis struct {
	Segment     string   `json:"segment"`               // "Length", "State", "LGA", "District", "Area", "BuildingUnit"
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
	Normalized    string             `json:"normalized,omitempty"`     // Hyphenated canonical form if valid or recoverable
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
		if b == ' ' || b == '-' {
			continue
		}
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		clean = append(clean, b)
	}
	report.CleanLength = len(clean)

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
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:  "LGA",
				Input:    lgaStr,
				Expected: "2 digits in range 01-99",
				Message:  fmt.Sprintf("LGA segment must be 2 numeric digits, but got %q", lgaStr),
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
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:  "Area",
				Input:    areaStr,
				Expected: "2 uppercase letters (A-Z, e.g. 'FK', 'TC')",
				Message:  fmt.Sprintf("Area segment must be 2 letters, but got %q", areaStr),
			})
		}
	}

	// Segment 5: Building Unit (positions 9..10)
	if len(clean) >= 11 {
		unitStr := string(clean[9:11])
		if !isDigit(clean[9]) || !isDigit(clean[10]) {
			report.Diagnoses = append(report.Diagnoses, SegmentDiagnosis{
				Segment:  "BuildingUnit",
				Input:    unitStr,
				Expected: "2 digits in range 01-99",
				Message:  fmt.Sprintf("Building unit segment must be 2 numeric digits, but got %q", unitStr),
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

// suggestValidStates finds closest Nigerian state codes based on character overlap and similarity.
func suggestValidStates(input string) []string {
	up := strings.ToUpper(input)

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
