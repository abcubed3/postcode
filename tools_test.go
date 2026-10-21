package postcode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestToolDeclarations(t *testing.T) {
	tools := DefaultAgentTools()
	if len(tools) == 0 {
		t.Fatalf("expected non-empty agent tools")
	}

	for _, tool := range tools {
		if tool.Name == "" || tool.Description == "" {
			t.Errorf("tool %q missing name or description", tool.Name)
		}

		// Test OpenAI format
		oai := tool.OpenAITool()
		if oai["type"] != "function" {
			t.Errorf("expected OpenAI type=function for %s", tool.Name)
		}

		// Test Anthropic format
		anth := tool.AnthropicTool()
		if anth["name"] != tool.Name {
			t.Errorf("expected Anthropic name=%s", tool.Name)
		}

		// Test Gemini format
		gem := tool.GeminiFunctionDeclaration()
		if gem["name"] != tool.Name {
			t.Errorf("expected Gemini name=%s", tool.Name)
		}
	}
}

func TestAgentDispatcherOffline(t *testing.T) {
	ctx := context.Background()
	dispatcher := NewAgentDispatcher(nil) // offline mode

	// 1. validate_postcode (valid)
	resVal, err := dispatcher.Dispatch(ctx, "validate_postcode", []byte(`{"code":"EK-01-A03-FK-01"}`))
	if err != nil {
		t.Fatalf("validate_postcode failed: %v", err)
	}
	valMap, ok := resVal.(map[string]any)
	if !ok || valMap["valid"] != true {
		t.Errorf("expected valid=true, got %+v", resVal)
	}
	if valMap["compact"] != "EK01A03FK01" {
		t.Errorf("expected compact EK01A03FK01, got %v", valMap["compact"])
	}

	// 2. validate_postcode (invalid)
	resInv, err := dispatcher.Dispatch(ctx, "validate_postcode", []byte(`{"code":"INVALID_CODE"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	invMap, ok := resInv.(map[string]any)
	if !ok || invMap["valid"] != false {
		t.Errorf("expected valid=false, got %+v", resInv)
	}

	// 2b. validate_postcode (temporary code)
	resTP, err := dispatcher.Dispatch(ctx, "validate_postcode", []byte(`{"code":"TP-01-A01-01"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tpMap, ok := resTP.(map[string]any)
	if !ok || tpMap["valid"] != false || tpMap["is_temporary"] != true {
		t.Errorf("expected is_temporary=true and valid=false, got %+v", resTP)
	}

	// 3. diagnose_postcode
	resDiag, err := dispatcher.Dispatch(ctx, "diagnose_postcode", []byte(`{"code":"EK 00 A03 FK 01"}`))
	if err != nil {
		t.Fatalf("diagnose_postcode failed: %v", err)
	}
	report, ok := resDiag.(DiagnosticReport)
	if !ok || report.Valid {
		t.Errorf("expected invalid report, got %+v", resDiag)
	}

	// 4. resolve_location offline
	resLoc, err := dispatcher.Dispatch(ctx, "resolve_location", []byte(`{"code":"EK-01-A03-FK-01"}`))
	if err != nil {
		t.Fatalf("resolve_location offline failed: %v", err)
	}
	loc, ok := resLoc.(*Location)
	if !ok || loc.StateCode != "EK" {
		t.Errorf("expected StateCode EK, got %+v", resLoc)
	}

	// 5. list_states
	resStates, err := dispatcher.Dispatch(ctx, "list_states", []byte(`{}`))
	if err != nil {
		t.Fatalf("list_states failed: %v", err)
	}
	states, ok := resStates.([]StateRecord)
	if !ok || len(states) != 37 {
		t.Errorf("expected 37 states, got %d", len(states))
	}

	// 6. assemble_postcode
	resAssemble, err := dispatcher.Dispatch(ctx, "assemble_postcode", []byte(`{"state":"ek","lga":"1","district":"a03","area":"fk","unit":"1"}`))
	if err != nil {
		t.Fatalf("assemble_postcode failed: %v", err)
	}
	asmb, ok := resAssemble.(AssembledPostcode)
	if !ok || asmb.Postcode != "EK-01-A03-FK-01" {
		t.Errorf("expected assembled EK-01-A03-FK-01, got %+v", resAssemble)
	}

	// 7. disassemble_postcode offline
	resDis, err := dispatcher.Dispatch(ctx, "disassemble_postcode", []byte(`{"code":"EK-01-A03-FK-01"}`))
	if err != nil {
		t.Fatalf("disassemble_postcode failed: %v", err)
	}
	segs, ok := resDis.(*Segments)
	if !ok || segs.State != "EK" || segs.LGA != "01" || segs.District != "A03" || segs.Area != "FK" || segs.Unit != "01" {
		t.Errorf("expected disassembled segments for EK-01-A03-FK-01, got %+v", resDis)
	}

	// 8. list_lgas offline
	resLGAs, err := dispatcher.Dispatch(ctx, "list_lgas", []byte(`{"state":"FC"}`))
	if err != nil {
		t.Fatalf("list_lgas failed: %v", err)
	}
	lgas, ok := resLGAs.([]NamedCode)
	if !ok || len(lgas) == 0 {
		t.Errorf("expected non-empty LGAs for FC, got %+v", resLGAs)
	}

	// 9. reverse_geocode offline
	resRev, err := dispatcher.Dispatch(ctx, "reverse_geocode", []byte(`{"latitude":7.6211,"longitude":5.2215,"max_distance_m":100}`))
	if err != nil {
		t.Fatalf("reverse_geocode offline failed: %v", err)
	}
	revResp, ok := resRev.(*ReverseResponse)
	if !ok || !revResp.Found {
		t.Errorf("expected found reverse geocode for Ado Ekiti coordinates, got %+v", resRev)
	}

	// 10. search_nearby offline
	resNearby, err := dispatcher.Dispatch(ctx, "search_nearby", []byte(`{"latitude":7.6211,"longitude":5.2215,"radius_m":500}`))
	if err != nil {
		t.Fatalf("search_nearby offline failed: %v", err)
	}
	nbResp, ok := resNearby.(*NearbyResponse)
	if !ok || len(nbResp.Results) == 0 {
		t.Errorf("expected nearby units found offline, got %+v", resNearby)
	}

	// 11. Online-only tool without client should return clear error
	_, errOnline := dispatcher.Dispatch(ctx, "lookup_postcode", []byte(`{"code":"EK-01-A03-FK-01"}`))
	if errOnline == nil || !strings.Contains(errOnline.Error(), "gateway client is required") {
		t.Errorf("expected gateway client required error, got %v", errOnline)
	}

	// 12. validate_postcode with unknown state code (e.g., ZZ)
	resZZ, err := dispatcher.Dispatch(ctx, "validate_postcode", []byte(`{"code":"ZZ-01-A03-FK-01"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	zzMap, ok := resZZ.(map[string]any)
	if !ok || zzMap["valid"] != false {
		t.Errorf("expected valid=false for unknown state ZZ, got %+v", resZZ)
	}

	// 13. repair_postcode
	resRepair, err := dispatcher.Dispatch(ctx, "repair_postcode", []byte(`{"code":"EK-1-AO3-FK-1"}`))
	if err != nil {
		t.Fatalf("repair_postcode failed: %v", err)
	}
	repMap, ok := resRepair.(map[string]any)
	if !ok || repMap["repaired"] != true || repMap["postcode"] != "EK-01-A03-FK-01" {
		t.Errorf("expected repaired EK-01-A03-FK-01, got %+v", resRepair)
	}

	// 14. extract_address
	resExt, err := dispatcher.Dispatch(ctx, "extract_address", []byte(`{"address":"Plot 14 Admiralty Way, beside Zenith Bank, Lekki Phase 1, Eti-Osa, Lagos 100001"}`))
	if err != nil {
		t.Fatalf("extract_address failed: %v", err)
	}
	ext, ok := resExt.(AddressExtraction)
	if !ok || ext.StateCode != "LA" || ext.LGACode != "08" || ext.LegacyPostcode != "100001" {
		t.Errorf("unexpected extract_address result: %+v", resExt)
	}

	// 15. calculate_distance
	resDist, err := dispatcher.Dispatch(ctx, "calculate_distance", []byte(`{"from_code":"EK-01-A03-FK-01","to_code":"LA-08-A86-RG-01"}`))
	if err != nil {
		t.Fatalf("calculate_distance failed: %v", err)
	}
	distMap, ok := resDist.(map[string]any)
	if !ok {
		t.Fatalf("expected map result from calculate_distance, got %T", resDist)
	}
	if distKm, ok := distMap["distance_km"].(float64); !ok || distKm < 150 || distKm > 350 {
		t.Errorf("expected distance_km ~218km, got %v", distMap["distance_km"])
	}

	// 16. generate_postcode_qr
	resQR, err := dispatcher.Dispatch(ctx, "generate_postcode_qr", []byte(`{"code":"EK-01-A03-FK-01","include_label":true}`))
	if err != nil {
		t.Fatalf("generate_postcode_qr failed: %v", err)
	}
	qrMap, ok := resQR.(map[string]any)
	if !ok || qrMap["valid"] != true || qrMap["postcode"] != "EK-01-A03-FK-01" {
		t.Errorf("unexpected generate_postcode_qr result: %+v", resQR)
	}
	if svgStr, _ := qrMap["svg"].(string); !strings.HasPrefix(svgStr, "<svg ") {
		t.Errorf("expected valid SVG string in generate_postcode_qr result")
	}

	// 17. DispatchString
	strRes, err := dispatcher.DispatchString(ctx, "validate_postcode", []byte(`{"code":"EK-01-A03-FK-01"}`))
	if err != nil {
		t.Fatalf("DispatchString failed: %v", err)
	}
	if !strings.Contains(strRes, `"valid":true`) {
		t.Errorf("expected valid:true in JSON string, got %s", strRes)
	}
}

func TestClientWithHandlerAndCacheNormalization(t *testing.T) {
	lookupHits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/lookup", func(w http.ResponseWriter, r *http.Request) {
		lookupHits++
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "999")
		w.Header().Set("X-Credit-Balance", "4995")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"postcode": "EK-01-A03-FK-01",
				"valid":    true,
				"level":    2,
			},
		})
	})

	client, err := NewClient(
		WithHandler(mux),
		WithAPIKey("test_key"),
		WithDefaultCache(),
	)
	if err != nil {
		t.Fatalf("NewClient with WithHandler failed: %v", err)
	}

	ctx := context.Background()
	// First call: spaced lowercase Level 2
	resp1, err := client.Lookup(ctx, "ek 01 a03 fk 01", Level2)
	if err != nil || !resp1.Valid {
		t.Fatalf("Lookup 1 failed: %v", err)
	}
	rl := client.RateLimit()
	if rl == nil || rl.CreditBalance != 4995 {
		t.Errorf("expected CreditBalance=4995, got %+v", rl)
	}

	// Second call: compact uppercase Level 1 (should hit cumulative normalized cache!)
	resp2, err := client.Lookup(ctx, "EK01A03FK01", Level1)
	if err != nil || !resp2.Valid {
		t.Fatalf("Lookup 2 failed: %v", err)
	}
	if lookupHits != 1 {
		t.Errorf("expected 1 HTTP lookup due to normalized cumulative cache hit, got %d", lookupHits)
	}
}

func TestAgentDispatcherOnline(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/lookup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"postcode": "EK-01-A03-FK-01",
				"valid":    true,
			},
		})
	})
	mux.HandleFunc("/v1/search/autocomplete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"segment": "district",
				"suggestions": []map[string]string{
					{"code": "A03", "label": "District A03"},
				},
			},
		})
	})
	mux.HandleFunc("/v1/search/reverse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"found": true,
				"unit": map[string]any{
					"postcode": "EK-01-A03-FK-01",
				},
			},
		})
	})
	mux.HandleFunc("/v1/search/nearby", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"results": []map[string]any{
					{"postcode": "EK-01-A03-FK-01", "distance_m": 12.5},
				},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewClient(
		WithBaseURL(srv.URL),
		WithAPIKey("test_key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	dispatcher := NewAgentDispatcher(client)

	// 1. lookup_postcode
	resLookup, err := dispatcher.Dispatch(ctx, "lookup_postcode", []byte(`{"code":"EK-01-A03-FK-01","level":1}`))
	if err != nil {
		t.Fatalf("lookup_postcode failed: %v", err)
	}
	lookup, ok := resLookup.(*LookupResponse)
	if !ok || !lookup.Valid {
		t.Errorf("expected valid lookup, got %+v", resLookup)
	}

	// 2. autocomplete_postcode
	resAuto, err := dispatcher.Dispatch(ctx, "autocomplete_postcode", []byte(`{"query":"EK 01 A"}`))
	if err != nil {
		t.Fatalf("autocomplete_postcode failed: %v", err)
	}
	auto, ok := resAuto.(*AutocompleteResponse)
	if !ok || len(auto.Suggestions) == 0 {
		t.Errorf("expected suggestions, got %+v", resAuto)
	}

	// 3. reverse_geocode
	resRev, err := dispatcher.Dispatch(ctx, "reverse_geocode", []byte(`{"latitude":7.6211,"longitude":5.2215}`))
	if err != nil {
		t.Fatalf("reverse_geocode failed: %v", err)
	}
	rev, ok := resRev.(*ReverseResponse)
	if !ok || !rev.Found {
		t.Errorf("expected found=true, got %+v", resRev)
	}

	// 4. search_nearby with coordinates
	resNear, err := dispatcher.Dispatch(ctx, "search_nearby", []byte(`{"latitude":7.6211,"longitude":5.2215,"radius_m":300}`))
	if err != nil {
		t.Fatalf("search_nearby failed: %v", err)
	}
	near, ok := resNear.(*NearbyResponse)
	if !ok || len(near.Results) == 0 {
		t.Errorf("expected nearby results, got %+v", resNear)
	}

	// 5. search_nearby with reference postcode
	resNearCode, err := dispatcher.Dispatch(ctx, "search_nearby", []byte(`{"code":"EK-01-A03-FK-01","radius_m":300}`))
	if err != nil {
		t.Fatalf("search_nearby with code failed: %v", err)
	}
	nearCode, ok := resNearCode.(*NearbyResponse)
	if !ok || len(nearCode.Results) == 0 {
		t.Errorf("expected nearby results with code, got %+v", resNearCode)
	}
}

func TestAgentDispatcherErrors(t *testing.T) {
	ctx := context.Background()
	dispatcher := NewAgentDispatcher(nil)

	// Unknown tool
	_, err := dispatcher.Dispatch(ctx, "unknown_tool", []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "unknown tool name") {
		t.Errorf("expected unknown tool error, got %v", err)
	}

	// Malformed JSON arguments
	_, errJSON := dispatcher.Dispatch(ctx, "validate_postcode", []byte(`{invalid-json`))
	if errJSON == nil {
		t.Errorf("expected unmarshal error for malformed json")
	}

	// DispatchString on error
	strErr, err := dispatcher.DispatchString(ctx, "unknown_tool", []byte(`{}`))
	if err == nil {
		t.Errorf("expected error returned by DispatchString")
	}
	var errMap map[string]string
	if errUnmarshal := json.Unmarshal([]byte(strErr), &errMap); errUnmarshal != nil || errMap["error"] == "" {
		t.Errorf("expected JSON error payload, got %s", strErr)
	}
}
