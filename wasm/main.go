//go:build js && wasm

package main

import (
	"cmp"
	"context"
	"encoding/json"
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
)

func getOrCreateClient() *postcode.Client {
	clientMu.Lock()
	defer clientMu.Unlock()
	if client == nil {
		var opts []postcode.ClientOption
		if currentKey != "" {
			opts = append(opts, postcode.WithAPIKey(currentKey))
		}
		if currentBase != "" {
			opts = append(opts, postcode.WithBaseURL(currentBase))
		}
		if currentGoogleMapsKey != "" {
			opts = append(opts, postcode.WithGoogleMapsKey(currentGoogleMapsKey))
		}
		c, err := postcode.NewClient(opts...)
		if err == nil {
			client = c
		}
	}
	return client
}

func resetClientLocked() {
	var opts []postcode.ClientOption
	if currentKey != "" {
		opts = append(opts, postcode.WithAPIKey(currentKey))
	}
	if currentBase != "" {
		opts = append(opts, postcode.WithBaseURL(currentBase))
	}
	if currentGoogleMapsKey != "" {
		opts = append(opts, postcode.WithGoogleMapsKey(currentGoogleMapsKey))
	}
	c, err := postcode.NewClient(opts...)
	if err == nil {
		client = c
	}
}

func toJS(v any) js.Value {
	b, err := json.Marshal(v)
	if err != nil {
		obj := js.Global().Get("Object").New()
		obj.Set("error", err.Error())
		return obj
	}
	return js.Global().Get("JSON").Call("parse", string(b))
}

func jsPromiseReject(msg string) js.Value {
	promiseConstructor := js.Global().Get("Promise")
	var rejectHandler js.Func
	rejectHandler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		defer rejectHandler.Release()
		errObj := js.Global().Get("Error").New(msg)
		promiseArgs[1].Invoke(errObj)
		return nil
	})
	return promiseConstructor.New(rejectHandler)
}

func validateSingle(code string) js.Value {
	p, err := postcode.Parse(code)
	if err != nil {
		diag := postcode.Diagnose(code)
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("input", code)
		obj.Set("error", err.Error())
		obj.Set("actionable_tip", diag.ActionableTip)
		obj.Set("clean_length", diag.CleanLength)
		return obj
	}

	stateName := ""
	if rec, ok := postcode.NigerianStates[p.State()]; ok {
		stateName = rec.Name
	}

	obj := js.Global().Get("Object").New()
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
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}
	return validateSingle(args[0].String())
}

func jsValidateBatch(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return js.Global().Get("Array").New()
	}
	arr := args[0]
	n := arr.Length()
	results := js.Global().Get("Array").New(n)
	for i := 0; i < n; i++ {
		code := arr.Index(i).String()
		results.SetIndex(i, validateSingle(code))
	}
	return results
}

func jsDiagnose(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := js.Global().Get("Object").New()
		obj.Set("error", "code parameter is required")
		return obj
	}
	report := postcode.Diagnose(args[0].String())
	return toJS(report)
}

func jsParse(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}

	raw := args[0].String()
	p, err := postcode.Parse(raw)
	if err != nil {
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("error", err.Error())
		return obj
	}

	loc := p.Location()
	capital := ""
	if st, ok := postcode.NigerianStates[p.State()]; ok {
		capital = st.Capital
	}

	obj := js.Global().Get("Object").New()
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
		obj := js.Global().Get("Object").New()
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
		res := js.Global().Get("Object").New()
		res.Set("valid", false)
		res.Set("error", err.Error())
		return res
	}

	res := js.Global().Get("Object").New()
	res.Set("valid", true)
	res.Set("postcode", assembled.Postcode)
	res.Set("display", assembled.Display)
	res.Set("compact", assembled.Compact)
	return res
}

func jsNormalizeSegments(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		obj := js.Global().Get("Object").New()
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
	norm := postcode.NormalizeSegments(segs)
	res := js.Global().Get("Object").New()
	res.Set("state", norm.State)
	res.Set("lga", norm.LGA)
	res.Set("district", norm.District)
	res.Set("area", norm.Area)
	res.Set("unit", norm.Unit)
	return res
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

	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			assembled, err := c.Assemble(context.Background(), segs)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			res := js.Global().Get("Object").New()
			res.Set("valid", true)
			res.Set("postcode", assembled.Postcode)
			res.Set("display", assembled.Display)
			res.Set("compact", assembled.Compact)
			resolve.Invoke(res)
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsDisassembleOnline(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("code parameter is required")
	}
	code := args[0].String()
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			segs, err := c.Disassemble(context.Background(), code)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			resolve.Invoke(toJS(segs))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsDisassemble(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}
	code := args[0].String()
	p, err := postcode.Parse(code)
	if err != nil {
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("error", err.Error())
		return obj
	}

	obj := js.Global().Get("Object").New()
	obj.Set("valid", true)
	obj.Set("state", p.State())
	obj.Set("lga", p.LGA())
	obj.Set("district", p.District())
	obj.Set("area", p.Area())
	obj.Set("unit", p.BuildingUnit())
	return obj
}

