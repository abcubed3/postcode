//go:build !js

package postcode

import "database/sql/driver"

var _ driver.Valuer = Postcode{}

// Value implements driver.Valuer, storing the canonical hyphenated postcode (or nil if zero).
func (p Postcode) Value() (driver.Value, error) {
	if p.IsZero() {
		return nil, nil
	}
	return p.Formatted(), nil
}
