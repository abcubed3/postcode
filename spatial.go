package postcode

import (
	"cmp"
	"math"
	"slices"
)

// AllBuildingRecords returns all known building benchmark records and loaded custom cached buildings.
func AllBuildingRecords() []BuildingRecord {
	ensureCacheLoaded()
	res := make([]BuildingRecord, 0, len(knownBuildings)+len(customBuildings))
	for _, b := range knownBuildings {
		res = append(res, b)
	}
	customMu.RLock()
	for _, b := range customBuildings {
		res = append(res, b)
	}
	customMu.RUnlock()
	return res
}

func haversineM(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000 // Earth's mean radius in meters
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

// SearchNearbyBuildingsOffline searches local embedded and cached buildings within radiusM meters of (lat, lng).
func SearchNearbyBuildingsOffline(lat, lng, radiusM float64) []NearbyUnit {
	all := AllBuildingRecords()
	if radiusM <= 0 {
		radiusM = 300.0
	}
	var results []NearbyUnit
	for _, b := range all {
		dist := haversineM(lat, lng, b.Latitude, b.Longitude)
		if dist <= radiusM {
			conf := "high"
			if dist > 50 {
				conf = "medium"
			}
			p, err := Parse(b.Postcode)
			canonical := b.Postcode
			display := b.Postcode
			if err == nil {
				canonical = p.Formatted()
				display = p.String()
			}
			results = append(results, NearbyUnit{
				Postcode:     canonical,
				Display:      display,
				DistanceM:    math.Round(dist*10) / 10,
				Confidence:   conf,
				StateName:    b.StateName,
				LGAName:      b.LGAName,
				LocalityName: b.LGAName,
				Address:      b.Address,
			})
		}
	}
	slices.SortFunc(results, func(a, b NearbyUnit) int {
		return cmp.Compare(a.DistanceM, b.DistanceM)
	})
	return results
}

// ReverseCoordinatesOffline resolves geographic coordinates to the closest building unit in the local database.
func ReverseCoordinatesOffline(lat, lng, maxDistanceM float64) *ReverseResponse {
	if maxDistanceM <= 0 {
		maxDistanceM = 25.0
	}
	if maxDistanceM > 250.0 {
		maxDistanceM = 250.0
	}
	units := SearchNearbyBuildingsOffline(lat, lng, maxDistanceM)
	res := &ReverseResponse{
		Coordinate: []float64{lng, lat},
		RadiusM:    maxDistanceM,
	}
	if len(units) > 0 {
		best := units[0]
		res.Found = true
		res.Unit = &best
		res.State = best.StateName
		if p, err := Parse(best.Postcode); err == nil {
			res.State = p.State()
			res.District = p.Formatted()[:8]
			res.Area = p.Formatted()[:11]
		}
	} else {
		res.Found = false
		res.Message = "no registered units found within search radius in offline database"
	}
	return res
}

// NearestLGA finds the single closest Nigerian LGA centroid to the given (lat, lng) coordinates.
func NearestLGA(lat, lng float64) (LGARecord, float64, bool) {
	if lat == 0 && lng == 0 {
		return LGARecord{}, 0, false
	}
	minDist := math.MaxFloat64
	var best LGARecord
	found := false
	for _, lga := range knownLGAs {
		if lga.Latitude == 0 && lga.Longitude == 0 {
			continue
		}
		dist := haversineM(lat, lng, lga.Latitude, lga.Longitude)
		if dist < minDist {
			minDist = dist
			best = lga
			found = true
		}
	}
	if found {
		return best, math.Round(minDist*10) / 10, true
	}
	return LGARecord{}, 0, false
}
