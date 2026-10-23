package postcode

import "strings"

// AddressExtraction holds the structured entities extracted from an unstructured Nigerian address
// using the 5-stage NIPOST address disambiguation pipeline.
type AddressExtraction struct {
	Raw             string  `json:"raw"`
	Postcode        string  `json:"postcode,omitempty"`         // Canonical hyphenated 11-char postcode (if found/repaired)
	PostcodeValid   bool    `json:"postcode_valid"`             // True if Postcode is a valid 11-char NIPOST code
	LegacyPostcode  string  `json:"legacy_postcode,omitempty"`  // Deprecated 6-digit numeric postal code (if found)
	SuggestedPrefix string  `json:"suggested_prefix,omitempty"` // Resolved State or State-LGA prefix (e.g. "LA-08", "FC-03")
	StateCode       string  `json:"state_code,omitempty"`       // 2-letter NIPOST state code (e.g. "LA", "FC")
	StateName       string  `json:"state_name,omitempty"`       // Full state name (e.g. "Lagos", "Federal Capital Territory")
	LGACode         string  `json:"lga_code,omitempty"`         // 2-digit LGA code (e.g. "08", "11")
	LGAName         string  `json:"lga_name,omitempty"`         // Full LGA name (e.g. "Eti-Osa", "Ikeja")
	Locality        string  `json:"locality,omitempty"`         // Neighborhood, district, or estate (e.g. "Lekki Phase 1", "Wuse II")
	Street          string  `json:"street,omitempty"`           // Street or house number component
	Landmark        string  `json:"landmark,omitempty"`         // Landmark anchor (e.g. "Beside First Bank", "Opposite Central Mosque")
	Latitude        float64 `json:"latitude,omitempty"`         // Best-available resolved latitude
	Longitude       float64 `json:"longitude,omitempty"`        // Best-available resolved longitude
	Confidence      string  `json:"confidence"`                 // "high", "medium", or "low"
}

type localityRule struct {
	phrase    string
	upper     string
	stateCode string
	lgaCode   string
}

