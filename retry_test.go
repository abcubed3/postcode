package postcode

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryTransport_RetriesOn503(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := attempts.Add(1)
		if att < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"status":"ok"}}`))
	}))
	defer server.Close()

	rt := NewRetryTransport(http.DefaultTransport, RetryConfig{
		MaxRetries: 3,
		BaseDelay:  5 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	})

	httpClient := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("httpClient.Do failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if attempts.Load() != 3 {
		t.Errorf("attempts = %d, want 3", attempts.Load())
	}
}

func TestRetryTransport_ContextCancellation(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	rt := NewRetryTransport(http.DefaultTransport, RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	client := &http.Client{Transport: rt}
	_, err = client.Do(req)
	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
}

func TestParseRateLimitInfo(t *testing.T) {
	t.Parallel()

	// 1. All nil / empty
	if info := ParseRateLimitInfo(nil); info != nil {
		t.Errorf("expected nil for nil header, got %+v", info)
	}
	if info := ParseRateLimitInfo(http.Header{}); info != nil {
		t.Errorf("expected nil for empty header, got %+v", info)
	}

	// 2. Standard NIPOST headers with integer Retry-After
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "600")
	h.Set("X-RateLimit-Remaining", "597")
	h.Set("Retry-After", "5")

	info := ParseRateLimitInfo(h)
	if info == nil {
		t.Fatal("expected non-nil RateLimitInfo")
	}
	if info.Limit != 600 {
		t.Errorf("info.Limit = %d, want 600", info.Limit)
	}
	if info.Remaining != 597 {
		t.Errorf("info.Remaining = %d, want 597", info.Remaining)
	}
	if info.RetryAfter != 5*time.Second {
		t.Errorf("info.RetryAfter = %v, want 5s", info.RetryAfter)
	}

	// 3. X-RateLimit-Reset Unix timestamp
	h2 := http.Header{}
	future := time.Now().Add(10 * time.Second).Unix()
	h2.Set("X-RateLimit-Reset", strconv.FormatInt(future, 10))
	info2 := ParseRateLimitInfo(h2)
	if info2 == nil {
		t.Fatal("expected non-nil RateLimitInfo for X-RateLimit-Reset")
	}
	if info2.RetryAfter <= 0 || info2.RetryAfter > 11*time.Second {
		t.Errorf("unexpected RetryAfter from reset: %v", info2.RetryAfter)
	}
}

func TestRetryTransport_RateLimitRetry(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := attempts.Add(1)
		if att < 3 {
			w.Header().Set("X-RateLimit-Limit", "600")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"too many requests"}}`))
			return
		}
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "597")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"status":"ok"}}`))
	}))
	defer server.Close()

	rt := NewRetryTransport(http.DefaultTransport, RetryConfig{
		MaxRetries:        3,
		BaseDelay:         5 * time.Millisecond,
		MaxDelay:          20 * time.Millisecond,
		MaxRateLimitDelay: 1 * time.Second,
	})

	httpClient := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("httpClient.Do failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if attempts.Load() != 3 {
		t.Errorf("attempts = %d, want 3", attempts.Load())
	}
}

func TestRetryTransport_RateLimit_ExceedsMaxRateLimitDelay(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("Retry-After", "60") // Server asks for 60s cooldown
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"rate limit exceeded"}}`))
	}))
	defer server.Close()

	rt := NewRetryTransport(http.DefaultTransport, RetryConfig{
		MaxRetries:        3,
		MaxRateLimitDelay: 100 * time.Millisecond, // Client only willing to wait 100ms
	})

	httpClient := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("httpClient.Do failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", resp.StatusCode)
	}
	// Should fail immediately without retrying because 60s > 100ms
	if attempts.Load() != 1 {
		t.Errorf("attempts = %d, want 1 (should fail fast)", attempts.Load())
	}
}

func TestRetryTransport_RateLimit_Disabled(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	rt := NewRetryTransport(http.DefaultTransport, RetryConfig{
		MaxRetries:       3,
		DisableRateLimit: true, // Rate limit retries explicitly disabled
	})

	httpClient := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("httpClient.Do failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", resp.StatusCode)
	}
	if attempts.Load() != 1 {
		t.Errorf("attempts = %d, want 1 when DisableRateLimit is true", attempts.Load())
	}
}
