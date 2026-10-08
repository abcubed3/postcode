package postcode

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestMCPServerLifecycle_Modern(t *testing.T) {
	server := NewMCPServer(nil) // offline mode

	// Construct simulated JSON-RPC client messages per MCP 2026-07-28 spec
	messages := []string{
		// 1. Mandatory server/discover RPC
		`{"jsonrpc":"2.0","id":"disc-1","method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`,
		// 2. Stateless tools/list with modern _meta
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`,
		// 3. Stateless tools/call
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"validate_postcode","arguments":{"code":"EK-01-A03-FK-01"},"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`,
		// 4. Stateless resources/list
		`{"jsonrpc":"2.0","id":4,"method":"resources/list"}`,
		// 5. Stateless resources/read
		`{"jsonrpc":"2.0","id":5,"method":"resources/read","params":{"uri":"postcode://states"}}`,
		// 6. Resource not found -> code -32602
		`{"jsonrpc":"2.0","id":6,"method":"resources/read","params":{"uri":"postcode://invalid_uri"}}`,
		// 7. Unsupported protocol version -> code -32022
		`{"jsonrpc":"2.0","id":7,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"1999-01-01"}}}`,
	}

	inputBuf := bytes.NewBufferString(strings.Join(messages, "\n") + "\n")
	outputBuf := &bytes.Buffer{}

	err := server.Serve(inputBuf, outputBuf)
	if err != nil && err != io.EOF {
		t.Fatalf("server.Serve error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outputBuf.String()), "\n")
	if len(lines) != 7 {
		t.Fatalf("expected 7 responses, got %d. Output:\n%s", len(lines), outputBuf.String())
	}

	parseResp := func(line string) map[string]any {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid json line %q: %v", line, err)
		}
		return m
	}

	// 1. Verify server/discover
	r1 := parseResp(lines[0])
	res1 := r1["result"].(map[string]any)
	if res1["resultType"] != "complete" {
		t.Errorf("expected resultType=complete, got %v", res1["resultType"])
	}
	versions := res1["supportedVersions"].([]any)
	if len(versions) == 0 || versions[0] != "2026-07-28" {
		t.Errorf("expected latest version 2026-07-28, got %+v", versions)
	}
	if res1["ttlMs"].(float64) <= 0 || res1["cacheScope"] != "public" {
		t.Errorf("expected valid caching hints on server/discover, got %+v", res1)
	}

	// 2. Verify tools/list deterministic sorting & caching hints
	r2 := parseResp(lines[1])
	res2 := r2["result"].(map[string]any)
	if res2["resultType"] != "complete" {
		t.Errorf("expected resultType=complete on tools/list")
	}
	tools := res2["tools"].([]any)
	for i := 1; i < len(tools); i++ {
		prev := tools[i-1].(map[string]any)["name"].(string)
		curr := tools[i].(map[string]any)["name"].(string)
		if prev > curr {
			t.Errorf("tools are not deterministically sorted: %s came after %s", curr, prev)
		}
	}

	// 3. Verify tools/call
	r3 := parseResp(lines[2])
	res3 := r3["result"].(map[string]any)
	if res3["resultType"] != "complete" || res3["isError"] == true {
		t.Errorf("expected valid tools/call result: %+v", res3)
	}

	// 4. Verify resources/list
	r4 := parseResp(lines[3])
	res4 := r4["result"].(map[string]any)
	if res4["resultType"] != "complete" {
		t.Errorf("expected resultType=complete on resources/list")
	}

	// 5. Verify resources/read
	r5 := parseResp(lines[4])
	res5 := r5["result"].(map[string]any)
	if res5["resultType"] != "complete" {
		t.Errorf("expected resultType=complete on resources/read")
	}

	// 6. Verify resource not found code -32602
	r6 := parseResp(lines[5])
	err6 := r6["error"].(map[string]any)
	if int(err6["code"].(float64)) != -32602 {
		t.Errorf("expected error code -32602 for resource not found, got %v", err6["code"])
	}

	// 7. Verify unsupported protocol version code -32022
	r7 := parseResp(lines[6])
	err7 := r7["error"].(map[string]any)
	if int(err7["code"].(float64)) != -32022 {
		t.Errorf("expected error code -32022 for unsupported version, got %v", err7["code"])
	}
}

func TestMCPServerLifecycle_LegacyCompatibility(t *testing.T) {
	server := NewMCPServer(nil) // offline mode

	// Legacy client handshake (initialize -> notifications/initialized -> tools/list)
	messages := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
	}

	inputBuf := bytes.NewBufferString(strings.Join(messages, "\n") + "\n")
	outputBuf := &bytes.Buffer{}

	err := server.Serve(inputBuf, outputBuf)
	if err != nil && err != io.EOF {
		t.Fatalf("server.Serve error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outputBuf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 responses for legacy client, got %d. Output:\n%s", len(lines), outputBuf.String())
	}

	var r1 map[string]any
	_ = json.Unmarshal([]byte(lines[0]), &r1)
	res1 := r1["result"].(map[string]any)
	if res1["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected negotiated legacy version 2024-11-05, got %v", res1["protocolVersion"])
	}
}

