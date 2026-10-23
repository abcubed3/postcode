import { componentStyles } from './styles';
import type {
  NigerianAddress,
  PostcodeWasmModule,
  StateOption,
  LGAOption,
  AddressChangeEventDetail,
  AddressResolveEventDetail,
  AddressDiagnoseEventDetail,
  AddressErrorEventDetail,
} from './types';
import { getLGAsForState, findNearestLGA } from './data/lgas';

const DEFAULT_WASM_VERSION = '1.0.0';
const DEFAULT_WASM_CDN_URL = `https://unpkg.com/@abcubed3/postcode-wasm@${DEFAULT_WASM_VERSION}/postcode.wasm`;
const DEFAULT_WASM_EXEC_CDN_URL = `https://unpkg.com/@abcubed3/postcode-wasm@${DEFAULT_WASM_VERSION}/wasm_exec.js`;

// Shared module-level singleton loader to avoid duplicate script injection and compilation across inputs
let sharedEnginePromise: Promise<PostcodeWasmModule> | null = null;

function escapeHTML(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function createOption(text: string, value: string): HTMLOptionElement {
  const opt = document.createElement('option');
  opt.value = value;
  opt.textContent = text;
  return opt;
}

// Fallback states list in case WASM is loading asynchronously
const DEFAULT_STATES: StateOption[] = [
  { code: 'AB', name: 'Abia' }, { code: 'AD', name: 'Adamawa' }, { code: 'AK', name: 'Akwa Ibom' },
  { code: 'AN', name: 'Anambra' }, { code: 'BA', name: 'Bauchi' }, { code: 'BY', name: 'Bayelsa' },
  { code: 'BN', name: 'Benue' }, { code: 'BO', name: 'Borno' }, { code: 'CR', name: 'Cross River' },
  { code: 'DE', name: 'Delta' }, { code: 'EB', name: 'Ebonyi' }, { code: 'ED', name: 'Edo' },
  { code: 'EK', name: 'Ekiti' }, { code: 'EN', name: 'Enugu' }, { code: 'FC', name: 'Federal Capital Territory' },
  { code: 'GO', name: 'Gombe' }, { code: 'IM', name: 'Imo' }, { code: 'JI', name: 'Jigawa' },
  { code: 'KD', name: 'Kaduna' }, { code: 'KN', name: 'Kano' }, { code: 'KT', name: 'Katsina' },
  { code: 'KE', name: 'Kebbi' }, { code: 'KO', name: 'Kogi' }, { code: 'KW', name: 'Kwara' },
  { code: 'LA', name: 'Lagos' }, { code: 'NA', name: 'Nasarawa' }, { code: 'NI', name: 'Niger' },
  { code: 'OG', name: 'Ogun' }, { code: 'ON', name: 'Ondo' }, { code: 'OS', name: 'Osun' },
  { code: 'OY', name: 'Oyo' }, { code: 'PL', name: 'Plateau' }, { code: 'RI', name: 'Rivers' },
  { code: 'SO', name: 'Sokoto' }, { code: 'TA', name: 'Taraba' }, { code: 'YO', name: 'Yobe' },
  { code: 'ZA', name: 'Zamfara' }
];

export class NipostAddressInputElement extends HTMLElement {
  static formAssociated = true;

  private _internals?: ElementInternals;
  private _engine: PostcodeWasmModule | null = null;
  private _isEngineLoading = false;
  private _isStructuredExpanded = false;
  private _rendered = false;
  private _renderedLgaState = '';
  private _lastResolvedPostcode: string | null = null;
  private _debounceTimer: ReturnType<typeof setTimeout> | null = null;

  private _address: NigerianAddress = {
    raw: '',
    postcodeValid: false,
    confidence: 'low',
  };

  // Shadow DOM elements
  private root!: ShadowRoot;
  private searchInput!: HTMLInputElement;
  private clearBtn!: HTMLButtonElement;
  private gpsBtn!: HTMLButtonElement;
  private statusBadge!: HTMLElement;
  private loadingSpan!: HTMLElement;
  private summaryBar!: HTMLElement;
  private postcodeDisplay!: HTMLElement;
  private diagnosticBox!: HTMLElement;
  private diagnosticMsg!: HTMLElement;
  private repairBtn!: HTMLButtonElement;
  private expandToggle!: HTMLButtonElement;
  private structuredGrid!: HTMLElement;

  // Structured form inputs
  private stateSelect!: HTMLSelectElement;
  private lgaSelect!: HTMLSelectElement;
  private localityInput!: HTMLInputElement;
  private landmarkInput!: HTMLInputElement;
  private streetInput!: HTMLInputElement;
  private unitInput!: HTMLInputElement;

  static get observedAttributes() {
    return [
      'name',
      'required',
      'disabled',
      'theme',
      'allow-gps',
      'wasm-url',
      'wasm-exec-url',
      'api-key',
      'endpoint',
      'placeholder',
      'value',
      'loading',
    ];
  }

  constructor() {
    super();
    if (typeof this.attachInternals === 'function') {
      this._internals = this.attachInternals();
    }
    this.root = this.attachShadow({ mode: 'open', delegatesFocus: true });
  }

  connectedCallback() {
    if (!this._rendered) {
      this.render();
      this._rendered = true;
    }

    this._upgradeProperty('name');
    this._upgradeProperty('required');
    this._upgradeProperty('disabled');
    this._upgradeProperty('value');
    this._upgradeProperty('engine');

    if (this.hasAttribute('disabled')) {
      this.setDisabled(true);
    }
    if (this.hasAttribute('value')) {
      const initVal = this.getAttribute('value');
      if (initVal) {
        this.setAddress(initVal);
      }
    }

    this.syncFormValue();
    this.setupEngineLoading();
  }

  formResetCallback() {
    this.reset();
  }

  formDisabledCallback(disabled: boolean) {
    this.setDisabled(disabled);
  }

  formStateRestoreCallback(state: unknown) {
    if (typeof state === 'string') {
      this.setAddress(state);
    }
  }

  private _upgradeProperty(prop: 'name' | 'required' | 'disabled' | 'value' | 'engine') {
    if (Object.prototype.hasOwnProperty.call(this, prop)) {
      const val = (this as any)[prop];
      delete (this as any)[prop];
      (this as any)[prop] = val;
    }
  }

  attributeChangedCallback(name: string, oldValue: string | null, newValue: string | null) {
    if (oldValue === newValue) return;

    if (name === 'disabled') {
      this.setDisabled(this.hasAttribute('disabled'));
    } else if (name === 'placeholder' && this.searchInput) {
      this.searchInput.placeholder = newValue || 'Paste or type Nigerian address (e.g. 14 Admiralty Way, Lekki, Lagos)';
    } else if (name === 'allow-gps' && this.gpsBtn) {
      this.gpsBtn.classList.toggle('hidden', newValue === 'false');
    } else if (name === 'value') {
      if (this.searchInput && newValue !== null && newValue !== this._address.raw) {
        this.setAddress(newValue);
      }
    } else if (name === 'required' || name === 'name') {
      this.syncFormValue();
    } else if (name === 'api-key') {
      if (newValue && this._engine && typeof this._engine.setAPIKey === 'function') {
        this._engine.setAPIKey(newValue);
      }
    } else if (name === 'endpoint') {
      if (newValue && this._engine && typeof (this._engine as any).configure === 'function') {
        (this._engine as any).configure({ baseURL: newValue });
      }
    }
  }

  // --- Public Getters & Setters ---

  get engine(): PostcodeWasmModule | null {
    return this._engine;
  }

  set engine(engineInstance: PostcodeWasmModule | null) {
    this._engine = engineInstance;
    if (this._engine) {
      this._isEngineLoading = false;
      this.updateLoadingState();
      this.populateStates();
      this.dispatchEvent(new CustomEvent('nipost-ready', { detail: { engine: this._engine } }));
      if (this._address.raw) {
        this.runDisambiguation(this._address.raw);
      }
    }
  }

  get address(): NigerianAddress {
    return { ...this._address };
  }

  get postcode(): string {
    return this._address.postcode || '';
  }

  get isValid(): boolean {
    return this._address.postcodeValid;
  }

  get value(): string {
    return this._address.raw;
  }

  set value(val: string) {
    this.setAddress(val);
  }

  get name(): string {
    return this.getAttribute('name') || 'address';
  }

  set name(val: string) {
    this.setAttribute('name', val);
  }

  get required(): boolean {
    return this.hasAttribute('required');
  }

  set required(val: boolean) {
    if (val) this.setAttribute('required', '');
    else this.removeAttribute('required');
  }

  get disabled(): boolean {
    return this.hasAttribute('disabled');
  }

  set disabled(val: boolean) {
    if (val) this.setAttribute('disabled', '');
    else this.removeAttribute('disabled');
  }

  // --- Public Methods ---

  public setAddress(rawText: string) {
    const val = rawText || '';
    if (this.searchInput) {
      this.searchInput.value = val;
      this.clearBtn?.classList.toggle('hidden', val.length === 0);
    }
    this.runDisambiguation(val);
  }

  public reset() {
    this._address = {
      raw: '',
      postcodeValid: false,
      confidence: 'low',
    };
    this._lastResolvedPostcode = null;
    if (this.searchInput) this.searchInput.value = '';
    if (this.clearBtn) this.clearBtn.classList.add('hidden');
    if (this.stateSelect) this.stateSelect.value = '';
    if (this.lgaSelect) {
      this.populateLGAs('', true);
      this.lgaSelect.value = '';
    }
    if (this.localityInput) this.localityInput.value = '';
    if (this.landmarkInput) this.landmarkInput.value = '';
    if (this.streetInput) this.streetInput.value = '';
    if (this.unitInput) this.unitInput.value = '';

    this.updateUI();
    this.syncFormValue();

    // Dispatch clear change event so listeners know the field was reset
    const detail: AddressChangeEventDetail = {
      address: { ...this._address },
      isValid: false,
      mode: 'smart',
    };
    this.dispatchEvent(new CustomEvent('nipost-change', { detail, bubbles: true, composed: true }));
  }

  // --- Internal Engine & WASM Loader ---

  private setupEngineLoading() {
    if (this._engine) return;
    const strategy = this.getAttribute('loading') || 'lazy';

    if (strategy === 'eager') {
      this.initEngine();
    } else if (strategy === 'idle') {
      if (typeof window !== 'undefined' && 'requestIdleCallback' in window) {
        (window as any).requestIdleCallback(() => this.initEngine(), { timeout: 3000 });
      } else {
        setTimeout(() => this.initEngine(), 800);
      }
    } else {
      // Lazy (default): load on first focus, hover, or input
      const trigger = () => {
        this.initEngine();
      };
      this.searchInput?.addEventListener('focusin', trigger, { once: true });
      this.searchInput?.addEventListener('pointerenter', trigger, { once: true });
      this.searchInput?.addEventListener('input', trigger, { once: true });
    }
  }

  private async initEngine() {
    if (this._engine) return;
    if ((globalThis as any).Postcode) {
      this.engine = (globalThis as any).Postcode;
      return;
    }

    const wasmUrl = this.getAttribute('wasm-url') || DEFAULT_WASM_CDN_URL;
    const wasmExecUrl = this.getAttribute('wasm-exec-url') || DEFAULT_WASM_EXEC_CDN_URL;
    const apiKey = this.getAttribute('api-key');
    const endpoint = this.getAttribute('endpoint');

    this._isEngineLoading = true;
    this.updateLoadingState();

    try {
      if (!sharedEnginePromise) {
        sharedEnginePromise = (async () => {
          // In browser environment, load wasm_exec.js if Go runtime is missing
          if (typeof (globalThis as any).Go === 'undefined') {
            if (typeof window !== 'undefined' && typeof document !== 'undefined' && document.head) {
              await new Promise<void>((resolve, reject) => {
                const script = document.createElement('script');
                script.src = wasmExecUrl;
                script.crossOrigin = 'anonymous';
                script.onload = () => resolve();
                script.onerror = () => reject(new Error(`Failed to load WebAssembly Go runtime from ${wasmExecUrl}`));
                document.head.appendChild(script);
              });
            }
          }

          let initWasmFn = (globalThis as any).initPostcode;
          if (!initWasmFn) {
            try {
              const mod = await import('@abcubed3/postcode-wasm');
              if (mod && mod.initPostcode) {
                initWasmFn = mod.initPostcode;
              }
            } catch {
              // Standalone fallback
            }
          }

          if (typeof initWasmFn !== 'function') {
            throw new Error('initPostcode loader function not found.');
          }

          const instance = await initWasmFn(wasmUrl, { apiKey, baseURL: endpoint });
          return instance;
        })().catch((err) => {
          sharedEnginePromise = null;
          throw err;
        });
      }

      const engine = await sharedEnginePromise;
      this.engine = engine;
    } catch (err: any) {
      this._isEngineLoading = false;
      this.updateLoadingState();
      this.populateStates();

      const errorDetail: AddressErrorEventDetail = {
        message: err?.message || 'Failed to initialize NIPOST Postcode WebAssembly engine.',
        context: 'engine-load',
        error: err,
      };
      this.dispatchEvent(new CustomEvent('nipost-error', { detail: errorDetail, bubbles: true, composed: true }));
    }
  }

  // --- Rendering UI Template ---

  private render() {
    const rawPlaceholder = this.getAttribute('placeholder') || 'Paste or type Nigerian address (e.g. 14 Admiralty Way, Lekki, Lagos)';
    const safePlaceholder = escapeHTML(rawPlaceholder);

    this.root.innerHTML = `
      <style>${componentStyles}</style>
      <div class="nipost-wrapper" part="container">
        <!-- Header -->
        <div class="nipost-header" part="header">
          <div class="nipost-title">
            <span class="flag">🇳🇬</span>
            <span>Nigerian Digital Address</span>
            <span class="loading-indicator hidden" id="loadingSpan">
              <span class="spinner"></span> <span>Engine loading...</span>
            </span>
          </div>
          <span class="nipost-badge" id="statusBadge" part="badge" role="status" aria-live="polite">Unverified</span>
        </div>

        <!-- Smart Search & Quick-Paste Bar -->
        <div class="search-container" part="search-container">
          <div class="search-input-wrapper">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input
              type="text"
              class="search-input"
              id="searchInput"
              part="search-input"
              aria-label="Nigerian address search"
              placeholder="${safePlaceholder}"
              autocomplete="street-address"
            />
            <button type="button" class="clear-btn hidden" id="clearBtn" title="Clear address" aria-label="Clear address">✕</button>
          </div>

          <button
            type="button"
            class="gps-btn ${this.getAttribute('allow-gps') === 'false' ? 'hidden' : ''}"
            id="gpsBtn"
            part="gps-btn"
            title="Use current device GPS location"
            aria-label="Use current device GPS location"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <polygon points="3 11 22 2 13 21 11 13 3 11"></polygon>
            </svg>
            <span>Current Location</span>
          </button>
        </div>

        <!-- Disambiguation Summary Pills -->
        <div class="summary-bar hidden" id="summaryBar" part="summary-bar" role="region" aria-label="Detected address components"></div>

        <!-- Postcode 5-Segment Chip Display -->
        <div class="postcode-display hidden" id="postcodeDisplay" part="postcode-display" role="status" aria-label="NIPOST 11-character postcode segments">
          <span class="postcode-title">NIPOST Code:</span>
          <div class="segment-chip" part="chip-state">
            <span class="segment-val" id="segState">--</span>
            <span class="segment-name">State</span>
          </div>
          <span class="segment-sep">-</span>
          <div class="segment-chip" part="chip-lga">
            <span class="segment-val" id="segLGA">--</span>
            <span class="segment-name">LGA</span>
          </div>
          <span class="segment-sep">-</span>
          <div class="segment-chip" part="chip-district">
            <span class="segment-val" id="segDistrict">--</span>
            <span class="segment-name">District</span>
          </div>
          <span class="segment-sep">-</span>
          <div class="segment-chip" part="chip-area">
            <span class="segment-val" id="segArea">--</span>
            <span class="segment-name">Area</span>
          </div>
          <span class="segment-sep">-</span>
          <div class="segment-chip" part="chip-unit">
            <span class="segment-val" id="segUnit">--</span>
            <span class="segment-name">Unit</span>
          </div>
        </div>

        <!-- Diagnostic Alert Box -->
        <div class="diagnostic-box hidden" id="diagnosticBox" part="diagnostic-box" role="status" aria-live="polite">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <div class="diagnostic-content">
            <span id="diagnosticMsg" part="diagnostic-tip"></span>
            <button type="button" class="repair-btn hidden" id="repairBtn" part="repair-btn"></button>
          </div>
        </div>

        <!-- Toggle Structured Breakdown Button -->
        <button
          type="button"
          class="expand-toggle"
          id="expandToggle"
          part="expand-btn"
          aria-expanded="false"
          aria-controls="structuredGrid"
        >
          <span>▼ Edit individual fields</span>
        </button>

        <!-- Structured Fields Grid -->
        <div class="structured-grid hidden" id="structuredGrid" part="form-grid" role="group" aria-label="Structured address breakdown">
          <div class="form-group">
            <label class="form-label" for="stateSelect">State</label>
            <select class="form-select" id="stateSelect" part="state-select"></select>
          </div>

          <div class="form-group">
            <label class="form-label" for="lgaSelect">LGA</label>
            <select class="form-select" id="lgaSelect" part="lga-select"></select>
          </div>

          <div class="form-group">
            <label class="form-label" for="localityInput">Locality / District / Area</label>
            <input type="text" class="form-input" id="localityInput" part="input" placeholder="e.g. Lekki Phase 1, Wuse II" />
          </div>

          <div class="form-group">
            <label class="form-label" for="landmarkInput">Landmark (Anchor)</label>
            <input type="text" class="form-input" id="landmarkInput" part="input" placeholder="e.g. Beside Total Filling Station" />
          </div>

          <div class="form-group full">
            <label class="form-label" for="streetInput">Street Address & House Number</label>
            <input type="text" class="form-input" id="streetInput" part="input" placeholder="e.g. Plot 14 Admiralty Way" />
          </div>

          <div class="form-group">
            <label class="form-label" for="unitInput">Building Unit (01-99)</label>
            <input type="text" class="form-input" id="unitInput" part="input" placeholder="e.g. 01" maxlength="2" />
          </div>
        </div>
      </div>
    `;

    // Cache DOM refs
    this.searchInput = this.root.querySelector('#searchInput') as HTMLInputElement;
    this.clearBtn = this.root.querySelector('#clearBtn') as HTMLButtonElement;
    this.gpsBtn = this.root.querySelector('#gpsBtn') as HTMLButtonElement;
    this.statusBadge = this.root.querySelector('#statusBadge') as HTMLElement;
    this.loadingSpan = this.root.querySelector('#loadingSpan') as HTMLElement;
    this.summaryBar = this.root.querySelector('#summaryBar') as HTMLElement;
    this.postcodeDisplay = this.root.querySelector('#postcodeDisplay') as HTMLElement;
    this.diagnosticBox = this.root.querySelector('#diagnosticBox') as HTMLElement;
    this.diagnosticMsg = this.root.querySelector('#diagnosticMsg') as HTMLElement;
    this.repairBtn = this.root.querySelector('#repairBtn') as HTMLButtonElement;
    this.expandToggle = this.root.querySelector('#expandToggle') as HTMLButtonElement;
    this.structuredGrid = this.root.querySelector('#structuredGrid') as HTMLElement;

    this.stateSelect = this.root.querySelector('#stateSelect') as HTMLSelectElement;
    this.lgaSelect = this.root.querySelector('#lgaSelect') as HTMLSelectElement;
    this.localityInput = this.root.querySelector('#localityInput') as HTMLInputElement;
    this.landmarkInput = this.root.querySelector('#landmarkInput') as HTMLInputElement;
    this.streetInput = this.root.querySelector('#streetInput') as HTMLInputElement;
    this.unitInput = this.root.querySelector('#unitInput') as HTMLInputElement;

    this.attachEventListeners();
    this.populateStates();
    this.populateLGAs('');
  }

  // --- Event Listeners ---

  private attachEventListeners() {
    this.searchInput.addEventListener('input', () => {
      const val = this.searchInput.value;
      this.clearBtn.classList.toggle('hidden', val.length === 0);

      if (this._debounceTimer) {
        clearTimeout(this._debounceTimer);
      }
      this._debounceTimer = setTimeout(() => {
        this.runDisambiguation(this.searchInput.value);
      }, 100);
    });

    this.clearBtn.addEventListener('click', () => {
      this.reset();
      this.searchInput.focus();
    });

    this.gpsBtn.addEventListener('click', () => {
      this.useCurrentLocation();
    });

    this.expandToggle.addEventListener('click', () => {
      this._isStructuredExpanded = !this._isStructuredExpanded;
      this.structuredGrid.classList.toggle('hidden', !this._isStructuredExpanded);
      this.expandToggle.setAttribute('aria-expanded', String(this._isStructuredExpanded));
      this.expandToggle.innerHTML = this._isStructuredExpanded
        ? '<span>▲ Hide individual fields</span>'
        : '<span>▼ Edit individual fields</span>';
    });

    // Structured fields change events
    this.stateSelect.addEventListener('change', () => {
      const stateCode = this.stateSelect.value;
      this.populateLGAs(stateCode);
      this.syncFromStructuredForm();
    });

    this.lgaSelect.addEventListener('change', () => this.syncFromStructuredForm());
    this.localityInput.addEventListener('input', () => this.syncFromStructuredForm());
    this.landmarkInput.addEventListener('input', () => this.syncFromStructuredForm());
    this.streetInput.addEventListener('input', () => this.syncFromStructuredForm());
    this.unitInput.addEventListener('input', () => this.syncFromStructuredForm());
  }

  private updateLoadingState() {
    if (this.loadingSpan) {
      this.loadingSpan.classList.toggle('hidden', !this._isEngineLoading);
    }
  }

  private populateStates() {
    if (!this.stateSelect) return;
    const currentVal = this.stateSelect.value;

    let states: StateOption[] = DEFAULT_STATES;
    if (this._engine) {
      try {
        if (typeof this._engine.referenceStatesOffline === 'function') {
          const offlineStates = this._engine.referenceStatesOffline();
          if (Array.isArray(offlineStates) && offlineStates.length > 0) {
            states = offlineStates;
          }
        } else if (typeof this._engine.listStates === 'function') {
          const res = this._engine.listStates();
          if (Array.isArray(res)) {
            states = res;
          } else if (res && typeof res === 'object') {
            states = Object.entries(res).map(([code, item]: [string, any]) => ({
              code,
              name: item.name || code,
              capital: item.capital,
              zone: item.zone,
            }));
          }
        }
      } catch (err) {
        console.warn('Failed to retrieve states from engine, using default list', err);
      }
    }

    this.stateSelect.replaceChildren(
      createOption('Select State...', ''),
      ...states.map((st) => createOption(`${st.name} (${st.code})`, st.code))
    );
    if (currentVal) {
      this.stateSelect.value = currentVal;
    }
  }

  private populateLGAs(stateCode: string, force = false) {
    if (!this.lgaSelect) return;
    const upperState = stateCode ? stateCode.trim().toUpperCase() : '';

    if (!force && this._renderedLgaState === upperState) {
      return;
    }
    this._renderedLgaState = upperState;

    if (!upperState) {
      this.lgaSelect.replaceChildren(createOption('Select LGA...', ''));
      return;
    }

    let lgas: LGAOption[] = getLGAsForState(upperState);

    // If engine provides additional dynamic metadata, merge cleanly
    if (this._engine && typeof this._engine.stateLGAs === 'function') {
      try {
        const engineLGAs = this._engine.stateLGAs(upperState);
        if (Array.isArray(engineLGAs) && engineLGAs.length > 0) {
          lgas = engineLGAs;
        }
      } catch {
        // Fallback to comprehensive 774 LGAs dataset
      }
    }

    this.lgaSelect.replaceChildren(
      createOption(`Select LGA (${lgas.length} available)...`, ''),
      ...lgas.map((lga) => createOption(`${lga.name} (${lga.code})`, lga.code))
    );
  }

  // --- Disambiguation Pipeline ---

  private runDisambiguation(rawText: string) {
    const trimmed = rawText.trim();
    if (!trimmed) {
      this.reset();
      return;
    }

    let result: NigerianAddress = {
      raw: trimmed,
      postcodeValid: false,
      confidence: 'low',
    };

    if (this._engine && typeof this._engine.extractAddress === 'function') {
      const extracted = this._engine.extractAddress(trimmed);
      result = {
        raw: trimmed,
        postcode: extracted.postcode,
        postcodeValid: !!extracted.postcode_valid,
        legacyPostcode: extracted.legacy_postcode,
        suggestedPrefix: extracted.suggested_prefix,
        stateCode: extracted.state_code,
        stateName: extracted.state_name,
        lgaCode: extracted.lga_code,
        lgaName: extracted.lga_name,
        locality: extracted.locality,
        street: extracted.street,
        landmark: extracted.landmark,
        latitude: extracted.latitude,
        longitude: extracted.longitude,
        confidence:
          extracted.confidence === 'high' || extracted.confidence === 'medium' || extracted.confidence === 'low'
            ? extracted.confidence
            : 'low',
      };
    } else {
      // Lightweight fallback while WASM loads
      result = this.lightweightDisambiguate(trimmed);
    }

    // Enhance with complete 774 LGA directory if LGA was not detected
    if (!result.lgaCode && result.stateCode) {
      const stateLgas = getLGAsForState(result.stateCode);
      const upper = trimmed.toUpperCase();
      for (const lga of stateLgas) {
        if (upper.includes(lga.name.toUpperCase())) {
          result.lgaCode = lga.code;
          result.lgaName = lga.name;
          result.suggestedPrefix = `${result.stateCode}-${result.lgaCode}`;
          break;
        }
      }
    }

    this._address = result;
    this.syncStructuredInputs(result);
    this.updateUI();
    this.syncFormValue();

    // Dispatch custom event
    const detail: AddressChangeEventDetail = {
      address: { ...this._address },
      isValid: this._address.postcodeValid,
      mode: 'smart',
    };
    this.dispatchEvent(new CustomEvent('nipost-change', { detail, bubbles: true, composed: true }));

    // Only fire nipost-resolve when a genuine, complete postcode is validated and changed
    if (this._address.postcodeValid && this._address.postcode) {
      if (this._lastResolvedPostcode !== this._address.postcode) {
        this._lastResolvedPostcode = this._address.postcode;
        const displayParts = [
          this._address.street,
          this._address.landmark,
          this._address.locality,
          this._address.lgaName,
          this._address.stateName,
        ].filter(Boolean);
        const resolveDetail: AddressResolveEventDetail = {
          address: { ...this._address },
          postcode: this._address.postcode,
          formattedDisplay: `${displayParts.join(', ')} [${this._address.postcode}]`,
        };
        this.dispatchEvent(new CustomEvent('nipost-resolve', { detail: resolveDetail, bubbles: true, composed: true }));
      }
    } else {
      this._lastResolvedPostcode = null;
    }
  }

  private lightweightDisambiguate(raw: string): NigerianAddress {
    const upper = raw.toUpperCase();
    let stateCode: string | undefined;
    let stateName: string | undefined;

    for (const st of DEFAULT_STATES) {
      if (upper.includes(st.name.toUpperCase())) {
        stateCode = st.code;
        stateName = st.name;
        break;
      }
    }

    // Check FCT / Abuja alias
    if (!stateCode && (upper.includes('ABUJA') || upper.includes('FCT'))) {
      stateCode = 'FC';
      stateName = 'Federal Capital Territory';
    }

    // Landmark detection
    const landmarkMatch = raw.match(/(?:opposite|beside|near|behind|close to|after|in front of|under)\s+[^,;]+/i);
    const landmark = landmarkMatch ? landmarkMatch[0].trim() : undefined;

    // Street detection (clauses starting with Plot/No/House/Flat/number or ending in street/road/way/avenue/etc.)
    let street: string | undefined;
    const clauses = raw.split(/[,;\n]/).map(c => c.trim()).filter(Boolean);
    for (const clause of clauses) {
      if (landmark && (clause.includes(landmark) || landmark.includes(clause))) continue;
      if (
        /^(?:(?:plot|no\.?|house|flat)\s+\d+|\d+)\s+/i.test(clause) ||
        /\b(?:street|str|st|road|rd|avenue|ave|crescent|cres|close|cl|way|boulevard|blvd|lane|ln|expressway|exp)\b/i.test(clause)
      ) {
        street = clause;
        break;
      }
    }

    // Locality detection for common high-density zones
    let locality: string | undefined;
    const commonLocalities = ['Lekki Phase 1', 'Lekki Phase 2', 'Victoria Island', 'Ikoyi', 'Wuse II', 'Wuse 2', 'Gwarinpa', 'Maitama', 'Asokoro', 'Festac Town'];
    for (const loc of commonLocalities) {
      if (upper.includes(loc.toUpperCase())) {
        locality = loc;
        break;
      }
    }

    // LGA detection from 774 LGA database
    let lgaCode: string | undefined;
    let lgaName: string | undefined;
    if (stateCode) {
      const stateLgas = getLGAsForState(stateCode);
      for (const lga of stateLgas) {
        if (upper.includes(lga.name.toUpperCase())) {
          lgaCode = lga.code;
          lgaName = lga.name;
          break;
        }
      }
    }

    return {
      raw,
      stateCode,
      stateName,
      lgaCode,
      lgaName,
      locality,
      street,
      landmark,
      suggestedPrefix: stateCode ? (lgaCode ? `${stateCode}-${lgaCode}` : stateCode) : undefined,
      postcodeValid: false,
      confidence: (stateCode && lgaCode) ? 'high' : (stateCode ? 'medium' : 'low'),
    };
  }

  private syncStructuredInputs(addr: NigerianAddress) {
    if (this.stateSelect) {
      this.stateSelect.value = addr.stateCode || '';
      if (addr.stateCode) {
        this.populateLGAs(addr.stateCode);
      }
    }
    if (this.lgaSelect) {
      this.lgaSelect.value = addr.lgaCode || '';
    }
    if (this.localityInput) {
      this.localityInput.value = addr.locality || '';
    }
    if (this.landmarkInput) {
      this.landmarkInput.value = addr.landmark || '';
    }
    if (this.streetInput) {
      this.streetInput.value = addr.street || '';
    }
  }

  private syncFromStructuredForm() {
    const stateCode = this.stateSelect?.value || '';
    const lgaCode = this.lgaSelect?.value || '';
    const locality = this.localityInput?.value.trim() || '';
    const landmark = this.landmarkInput?.value.trim() || '';
    const street = this.streetInput?.value.trim() || '';
    const unit = this.unitInput?.value.trim() || '';

    // Construct unified address string
    const parts = [street, landmark, locality];
    const lgaName = lgaCode && stateCode ? getLGAsForState(stateCode).find(l => l.code === lgaCode)?.name : undefined;
    const stateName = DEFAULT_STATES.find(s => s.code === stateCode)?.name || stateCode;

    if (lgaName) parts.push(lgaName);
    if (stateName) parts.push(stateName);

    const compositeRaw = parts.filter(Boolean).join(', ');
    if (this.searchInput) {
      this.searchInput.value = compositeRaw;
      this.clearBtn?.classList.toggle('hidden', compositeRaw.length === 0);
    }

    let postcode: string | undefined;
    let postcodeValid = false;
    let suggestedPrefix: string | undefined;

    if (stateCode && lgaCode) {
      suggestedPrefix = `${stateCode}-${lgaCode}`;
      // Only synthesize and validate a postcode candidate if a valid building unit (01-99) was provided
      if (/^(0[1-9]|[1-9][0-9])$/.test(unit)) {
        const candidate = `${stateCode}-${lgaCode}-A01-AA-${unit}`;
        if (this._engine && typeof this._engine.validate === 'function') {
          const valRes = this._engine.validate(candidate);
          postcode = candidate;
          postcodeValid = valRes.valid;
        } else {
          postcode = candidate;
          postcodeValid = true;
        }
      }
    }

    this._address = {
      raw: compositeRaw,
      stateCode: stateCode || undefined,
      stateName: stateName || undefined,
      lgaCode: lgaCode || undefined,
      lgaName,
      locality: locality || undefined,
      landmark: landmark || undefined,
      street: street || undefined,
      unit: unit || undefined,
      suggestedPrefix,
      postcode,
      postcodeValid,
      confidence: (stateCode && lgaCode) ? 'high' : 'medium',
    };

    this.updateUI();
    this.syncFormValue();

    const detail: AddressChangeEventDetail = {
      address: { ...this._address },
      isValid: this._address.postcodeValid,
      mode: 'structured',
    };
    this.dispatchEvent(new CustomEvent('nipost-change', { detail, bubbles: true, composed: true }));
  }

  // --- UI Update & Diagnostic Feedback ---

  private updateUI() {
    const addr = this._address;

    // Status Badge
    if (this.statusBadge) {
      if (addr.postcodeValid) {
        this.statusBadge.textContent = 'Verified Postcode';
        this.statusBadge.className = 'nipost-badge valid';
      } else if (addr.stateCode || addr.lgaCode) {
        this.statusBadge.textContent = 'Partial Match';
        this.statusBadge.className = 'nipost-badge';
      } else {
        this.statusBadge.textContent = 'Unverified';
        this.statusBadge.className = 'nipost-badge';
      }
    }

    // Summary Pills - Safe DOM Node Creation to Prevent Stored XSS
    if (this.summaryBar) {
      this.summaryBar.replaceChildren();
      const items: HTMLElement[] = [];

      const createPill = (label: string, value: string, isLandmark = false) => {
        const span = document.createElement('span');
        span.className = isLandmark ? 'tag-item tag-landmark' : 'tag-item';

        if (isLandmark) {
          span.textContent = '📍 ';
          const b = document.createElement('b');
          b.textContent = value;
          span.appendChild(b);
        } else {
          const lSpan = document.createElement('span');
          lSpan.className = 'tag-label';
          lSpan.textContent = label + ':';
          const b = document.createElement('b');
          b.textContent = value;
          span.append(lSpan, ' ', b);
        }
        return span;
      };

      if (addr.stateName) {
        items.push(createPill('State', addr.stateName));
      }
      if (addr.lgaName) {
        items.push(createPill('LGA', addr.lgaName));
      }
      if (addr.locality) {
        items.push(createPill('Area', addr.locality));
      }
      if (addr.landmark) {
        items.push(createPill('Landmark', addr.landmark, true));
      }

      if (items.length > 0) {
        this.summaryBar.append(...items);
        this.summaryBar.classList.remove('hidden');
      } else {
        this.summaryBar.classList.add('hidden');
      }
    }

    // 5-Segment Postcode Chip Display
    if (this.postcodeDisplay) {
      if (addr.postcode) {
        const segs = addr.postcode.split('-');
        if (segs.length === 5) {
          (this.root.querySelector('#segState') as HTMLElement).textContent = segs[0];
          (this.root.querySelector('#segLGA') as HTMLElement).textContent = segs[1];
          (this.root.querySelector('#segDistrict') as HTMLElement).textContent = segs[2];
          (this.root.querySelector('#segArea') as HTMLElement).textContent = segs[3];
          (this.root.querySelector('#segUnit') as HTMLElement).textContent = segs[4];
          this.postcodeDisplay.classList.remove('hidden');
        } else {
          this.postcodeDisplay.classList.add('hidden');
        }
      } else {
        this.postcodeDisplay.classList.add('hidden');
      }
    }

    // Diagnostics & Self-Healing Auto-Repair Tips
    if (this.diagnosticBox && this.diagnosticMsg) {
      let tip: string | undefined;
      let repairedCode: string | undefined;

      if (addr.legacyPostcode && !addr.postcodeValid) {
        const prefix = addr.suggestedPrefix || (addr.stateCode ? `${addr.stateCode}-${addr.lgaCode || '01'}` : 'EK-01');
        tip = `Legacy 6-digit code detected (${addr.legacyPostcode}). NIPOST modern postcodes use 11 characters (e.g. ${prefix}-A01-AA-01).`;
      } else if (addr.postcode && !addr.postcodeValid && this._engine) {
        if (typeof this._engine.diagnose === 'function') {
          const diag = this._engine.diagnose(addr.postcode);
          if (!diag.valid && diag.actionable_tip) {
            tip = diag.actionable_tip;
          }
        }
        if (typeof this._engine.repair === 'function') {
          const rep = this._engine.repair(addr.postcode);
          if (rep.repaired && rep.postcode) {
            repairedCode = rep.postcode;
            tip = (tip ? `${tip} ` : '') + `Did you mean ${rep.postcode}?`;
          }
        }
      }

      if (tip) {
        this.diagnosticMsg.textContent = tip;
        this.diagnosticBox.className = 'diagnostic-box warning';
        this.diagnosticBox.classList.remove('hidden');

        if (repairedCode && this.repairBtn) {
          this.repairBtn.textContent = `Auto-repair: Apply ${repairedCode}`;
          this.repairBtn.classList.remove('hidden');
          this.repairBtn.onclick = () => {
            if (this.searchInput) {
              const currentText = this.searchInput.value;
              const rep = repairedCode!;
              const target = addr.postcode || '';
              const updated = target && currentText.includes(target)
                ? currentText.replace(target, rep)
                : (currentText ? `${currentText} ${rep}` : rep);
              this.setAddress(updated);
              this.searchInput.focus();
            }
          };
        } else if (this.repairBtn) {
          this.repairBtn.classList.add('hidden');
          this.repairBtn.onclick = null;
        }

        const diagDetail: AddressDiagnoseEventDetail = {
          raw: addr.raw,
          postcodeCandidate: addr.postcode,
          tip,
          isValid: false,
        };
        this.dispatchEvent(new CustomEvent('nipost-diagnose', { detail: diagDetail, bubbles: true, composed: true }));
      } else {
        this.diagnosticBox.classList.add('hidden');
        if (this.repairBtn) {
          this.repairBtn.classList.add('hidden');
          this.repairBtn.onclick = null;
        }
      }
    }
  }

  // --- GPS Location Reverse Geocoding ---

  private useCurrentLocation() {
    if (!navigator.geolocation) {
      const msg = 'Geolocation is not supported by your browser.';
      this.dispatchEvent(
        new CustomEvent('nipost-error', {
          detail: { message: msg, context: 'geolocation' },
          bubbles: true,
          composed: true,
        })
      );
      if (this.diagnosticBox && this.diagnosticMsg) {
        this.diagnosticMsg.textContent = msg;
        this.diagnosticBox.className = 'diagnostic-box warning';
        this.diagnosticBox.classList.remove('hidden');
      }
      return;
    }

    this.gpsBtn.disabled = true;
    const originalText = this.gpsBtn.innerHTML;
    this.gpsBtn.innerHTML = '<span class="spinner" aria-hidden="true"></span> <span>Locating...</span>';

    navigator.geolocation.getCurrentPosition(
      (pos) => {
        this.gpsBtn.disabled = false;
        this.gpsBtn.innerHTML = originalText;
        const lat = pos.coords.latitude;
        const lng = pos.coords.longitude;

        let foundStateCode = '';
        let foundStateName = '';
        let foundLgaCode = '';
        let foundLgaName = '';
        let foundBuilding = '';

        // 1. Try offline building lookup from WASM database
        if (this._engine && typeof this._engine.reverseCoordinatesOffline === 'function') {
          try {
            const rev = this._engine.reverseCoordinatesOffline(lat, lng, 150);
            if (rev && rev.found && rev.unit) {
              foundStateCode = rev.state_code || '';
              foundStateName = rev.state_name || rev.unit.state_name || '';
              foundLgaCode = rev.lga_code || '';
              foundLgaName = rev.lga_name || rev.unit.lga_name || '';
              foundBuilding = rev.unit.display || rev.unit.address || '';
            }
          } catch (e) {
            console.warn('reverseCoordinatesOffline lookup failed:', e);
          }
        }

        // 2. If no registered building within search radius, find nearest LGA centroid
        if (!foundLgaCode) {
          if (this._engine && typeof this._engine.nearestLGA === 'function') {
            try {
              const nearest = this._engine.nearestLGA(lat, lng);
              if (nearest && nearest.found && nearest.lga_code) {
                foundStateCode = nearest.state_code || '';
                foundStateName = nearest.state_name || '';
                foundLgaCode = nearest.lga_code || '';
                foundLgaName = nearest.lga_name || '';
              }
            } catch (e) {
              console.warn('nearestLGA lookup failed:', e);
            }
          }

          if (!foundLgaCode) {
            const nearestFallback = findNearestLGA(lat, lng);
            if (nearestFallback) {
              foundStateCode = nearestFallback.lga.state_code || '';
              foundStateName = DEFAULT_STATES.find(s => s.code === foundStateCode)?.name || foundStateCode;
              foundLgaCode = nearestFallback.lga.code;
              foundLgaName = nearestFallback.lga.name;
            }
          }
        }

        if (foundStateCode && foundLgaCode) {
          this._address.stateCode = foundStateCode;
          this._address.stateName = foundStateName;
          this._address.lgaCode = foundLgaCode;
          this._address.lgaName = foundLgaName;
          this._address.latitude = lat;
          this._address.longitude = lng;

          if (this.stateSelect) {
            this.stateSelect.value = foundStateCode;
            this.populateLGAs(foundStateCode);
          }
          if (this.lgaSelect) {
            this.lgaSelect.value = foundLgaCode;
          }

          const locText = foundBuilding
            ? `${foundBuilding}, ${foundLgaName}, ${foundStateName}`
            : `${foundLgaName}, ${foundStateName}`;
          if (this.searchInput) {
            this.searchInput.value = locText;
          }
          this.runDisambiguation(locText);
          return;
        }

        // Generic fallback if coordinates fall outside Nigerian cadastral coverage
        if (this.searchInput) {
          this.searchInput.value = `[GPS: ${lat.toFixed(4)}, ${lng.toFixed(4)}]`;
        }
      },
      (err) => {
        this.gpsBtn.disabled = false;
        this.gpsBtn.innerHTML = originalText;
        console.warn('Geolocation error:', err.message);
        this.dispatchEvent(
          new CustomEvent('nipost-error', {
            detail: {
              message: err.message || 'Geolocation access denied or timed out.',
              context: 'geolocation',
              error: err,
            },
            bubbles: true,
            composed: true,
          })
        );
      },
      { timeout: 10000, enableHighAccuracy: true }
    );
  }

  // --- Form Associated Callback Support ---

  private syncFormValue() {
    if (!this._internals) return;

    const name = this.name;
    const addr = this._address;

    const data = new FormData();
    data.append(`${name}[raw]`, addr.raw || '');
    if (addr.stateCode) data.append(`${name}[state]`, addr.stateCode);
    if (addr.lgaCode) data.append(`${name}[lga]`, addr.lgaCode);
    if (addr.locality) data.append(`${name}[locality]`, addr.locality);
    if (addr.landmark) data.append(`${name}[landmark]`, addr.landmark);
    if (addr.street) data.append(`${name}[street]`, addr.street);
    if (addr.postcode) data.append(`${name}[postcode]`, addr.postcode);
    if (addr.latitude) data.append(`${name}[latitude]`, String(addr.latitude));
    if (addr.longitude) data.append(`${name}[longitude]`, String(addr.longitude));

    this._internals.setFormValue(data);

    if (this.required && !addr.raw.trim()) {
      this._internals.setValidity(
        { valueMissing: true },
        'Please provide a Nigerian address.',
        this.searchInput
      );
    } else {
      this._internals.setValidity({});
    }
  }

  private setDisabled(disabled: boolean) {
    if (this.searchInput) this.searchInput.disabled = disabled;
    if (this.clearBtn) this.clearBtn.disabled = disabled;
    if (this.gpsBtn) this.gpsBtn.disabled = disabled;
    if (this.stateSelect) this.stateSelect.disabled = disabled;
    if (this.lgaSelect) this.lgaSelect.disabled = disabled;
    if (this.localityInput) this.localityInput.disabled = disabled;
    if (this.landmarkInput) this.landmarkInput.disabled = disabled;
    if (this.streetInput) this.streetInput.disabled = disabled;
    if (this.unitInput) this.unitInput.disabled = disabled;
  }
}
