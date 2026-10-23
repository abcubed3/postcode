//go:build js && wasm

package postcode

import (
	"context"
	"fmt"
	rand "math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall/js"
	"time"
)

const (
	DefaultBaseURL = "https://api.postcode.gov.ng"
	defaultTimeout = 10 * time.Second
)

type ClientOption func(*Client)

type Client struct {
	baseURL     *url.URL
	apiKey      string
	timeout     time.Duration
	retryConfig RetryConfig
	telemetry   Telemetry
	rateLimit   atomic.Pointer[RateLimitInfo]
	cache       Cache
	agentGuard  *agentGuardState
	geocoder    Geocoder
}

// NewClient instantiates a WebAssembly HTTP client backed by browser/Node fetch().
func NewClient(opts ...ClientOption) (*Client, error) {
	parsedURL, err := url.Parse(DefaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("postcode: invalid base URL: %w", err)
	}

	c := &Client{
		baseURL: parsedURL,
		timeout: defaultTimeout,
		retryConfig: RetryConfig{
			MaxRetries:        3,
			BaseDelay:         100 * time.Millisecond,
			MaxDelay:          2 * time.Second,
			MaxRateLimitDelay: 10 * time.Second,
		},
		telemetry: NopTelemetry(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.geocoder == nil {
		c.geocoder = DefaultGeocoder("")
	}

	return c, nil
}

func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

func WithBaseURL(customURL string) ClientOption {
	return func(c *Client) {
		if u, err := url.Parse(customURL); err == nil {
			c.baseURL = u
		}
	}
}

func WithRetryConfig(cfg RetryConfig) ClientOption {
	return func(c *Client) {
		if cfg.MaxRetries > 0 {
			c.retryConfig.MaxRetries = cfg.MaxRetries
		}
		if cfg.BaseDelay > 0 {
			c.retryConfig.BaseDelay = cfg.BaseDelay
		}
		if cfg.MaxDelay > 0 {
			c.retryConfig.MaxDelay = cfg.MaxDelay
		}
		if cfg.MaxRateLimitDelay > 0 {
			c.retryConfig.MaxRateLimitDelay = cfg.MaxRateLimitDelay
		}
		c.retryConfig.DisableRateLimit = cfg.DisableRateLimit
	}
}

func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = d
	}
}

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

func WithRateLimitRetry(enabled bool, maxDelay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryConfig.DisableRateLimit = !enabled
		if maxDelay > 0 {
			c.retryConfig.MaxRateLimitDelay = maxDelay
		}
	}
}

func WithCache(cache Cache) ClientOption {
	return func(c *Client) {
		c.cache = cache
	}
}

func WithDefaultCache() ClientOption {
	return func(c *Client) {
		c.cache = NewMemoryCache(1000, 30*time.Minute)
	}
}

func WithGeocoder(g Geocoder) ClientOption {
	return func(c *Client) {
		c.geocoder = g
	}
}

func WithGoogleMapsKey(key string) ClientOption {
	return func(c *Client) {
		c.geocoder = DefaultGeocoder(key)
	}
}

func WithAgentGuard(cfg AgentGuardConfig) ClientOption {
	return func(c *Client) {
		c.agentGuard = newAgentGuardState(cfg)
	}
}

func (c *Client) AgentMetrics() AgentGuardMetrics {
	if c.agentGuard == nil {
		return AgentGuardMetrics{}
	}
	return c.agentGuard.metrics()
}

func (c *Client) RateLimit() *RateLimitInfo {
	return c.rateLimit.Load()
}

type jsFetchResponse struct {
	statusCode int
	statusText string
	body       string
	headers    map[string]string
}

