package postcode

import (
	"encoding"
	"errors"
	"fmt"
	"iter"
)

var (
	ErrInvalidLength = errors.New("postcode: invalid length, must be exactly 11 alphanumeric characters")
	ErrInvalidFormat = errors.New("postcode: invalid format structure")
)

var (
	_ encoding.TextAppender    = Postcode{}
	_ encoding.TextMarshaler   = Postcode{}
	_ encoding.TextUnmarshaler = (*Postcode)(nil)
)

// Postcode represents an immutable, validated 11-byte Nigerian postcode.
// Backed by a fixed byte array to ensure zero heap escapes.
type Postcode struct {
	raw [11]byte
}

// Parse extracts and validates segments with 0 heap allocations.
func Parse(raw string) (Postcode, error) {
	var p Postcode
	idx := 0

	for i := 0; i < len(raw); i++ {
		b := raw[i]
		if b == ' ' || b == '-' {
			continue
		}
		if idx >= 11 {
			return Postcode{}, ErrInvalidLength
		}

		// Fast branch-free ASCII uppercase conversion
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		p.raw[idx] = b
		idx++
	}

	if idx != 11 {
		return Postcode{}, ErrInvalidLength
	}

	// Grammar: State(2A) LGA(2D) District(3AN) Area(2A) Unit(2D)
	if !isAlpha(p.raw[0]) || !isAlpha(p.raw[1]) {
		return Postcode{}, fmt.Errorf("%w: state must be 2 letters", ErrInvalidFormat)
	}
	if !isDigit(p.raw[2]) || !isDigit(p.raw[3]) {
		return Postcode{}, fmt.Errorf("%w: LGA must be 2 digits", ErrInvalidFormat)
	}
	// Per NIPOST spec: numeric segments range 01-99 (never 00)
	if p.raw[2] == '0' && p.raw[3] == '0' {
		return Postcode{}, fmt.Errorf("%w: LGA cannot be 00 (valid range 01-99)", ErrInvalidFormat)
	}
	if !isAlnum(p.raw[4]) || !isAlnum(p.raw[5]) || !isAlnum(p.raw[6]) {
		return Postcode{}, fmt.Errorf("%w: district must be 3 alphanumeric characters", ErrInvalidFormat)
	}
	if !isAlpha(p.raw[7]) || !isAlpha(p.raw[8]) {
		return Postcode{}, fmt.Errorf("%w: area must be 2 letters", ErrInvalidFormat)
	}
	if !isDigit(p.raw[9]) || !isDigit(p.raw[10]) {
		return Postcode{}, fmt.Errorf("%w: building unit must be 2 digits", ErrInvalidFormat)
	}
	if p.raw[9] == '0' && p.raw[10] == '0' {
		return Postcode{}, fmt.Errorf("%w: building unit cannot be 00 (valid range 01-99)", ErrInvalidFormat)
	}

	return p, nil
}

// Validate checks that raw is a grammatically valid Nigerian postcode
// with a recognized Nigerian state or FCT code.
func Validate(raw string) error {
	p, err := Parse(raw)
	if err != nil {
		return err
	}
	if _, ok := NigerianStates[p.State()]; !ok {
		return fmt.Errorf("%w: unknown state code %q (not in 36 States + FCT)", ErrInvalidFormat, p.State())
	}
	return nil
}

// IsValid reports whether raw is a valid Nigerian postcode with a recognized state code.
func IsValid(raw string) bool {
	return Validate(raw) == nil
}

// Safe segment accessors.
func (p Postcode) State() string {
	if p.IsZero() {
		return ""
	}
	return string(p.raw[0:2])
}

func (p Postcode) LGA() string {
	if p.IsZero() {
		return ""
	}
	return string(p.raw[2:4])
}

func (p Postcode) District() string {
	if p.IsZero() {
		return ""
	}
	return string(p.raw[4:7])
}

func (p Postcode) Area() string {
	if p.IsZero() {
		return ""
	}
	return string(p.raw[7:9])
}

func (p Postcode) BuildingUnit() string {
	if p.IsZero() {
		return ""
	}
	return string(p.raw[9:11])
}

