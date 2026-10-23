package postcode

import (
	"strings"
)

// Masked returns the NDPA data-minimized representation of the postcode (e.g. "EK-01-A03-FK-**"),
// preserving State, LGA, District, and Area for routing and analytics while masking the
// exact building unit ("UU") to protect residential privacy under the Nigeria Data Protection Act (NDPA).
func (p Postcode) Masked() string {
	if p.IsZero() {
		return ""
	}
	var b [15]byte
	copy(b[0:2], p.raw[0:2])
	b[2] = '-'
	copy(b[3:5], p.raw[2:4])
	b[5] = '-'
	copy(b[6:9], p.raw[4:7])
	b[9] = '-'
	copy(b[10:12], p.raw[7:9])
	b[12] = '-'
	b[13] = '*'
	b[14] = '*'
	return string(b[:])
}

// AreaPrefix returns the 4-segment Area-level prefix (e.g. "EK-01-A03-FK") without the building unit.
func (p Postcode) AreaPrefix() string {
	if p.IsZero() {
		return ""
	}
	var b [12]byte
	copy(b[0:2], p.raw[0:2])
	b[2] = '-'
	copy(b[3:5], p.raw[2:4])
	b[5] = '-'
	copy(b[6:9], p.raw[4:7])
	b[9] = '-'
	copy(b[10:12], p.raw[7:9])
	return string(b[:])
}

// RedactedAddress holds the result of NDPA-compliant PII redaction on an unstructured Nigerian address or log line.
type RedactedAddress struct {
	RedactedText     string   `json:"redacted_text"`
	MaskedPostcode   string   `json:"masked_postcode,omitempty"`
	MaskedPostcodes  []string `json:"masked_postcodes,omitempty"`
	AreaPrefix       string   `json:"area_prefix,omitempty"`
	StateCode        string   `json:"state_code,omitempty"`
	LGACode          string   `json:"lga_code,omitempty"`
	PIIRedactedCount int      `json:"pii_redacted_count"`
	RedactedTypes    []string `json:"redacted_types,omitempty"`
	NDPACompliant    bool     `json:"ndpa_compliant"`
}

// RedactPII scrubs personal identifiable information (Nigerian phone numbers, email addresses,
// 11-digit BVN/NIN numbers, and exact building unit numbers in NIPOST postcodes) from raw text
// while retaining State, LGA, District, and Area routing context in line with the Nigeria Data Protection Act (NDPA).
func RedactPII(raw string) RedactedAddress {
	out := raw
	var types []string
	count := 0

	addType := func(t string) {
		for _, existing := range types {
			if existing == t {
				return
			}
		}
		types = append(types, t)
	}

	// 1. Redact emails
	var emailCount int
	out, emailCount = redactEmails(out)
	if emailCount > 0 {
		count += emailCount
		addType("email")
	}

	// 2. Redact Nigerian phone numbers
	var phoneCount int
	out, phoneCount = redactNigerianPhones(out)
	if phoneCount > 0 {
		count += phoneCount
		addType("phone")
	}

	// 3. Redact 11-digit NIN/BVN numbers
	var ninCount int
	out, ninCount = redactElevenDigitNumbers(out)
	if ninCount > 0 {
		count += ninCount
		addType("nin_bvn")
	}

	// 4. Mask exact building unit (UU) in any embedded 11-character NIPOST postcode
	var maskedPC, areaPrefix, stateCode, lgaCode string
	var maskedList []string
	if candidates := findPostcodeCandidates(out); len(candidates) > 0 {
		for _, cand := range candidates {
			if p, ok := TryRepair(cand); ok {
				masked := p.Masked()
				maskedList = append(maskedList, masked)
				if maskedPC == "" {
					maskedPC = masked
					areaPrefix = p.AreaPrefix()
					stateCode = p.State()
					lgaCode = p.LGA()
				}
				out = strings.ReplaceAll(out, cand, masked)
				count++
				addType("building_unit")
			}
		}
	}

	// If no full postcode was present, still extract State/LGA routing context via ExtractAddress
	if stateCode == "" {
		ext := ExtractAddress(out)
		stateCode = ext.StateCode
		lgaCode = ext.LGACode
	}

	return RedactedAddress{
		RedactedText:     out,
		MaskedPostcode:   maskedPC,
		MaskedPostcodes:  maskedList,
		AreaPrefix:       areaPrefix,
		StateCode:        stateCode,
		LGACode:          lgaCode,
		PIIRedactedCount: count,
		RedactedTypes:    types,
		NDPACompliant:    true,
	}
}

func isEmailLocalByte(b byte) bool {
	return isASCIIAlnum(b) || b == '.' || b == '_' || b == '%' || b == '+' || b == '-'
}

func isEmailDomainByte(b byte) bool {
	return isASCIIAlnum(b) || b == '.' || b == '-'
}

func redactEmails(s string) (string, int) {
	count := 0
	var sb strings.Builder
	i := 0
	n := len(s)
	for i < n {
		at := strings.IndexByte(s[i:], '@')
		if at < 0 {
			sb.WriteString(s[i:])
			break
		}
		atIdx := i + at
		l := atIdx
		for l > i && isEmailLocalByte(s[l-1]) {
			l--
		}
		r := atIdx + 1
		dotIdx := -1
		for r < n && isEmailDomainByte(s[r]) {
			if s[r] == '.' {
				dotIdx = r
			}
			r++
		}
		// Trim trailing dots/hyphens on domain
		for r > atIdx+1 && (s[r-1] == '.' || s[r-1] == '-') {
			if r-1 == dotIdx {
				dotIdx = -1
			}
			r--
		}
		tldLen := r - (dotIdx + 1)
		if l < atIdx && dotIdx > atIdx+1 && tldLen >= 2 {
			sb.WriteString(s[i:l])
			sb.WriteString("[REDACTED_EMAIL]")
			count++
			i = r
		} else {
			sb.WriteString(s[i : atIdx+1])
			i = atIdx + 1
		}
	}
	if count == 0 {
		return s, 0
	}
	return sb.String(), count
}