func doJSFetch(ctx context.Context, method, fullURL string, headers map[string]string, body string) (*jsFetchResponse, error) {
	global := js.Global()
	fetchFn := global.Get("fetch")
	if fetchFn.Type() != js.TypeFunction {
		return nil, fmt.Errorf("postcode: global fetch() is not available in this WebAssembly environment")
	}

	opts := global.Get("Object").New()
	opts.Set("method", method)

	hdrObj := global.Get("Object").New()
	for k, v := range headers {
		hdrObj.Set(k, v)
	}
	opts.Set("headers", hdrObj)
	if body != "" {
		opts.Set("body", body)
	}

	type fetchResult struct {
		resp *jsFetchResponse
		err  error
	}
	ch := make(chan fetchResult, 1)

	var thenFunc, catchFunc js.Func
	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer thenFunc.Release()
		respVal := args[0]
		statusCode := respVal.Get("status").Int()
		statusText := ""
		if st := respVal.Get("statusText"); st.Type() == js.TypeString {
			statusText = st.String()
		}

		hdrMap := make(map[string]string, 5)
		jsHdrs := respVal.Get("headers")
		if jsHdrs.Type() == js.TypeObject && jsHdrs.Get("get").Type() == js.TypeFunction {
			for _, hName := range []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After", "X-Credit-Balance"} {
				v := jsHdrs.Call("get", hName)
				if v.Type() == js.TypeString && v.String() != "" {
					hdrMap[hName] = v.String()
				}
			}
		}

		textPromise := respVal.Call("text")
		var textThen, textCatch js.Func
		textThen = js.FuncOf(func(this js.Value, tArgs []js.Value) any {
			defer textThen.Release()
			defer textCatch.Release()
			bodyStr := ""
			if len(tArgs) > 0 && tArgs[0].Type() == js.TypeString {
				bodyStr = tArgs[0].String()
			}
			ch <- fetchResult{
				resp: &jsFetchResponse{
					statusCode: statusCode,
					statusText: statusText,
					body:       bodyStr,
					headers:    hdrMap,
				},
			}
			return nil
		})
		textCatch = js.FuncOf(func(this js.Value, cArgs []js.Value) any {
			defer textThen.Release()
			defer textCatch.Release()
			msg := "failed reading response text"
			if len(cArgs) > 0 {
				msg = cArgs[0].Call("toString").String()
			}
			ch <- fetchResult{err: fmt.Errorf("postcode: %s", msg)}
			return nil
		})
		textPromise.Call("then", textThen).Call("catch", textCatch)
		return nil
	})

	catchFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer catchFunc.Release()
		defer thenFunc.Release()
		msg := "fetch error"
		if len(args) > 0 {
			msg = args[0].Call("toString").String()
		}
		ch <- fetchResult{err: fmt.Errorf("postcode: transport: %s", msg)}
		return nil
	})

	fetchFn.Invoke(fullURL, opts).Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case res := <-ch:
		return res.resp, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func parseJSRateLimitInfo(h map[string]string) *RateLimitInfo {
	if len(h) == 0 {
		return nil
	}
	limitStr := h["X-RateLimit-Limit"]
	remStr := h["X-RateLimit-Remaining"]
	resetStr := h["X-RateLimit-Reset"]
	retryAfterStr := h["Retry-After"]
	creditStr := h["X-Credit-Balance"]

	if limitStr == "" && remStr == "" && resetStr == "" && retryAfterStr == "" && creditStr == "" {
		return nil
	}

	info := &RateLimitInfo{CreditBalance: -1}
	if v, err := strconv.Atoi(limitStr); err == nil {
		info.Limit = v
	}
	if v, err := strconv.Atoi(remStr); err == nil {
		info.Remaining = v
	}
	if v, err := strconv.Atoi(creditStr); err == nil {
		info.CreditBalance = v
	}
	if retryAfterStr != "" {
		if sec, err := strconv.Atoi(retryAfterStr); err == nil && sec >= 0 {
			info.RetryAfter = time.Duration(sec) * time.Second
			info.ResetAt = time.Now().Add(info.RetryAfter)
		}
	}
	if info.ResetAt.IsZero() && resetStr != "" {
		if val, err := strconv.ParseInt(resetStr, 10, 64); err == nil && val > 0 {
			if val > 1_000_000_000 {
				info.ResetAt = time.Unix(val, 0)
				if d := time.Until(info.ResetAt); d > 0 {
					info.RetryAfter = d
				}
			} else {
				info.RetryAfter = time.Duration(val) * time.Second
				info.ResetAt = time.Now().Add(info.RetryAfter)
			}
		}
	}
	return info
}

