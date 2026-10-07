# Postcode Nigeria WebAssembly (WASM)

Zero-dependency, offline-first Nigerian Postcode engine compiled to WebAssembly. Enables AI agents, edge workers, mobile apps, and web browsers to perform high-speed validation, diagnostics, formatting, administrative segment parsing/assembly, geocoding, and agent tool execution without network overhead.

## Features

- **Blazing Fast**: Microsecond execution on edge runtimes (Cloudflare Workers, Fastly, Vercel Edge, AWS Lambda@Edge).
- **Full Parity with Go Native**: Exposes all 19 Go library and CLI capabilities directly to Node.js, Bun, Deno, and web browsers.
- **Zero Network Overhead (Offline Core)**: Validation, batch validation, format repair tips, parsing, formatting, segment assembly/disassembly, geocoding, and synthetic address benchmarking run 100% offline in WASM memory.
- **NIPOST Gateway Integration (Online)**: Level 1–3 lookups, address autocompletion, radius-based nearby search, and reverse geocoding via standard async Promises.
- **Native AI Agent Protocols**: Pre-formatted tool definitions ready for OpenAI (`function`), Anthropic (`input_schema`), and Google Gemini (`functionDeclarations`), with a unified `executeTool` dispatcher and telemetry guardrails.

## Building the WASM Binary

```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o wasm/postcode.wasm ./wasm
```

## Quick Start (Node.js / Edge)

```javascript
const { initPostcode } = require('@abcubed3/postcode-wasm');

async function main() {
  const postcode = await initPostcode();

  // 1. Single & Batch Validation
  const valid = postcode.validate('EK 01 A03 FK 01');
  console.log(valid.valid, valid.state_name); // true, 'Ekiti'

  const batch = postcode.validateBatch(['EK 01 A03 FK 01', 'LA 11 W06 TC 10', 'INVALID']);
  console.log(batch.map(r => r.valid)); // [true, true, false]

  // 2. Formatting & Styles (canonical, spaced, compact, slug)
  console.log(postcode.format('EK 01 A03 FK 01', 'compact')); // 'EK01A03FK01'
  console.log(postcode.format('EK01A03FK01', 'canonical'));   // 'EK-01-A03-FK-01'
  console.log(postcode.format('EK 01 A03 FK 01', 'slug'));    // 'ek-01-a03-fk-01'

  // 3. Administrative Disassembly & Assembly
  const segs = postcode.disassemble('EK-01-A03-FK-01');
  // { state: 'EK', lga: '01', district: 'A03', area: 'FK', unit: '01' }

  const assembled = postcode.assemble({
    state: 'LA', lga: '11', district: 'W06', area: 'TC', unit: '10'
  });
  console.log(assembled.postcode, assembled.display); // 'LA-11-W06-TC-10', 'LA 11 W06 TC 10'

  // 4. Intelligent Diagnostic Engine
  const diag = postcode.diagnose('10001');
  console.log(diag.actionable_tip);
  // "Fix Length: Exactly 11 alphanumeric characters (e.g. 'EK 01 A03 FK 01'); Fix State: consider LA (Lagos)..."

  // 5. Offline Geolocation & Universal Maps
  const loc = postcode.resolveLocation('EK 01 A03 FK 01');
  console.log(loc.latitude, loc.longitude);
  console.log(loc.google_maps_directions_url); // Navigation URL
  console.log(loc.apple_maps_url);              // Apple Maps URL

  // 6. AI Agent Protocol (OpenAI, Anthropic, Gemini)
  const openAITools = postcode.getAgentTools('openai');
  const toolResult = await postcode.executeTool('validate_postcode', { code: 'EK 01 A03 FK 01' });
  console.log('Agent tool execution:', toolResult.valid, toolResult.state_name);

  // 7. Synthetic Address Generation for LLM Benchmarking
  const dataset = postcode.generateSyntheticAddresses({ count: 5, noiseRate: 0.2 });
  console.log(dataset[0].raw_text, dataset[0].expected_postcode);

  // 8. Live Gateway Operations (L1-L3 Lookup, Autocomplete, Nearby, Reverse Geocoding)
  postcode.setAPIKey('nipost_live_your_api_key_here');

  // Query live NIPOST Gateway for commercial building data
  const details = await postcode.lookup('EK 01 A03 FK 01', 3);
  console.log(details.building_use_status);

  // Autocomplete search
  const suggestions = await postcode.autocomplete('Adetokunbo');
  console.log(suggestions.results);

  // Search nearby units (by reference postcode or coordinates)
  const nearby = await postcode.nearby('LA-08-A86-RG-01', 250);
  // Or with coordinates directly: await postcode.nearby(6.6018, 3.3515, 250);
  console.log(nearby.results);

  // Reverse geocode coordinates to postcode
  const reverse = await postcode.reverseGeocode(6.6018, 3.3515, 25);
  console.log(reverse.found, reverse.postcode);
}

main();
```