var (
	// nigerianLocalityRules maps well-known Nigerian neighborhoods, estates, and districts
	// to their canonical StateCode and LGACode. Ordered longest-first at init time.
	nigerianLocalityRules = []localityRule{
		{phrase: "Onitsha Main Market", stateCode: "AN", lgaCode: "02"},
		{phrase: "Nnewi Industrial Area", stateCode: "AN", lgaCode: "03"},
		{phrase: "Ota Industrial Zone", stateCode: "OG", lgaCode: "02"},
		{phrase: "Independence Layout", stateCode: "EN", lgaCode: "05"},
		{phrase: "Port Harcourt GRA", stateCode: "RI", lgaCode: "01"},
		{phrase: "Victoria Island", stateCode: "LA", lgaCode: "08"},
		{phrase: "Lekki Phase 1", stateCode: "LA", lgaCode: "08"},
		{phrase: "Lekki Phase 2", stateCode: "LA", lgaCode: "08"},
		{phrase: "Achara Layout", stateCode: "EN", lgaCode: "02"},
		{phrase: "Abeokuta GRA", stateCode: "OG", lgaCode: "14"},
		{phrase: "Festac Town", stateCode: "LA", lgaCode: "04"},
		{phrase: "Trans-Amadi", stateCode: "RI", lgaCode: "01"},
		{phrase: "Sabon Tasha", stateCode: "KD", lgaCode: "03"},
		{phrase: "Ikpoba Hill", stateCode: "ED", lgaCode: "03"},
		{phrase: "Sabon Gari", stateCode: "KN", lgaCode: "01"},
		{phrase: "Tudun Wada", stateCode: "KD", lgaCode: "02"},
		{phrase: "GRA Benin", stateCode: "ED", lgaCode: "01"},
		{phrase: "Agodi GRA", stateCode: "OY", lgaCode: "01"},
		{phrase: "New Haven", stateCode: "EN", lgaCode: "05"},
		{phrase: "Rumuokoro", stateCode: "RI", lgaCode: "02"},
		{phrase: "Kano City", stateCode: "KN", lgaCode: "31"},
		{phrase: "Ijebu Ode", stateCode: "OG", lgaCode: "01"},
		{phrase: "Gwarinpa", stateCode: "FC", lgaCode: "03"},
		{phrase: "Surulere", stateCode: "LA", lgaCode: "20"},
		{phrase: "Alimosho", stateCode: "LA", lgaCode: "03"},
		{phrase: "Awka GRA", stateCode: "AN", lgaCode: "01"},
		{phrase: "Iwo Road", stateCode: "OY", lgaCode: "02"},
		{phrase: "Barnawa", stateCode: "KD", lgaCode: "02"},
		{phrase: "Maitama", stateCode: "FC", lgaCode: "03"},
		{phrase: "Asokoro", stateCode: "FC", lgaCode: "03"},
		{phrase: "Wuse II", stateCode: "FC", lgaCode: "03"},
		{phrase: "Wuse 2", stateCode: "FC", lgaCode: "03"},
		{phrase: "Kakuri", stateCode: "KD", lgaCode: "02"},
		{phrase: "Bodija", stateCode: "OY", lgaCode: "01"},
		{phrase: "Mokola", stateCode: "OY", lgaCode: "01"},
		{phrase: "Ugbowo", stateCode: "ED", lgaCode: "02"},
		{phrase: "Sagamu", stateCode: "OG", lgaCode: "01"},
		{phrase: "Bompai", stateCode: "KN", lgaCode: "02"},
		{phrase: "D/Line", stateCode: "RI", lgaCode: "01"},
		{phrase: "Festac", stateCode: "LA", lgaCode: "04"},
		{phrase: "Ikoyi", stateCode: "LA", lgaCode: "08"},
		{phrase: "Lekki", stateCode: "LA", lgaCode: "08"},
		{phrase: "Ikeja", stateCode: "LA", lgaCode: "11"},
		{phrase: "Agege", stateCode: "LA", lgaCode: "01"},
		{phrase: "Garki", stateCode: "FC", lgaCode: "03"},
		{phrase: "Kubwa", stateCode: "FC", lgaCode: "02"},
		{phrase: "Utako", stateCode: "FC", lgaCode: "03"},
		{phrase: "Diobu", stateCode: "RI", lgaCode: "01"},
		{phrase: "Dugbe", stateCode: "OY", lgaCode: "05"},
		{phrase: "Fagge", stateCode: "KN", lgaCode: "01"},
		{phrase: "Yaba", stateCode: "LA", lgaCode: "15"},
		{phrase: "Jabi", stateCode: "FC", lgaCode: "03"},
		{phrase: "Wuse", stateCode: "FC", lgaCode: "03"},
		{phrase: "Woji", stateCode: "RI", lgaCode: "02"},
		{phrase: "Ogui", stateCode: "EN", lgaCode: "05"},
		{phrase: "Dala", stateCode: "KN", lgaCode: "31"},
	}

	// landmarkPrefixes identifies Nigerian landmark anchor phrases (checked longest-first).
	landmarkPrefixes = []string{
		"IN FRONT OF",
		"OPPOSITE",
		"ADJACENT",
		"CLOSE TO",
		"BACK OF",
		"BESIDE",
		"BEHIND",
		"AFTER",
		"UNDER",
		"NEAR",
	}

	// streetSuffixes identifies street names and road types.
	streetSuffixes = []string{
		"EXPRESSWAY",
		"BOULEVARD",
		"CRESCENT",
		"AVENUE",
		"STREET",
		"CLOSE",
		"DRIVE",
		"ROAD",
		"LANE",
		"WAY",
		"AVE",
		"STR",
		"ST",
		"RD",
	}

	// englishTwoLetterStopwords prevents 2-letter English prepositions/words in addresses
	// (such as "BY", "ON", "AN", "SO", "GO") from being misclassified as Nigerian state codes.
	englishTwoLetterStopwords = map[string]bool{
		"BY": true,
		"ON": true,
		"AN": true,
		"SO": true,
		"GO": true,
	}
)

