package postcode

import (
	"regexp"
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

var (
	// Matches Nigerian mobile & landline formats: +234 803 123 4567, 2348031234567, 0803-123-4567, 09012345678
	ngPhoneRegex = regexp.MustCompile(`(?:\+?234[\s\-]?|0)(?:70|80|81|90|91)[0-9][\s\-]?[0-9]{3}[\s\-]?[0-9]{4}\b`)
	// Matches standard email addresses
	emailRegex = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)
	// Matches 11-digit numeric NIN / BVN identifiers
	ninBVNRegex = regexp.MustCompile(`\b[0-9]{11}\b`)
	// Matches embedded 5-segment NIPOST postcode candidates (including FCT alias or unpadded digits)
	postcodeCandidateRegex = regexp.MustCompile(`(?i)\b[A-Z]{2,3}[\-\s][0-9OI]{1,2}[\-\s][A-Z][0-9OI]{2}[\-\s][A-Z]{2}[\-\s][0-9OI]{1,2}\b`)
)

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
	if matches := emailRegex.FindAllString(out, -1); len(matches) > 0 {
		count += len(matches)
		addType("email")
		out = emailRegex.ReplaceAllString(out, "[REDACTED_EMAIL]")
	}

	// 2. Redact Nigerian phone numbers
	if matches := ngPhoneRegex.FindAllString(out, -1); len(matches) > 0 {
		count += len(matches)
		addType("phone")
		out = ngPhoneRegex.ReplaceAllString(out, "[REDACTED_PHONE]")
	}

	// 3. Redact 11-digit NIN/BVN numbers
	if matches := ninBVNRegex.FindAllString(out, -1); len(matches) > 0 {
		count += len(matches)
		addType("nin_bvn")
		out = ninBVNRegex.ReplaceAllString(out, "[REDACTED_NIN_BVN]")
	}

	// 4. Mask exact building unit (UU) in any embedded 11-character NIPOST postcode
	var maskedPC, areaPrefix, stateCode, lgaCode string
	var maskedList []string
	if candidates := postcodeCandidateRegex.FindAllString(out, -1); len(candidates) > 0 {
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