func redactNigerianPhones(s string) (string, int) {
	count := 0
	var sb strings.Builder
	i := 0
	n := len(s)
	for i < n {
		if end, ok := matchNigerianPhoneAt(s, i); ok {
			sb.WriteString("[REDACTED_PHONE]")
			count++
			i = end
			continue
		}
		sb.WriteByte(s[i])
		i++
	}
	if count == 0 {
		return s, 0
	}
	return sb.String(), count
}

// Matches Nigerian mobile & landline formats: +234 803 123 4567, 2348031234567, 0803-123-4567, 09012345678
func matchNigerianPhoneAt(s string, start int) (int, bool) {
	n := len(s)
	if start > 0 && isWordChar(s[start-1]) {
		return 0, false
	}
	p := start
	if p < n && s[p] == '+' {
		p++
		if p+3 > n || s[p:p+3] != "234" {
			return 0, false
		}
		p += 3
		if p < n && (s[p] == ' ' || s[p] == '-') {
			p++
		}
	} else if p+3 <= n && s[p:p+3] == "234" {
		p += 3
		if p < n && (s[p] == ' ' || s[p] == '-') {
			p++
		}
	} else if p < n && s[p] == '0' {
		p++
	} else {
		return 0, false
	}

	if p+3 > n {
		return 0, false
	}
	pfx := s[p : p+2]
	if pfx != "70" && pfx != "80" && pfx != "81" && pfx != "90" && pfx != "91" {
		return 0, false
	}
	if !isASCIIDigit(s[p+2]) {
		return 0, false
	}
	p += 3
	if p < n && (s[p] == ' ' || s[p] == '-') {
		p++
	}
	if p+3 > n || !isASCIIDigit(s[p]) || !isASCIIDigit(s[p+1]) || !isASCIIDigit(s[p+2]) {
		return 0, false
	}
	p += 3
	if p < n && (s[p] == ' ' || s[p] == '-') {
		p++
	}
	if p+4 > n || !isASCIIDigit(s[p]) || !isASCIIDigit(s[p+1]) || !isASCIIDigit(s[p+2]) || !isASCIIDigit(s[p+3]) {
		return 0, false
	}
	p += 4
	if p < n && isWordChar(s[p]) {
		return 0, false
	}
	return p, true
}

func redactElevenDigitNumbers(s string) (string, int) {
	count := 0
	var sb strings.Builder
	i := 0
	n := len(s)
	for i < n {
		if i+11 <= n && (i == 0 || !isWordChar(s[i-1])) && (i+11 == n || !isWordChar(s[i+11])) {
			allDigits := true
			for j := 0; j < 11; j++ {
				if !isASCIIDigit(s[i+j]) {
					allDigits = false
					break
				}
			}
			if allDigits {
				sb.WriteString("[REDACTED_NIN_BVN]")
				count++
				i += 11
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	if count == 0 {
		return s, 0
	}
	return sb.String(), count
}

func isDigitOrOI(b byte) bool {
	return isASCIIDigit(b) || b == 'O' || b == 'o' || b == 'I' || b == 'i'
}

// findPostcodeCandidates matches embedded 5-segment NIPOST postcode candidates:
// (?i)\b[A-Z]{2,3}[\-\s][0-9OI]{1,2}[\-\s][A-Z][0-9OI]{2}[\-\s][A-Z]{2}[\-\s][0-9OI]{1,2}\b
func findPostcodeCandidates(s string) []string {
	var out []string
	n := len(s)
	for i := 0; i < n; i++ {
		if i > 0 && isWordChar(s[i-1]) {
			continue
		}
		p := i
		letters := 0
		for p < n && isASCIILetter(s[p]) && letters < 4 {
			p++
			letters++
		}
		if letters < 2 || letters > 3 || p >= n || (s[p] != '-' && s[p] != ' ') {
			continue
		}
		p++
		lgaLen := 0
		for p < n && isDigitOrOI(s[p]) && lgaLen < 3 {
			p++
			lgaLen++
		}
		if lgaLen < 1 || lgaLen > 2 || p >= n || (s[p] != '-' && s[p] != ' ') {
			continue
		}
		p++
		if p+3 > n || !isASCIILetter(s[p]) || !isDigitOrOI(s[p+1]) || !isDigitOrOI(s[p+2]) {
			continue
		}
		p += 3
		if p >= n || (s[p] != '-' && s[p] != ' ') {
			continue
		}
		p++
		if p+2 > n || !isASCIILetter(s[p]) || !isASCIILetter(s[p+1]) {
			continue
		}
		p += 2
		if p >= n || (s[p] != '-' && s[p] != ' ') {
			continue
		}
		p++
		unitLen := 0
		for p < n && isDigitOrOI(s[p]) && unitLen < 3 {
			p++
			unitLen++
		}
		if unitLen < 1 || unitLen > 2 {
			continue
		}
		if p < n && isWordChar(s[p]) {
			continue
		}
		out = append(out, s[i:p])
		i = p - 1
	}
	return out
}
