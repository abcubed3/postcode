//go:build js && wasm

package main

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"sync"
	"syscall/js"

	"github.com/abcubed3/postcode"
)

var (
	clientMu             sync.RWMutex
	currentKey           string
	currentBase          string
	currentGoogleMapsKey string
	client               *postcode.Client

	// Cached global JS constructors to eliminate repeated WASM->JS boundary lookups
	jsGlobal  js.Value
	jsObject  js.Value
	jsArray   js.Value
	jsPromise js.Value
	jsError   js.Value
	jsJSON    js.Value

	// Memoized static reference data and tool schemas
	staticOnce        sync.Once
	cachedStatesMap   js.Value
	cachedRefStates   js.Value
	cachedOpenAITools js.Value
	cachedAnthroTools js.Value
	cachedGeminiTools js.Value
)

func initJSGlobals() {
	jsGlobal = js.Global()
	jsObject = jsGlobal.Get("Object")
	jsArray = jsGlobal.Get("Array")
	jsPromise = jsGlobal.Get("Promise")
	jsError = jsGlobal.Get("Error")
	jsJSON = jsGlobal.Get("JSON")
}

func newJSObject() js.Value {
	return jsObject.New()
}

func newJSArray(n int) js.Value {
	if n <= 0 {
		return jsArray.New()
	}
	return jsArray.New(n)
}

func initStaticCatalogs() {
	staticOnce.Do(func() {
		// 1. listStates() dictionary
		statesObj := newJSObject()
		statesList := make([]postcode.NamedCode, 0, len(postcode.NigerianStates))
		for code, rec := range postcode.NigerianStates {
			st := newJSObject()
			st.Set("code", rec.Code)
			st.Set("name", rec.Name)
			st.Set("capital", rec.Capital)
			st.Set("latitude", rec.Latitude)
			st.Set("longitude", rec.Longitude)
			st.Set("zone", rec.Zone)
			st.Set("Code", rec.Code)
			st.Set("Name", rec.Name)
			st.Set("Capital", rec.Capital)
			st.Set("Latitude", rec.Latitude)
			st.Set("Longitude", rec.Longitude)
			st.Set("Zone", rec.Zone)
			statesObj.Set(code, st)

			statesList = append(statesList, postcode.NamedCode{
				Code: code,
				Name: rec.Name,
			})
		}
		cachedStatesMap = statesObj

		// 2. referenceStatesOffline() sorted array
		slices.SortFunc(statesList, func(a, b postcode.NamedCode) int {
			return cmp.Compare(a.Code, b.Code)
		})
		cachedRefStates = namedCodesToJS(statesList)

		// 3. Pre-built Agent Tool declarations
		tools := postcode.DefaultAgentTools()
		n := len(tools)
		oai := newJSArray(n)
		ant := newJSArray(n)
		gem := newJSArray(n)
		for i, t := range tools {
			oai.SetIndex(i, toJS(t.OpenAITool()))
			ant.SetIndex(i, toJS(t.AnthropicTool()))
			gem.SetIndex(i, toJS(t.GeminiFunctionDeclaration()))
		}
		cachedOpenAITools = oai
		cachedAnthroTools = ant
		cachedGeminiTools = gem
	})
}

func buildClientOptions() []postcode.ClientOption {
	opts := []postcode.ClientOption{
		postcode.WithDefaultCache(),
	}
	if currentKey != "" {
		opts = append(opts, postcode.WithAPIKey(currentKey))
	}
	if currentBase != "" {
		opts = append(opts, postcode.WithBaseURL(currentBase))
	}
	if currentGoogleMapsKey != "" {
		opts = append(opts, postcode.WithGoogleMapsKey(currentGoogleMapsKey))
	}
	return opts
}

func getOrCreateClient() *postcode.Client {
	clientMu.RLock()
	c := client
	clientMu.RUnlock()
	if c != nil {
		return c
	}

	clientMu.Lock()
	defer clientMu.Unlock()
	if client == nil {
		if created, err := postcode.NewClient(buildClientOptions()...); err == nil {
			client = created
		}
	}
	return client
}

func resetClientLocked() {
	if created, err := postcode.NewClient(buildClientOptions()...); err == nil {
		client = created
	}
}

// ============================================================================
// Direct, Zero-Reflection Go -> js.Value Builders
// ============================================================================

func stringsToJS(list []string) js.Value {
	arr := newJSArray(len(list))
	for i, s := range list {
		arr.SetIndex(i, s)
	}
	return arr
}

func floatsToJS(list []float64) js.Value {
	arr := newJSArray(len(list))
	for i, f := range list {
		arr.SetIndex(i, f)
	}
	return arr
}

func namedCodesToJS(list []postcode.NamedCode) js.Value {
	arr := newJSArray(len(list))
	for i, item := range list {
		obj := newJSObject()
		obj.Set("code", item.Code)
		if item.Name != "" {
			obj.Set("name", item.Name)
		}
		arr.SetIndex(i, obj)
	}
	return arr
}

func segmentsToJS(s postcode.Segments) js.Value {
	obj := newJSObject()
	obj.Set("state", s.State)
	obj.Set("lga", s.LGA)
	obj.Set("district", s.District)
	obj.Set("area", s.Area)
	obj.Set("unit", s.Unit)
	return obj
}

func assembledToJS(a postcode.AssembledPostcode) js.Value {
	obj := newJSObject()
	obj.Set("postcode", a.Postcode)
	obj.Set("display", a.Display)
	obj.Set("compact", a.Compact)
	return obj
}

func segmentDiagnosesToJS(list []postcode.SegmentDiagnosis) js.Value {
	arr := newJSArray(len(list))
	for i, d := range list {
		obj := newJSObject()
		obj.Set("segment", d.Segment)
		obj.Set("input", d.Input)
		obj.Set("expected", d.Expected)
		obj.Set("message", d.Message)
		if len(d.Suggestions) > 0 {
			obj.Set("suggestions", stringsToJS(d.Suggestions))
		}
		arr.SetIndex(i, obj)
	}
	return arr
}

