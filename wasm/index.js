/**
 * Postcode Nigeria WebAssembly Loader
 * Supports Node.js, Deno, Bun, Edge Workers, and browser environments.
 */

const fs = typeof window === 'undefined' ? require('fs') : null;
const path = typeof window === 'undefined' ? require('path') : null;

// Require wasm_exec.js if running in Node.js / Bun / Deno
if (typeof window === 'undefined' && typeof globalThis.Go === 'undefined') {
  require('./wasm_exec.js');
}

let initPromise = null;

function isPlainOptionsObject(val) {
  if (!val || typeof val !== 'object') return false;
  if (val instanceof ArrayBuffer || ArrayBuffer.isView(val)) return false;
  if (typeof Buffer !== 'undefined' && Buffer.isBuffer(val)) return false;
  if (typeof Response !== 'undefined' && val instanceof Response) return false;
  if (typeof WebAssembly !== 'undefined' && val instanceof WebAssembly.Module) return false;
  if (typeof val.then === 'function') return false;
  return true;
}

function applyConfiguration(engine, options = {}) {
  if (!engine) return engine;
  const envKey =
    typeof process !== 'undefined' && process.env
      ? process.env.NIPOST_API_KEY || process.env.POSTCODE_API_KEY
      : undefined;
  const apiKey = options.apiKey || envKey;
  if (apiKey) {
    engine.setAPIKey(apiKey);
  }
  if (options.baseURL || options.googleMapsApiKey || options.googleMapsKey) {
    engine.configure(options);
  }
  return engine;
}

async function instantiateWasm(wasmSource, importObject) {
  if (!wasmSource) {
    if (typeof window === 'undefined' && fs && path) {
      const defaultPath = path.resolve(__dirname, 'postcode.wasm');
      if (fs.existsSync(defaultPath)) {
        const wasmBytes = fs.readFileSync(defaultPath);
        const result = await WebAssembly.instantiate(wasmBytes, importObject);
        return result.instance;
      }
      throw new Error(
        `postcode.wasm not found at ${defaultPath}. Build it first with: GOOS=js GOARCH=wasm go build -o wasm/postcode.wasm ./wasm`
      );
    }
    wasmSource = 'postcode.wasm';
  }

  // 1. Pre-compiled WebAssembly.Module (Cloudflare Workers / Edge runtimes)
  if (typeof WebAssembly !== 'undefined' && wasmSource instanceof WebAssembly.Module) {
    return await WebAssembly.instantiate(wasmSource, importObject);
  }

  // 2. Binary buffer (ArrayBuffer, TypedArray, Node Buffer)
  if (
    wasmSource instanceof ArrayBuffer ||
    ArrayBuffer.isView(wasmSource) ||
    (typeof Buffer !== 'undefined' && Buffer.isBuffer(wasmSource))
  ) {
    const result = await WebAssembly.instantiate(wasmSource, importObject);
    return result.instance;
  }

  // 3. URL string or fetch Response / Promise<Response>
  let responsePromise = null;
  if (typeof wasmSource === 'string' && typeof fetch !== 'undefined') {
    responsePromise = fetch(wasmSource);
  } else if (
    (typeof Response !== 'undefined' && wasmSource instanceof Response) ||
    (wasmSource && typeof wasmSource.then === 'function')
  ) {
    responsePromise = Promise.resolve(wasmSource);
  }

  if (responsePromise) {
    if (typeof WebAssembly.instantiateStreaming === 'function') {
      try {
        const response = await responsePromise;
        const cloned = typeof response.clone === 'function' ? response.clone() : null;
        try {
          const result = await WebAssembly.instantiateStreaming(response, importObject);
          return result.instance;
        } catch (_streamErr) {
          // Fallback if server omits 'application/wasm' Content-Type header
          const fallbackResp = cloned || (typeof wasmSource === 'string' ? await fetch(wasmSource) : response);
          const bytes = await fallbackResp.arrayBuffer();
          const result = await WebAssembly.instantiate(bytes, importObject);
          return result.instance;
        }
      } catch (err) {
        throw err;
      }
    }
    const response = await responsePromise;
    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, importObject);
    return result.instance;
  }

  throw new Error('Unsupported wasmSource format');
}

/**
 * Initializes the Postcode WebAssembly module (deduplicating concurrent calls).
 * @param {string|Buffer|ArrayBuffer|Uint8Array|Response|WebAssembly.Module|{ apiKey?: string, baseURL?: string, googleMapsApiKey?: string }} [wasmSource] Optional path, buffer, fetch response, WebAssembly.Module, or options.
 * @param {{ apiKey?: string, baseURL?: string, googleMapsApiKey?: string }} [options] Optional configuration options.
 * @returns {Promise<typeof globalThis.Postcode>} The initialized Postcode module.
 */
async function initPostcode(wasmSource, options = {}) {
  if (isPlainOptionsObject(wasmSource)) {
    options = wasmSource;
    wasmSource = undefined;
  }

  if (globalThis.Postcode) {
    return applyConfiguration(globalThis.Postcode, options);
  }

  if (!initPromise) {
    initPromise = (async () => {
      const go = new globalThis.Go();
      const instance = await instantiateWasm(wasmSource, go.importObject);
      // Start the Go WASM runtime (registers globalThis.Postcode synchronously before yielding)
      go.run(instance);
      return globalThis.Postcode;
    })().catch((err) => {
      initPromise = null;
      throw err;
    });
  }

  const engine = await initPromise;
  return applyConfiguration(engine, options);
}

module.exports = {
  initPostcode,
};