func parseJSONSafe(raw string) (val js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid json: %v", r)
		}
	}()
	val = js.Global().Get("JSON").Call("parse", raw)
	return val, nil
}

func jsStringField(obj js.Value, key string) string {
	if obj.Type() != js.TypeObject {
		return ""
	}
	v := obj.Get(key)
	if v.Type() == js.TypeString {
		return v.String()
	}
	return ""
}

func jsFloatField(obj js.Value, key string) float64 {
	if obj.Type() != js.TypeObject {
		return 0
	}
	v := obj.Get(key)
	if v.Type() == js.TypeNumber {
		return v.Float()
	}
	return 0
}

func jsBoolField(obj js.Value, key string) bool {
	if obj.Type() != js.TypeObject {
		return false
	}
	v := obj.Get(key)
	if v.Type() == js.TypeBoolean {
		return v.Bool()
	}
	return false
}

func (c *Client) execute(ctx context.Context, method string, pathParts []string, q url.Values, reqBody any, dst any) error {
	u := c.baseURL.JoinPath(pathParts...)
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}

	endpointPath := u.Path
	ctx, span := c.telemetry.Tracer.Start(ctx, "HTTP "+method+" "+endpointPath)
	defer span.End()

	headers := map[string]string{
		"Accept": "application/json",
	}
	var bodyStr string
	if reqBody != nil {
		headers["Content-Type"] = "application/json"
		if segs, ok := reqBody.(Segments); ok {
			bodyStr = `{"state":` + strconv.Quote(segs.State) +
				`,"lga":` + strconv.Quote(segs.LGA) +
				`,"district":` + strconv.Quote(segs.District) +
				`,"area":` + strconv.Quote(segs.Area) +
				`,"unit":` + strconv.Quote(segs.Unit) + `}`
		}
	}
	if c.apiKey != "" {
		headers["X-API-Key"] = c.apiKey
	}

	var resp *jsFetchResponse
	var err error
	maxRetries := c.retryConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		start := time.Now()
		resp, err = doJSFetch(ctx, method, u.String(), headers, bodyStr)
		duration := time.Since(start)

		if err == nil && resp != nil {
			c.telemetry.Metrics.RecordDuration(ctx, method, endpointPath, resp.statusCode, duration)
		} else {
			c.telemetry.Metrics.RecordDuration(ctx, method, endpointPath, 0, duration)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		shouldRetry := err != nil || (resp != nil && (resp.statusCode == 502 || resp.statusCode == 503 || resp.statusCode == 504 || (resp.statusCode == 429 && !c.retryConfig.DisableRateLimit)))
		if !shouldRetry || attempt == maxRetries {
			break
		}

		factor := time.Duration(1 << attempt)
		baseDelay := c.retryConfig.BaseDelay
		if baseDelay <= 0 {
			baseDelay = 100 * time.Millisecond
		}
		maxDelay := c.retryConfig.MaxDelay
		if maxDelay <= 0 {
			maxDelay = 2 * time.Second
		}
		backoff := min(maxDelay, baseDelay*factor)
		delay := time.Duration(float64(backoff) * (0.8 + 0.4*rand.Float64()))
		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
	}

	if err != nil {
		span.RecordError(err)
		return err
	}

	rl := parseJSRateLimitInfo(resp.headers)
	if rl != nil {
		c.rateLimit.Store(rl)
	}

	if resp.statusCode < 200 || resp.statusCode >= 300 {
		apiErr := APIError{
			StatusCode: resp.statusCode,
			Message:    resp.statusText,
			RateLimit:  rl,
		}
		if rl != nil {
			apiErr.RetryAfter = rl.RetryAfter
		}
		if parsed, pErr := parseJSONSafe(resp.body); pErr == nil && parsed.Type() == js.TypeObject {
			if errObj := parsed.Get("error"); errObj.Type() == js.TypeObject {
				apiErr.Code = jsStringField(errObj, "code")
				if msg := jsStringField(errObj, "message"); msg != "" {
					apiErr.Message = msg
				}
			} else {
				apiErr.Code = jsStringField(parsed, "code")
				if msg := jsStringField(parsed, "message"); msg != "" {
					apiErr.Message = msg
				}
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = fmt.Sprintf("HTTP %d", resp.statusCode)
		}
		span.RecordError(&apiErr)
		return &apiErr
	}

	if dst == nil {
		return nil
	}

	parsed, pErr := parseJSONSafe(resp.body)
	if pErr != nil {
		span.RecordError(pErr)
		return fmt.Errorf("postcode: decode response: %w", pErr)
	}

	if parsed.Type() == js.TypeObject {
		if dataField := parsed.Get("data"); dataField.Type() == js.TypeObject {
			parsed = dataField
		}
	}

	decodeJSResponse(parsed, dst)
	return nil
}

