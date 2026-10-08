package postcode

import (
	"bufio"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"sort"
	"strings"
	"sync"
)

// MCP Protocol Revisions & Server Metadata per MCP 2026-07-28 Specification
const (
	// MCPProtocolVersionLatest is the modern stateless protocol revision (SEP-2575).
	MCPProtocolVersionLatest = "2026-07-28"
	// MCPProtocolVersion2025 is the intermediate 2025 specification revision.
	MCPProtocolVersion2025 = "2025-11-25"
	// MCPProtocolVersionLegacy is the original initialization-based protocol revision.
	MCPProtocolVersionLegacy = "2024-11-05"

	MCPServerName    = "postcode-mcp"
	MCPServerVersion = "1.0.0"

	// Standard MCP error codes (JSON-RPC reserved range per 2026-07-28 spec)
	errCodeUnsupportedProtocolVersion = -32022
	errCodeInvalidParams              = -32602
	errCodeMethodNotFound             = -32601
	errCodeParseError                 = -32700
)

// MCPSupportedVersions lists all supported protocol revisions in order of preference.
var MCPSupportedVersions = []string{
	MCPProtocolVersionLatest,
	MCPProtocolVersion2025,
	MCPProtocolVersionLegacy,
}

// MCPOption configures an MCPServer instance.
type MCPOption func(*MCPServer)

// WithMCPLogger specifies a logger for internal MCP diagnostics (must write to stderr).
func WithMCPLogger(logger *log.Logger) MCPOption {
	return func(s *MCPServer) {
		if logger != nil {
			s.logger = logger
		}
	}
}

// MCPServer implements a dual-era Model Context Protocol (MCP) server conforming to
// the MCP 2026-07-28 specification, with backward compatibility for legacy clients.
// Operates statelessly over stdio or any streaming io.ReadWriter.
type MCPServer struct {
	dispatcher *AgentDispatcher
	logger     *log.Logger
	writeMu    sync.Mutex
}

