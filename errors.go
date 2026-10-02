package postcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

// UnmarshalJSON handles both the official nested envelope {"error": {"code": ..., "message": ...}}
// and flat payloads {"code": ..., "message": ...}.
func (e *APIError) UnmarshalJSON(data []byte) error {
	var nested struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &nested); err == nil && (nested.Error.Code != "" || nested.Error.Message != "") {
		e.Code = nested.Error.Code
		e.Message = nested.Error.Message
		return nil
	}

	var flat struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &flat); err == nil {
		e.Code = flat.Code
		e.Message = flat.Message
	}
	return nil
}

func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.Code == "rate_limit_exceeded"
}

func (e *APIError) IsInsufficientCredits() bool {
	return e.StatusCode == http.StatusPaymentRequired || e.Code == "insufficient_credits"
}

func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
