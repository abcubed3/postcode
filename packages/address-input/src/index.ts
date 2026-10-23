/**
 * @abcubed3/nipost-address-input
 * Lightweight embeddable Nigerian address & NIPOST postcode Web Component
 */

import { NipostAddressInputElement } from './component';
export * from './types';
export * from './data/lgas';
export { NipostAddressInputElement };

// Auto-register custom element if in a browser environment
if (typeof window !== 'undefined' && typeof customElements !== 'undefined') {
  if (!customElements.get('nipost-address-input')) {
    customElements.define('nipost-address-input', NipostAddressInputElement);
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'nipost-address-input': NipostAddressInputElement;
  }
}
