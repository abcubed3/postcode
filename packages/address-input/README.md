# @abcubed3/nipost-address-input 🇳🇬

A lightweight, modern, open-source embeddable address collection Web Component for Nigerian web applications. Powered by [`@abcubed3/postcode-wasm`](https://github.com/abcubed3/postcode/tree/main/wasm) to bring official NIPOST 11-character digital postcodes, landmark detection, 774 LGA centroids, and self-healing address disambiguation directly to any frontend stack.

---

## Highlights

- **Framework-Agnostic Web Component**: Works in plain HTML, React, Next.js, Vue, Angular, Svelte, or any modern web framework.
- **Hybrid Dual-Mode UX**:
  - **Smart Quick-Paste / Search**: Users can paste freeform addresses (e.g. *"Plot 14 Admiralty Way, Lekki Phase 1, Opposite Domino's Pizza, Lagos"*). The 5-stage pipeline automatically isolates the **State**, **LGA**, **Locality**, **Landmark**, **Street**, and canonical **11-character NIPOST Postcode**.
  - **Expandable Structured Breakdown**: Toggle to manual cascading dropdowns for State (37 states + FCT), 774 LGAs, Locality, Landmark, and Street/Unit.
- **Form-Associated Custom Element**: Automatically registers with native `<form>` submissions, sending structured sub-keys (`address[state]`, `address[lga]`, `address[postcode]`, `address[landmark]`, `address[street]`) with zero extra JS.
- **Three-Tier Customization**:
  - **Preset Themes**: `theme="light"` (default Nigerian green/slate), `theme="dark"`, `theme="minimal"`, and `theme="unstyled"`.
  - **CSS Design Tokens**: Semantic CSS variables (`--nipost-primary`, `--nipost-radius`, `--nipost-bg`, etc.).
  - **Shadow Parts (`::part()`)**: Style internal elements directly with custom CSS or Tailwind `@apply`.
  - **Unstyled Mode**: Renders for direct Tailwind CSS utility styling.
- **GPS One-Click Geolocation**: Device GPS button reverse-geocodes nearest Nigerian LGA centroid.
- **Self-Healing Diagnostics**: Visual 5-segment NIPOST chips (`SS-LL-DDD-AA-UU`) with instant tips for legacy 6-digit codes or illegal units (`00`).

---

## Quick Start

### Option 1: Zero-Install CDN (Plain HTML)

```html
<!-- Drop the script into your <head> or before </body> -->
<script type="module" src="https://unpkg.com/@abcubed3/nipost-address-input"></script>

<!-- Use inside any standard HTML form -->
<form action="/checkout" method="POST">
  <nipost-address-input
    name="shipping_address"
    placeholder="Enter street, landmark, or estate..."
    required
  ></nipost-address-input>

  <button type="submit">Place Order</button>
</form>

<script>
  const widget = document.querySelector('nipost-address-input');
  widget.addEventListener('nipost-resolve', (e) => {
    console.log('Verified Postcode:', e.detail.postcode);
    console.log('Structured Address:', e.detail.address);
  });
</script>
```

### Option 2: npm Package (React, Vue, Vite, Next.js)

```bash
npm install @abcubed3/nipost-address-input
```

```javascript
import '@abcubed3/nipost-address-input';
```

---

## Form Submission (`FormData` Contract)

Because `<nipost-address-input>` implements the HTML Form-Associated Custom Element standard, the server receives standard form keys on submission:

| Key | Example Value | Description |
|---|---|---|
| `address[raw]` | `"14 Admiralty Way, Lekki, Lagos"` | Full user input string |
| `address[state]` | `"LA"` | 2-letter NIPOST state code |
| `address[lga]` | `"08"` | 2-digit LGA code |
| `address[locality]` | `"Lekki Phase 1"` | Neighborhood or estate |
| `address[landmark]` | `"Opposite Domino's"` | Anchor phrase |
| `address[street]` | `"14 Admiralty Way"` | Street and house number |
| `address[postcode]` | `"LA-08-A86-RG-01"` | Canonical 11-char NIPOST code |
| `address[latitude]` | `6.4474` | Centroid latitude |
| `address[longitude]` | `3.4723` | Centroid longitude |

---

## Framework Integration Recipes

### React / Next.js

```tsx
import React, { useEffect, useRef } from 'react';
import '@abcubed3/nipost-address-input';
import type { AddressResolveEventDetail } from '@abcubed3/nipost-address-input';

export function CheckoutForm() {
  const addressRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const el = addressRef.current;
    if (!el) return;

    const onResolve = (e: CustomEvent<AddressResolveEventDetail>) => {
      console.log('NIPOST Code:', e.detail.postcode);
      console.log('Address Details:', e.detail.address);
    };

    el.addEventListener('nipost-resolve', onResolve as EventListener);
    return () => el.removeEventListener('nipost-resolve', onResolve as EventListener);
  }, []);

  return (
    <form onSubmit={(e) => { e.preventDefault(); /* ... */ }}>
      <nipost-address-input
        ref={addressRef}
        name="delivery_address"
        theme="light"
        required
      />
      <button type="submit" className="mt-4 px-4 py-2 bg-green-700 text-white rounded">
        Submit Order
      </button>
    </form>
  );
}
```

### Vue 3

```vue
<template>
  <form @submit.prevent="submitOrder">
    <nipost-address-input
      name="delivery_address"
      theme="light"
      @nipost-resolve="onResolved"
      @nipost-change="onChanged"
    />
    <button type="submit">Submit</button>
  </form>
</template>

<script setup>
import '@abcubed3/nipost-address-input';

function onResolved(event) {
  console.log('Resolved Postcode:', event.detail.postcode);
}

function onChanged(event) {
  console.log('Address draft:', event.detail.address);
}
</script>
```

### Angular

Add `CUSTOM_ELEMENTS_SCHEMA` to your component or module schemas:

```typescript
import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import '@abcubed3/nipost-address-input';

@Component({
  selector: 'app-shipping',
  template: `
    <nipost-address-input
      name="shipping_address"
      (nipost-resolve)="onAddressResolved($event)"
    ></nipost-address-input>
  `,
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class ShippingComponent {
  onAddressResolved(event: any) {
    console.log('Verified Postcode:', event.detail.postcode);
  }
}
```

---

## Styling & Design Tokens

### CSS Custom Properties

```css
nipost-address-input {
  --nipost-primary: #15803d;        /* Accent & button color */
  --nipost-primary-hover: #166534;  /* Hover state */
  --nipost-bg: #ffffff;             /* Card background */
  --nipost-surface: #f8fafc;        /* Input background */
  --nipost-border: #e2e8f0;         /* Border color */
  --nipost-border-focus: #15803d;   /* Active border */
  --nipost-text: #0f172a;           /* Text color */
  --nipost-muted: #64748b;          /* Placeholder & labels */
  --nipost-radius: 8px;             /* Corner radius */
  --nipost-font: system-ui, sans-serif;
  --nipost-mono: ui-monospace, monospace;
}
```

### Shadow Parts (`::part()`)

Target internal elements directly with custom CSS or Tailwind `@apply`:

```css
nipost-address-input::part(container) {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
}

nipost-address-input::part(search-input) {
  font-size: 1rem;
}

nipost-address-input::part(gps-btn) {
  background-color: #f0fdf4;
}

nipost-address-input::part(badge) {
  text-transform: uppercase;
}
```

---

## Component Attributes

| Attribute | Type | Default | Description |
|---|---|---|---|
| `name` | string | `"address"` | Name prefix used for form submission fields |
| `required` | boolean | `false` | HTML form requirement validation |
| `disabled` | boolean | `false` | Disables all child inputs |
| `theme` | `"light" \| "dark" \| "minimal" \| "unstyled"` | `"light"` | Preset visual theme |
| `allow-gps` | `"true" \| "false"` | `"true"` | Toggle "Current Location" GPS button |
| `loading` | `"lazy" \| "idle" \| "eager"` | `"lazy"` | WASM loading strategy (lazy loads on first interaction) |
| `wasm-url` | string | Unpkg CDN | Path or URL to `postcode.wasm` |
| `wasm-exec-url` | string | Unpkg CDN | Path or URL to `wasm_exec.js` |
| `endpoint` | string | `""` | Backend proxy endpoint to protect commercial API keys |
| `api-key` | string | `""` | Direct NIPOST API key (recommend `endpoint` for public HTML) |
| `placeholder` | string | Nigerian hint | Custom placeholder for smart search input |
| `value` | string | `""` | Initial raw address string |

> [!SECURITY]
> **API Key Protection:** In client-facing production applications, avoid putting commercial `api-key` attributes directly in client HTML source code. Instead, point `endpoint="/api/nipost-proxy"` to your own backend route that adds the private `NIPOST_API_KEY` header before forwarding to NIPOST gateway.

---

## Custom DOM Events

| Event Name | `event.detail` | Description |
|---|---|---|
| `nipost-change` | `{ address, isValid, mode }` | Fired whenever any input or selection changes |
| `nipost-resolve` | `{ address, postcode, formattedDisplay }` | Fired when a verified NIPOST postcode is generated |
| `nipost-diagnose` | `{ raw, postcodeCandidate, tip, isValid }` | Fired when an address typo or legacy code tip is shown |
| `nipost-ready` | `{ engine }` | Fired when the WebAssembly engine finishes loading |

---

## JavaScript Methods & Properties

```javascript
const el = document.querySelector('nipost-address-input');

// Properties
console.log(el.address);   // Structured NigerianAddress object
console.log(el.postcode);  // "LA-08-A86-RG-01"
console.log(el.isValid);   // true / false
console.log(el.value);     // Raw address string

// Methods
el.setAddress("Plot 12, Wuse II, Abuja"); // Auto-parses and updates UI
el.reset();                              // Clears all fields and reset state

// Manual engine injection (optional)
import { initPostcode } from '@abcubed3/postcode-wasm';
el.engine = await initPostcode('/assets/postcode.wasm');
```

---

## Development & Testing

```bash
# Install dependencies
npm install

# Start Vite dev server & interactive demo
npm run dev

# Run Vitest test suite
npm test

# Build production library (ESM, CJS, and standalone UMD)
npm run build
```

---

## License

MIT © [abcubed3](https://github.com/abcubed3)