func diagnosticReportToJS(r postcode.DiagnosticReport) js.Value {
	obj := newJSObject()
	obj.Set("input", r.Input)
	obj.Set("valid", r.Valid)
	if r.IsTemporary {
		obj.Set("is_temporary", true)
	}
	if r.Normalized != "" {
		obj.Set("normalized", r.Normalized)
	}
	if r.SuggestedFix != "" {
		obj.Set("suggested_fix", r.SuggestedFix)
	}
	obj.Set("clean_length", r.CleanLength)
	if len(r.Diagnoses) > 0 {
		obj.Set("diagnoses", segmentDiagnosesToJS(r.Diagnoses))
	}
	if r.ActionableTip != "" {
		obj.Set("actionable_tip", r.ActionableTip)
	}
	obj.Set("format_score", r.FormatScore)
	return obj
}

func locationToJS(loc postcode.Location) js.Value {
	obj := newJSObject()
	obj.Set("postcode", loc.Postcode)
	obj.Set("compact", loc.Compact)
	obj.Set("latitude", loc.Latitude)
	obj.Set("longitude", loc.Longitude)
	if loc.Address != "" {
		obj.Set("address", loc.Address)
	}
	obj.Set("state_code", loc.StateCode)
	obj.Set("state_name", loc.StateName)
	if loc.LGACode != "" {
		obj.Set("lga_code", loc.LGACode)
	}
	if loc.LGAName != "" {
		obj.Set("lga_name", loc.LGAName)
	}
	if loc.Zone != "" {
		obj.Set("zone", loc.Zone)
	}
	obj.Set("precision", loc.Precision.String())
	return obj
}

func nearbyUnitToJS(u postcode.NearbyUnit) js.Value {
	obj := newJSObject()
	obj.Set("postcode", u.Postcode)
	obj.Set("display", u.Display)
	obj.Set("distance_m", u.DistanceM)
	obj.Set("confidence", u.Confidence)
	if u.StateName != "" {
		obj.Set("state_name", u.StateName)
	}
	if u.LGAName != "" {
		obj.Set("lga_name", u.LGAName)
	}
	if u.LocalityName != "" {
		obj.Set("locality_name", u.LocalityName)
	}
	if u.Address != "" {
		obj.Set("address", u.Address)
	}
	return obj
}

func nearbyUnitsToJS(units []postcode.NearbyUnit) js.Value {
	arr := newJSArray(len(units))
	for i, u := range units {
		arr.SetIndex(i, nearbyUnitToJS(u))
	}
	return arr
}

func reverseResponseToJS(r *postcode.ReverseResponse) js.Value {
	obj := newJSObject()
	if r == nil {
		return obj
	}
	obj.Set("found", r.Found)
	obj.Set("coordinate", floatsToJS(r.Coordinate))
	if r.Unit != nil {
		obj.Set("unit", nearbyUnitToJS(*r.Unit))
		if p, err := postcode.Parse(r.Unit.Postcode); err == nil {
			obj.Set("state_code", p.State())
			if st, ok := postcode.LookupState(p.State()); ok {
				obj.Set("state_name", st.Name)
			}
			obj.Set("lga_code", p.LGA())
			if lga, ok := postcode.LookupLGA(p.State(), p.LGA()); ok {
				obj.Set("lga_name", lga.Name)
			}
			obj.Set("postcode", p.Formatted())
		}
	}
	if r.Area != "" {
		obj.Set("area", r.Area)
	}
	if r.District != "" {
		obj.Set("district", r.District)
	}
	if r.State != "" {
		obj.Set("state", r.State)
	}
	if r.Message != "" {
		obj.Set("message", r.Message)
	}
	obj.Set("radius_m", r.RadiusM)
	return obj
}

func lookupResponseToJS(r *postcode.LookupResponse) js.Value {
	obj := newJSObject()
	if r == nil {
		return obj
	}
	obj.Set("postcode", r.Postcode)
	obj.Set("valid", r.Valid)
	if r.Temporary {
		obj.Set("temporary", true)
	}
	if r.AdministrativeAddress != nil {
		adm := newJSObject()
		a := r.AdministrativeAddress
		if a.State != "" {
			adm.Set("state", a.State)
		}
		if a.StateName != "" {
			adm.Set("state_name", a.StateName)
		}
		if a.LGA != "" {
			adm.Set("lga", a.LGA)
		}
		if a.LGAName != "" {
			adm.Set("lga_name", a.LGAName)
		}
		if a.District != "" {
			adm.Set("district", a.District)
		}
		if a.DistrictName != "" {
			adm.Set("district_name", a.DistrictName)
		}
		if a.Area != "" {
			adm.Set("area", a.Area)
		}
		if a.AreaName != "" {
			adm.Set("area_name", a.AreaName)
		}
		if a.Unit != "" {
			adm.Set("unit", a.Unit)
		}
		if a.Zone != "" {
			adm.Set("zone", a.Zone)
		}
		if a.LocalityName != "" {
			adm.Set("locality_name", a.LocalityName)
		}
		obj.Set("administrative_address", adm)
	}
	if r.RecentHouseAddress != nil {
		rha := newJSObject()
		h := r.RecentHouseAddress
		if h.Address != "" {
			rha.Set("address", h.Address)
		}
		if h.Recent != "" {
			rha.Set("recent", h.Recent)
		}
		if h.HouseNumber != "" {
			rha.Set("house_number", h.HouseNumber)
		}
		if h.StreetName != "" {
			rha.Set("street_name", h.StreetName)
		}
		if h.LocalityName != "" {
			rha.Set("locality_name", h.LocalityName)
		}
		obj.Set("recent_house_address", rha)
	}
	if r.BuildingUseStatus != "" {
		obj.Set("building_use_status", r.BuildingUseStatus)
	}
	if len(r.OtherBuildingInfo) > 0 {
		obj.Set("other_building_info", toJS(r.OtherBuildingInfo))
	}
	if r.PointGeometry != nil {
		pg := newJSObject()
		if r.PointGeometry.Type != "" {
			pg.Set("type", r.PointGeometry.Type)
		}
		pg.Set("coordinates", floatsToJS(r.PointGeometry.Coordinates))
		obj.Set("point_geometry", pg)
	}
	return obj
}

func autocompleteResponseToJS(r *postcode.AutocompleteResponse) js.Value {
	obj := newJSObject()
	if r == nil {
		return obj
	}
	obj.Set("segment", r.Segment)
	sug := newJSArray(len(r.Suggestions))
	for i, s := range r.Suggestions {
		item := newJSObject()
		item.Set("code", s.Code)
		item.Set("label", s.Label)
		sug.SetIndex(i, item)
	}
	obj.Set("suggestions", sug)
	return obj
}

