package postcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultBaseURL    = "https://api.postcode.gov.ng"
	maxDrainBytes     = 4096             // Drain limit to safely allow TCP connection reuse
	defaultTimeout    = 10 * time.Second // Default overall HTTP timeout
	maxPoolBufferSize = 65536            // Do not retain buffers larger than 64KB in pool
	maxResponseBytes  = 4 * 1024 * 1024  // 4MB safety ceiling against unbounded reads / OOM
)

var bufPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 8192))
	},
}

type ClientOption func(*Client)

type Client struct {
	baseURL       *url.URL
	apiKey        string
	httpClient    *http.Client
	baseTransport http.RoundTripper
	retryConfig   RetryConfig
	telemetry     Telemetry
	rateLimit     atomic.Pointer[RateLimitInfo]
	cache         Cache
	agentGuard    *agentGuardState
}

// NewClient instantiates a production-tuned HTTP client.
func NewClient(opts ...ClientOption) (*Client, error) {
	parsedURL, err := url.Parse(DefaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("postcode: invalid base URL: %w", err)
	}

	c := &Client{
		baseURL:     parsedURL,
		retryConfig: RetryConfig{},
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		telemetry: NopTelemetry(),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Wire transport stack: If a custom top-level transport was not set via WithTransport,
	// wrap the base transport with RetryTransport.
	if c.httpClient.Transport == nil {
		base := c.baseTransport
		if base == nil {
			if runtime.GOOS == "js" {
				base = http.DefaultTransport
			} else {
				base = &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   5 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					MaxIdleConnsPerHost:   100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   5 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				}
			}
		}
		c.httpClient.Transport = NewRetryTransport(base, c.retryConfig)
	}

	return c, nil
}

// WithAPIKey sets the NIPOST API key sent in the X-API-Key header.
// An apiKey is optional for public endpoints (L1 lookup, search, assembly).
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithBaseURL overrides the default gateway URL (e.g. for testing with httptest.Server).
func WithBaseURL(customURL string) ClientOption {
	return func(c *Client) {
		if u, err := url.Parse(customURL); err == nil {
			c.baseURL = u
		}
	}
}

// WithTransport explicitly overrides the entire round tripper stack (bypassing retries).
func WithTransport(rt http.RoundTripper) ClientOption {
	return func(c *Client) {
		if rt != nil {
			c.httpClient.Transport = rt
		}
	}
}

// WithBaseTransport sets the underlying round tripper underneath the retry middleware.
func WithBaseTransport(rt http.RoundTripper) ClientOption {
	return func(c *Client) {
		if rt != nil {
			c.baseTransport = rt
		}
	}
}

// WithRetryConfig configures backoff, jitter, and retry limits.
func WithRetryConfig(cfg RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = cfg
	}
}

// WithTimeout sets the overall client request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithTelemetry registers custom observability hooks.
func WithTelemetry(t Telemetry) ClientOption {
	return func(c *Client) {
		if t.Tracer != nil {
			c.telemetry.Tracer = t.Tracer
		}
		if t.Metrics != nil {
			c.telemetry.Metrics = t.Metrics
		}
	}
}

// WithRateLimitRetry configures automatic retry on 429 Too Many Requests responses.
// If enabled is false, 429 responses will not be retried.
// maxDelay sets the maximum duration the client is willing to sleep for a rate limit cooldown.
func WithRateLimitRetry(enabled bool, maxDelay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryConfig.DisableRateLimit = !enabled
		if maxDelay > 0 {
			c.retryConfig.MaxRateLimitDelay = maxDelay
		}
	}
}

// WithCache configures a custom cache implementation for the client.
func WithCache(cache Cache) ClientOption {
	return func(c *Client) {
		c.cache = cache
	}
}

// WithDefaultCache configures a standard in-memory TTL cache (1000 items, 30m TTL).
func WithDefaultCache() ClientOption {
	return func(c *Client) {
		c.cache = NewMemoryCache(1000, 30*time.Minute)
	}
}

