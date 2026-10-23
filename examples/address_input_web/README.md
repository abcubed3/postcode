# NIPOST Address Input — Web Client Component Example 🇳🇬

A standalone, runnable example showcasing the official **`<nipost-address-input>`** Web Component embedded inside a Nigerian delivery checkout form.

Powered by [`@abcubed3/postcode-wasm`](https://github.com/abcubed3/postcode/tree/main/wasm), this component brings official NIPOST 11-character digital postcodes, 774 LGA centroids, landmark detection, and self-healing address disambiguation directly to browser frontends.

---

## What This Example Demonstrates

1. **Embedded Web Component (`<nipost-address-input>`)**:
   - Zero-dependency custom element running locally or via CDN.
   - Dual-mode UX: One-box smart search/paste + expandable cascading dropdowns across all **774 Nigerian LGAs**.
2. **Form-Associated Custom Element**:
   - Submits structured NIPOST sub-keys with native HTML `<form>` submissions (`delivery_address[state]`, `delivery_address[lga]`, `delivery_address[postcode]`, etc.) without requiring custom serializer JavaScript.
3. **Real-Time Telemetry & Event Streams**:
   - Listens to `nipost-change` on keystroke and `nipost-resolve` when a canonical 11-character postcode or structured address is verified.
4. **Interactive Presets**:
   - Instant 1-click test buttons for Lagos, Abuja, Ekiti, and canonical 11-digit codes.

---

## Running the Example

### Option 1: Open Directly in Browser (No Server Required)

Since the example references the bundled standalone library in `packages/address-input/dist/nipost-address-input.min.js`, you can open the file directly in your browser:

```bash
# macOS
open examples/address_input_web/index.html

# Linux
xdg-open examples/address_input_web/index.html

# Windows
start examples/address_input_web/index.html
```

### Option 2: Run with Any Local HTTP Server

```bash
# Using npx serve
npx serve examples/address_input_web

# Or Python 3 built-in server
python3 -m http.server 8080 --directory examples/address_input_web
```

Navigate to `http://localhost:8080` in your browser.

---

## Embedding in Your Own Projects

### 1. Plain HTML / Zero-Install CDN

```html
<!-- Include script in <head> or before </body> -->
<script type="module" src="https://unpkg.com/@abcubed3/nipost-address-input"></script>

<!-- Embed inside any standard HTML form -->
<form action="/api/checkout" method="POST">
  <nipost-address-input
    name="shipping_address"
    placeholder="Enter street, landmark, or 11-digit NIPOST code..."
    theme="light"
    required
  ></nipost-address-input>

  <button type="submit">Place Order</button>
</form>

<script>
  const input = document.querySelector('nipost-address-input');
  
  // Real-time keystroke & parsing updates
  input.addEventListener('nipost-change', (e) => {
    console.log('Address draft:', e.detail.address);
    console.log('Valid:', e.detail.isValid);
  });

  // Canonical resolution event
  input.addEventListener('nipost-resolve', (e) => {
    console.log('Verified NIPOST Postcode:', e.detail.postcode);
    console.log('Centroid Coords:', e.detail.address.latitude, e.detail.address.longitude);
  });
</script>
```

### 2. Modern Frontend Bundlers (React, Vue, Vite, Next.js)

Install the npm package:

```bash
npm install @abcubed3/nipost-address-input
```

Import once in your app's entry point:

```typescript
import '@abcubed3/nipost-address-input';
```

#### React / Next.js Example

```tsx
import React, { useEffect, useRef } from 'react';
import '@abcubed3/nipost-address-input';
import type { AddressResolveEventDetail } from '@abcubed3/nipost-address-input';

export function ShippingForm() {
  const addressRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const el = addressRef.current;
    if (!el) return;

    const handleResolve = (e: CustomEvent<AddressResolveEventDetail>) => {
      console.log('Verified Postcode:', e.detail.postcode);
      console.log('Full Address:', e.detail.address);
    };

    el.addEventListener('nipost-resolve', handleResolve as EventListener);
    return () => el.removeEventListener('nipost-resolve', handleResolve as EventListener);
  }, []);

  return (
    <form onSubmit={(e) => { e.preventDefault(); }}>
      <nipost-address-input
        ref={addressRef}
        name="shipping_address"
        theme="light"
        required
      />
      <button type="submit">Submit Order</button>
    </form>
  );
}
```

---

## Form Submission Contract (`FormData`)

When submitted inside a `<form>`, `<nipost-address-input>` exposes structured fields under its `name` attribute:

| Field Name | Example Value | Description |
| :--- | :--- | :--- |
| `[name][raw]` | `"Plot 14 Admiralty Way, Lekki Phase 1, Lagos"` | Raw input text |
| `[name][state]` | `"LA"` | 2-letter NIPOST state code |
| `[name][lga]` | `"08"` | 2-digit LGA code |
| `[name][locality]` | `"Lekki Phase 1"` | Extracted locality / estate |
| `[name][landmark]` | `"Opposite Domino's Pizza"` | Detected landmark phrase |
| `[name][street]` | `"Plot 14 Admiralty Way"` | Street address |
| `[name][postcode]` | `"LA-08-A86-RG-01"` | Canonical 11-character NIPOST code |
| `[name][latitude]` | `6.4474` | Centroid latitude |
| `[name][longitude]` | `3.4723` | Centroid longitude |

---

## Related Packages & Documentation

- [`packages/address-input/`](../../packages/address-input) — Component source code, Vitest suite, and styling docs.
- [`wasm/`](../../wasm) — Zero-allocation WebAssembly core engine.
- [Root Documentation](../../README.md) — Nigerian Postcode Go SDK & Toolkit.