func agentMetricsToJS(m postcode.AgentGuardMetrics) js.Value {
	obj := newJSObject()
	obj.Set("total_calls", m.TotalCalls)
	obj.Set("commercial_calls", m.CommercialCalls)
	obj.Set("downgraded_calls", m.DowngradedCalls)
	obj.Set("offline_fallbacks", m.OfflineFallbacks)
	return obj
}

func syntheticAddressesToJS(list []postcode.SyntheticAddress) js.Value {
	arr := newJSArray(len(list))
	for i, a := range list {
		obj := newJSObject()
		obj.Set("raw_text", a.RawText)
		obj.Set("expected_state", a.ExpectedState)
		obj.Set("expected_state_code", a.ExpectedStateCode)
		if a.ExpectedLGA != "" {
			obj.Set("expected_lga", a.ExpectedLGA)
		}
		if a.ExpectedPostcode != "" {
			obj.Set("expected_postcode", a.ExpectedPostcode)
		}
		if a.ExpectedLat != 0 {
			obj.Set("expected_lat", a.ExpectedLat)
		}
		if a.ExpectedLng != 0 {
			obj.Set("expected_lng", a.ExpectedLng)
		}
		if a.NoiseType != "" {
			obj.Set("noise_type", a.NoiseType)
		}
		arr.SetIndex(i, obj)
	}
	return arr
}

func addressExtractionToJS(e postcode.AddressExtraction) js.Value {
	obj := newJSObject()
	obj.Set("raw", e.Raw)
	if e.Postcode != "" {
		obj.Set("postcode", e.Postcode)
	}
	obj.Set("postcode_valid", e.PostcodeValid)
	if e.LegacyPostcode != "" {
		obj.Set("legacy_postcode", e.LegacyPostcode)
	}
	if e.SuggestedPrefix != "" {
		obj.Set("suggested_prefix", e.SuggestedPrefix)
	}
	if e.StateCode != "" {
		obj.Set("state_code", e.StateCode)
	}
	if e.StateName != "" {
		obj.Set("state_name", e.StateName)
	}
	if e.LGACode != "" {
		obj.Set("lga_code", e.LGACode)
	}
	if e.LGAName != "" {
		obj.Set("lga_name", e.LGAName)
	}
	if e.Locality != "" {
		obj.Set("locality", e.Locality)
	}
	if e.Street != "" {
		obj.Set("street", e.Street)
	}
	if e.Landmark != "" {
		obj.Set("landmark", e.Landmark)
	}
	if e.Latitude != 0 {
		obj.Set("latitude", e.Latitude)
	}
	if e.Longitude != 0 {
		obj.Set("longitude", e.Longitude)
	}
	obj.Set("confidence", e.Confidence)
	return obj
}

func stateRecordsToJS(list []postcode.StateRecord) js.Value {
	arr := newJSArray(len(list))
	for i, rec := range list {
		st := newJSObject()
		st.Set("code", rec.Code)
		st.Set("name", rec.Name)
		st.Set("capital", rec.Capital)
		st.Set("latitude", rec.Latitude)
		st.Set("longitude", rec.Longitude)
		st.Set("zone", rec.Zone)
		st.Set("Code", rec.Code)
		st.Set("Name", rec.Name)
		st.Set("Capital", rec.Capital)
		st.Set("Latitude", rec.Latitude)
		st.Set("Longitude", rec.Longitude)
		st.Set("Zone", rec.Zone)
		arr.SetIndex(i, st)
	}
	return arr
}

func toJS(v any) js.Value {
	if v == nil {
		return js.Null()
	}
	switch val := v.(type) {
	case js.Value:
		return val
	case string:
		return js.ValueOf(val)
	case bool:
		return js.ValueOf(val)
	case int:
		return js.ValueOf(val)
	case int64:
		return js.ValueOf(val)
	case float64:
		return js.ValueOf(val)
	case postcode.QRPayloadType:
		return js.ValueOf(string(val))
	case postcode.QRErrorCorrectionLevel:
		return js.ValueOf(string(val))
	case []string:
		return stringsToJS(val)
	case []float64:
		return floatsToJS(val)
	case []any:
		arr := newJSArray(len(val))
		for i, elem := range val {
			arr.SetIndex(i, toJS(elem))
		}
		return arr
	case map[string]any:
		obj := newJSObject()
		for k, elem := range val {
			obj.Set(k, toJS(elem))
		}
		return obj
	case postcode.DiagnosticReport:
		return diagnosticReportToJS(val)
	case []postcode.SegmentDiagnosis:
		return segmentDiagnosesToJS(val)
	case postcode.Segments:
		return segmentsToJS(val)
	case *postcode.Segments:
		if val == nil {
			return js.Null()
		}
		return segmentsToJS(*val)
	case postcode.AssembledPostcode:
		return assembledToJS(val)
	case *postcode.AssembledPostcode:
		if val == nil {
			return js.Null()
		}
		return assembledToJS(*val)
	case postcode.Location:
		return locationToJS(val)
	case *postcode.Location:
		if val == nil {
			return js.Null()
		}
		return locationToJS(*val)
	case []postcode.NamedCode:
		return namedCodesToJS(val)
	case []postcode.NearbyUnit:
		return nearbyUnitsToJS(val)
	case *postcode.NearbyResponse:
		if val == nil {
			return js.Null()
		}
		obj := newJSObject()
		obj.Set("results", nearbyUnitsToJS(val.Results))
		return obj
	case *postcode.ReverseResponse:
		return reverseResponseToJS(val)
	case *postcode.LookupResponse:
		return lookupResponseToJS(val)
	case *postcode.AutocompleteResponse:
		return autocompleteResponseToJS(val)
	case postcode.AgentGuardMetrics:
		return agentMetricsToJS(val)
	case []postcode.SyntheticAddress:
		return syntheticAddressesToJS(val)
	case postcode.AddressExtraction:
		return addressExtractionToJS(val)
	case []postcode.StateRecord:
		return stateRecordsToJS(val)
	default:
		return newJSObject()
	}
}

func jsPromiseReject(msg string) js.Value {
	var rejectHandler js.Func
	rejectHandler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		defer rejectHandler.Release()
		errObj := jsError.New(msg)
		promiseArgs[1].Invoke(errObj)
		return nil
	})
	return jsPromise.New(rejectHandler)
}

