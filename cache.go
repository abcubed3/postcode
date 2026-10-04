package postcode

import (
	"sync"
	"time"
)

// Cache defines the interface for caching postcode gateway lookups and resolutions.
// Enables idempotency and prevents redundant paid API calls during AI agent loops.
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, val any, ttl time.Duration)
	Delete(key string)
	Clear()
	Len() int
}

type cacheItem struct {
	value     any
	expiresAt time.Time
}

func (i cacheItem) isExpired(now time.Time) bool {
	if i.expiresAt.IsZero() {
		return false
	}
	return now.After(i.expiresAt)
}

// MemoryCache provides a thread-safe, in-memory TTL cache with size bounding.
type MemoryCache struct {
	mu         sync.RWMutex
	items      map[string]cacheItem
	maxEntries int
	defaultTTL time.Duration
}

// NewMemoryCache creates a memory cache. If maxEntries <= 0, defaults to 1000 entries.
// If defaultTTL <= 0, defaults to 30 minutes.
func NewMemoryCache(maxEntries int, defaultTTL time.Duration) *MemoryCache {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	if defaultTTL <= 0 {
		defaultTTL = 30 * time.Minute
	}
	return &MemoryCache{
		items:      make(map[string]cacheItem, maxEntries),
		maxEntries: maxEntries,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves an item from the cache if present and unexpired.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	now := time.Now()
	if item.isExpired(now) {
		// Lazily clean up expired entry with double-checked lock
		c.mu.Lock()
		if it, exists := c.items[key]; exists && it.isExpired(time.Now()) {
			delete(c.items, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return item.value, true
}

// Set stores a value in the cache with an expiration duration.
// If ttl <= 0, the cache's defaultTTL is used.
func (c *MemoryCache) Set(key string, val any, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	// If at capacity and key is new, evict expired items first
	if len(c.items) >= c.maxEntries {
		if _, exists := c.items[key]; !exists {
			evicted := false
			for k, it := range c.items {
				if it.isExpired(now) {
					delete(c.items, k)
					evicted = true
				}
			}
			// If still at capacity, evict an arbitrary entry
			if !evicted && len(c.items) >= c.maxEntries {
				for k := range c.items {
					delete(c.items, k)
					break
				}
			}
		}
	}

	c.items[key] = cacheItem{
		value:     val,
		expiresAt: now.Add(ttl),
	}
}

// Delete removes a specific key from the cache.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Clear flushes all entries from the cache.
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]cacheItem, c.maxEntries)
	c.mu.Unlock()
}

// Len returns the current count of items in the cache.
func (c *MemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