// WithAgentGuard configures budget and safety ceilings for autonomous agents.
func WithAgentGuard(cfg AgentGuardConfig) ClientOption {
	return func(c *Client) {
		c.agentGuard = newAgentGuardState(cfg)
	}
}

// AgentMetrics returns live telemetry metrics for calls executed under agent guardrails.
func (c *Client) AgentMetrics() AgentGuardMetrics {
	if c.agentGuard == nil {
		return AgentGuardMetrics{}
	}
	return c.agentGuard.metrics()
}

// RateLimit returns the most recent rate limit quotas reported by the NIPOST gateway,
// or nil if no request has yet returned rate limit headers.
func (c *Client) RateLimit() *RateLimitInfo {
	return c.rateLimit.Load()
}

func (c *Client) execute(ctx context.Context, method string, pathParts []string, q url.Values, reqBody any, dst any) error {
	u := c.baseURL.JoinPath(pathParts...)
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}

	endpointPath := u.Path
	ctx, span := c.telemetry.Tracer.Start(ctx, "HTTP "+method+" "+endpointPath)
	defer span.End()

	span.SetAttribute("http.request.method", method)
	span.SetAttribute("server.address", u.Hostname())
	span.SetAttribute("url.full", u.String())

	var bodyReader io.Reader = http.NoBody
	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("postcode: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("postcode: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	c.telemetry.Metrics.AddInflight(ctx, 1)
	defer c.telemetry.Metrics.AddInflight(ctx, -1)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(http.StatusInternalServerError, err.Error())
		c.telemetry.Metrics.RecordDuration(ctx, method, endpointPath, 0, duration)
		return fmt.Errorf("postcode: transport: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrainBytes))
		_ = resp.Body.Close()
	}()

	statusCode := resp.StatusCode
	span.SetAttribute("http.response.status_code", strconv.Itoa(statusCode))
	span.SetStatus(statusCode, http.StatusText(statusCode))
	c.telemetry.Metrics.RecordDuration(ctx, method, endpointPath, statusCode, duration)

	rl := ParseRateLimitInfo(resp.Header)
	if rl != nil {
		c.rateLimit.Store(rl)
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPoolBufferSize {
			bufPool.Put(buf)
		}
	}()

	// Read stream bounded by maxResponseBytes to defend against OOM/DoS attacks
	limitedReader := io.LimitReader(resp.Body, maxResponseBytes)
	if _, err := buf.ReadFrom(limitedReader); err != nil {
		span.RecordError(err)
		return fmt.Errorf("postcode: read stream: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		var apiErr APIError
		if err := json.Unmarshal(buf.Bytes(), &apiErr); err == nil && (apiErr.Message != "" || apiErr.Code != "") {
			apiErr.StatusCode = statusCode
			apiErr.RateLimit = rl
			if rl != nil {
				apiErr.RetryAfter = rl.RetryAfter
			}
			span.RecordError(&apiErr)
			return &apiErr
		}
		apiErr = APIError{
			StatusCode: statusCode,
			Message:    http.StatusText(statusCode),
			RateLimit:  rl,
		}
		if rl != nil {
			apiErr.RetryAfter = rl.RetryAfter
		}
		span.RecordError(&apiErr)
		return &apiErr
	}

	if dst != nil {
		var env struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(buf.Bytes(), &env); err == nil && len(env.Data) > 0 && string(env.Data) != "null" {
			if err := json.Unmarshal(env.Data, dst); err != nil {
				span.RecordError(err)
				return fmt.Errorf("postcode: decode response payload: %w", err)
			}
			return nil
		}

		if err := json.Unmarshal(buf.Bytes(), dst); err != nil {
			span.RecordError(err)
			return fmt.Errorf("postcode: decode response: %w", err)
		}
	}

	return nil
}