func validateSingle(code string) js.Value {
	p, err := postcode.Parse(code)
	if err != nil {
		diag := postcode.Diagnose(code)
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("input", code)
		obj.Set("error", err.Error())
		obj.Set("actionable_tip", diag.ActionableTip)
		obj.Set("clean_length", diag.CleanLength)
		if diag.IsTemporary {
			obj.Set("is_temporary", true)
		}
		return obj
	}

	stateName := ""
	if rec, ok := postcode.NigerianStates[p.State()]; ok {
		stateName = rec.Name
	}

	obj := newJSObject()
	obj.Set("valid", true)
	obj.Set("postcode", p.Formatted())
	obj.Set("compact", p.Compact())
	obj.Set("state_code", p.State())
	obj.Set("state_name", stateName)
	obj.Set("lga_code", p.LGA())
	obj.Set("district", p.District())
	obj.Set("area", p.Area())
	obj.Set("unit", p.BuildingUnit())
	return obj
}

func jsValidate(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}
	return validateSingle(args[0].String())
}

func jsValidateBatch(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return newJSArray(0)
	}
	arr := args[0]
	n := arr.Length()
	results := newJSArray(n)
	for i := 0; i < n; i++ {
		code := arr.Index(i).String()
		results.SetIndex(i, validateSingle(code))
	}
	return results
}

func jsDiagnose(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("error", "code parameter is required")
		return obj
	}
	return diagnosticReportToJS(postcode.Diagnose(args[0].String()))
}

func jsIsTemporary(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return false
	}
	return postcode.IsTemporary(args[0].String())
}

func jsFormatTemporary(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return ""
	}
	return postcode.FormatTemporary(args[0].String())
}

func jsParse(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}

	raw := args[0].String()
	p, err := postcode.Parse(raw)
	if err != nil {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", err.Error())
		return obj
	}

	loc := p.Location()
	capital := ""
	if st, ok := postcode.NigerianStates[p.State()]; ok {
		capital = st.Capital
	}

	obj := newJSObject()
	obj.Set("valid", true)
	obj.Set("input", raw)
	obj.Set("postcode", p.Formatted())
	obj.Set("formatted", p.Formatted())
	obj.Set("compact", p.Compact())
	obj.Set("spaced", p.String())
	obj.Set("state_code", p.State())
	obj.Set("state_name", loc.StateName)
	obj.Set("lga_code", p.LGA())
	obj.Set("lga_name", loc.LGAName)
	obj.Set("district", p.District())
	obj.Set("area", p.Area())
	obj.Set("unit", p.BuildingUnit())
	obj.Set("zone", loc.Zone)
	obj.Set("state_capital", capital)
	return obj
}

func jsFormat(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return ""
	}
	code := args[0].String()
	style := "canonical"
	if len(args) > 1 && args[1].Type() == js.TypeString {
		style = args[1].String()
	}

	p, err := postcode.Parse(code)
	if err != nil {
		return ""
	}

	switch style {
	case "compact":
		return p.Compact()
	case "spaced":
		return p.String()
	case "hyphenated":
		return p.Formatted()
	case "slug":
		return strings.ToLower(p.Formatted())
	case "canonical":
		fallthrough
	default:
		return p.Formatted()
	}
}

func jsAssemble(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", "segments object is required with state, lga, district, area, unit")
		return obj
	}
	obj := args[0]
	segs := postcode.Segments{
		State:    obj.Get("state").String(),
		LGA:      obj.Get("lga").String(),
		District: obj.Get("district").String(),
		Area:     obj.Get("area").String(),
		Unit:     obj.Get("unit").String(),
	}

	assembled, err := postcode.AssembleSegments(segs)
	if err != nil {
		res := newJSObject()
		res.Set("valid", false)
		res.Set("error", err.Error())
		return res
	}

	res := newJSObject()
	res.Set("valid", true)
	res.Set("postcode", assembled.Postcode)
	res.Set("display", assembled.Display)
	res.Set("compact", assembled.Compact)
	return res
}

func jsNormalizeSegments(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		obj := newJSObject()
		obj.Set("error", "segments object is required with state, lga, district, area, unit")
		return obj
	}
	obj := args[0]
	segs := postcode.Segments{
		State:    obj.Get("state").String(),
		LGA:      obj.Get("lga").String(),
		District: obj.Get("district").String(),
		Area:     obj.Get("area").String(),
		Unit:     obj.Get("unit").String(),
	}
	return segmentsToJS(postcode.NormalizeSegments(segs))
}

func jsAssembleOnline(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return jsPromiseReject("segments object is required with state, lga, district, area, unit")
	}
	obj := args[0]
	segs := postcode.Segments{
		State:    obj.Get("state").String(),
		LGA:      obj.Get("lga").String(),
		District: obj.Get("district").String(),
		Area:     obj.Get("area").String(),
		Unit:     obj.Get("unit").String(),
	}

	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			assembled, err := c.Assemble(context.Background(), segs)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			res := newJSObject()
			res.Set("valid", true)
			res.Set("postcode", assembled.Postcode)
			res.Set("display", assembled.Display)
			res.Set("compact", assembled.Compact)
			resolve.Invoke(res)
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsDisassembleOnline(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("code parameter is required")
	}
	code := args[0].String()
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			segs, err := c.Disassemble(context.Background(), code)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			resolve.Invoke(segmentsToJS(*segs))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsDisassemble(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}
	code := args[0].String()
	p, err := postcode.Parse(code)
	if err != nil {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", err.Error())
		return obj
	}

	obj := newJSObject()
	obj.Set("valid", true)
	obj.Set("state", p.State())
	obj.Set("lga", p.LGA())
	obj.Set("district", p.District())
	obj.Set("area", p.Area())
	obj.Set("unit", p.BuildingUnit())
	return obj
}

func enrichLocationJS(loc postcode.Location) js.Value {
	jsLoc := locationToJS(loc)
	jsLoc.Set("state", loc.StateName)
	jsLoc.Set("lga", loc.LGAName)
	jsLoc.Set("google_maps_url", loc.GoogleMapsURL())
	jsLoc.Set("google_maps_directions_url", loc.GoogleMapsDirectionsURL())
	jsLoc.Set("apple_maps_url", loc.AppleMapsURL())
	jsLoc.Set("osm_url", loc.OpenStreetMapURL())
	jsLoc.Set("search_query", loc.SearchQuery())
	return jsLoc
}

