package postcode

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryCache_Basic(t *testing.T) {
	c := NewMemoryCache(10, 50*time.Millisecond)

	// Set & Get
	c.Set("key1", "val1", 0)
	val, ok := c.Get("key1")
	if !ok || val != "val1" {
		t.Fatalf("expected val1, got %v", val)
	}

	if c.Len() != 1 {
		t.Errorf("expected len 1, got %d", c.Len())
	}

	// Delete
	c.Delete("key1")
	_, ok = c.Get("key1")
	if ok {
		t.Errorf("expected key1 to be deleted")
	}

	// Clear
	c.Set("k2", "v2", 0)
	c.Set("k3", "v3", 0)
	c.Clear()
	if c.Len() != 0 {
		t.Errorf("expected len 0 after clear, got %d", c.Len())
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	c := NewMemoryCache(10, 20*time.Millisecond)

	c.Set("expire_me", "val", 10*time.Millisecond)
	time.Sleep(25 * time.Millisecond)

	_, ok := c.Get("expire_me")
	if ok {
		t.Errorf("expected expire_me to have expired")
	}
}

func TestMemoryCache_CapacityEviction(t *testing.T) {
	c := NewMemoryCache(2, 1*time.Minute)

	c.Set("k1", "v1", 0)
	c.Set("k2", "v2", 0)
	c.Set("k3", "v3", 0) // exceeds capacity 2

	if c.Len() > 2 {
		t.Errorf("cache exceeded capacity 2, len=%d", c.Len())
	}
}

func TestMemoryCache_Concurrency(t *testing.T) {
	c := NewMemoryCache(100, 1*time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key"
			c.Set(key, idx, 0)
			_, _ = c.Get(key)
		}(i)
	}

	wg.Wait()
}