func decodeJSNearbyUnit(u js.Value) NearbyUnit {
	return NearbyUnit{
		Postcode:     jsStringField(u, "postcode"),
		Display:      jsStringField(u, "display"),
		DistanceM:    jsFloatField(u, "distance_m"),
		Confidence:   jsStringField(u, "confidence"),
		StateName:    jsStringField(u, "state_name"),
		LGAName:      jsStringField(u, "lga_name"),
		LocalityName: jsStringField(u, "locality_name"),
		Address:      jsStringField(u, "address"),
	}
}

func decodeJSNamedCodes(arr js.Value) []NamedCode {
	if arr.Type() != js.TypeObject {
		return nil
	}
	n := arr.Length()
	out := make([]NamedCode, 0, n)
	for i := 0; i < n; i++ {
		item := arr.Index(i)
		out = append(out, NamedCode{
			Code: jsStringField(item, "code"),
			Name: jsStringField(item, "name"),
		})
	}
	return out
}

func decodeJSResponse(v js.Value, dst any) {
	switch out := dst.(type) {
	case *LookupResponse:
		out.Postcode = jsStringField(v, "postcode")
		out.Valid = jsBoolField(v, "valid")
		out.Temporary = jsBoolField(v, "temporary")
		out.BuildingUseStatus = jsStringField(v, "building_use_status")
		if adm := v.Get("administrative_address"); adm.Type() == js.TypeObject {
			out.AdministrativeAddress = &AdministrativeAddress{
				State:        jsStringField(adm, "state"),
				StateName:    jsStringField(adm, "state_name"),
				LGA:          jsStringField(adm, "lga"),
				LGAName:      jsStringField(adm, "lga_name"),
				District:     jsStringField(adm, "district"),
				DistrictName: jsStringField(adm, "district_name"),
				Area:         jsStringField(adm, "area"),
				AreaName:     jsStringField(adm, "area_name"),
				Unit:         jsStringField(adm, "unit"),
				Zone:         jsStringField(adm, "zone"),
				LocalityName: jsStringField(adm, "locality_name"),
			}
		}
		if rha := v.Get("recent_house_address"); rha.Type() == js.TypeObject {
			out.RecentHouseAddress = &RecentHouseAddress{
				Address:      jsStringField(rha, "address"),
				Recent:       jsStringField(rha, "recent"),
				HouseNumber:  jsStringField(rha, "house_number"),
				StreetName:   jsStringField(rha, "street_name"),
				LocalityName: jsStringField(rha, "locality_name"),
			}
		}
		if pg := v.Get("point_geometry"); pg.Type() == js.TypeObject {
			coordsVal := pg.Get("coordinates")
			var coords []float64
			if coordsVal.Type() == js.TypeObject && coordsVal.Length() >= 2 {
				coords = []float64{coordsVal.Index(0).Float(), coordsVal.Index(1).Float()}
			}
			out.PointGeometry = &PointGeometry{
				Type:        jsStringField(pg, "type"),
				Coordinates: coords,
			}
		}

	case *AutocompleteResponse:
		out.Segment = jsStringField(v, "segment")
		if sug := v.Get("suggestions"); sug.Type() == js.TypeObject {
			n := sug.Length()
			out.Suggestions = make([]AutocompleteSuggestion, 0, n)
			for i := 0; i < n; i++ {
				item := sug.Index(i)
				out.Suggestions = append(out.Suggestions, AutocompleteSuggestion{
					Code:  jsStringField(item, "code"),
					Label: jsStringField(item, "label"),
				})
			}
		}

	case *NearbyResponse:
		arr := v
		if resArr := v.Get("results"); resArr.Type() == js.TypeObject {
			arr = resArr
		}
		if arr.Type() == js.TypeObject {
			n := arr.Length()
			out.Results = make([]NearbyUnit, 0, n)
			for i := 0; i < n; i++ {
				out.Results = append(out.Results, decodeJSNearbyUnit(arr.Index(i)))
			}
		}

	case *ReverseResponse:
		out.Found = jsBoolField(v, "found")
		out.Area = jsStringField(v, "area")
		out.District = jsStringField(v, "district")
		out.State = jsStringField(v, "state")
		out.Message = jsStringField(v, "message")
		out.RadiusM = jsFloatField(v, "radius_m")
		if coordsVal := v.Get("coordinate"); coordsVal.Type() == js.TypeObject && coordsVal.Length() >= 2 {
			out.Coordinate = []float64{coordsVal.Index(0).Float(), coordsVal.Index(1).Float()}
		}
		if u := v.Get("unit"); u.Type() == js.TypeObject {
			unit := decodeJSNearbyUnit(u)
			out.Unit = &unit
		}

	case *AssembledPostcode:
		out.Postcode = jsStringField(v, "postcode")
		out.Display = jsStringField(v, "display")
		out.Compact = jsStringField(v, "compact")

	case *Segments:
		out.State = jsStringField(v, "state")
		out.LGA = jsStringField(v, "lga")
		out.District = jsStringField(v, "district")
		out.Area = jsStringField(v, "area")
		out.Unit = jsStringField(v, "unit")

	case *referenceStatesResponse:
		if arr := v.Get("states"); arr.Type() == js.TypeObject {
			out.States = decodeJSNamedCodes(arr)
		} else {
			out.States = decodeJSNamedCodes(v)
		}

	case *referenceLGAsResponse:
		if arr := v.Get("lgas"); arr.Type() == js.TypeObject {
			out.LGAs = decodeJSNamedCodes(arr)
		} else {
			out.LGAs = decodeJSNamedCodes(v)
		}

	case *referenceDistrictsResponse:
		if arr := v.Get("districts"); arr.Type() == js.TypeObject {
			out.Districts = decodeJSNamedCodes(arr)
		} else {
			out.Districts = decodeJSNamedCodes(v)
		}

	case *referenceAreasResponse:
		if arr := v.Get("areas"); arr.Type() == js.TypeObject {
			out.Areas = decodeJSNamedCodes(arr)
		} else {
			out.Areas = decodeJSNamedCodes(v)
		}
	}
}

