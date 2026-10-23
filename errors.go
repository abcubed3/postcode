package postcode

import (
	"fmt"
	"time"
)

// RateLimitInfo contains the rate-limiting quotas returned by the NIPOST API gateway
// as documented in https://docs.postcode.gov.ng/authentication.
type RateLimitInfo struct {
	Limit         int           // Maximum requests allowed in current window (from X-RateLimit-Limit)
	Remaining     int           // Remaining requests in current window (from X-RateLimit-Remaining)
	CreditBalance int           // Remaining commercial credit balance, or -1 if header not present (from X-Credit-Balance)
	ResetAt       time.Time     // Timestamp when quota resets (from X-RateLimit-Reset or Retry-After)
	RetryAfter    time.Duration // Cooldown duration requested by server (from Retry-After or X-RateLimit-Reset)
}

// RetryConfig configures backoff, jitter, and rate-limit handling for HTTP requests.
type RetryConfig struct {
	MaxRetries        int           // Maximum retry attempts for transient errors (default: 3)
	BaseDelay         time.Duration // Initial exponential backoff delay (default: 100ms)
	MaxDelay          time.Duration // Maximum exponential backoff delay (default: 2s)
	DisableRateLimit  bool          // If true, 429 Too Many Requests will not be retried automatically
	MaxRateLimitDelay time.Duration // Maximum sleep duration client will wait for a 429 cooldown (default: 10s)
}

// APIError represents an error returned by the NIPOST Postcode API gateway.
type APIError struct {
	StatusCode int            `json:"-"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	RateLimit  *RateLimitInfo `json:"-"`
	RetryAfter time.Duration  `json:"-"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		if e.RetryAfter > 0 {
			return fmt.Sprintf("postcode: http %d [%s]: %s (retry after %v)", e.StatusCode, e.Code, e.Message, e.RetryAfter)
		}
		return fmt.Sprintf("postcode: http %d [%s]: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("postcode: http %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == 429 || e.Code == "rate_limit_exceeded"
}

func (e *APIError) IsInsufficientCredits() bool {
	return e.StatusCode == 402 || e.Code == "insufficient_credits"
}

func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}
