package postcode

import "sort"

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
