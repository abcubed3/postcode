package postcode

import (
	"context"
	"fmt"
)

// Geocoder resolves a geographic textual address or landmark query to latitude and longitude coordinates.
type Geocoder interface {
	Geocode(ctx context.Context, query string) (lat, lng float64, err error)
}

// ChainGeocoder attempts each provided Geocoder in sequence until one succeeds.
type ChainGeocoder struct {
	Geocoders []Geocoder
}

// NewChainGeocoder creates a composite geocoder.
func NewChainGeocoder(geocoders ...Geocoder) *ChainGeocoder {
	valid := make([]Geocoder, 0, len(geocoders))
	for _, g := range geocoders {
		if g != nil {
			valid = append(valid, g)
		}
	}
	return &ChainGeocoder{Geocoders: valid}
}

// Geocode executes the geocoder chain in priority order.
func (c *ChainGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	var lastErr error
	for _, g := range c.Geocoders {
		lat, lng, err := g.Geocode(ctx, query)
		if err == nil && (lat != 0 || lng != 0) {
			return lat, lng, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return 0, 0, lastErr
	}
	return 0, 0, fmt.Errorf("chain geocoder: no geocoder succeeded")
}
