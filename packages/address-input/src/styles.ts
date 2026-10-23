/**
 * CSS stylesheet and theme tokens for <nipost-address-input>
 */

export const componentStyles = `
:host {
  display: block;
  box-sizing: border-box;
  font-family: var(--nipost-font, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif);
  color: var(--nipost-text, #0f172a);
  --nipost-primary: #15803d;
  --nipost-primary-hover: #166534;
  --nipost-primary-light: #f0fdf4;
  --nipost-bg: #ffffff;
  --nipost-surface: #f8fafc;
  --nipost-border: #e2e8f0;
  --nipost-border-focus: #15803d;
  --nipost-text: #0f172a;
  --nipost-muted: #64748b;
  --nipost-radius: 8px;
  --nipost-amber: #b45309;
  --nipost-amber-bg: #fffbeb;
  --nipost-amber-border: #fef3c7;
  --nipost-red: #b91c1c;
  --nipost-red-bg: #fef2f2;
  --nipost-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

:host([theme="dark"]) {
  --nipost-primary: #22c55e;
  --nipost-primary-hover: #4ade80;
  --nipost-primary-light: #064e3b;
  --nipost-bg: #0f172a;
  --nipost-surface: #1e293b;
  --nipost-border: #334155;
  --nipost-border-focus: #22c55e;
  --nipost-text: #f8fafc;
  --nipost-muted: #94a3b8;
  --nipost-amber: #fbbf24;
  --nipost-amber-bg: #451a03;
  --nipost-amber-border: #78350f;
  --nipost-red: #f87171;
  --nipost-red-bg: #450a0a;
}

:host([theme="minimal"]) {
  --nipost-surface: #ffffff;
  --nipost-border: #d4d4d8;
  --nipost-radius: 4px;
}

:host([theme="unstyled"]) {
  /* Inherits everything from outer DOM without forced theme styles */
}

*, *::before, *::after {
  box-sizing: border-box;
}

.nipost-wrapper {
  background: var(--nipost-bg);
  border: 1px solid var(--nipost-border);
  border-radius: var(--nipost-radius);
  padding: 1rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.nipost-wrapper:focus-within {
  border-color: var(--nipost-border-focus);
}

/* Header & Status Bar */
.nipost-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.65rem;
  font-size: 0.85rem;
}

.nipost-title {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
  color: var(--nipost-text);
}

.flag {
  font-size: 1.1rem;
}

.nipost-badge {
  font-size: 0.72rem;
  font-weight: 500;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  background: var(--nipost-surface);
  border: 1px solid var(--nipost-border);
  color: var(--nipost-muted);
}

.nipost-badge.valid {
  background: var(--nipost-primary-light);
  border-color: transparent;
  color: var(--nipost-primary);
}

/* Primary Search / Quick-Paste Box */
.search-container {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  width: 16px;
  height: 16px;
  color: var(--nipost-muted);
  pointer-events: none;
}

.search-input {
  width: 100%;
  font-family: inherit;
  font-size: 0.92rem;
  color: var(--nipost-text);
  background: var(--nipost-surface);
  border: 1px solid var(--nipost-border);
  border-radius: var(--nipost-radius);
  padding: 0.65rem 2.2rem 0.65rem 2.35rem;
  outline: none;
  transition: all 0.15s ease;
}

.search-input:focus {
  background: var(--nipost-bg);
  border-color: var(--nipost-border-focus);
  box-shadow: 0 0 0 1px var(--nipost-border-focus);
}

.search-input::placeholder {
  color: var(--nipost-muted);
  opacity: 0.8;
}

.clear-btn {
  position: absolute;
  right: 0.6rem;
  background: none;
  border: none;
  color: var(--nipost-muted);
  cursor: pointer;
  padding: 0.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.clear-btn:hover {
  color: var(--nipost-text);
}

/* GPS Button */
.gps-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.65rem 0.85rem;
  background: var(--nipost-surface);
  border: 1px solid var(--nipost-border);
  border-radius: var(--nipost-radius);
  color: var(--nipost-text);
  font-family: inherit;
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.gps-btn:hover:not(:disabled) {
  border-color: var(--nipost-primary);
  color: var(--nipost-primary);
  background: var(--nipost-primary-light);
}

.gps-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Disambiguation summary pill */
.summary-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: 0.6rem;
  font-size: 0.8rem;
}

.tag-item {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  background: var(--nipost-surface);
  border: 1px solid var(--nipost-border);
  color: var(--nipost-text);
  font-size: 0.78rem;
}

.tag-label {
  color: var(--nipost-muted);
  font-size: 0.72rem;
}

.tag-landmark {
  background: #fef3c7;
  border-color: #fde68a;
  color: #92400e;
}

/* Postcode 5-Segment Chips */
.postcode-display {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.75rem;
  padding: 0.6rem 0.75rem;
  background: var(--nipost-surface);
  border: 1px dashed var(--nipost-border);
  border-radius: var(--nipost-radius);
  overflow-x: auto;
}

.postcode-title {
  font-size: 0.75rem;
  color: var(--nipost-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 600;
  margin-right: 0.25rem;
}

.segment-chip {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  background: var(--nipost-bg);
  border: 1px solid var(--nipost-border);
  border-radius: 6px;
  padding: 0.25rem 0.55rem;
  min-width: 48px;
}

.segment-val {
  font-family: var(--nipost-mono);
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--nipost-text);
  line-height: 1.1;
}

.segment-name {
  font-size: 0.65rem;
  color: var(--nipost-muted);
  margin-top: 2px;
}

.segment-sep {
  color: var(--nipost-muted);
  font-weight: 700;
  font-family: var(--nipost-mono);
}

/* Diagnostic Alert Box */
.diagnostic-box {
  margin-top: 0.65rem;
  padding: 0.6rem 0.8rem;
  border-radius: var(--nipost-radius);
  font-size: 0.82rem;
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.diagnostic-box.warning {
  background: var(--nipost-amber-bg);
  border: 1px solid var(--nipost-amber-border);
  color: var(--nipost-amber);
}

.diagnostic-box.error {
  background: var(--nipost-red-bg);
  border: 1px solid var(--nipost-red);
  color: var(--nipost-red);
}

.diagnostic-box svg {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  margin-top: 1px;
}

.diagnostic-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
  flex: 1;
}

.repair-btn {
  display: inline-flex;
  align-items: center;
  background: var(--nipost-primary);
  color: #ffffff;
  border: none;
  border-radius: var(--nipost-radius);
  font-family: inherit;
  font-size: 0.76rem;
  font-weight: 600;
  padding: 0.25rem 0.6rem;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.repair-btn:hover {
  opacity: 0.9;
}

/* Mode Switch / Toggle Button */
.expand-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.75rem;
  background: none;
  border: none;
  color: var(--nipost-primary);
  font-family: inherit;
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0.2rem 0;
}

.expand-toggle:hover {
  text-decoration: underline;
}

/* Structured Form Grid */
.structured-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-top: 0.85rem;
  padding-top: 0.85rem;
  border-top: 1px solid var(--nipost-border);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-group.full {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--nipost-muted);
}

.form-input, .form-select {
  font-family: inherit;
  font-size: 0.88rem;
  color: var(--nipost-text);
  background: var(--nipost-surface);
  border: 1px solid var(--nipost-border);
  border-radius: var(--nipost-radius);
  padding: 0.55rem 0.7rem;
  outline: none;
  transition: border-color 0.15s ease;
}

.form-input:focus, .form-select:focus {
  background: var(--nipost-bg);
  border-color: var(--nipost-border-focus);
}

.hidden {
  display: none !important;
}

/* Loading / Ready Indicator */
.loading-indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  color: var(--nipost-muted);
}

.spinner {
  width: 12px;
  height: 12px;
  border: 2px solid var(--nipost-border);
  border-top-color: var(--nipost-primary);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .structured-grid {
    grid-template-columns: 1fr;
  }
  .search-container {
    flex-direction: column;
    align-items: stretch;
  }
}
`;
