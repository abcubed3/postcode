const { initPostcode } = require('./index.js');
const assert = require('assert');

async function run() {
  console.log('Initializing WebAssembly Postcode Engine...');
  const postcode = await initPostcode();

  // 1. Instant validation & parsing
  const valid = postcode.validate('EK 01 A03 FK 01');
  assert.strictEqual(valid.valid, true);
  assert.strictEqual(valid.state_name, 'Ekiti');
  console.log('✔ validate() passed');

  const parsed = postcode.parse('EK 01 A03 FK 01');
  assert.strictEqual(parsed.valid, true);
  assert.strictEqual(parsed.state_code, 'EK');
  assert.strictEqual(parsed.lga_code, '01');
  console.log('✔ parse() passed');

  const states = postcode.listStates();
  assert(states['EK'] !== undefined);
  assert.strictEqual(states['EK'].name, 'Ekiti');
  console.log('✔ listStates() passed');

  // 2. Batch validation
  const batch = postcode.validateBatch(['EK 01 A03 FK 01', 'LA 11 W06 TC 10', 'INVALID']);
  assert.strictEqual(batch.length, 3);
  assert.strictEqual(batch[0].valid, true);
  assert.strictEqual(batch[1].valid, true);
  assert.strictEqual(batch[2].valid, false);
  console.log('✔ validateBatch() passed');

  // 3. Formatting
  const compact = postcode.format('EK 01 A03 FK 01', 'compact');
  const hyphenated = postcode.format('EK01A03FK01', 'hyphenated');
  const slug = postcode.format('EK 01 A03 FK 01', 'slug');
  assert.strictEqual(compact, 'EK01A03FK01');
  assert.strictEqual(hyphenated, 'EK-01-A03-FK-01');
  assert.strictEqual(slug, 'ek-01-a03-fk-01');
  console.log('✔ format() passed');

  // 4. Disassemble and Assemble
  const parts = postcode.disassemble('EK-01-A03-FK-01');
  assert.strictEqual(parts.state, 'EK');
  assert.strictEqual(parts.lga, '01');
  assert.strictEqual(parts.district, 'A03');
  assert.strictEqual(parts.area, 'FK');
  assert.strictEqual(parts.unit, '01');
  const assembled = postcode.assemble(parts);
  assert.strictEqual(assembled.valid, true);
  assert.strictEqual(assembled.postcode, 'EK-01-A03-FK-01');
  assert.strictEqual(assembled.display, 'EK 01 A03 FK 01');
  assert.strictEqual(assembled.compact, 'EK01A03FK01');
  console.log('✔ disassemble() & assemble() passed');

  // 5. Format diagnostics with actionable suggestions
  const diag = postcode.diagnose('ZZ 00 A03 FK 00');
  assert.strictEqual(typeof diag.format_score, 'number');
  assert(diag.actionable_tip.length > 0);
  console.log('✔ diagnose() passed');

  // 6. Offline geocoding & mapping
  const loc = postcode.resolveLocation('EK 01 A03 FK 01');
  assert.strictEqual(loc.state, 'Ekiti');
  assert(loc.latitude > 0 && loc.longitude > 0);
  assert(loc.google_maps_directions_url.includes('destination='));
  assert(loc.apple_maps_url.includes('maps.apple.com'));
  assert(loc.search_query.includes('Ekiti'));
  console.log('✔ resolveLocation() passed');

  // 7. API Key Configuration & Updates
  postcode.setAPIKey('nipost_live_test_key_123');
  assert.strictEqual(postcode.getAPIKey(), 'nipost_live_test_key_123');
  postcode.configure({ apiKey: 'nipost_live_updated_456', baseURL: 'https://api.postcode.gov.ng' });
  assert.strictEqual(postcode.getAPIKey(), 'nipost_live_updated_456');
  console.log('✔ configure() & setAPIKey() passed');

  // 8. AI Agent Protocol (OpenAI, Anthropic, Gemini schemas)
  const openaiTools = postcode.getAgentTools('openai');
  assert(Array.isArray(openaiTools) && openaiTools.length > 0);
  assert.strictEqual(openaiTools[0].type, 'function');
  const anthropicTools = postcode.getAgentTools('anthropic');
  assert(Array.isArray(anthropicTools) && anthropicTools.length > 0);
  assert(anthropicTools[0].input_schema !== undefined);
  const geminiTools = postcode.getAgentTools('gemini');
  assert(Array.isArray(geminiTools) && geminiTools.length > 0);
  console.log('✔ getAgentTools() passed for openai, anthropic, gemini');

  // 9. AI Agent Tool Execution
  const toolResult = await postcode.executeTool('validate_postcode', { code: 'EK 01 A03 FK 01' });
  assert.strictEqual(toolResult.valid, true);
  assert.strictEqual(toolResult.state_name, 'Ekiti');
  assert.strictEqual(toolResult.state_code, 'EK');
  console.log('✔ executeTool() passed');

  // 10. AI Agent Metrics & Guardrails
  const metrics = postcode.getAgentMetrics();
  assert.strictEqual(typeof metrics.total_calls, 'number');
  assert.strictEqual(typeof metrics.commercial_calls, 'number');
  console.log('✔ getAgentMetrics() passed');

  // 11. Synthetic Benchmark Generator
  const synthetic = postcode.generateSyntheticAddresses({ count: 5 });
  assert.strictEqual(synthetic.length, 5);
  assert(synthetic[0].raw_text.length > 0);
  assert(synthetic[0].expected_state.length > 0);
  assert(synthetic[0].expected_state_code.length > 0);
  console.log('✔ generateSyntheticAddresses() passed');

  // 12. Flexible Nearby & Reverse Geocoding API Signatures
  // Verify argument validation on invalid input
  await assert.rejects(
    async () => { await postcode.nearby('INVALID_CODE'); },
    /invalid reference postcode/
  );
  await assert.rejects(
    async () => { await postcode.reverseGeocode(0, 0); },
    /latitude and longitude coordinates are required/
  );
  await assert.rejects(
    async () => { await postcode.reverseGeocode(); },
    /parameters required/
  );
  console.log('✔ nearby() & reverseGeocode() flexible signature validation passed');

  // 13. Single-digit Segment Normalization & Padding in assemble()
  const unpadded = postcode.assemble({ state: 'ek', lga: '1', district: 'a03', area: 'fk', unit: '1' });
  assert.strictEqual(unpadded.valid, true);
  assert.strictEqual(unpadded.postcode, 'EK-01-A03-FK-01');
  assert.strictEqual(unpadded.display, 'EK 01 A03 FK 01');
  assert.strictEqual(unpadded.compact, 'EK01A03FK01');

  const normalized = postcode.normalizeSegments({ state: 'ek', lga: '1', district: 'a03', area: 'fk', unit: '1' });
  assert.strictEqual(normalized.state, 'EK');
  assert.strictEqual(normalized.lga, '01');
  assert.strictEqual(normalized.district, 'A03');
  assert.strictEqual(normalized.area, 'FK');
  assert.strictEqual(normalized.unit, '01');
  console.log('✔ assemble() single-digit padding & normalizeSegments() passed');

  // 14. Offline Reference Catalogs (States and LGAs)
  const refStates = postcode.referenceStatesOffline();
  assert.strictEqual(refStates.length, 37);
  assert.strictEqual(refStates[0].code, 'AB');
  assert.strictEqual(refStates[0].name, 'Abia');

  const refLGAs = postcode.referenceLGAsOffline('LA');
  assert(refLGAs.length >= 20);
  assert.strictEqual(refLGAs[0].code, '01');

  const stateLGAs = postcode.stateLGAs('LA');
  assert.strictEqual(stateLGAs.length, refLGAs.length);
  console.log('✔ referenceStatesOffline() & referenceLGAsOffline() passed');

  // 15. Offline Spatial Geocoding & Radius Search
  const nearbyOffline = postcode.searchNearbyBuildingsOffline(7.6211, 5.2215, 300);
  assert(Array.isArray(nearbyOffline) && nearbyOffline.length > 0);
  assert.strictEqual(nearbyOffline[0].postcode, 'EK-01-A03-FK-01');

  const reverseOffline = postcode.reverseCoordinatesOffline(7.6211, 5.2215, 50);
  assert.strictEqual(reverseOffline.found, true);
  assert.strictEqual(reverseOffline.unit.postcode, 'EK-01-A03-FK-01');
  console.log('✔ searchNearbyBuildingsOffline() & reverseCoordinatesOffline() passed');

  // 16. Google Maps Geocoding Configuration
  postcode.setGoogleMapsAPIKey('test_gmaps_key_123');
  assert.strictEqual(postcode.getGoogleMapsAPIKey(), 'test_gmaps_key_123');
  postcode.configure({ googleMapsApiKey: 'updated_gmaps_key_456' });
  assert.strictEqual(postcode.getGoogleMapsAPIKey(), 'updated_gmaps_key_456');
  console.log('✔ setGoogleMapsAPIKey() & configure({googleMapsApiKey}) passed');

  // 17. Async Reference Catalogs (with offline fallback)
  const asyncStates = await postcode.referenceStates();
  assert.strictEqual(asyncStates.length, 37);
  const asyncLGAs = await postcode.referenceLGAs('LA');
  assert(asyncLGAs.length >= 20);
  console.log('✔ async referenceStates() & referenceLGAs() passed');

  // 18. AI Agent Tool: assemble_postcode
  const toolAssembled = await postcode.executeTool('assemble_postcode', {
    state: 'ek', lga: '1', district: 'a03', area: 'fk', unit: '1'
  });
  assert.strictEqual(toolAssembled.postcode, 'EK-01-A03-FK-01');
  console.log('✔ executeTool("assemble_postcode") passed');

  // 19. AI Agent Tools: disassemble_postcode & list_lgas
  const toolDis = await postcode.executeTool('disassemble_postcode', { code: 'EK-01-A03-FK-01' });
  assert.strictEqual(toolDis.state, 'EK');
  assert.strictEqual(toolDis.district, 'A03');

  const toolLGAs = await postcode.executeTool('list_lgas', { state: 'LA' });
  assert(Array.isArray(toolLGAs) && toolLGAs.length >= 20);
  console.log('✔ executeTool("disassemble_postcode") & executeTool("list_lgas") passed');

  // 20. Auto-Repair, Address Extraction, and Distance Calculation
  const repaired = postcode.repair('EK-1-AO3-FK-1');
  assert.strictEqual(repaired.repaired, true);
  assert.strictEqual(repaired.postcode, 'EK-01-A03-FK-01');
  assert.strictEqual(repaired.slug, 'ek-01-a03-fk-01');

  const extracted = postcode.extractAddress('Plot 14 Admiralty Way, beside Zenith Bank, Lekki Phase 1, Eti-Osa, Lagos 100001');
  assert.strictEqual(extracted.state_code, 'LA');
  assert.strictEqual(extracted.lga_code, '08');
  assert.strictEqual(extracted.legacy_postcode, '100001');

  const dist = postcode.distanceMeters('EK-01-A03-FK-01', 'LA-08-A86-RG-01');
  assert.strictEqual(dist.valid, true);
  assert(dist.distance_km > 150 && dist.distance_km < 350);
  console.log('✔ repair(), extractAddress(), & distanceMeters() passed');

  // 21. QR Code Generation (SVG, Printable Label, PNG Data URI, Terminal, and MCP Tool)
  const qrDefault = postcode.generateQR('EK 01 A03 FK 01');
  assert.strictEqual(qrDefault.valid, true);
  assert.strictEqual(qrDefault.postcode, 'EK-01-A03-FK-01');
  assert.strictEqual(qrDefault.payload, 'EK-01-A03-FK-01');
  assert(qrDefault.svg.startsWith('<svg '));
  assert(qrDefault.data_uri.startsWith('data:image/png;base64,'));
  assert(qrDefault.terminal.length > 0);

  const qrLabel = postcode.qr('LA-11-W06-TC-10', {
    payload: 'google-maps',
    includeLabel: true,
    title: 'NIPOST WAYBILL',
  });
  assert.strictEqual(qrLabel.valid, true);
  assert.strictEqual(qrLabel.payload_type, 'google-maps');
  assert(qrLabel.svg.includes('NIPOST WAYBILL'));
  assert(qrLabel.printable_html.includes('<!DOCTYPE html>'));

  const qrTool = await postcode.executeTool('generate_postcode_qr', {
    code: 'FC-03-B06-AG-12',
    include_label: true,
  });
  assert.strictEqual(qrTool.valid, true);
  assert.strictEqual(qrTool.postcode, 'FC-03-B06-AG-12');
  assert(qrTool.svg.startsWith('<svg '));
  console.log('✔ generateQR(), qr(), & executeTool("generate_postcode_qr") passed');

  // 22. Temporary Code (TP_... / TP-...) Awareness
  assert.strictEqual(postcode.isTemporary('TP-01-A01-01'), true);
  assert.strictEqual(postcode.isTemporary('TP_01_A01_01'), true);
  assert.strictEqual(postcode.isTemporary('tp-123456'), true);
  assert.strictEqual(postcode.isTemporary('TP 01 A01 01'), true);
  assert.strictEqual(postcode.isTemporary('EK-01-A03-FK-01'), false);
  assert.strictEqual(postcode.isTemporary('100001'), false);

  assert.strictEqual(postcode.formatTemporary('tp_01_a01_01'), 'TP-01-A01-01');
  assert.strictEqual(postcode.formatTemporary('tp-123456'), 'TP-123456');

  const tpDiag = postcode.diagnose('TP_01_A01_01');
  assert.strictEqual(tpDiag.valid, false);
  assert.strictEqual(tpDiag.is_temporary, true);
  assert.strictEqual(tpDiag.format_score, 50);
  assert.strictEqual(tpDiag.diagnoses[0].segment, 'TemporaryCode');
  assert(tpDiag.actionable_tip.includes('GET /v1/lookup'));

  const tpVal = postcode.validate('TP-01-A01-01');
  assert.strictEqual(tpVal.valid, false);
  assert.strictEqual(tpVal.is_temporary, true);

  const tpTool = await postcode.executeTool('validate_postcode', { code: 'TP_01_A01_01' });
  assert.strictEqual(tpTool.valid, false);
  assert.strictEqual(tpTool.is_temporary, true);
  console.log('✔ Temporary Code (TP-...) awareness passed');

  // 23. Zero-dependency PNG Data URI Verification & Concurrent Loader Deduplication
  const zlib = require('zlib');
  const pngB64 = qrDefault.data_uri.slice('data:image/png;base64,'.length);
  const pngBuf = Buffer.from(pngB64, 'base64');
  // Verify PNG 8-byte magic header: 89 50 4E 47 0D 0A 1A 0A
  assert.strictEqual(pngBuf.subarray(0, 8).toString('hex'), '89504e470d0a1a0a');
  assert.strictEqual(pngBuf.subarray(12, 16).toString('ascii'), 'IHDR');
  // Locate IDAT chunk and verify zlib decompression succeeds
  const idatIdx = pngBuf.indexOf(Buffer.from('IDAT', 'ascii'));
  assert(idatIdx > 0);
  const idatLen = pngBuf.readUInt32BE(idatIdx - 4);
  const idatPayload = pngBuf.subarray(idatIdx + 4, idatIdx + 4 + idatLen);
  const inflated = zlib.inflateSync(idatPayload);
  assert(inflated.length > 0);

  // Verify concurrent initPostcode() calls return the same singleton instance
  const [p1, p2] = await Promise.all([initPostcode(), initPostcode()]);
  assert.strictEqual(p1, postcode);
  assert.strictEqual(p2, postcode);
  console.log('✔ PNG binary integrity & concurrent initPostcode() deduplication passed');

  console.log('\nAll parity verification tests passed successfully! 🎉');
}

run()
  .then(() => {
    process.exit(0);
  })
  .catch((err) => {
    console.error('Test failed:', err);
    process.exit(1);
  });