func jsResolveLocation(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := js.Global().Get("Object").New()
		obj.Set("error", "code parameter is required")
		return obj
	}

	p, err := postcode.Parse(args[0].String())
	if err != nil {
		obj := js.Global().Get("Object").New()
		obj.Set("error", err.Error())
		return obj
	}

	loc := p.Location()
	jsLoc := toJS(loc)
	jsLoc.Set("state", loc.StateName)
	jsLoc.Set("lga", loc.LGAName)
	jsLoc.Set("google_maps_url", loc.GoogleMapsURL())
	jsLoc.Set("google_maps_directions_url", loc.GoogleMapsDirectionsURL())
	jsLoc.Set("apple_maps_url", loc.AppleMapsURL())
	jsLoc.Set("osm_url", loc.OpenStreetMapURL())
	jsLoc.Set("search_query", loc.SearchQuery())
	return jsLoc
}

func jsResolveLocationOnline(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("code parameter is required")
	}

	code := args[0].String()
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()

			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}

			loc, err := c.ResolveLocation(context.Background(), code)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}

			jsLoc := toJS(loc)
			jsLoc.Set("state", loc.StateName)
			jsLoc.Set("lga", loc.LGAName)
			jsLoc.Set("google_maps_url", loc.GoogleMapsURL())
			jsLoc.Set("google_maps_directions_url", loc.GoogleMapsDirectionsURL())
			jsLoc.Set("apple_maps_url", loc.AppleMapsURL())
			jsLoc.Set("osm_url", loc.OpenStreetMapURL())
			jsLoc.Set("search_query", loc.SearchQuery())
			resolve.Invoke(jsLoc)
		}()
		return nil
	})

	return promiseConstructor.New(handler)
}

func jsRegisterBuilding(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return false
	}
	obj := args[0]
	rec := postcode.BuildingRecord{
		Postcode:  obj.Get("postcode").String(),
		Latitude:  obj.Get("latitude").Float(),
		Longitude: obj.Get("longitude").Float(),
	}
	if obj.Get("address").Type() == js.TypeString {
		rec.Address = obj.Get("address").String()
	}
	if obj.Get("state_code").Type() == js.TypeString {
		rec.StateCode = obj.Get("state_code").String()
	}
	if obj.Get("state_name").Type() == js.TypeString {
		rec.StateName = obj.Get("state_name").String()
	}
	if obj.Get("lga_code").Type() == js.TypeString {
		rec.LGACode = obj.Get("lga_code").String()
	}
	if obj.Get("lga_name").Type() == js.TypeString {
		rec.LGAName = obj.Get("lga_name").String()
	}
	if obj.Get("zone").Type() == js.TypeString {
		rec.Zone = obj.Get("zone").String()
	}

	postcode.RegisterKnownBuilding(rec)
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
			rec := postcode.BuildingRecord{
				Postcode:  item.Get("postcode").String(),
				Latitude:  item.Get("latitude").Float(),
				Longitude: item.Get("longitude").Float(),
			}
			if item.Get("address").Type() == js.TypeString {
				rec.Address = item.Get("address").String()
			}
			if item.Get("state_code").Type() == js.TypeString {
				rec.StateCode = item.Get("state_code").String()
			}
			if item.Get("state_name").Type() == js.TypeString {
				rec.StateName = item.Get("state_name").String()
			}
			if item.Get("lga_code").Type() == js.TypeString {
				rec.LGACode = item.Get("lga_code").String()
			}
			if item.Get("lga_name").Type() == js.TypeString {
				rec.LGAName = item.Get("lga_name").String()
			}
			if item.Get("zone").Type() == js.TypeString {
				rec.Zone = item.Get("zone").String()
			}
			postcode.RegisterKnownBuilding(rec)
			count++
		}
	}
	return count
}

func jsListStates(this js.Value, args []js.Value) any {
	obj := js.Global().Get("Object").New()
	for code, rec := range postcode.NigerianStates {
		st := js.Global().Get("Object").New()
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
		obj.Set(code, st)
	}
	return obj
}

func jsReferenceStatesOffline(this js.Value, args []js.Value) any {
	states := make([]postcode.NamedCode, 0, len(postcode.NigerianStates))
	for code, rec := range postcode.NigerianStates {
		states = append(states, postcode.NamedCode{
			Code: code,
			Name: rec.Name,
		})
	}
	slices.SortFunc(states, func(a, b postcode.NamedCode) int {
		return cmp.Compare(a.Code, b.Code)
	})
	return toJS(states)
}

