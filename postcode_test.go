package postcode

import (
	"errors"
	"testing"
)

func TestParse_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		input        string
		wantState    string
		wantLGA      string
		wantDistrict string
		wantArea     string
		wantUnit     string
		wantRaw      string
		wantFmt      string
		wantDisplay  string
	}{
		{
			name:         "spaced uppercase",
			input:        "EK 01 A03 FK 01",
			wantState:    "EK",
			wantLGA:      "01",
			wantDistrict: "A03",
			wantArea:     "FK",
			wantUnit:     "01",
			wantRaw:      "EK01A03FK01",
			wantFmt:      "EK-01-A03-FK-01",
			wantDisplay:  "EK 01 A03 FK 01",
		},
		{
			name:         "hyphenated uppercase",
			input:        "EK-01-A03-FK-01",
			wantState:    "EK",
			wantLGA:      "01",
			wantDistrict: "A03",
			wantArea:     "FK",
			wantUnit:     "01",
			wantRaw:      "EK01A03FK01",
			wantFmt:      "EK-01-A03-FK-01",
			wantDisplay:  "EK 01 A03 FK 01",
		},
		{
			name:         "compact lowercase",
			input:        "ek01a03fk01",
			wantState:    "EK",
			wantLGA:      "01",
			wantDistrict: "A03",
			wantArea:     "FK",
			wantUnit:     "01",
			wantRaw:      "EK01A03FK01",
			wantFmt:      "EK-01-A03-FK-01",
			wantDisplay:  "EK 01 A03 FK 01",
		},
		{
			name:         "Abuja FCT code",
			input:        "FC-02-A09-DB-09",
			wantState:    "FC",
			wantLGA:      "02",
			wantDistrict: "A09",
			wantArea:     "DB",
			wantUnit:     "09",
			wantRaw:      "FC02A09DB09",
			wantFmt:      "FC-02-A09-DB-09",
			wantDisplay:  "FC 02 A09 DB 09",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			if got := p.State(); got != tt.wantState {
				t.Errorf("State() = %q, want %q", got, tt.wantState)
			}
			if got := p.LGA(); got != tt.wantLGA {
				t.Errorf("LGA() = %q, want %q", got, tt.wantLGA)
			}
			if got := p.District(); got != tt.wantDistrict {
				t.Errorf("District() = %q, want %q", got, tt.wantDistrict)
			}
			if got := p.Area(); got != tt.wantArea {
				t.Errorf("Area() = %q, want %q", got, tt.wantArea)
			}
			if got := p.BuildingUnit(); got != tt.wantUnit {
				t.Errorf("BuildingUnit() = %q, want %q", got, tt.wantUnit)
			}
			if got := p.Raw(); got != tt.wantRaw {
				t.Errorf("Raw() = %q, want %q", got, tt.wantRaw)
			}
			if got := p.Formatted(); got != tt.wantFmt {
				t.Errorf("Formatted() = %q, want %q", got, tt.wantFmt)
			}
			if got := p.String(); got != tt.wantDisplay {
				t.Errorf("String() = %q, want %q", got, tt.wantDisplay)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty input", "", ErrInvalidLength},
		{"too short", "EK 01 A03", ErrInvalidLength},
		{"too long", "EK 01 A03 FK 01 99", ErrInvalidLength},
		{"invalid state letters", "12 01 A03 FK 01", ErrInvalidFormat},
		{"invalid LGA non-digits", "EK AB A03 FK 01", ErrInvalidFormat},
		{"invalid LGA zero (00)", "EK 00 A03 FK 01", ErrInvalidFormat},
		{"invalid Unit zero (00)", "EK 01 A03 FK 00", ErrInvalidFormat},
		{"invalid area non-alpha", "EK 01 A03 12 01", ErrInvalidFormat},
		{"invalid unit non-digits", "EK 01 A03 FK AB", ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(tt.input)
			if err == nil {
				t.Fatalf("Parse(%q) expected error, got nil", tt.input)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Parse(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestPostcode_ZeroValue(t *testing.T) {
	t.Parallel()
	var p Postcode
	if !p.IsZero() {
		t.Errorf("p.IsZero() = false, want true")
	}
	if got := p.State(); got != "" {
		t.Errorf("p.State() = %q, want empty string", got)
	}
	if got := p.LGA(); got != "" {
		t.Errorf("p.LGA() = %q, want empty string", got)
	}
	if got := p.District(); got != "" {
		t.Errorf("p.District() = %q, want empty string", got)
	}
	if got := p.Area(); got != "" {
		t.Errorf("p.Area() = %q, want empty string", got)
	}
	if got := p.BuildingUnit(); got != "" {
		t.Errorf("p.BuildingUnit() = %q, want empty string", got)
	}
	if got := p.Raw(); got != "" {
		t.Errorf("p.Raw() = %q, want empty string", got)
	}
	if got := p.Formatted(); got != "" {
		t.Errorf("p.Formatted() = %q, want empty string", got)
	}
	if got := p.String(); got != "" {
		t.Errorf("p.String() = %q, want empty string", got)
	}
}

func TestTextMarshaling(t *testing.T) {
	t.Parallel()
	const raw = "EK 01 A03 FK 01"
	p, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	marshaled, err := p.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText failed: %v", err)
	}
	if string(marshaled) != raw {
		t.Errorf("MarshalText = %q, want %q", string(marshaled), raw)
	}

	var p2 Postcode
	if err := p2.UnmarshalText([]byte("ek-01-a03-fk-01")); err != nil {
		t.Fatalf("UnmarshalText failed: %v", err)
	}
	if p2.Raw() != "EK01A03FK01" {
		t.Errorf("UnmarshalText parsed = %q, want EK01A03FK01", p2.Raw())
	}
}

func TestParseSeq(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"EK 01 A03 FK 01",
		"invalid",
		"FC 02 A09 DB 09",
	}

	var parsed []Postcode
	var errCount int

	for p, err := range ParseSeq(inputs) {
		if err != nil {
			errCount++
		} else {
			parsed = append(parsed, p)
		}
	}

	if errCount != 1 {
		t.Errorf("expected 1 error, got %d", errCount)
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 parsed postcodes, got %d", len(parsed))
	}
}

func BenchmarkParse(b *testing.B) {
	const raw = "EK 01 A03 FK 01"
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, _ = Parse(raw)
	}
}

func BenchmarkFormatted(b *testing.B) {
	p, _ := Parse("EK 01 A03 FK 01")
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = p.Formatted()
	}
}

func BenchmarkAppendFormatted(b *testing.B) {
	p, _ := Parse("EK 01 A03 FK 01")
	buf := make([]byte, 0, 15)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		buf = p.AppendFormatted(buf[:0])
	}
}

func TestPostcode_NewCoreMethods(t *testing.T) {
	t.Parallel()

	p1 := MustParse("\tLA-11-W06-TC-10\n")
	p2 := MustParse("la 11 w06 tc 10")
	p3 := MustParse("EK-01-A03-FK-01")

	if !p1.Equal(p2) {
		t.Errorf("expected p1.Equal(p2) to be true")
	}
	if p1.Equal(p3) {
		t.Errorf("expected p1.Equal(p3) to be false")
	}
	if p3.Compare(p1) >= 0 || p1.Compare(p3) <= 0 || p1.Compare(p2) != 0 {
		t.Errorf("unexpected Compare ordering between %s and %s", p3, p1)
	}
	if got := p1.Slug(); got != "la-11-w06-tc-10" {
		t.Errorf("p1.Slug() = %q, want 'la-11-w06-tc-10'", got)
	}
	buf := p1.AppendFormatted(make([]byte, 0, 15))
	if string(buf) != "LA-11-W06-TC-10" {
		t.Errorf("AppendFormatted = %q, want 'LA-11-W06-TC-10'", string(buf))
	}

	// Distance calculations
	dist := p1.DistanceTo(p3)
	if dist < 200_000 || dist > 300_000 {
		t.Errorf("expected distance between Lagos (Ikeja) and Ado Ekiti to be ~235km, got %.1fm", dist)
	}
	distFn, err := DistanceMeters("LA-11-W06-TC-10", "EK-01-A03-FK-01")
	if err != nil || distFn != dist {
		t.Errorf("DistanceMeters() = %.1f (err: %v), want %.1f", distFn, err, dist)
	}

	// PointGeometry
	pg := p1.Location().PointGeometry()
	if pg.Type != "Point" || len(pg.Coordinates) != 2 || pg.Coordinates[0] != 3.3515 || pg.Coordinates[1] != 6.6018 {
		t.Errorf("unexpected PointGeometry: %+v", pg)
	}

	// SQL driver.Valuer & sql.Scanner
	val, err := p1.Value()
	if err != nil || val != "LA-11-W06-TC-10" {
		t.Errorf("p1.Value() = %v (err: %v), want 'LA-11-W06-TC-10'", val, err)
	}
	var scanned Postcode
	if err := scanned.Scan("la 11 w06 tc 10"); err != nil || !scanned.Equal(p1) {
		t.Errorf("Scan(string) failed: got %v, err %v", scanned, err)
	}
	if err := scanned.Scan([]byte("EK-01-A03-FK-01")); err != nil || !scanned.Equal(p3) {
		t.Errorf("Scan([]byte) failed: got %v, err %v", scanned, err)
	}
	if err := scanned.Scan(nil); err != nil || !scanned.IsZero() {
		t.Errorf("Scan(nil) failed: got %v, err %v", scanned, err)
	}

	// LookupState & LookupLGA
	if st, ok := LookupState("Abuja"); !ok || st.Code != "FC" {
		t.Errorf("LookupState('Abuja') = %+v, ok=%v; want FC", st, ok)
	}
	if st, ok := LookupState("BE"); !ok || st.Code != "BN" {
		t.Errorf("LookupState('BE') = %+v, ok=%v; want BN (Benue)", st, ok)
	}
	if lga, ok := LookupLGA("Lagos", "Eti Osa"); !ok || lga.LGACode != "08" {
		t.Errorf("LookupLGA('Lagos', 'Eti Osa') = %+v, ok=%v; want 08", lga, ok)
	}
	if lga, ok := LookupLGA("FC", "AMAC"); !ok || lga.LGACode != "03" {
		t.Errorf("LookupLGA('FC', 'AMAC') = %+v, ok=%v; want 03", lga, ok)
	}

	// Every one of the 37 states must have at least 1 LGA in StateLGAs
	for code := range NigerianStates {
		if lgas := StateLGAs(code); len(lgas) == 0 {
			t.Errorf("expected non-empty StateLGAs(%q), got 0", code)
		}
	}
}

func TestNDPAMaskingAndRedaction(t *testing.T) {
	p := MustParse("EK-01-A03-FK-01")
	if got := p.Masked(); got != "EK-01-A03-FK-**" {
		t.Errorf("p.Masked() = %q, want 'EK-01-A03-FK-**'", got)
	}
	if got := p.AreaPrefix(); got != "EK-01-A03-FK" {
		t.Errorf("p.AreaPrefix() = %q, want 'EK-01-A03-FK'", got)
	}

	raw := "Deliver to EK-01-A03-FK-01 beside Shoprite, call 08031234567 or email buyer@example.ng (BVN 22345678901)"
	red := RedactPII(raw)
	if len(red.MaskedPostcodes) != 1 || red.MaskedPostcodes[0] != "EK-01-A03-FK-**" {
		t.Errorf("unexpected MaskedPostcodes: %+v", red.MaskedPostcodes)
	}
	if red.PIIRedactedCount < 4 {
		t.Errorf("expected at least 4 PII redactions (unit, phone, email, nin/bvn), got %d (%+v)", red.PIIRedactedCount, red)
	}
}

func TestTemporaryCodes(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"TP-01-A01-01", true},
		{"TP_01_A01_01", true},
		{"tp-123456", true},
		{"tp_998877", true},
		{"TP 01 A01 01", true},
		{"TP01A0101", true},
		{"EK-01-A03-FK-01", false},
		{"100001", false},
		{"LA-11-W06-TC-10", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := IsTemporary(tc.input); got != tc.want {
			t.Errorf("IsTemporary(%q) = %v, want %v", tc.input, got, tc.want)
		}
		if got := IsTemporaryCode(tc.input); got != tc.want {
			t.Errorf("IsTemporaryCode(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}

	if got := FormatTemporary("tp_01_a01_01"); got != "TP-01-A01-01" {
		t.Errorf("FormatTemporary('tp_01_a01_01') = %q, want 'TP-01-A01-01'", got)
	}
	if got := FormatTemporary("tp-123456"); got != "TP-123456" {
		t.Errorf("FormatTemporary('tp-123456') = %q, want 'TP-123456'", got)
	}
	if got := FormatTemporary("tp 01 a01 01"); got != "TP-01-A01-01" {
		t.Errorf("FormatTemporary('tp 01 a01 01') = %q, want 'TP-01-A01-01'", got)
	}

	errVal := Validate("TP-01-A01-01")
	if !errors.Is(errVal, ErrTemporaryCode) {
		t.Errorf("Validate('TP-01-A01-01') error = %v, want errors.Is(..., ErrTemporaryCode)", errVal)
	}

	_, errParse := Parse("TP-01-A01-AA-01")
	if !errors.Is(errParse, ErrTemporaryCode) {
		t.Errorf("Parse('TP-01-A01-AA-01') error = %v, want errors.Is(..., ErrTemporaryCode)", errParse)
	}
}

