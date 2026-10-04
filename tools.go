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
			Description: "Searches for active postcode units within a radius (default 300m) of geographic coordinates.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"latitude": map[string]any{
						"type":        "number",
						"description": "Geographic latitude in Nigeria.",
					},
					"longitude": map[string]any{
						"type":        "number",
						"description": "Geographic longitude in Nigeria.",
					},
					"radius_m": map[string]any{
						"type":        "number",
						"description": "Search radius in meters (default 300, max 300).",
					},
				},
				"required": []string{"latitude", "longitude"},
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
			Description: "Queries the NIPOST gateway for graded attributes of a postcode (Level 1 free validity check, Levels 2-3 commercial addresses and building use).",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Nigerian postcode to look up.",
					},
					"level": map[string]any{
						"type":        "integer",
						"description": "Lookup tier: 1 (Free validity check), 2 (Commercial address), 3 (Commercial building use). Default is 1.",
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

	case "reverse_geocode":
		if d.client == nil {
			return nil, fmt.Errorf("gateway client is required for reverse geocoding")
		}
		var args struct {
			Latitude     float64 `json:"latitude"`
			Longitude    float64 `json:"longitude"`
			MaxDistanceM float64 `json:"max_distance_m"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		return d.client.Reverse(ctx, ReverseParams{
			Latitude:     args.Latitude,
			Longitude:    args.Longitude,
			MaxDistanceM: args.MaxDistanceM,
		})

	case "search_nearby":
		if d.client == nil {
			return nil, fmt.Errorf("gateway client is required for nearby search")
		}
		var args struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			RadiusM   float64 `json:"radius_m"`
		}
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		return d.client.Nearby(ctx, NearbyParams{
			Latitude:  args.Latitude,
			Longitude: args.Longitude,
			RadiusM:   args.RadiusM,
		})

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
