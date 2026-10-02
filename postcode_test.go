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
