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
}

main();
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