// GoogleMapsGeocoder resolves coordinates on js/wasm using fetch().
type GoogleMapsGeocoder struct {
	APIKey  string
	BaseURL string
}

func NewGoogleMapsGeocoder(apiKey string) *GoogleMapsGeocoder {
	return &GoogleMapsGeocoder{APIKey: apiKey}
}

func (g *GoogleMapsGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	if g.APIKey == "" {
		return 0, 0, fmt.Errorf("google geocoder: API key is empty")
	}
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return 0, 0, fmt.Errorf("google geocoder: query cannot be empty")
	}
	baseURL := g.BaseURL
	if baseURL == "" {
		baseURL = "https://geocode.googleapis.com"
	}
	endpoint := fmt.Sprintf("%s/v4/geocode/address/%s?regionCode=NG", baseURL, url.PathEscape(cleanQuery))
	resp, err := doJSFetch(ctx, "GET", endpoint, map[string]string{
		"X-Goog-Api-Key":   g.APIKey,
		"X-Goog-FieldMask": "results.location,results.formattedAddress,results.placeId",
	}, "")
	if err != nil {
		return 0, 0, err
	}
	if resp.statusCode != 200 {
		return 0, 0, fmt.Errorf("google geocoder v4 returned http status %d", resp.statusCode)
	}
	parsed, pErr := parseJSONSafe(resp.body)
	if pErr != nil {
		return 0, 0, pErr
	}
	results := parsed.Get("results")
	if results.Type() != js.TypeObject || results.Length() == 0 {
		return 0, 0, fmt.Errorf("google geocoder v4: no results found for %q", cleanQuery)
	}
	first := results.Index(0)
	loc := first.Get("location")
	lat := jsFloatField(loc, "lat")
	lng := jsFloatField(loc, "lng")
	if lat == 0 && lng == 0 {
		if geom := first.Get("geometry"); geom.Type() == js.TypeObject {
			gLoc := geom.Get("location")
			lat = jsFloatField(gLoc, "lat")
			lng = jsFloatField(gLoc, "lng")
		}
	}
	if lat == 0 && lng == 0 {
		return 0, 0, fmt.Errorf("google geocoder v4: zero coordinates returned for %q", cleanQuery)
	}
	return lat, lng, nil
}