func jsResolveLocation(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("error", "code parameter is required")
		return obj
	}

	p, err := postcode.Parse(args[0].String())
	if err != nil {
		obj := newJSObject()
		obj.Set("error", err.Error())
		return obj
	}

	return enrichLocationJS(p.Location())
}

func jsResolveLocationOnline(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("code parameter is required")
	}

	code := args[0].String()
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()

			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}

			loc, err := c.ResolveLocation(context.Background(), code)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}

			resolve.Invoke(enrichLocationJS(*loc))
		}()
		return nil
	})

	return jsPromise.New(handler)
}

func parseJSBuildingRecord(obj js.Value) postcode.BuildingRecord {
	rec := postcode.BuildingRecord{
		Postcode:  obj.Get("postcode").String(),
		Latitude:  obj.Get("latitude").Float(),
		Longitude: obj.Get("longitude").Float(),
	}
	if v := obj.Get("address"); v.Type() == js.TypeString {
		rec.Address = v.String()
	}
	if v := obj.Get("state_code"); v.Type() == js.TypeString {
		rec.StateCode = v.String()
	}
	if v := obj.Get("state_name"); v.Type() == js.TypeString {
		rec.StateName = v.String()
	}
	if v := obj.Get("lga_code"); v.Type() == js.TypeString {
		rec.LGACode = v.String()
	}
	if v := obj.Get("lga_name"); v.Type() == js.TypeString {
		rec.LGAName = v.String()
	}
	if v := obj.Get("zone"); v.Type() == js.TypeString {
		rec.Zone = v.String()
	}
	return rec
}

func jsRegisterBuilding(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return false
	}
	postcode.RegisterKnownBuilding(parseJSBuildingRecord(args[0]))
	return true
}

func jsRegisterBuildings(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return 0
	}
	arr := args[0]
	n := arr.Length()
	count := 0
	for i := 0; i < n; i++ {
		item := arr.Index(i)
		if item.Type() == js.TypeObject {
			postcode.RegisterKnownBuilding(parseJSBuildingRecord(item))
			count++
		}
	}
	return count
}

func jsListStates(this js.Value, args []js.Value) any {
	initStaticCatalogs()
	return cachedStatesMap
}

func jsReferenceStatesOffline(this js.Value, args []js.Value) any {
	initStaticCatalogs()
	return cachedRefStates
}

func jsReferenceStates(this js.Value, args []js.Value) any {
	if len(args) > 0 && args[0].Type() == js.TypeBoolean && !args[0].Bool() {
		return jsReferenceStatesOffline(this, args)
	}
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c != nil {
				if states, err := c.ReferenceStates(context.Background()); err == nil {
					resolve.Invoke(namedCodesToJS(states))
					return
				}
			}
			initStaticCatalogs()
			resolve.Invoke(cachedRefStates)
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsReferenceLGAsOffline(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return newJSArray(0)
	}
	return namedCodesToJS(postcode.StateLGAs(args[0].String()))
}

func jsReferenceLGAs(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("state parameter is required")
	}
	state := args[0].String()
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c != nil {
				if lgas, err := c.ReferenceLGAs(context.Background(), state); err == nil {
					resolve.Invoke(namedCodesToJS(lgas))
					return
				}
			}
			resolve.Invoke(namedCodesToJS(postcode.StateLGAs(state)))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsReferenceDistricts(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return jsPromiseReject("state and lga parameters are required")
	}
	state := args[0].String()
	lga := args[1].String()
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			districts, err := c.ReferenceDistricts(context.Background(), state, lga)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			resolve.Invoke(namedCodesToJS(districts))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsReferenceAreas(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return jsPromiseReject("state, lga, and district parameters are required")
	}
	state := args[0].String()
	lga := args[1].String()
	district := args[2].String()
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			areas, err := c.ReferenceAreas(context.Background(), state, lga, district)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			resolve.Invoke(namedCodesToJS(areas))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsSearchNearbyBuildingsOffline(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return newJSArray(0)
	}
	lat := args[0].Float()
	lng := args[1].Float()
	radiusM := 300.0
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		radiusM = args[2].Float()
	}
	return nearbyUnitsToJS(postcode.SearchNearbyBuildingsOffline(lat, lng, radiusM))
}

func jsReverseCoordinatesOffline(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		obj := newJSObject()
		obj.Set("found", false)
		obj.Set("error", "latitude and longitude arguments are required")
		return obj
	}
	lat := args[0].Float()
	lng := args[1].Float()
	maxDist := 25.0
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		maxDist = args[2].Float()
	}
	return reverseResponseToJS(postcode.ReverseCoordinatesOffline(lat, lng, maxDist))
}

func jsNearestLGA(this js.Value, args []js.Value) any {
	obj := newJSObject()
	if len(args) < 2 {
		obj.Set("found", false)
		obj.Set("error", "latitude and longitude arguments are required")
		return obj
	}
	lat := args[0].Float()
	lng := args[1].Float()
	rec, distM, found := postcode.NearestLGA(lat, lng)
	obj.Set("found", found)
	if !found {
		return obj
	}
	obj.Set("state_code", rec.StateCode)
	if st, ok := postcode.LookupState(rec.StateCode); ok {
		obj.Set("state_name", st.Name)
	}
	obj.Set("lga_code", rec.LGACode)
	obj.Set("lga_name", rec.Name)
	obj.Set("latitude", rec.Latitude)
	obj.Set("longitude", rec.Longitude)
	obj.Set("distance_meters", distM)
	obj.Set("postcode_prefix", rec.StateCode+"-"+rec.LGACode)
	return obj
}

func jsHealth(this js.Value, args []js.Value) any {
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			if err := c.Health(context.Background()); err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			res := newJSObject()
			res.Set("status", "ok")
			resolve.Invoke(res)
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsSetAPIKey(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return false
	}
	key := args[0].String()
	clientMu.Lock()
	currentKey = key
	resetClientLocked()
	clientMu.Unlock()
	return true
}

func jsGetAPIKey(this js.Value, args []js.Value) any {
	clientMu.RLock()
	defer clientMu.RUnlock()
	return currentKey
}

func jsSetGoogleMapsAPIKey(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return false
	}
	key := args[0].String()
	clientMu.Lock()
	currentGoogleMapsKey = key
	resetClientLocked()
	clientMu.Unlock()
	return true
}