## API Reference

### Core Offline Engine (Sync)
| Method | Description |
|---|---|
| `validate(code: string): ValidationResult` | Instant 11-char grammar & state validation. |
| `validateBatch(codes: string[]): ValidationResult[]` | High-throughput batch validation in WebAssembly memory. |
| `diagnose(code: string): DiagnosticReport` | Granular per-segment error detection with fuzzy repair tips. |
| `parse(code: string): ParsedPostcode` | Extracts full parsed model with capital, zone, and components. |
| `format(code: string, style?: string): string` | Normalizes to `canonical`, `compact`, `spaced`, `hyphenated`, or `slug`. |
| `assemble(segments: Segments): AssembledPostcode` | Assembles 5 administrative segments into canonical, display, and compact codes. |
| `disassemble(code: string): Segments` | Decomposes code into state, LGA, district, area, and building unit. |
| `resolveLocation(code: string): LocationResult` | Resolves offline centroid coordinates, Google Maps, Apple Maps, and OSM links. |
| `resolveLocationOnline(code: string): Promise<LocationResult>` | Fetches pinpoint building coordinates online via NIPOST cadastral discovery with local caching. |
| `registerBuilding(record: BuildingRecord): boolean` | Registers custom building unit into local geocoding registry. |
| `registerBuildings(records: BuildingRecord[]): number` | Bulk registers multiple building records for instant offline resolution. |
| `listStates(): Record<string, StateRecord>` | Returns static dictionary of all 36 Nigerian States + FCT. |

### AI Agent Protocol & Benchmarking
| Method | Description |
|---|---|
| `getAgentTools(format?: 'openai' \| 'anthropic' \| 'gemini'): any[]` | Exports JSON schemas for agent function calling. |
| `executeTool(name: string, args: any): Promise<any>` | Directly invokes agent tool handlers with structured responses. |
| `getAgentMetrics(): AgentGuardMetrics` | Telemetry for commercial vs. offline calls and downgrade rates. |
| `generateSyntheticAddresses(options?: GeneratorOptions): SyntheticAddress[]` | Generates realistic dirty/clean addresses with ground truth for LLM evaluation. |

### Live Gateway Operations (Async)
| Method | Description |
|---|---|
| `lookup(code: string, level?: number): Promise<LookupResponse>` | Queries NIPOST Gateway for Level 1, 2, or 3 commercial metadata. |
| `autocomplete(query: string): Promise<AutocompleteResponse>` | Real-time address and street suggestions. |
| `nearby(postcodeOrParams, radiusM?)` | Radius search using either a reference postcode or coordinate options. |
| `reverseGeocode(latOrParams, lng?, maxDistanceM?)` | Reverse-resolves latitude and longitude into nearest valid postcode. |

## Configuring the NIPOST API Key

### 1. Offline vs. Online Boundary
- **Offline Methods (No API Key Required)**: `validate()`, `validateBatch()`, `diagnose()`, `parse()`, `format()`, `assemble()`, `disassemble()`, `resolveLocation()`, `listStates()`, `getAgentTools()`, and `generateSyntheticAddresses()` run 100% offline in WebAssembly memory with zero network latency.
- **Online Methods (API Key Configurable)**: `lookup()`, `autocomplete()`, `nearby()`, and `reverseGeocode()` connect to the live NIPOST Gateway (for commercial Level 2/3 street names and building use classification; note that commercial API keys are strictly capped at Level 3).

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

    const loc = window.Postcode.resolveLocation("EK 01 A03 FK 01");
    console.log("Directions:", loc.google_maps_directions_url);
  }
  loadWasm();
</script>
```