func jsReferenceStates(this js.Value, args []js.Value) any {
	if len(args) > 0 && args[0].Type() == js.TypeBoolean && !args[0].Bool() {
		return jsReferenceStatesOffline(this, args)
	}
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c != nil {
				states, err := c.ReferenceStates(context.Background())
				if err == nil {
					resolve.Invoke(toJS(states))
					return
				}
			}
			// Fallback to offline catalog
			states := make([]postcode.NamedCode, 0, len(postcode.NigerianStates))
			for code, rec := range postcode.NigerianStates {
				states = append(states, postcode.NamedCode{
					Code: code,
					Name: rec.Name,
				})
			}
			slices.SortFunc(states, func(a, b postcode.NamedCode) int {
				return cmp.Compare(a.Code, b.Code)
			})
			resolve.Invoke(toJS(states))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsReferenceLGAsOffline(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return js.Global().Get("Array").New()
	}
	state := args[0].String()
	lgas := postcode.StateLGAs(state)
	return toJS(lgas)
}

func jsReferenceLGAs(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("state parameter is required")
	}
	state := args[0].String()
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c != nil {
				lgas, err := c.ReferenceLGAs(context.Background(), state)
				if err == nil {
					resolve.Invoke(toJS(lgas))
					return
				}
			}
			// Fallback to offline
			lgas := postcode.StateLGAs(state)
			resolve.Invoke(toJS(lgas))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsReferenceDistricts(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return jsPromiseReject("state and lga parameters are required")
	}
	state := args[0].String()
	lga := args[1].String()
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			districts, err := c.ReferenceDistricts(context.Background(), state, lga)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			resolve.Invoke(toJS(districts))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsReferenceAreas(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return jsPromiseReject("state, lga, and district parameters are required")
	}
	state := args[0].String()
	lga := args[1].String()
	district := args[2].String()
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			areas, err := c.ReferenceAreas(context.Background(), state, lga, district)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			resolve.Invoke(toJS(areas))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsSearchNearbyBuildingsOffline(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return js.Global().Get("Array").New()
	}
	lat := args[0].Float()
	lng := args[1].Float()
	radiusM := 300.0
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		radiusM = args[2].Float()
	}
	results := postcode.SearchNearbyBuildingsOffline(lat, lng, radiusM)
	return toJS(results)
}

func jsReverseCoordinatesOffline(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		obj := js.Global().Get("Object").New()
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
	resp := postcode.ReverseCoordinatesOffline(lat, lng, maxDist)
	return toJS(resp)
}

func jsHealth(this js.Value, args []js.Value) any {
	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			if err := c.Health(context.Background()); err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			res := js.Global().Get("Object").New()
			res.Set("status", "ok")
			resolve.Invoke(res)
		}()
		return nil
	})
	return promiseConstructor.New(handler)
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
	if optObj.Get("apiKey").Type() == js.TypeString {
		currentKey = optObj.Get("apiKey").String()
	}
	if optObj.Get("baseURL").Type() == js.TypeString {
		currentBase = optObj.Get("baseURL").String()
	}
	if optObj.Get("googleMapsApiKey").Type() == js.TypeString {
		currentGoogleMapsKey = optObj.Get("googleMapsApiKey").String()
	} else if optObj.Get("googleMapsKey").Type() == js.TypeString {
		currentGoogleMapsKey = optObj.Get("googleMapsKey").String()
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

	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()

			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}

			resp, err := c.Lookup(context.Background(), code, postcode.LookupLevel(level))
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}

			resolve.Invoke(toJS(resp))
		}()
		return nil
	})

	return promiseConstructor.New(handler)
}