func init() {
	for i := range nigerianLocalityRules {
		nigerianLocalityRules[i].upper = strings.ToUpper(nigerianLocalityRules[i].phrase)
	}
}

// ExtractAddress runs the 5-stage Nigerian address disambiguation pipeline on an unstructured
// address string, extracting postcodes (modern or legacy 6-digit), administrative state/LGA,
// neighborhood locality, street, landmark anchors, and resolved coordinates.
func ExtractAddress(raw string) AddressExtraction {
	res := AddressExtraction{
		Raw:        strings.TrimSpace(raw),
		Confidence: "low",
	}
	if res.Raw == "" {
		return res
	}

	upper := strings.ToUpper(res.Raw)

	// Stage 0: Extract modern 11-character postcode (or auto-repairable candidate)
	if match, ok := findPostcodeSubmatch(res.Raw); ok {
		candidate := strings.ToUpper(match[0]) + " " + match[1] + " " + strings.ToUpper(match[2]) + " " + strings.ToUpper(match[3]) + " " + match[4]
		if p, ok := TryRepair(candidate); ok {
			res.Postcode = p.Formatted()
			res.PostcodeValid = true
			res.StateCode = p.State()
			res.LGACode = p.LGA()
			res.SuggestedPrefix = p.State() + "-" + p.LGA()
			if stRec, okSt := NigerianStates[res.StateCode]; okSt {
				res.StateName = stRec.Name
			}
			if lgaRec, okLGA := knownLGAs[res.StateCode+res.LGACode]; okLGA {
				res.LGAName = lgaRec.Name
			}
		} else {
			res.Postcode = strings.ReplaceAll(candidate, " ", "-")
		}
	}

	// Check for legacy 6-digit postal code
	if legacy := findLegacyPostcode(res.Raw); legacy != "" {
		res.LegacyPostcode = legacy
		if prefix, _, ok := LookupLegacyPrefix(res.LegacyPostcode); ok {
			if res.SuggestedPrefix == "" {
				res.SuggestedPrefix = prefix
			}
			parts := strings.Split(prefix, "-")
			if len(parts) == 2 {
				if res.StateCode == "" {
					res.StateCode = parts[0]
					if stRec, okSt := NigerianStates[res.StateCode]; okSt {
						res.StateName = stRec.Name
					}
				}
				if res.LGACode == "" {
					res.LGACode = parts[1]
					if lgaRec, okLGA := knownLGAs[res.StateCode+res.LGACode]; okLGA {
						res.LGAName = lgaRec.Name
					}
				}
			}
		}
	}

	// Stage 1: State Extraction & Normalization (names, capitals, nicknames, and safe codes)
	var textStateCode string
	for _, entry := range sortedStateEntries {
		if strings.Contains(upper, entry.upper) {
			if stRec, ok := LookupState(entry.name); ok {
				textStateCode = stRec.Code
				break
			}
		}
	}
	if textStateCode == "" {
		// Check common aliases & nicknames ("ABUJA", "FCT", "PORT HARCOURT", "PITAKWA", "BENIN CITY")
		switch {
		case strings.Contains(upper, "ABUJA") || strings.Contains(upper, "FCT") || strings.Contains(upper, "F.C.T"):
			textStateCode = "FC"
		case strings.Contains(upper, "PORT HARCOURT") || strings.Contains(upper, "PITAKWA"):
			textStateCode = "RI"
		case strings.Contains(upper, "BENIN CITY"):
			textStateCode = "ED"
		case strings.Contains(upper, "IBADAN"):
			textStateCode = "OY"
		case strings.Contains(upper, "CALABAR"):
			textStateCode = "CR"
		case strings.Contains(upper, "ABEOKUTA"):
			textStateCode = "OG"
		case strings.Contains(upper, "ONITSHA") || strings.Contains(upper, "NNEWI") || strings.Contains(upper, "AWKA"):
			textStateCode = "AN"
		case strings.Contains(upper, "UMUAHIA"):
			textStateCode = "AB"
		case strings.Contains(upper, "MAKURDI"):
			textStateCode = "BN"
		case strings.Contains(upper, "MAIDUGURI"):
			textStateCode = "BO"
		case strings.Contains(upper, "ASABA") || strings.Contains(upper, "WARRI"):
			textStateCode = "DE"
		case strings.Contains(upper, "ABAKALIKI"):
			textStateCode = "EB"
		case strings.Contains(upper, "ADO EKITI") || strings.Contains(upper, "ADO-EKITI"):
			textStateCode = "EK"
		case strings.Contains(upper, "OWERRI"):
			textStateCode = "IM"
		case strings.Contains(upper, "DUTSE"):
			textStateCode = "JI"
		case strings.Contains(upper, "ZARIA"):
			textStateCode = "KD"
		case strings.Contains(upper, "LOKOJA"):
			textStateCode = "KO"
		case strings.Contains(upper, "ILORIN"):
			textStateCode = "KW"
		case strings.Contains(upper, "IKEJA") || strings.Contains(upper, "LEKKI"):
			textStateCode = "LA"
		case strings.Contains(upper, "LAFIA"):
			textStateCode = "NA"
		case strings.Contains(upper, "MINNA") || strings.Contains(upper, "SULEJA"):
			textStateCode = "NI"
		case strings.Contains(upper, "AKURE"):
			textStateCode = "ON"
		case strings.Contains(upper, "OSOGBO") || strings.Contains(upper, "IFE"):
			textStateCode = "OS"
		case strings.Contains(upper, "JOS"):
			textStateCode = "PL"
		case strings.Contains(upper, "JALINGO"):
			textStateCode = "TA"
		case strings.Contains(upper, "DAMATURU"):
			textStateCode = "YO"
		case strings.Contains(upper, "GUSAU"):
			textStateCode = "ZA"
		case strings.Contains(upper, "UYO"):
			textStateCode = "AK"
		case strings.Contains(upper, "YOLA"):
			textStateCode = "AD"
		case strings.Contains(upper, "YENAGOA"):
			textStateCode = "BY"
		}
	}

	// Stage 2 & 3: Locality & LGA Identification
	lgaFromText := false
	for _, rule := range nigerianLocalityRules {
		if strings.Contains(upper, rule.upper) {
			if textStateCode != "" && rule.stateCode != textStateCode {
				continue
			}
			res.Locality = rule.phrase
			if textStateCode == "" {
				textStateCode = rule.stateCode
			}
			if !res.PostcodeValid {
				res.LGACode = rule.lgaCode
				if lgaRec, ok := knownLGAs[rule.stateCode+rule.lgaCode]; ok {
					res.LGAName = lgaRec.Name
				}
				res.SuggestedPrefix = rule.stateCode + "-" + rule.lgaCode
				lgaFromText = true
			}
			break
		}
	}

	// Also check knownLGAs by name if LGA was not already set by an explicit 11-char postcode or locality rule
	if !res.PostcodeValid && !lgaFromText {
		for _, lgaRec := range knownLGAs {
			if textStateCode != "" && lgaRec.StateCode != textStateCode {
				continue
			}
			if strings.Contains(upper, strings.ToUpper(lgaRec.Name)) {
				res.LGACode = lgaRec.LGACode
				res.LGAName = lgaRec.Name
				if textStateCode == "" {
					textStateCode = lgaRec.StateCode
				}
				res.SuggestedPrefix = lgaRec.StateCode + "-" + lgaRec.LGACode
				break
			}
		}
	}

	// Fallback to non-ambiguous 2-letter state code only if still unknown
	if textStateCode == "" && res.StateCode == "" {
		matches := findStateCodeTokens(upper)
		for _, m := range matches {
			if englishTwoLetterStopwords[m] {
				continue
			}
			if _, ok := NigerianStates[m]; ok {
				textStateCode = m
				break
			}
		}
	}

	if res.StateCode == "" && textStateCode != "" {
		res.StateCode = textStateCode
	}
	if res.StateCode != "" && res.StateName == "" {
		if stRec, ok := NigerianStates[res.StateCode]; ok {
			res.StateName = stRec.Name
		}
	}
	if res.StateCode != "" && res.LGACode != "" && res.LGAName == "" {
		if lgaRec, ok := knownLGAs[res.StateCode+res.LGACode]; ok {
			res.LGAName = lgaRec.Name
		}
	}
	if res.SuggestedPrefix == "" && res.StateCode != "" {
		if res.LGACode != "" {
			res.SuggestedPrefix = res.StateCode + "-" + res.LGACode
		} else {
			res.SuggestedPrefix = res.StateCode
		}
	}

	// Stage 4: Street and Landmark Extraction
	if lm := extractLandmarkPhrase(res.Raw, upper); lm != "" {
		res.Landmark = lm
	}
	if st := extractStreetPhrase(res.Raw, upper); st != "" {
		res.Street = st
	}

	// Stage 5: Coordinate & Confidence Resolution
	switch {
	case res.PostcodeValid:
		if loc, err := ResolveLocation(res.Postcode); err == nil {
			res.Latitude = loc.Latitude
			res.Longitude = loc.Longitude
		}
		res.Confidence = "high"
	case res.StateCode != "" && res.LGACode != "":
		if lgaRec, ok := knownLGAs[res.StateCode+res.LGACode]; ok {
			res.Latitude = lgaRec.Latitude
			res.Longitude = lgaRec.Longitude
		} else if stRec, ok := NigerianStates[res.StateCode]; ok {
			res.Latitude = stRec.Latitude
			res.Longitude = stRec.Longitude
		}
		res.Confidence = "medium"
	case res.StateCode != "":
		if stRec, ok := NigerianStates[res.StateCode]; ok {
			res.Latitude = stRec.Latitude
			res.Longitude = stRec.Longitude
		}
		res.Confidence = "medium"
	}

	return res
}

func isASCIILetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isASCIIDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isASCIIAlnum(b byte) bool {
	return isASCIILetter(b) || isASCIIDigit(b)
}

func isWordChar(b byte) bool {
	return isASCIIAlnum(b) || b == '_'
}

func isSepByte(b byte) bool {
	return b == ' ' || b == '-' || b == '\t' || b == '\n' || b == '\r'
}

// findPostcodeSubmatch scans s for the first word-bounded 5-segment Nigerian postcode pattern:
// ([A-Za-z]{2})[\s\-]?([0-9]{2})[\s\-]?([A-Za-z0-9]{3})[\s\-]?([A-Za-z]{2})[\s\-]?([0-9]{2})
func findPostcodeSubmatch(s string) ([5]string, bool) {
	n := len(s)
	for i := 0; i <= n-11; i++ {
		if i > 0 && isWordChar(s[i-1]) {
			continue
		}
		if !isASCIILetter(s[i]) || !isASCIILetter(s[i+1]) {
			continue
		}
		p := i + 2
		if p < n && isSepByte(s[p]) {
			p++
		}
		if p+2 > n || !isASCIIDigit(s[p]) || !isASCIIDigit(s[p+1]) {
			continue
		}
		seg2 := s[p : p+2]
		p += 2
		if p < n && isSepByte(s[p]) {
			p++
		}
		if p+3 > n || !isASCIIAlnum(s[p]) || !isASCIIAlnum(s[p+1]) || !isASCIIAlnum(s[p+2]) {
			continue
		}
		seg3 := s[p : p+3]
		p += 3
		if p < n && isSepByte(s[p]) {
			p++
		}
		if p+2 > n || !isASCIILetter(s[p]) || !isASCIILetter(s[p+1]) {
			continue
		}
		seg4 := s[p : p+2]
		p += 2
		if p < n && isSepByte(s[p]) {
			p++
		}
		if p+2 > n || !isASCIIDigit(s[p]) || !isASCIIDigit(s[p+1]) {
			continue
		}
		seg5 := s[p : p+2]
		p += 2
		if p < n && isWordChar(s[p]) {
			continue
		}
		return [5]string{s[i : i+2], seg2, seg3, seg4, seg5}, true
	}
	return [5]string{}, false
}

