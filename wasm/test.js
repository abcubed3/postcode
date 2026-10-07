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

  console.log('\nAll parity verification tests passed successfully! 🎉');
}

run().catch((err) => {
  console.error('Test failed:', err);
  process.exit(1);
});
