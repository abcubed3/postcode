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
			Description: "Queries the NIPOST gateway for graded attributes of a postcode (Level 1 free validity check, Levels 2-3 commercial addresses and building use, Levels 4-5 building metadata and geometry).",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Nigerian postcode to look up.",
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
		p, err := Parse(args.Code)
		if err != nil {
			report := Diagnose(args.Code)
			return map[string]any{
				"valid":          false,
				"input":          args.Code,
				"error":          err.Error(),
				"actionable_tip": report.ActionableTip,
				"clean_length":   report.CleanLength,
			}, nil
		}
		stateName := ""
		if rec, ok := NigerianStates[p.State()]; ok {
			stateName = rec.Name
		}
		return map[string]any{
			"valid":      true,
			"postcode":   p.Formatted(),
			"compact":    p.Compact(),
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