// NominatimGeocoder resolves coordinates on js/wasm using OpenStreetMap Nominatim.
type NominatimGeocoder struct {
	UserAgent string
	BaseURL   string
}

func NewNominatimGeocoder(userAgent string) *NominatimGeocoder {
	if userAgent == "" {
		userAgent = "PostcodeNigeria-SDK/1.0"
	}
	return &NominatimGeocoder{UserAgent: userAgent}
}

func (n *NominatimGeocoder) Geocode(ctx context.Context, query string) (float64, float64, error) {
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return 0, 0, fmt.Errorf("nominatim geocoder: query cannot be empty")
	}
	baseURL := n.BaseURL
	if baseURL == "" {
		baseURL = "https://nominatim.openstreetmap.org"
	}
	endpoint := fmt.Sprintf("%s/search?format=json&countrycodes=ng&limit=1&q=%s", baseURL, url.QueryEscape(cleanQuery))
	resp, err := doJSFetch(ctx, "GET", endpoint, map[string]string{
		"User-Agent": n.UserAgent,
	}, "")
	if err != nil {
		return 0, 0, err
	}
	if resp.statusCode != 200 {
		return 0, 0, fmt.Errorf("nominatim geocoder returned http status %d", resp.statusCode)
	}
	parsed, pErr := parseJSONSafe(resp.body)
	if pErr != nil {
		return 0, 0, pErr
	}
	if parsed.Type() != js.TypeObject || parsed.Length() == 0 {
		return 0, 0, fmt.Errorf("nominatim geocoder: no results found for %q", cleanQuery)
	}
	first := parsed.Index(0)
	lat, errLat := strconv.ParseFloat(jsStringField(first, "lat"), 64)
	lng, errLng := strconv.ParseFloat(jsStringField(first, "lon"), 64)
	if errLat != nil || errLng != nil {
		return 0, 0, fmt.Errorf("nominatim geocoder parsing coordinates failed")
	}
	return lat, lng, nil
}

func DefaultGeocoder(googleMapsKey string) Geocoder {
	geocoders := make([]Geocoder, 0, 2)
	if googleMapsKey != "" {
		geocoders = append(geocoders, NewGoogleMapsGeocoder(googleMapsKey))
	}
	geocoders = append(geocoders, NewNominatimGeocoder("PostcodeNigeria-SDK/1.0"))
	return NewChainGeocoder(geocoders...)
}