// findLegacyPostcode scans s for the first word-bounded 6-digit number (\b[0-9]{6}\b).
func findLegacyPostcode(s string) string {
	n := len(s)
	for i := 0; i <= n-6; i++ {
		if i > 0 && isWordChar(s[i-1]) {
			continue
		}
		allDigits := true
		for j := 0; j < 6; j++ {
			if !isASCIIDigit(s[i+j]) {
				allDigits = false
				break
			}
		}
		if !allDigits {
			continue
		}
		if i+6 < n && isWordChar(s[i+6]) {
			continue
		}
		return s[i : i+6]
	}
	return ""
}

// findStateCodeTokens returns all word-bounded 2-letter uppercase Nigerian state code tokens in upper.
func findStateCodeTokens(upper string) []string {
	var out []string
	n := len(upper)
	for i := 0; i <= n-2; i++ {
		if i > 0 && isWordChar(upper[i-1]) {
			continue
		}
		if upper[i] < 'A' || upper[i] > 'Z' || upper[i+1] < 'A' || upper[i+1] > 'Z' {
			continue
		}
		if i+2 < n && isWordChar(upper[i+2]) {
			continue
		}
		tok := upper[i : i+2]
		if _, ok := NigerianStates[tok]; ok {
			out = append(out, tok)
		}
	}
	return out
}

