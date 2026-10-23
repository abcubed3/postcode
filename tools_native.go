//go:build !js

package postcode

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// Dispatch executes the specified tool by name with unparsed JSON arguments.
func (d *AgentDispatcher) Dispatch(ctx context.Context, name string, argsJSON []byte) (any, error) {
	switch name {
	case "assemble_postcode":
		var args Segments
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		assembled, err := AssembleSegments(args)
		if err != nil {
			return nil, err
		}
		return assembled, nil

	case "validate_postcode":
		var args struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if err := Validate(args.Code); err != nil {
			report := Diagnose(args.Code)
			res := map[string]any{
				"valid":          false,
				"input":          args.Code,
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
		p, _ := Parse(args.Code)
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
		var args struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		return Diagnose(args.Code), nil

	case "repair_postcode":
		var args struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		p, ok := TryRepair(args.Code)
		if !ok {
			report := Diagnose(args.Code)
			return map[string]any{
				"repaired":       false,
				"input":          args.Code,
				"actionable_tip": report.ActionableTip,
				"diagnoses":      report.Diagnoses,
			}, nil
		}
		return map[string]any{
			"repaired": true,
			"input":    args.Code,
			"postcode": p.Formatted(),
			"display":  p.String(),
			"compact":  p.Compact(),
			"slug":     p.Slug(),
		}, nil

	case "extract_address":
		var args struct {
			Address string `json:"address"`
			Raw     string `json:"raw"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		input := args.Address
		if input == "" {
			input = args.Raw
		}
		return ExtractAddress(input), nil

	case "calculate_distance":
		var args struct {
			FromCode string `json:"from_code"`
			ToCode   string `json:"to_code"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		distM, err := DistanceMeters(args.FromCode, args.ToCode)
		if err != nil {
			return nil, err
		}
		fromLoc, _ := ResolveLocation(args.FromCode)
		toLoc, _ := ResolveLocation(args.ToCode)
		return map[string]any{
			"from_postcode": fromLoc.Postcode,
			"to_postcode":   toLoc.Postcode,
			"distance_m":    distM,
			"distance_km":   distM / 1000.0,
		}, nil

	case "resolve_location":
		var args struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if d.client != nil {
			return d.client.ResolveLocation(ctx, args.Code)
		}
		loc, err := ResolveLocation(args.Code)
		if err != nil {
			return nil, err
		}
		return &loc, nil

	case "disassemble_postcode":
		var args struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if d.client != nil {
			segs, err := d.client.Disassemble(ctx, args.Code)
			if err == nil {
				return segs, nil
			}
		}
		p, err := Parse(args.Code)
		if err != nil {
			return nil, err
		}
		segs := p.Disassemble()
		return &segs, nil

	case "reverse_geocode":
		var args struct {
			Latitude     float64 `json:"latitude"`
			Longitude    float64 `json:"longitude"`
			MaxDistanceM float64 `json:"max_distance_m"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if d.client != nil {
			resp, err := d.client.Reverse(ctx, ReverseParams{
				Latitude:     args.Latitude,
				Longitude:    args.Longitude,
				MaxDistanceM: args.MaxDistanceM,
			})
			if err == nil {
				return resp, nil
			}
		}
		return ReverseCoordinatesOffline(args.Latitude, args.Longitude, args.MaxDistanceM), nil

	case "search_nearby":
		var args struct {
			Code      string  `json:"code"`
			Postcode  string  `json:"postcode"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			RadiusM   float64 `json:"radius_m"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		targetCode := args.Code
		if targetCode == "" {
			targetCode = args.Postcode
		}
		lat := args.Latitude
		lng := args.Longitude
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
				RadiusM:   args.RadiusM,
			})
			if err == nil {
				return resp, nil
			}
		}
		radius := args.RadiusM
		if radius <= 0 {
			radius = 300
		}
		units := SearchNearbyBuildingsOffline(lat, lng, radius)
		return &NearbyResponse{
			Results: units,
		}, nil

	case "autocomplete_postcode":
		if d.client == nil {
			return nil, fmt.Errorf("gateway client is required for autocomplete")
		}
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		return d.client.Autocomplete(ctx, args.Query)

	case "lookup_postcode":
		if d.client == nil {
			return nil, fmt.Errorf("gateway client is required for live lookup")
		}
		var args struct {
			Code  string `json:"code"`
			Level int    `json:"level"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		lvl := Level1
		if args.Level > 0 {
			lvl = LookupLevel(args.Level)
		}
		return d.client.Lookup(ctx, args.Code, lvl)

	case "list_states":
		var states []StateRecord
		for _, rec := range NigerianStates {
			states = append(states, rec)
		}
		sort.Slice(states, func(i, j int) bool {
			return states[i].Code < states[j].Code
		})
		return states, nil

	case "list_lgas":
		var args struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if d.client != nil {
			lgas, err := d.client.ReferenceLGAs(ctx, args.State)
			if err == nil {
				return lgas, nil
			}
		}
		return StateLGAs(args.State), nil

	case "generate_postcode_qr":
		var args struct {
			Code            string `json:"code"`
			Payload         string `json:"payload"`
			IncludeLabel    bool   `json:"include_label"`
			Title           string `json:"title"`
			Subtitle        string `json:"subtitle"`
			Size            int    `json:"size"`
			Level           string `json:"level"`
			ForegroundColor string `json:"foreground_color"`
			BackgroundColor string `json:"background_color"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		opts := QROptions{
			Payload:         QRPayloadType(args.Payload),
			Level:           QRErrorCorrectionLevel(args.Level),
			Size:            args.Size,
			IncludeLabel:    args.IncludeLabel,
			Title:           args.Title,
			Subtitle:        args.Subtitle,
			ForegroundColor: args.ForegroundColor,
			BackgroundColor: args.BackgroundColor,
		}
		var qr *QRCode
		var err error
		if d.client != nil {
			qr, err = d.client.GenerateQR(ctx, args.Code, opts)
		} else {
			qr, err = GenerateQR(args.Code, opts)
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

// DispatchString executes the tool and returns the result serialized as a JSON string,
// safe for direct injection into LLM response messages.
func (d *AgentDispatcher) DispatchString(ctx context.Context, name string, argsJSON []byte) (string, error) {
	res, err := d.Dispatch(ctx, name, argsJSON)
	if err != nil {
		errPayload, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(errPayload), err
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
