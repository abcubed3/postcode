package postcode

import (
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// RateLimitInfo contains the rate-limiting quotas returned by the NIPOST API gateway
// as documented in https://docs.postcode.gov.ng/authentication.
type RateLimitInfo struct {
	Limit      int           // Maximum requests allowed in current window (from X-RateLimit-Limit)
	Remaining  int           // Remaining requests in current window (from X-RateLimit-Remaining)
	ResetAt    time.Time     // Timestamp when quota resets (from X-RateLimit-Reset or Retry-After)
	RetryAfter time.Duration // Cooldown duration requested by server (from Retry-After or X-RateLimit-Reset)
}

// ParseRateLimitInfo extracts rate limit quotas and cooldown durations from HTTP response headers.
func ParseRateLimitInfo(h http.Header) *RateLimitInfo {
	if h == nil {
		return nil
	}
	limitStr := h.Get("X-RateLimit-Limit")
	remStr := h.Get("X-RateLimit-Remaining")
	resetStr := h.Get("X-RateLimit-Reset")
	retryAfterStr := h.Get("Retry-After")

	if limitStr == "" && remStr == "" && resetStr == "" && retryAfterStr == "" {
		return nil
	}

	info := &RateLimitInfo{}
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			info.Limit = v
		}
	}
	if remStr != "" {
		if v, err := strconv.Atoi(remStr); err == nil {
			info.Remaining = v
		}
	}

	// 1. Parse Retry-After (seconds integer or RFC1123 HTTP date)
	if retryAfterStr != "" {
		if sec, err := strconv.Atoi(retryAfterStr); err == nil && sec >= 0 {
			info.RetryAfter = time.Duration(sec) * time.Second
			info.ResetAt = time.Now().Add(info.RetryAfter)
		} else if targetTime, err := http.ParseTime(retryAfterStr); err == nil {
			if d := time.Until(targetTime); d > 0 {
				info.RetryAfter = d
				info.ResetAt = targetTime
			}
		}
	}

	// 2. Fall back to X-RateLimit-Reset if ResetAt not yet determined
	if info.ResetAt.IsZero() && resetStr != "" {
		if val, err := strconv.ParseInt(resetStr, 10, 64); err == nil && val > 0 {
			if val > 1_000_000_000 {
				// Unix epoch timestamp in seconds
				info.ResetAt = time.Unix(val, 0)
				if d := time.Until(info.ResetAt); d > 0 {
					info.RetryAfter = d
				}
			} else {
				// Relative delta seconds until reset
				info.RetryAfter = time.Duration(val) * time.Second
				info.ResetAt = time.Now().Add(info.RetryAfter)
			}
		}
	}

	return info
}

// RetryConfig configures backoff, jitter, and rate-limit handling for HTTP requests.
type RetryConfig struct {
	MaxRetries        int           // Maximum retry attempts for transient errors (default: 3)
	BaseDelay         time.Duration // Initial exponential backoff delay (default: 100ms)
	MaxDelay          time.Duration // Maximum exponential backoff delay (default: 2s)
	DisableRateLimit  bool          // If true, 429 Too Many Requests will not be retried automatically
	MaxRateLimitDelay time.Duration // Maximum sleep duration client will wait for a 429 cooldown (default: 10s)
}

type RetryTransport struct {
	Base   http.RoundTripper
	Config RetryConfig
}

func NewRetryTransport(base http.RoundTripper, cfg RetryConfig) *RetryTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 100 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 2 * time.Second
	}
	if cfg.MaxRateLimitDelay <= 0 {
		cfg.MaxRateLimitDelay = 10 * time.Second
	}
	return &RetryTransport{Base: base, Config: cfg}
}

func (r *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// If the request has a non-empty body that cannot be re-read, we cannot safely retry.
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return r.Base.RoundTrip(req)
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= r.Config.MaxRetries; attempt++ {
		// Rewind body if retrying
		if attempt > 0 && req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, bodyErr
			}
			req.Body = body
		}

		resp, err = r.Base.RoundTrip(req)

		if req.Context().Err() != nil {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			return nil, req.Context().Err()
		}

		if !r.shouldRetry(resp, err) || attempt == r.Config.MaxRetries {
			return resp, err
		}

		delay := r.calculateDelay(attempt, resp)

		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-req.Context().Done():
			t.Stop()
			return nil, req.Context().Err()
		}
	}

	return resp, err
}

func (r *RetryTransport) shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true // Connection reset, dial timeout, or network transient error
	}
	if resp == nil {
		return false
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		if r.Config.DisableRateLimit {
			return false
		}
		// If the server-requested cooldown exceeds MaxRateLimitDelay, do not retry and fail fast
		info := ParseRateLimitInfo(resp.Header)
		if info != nil && info.RetryAfter > 0 && info.RetryAfter > r.Config.MaxRateLimitDelay {
			return false
		}
		return true
	}
	return resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout
}

func (r *RetryTransport) calculateDelay(attempt int, resp *http.Response) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		info := ParseRateLimitInfo(resp.Header)
		if info != nil && info.RetryAfter > 0 {
			// Add a minor jitter (up to 5% of cooldown, max 200ms) to prevent thundering herd when rate limit window resets
			jitter := time.Duration(float64(info.RetryAfter) * 0.05 * rand.Float64())
			if jitter > 200*time.Millisecond {
				jitter = 200 * time.Millisecond
			}
			delay := info.RetryAfter + jitter
			if delay > r.Config.MaxRateLimitDelay {
				return r.Config.MaxRateLimitDelay
			}
			return delay
		}
	}

	// Exponential backoff with rand/v2 jitter (Full Jitter pattern)
	factor := time.Duration(1 << attempt)
	backoff := min(r.Config.MaxDelay, r.Config.BaseDelay*factor)
	jitter := time.Duration(float64(backoff) * (0.8 + 0.4*rand.Float64()))
	return jitter
}
