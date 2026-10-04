const { initPostcode } = require('./index.js');

async function run() {
  const postcode = await initPostcode();

  // 1. Instant validation
  const valid = postcode.validate('EK 01 A03 FK 01');
  console.log('Valid:', valid.valid);
  console.log('State:', valid.state_name);

  // 2. Format diagnostics with actionable suggestions
  const diag = postcode.diagnose('ZZ 00 A03 FK 00');
  console.log('Format score:', diag.format_score);
  console.log('Actionable tip:', diag.actionable_tip);

  // 3. Offline geocoding & mapping
  const loc = postcode.resolveLocation('EK 01 A03 FK 01');
  console.log('Coordinates:', loc.latitude, loc.longitude);
  console.log('Maps URL:', loc.google_maps_url);
}

run().catch(console.error);