func TestMCPServer_ExtendedCapabilities(t *testing.T) {
	server := NewMCPServer(nil) // offline mode

	messages := []string{
		// 1. tools/call assemble_postcode
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"assemble_postcode","arguments":{"state":"ek","lga":"1","district":"a03","area":"fk","unit":"1"}}}`,
		// 2. tools/call disassemble_postcode
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"disassemble_postcode","arguments":{"code":"EK-01-A03-FK-01"}}}`,
		// 3. tools/call list_lgas
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_lgas","arguments":{"state":"LA"}}}`,
		// 4. resources/read postcode://lgas
		`{"jsonrpc":"2.0","id":4,"method":"resources/read","params":{"uri":"postcode://lgas"}}`,
		// 5. resources/read postcode://lgas/LA
		`{"jsonrpc":"2.0","id":5,"method":"resources/read","params":{"uri":"postcode://lgas/LA"}}`,
		// 6. prompts/list
		`{"jsonrpc":"2.0","id":6,"method":"prompts/list"}`,
		// 7. prompts/get assemble-nigerian-postcode
		`{"jsonrpc":"2.0","id":7,"method":"prompts/get","params":{"name":"assemble-nigerian-postcode","arguments":{"state":"LA","lga":"01"}}}`,
	}

	inputBuf := bytes.NewBufferString(strings.Join(messages, "\n") + "\n")
	outputBuf := &bytes.Buffer{}

	err := server.Serve(inputBuf, outputBuf)
	if err != nil && err != io.EOF {
		t.Fatalf("server.Serve error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outputBuf.String()), "\n")
	if len(lines) != 7 {
		t.Fatalf("expected 7 responses, got %d. Output:\n%s", len(lines), outputBuf.String())
	}

	parseResp := func(line string) map[string]any {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid json line %q: %v", line, err)
		}
		return m
	}

	// 1. assemble_postcode response check
	r1 := parseResp(lines[0])
	res1 := r1["result"].(map[string]any)
	content1 := res1["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(content1, "EK-01-A03-FK-01") {
		t.Errorf("expected assembled code in response, got %s", content1)
	}

	// 2. disassemble_postcode response check
	r2 := parseResp(lines[1])
	res2 := r2["result"].(map[string]any)
	content2 := res2["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(content2, `"state":"EK"`) || !strings.Contains(content2, `"district":"A03"`) {
		t.Errorf("expected disassembled segments, got %s", content2)
	}

	// 3. list_lgas response check
	r3 := parseResp(lines[2])
	res3 := r3["result"].(map[string]any)
	content3 := res3["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(content3, "Ikeja") && !strings.Contains(content3, "Agege") {
		t.Errorf("expected Lagos LGAs in response, got %s", content3)
	}

	// 4. resources/read postcode://lgas
	r4 := parseResp(lines[3])
	res4 := r4["result"].(map[string]any)
	text4 := res4["contents"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text4, "Ado Ekiti") {
		t.Errorf("expected LGAs registry content, got %s", text4)
	}

	// 5. resources/read postcode://lgas/LA
	r5 := parseResp(lines[4])
	res5 := r5["result"].(map[string]any)
	text5 := res5["contents"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text5, "Ikeja") {
		t.Errorf("expected Lagos state LGAs content, got %s", text5)
	}

	// 6. prompts/list check
	r6 := parseResp(lines[5])
	res6 := r6["result"].(map[string]any)
	prompts := res6["prompts"].([]any)
	if len(prompts) < 2 {
		t.Errorf("expected at least 2 prompts, got %d", len(prompts))
	}

	// 7. prompts/get assemble-nigerian-postcode
	r7 := parseResp(lines[6])
	res7 := r7["result"].(map[string]any)
	msgs := res7["messages"].([]any)
	promptBody := msgs[0].(map[string]any)["content"].(map[string]any)["text"].(string)
	if !strings.Contains(promptBody, "assemble_postcode") {
		t.Errorf("expected assemble prompt body, got %s", promptBody)
	}
}
