/**
 * Postcode Nigeria WebAssembly Loader
 * Supports Node.js, Deno, Bun, and browser environments.
 */

const fs = typeof window === 'undefined' ? require('fs') : null;
const path = typeof window === 'undefined' ? require('path') : null;

// Require wasm_exec.js if running in Node.js
if (typeof window === 'undefined' && typeof globalThis.Go === 'undefined') {
  require('./wasm_exec.js');
}

/**
 * Initializes the Postcode WebAssembly module.
 * @param {string|Buffer|ArrayBuffer|Response|{ apiKey?: string, baseURL?: string }} [wasmSource] Optional path, buffer, fetch response, or options.
 * @param {{ apiKey?: string, baseURL?: string }} [options] Optional configuration options.
 * @returns {Promise<typeof globalThis.Postcode>} The initialized Postcode module.
 */
async function initPostcode(wasmSource, options = {}) {
  if (wasmSource && typeof wasmSource === 'object' && !(wasmSource instanceof ArrayBuffer) && !(typeof Buffer !== 'undefined' && Buffer.isBuffer(wasmSource)) && !(typeof Response !== 'undefined' && wasmSource instanceof Response)) {
    options = wasmSource;
    wasmSource = undefined;
  }

  if (globalThis.Postcode) {
    if (options.apiKey) globalThis.Postcode.setAPIKey(options.apiKey);
    if (options.baseURL) globalThis.Postcode.configure({ baseURL: options.baseURL });
    return globalThis.Postcode;
  }

  const go = new globalThis.Go();
  let wasmBytes;

  if (!wasmSource) {
    if (typeof window === 'undefined') {
      const defaultPath = path.resolve(__dirname, 'postcode.wasm');
      if (fs.existsSync(defaultPath)) {
        wasmBytes = fs.readFileSync(defaultPath);
      } else {
        throw new Error(`postcode.wasm not found at ${defaultPath}. Build it first with: GOOS=js GOARCH=wasm go build -o wasm/postcode.wasm ./wasm`);
      }
    } else {
      wasmSource = 'postcode.wasm';
    }
  }

  let instance;
  if (wasmBytes) {
    const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
    instance = result.instance;
  } else if (typeof wasmSource === 'string' && typeof fetch !== 'undefined') {
    const response = await fetch(wasmSource);
    const result = await WebAssembly.instantiateStreaming(response, go.importObject);
    instance = result.instance;
  } else if (wasmSource instanceof ArrayBuffer || (typeof Buffer !== 'undefined' && Buffer.isBuffer(wasmSource))) {
    const result = await WebAssembly.instantiate(wasmSource, go.importObject);
    instance = result.instance;
  } else {
    throw new Error('Unsupported wasmSource format');
  }

  // Run the Go runtime in background (it will register globalThis.Postcode)
  go.run(instance);

  // Configure API key if provided via options or environment variable
  const envKey = typeof process !== 'undefined' && process.env ? (process.env.NIPOST_API_KEY || process.env.POSTCODE_API_KEY) : undefined;
  const apiKey = options.apiKey || envKey;
  if (apiKey) {
    globalThis.Postcode.setAPIKey(apiKey);
  }
  if (options.baseURL) {
    globalThis.Postcode.configure({ baseURL: options.baseURL });
  }

  return globalThis.Postcode;
}

module.exports = {
  initPostcode,
};