// Segments returns the 5 constituent administrative segments of the postcode.
func (p Postcode) Segments() Segments {
	if p.IsZero() {
		return Segments{}
	}
	return Segments{
		State:    p.State(),
		LGA:      p.LGA(),
		District: p.District(),
		Area:     p.Area(),
		Unit:     p.BuildingUnit(),
	}
}

// Disassemble returns the 5 constituent administrative segments of the postcode.
// Alias for Segments().
func (p Postcode) Disassemble() Segments {
	return p.Segments()
}

// Raw returns the compact 11-character representation (e.g. "EK01A03FK01").
func (p Postcode) Raw() string {
	if p.IsZero() {
		return ""
	}
	return string(p.raw[:])
}

// Compact returns the compact 11-character representation without delimiters (e.g. "EK01A03FK01").
// Alias for Raw().
func (p Postcode) Compact() string {
	return p.Raw()
}

// Formatted returns the canonical hyphenated representation (e.g. "EK-01-A03-FK-01").
func (p Postcode) Formatted() string {
	if p.IsZero() {
		return ""
	}
	b := make([]byte, 0, 15)
	b = append(b, p.raw[0:2]...)
	b = append(b, '-')
	b = append(b, p.raw[2:4]...)
	b = append(b, '-')
	b = append(b, p.raw[4:7]...)
	b = append(b, '-')
	b = append(b, p.raw[7:9]...)
	b = append(b, '-')
	b = append(b, p.raw[9:11]...)
	return string(b)
}

// Location returns the resolved geographic location and mapping URLs for this postcode.
// Resolves offline via built-in administrative and building registries without network calls.
func (p Postcode) Location() Location {
	return resolvePostcodeLocation(p)
}

// GoogleMapsURL returns a direct Google Maps search URL with exact coordinates.
func (p Postcode) GoogleMapsURL() string {
	return p.Location().GoogleMapsURL()
}

// Coordinates returns the latitude and longitude coordinates for this postcode.
func (p Postcode) Coordinates() (lat, lng float64) {
	loc := p.Location()
	return loc.Latitude, loc.Longitude
}

func (p Postcode) IsZero() bool {
	return p.raw == [11]byte{}
}

// AppendText implements encoding.TextAppender for zero-alloc formatting.
// Outputs the spaced format: "AA 99 H77 BB 55".
func (p Postcode) AppendText(b []byte) ([]byte, error) {
	if p.IsZero() {
		return b, nil
	}
	b = append(b, p.raw[0:2]...)
	b = append(b, ' ')
	b = append(b, p.raw[2:4]...)
	b = append(b, ' ')
	b = append(b, p.raw[4:7]...)
	b = append(b, ' ')
	b = append(b, p.raw[7:9]...)
	b = append(b, ' ')
	b = append(b, p.raw[9:11]...)
	return b, nil
}

func (p Postcode) MarshalText() ([]byte, error) {
	return p.AppendText(make([]byte, 0, 15))
}

func (p *Postcode) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

// String returns the spaced format (e.g. "EK 01 A03 FK 01").
func (p Postcode) String() string {
	if p.IsZero() {
		return ""
	}
	buf, _ := p.AppendText(make([]byte, 0, 15))
	return string(buf)
}

// ParseSeq streams parsed postcodes as (Postcode, error) pairs using Go iterators.
func ParseSeq(inputs []string) iter.Seq2[Postcode, error] {
	return func(yield func(Postcode, error) bool) {
		for _, raw := range inputs {
			p, err := Parse(raw)
			if !yield(p, err) {
				return
			}
		}
	}
}

// BatchParseSeq yields an iter.Seq2 iterator with slice index and Result wrapper.
func BatchParseSeq(inputs []string) iter.Seq2[int, Result[Postcode]] {
	return func(yield func(int, Result[Postcode]) bool) {
		for i, raw := range inputs {
			p, err := Parse(raw)
			if !yield(i, Result[Postcode]{Value: p, Err: err}) {
				return
			}
		}
	}
}

type Result[T any] struct {
	Value T
	Err   error
}

func isAlpha(b byte) bool { return (b >= 'A' && b <= 'Z') }
func isDigit(b byte) bool { return (b >= '0' && b <= '9') }
func isAlnum(b byte) bool { return isAlpha(b) || isDigit(b) }