// NewMCPServer creates a ready-to-run MCP server. If client is nil, the server
// operates in offline mode, servicing local validation, diagnostics, and reference geocoding.
func NewMCPServer(client *Client, opts ...MCPOption) *MCPServer {
	s := &MCPServer{
		dispatcher: NewAgentDispatcher(client),
		logger:     log.New(io.Discard, "[postcode-mcp] ", 0),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Wire structures for JSON-RPC 2.0 with MCP metadata
type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpParamsWithMeta struct {
	Meta map[string]any `json:"_meta,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// standardServerInfoMeta returns the standard _meta envelope with server identity.
func standardServerInfoMeta() map[string]any {
	return map[string]any{
		"io.modelcontextprotocol/serverInfo": map[string]any{
			"name":    MCPServerName,
			"version": MCPServerVersion,
		},
	}
}

// isVersionSupported checks whether a given protocol version is recognized.
func isVersionSupported(v string) bool {
	for _, supported := range MCPSupportedVersions {
		if v == supported {
			return true
		}
	}
	return false
}

// Serve reads JSON-RPC 2.0 messages from in, executes corresponding MCP methods,
// and writes newline-delimited JSON-RPC responses to out.
func (s *MCPServer) Serve(in io.Reader, out io.Writer) error {
	reader := bufio.NewReaderSize(in, 1024*1024)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("mcp read error: %w", err)
		}

		trimmed := strings.TrimSpace(string(line))
		if len(trimmed) == 0 {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal([]byte(trimmed), &req); err != nil {
			s.sendError(out, nil, errCodeParseError, "Parse error: invalid JSON", nil)
			continue
		}

		// Handle notifications (no ID provided)
		if len(req.ID) == 0 || string(req.ID) == "null" {
			s.handleNotification(req)
			continue
		}

		// Handle standard JSON-RPC calls
		resp := s.handleRequest(context.Background(), req)
		if err := s.sendResponse(out, resp); err != nil {
			return fmt.Errorf("mcp write error: %w", err)
		}
	}
}

func (s *MCPServer) handleNotification(req mcpRequest) {
	switch req.Method {
	case "notifications/initialized":
		s.logger.Printf("client initialized")
	case "notifications/cancelled":
		s.logger.Printf("client cancelled operation")
	default:
		s.logger.Printf("received notification: %s", req.Method)
	}
}

func (s *MCPServer) handleRequest(ctx context.Context, req mcpRequest) mcpResponse {
	// 1. Inspect _meta for protocol version validation in modern requests
	if len(req.Params) > 0 {
		var metaHolder mcpParamsWithMeta
		if err := json.Unmarshal(req.Params, &metaHolder); err == nil && metaHolder.Meta != nil {
			if reqVer, ok := metaHolder.Meta["io.modelcontextprotocol/protocolVersion"].(string); ok && reqVer != "" {
				if !isVersionSupported(reqVer) {
					return mcpResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error: &mcpError{
							Code:    errCodeUnsupportedProtocolVersion,
							Message: "Unsupported protocol version",
							Data: map[string]any{
								"supported": MCPSupportedVersions,
								"requested": reqVer,
							},
						},
					}
				}
			}
		}
	}

	switch req.Method {
	// 2. Modern MCP 2026-07-28 mandatory Discovery RPC (SEP-2575)
	case "server/discover":
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType":        "complete",
				"supportedVersions": MCPSupportedVersions,
				"capabilities": map[string]any{
					"tools":     map[string]any{},
					"resources": map[string]any{},
					"prompts":   map[string]any{},
				},
				"_meta":        standardServerInfoMeta(),
				"instructions": "This server provides comprehensive tools and reference data for Nigeria's National Digital Alphanumeric Postcodes (NIPOST): validation, diagnostic guidance, location resolution, reverse geocoding, nearby radius search, autocomplete, graded gateway lookups (Levels 1–5), states and LGA reference catalogs, and segment-level postcode assembly and disassembly.",
				"ttlMs":        3600000, // 1 hour caching recommendation (SEP-2549)
				"cacheScope":   "public",
			},
		}

	// 3. Dual-era backward-compatibility: legacy initialization handshake
	case "initialize":
		var initParams struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &initParams)

		negotiatedVersion := MCPProtocolVersionLatest
		if initParams.ProtocolVersion != "" {
			if isVersionSupported(initParams.ProtocolVersion) {
				negotiatedVersion = initParams.ProtocolVersion
			} else {
				return mcpResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &mcpError{
						Code:    errCodeUnsupportedProtocolVersion,
						Message: "Unsupported protocol version",
						Data: map[string]any{
							"supported": MCPSupportedVersions,
							"requested": initParams.ProtocolVersion,
						},
					},
				}
			}
		}

		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType":      "complete",
				"protocolVersion": negotiatedVersion,
				"capabilities": map[string]any{
					"tools":     map[string]any{},
					"resources": map[string]any{},
					"prompts":   map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    MCPServerName,
					"version": MCPServerVersion,
				},
				"_meta": standardServerInfoMeta(),
			},
		}

	// Legacy ping support
	case "ping":
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType": "complete",
			},
		}

	// 4. Tools List (deterministic ordering per SEP-2549)
	case "tools/list":
		var mcpTools []map[string]any
		for _, tool := range s.dispatcher.Tools() {
			mcpTools = append(mcpTools, map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.Parameters,
			})
		}
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType": "complete",
				"tools":      mcpTools,
				"ttlMs":      3600000,
				"cacheScope": "public",
				"_meta":      standardServerInfoMeta(),
			},
		}

	// 5. Tools Call
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: errCodeInvalidParams, Message: fmt.Sprintf("invalid params: %v", err)},
			}
		}

		args := params.Arguments
		if len(args) == 0 {
			args = []byte(`{}`)
		}

		resStr, err := s.dispatcher.DispatchString(ctx, params.Name, args)
		if err != nil {
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"resultType": "complete",
					"content": []map[string]any{
						{
							"type": "text",
							"text": fmt.Sprintf("Error executing tool %q: %v", params.Name, err),
						},
					},
					"isError": true,
					"_meta":   standardServerInfoMeta(),
				},
			}
		}

		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType": "complete",
				"content": []map[string]any{
					{
						"type": "text",
						"text": resStr,
					},
				},
				"isError": false,
				"_meta":   standardServerInfoMeta(),
			},
		}

	// 6. Resources List
	case "resources/list":
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType": "complete",
				"resources": []map[string]any{
					{
						"uri":         "postcode://states",
						"name":        "Nigerian States & FCT Reference Registry",
						"description": "Complete directory of all 36 Nigerian states and the Federal Capital Territory with capitals, zones, and geographic centroids",
						"mimeType":    "application/json",
					},
					{
						"uri":         "postcode://lgas",
						"name":        "Nigerian Local Government Areas (LGAs) Reference Registry",
						"description": "Directory of registered Local Government Areas mapped across Nigerian states with codes and centroid coordinates",
						"mimeType":    "application/json",
					},
					{
						"uri":         "postcode://grammar",
						"name":        "Nigerian Postcode Grammar Specification",
						"description": "NIPOST standard specification for 11-digit alphanumeric postcode structure",
						"mimeType":    "text/markdown",
					},
				},
				"ttlMs":      3600000,
				"cacheScope": "public",
				"_meta":      standardServerInfoMeta(),
			},
		}

	// 7. Resources Read
	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: errCodeInvalidParams, Message: "invalid params: uri required"},
			}
		}

		switch params.URI {
		case "postcode://states":
			var states []StateRecord
			for _, rec := range NigerianStates {
				states = append(states, rec)
			}
			sort.Slice(states, func(i, j int) bool {
				return states[i].Code < states[j].Code
			})
			bytes, _ := json.MarshalIndent(states, "", "  ")
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"resultType": "complete",
					"contents": []map[string]any{
						{
							"uri":      params.URI,
							"mimeType": "application/json",
							"text":     string(bytes),
						},
					},
					"ttlMs":      3600000,
					"cacheScope": "public",
				},
			}

		case "postcode://lgas":
			var lgas []LGARecord
			for _, rec := range knownLGAs {
				lgas = append(lgas, rec)
			}
			slices.SortFunc(lgas, func(a, b LGARecord) int {
				if a.StateCode != b.StateCode {
					return cmp.Compare(a.StateCode, b.StateCode)
				}
				return cmp.Compare(a.LGACode, b.LGACode)
			})
			bytes, _ := json.MarshalIndent(lgas, "", "  ")
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"resultType": "complete",
					"contents": []map[string]any{
						{
							"uri":      params.URI,
							"mimeType": "application/json",
							"text":     string(bytes),
						},
					},
					"ttlMs":      3600000,
					"cacheScope": "public",
				},
			}

		case "postcode://grammar":
			grammarDoc := `# NIPOST National Digital Alphanumeric Postcode Specification

A Nigerian postcode consists of exactly 11 alphanumeric characters divided into 5 administrative segments:

1. **State** (2 letters, A-Z): Official 2-letter state code (e.g. EK, LA, FC).
2. **LGA** (2 digits, 01-99): Local Government Area code (cannot be 00).
3. **District** (3 alphanumeric chars): Administrative district (e.g. A03, W06).
4. **Area** (2 letters, A-Z): Street or neighborhood axis code (e.g. FK, TC).
5. **Building Unit** (2 digits, 01-99): Specific building or parcel unit (cannot be 00).

Canonical hyphenated representation: AA-99-A00-AA-99 (e.g. EK-01-A03-FK-01).
Spaced display representation: AA 99 A00 AA 99.
Compact representation: AA99A00AA99.
`
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"resultType": "complete",
					"contents": []map[string]any{
						{
							"uri":      params.URI,
							"mimeType": "text/markdown",
							"text":     grammarDoc,
						},
					},
					"ttlMs":      3600000,
					"cacheScope": "public",
				},
			}

		default:
			if strings.HasPrefix(params.URI, "postcode://lgas/") {
				st := strings.ToUpper(strings.TrimPrefix(params.URI, "postcode://lgas/"))
				items := StateLGAs(st)
				bytes, _ := json.MarshalIndent(items, "", "  ")
				return mcpResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]any{
						"resultType": "complete",
						"contents": []map[string]any{
							{
								"uri":      params.URI,
								"mimeType": "application/json",
								"text":     string(bytes),
							},
						},
						"ttlMs":      3600000,
						"cacheScope": "public",
					},
				}
			}
			// MCP 2026-07-28 uses -32602 (Invalid Params) for resource not found
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: errCodeInvalidParams, Message: fmt.Sprintf("resource not found: %s", params.URI)},
			}
		}

	// 8. Prompts List
	case "prompts/list":
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"resultType": "complete",
				"prompts": []map[string]any{
					{
						"name":        "normalize-nigerian-address",
						"description": "Expert instructions for extracting and resolving informal Nigerian addresses to canonical postcodes",
						"arguments": []map[string]any{
							{
								"name":        "raw_address",
								"description": "Informal address text (e.g. '14 Admiralty Way, Lekki Phase 1, Lagos')",
								"required":    true,
							},
						},
					},
					{
						"name":        "assemble-nigerian-postcode",
						"description": "Instructions for assembling and validating an 11-digit NIPOST postcode from individual administrative segments",
						"arguments": []map[string]any{
							{
								"name":        "state",
								"description": "2-letter state code (e.g. 'LA', 'EK', 'FC')",
								"required":    true,
							},
							{
								"name":        "lga",
								"description": "Local Government Area code or number (e.g. '01', '1', '11')",
								"required":    true,
							},
						},
					},
				},
				"ttlMs":      3600000,
				"cacheScope": "public",
			},
		}

	// 9. Prompts Get
	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: errCodeInvalidParams, Message: "invalid params"},
			}
		}

		if params.Name == "normalize-nigerian-address" {
			rawAddr := params.Arguments["raw_address"]
			promptText := fmt.Sprintf(`You are a Nigerian Address & Geocoding Expert.
Your goal is to normalize the following raw Nigerian address into an official 11-digit NIPOST postcode:
Address: %q

Follow these steps:
1. Extract the State (e.g. Lagos -> LA, Ekiti -> EK, Abuja -> FC).
2. Extract the Local Government Area (LGA) and district/neighborhood.
3. Call the 'validate_postcode' or 'lookup_postcode' tools to confirm exact validity.
4. If invalid or ambiguous, call 'diagnose_postcode' to receive segment-level feedback and suggestions.
5. Provide the user with the canonical format (e.g. LA-11-W06-TC-10) and a Google Maps confirmation link.`, rawAddr)

			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"resultType":  "complete",
					"description": "Nigerian address normalization instructions",
					"messages": []map[string]any{
						{
							"role": "user",
							"content": map[string]any{
								"type": "text",
								"text": promptText,
							},
						},
					},
				},
			}
		}

		if params.Name == "assemble-nigerian-postcode" {
			stateVal := params.Arguments["state"]
			lgaVal := params.Arguments["lga"]
			promptText := fmt.Sprintf(`You are an Assistant specialized in Nigerian Postcodes.
Your objective is to construct a canonical 11-digit postcode for:
State: %s
LGA: %s

Follow these instructions:
1. Call 'list_lgas' for state %q if you need to verify the exact 2-digit LGA code.
2. Determine or solicit the 3-character district, 2-letter area, and 2-digit building unit.
3. Call 'assemble_postcode' with the 5 segments (single digits will be zero-padded automatically).
4. Call 'validate_postcode' or 'lookup_postcode' to confirm accuracy.
5. If any validation error occurs, call 'diagnose_postcode' for actionable correction guidance.`, stateVal, lgaVal, stateVal)

			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"resultType":  "complete",
					"description": "Nigerian postcode assembly guidance",
					"messages": []map[string]any{
						{
							"role": "user",
							"content": map[string]any{
								"type": "text",
								"text": promptText,
							},
						},
					},
				},
			}
		}

		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: errCodeInvalidParams, Message: fmt.Sprintf("prompt not found: %s", params.Name)},
		}

	default:
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: errCodeMethodNotFound, Message: fmt.Sprintf("method not found: %s", req.Method)},
		}
	}
}

func (s *MCPServer) sendResponse(out io.Writer, resp mcpResponse) error {
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err = out.Write(append(bytes, '\n'))
	return err
}

func (s *MCPServer) sendError(out io.Writer, id json.RawMessage, code int, msg string, data any) {
	if len(id) == 0 {
		id = []byte("null")
	}
	resp := mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcpError{
			Code:    code,
			Message: msg,
			Data:    data,
		},
	}
	_ = s.sendResponse(out, resp)
}
