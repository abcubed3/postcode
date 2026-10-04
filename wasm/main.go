//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"sync"
	"syscall/js"

	"github.com/abcubed3/postcode"
)

var (
	clientMu    sync.RWMutex
	currentKey  string
	currentBase string
	client      *postcode.Client
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
	c, err := postcode.NewClient(opts...)
	if err == nil {
		client = c
	}
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
	resetClientLocked()
	return true
}

func jsLookup(this js.Value, args []js.Value) any {
	promiseConstructor := js.Global().Get("Promise")

	if len(args) == 0 {
		var rejectHandler js.Func
		rejectHandler = js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
			defer rejectHandler.Release()
			errObj := js.Global().Get("Error").New("code parameter is required")
			promiseArgs[1].Invoke(errObj)
			return nil
		})
		return promiseConstructor.New(rejectHandler)
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
	postcodeObj.Set("setAPIKey", js.FuncOf(jsSetAPIKey))
	postcodeObj.Set("getAPIKey", js.FuncOf(jsGetAPIKey))
	postcodeObj.Set("configure", js.FuncOf(jsConfigure))
	postcodeObj.Set("lookup", js.FuncOf(jsLookup))

	js.Global().Set("Postcode", postcodeObj)

	// Keep WebAssembly event loop running
	<-make(chan struct{})
}