func jsGetGoogleMapsAPIKey(this js.Value, args []js.Value) any {
	clientMu.RLock()
	defer clientMu.RUnlock()
	return currentGoogleMapsKey
}

func jsConfigure(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return false
	}
	optObj := args[0]
	clientMu.Lock()
	defer clientMu.Unlock()
	if v := optObj.Get("apiKey"); v.Type() == js.TypeString {
		currentKey = v.String()
	}
	if v := optObj.Get("baseURL"); v.Type() == js.TypeString {
		currentBase = v.String()
	}
	if v := optObj.Get("googleMapsApiKey"); v.Type() == js.TypeString {
		currentGoogleMapsKey = v.String()
	} else if v := optObj.Get("googleMapsKey"); v.Type() == js.TypeString {
		currentGoogleMapsKey = v.String()
	}
	resetClientLocked()
	return true
}

func jsLookup(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("code parameter is required")
	}

	code := args[0].String()
	level := 1
	if len(args) > 1 && args[1].Type() == js.TypeNumber {
		level = args[1].Int()
	}

	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()

			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}

			resp, err := c.Lookup(context.Background(), code, postcode.LookupLevel(level))
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}

			resolve.Invoke(lookupResponseToJS(resp))
		}()
		return nil
	})

	return jsPromise.New(handler)
}

func jsAutocomplete(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("query parameter is required")
	}
	query := args[0].String()

	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			resp, err := c.Autocomplete(context.Background(), query)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			resolve.Invoke(autocompleteResponseToJS(resp))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsNearby(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("parameters required: specify a postcode string or an options object with {postcode} or {latitude, longitude}")
	}

	var targetCode string
	var lat, lng float64
	radiusM := 300.0
	limit := 10

	if args[0].Type() == js.TypeString {
		targetCode = args[0].String()
		p, err := postcode.Parse(targetCode)
		if err != nil {
			return jsPromiseReject("invalid reference postcode: " + err.Error())
		}
		loc := p.Location()
		lat = loc.Latitude
		lng = loc.Longitude
		if len(args) > 1 && args[1].Type() == js.TypeNumber {
			radiusM = args[1].Float()
		}
	} else if len(args) >= 2 && args[0].Type() == js.TypeNumber && args[1].Type() == js.TypeNumber {
		lat = args[0].Float()
		lng = args[1].Float()
		if len(args) > 2 && args[2].Type() == js.TypeNumber {
			radiusM = args[2].Float()
		}
	} else if args[0].Type() == js.TypeObject {
		paramsObj := args[0]
		if v := paramsObj.Get("postcode"); v.Type() == js.TypeString {
			targetCode = v.String()
		} else if v := paramsObj.Get("code"); v.Type() == js.TypeString {
			targetCode = v.String()
		}

		if v := paramsObj.Get("latitude"); v.Type() == js.TypeNumber {
			lat = v.Float()
		}
		if v := paramsObj.Get("longitude"); v.Type() == js.TypeNumber {
			lng = v.Float()
		}

		if lat == 0 && lng == 0 && targetCode != "" {
			p, err := postcode.Parse(targetCode)
			if err != nil {
				return jsPromiseReject("invalid reference postcode: " + err.Error())
			}
			loc := p.Location()
			lat = loc.Latitude
			lng = loc.Longitude
		}

		if v := paramsObj.Get("radius_m"); v.Type() == js.TypeNumber {
			radiusM = v.Float()
		} else if v := paramsObj.Get("radiusKm"); v.Type() == js.TypeNumber {
			radiusM = v.Float() * 1000.0
		} else if v := paramsObj.Get("radius_km"); v.Type() == js.TypeNumber {
			radiusM = v.Float() * 1000.0
		} else if v := paramsObj.Get("radius"); v.Type() == js.TypeNumber {
			radiusM = v.Float()
		}

		if v := paramsObj.Get("limit"); v.Type() == js.TypeNumber {
			limit = v.Int()
		}
	}

	if lat == 0 && lng == 0 {
		return jsPromiseReject("specify a reference postcode or latitude/longitude coordinates")
	}

	params := postcode.NearbyParams{
		Postcode:  targetCode,
		Latitude:  lat,
		Longitude: lng,
		RadiusM:   radiusM,
	}

	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			resp, err := c.Nearby(context.Background(), params)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			if limit > 0 && len(resp.Results) > limit {
				resp.Results = resp.Results[:limit]
			}
			resolve.Invoke(toJS(resp))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsReverseGeocode(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("parameters required: specify lat and lng as numbers, or an options object with {latitude, longitude}")
	}

	var lat, lng float64
	maxDist := 25.0

	if len(args) >= 2 && args[0].Type() == js.TypeNumber && args[1].Type() == js.TypeNumber {
		lat = args[0].Float()
		lng = args[1].Float()
		if len(args) > 2 && args[2].Type() == js.TypeNumber {
			maxDist = args[2].Float()
		}
	} else if args[0].Type() == js.TypeObject {
		paramsObj := args[0]
		lat = paramsObj.Get("latitude").Float()
		lng = paramsObj.Get("longitude").Float()
		if v := paramsObj.Get("max_distance_m"); v.Type() == js.TypeNumber {
			maxDist = v.Float()
		} else if v := paramsObj.Get("maxDistanceKm"); v.Type() == js.TypeNumber {
			maxDist = v.Float() * 1000.0
		} else if v := paramsObj.Get("maxDistanceM"); v.Type() == js.TypeNumber {
			maxDist = v.Float()
		} else if v := paramsObj.Get("max_dist"); v.Type() == js.TypeNumber {
			maxDist = v.Float()
		}
	}

	if lat == 0 && lng == 0 {
		return jsPromiseReject("latitude and longitude coordinates are required")
	}

	params := postcode.ReverseParams{
		Latitude:     lat,
		Longitude:    lng,
		MaxDistanceM: maxDist,
	}

	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				reject.Invoke(jsError.New("failed to initialize NIPOST client"))
				return
			}
			resp, err := c.Reverse(context.Background(), params)
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			resolve.Invoke(reverseResponseToJS(resp))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsGetAgentTools(this js.Value, args []js.Value) any {
	initStaticCatalogs()
	format := "openai"
	if len(args) > 0 && args[0].Type() == js.TypeString {
		format = args[0].String()
	}
	switch format {
	case "anthropic":
		return cachedAnthroTools
	case "gemini":
		return cachedGeminiTools
	default:
		return cachedOpenAITools
	}
}

