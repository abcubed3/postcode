import { defineConfig } from 'vite';
import { resolve } from 'path';
import { execSync } from 'child_process';

function resolveLatestWasmVersion(): string {
  const envVer = process.env.WASM_VERSION?.trim().replace(/^v/, '');
  if (envVer && /^\d+\.\d+\.\d+/.test(envVer)) {
    return envVer;
  }
  try {
    const out = execSync("git tag --list 'v[0-9]*' --sort=-v:refname", {
      cwd: __dirname,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    const latestTag = out
      .split(/\r?\n/)
      .map((line) => line.trim())
      .find((line) => /^v\d+\.\d+\.\d+/.test(line));
    if (latestTag) {
      return latestTag.replace(/^v/, '');
    }
  } catch {
    // Fallback below when git is unavailable
  }
  return '0.2.5';
}

export default defineConfig({
  root: __dirname,
  define: {
    __DEFAULT_WASM_VERSION__: JSON.stringify(resolveLatestWasmVersion()),
  },
  resolve: {
    alias: {
      'happy-dom': resolve(__dirname, 'node_modules/happy-dom'),
      '@abcubed3/postcode-wasm': resolve(__dirname, '../../wasm/index.js'),
    },
  },
  build: {
    copyPublicDir: false,
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'NipostAddressInput',
      formats: ['es', 'cjs', 'umd'],
      fileName: (format) => {
        if (format === 'es') return 'index.js';
        if (format === 'cjs') return 'index.cjs';
        return 'nipost-address-input.min.js';
      },
    },
    rollupOptions: {
      output: {
        exports: 'named',
      },
    },
    sourcemap: true,
  },
  test: {
    environment: 'happy-dom',
    globals: true,
  },
});