func jsAutocomplete(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("query parameter is required")
	}
	query := args[0].String()

	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			resp, err := c.Autocomplete(context.Background(), query)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			resolve.Invoke(toJS(resp))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsNearby(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("parameters required: specify a postcode string or an options object with {postcode} or {latitude, longitude}")
	}

	var targetCode string
	var lat, lng float64
	radiusM := 300.0 // default 300m
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
		if paramsObj.Get("postcode").Type() == js.TypeString {
			targetCode = paramsObj.Get("postcode").String()
		} else if paramsObj.Get("code").Type() == js.TypeString {
			targetCode = paramsObj.Get("code").String()
		}

		if paramsObj.Get("latitude").Type() == js.TypeNumber {
			lat = paramsObj.Get("latitude").Float()
		}
		if paramsObj.Get("longitude").Type() == js.TypeNumber {
			lng = paramsObj.Get("longitude").Float()
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

		if paramsObj.Get("radius_m").Type() == js.TypeNumber {
			radiusM = paramsObj.Get("radius_m").Float()
		} else if paramsObj.Get("radiusKm").Type() == js.TypeNumber {
			radiusM = paramsObj.Get("radiusKm").Float() * 1000.0
		} else if paramsObj.Get("radius_km").Type() == js.TypeNumber {
			radiusM = paramsObj.Get("radius_km").Float() * 1000.0
		} else if paramsObj.Get("radius").Type() == js.TypeNumber {
			radiusM = paramsObj.Get("radius").Float()
		}

		if paramsObj.Get("limit").Type() == js.TypeNumber {
			limit = paramsObj.Get("limit").Int()
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

	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			resp, err := c.Nearby(context.Background(), params)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			if limit > 0 && len(resp.Results) > limit {
				resp.Results = resp.Results[:limit]
			}
			resolve.Invoke(toJS(resp))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
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
		if paramsObj.Get("max_distance_m").Type() == js.TypeNumber {
			maxDist = paramsObj.Get("max_distance_m").Float()
		} else if paramsObj.Get("maxDistanceKm").Type() == js.TypeNumber {
			maxDist = paramsObj.Get("maxDistanceKm").Float() * 1000.0
		} else if paramsObj.Get("maxDistanceM").Type() == js.TypeNumber {
			maxDist = paramsObj.Get("maxDistanceM").Float()
		} else if paramsObj.Get("max_dist").Type() == js.TypeNumber {
			maxDist = paramsObj.Get("max_dist").Float()
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

	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			if c == nil {
				errObj := js.Global().Get("Error").New("failed to initialize NIPOST client")
				reject.Invoke(errObj)
				return
			}
			resp, err := c.Reverse(context.Background(), params)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			resolve.Invoke(toJS(resp))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsGetAgentTools(this js.Value, args []js.Value) any {
	format := "openai"
	if len(args) > 0 && args[0].Type() == js.TypeString {
		format = args[0].String()
	}
	tools := postcode.DefaultAgentTools()
	var out []any
	switch format {
	case "anthropic":
		for _, t := range tools {
			out = append(out, t.AnthropicTool())
		}
	case "gemini":
		for _, t := range tools {
			out = append(out, t.GeminiFunctionDeclaration())
		}
	default:
		for _, t := range tools {
			out = append(out, t.OpenAITool())
		}
	}
	return toJS(out)
}

func jsExecuteTool(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return jsPromiseReject("tool name parameter is required")
	}
	name := args[0].String()
	var argsJSON []byte
	if len(args) > 1 {
		if args[1].Type() == js.TypeString {
			argsJSON = []byte(args[1].String())
		} else if args[1].Type() == js.TypeObject {
			jsonStr := js.Global().Get("JSON").Call("stringify", args[1]).String()
			argsJSON = []byte(jsonStr)
		}
	}
	if len(argsJSON) == 0 {
		argsJSON = []byte("{}")
	}

	promiseConstructor := js.Global().Get("Promise")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer handler.Release()
			c := getOrCreateClient()
			dispatcher := postcode.NewAgentDispatcher(c)
			res, err := dispatcher.Dispatch(context.Background(), name, argsJSON)
			if err != nil {
				errObj := js.Global().Get("Error").New(err.Error())
				reject.Invoke(errObj)
				return
			}
			resolve.Invoke(toJS(res))
		}()
		return nil
	})
	return promiseConstructor.New(handler)
}

func jsGetAgentMetrics(this js.Value, args []js.Value) any {
	c := getOrCreateClient()
	if c == nil {
		return toJS(postcode.AgentGuardMetrics{})
	}
	return toJS(c.AgentMetrics())
}

func jsGenerateSyntheticAddresses(this js.Value, args []js.Value) any {
	opts := postcode.GeneratorOptions{Count: 10, NoiseRate: 0.3}
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		obj := args[0]
		if obj.Get("count").Type() == js.TypeNumber {
			opts.Count = obj.Get("count").Int()
		}
		if obj.Get("noiseRate").Type() == js.TypeNumber {
			opts.NoiseRate = obj.Get("noiseRate").Float()
		} else if obj.Get("noise_rate").Type() == js.TypeNumber {
			opts.NoiseRate = obj.Get("noise_rate").Float()
		}
		if obj.Get("seed").Type() == js.TypeNumber {
			opts.Seed = uint64(obj.Get("seed").Int())
		}
	}
	addresses := postcode.GenerateSyntheticDataset(opts)
	return toJS(addresses)
}

func main() {
	postcodeObj := js.Global().Get("Object").New()

	// 1. Core Offline (Sync)
	postcodeObj.Set("validate", js.FuncOf(jsValidate))
	postcodeObj.Set("validateBatch", js.FuncOf(jsValidateBatch))
	postcodeObj.Set("diagnose", js.FuncOf(jsDiagnose))
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

	js.Global().Set("Postcode", postcodeObj)

	// Keep WebAssembly event loop running
	<-make(chan struct{})
}