func jsExecuteTool(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("tool name parameter is required")
	}
	name := args[0].String()
	var jsArgsObj js.Value
	var rawArgsJSON string
	hasDirectObj := false

	if len(args) > 1 {
		if args[1].Type() == js.TypeObject {
			jsArgsObj = args[1]
			hasDirectObj = true
		} else if args[1].Type() == js.TypeString {
			rawArgsJSON = args[1].String()
		}
	}
	if !hasDirectObj && rawArgsJSON == "" {
		rawArgsJSON = "{}"
	}

	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			dispatcher := postcode.NewAgentDispatcher(c)
			var res any
			var err error
			if hasDirectObj {
				res, err = dispatcher.DispatchJSValue(context.Background(), name, jsArgsObj)
			} else {
				res, err = dispatcher.Dispatch(context.Background(), name, []byte(rawArgsJSON))
			}
			if err != nil {
				reject.Invoke(jsError.New(err.Error()))
				return
			}
			resolve.Invoke(toJS(res))
		}()
		return nil
	})
	return jsPromise.New(handler)
}

func jsGetAgentMetrics(this js.Value, args []js.Value) any {
	c := getOrCreateClient()
	if c == nil {
		return agentMetricsToJS(postcode.AgentGuardMetrics{})
	}
	return agentMetricsToJS(c.AgentMetrics())
}

func jsGenerateSyntheticAddresses(this js.Value, args []js.Value) any {
	opts := postcode.GeneratorOptions{Count: 10, NoiseRate: 0.3}
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		obj := args[0]
		if v := obj.Get("count"); v.Type() == js.TypeNumber {
			opts.Count = v.Int()
		}
		if v := obj.Get("noiseRate"); v.Type() == js.TypeNumber {
			opts.NoiseRate = v.Float()
		} else if v := obj.Get("noise_rate"); v.Type() == js.TypeNumber {
			opts.NoiseRate = v.Float()
		}
		if v := obj.Get("seed"); v.Type() == js.TypeNumber {
			opts.Seed = uint64(v.Int())
		}
	}
	return syntheticAddressesToJS(postcode.GenerateSyntheticDataset(opts))
}

func jsRepair(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("repaired", false)
		obj.Set("error", "code parameter is required")
		return obj
	}
	raw := args[0].String()
	p, ok := postcode.TryRepair(raw)
	if !ok {
		report := postcode.Diagnose(raw)
		obj := newJSObject()
		obj.Set("repaired", false)
		obj.Set("input", raw)
		obj.Set("actionable_tip", report.ActionableTip)
		obj.Set("diagnoses", segmentDiagnosesToJS(report.Diagnoses))
		return obj
	}
	obj := newJSObject()
	obj.Set("repaired", true)
	obj.Set("input", raw)
	obj.Set("postcode", p.Formatted())
	obj.Set("display", p.String())
	obj.Set("compact", p.Compact())
	obj.Set("slug", p.Slug())
	return obj
}

func jsExtractAddress(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return addressExtractionToJS(postcode.AddressExtraction{Confidence: "low"})
	}
	return addressExtractionToJS(postcode.ExtractAddress(args[0].String()))
}

func jsDistanceMeters(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", "two postcode arguments (codeA, codeB) are required")
		return obj
	}
	codeA := args[0].String()
	codeB := args[1].String()
	distM, err := postcode.DistanceMeters(codeA, codeB)
	if err != nil {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", err.Error())
		return obj
	}
	locA, _ := postcode.ResolveLocation(codeA)
	locB, _ := postcode.ResolveLocation(codeB)
	fromPostcode := codeA
	if locA.Postcode != "" {
		fromPostcode = locA.Postcode
	}
	toPostcode := codeB
	if locB.Postcode != "" {
		toPostcode = locB.Postcode
	}
	obj := newJSObject()
	obj.Set("valid", true)
	obj.Set("distance_m", distM)
	obj.Set("distance_km", distM/1000.0)
	obj.Set("from_postcode", fromPostcode)
	obj.Set("to_postcode", toPostcode)
	return obj
}

func jsGenerateQR(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}
	code := args[0].String()
	var opts postcode.QROptions
	if len(args) > 1 {
		if args[1].Type() == js.TypeString {
			opts.Payload = postcode.QRPayloadType(args[1].String())
		} else if args[1].Type() == js.TypeObject {
			obj := args[1]
			if v := obj.Get("payload"); v.Type() == js.TypeString {
				opts.Payload = postcode.QRPayloadType(v.String())
			}
			if v := obj.Get("level"); v.Type() == js.TypeString {
				opts.Level = postcode.QRErrorCorrectionLevel(v.String())
			}
			if v := obj.Get("size"); v.Type() == js.TypeNumber {
				opts.Size = v.Int()
			}
			if v := obj.Get("quietZone"); v.Type() == js.TypeNumber {
				opts.QuietZone = v.Int()
			} else if v := obj.Get("quiet_zone"); v.Type() == js.TypeNumber {
				opts.QuietZone = v.Int()
			}
			if v := obj.Get("foregroundColor"); v.Type() == js.TypeString {
				opts.ForegroundColor = v.String()
			} else if v := obj.Get("foreground_color"); v.Type() == js.TypeString {
				opts.ForegroundColor = v.String()
			}
			if v := obj.Get("backgroundColor"); v.Type() == js.TypeString {
				opts.BackgroundColor = v.String()
			} else if v := obj.Get("background_color"); v.Type() == js.TypeString {
				opts.BackgroundColor = v.String()
			}
			if v := obj.Get("includeLabel"); v.Type() == js.TypeBoolean {
				opts.IncludeLabel = v.Bool()
			} else if v := obj.Get("include_label"); v.Type() == js.TypeBoolean {
				opts.IncludeLabel = v.Bool()
			}
			if v := obj.Get("title"); v.Type() == js.TypeString {
				opts.Title = v.String()
			}
			if v := obj.Get("subtitle"); v.Type() == js.TypeString {
				opts.Subtitle = v.String()
			}
			if v := obj.Get("invertAscii"); v.Type() == js.TypeBoolean {
				opts.InvertASCII = v.Bool()
			} else if v := obj.Get("invert_ascii"); v.Type() == js.TypeBoolean {
				opts.InvertASCII = v.Bool()
			}
		}
	}

	qr, err := postcode.GenerateQR(code, opts)
	if err != nil {
		obj := newJSObject()
		obj.Set("valid", false)
		obj.Set("input", code)
		obj.Set("error", err.Error())
		return obj
	}

	dataURI, _ := qr.DataURI()
	obj := newJSObject()
	obj.Set("valid", true)
	obj.Set("postcode", qr.Postcode)
	obj.Set("compact", qr.Compact)
	obj.Set("state_code", qr.StateCode)
	obj.Set("state_name", qr.StateName)
	obj.Set("lga_code", qr.LGACode)
	obj.Set("lga_name", qr.LGAName)
	obj.Set("latitude", qr.Latitude)
	obj.Set("longitude", qr.Longitude)
	obj.Set("google_maps_url", qr.GoogleMapsURL)
	obj.Set("payload", qr.Payload)
	obj.Set("payload_type", string(qr.PayloadType))
	obj.Set("level", string(qr.Level))
	obj.Set("version", qr.Version)
	obj.Set("dimension", qr.Dimension)
	obj.Set("include_label", qr.IncludeLabel)
	obj.Set("svg", qr.SVG())
	obj.Set("matrix_svg", qr.MatrixSVG())
	obj.Set("label_svg", qr.LabelSVG())
	obj.Set("printable_html", qr.PrintableHTML())
	obj.Set("data_uri", dataURI)
	obj.Set("svg_data_uri", qr.SVGDataURI())
	obj.Set("terminal", qr.Terminal())
	obj.Set("ascii", qr.ASCII())
	return obj
}

