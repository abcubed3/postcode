# Postcode Nigeria WebAssembly (WASM)

Zero-dependency, offline-first Nigerian Postcode engine compiled to WebAssembly. Enables AI agents, edge workers, mobile apps, and web browsers to perform high-speed validation, diagnostics, and geocoding without network overhead.

## Features

- **Blazing Fast**: Microsecond execution on edge runtimes (Cloudflare Workers, Fastly, Vercel Edge, AWS Lambda@Edge).
- **Zero Network Overhead**: Performs full validation, format repair tips, and geocoding offline.
- **Cross-Platform**: Runs in Node.js, Deno, Bun, web browsers, and Python (via `wasmtime` / `wasmer`).

## Building the WASM Binary

```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o wasm/postcode.wasm ./wasm
```

## Quick Start (Node.js / Edge)

```javascript
const { initPostcode } = require('./wasm/index.js');

async function main() {
  const postcode = await initPostcode();

  // 1. Validate Postcode
  const valid = postcode.validate('EK 01 A03 FK 01');
  console.log(valid.valid); // true
  console.log(valid.state_name); // 'Ekiti'

  // 2. Intelligent Diagnostic Engine
  const diag = postcode.diagnose('10001');
  console.log(diag.actionable_tip);
  // "Fix Length: Exactly 11 alphanumeric characters (e.g. 'EK 01 A03 FK 01'); Fix State: consider LA (Lagos)..."

  // 3. Offline Geolocation & Coordinates
  const loc = postcode.resolveLocation('EK 01 A03 FK 01');
  console.log(loc.latitude, loc.longitude, loc.google_maps_url);

  // 4. Reference Data for all 36 States + FCT
  const states = postcode.listStates();
  console.log(states['LA']); // Lagos state info

  // 5. Updating the NIPOST API Key & Online Lookup (L1-L3)
  postcode.setAPIKey('nipost_live_your_api_key_here');
  // Or: postcode.configure({ apiKey: '...', baseURL: 'https://api.postcode.gov.ng' });

  // Query live NIPOST Gateway for commercial building data
  const details = await postcode.lookup('EK 01 A03 FK 01', 3);
  console.log(details.building_use_status);
}

main();
```

## Configuring the NIPOST API Key

### 1. Offline vs. Online Boundary
- **Offline Methods (No API Key Required)**: `validate()`, `diagnose()`, `parse()`, `resolveLocation()`, and `listStates()` run 100% offline in WebAssembly memory with zero network latency.
- **Online Methods (API Key Configurable)**: `lookup()` connects to the live NIPOST Gateway (for commercial Level 2/3 street names, building use status, and GIS point geometry).

### 2. Ways to Set or Update the API Key

#### Option A: Dynamically at Runtime
```javascript
postcode.setAPIKey('nipost_live_your_new_key');
```

#### Option B: Full Configuration Object
```javascript
postcode.configure({
  apiKey: 'nipost_live_your_key',
  baseURL: 'https://api.postcode.gov.ng' // optional custom gateway
});
```

#### Option C: At Engine Initialization
```javascript
const postcode = await initPostcode(undefined, {
  apiKey: 'nipost_live_your_key'
});
```

#### Option D: Automatic Environment Variable (Node.js)
In Node.js, `initPostcode()` automatically detects `NIPOST_API_KEY` or `POSTCODE_API_KEY` from `process.env`.
```bash
export NIPOST_API_KEY=nipost_live_...
node app.js
```

## Browser Usage

```html
<script src="wasm_exec.js"></script>
<script>
  async function loadWasm() {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(fetch("postcode.wasm"), go.importObject);
    go.run(result.instance);

    // window.Postcode is now globally available
    const res = window.Postcode.validate("EK 01 A03 FK 01");
    console.log(res);
  }
  loadWasm();
</script>
```
