/**
 * Demonstration of consuming the Nigerian Postcode WebAssembly module in Node.js.
 * Provides microsecond, offline-first geocoding, format diagnosis, and validation.
 */

const { initPostcode } = require('../../wasm/index.js');

async function main() {
  console.log('==================================================================');
  console.log('  Postcode Nigeria — WebAssembly (WASM) Edge Runtime Demo');
  console.log('==================================================================');

  // Initialize the WebAssembly runtime
  const postcode = await initPostcode();
  console.log('✅ WebAssembly engine loaded successfully.\n');

  // 1. Instant Offline Validation
  const testCodes = ['EK 01 A03 FK 01', 'LA 11 W06 TC 10', 'INVALID_CODE'];
  console.log('1. Offline Validation:');
  for (const code of testCodes) {
    const res = postcode.validate(code);
    if (res.valid) {
      console.log(`   ✓ ${code} -> State: ${res.state_name} (${res.state_code}), Canonical: ${res.postcode}`);
    } else {
      console.log(`   ✗ ${code} -> Error: ${res.error}`);
    }
  }

  // 2. Intelligent Format Diagnosis & Typo Suggestion
  console.log('\n2. Diagnostic Engine with Actionable Agent Tips:');
  const malformed = 'ZZ 00 A03 FK 00';
  const diag = postcode.diagnose(malformed);
  console.log(`   Input:            ${diag.input}`);
  console.log(`   Structural Score: ${diag.format_score} / 100`);
  console.log(`   Actionable Tip:   ${diag.actionable_tip}`);
  for (const d of diag.diagnoses || []) {
    console.log(`     • [${d.segment}] ${d.message}`);
    if (d.suggestions && d.suggestions.length > 0) {
      console.log(`       Suggestions: ${d.suggestions.join(', ')}`);
    }
  }

  // 3. Offline Geolocation & Coordinates
  console.log('\n3. High-Precision Offline Geocoding:');
  const loc = postcode.resolveLocation('EK 01 A03 FK 01');
  console.log(`   Postcode:    ${loc.postcode}`);
  console.log(`   Coordinates: Latitude=${loc.latitude}, Longitude=${loc.longitude}`);
  console.log(`   State / LGA: ${loc.state_name} / ${loc.lga_name || 'Centroid'}`);
  console.log(`   Google Maps: ${loc.google_maps_url}`);

  console.log('\n==================================================================');
  console.log('  Edge WASM execution completed with 0 network requests!');
  console.log('==================================================================');
}

main().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
