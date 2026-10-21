package postcode

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// ToolDef describes an agent-executable function schema compliant with JSON Schema Draft-07.
type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// OpenAITool formats the tool declaration for OpenAI, Mistral, Groq, and Ollama APIs.
func (t ToolDef) OpenAITool() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  t.Parameters,
		},
	}
}

// AnthropicTool formats the tool declaration for the Anthropic Claude Messages API.
func (t ToolDef) AnthropicTool() map[string]any {
	return map[string]any{
		"name":         t.Name,
		"description":  t.Description,
		"input_schema": t.Parameters,
	}
}

// GeminiFunctionDeclaration formats the tool declaration for Google Gemini / Google GenAI SDK.
func (t ToolDef) GeminiFunctionDeclaration() map[string]any {
	return map[string]any{
		"name":        t.Name,
		"description": t.Description,
		"parameters":  t.Parameters,
	}
}

// DefaultAgentTools returns the standard suite of tool declarations for Nigerian postcode operations.
func DefaultAgentTools() []ToolDef {
	tools := []ToolDef{
		{
			Name:        "validate_postcode",
			Description: "Validates an 11-character Nigerian postcode for grammar, structure, and recognized state code.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Candidate Nigerian postcode (e.g. 'EK-01-A03-FK-01', 'LA 11 W06 TC 10', or compact 'EK01A03FK01').",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "diagnose_postcode",
			Description: "Deeply analyzes a potentially invalid Nigerian postcode, returning segment-by-segment errors, suggestions, and actionable guidance for self-correction.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Candidate postcode to inspect and diagnose.",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "resolve_location",
			Description: "Resolves an 11-digit Nigerian postcode to geographic coordinates (lat/long), state, LGA, precision tier, and Google Maps URL.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Valid 11-character Nigerian postcode.",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "reverse_geocode",
			Description: "Snaps latitude and longitude coordinates to the nearest active Nigerian postcode unit.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"latitude": map[string]any{
						"type":        "number",
						"description": "Geographic latitude in Nigeria (approx 4.0 to 14.0).",
					},
					"longitude": map[string]any{
						"type":        "number",
						"description": "Geographic longitude in Nigeria (approx 2.5 to 14.5).",
					},
					"max_distance_m": map[string]any{
						"type":        "number",
						"description": "Maximum search radius in meters (default 25, max 250).",
					},
				},
				"required": []string{"latitude", "longitude"},
			},
		},
		{
			Name:        "search_nearby",
			Description: "Searches for active postcode units within a radius (default 300m) of geographic coordinates or reference postcode.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Optional reference Nigerian postcode (e.g. 'LA-08-A86-RG-01'). If provided, centroid coordinates are resolved automatically.",
					},
					"latitude": map[string]any{
						"type":        "number",
						"description": "Optional geographic latitude in Nigeria.",
					},
					"longitude": map[string]any{
						"type":        "number",
						"description": "Optional geographic longitude in Nigeria.",
					},
					"radius_m": map[string]any{
						"type":        "number",
						"description": "Search radius in meters (default 300, max 300).",
					},
				},
			},
		},
		{
			Name:        "autocomplete_postcode",
			Description: "Retrieves segment-aware autocomplete suggestions for partial Nigerian postcode inputs (e.g. 'EK 01 A').",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Partial postcode query prefix.",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "lookup_postcode",
			Description: "Queries the NIPOST gateway for graded attributes of an official 11-character postcode or temporary building code (TP-...) (Level 1 free validity check, Levels 2-3 commercial addresses and building use, Levels 4-5 building metadata and geometry).",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Nigerian postcode or temporary code (TP-...) to look up.",
					},
					"level": map[string]any{
						"type":        "integer",
						"description": "Lookup tier: 1 (Free validity check), 2 (Commercial address), 3 (Commercial building use), 4 (Building metadata), 5 (Point geometry). Default is 1.",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "list_states",
			Description: "Returns reference information for all 36 Nigerian states and the Federal Capital Territory (FCT), including 2-letter codes, capitals, zones, and centroids.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "list_lgas",
			Description: "Returns all Local Government Areas (LGAs) for a given Nigerian state code (e.g. 'LA', 'FC', 'EK') from official reference data.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"state": map[string]any{
						"type":        "string",
						"description": "2-letter Nigerian state code (e.g. 'LA', 'FC', 'EK').",
					},
				},
				"required": []string{"state"},
			},
		},
		{
			Name:        "assemble_postcode",
			Description: "Assembles 5 administrative segments (state, LGA, district, area, building unit) into canonical, display, and compact Nigerian postcodes, with automatic zero-padding for single-digit numbers.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"state":    map[string]any{"type": "string", "description": "2-letter state code (e.g. 'EK', 'LA')."},
					"lga":      map[string]any{"type": "string", "description": "LGA numeric code (e.g. '01' or '1')."},
					"district": map[string]any{"type": "string", "description": "3-character district code (e.g. 'A03')."},
					"area":     map[string]any{"type": "string", "description": "2-letter area code (e.g. 'FK')."},
					"unit":     map[string]any{"type": "string", "description": "Building unit code (e.g. '01' or '1')."},
				},
				"required": []string{"state", "lga", "district", "area", "unit"},
			},
		},
		{
			Name:        "disassemble_postcode",
			Description: "Deconstructs an 11-digit Nigerian postcode into its 5 constituent administrative segments (State, LGA, District, Area, Unit) and structural components.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "11-character Nigerian postcode to disassemble (e.g. 'EK-01-A03-FK-01').",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "repair_postcode",
			Description: "Deterministically self-heals a malformed or noisy Nigerian postcode (padding single-digit segments, fixing '00' to '01', resolving state aliases like 'BE'->'BN', and correcting OCR lookalikes).",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Malformed or noisy Nigerian postcode candidate (e.g. 'EK-1-A03-FK-1', 'LA-00-W06-TC-00', 'BE-01-A01-AA-01').",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "extract_address",
			Description: "Extracts structured Nigerian address entities (postcode, legacy 6-digit code, state, LGA, neighborhood locality, street, landmark anchor, and coordinates) from unstructured text using the 5-stage disambiguation pipeline.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"address": map[string]any{
						"type":        "string",
						"description": "Raw unstructured Nigerian address string.",
					},
				},
				"required": []string{"address"},
			},
		},
		{
			Name:        "calculate_distance",
			Description: "Calculates the Haversine distance in meters and kilometers between two Nigerian postcodes.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from_code": map[string]any{
						"type":        "string",
						"description": "Origin 11-character Nigerian postcode.",
					},
					"to_code": map[string]any{
						"type":        "string",
						"description": "Destination 11-character Nigerian postcode.",
					},
				},
				"required": []string{"from_code", "to_code"},
			},
		},
		{
			Name:        "generate_postcode_qr",
			Description: "Converts a valid 11-character Nigerian postcode into a shareable and printable QR code (SVG, printable NIPOST label SVG, PNG base64 Data URI, and terminal Unicode blocks).",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Valid 11-character Nigerian postcode (e.g. 'EK-01-A03-FK-01').",
					},
					"payload": map[string]any{
						"type":        "string",
						"description": "Payload to encode when scanned: 'postcode' (default canonical string), 'compact', 'spaced', 'google-maps', 'directions', 'apple-maps', 'osm', 'geo', 'json', or 'text'.",
					},
					"include_label": map[string]any{
						"type":        "boolean",
						"description": "When true, renders a printable NIPOST postcode badge/label with the formatted postcode, State/LGA, and coordinates.",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "Optional header title for printable labels (defaults to 'NIGERIA DIGITAL POSTCODE').",
					},
					"subtitle": map[string]any{
						"type":        "string",
						"description": "Optional street address or recipient subtitle for printable labels.",
					},
					"size": map[string]any{
						"type":        "integer",
						"description": "Target image width/height in pixels (default 256, or 400 with label).",
					},
					"level": map[string]any{
						"type":        "string",
						"description": "QR error correction level: 'L', 'M', 'Q', or 'H'.",
					},
				},
				"required": []string{"code"},
			},
		},
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
	return tools
}

// AgentDispatcher handles incoming tool executions from AI agent frameworks.
type AgentDispatcher struct {
	client *Client
}

// NewAgentDispatcher creates a dispatcher. If client is nil, the dispatcher
// operates in offline mode, servicing local validation, diagnostics, state listing,
// and reference geocoding.
func NewAgentDispatcher(client *Client) *AgentDispatcher {
	return &AgentDispatcher{client: client}
}

// Tools returns the list of available tools.
func (d *AgentDispatcher) Tools() []ToolDef {
	return DefaultAgentTools()
}

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
