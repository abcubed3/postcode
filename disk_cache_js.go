//go:build js && wasm

package postcode

import "sync"

var (
	diskCacheMu   sync.Mutex
	diskCachePath = ":memory:"
)

// DefaultCachePath returns a virtual in-memory cache identifier on js/wasm.
func DefaultCachePath() string {
	return ":memory:"
}

// SetCachePath sets the virtual cache identifier on js/wasm.
func SetCachePath(path string) error {
	diskCacheMu.Lock()
	diskCachePath = path
	diskCacheMu.Unlock()
	return nil
}

// GetCachePath returns the active cache identifier on js/wasm.
func GetCachePath() string {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()
	return diskCachePath
}

func ensureCacheLoaded() {}

// LoadDiskCache is a no-op on js/wasm (buildings are stored in memory).
func LoadDiskCache() error {
	return nil
}

// SaveDiskCache is a no-op on js/wasm (buildings are stored in memory).
func SaveDiskCache() error {
	return nil
}

// ClearDiskCache clears all custom registered building records in memory on js/wasm.
func ClearDiskCache() error {
	customMu.Lock()
	customBuildings = make(map[string]BuildingRecord)
	customMu.Unlock()
	return nil
}

// GetCacheStats returns the number of custom cached building records in memory on js/wasm.
func GetCacheStats() (int, string) {
	customMu.RLock()
	count := len(customBuildings)
	customMu.RUnlock()
	return count, GetCachePath()
}
