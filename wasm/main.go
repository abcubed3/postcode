//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/abcubed3/postcode"
)

func toJS(v any) js.Value {
	b, err := json.Marshal(v)
	if err != nil {
		obj := js.Global().Get("Object").New()
		obj.Set("error", err.Error())
		return obj
	}
	return js.Global().Get("JSON").Call("parse", string(b))
}

func jsValidate(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		obj := js.Global().Get("Object").New()
		obj.Set("valid", false)
		obj.Set("error", "code parameter is required")
		return obj
	}

	code := args[0].String()
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
		obj.Set("error", "code parameter is required")
		return obj
	}

	p, err := postcode.Parse(args[0].String())
	if err != nil {
		obj := js.Global().Get("Object").New()
		obj.Set("error", err.Error())
		return obj
	}

	return toJS(p)
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
	jsLoc.Set("google_maps_url", loc.GoogleMapsURL())
	jsLoc.Set("apple_maps_url", loc.AppleMapsURL())
	jsLoc.Set("osm_url", loc.OpenStreetMapURL())
	return jsLoc
}

func jsListStates(this js.Value, args []js.Value) any {
	return toJS(postcode.NigerianStates)
}

func main() {
	postcodeObj := js.Global().Get("Object").New()
	postcodeObj.Set("validate", js.FuncOf(jsValidate))
	postcodeObj.Set("diagnose", js.FuncOf(jsDiagnose))
	postcodeObj.Set("parse", js.FuncOf(jsParse))
	postcodeObj.Set("resolveLocation", js.FuncOf(jsResolveLocation))
	postcodeObj.Set("listStates", js.FuncOf(jsListStates))

	js.Global().Set("Postcode", postcodeObj)

	// Keep WebAssembly event loop running
	<-make(chan struct{})
}
