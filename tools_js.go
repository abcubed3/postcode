//go:build js && wasm

package postcode

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"syscall/js"
)

func jsIntField(obj js.Value, key string) int {
	if obj.Type() != js.TypeObject {
		return 0
	}
	v := obj.Get(key)
	if v.Type() == js.TypeNumber {
		return v.Int()
	}
	return 0
}

// Dispatch executes the specified tool by name with JSON arguments on js/wasm
// using native JS JSON.parse() without linking Go's encoding/json package.
func (d *AgentDispatcher) Dispatch(ctx context.Context, name string, argsJSON []byte) (any, error) {
	rawArgs := string(argsJSON)
	if rawArgs == "" {
		rawArgs = "{}"
	}
	argsObj, err := parseJSONSafe(rawArgs)
	if err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	return d.DispatchJSValue(ctx, name, argsObj)
}

// DispatchJSValue executes the specified tool directly from a js.Value arguments object,
// avoiding JSON stringification and parsing overhead across the WASM boundary.
func (d *AgentDispatcher) DispatchJSValue(ctx context.Context, name string, argsObj js.Value) (any, error) {
	switch name {
	case "assemble_postcode":
		args := Segments{
			State:    jsStringField(argsObj, "state"),
			LGA:      jsStringField(argsObj, "lga"),
			District: jsStringField(argsObj, "district"),
			Area:     jsStringField(argsObj, "area"),
			Unit:     jsStringField(argsObj, "unit"),
		}
		assembled, err := AssembleSegments(args)
		if err != nil {
			return nil, err
		}
		return assembled, nil

	case "validate_postcode":
		code := jsStringField(argsObj, "code")
		if err := Validate(code); err != nil {
			report := Diagnose(code)
			res := map[string]any{
				"valid":          false,
				"input":          code,
				"error":          err.Error(),
				"actionable_tip": report.ActionableTip,
				"clean_length":   report.CleanLength,
			}
			if report.IsTemporary {
				res["is_temporary"] = true
			}
			if report.SuggestedFix != "" {
				res["suggested_fix"] = report.SuggestedFix
			}
			return res, nil
		}
		p, _ := Parse(code)
		stateName := ""
		if rec, ok := NigerianStates[p.State()]; ok {
			stateName = rec.Name
		}
		return map[string]any{
			"valid":      true,
			"postcode":   p.Formatted(),
			"compact":    p.Compact(),
			"slug":       p.Slug(),
			"state_code": p.State(),
			"state_name": stateName,
			"lga_code":   p.LGA(),
			"district":   p.District(),
			"area":       p.Area(),
			"unit":       p.BuildingUnit(),
		}, nil

	case "diagnose_postcode":
		return Diagnose(jsStringField(argsObj, "code")), nil

	case "repair_postcode":
		code := jsStringField(argsObj, "code")
		p, ok := TryRepair(code)
		if !ok {
			report := Diagnose(code)
			return map[string]any{
				"repaired":       false,
				"input":          code,
				"actionable_tip": report.ActionableTip,
				"diagnoses":      report.Diagnoses,
			}, nil
		}
		return map[string]any{
			"repaired": true,
			"input":    code,
			"postcode": p.Formatted(),
			"display":  p.String(),
			"compact":  p.Compact(),
			"slug":     p.Slug(),
		}, nil

	case "extract_address":
		input := jsStringField(argsObj, "address")
		if input == "" {
			input = jsStringField(argsObj, "raw")
		}
		return ExtractAddress(input), nil

	case "calculate_distance":
		fromCode := jsStringField(argsObj, "from_code")
		toCode := jsStringField(argsObj, "to_code")
		distM, err := DistanceMeters(fromCode, toCode)
		if err != nil {
			return nil, err
		}
		fromLoc, _ := ResolveLocation(fromCode)
		toLoc, _ := ResolveLocation(toCode)
		return map[string]any{
			"from_postcode": fromLoc.Postcode,
			"to_postcode":   toLoc.Postcode,
			"distance_m":    distM,
			"distance_km":   distM / 1000.0,
		}, nil

	case "resolve_location":
		code := jsStringField(argsObj, "code")
		if d.client != nil {
			return d.client.ResolveLocation(ctx, code)
		}
		loc, err := ResolveLocation(code)
		if err != nil {
			return nil, err
		}
		return &loc, nil

	case "disassemble_postcode":
		code := jsStringField(argsObj, "code")
		if d.client != nil {
			segs, err := d.client.Disassemble(ctx, code)
			if err == nil {
				return segs, nil
			}
		}
		p, err := Parse(code)
		if err != nil {
			return nil, err
		}
		segs := p.Disassemble()
		return &segs, nil

	case "reverse_geocode":
		lat := jsFloatField(argsObj, "latitude")
		lng := jsFloatField(argsObj, "longitude")
		maxDist := jsFloatField(argsObj, "max_distance_m")
		if d.client != nil {
			resp, err := d.client.Reverse(ctx, ReverseParams{
				Latitude:     lat,
				Longitude:    lng,
				MaxDistanceM: maxDist,
			})
			if err == nil {
				return resp, nil
			}
		}
		return ReverseCoordinatesOffline(lat, lng, maxDist), nil

	case "search_nearby":
		targetCode := jsStringField(argsObj, "code")
		if targetCode == "" {
			targetCode = jsStringField(argsObj, "postcode")
		}
		lat := jsFloatField(argsObj, "latitude")
		lng := jsFloatField(argsObj, "longitude")
		radius := jsFloatField(argsObj, "radius_m")
		if lat == 0 && lng == 0 && targetCode != "" {
			p, err := Parse(targetCode)
			if err != nil {
				return nil, fmt.Errorf("invalid reference postcode %q: %w", targetCode, err)
			}
			loc := p.Location()
			lat = loc.Latitude
			lng = loc.Longitude
		}
		if lat == 0 && lng == 0 {
			return nil, fmt.Errorf("either reference postcode ('code') or 'latitude' and 'longitude' must be provided")
		}
		if d.client != nil {
			resp, err := d.client.Nearby(ctx, NearbyParams{
				Postcode:  targetCode,
				Latitude:  lat,
				Longitude: lng,
				RadiusM:   radius,
			})
			if err == nil {
				return resp, nil
			}
		}
		if radius <= 0 {
			radius = 300
		}
		units := SearchNearbyBuildingsOffline(lat, lng, radius)
		return &NearbyResponse{Results: units}, nil

	case "autocomplete_postcode":
		if d.client == nil {
			return nil, fmt.Errorf("gateway client is required for autocomplete")
		}
		return d.client.Autocomplete(ctx, jsStringField(argsObj, "query"))

	case "lookup_postcode":
		if d.client == nil {
			return nil, fmt.Errorf("gateway client is required for live lookup")
		}
		lvl := Level1
		if n := jsIntField(argsObj, "level"); n > 0 {
			lvl = LookupLevel(n)
		}
		return d.client.Lookup(ctx, jsStringField(argsObj, "code"), lvl)

	case "list_states":
		states := make([]StateRecord, 0, len(NigerianStates))
		for _, rec := range NigerianStates {
			states = append(states, rec)
		}
		sort.Slice(states, func(i, j int) bool {
			return states[i].Code < states[j].Code
		})
		return states, nil

	case "list_lgas":
		state := jsStringField(argsObj, "state")
		if d.client != nil {
			lgas, err := d.client.ReferenceLGAs(ctx, state)
			if err == nil {
				return lgas, nil
			}
		}
		return StateLGAs(state), nil

	case "generate_postcode_qr":
		code := jsStringField(argsObj, "code")
		opts := QROptions{
			Payload:         QRPayloadType(jsStringField(argsObj, "payload")),
			Level:           QRErrorCorrectionLevel(jsStringField(argsObj, "level")),
			Size:            jsIntField(argsObj, "size"),
			IncludeLabel:    jsBoolField(argsObj, "include_label"),
			Title:           jsStringField(argsObj, "title"),
			Subtitle:        jsStringField(argsObj, "subtitle"),
			ForegroundColor: jsStringField(argsObj, "foreground_color"),
			BackgroundColor: jsStringField(argsObj, "background_color"),
		}
		var qr *QRCode
		var err error
		if d.client != nil {
			qr, err = d.client.GenerateQR(ctx, code, opts)
		} else {
			qr, err = GenerateQR(code, opts)
		}
		if err != nil {
			return nil, err
		}
		dataURI, _ := qr.DataURI()
		return map[string]any{
			"valid":           true,
			"postcode":        qr.Postcode,
			"compact":         qr.Compact,
			"state_code":      qr.StateCode,
			"state_name":      qr.StateName,
			"lga_code":        qr.LGACode,
			"lga_name":        qr.LGAName,
			"latitude":        qr.Latitude,
			"longitude":       qr.Longitude,
			"google_maps_url": qr.GoogleMapsURL,
			"payload":         qr.Payload,
			"payload_type":    qr.PayloadType,
			"level":           qr.Level,
			"version":         qr.Version,
			"dimension":       qr.Dimension,
			"include_label":   qr.IncludeLabel,
			"svg":             qr.SVG(),
			"label_svg":       qr.LabelSVG(),
			"data_uri":        dataURI,
			"svg_data_uri":    qr.SVGDataURI(),
			"terminal":        qr.Terminal(),
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool name: %q", name)
	}
}

// DispatchString executes the tool and returns a JSON string error or result on js/wasm.
func (d *AgentDispatcher) DispatchString(ctx context.Context, name string, argsJSON []byte) (string, error) {
	res, err := d.Dispatch(ctx, name, argsJSON)
	if err != nil {
		return `{"error":` + strconv.Quote(err.Error()) + `}`, err
	}
	return encodeJSJSON(res), nil
}

func encodeJSJSON(v any) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		return strconv.Quote(val)
	case bool:
		return strconv.FormatBool(val)
	case int:
		return strconv.Itoa(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case QRPayloadType:
		return strconv.Quote(string(val))
	case QRErrorCorrectionLevel:
		return strconv.Quote(string(val))
	case []string:
		var sb strings.Builder
		sb.WriteByte('[')
		for i, s := range val {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(strconv.Quote(s))
		}
		sb.WriteByte(']')
		return sb.String()
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sb strings.Builder
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(strconv.Quote(k))
			sb.WriteByte(':')
			sb.WriteString(encodeJSJSON(val[k]))
		}
		sb.WriteByte('}')
		return sb.String()
	case AssembledPostcode:
		return `{"postcode":` + strconv.Quote(val.Postcode) + `,"display":` + strconv.Quote(val.Display) + `,"compact":` + strconv.Quote(val.Compact) + `}`
	case *AssembledPostcode:
		if val == nil {
			return "null"
		}
		return encodeJSJSON(*val)
	case Segments:
		return `{"state":` + strconv.Quote(val.State) + `,"lga":` + strconv.Quote(val.LGA) + `,"district":` + strconv.Quote(val.District) + `,"area":` + strconv.Quote(val.Area) + `,"unit":` + strconv.Quote(val.Unit) + `}`
	case *Segments:
		if val == nil {
			return "null"
		}
		return encodeJSJSON(*val)
	case DiagnosticReport:
		return encodeJSJSON(map[string]any{
			"input":          val.Input,
			"valid":          val.Valid,
			"clean_length":   val.CleanLength,
			"is_temporary":   val.IsTemporary,
			"suggested_fix":  val.SuggestedFix,
			"actionable_tip": val.ActionableTip,
			"diagnoses":      val.Diagnoses,
		})
	case AddressExtraction:
		return encodeJSJSON(map[string]any{
			"raw":              val.Raw,
			"postcode":         val.Postcode,
			"postcode_valid":   val.PostcodeValid,
			"legacy_postcode":  val.LegacyPostcode,
			"suggested_prefix": val.SuggestedPrefix,
			"state_code":       val.StateCode,
			"state_name":       val.StateName,
			"lga_code":         val.LGACode,
			"lga_name":         val.LGAName,
			"locality":         val.Locality,
			"street":           val.Street,
			"landmark":         val.Landmark,
			"latitude":         val.Latitude,
			"longitude":        val.Longitude,
			"confidence":       val.Confidence,
		})
	case Location:
		return encodeJSJSON(map[string]any{
			"postcode":   val.Postcode,
			"compact":    val.Compact,
			"latitude":   val.Latitude,
			"longitude":  val.Longitude,
			"address":    val.Address,
			"state_code": val.StateCode,
			"state_name": val.StateName,
			"lga_code":   val.LGACode,
			"lga_name":   val.LGAName,
			"zone":       val.Zone,
			"precision":  val.Precision.String(),
		})
	case *Location:
		if val == nil {
			return "null"
		}
		return encodeJSJSON(*val)
	default:
		return `{"result":` + strconv.Quote(fmt.Sprintf("%v", v)) + `}`
	}
}