func main() {
	initJSGlobals()
	postcodeObj := newJSObject()

	// 1. Core Offline (Sync)
	postcodeObj.Set("validate", js.FuncOf(jsValidate))
	postcodeObj.Set("validateBatch", js.FuncOf(jsValidateBatch))
	postcodeObj.Set("diagnose", js.FuncOf(jsDiagnose))
	postcodeObj.Set("isTemporary", js.FuncOf(jsIsTemporary))
	postcodeObj.Set("formatTemporary", js.FuncOf(jsFormatTemporary))
	postcodeObj.Set("repair", js.FuncOf(jsRepair))
	postcodeObj.Set("extractAddress", js.FuncOf(jsExtractAddress))
	postcodeObj.Set("distanceMeters", js.FuncOf(jsDistanceMeters))
	postcodeObj.Set("generateQR", js.FuncOf(jsGenerateQR))
	postcodeObj.Set("qr", js.FuncOf(jsGenerateQR))
	postcodeObj.Set("parse", js.FuncOf(jsParse))
	postcodeObj.Set("format", js.FuncOf(jsFormat))
	postcodeObj.Set("assemble", js.FuncOf(jsAssemble))
	postcodeObj.Set("normalizeSegments", js.FuncOf(jsNormalizeSegments))
	postcodeObj.Set("disassemble", js.FuncOf(jsDisassemble))
	postcodeObj.Set("resolveLocation", js.FuncOf(jsResolveLocation))
	postcodeObj.Set("registerBuilding", js.FuncOf(jsRegisterBuilding))
	postcodeObj.Set("registerBuildings", js.FuncOf(jsRegisterBuildings))
	postcodeObj.Set("listStates", js.FuncOf(jsListStates))
	postcodeObj.Set("referenceStatesOffline", js.FuncOf(jsReferenceStatesOffline))
	postcodeObj.Set("referenceLGAsOffline", js.FuncOf(jsReferenceLGAsOffline))
	postcodeObj.Set("stateLGAs", js.FuncOf(jsReferenceLGAsOffline))
	postcodeObj.Set("searchNearbyBuildingsOffline", js.FuncOf(jsSearchNearbyBuildingsOffline))
	postcodeObj.Set("reverseCoordinatesOffline", js.FuncOf(jsReverseCoordinatesOffline))
	postcodeObj.Set("nearestLGA", js.FuncOf(jsNearestLGA))

	// 2. Configuration & State (Sync)
	postcodeObj.Set("setAPIKey", js.FuncOf(jsSetAPIKey))
	postcodeObj.Set("getAPIKey", js.FuncOf(jsGetAPIKey))
	postcodeObj.Set("setGoogleMapsAPIKey", js.FuncOf(jsSetGoogleMapsAPIKey))
	postcodeObj.Set("getGoogleMapsAPIKey", js.FuncOf(jsGetGoogleMapsAPIKey))
	postcodeObj.Set("configure", js.FuncOf(jsConfigure))

	// 3. Online Gateway API & Live Catalogs (Async Promises)
	postcodeObj.Set("resolveLocationOnline", js.FuncOf(jsResolveLocationOnline))
	postcodeObj.Set("lookup", js.FuncOf(jsLookup))
	postcodeObj.Set("autocomplete", js.FuncOf(jsAutocomplete))
	postcodeObj.Set("nearby", js.FuncOf(jsNearby))
	postcodeObj.Set("reverseGeocode", js.FuncOf(jsReverseGeocode))
	postcodeObj.Set("assembleOnline", js.FuncOf(jsAssembleOnline))
	postcodeObj.Set("disassembleOnline", js.FuncOf(jsDisassembleOnline))
	postcodeObj.Set("referenceStates", js.FuncOf(jsReferenceStates))
	postcodeObj.Set("referenceLGAs", js.FuncOf(jsReferenceLGAs))
	postcodeObj.Set("referenceDistricts", js.FuncOf(jsReferenceDistricts))
	postcodeObj.Set("referenceAreas", js.FuncOf(jsReferenceAreas))
	postcodeObj.Set("health", js.FuncOf(jsHealth))

	// 4. AI Agent Tooling & Guardrails (LLM Protocol Bridge)
	postcodeObj.Set("getAgentTools", js.FuncOf(jsGetAgentTools))
	postcodeObj.Set("executeTool", js.FuncOf(jsExecuteTool))
	postcodeObj.Set("getAgentMetrics", js.FuncOf(jsGetAgentMetrics))

	// 5. Synthetic Address Generator & Evaluation Benchmark
	postcodeObj.Set("generateSyntheticAddresses", js.FuncOf(jsGenerateSyntheticAddresses))

	jsGlobal.Set("Postcode", postcodeObj)

	// Keep WebAssembly event loop running
	<-make(chan struct{})
}