func extractLandmarkPhrase(raw, upper string) string {
	n := len(upper)
	bestIdx := -1
	for _, kw := range landmarkPrefixes {
		searchFrom := 0
		for searchFrom < n {
			rel := strings.Index(upper[searchFrom:], kw)
			if rel < 0 {
				break
			}
			pos := searchFrom + rel
			endKw := pos + len(kw)
			if (pos == 0 || !isWordChar(upper[pos-1])) && endKw < n && (upper[endKw] == ' ' || upper[endKw] == '\t') {
				if bestIdx == -1 || pos < bestIdx {
					bestIdx = pos
				}
				break
			}
			searchFrom = pos + 1
		}
	}
	if bestIdx < 0 {
		return ""
	}
	end := bestIdx
	for end < len(raw) && raw[end] != ',' && raw[end] != ';' {
		end++
	}
	return strings.TrimSpace(raw[bestIdx:end])
}

func extractStreetPhrase(raw, upper string) string {
	clauses := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})
	offset := 0
	for _, clause := range clauses {
		idx := strings.Index(raw[offset:], clause)
		clauseStart := offset + max(0, idx)
		offset = clauseStart + len(clause)

		trimmed := strings.TrimSpace(clause)
		if len(trimmed) < 4 {
			continue
		}
		upClause := strings.ToUpper(trimmed)
		for _, suf := range streetSuffixes {
			searchFrom := 0
			for searchFrom < len(upClause) {
				rel := strings.Index(upClause[searchFrom:], suf)
				if rel < 0 {
					break
				}
				pos := searchFrom + rel
				endPos := pos + len(suf)
				if pos > 1 && !isWordChar(upClause[pos-1]) && (endPos == len(upClause) || !isWordChar(upClause[endPos])) {
					return strings.TrimSpace(trimmed[:endPos])
				}
				searchFrom = pos + 1
			}
		}
	}
	return ""
}
