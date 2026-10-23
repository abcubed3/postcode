import { describe, it, expect, beforeEach, vi } from 'vitest';
import '../src/index';
import { NipostAddressInputElement } from '../src/component';
import type { PostcodeWasmModule } from '../src/types';
import { NIGERIAN_774_LGAS, getTotalLGACount, getLGAsForState, findNearestLGA } from '../src/data/lgas';

describe('<nipost-address-input>', () => {
  let element: NipostAddressInputElement;

  beforeEach(() => {
    document.body.innerHTML = '';
    element = document.createElement('nipost-address-input');
    document.body.appendChild(element);
  });

  it('registers custom element correctly', () => {
    expect(customElements.get('nipost-address-input')).toBeDefined();
    expect(element).toBeInstanceOf(NipostAddressInputElement);
    expect(element.shadowRoot).toBeTruthy();
  });

  it('has proper default attributes and properties', () => {
    expect(element.name).toBe('address');
    expect(element.required).toBe(false);
    expect(element.disabled).toBe(false);
    expect(element.value).toBe('');
    expect(element.isValid).toBe(false);
  });

  it('handles name, required, and disabled changes', () => {
    element.name = 'shipping_address';
    expect(element.getAttribute('name')).toBe('shipping_address');

    element.required = true;
    expect(element.hasAttribute('required')).toBe(true);

    element.disabled = true;
    expect(element.hasAttribute('disabled')).toBe(true);
    const searchInput = element.shadowRoot?.querySelector('#searchInput') as HTMLInputElement;
    expect(searchInput.disabled).toBe(true);
  });

  it('dispatches nipost-change event on setting address', () => {
    const changeSpy = vi.fn();
    element.addEventListener('nipost-change', changeSpy);

    element.setAddress('14 Admiralty Way, Lekki, Lagos');

    expect(changeSpy).toHaveBeenCalled();
    const eventDetail = changeSpy.mock.calls[0][0].detail;
    expect(eventDetail.address.raw).toBe('14 Admiralty Way, Lekki, Lagos');
    expect(eventDetail.address.stateCode).toBe('LA');
    expect(element.value).toBe('14 Admiralty Way, Lekki, Lagos');
  });

  it('handles real WASM engine shape where listStates returns an Object map', () => {
    const mockEngine: PostcodeWasmModule = {
      validate: vi.fn().mockReturnValue({ valid: true }),
      diagnose: vi.fn().mockReturnValue({ valid: true, input: 'LA-08-A01-AA-01' }),
      repair: vi.fn().mockReturnValue({ repaired: false, original: 'LA-08-A01-AA-01', postcode: 'LA-08-A01-AA-01' }),
      extractAddress: vi.fn().mockReturnValue({
        raw: 'Plot 14 Admiralty Way, Lekki Phase 1, Opposite Dominos, Lagos',
        postcode: 'LA-08-A86-RG-01',
        postcode_valid: true,
        state_code: 'LA',
        state_name: 'Lagos',
        lga_code: '08',
        lga_name: 'Eti-Osa',
        locality: 'Lekki Phase 1',
        street: '14 Admiralty Way',
        landmark: "Opposite Domino's",
        confidence: 'high',
      }),
      // Real Go WASM jsListStates returns a map of { LA: { name: 'Lagos' } }, not an array!
      listStates: vi.fn().mockReturnValue({
        LA: { name: 'Lagos', capital: 'Ikeja', zone: 'SOUTH WEST' },
        EK: { name: 'Ekiti', capital: 'Ado-Ekiti', zone: 'SOUTH WEST' },
      }),
      referenceStatesOffline: vi.fn().mockReturnValue([
        { code: 'LA', name: 'Lagos' },
        { code: 'EK', name: 'Ekiti' },
      ]),
      stateLGAs: vi.fn().mockReturnValue([{ code: '08', name: 'Eti-Osa' }]),
      nearestLGA: vi.fn().mockReturnValue({
        found: true,
        state_code: 'LA',
        state_name: 'Lagos',
        lga_code: '08',
        lga_name: 'Eti-Osa',
        distance_meters: 10,
      }),
      reverseCoordinatesOffline: vi.fn().mockReturnValue({
        found: true,
        state_code: 'LA',
        state_name: 'Lagos',
        lga_code: '08',
        lga_name: 'Eti-Osa',
      }),
      setAPIKey: vi.fn(),
    };

    const resolveSpy = vi.fn();
    element.addEventListener('nipost-resolve', resolveSpy);

    // Setting engine should NOT crash when listStates() returns an Object
    expect(() => {
      element.engine = mockEngine;
    }).not.toThrow();

    element.setAddress('Plot 14 Admiralty Way, Lekki Phase 1, Opposite Dominos, Lagos');

    expect(mockEngine.extractAddress).toHaveBeenCalledWith(
      'Plot 14 Admiralty Way, Lekki Phase 1, Opposite Dominos, Lagos'
    );
    expect(element.isValid).toBe(true);
    expect(element.postcode).toBe('LA-08-A86-RG-01');
    expect(resolveSpy).toHaveBeenCalled();
  });

  it('prevents stored XSS when address text contains HTML tags', () => {
    const xssPayload = 'near <img src=x onerror=alert(1)> Lekki, Lagos';
    element.setAddress(xssPayload);

    const shadowRoot = element.shadowRoot!;
    // Ensure no <img> element was parsed into the ShadowRoot DOM
    const imgEl = shadowRoot.querySelector('img');
    expect(imgEl).toBeNull();

    const summaryBar = shadowRoot.querySelector('#summaryBar');
    expect(summaryBar).toBeTruthy();
    // Raw HTML tag should be safely escaped as plain text
    expect(summaryBar?.innerHTML).not.toContain('<img');
    expect(summaryBar?.textContent).toContain('<img src=x onerror=alert(1)>');
  });

  it('resets inputs cleanly and dispatches nipost-change on element.reset()', () => {
    const changeSpy = vi.fn();
    element.addEventListener('nipost-change', changeSpy);

    element.setAddress('Plot 12, Wuse II, Abuja');
    expect(element.value).toBe('Plot 12, Wuse II, Abuja');

    element.reset();
    expect(element.value).toBe('');
    expect(element.postcode).toBe('');
    expect(element.isValid).toBe(false);

    // Verify reset notified listeners
    expect(changeSpy).toHaveBeenCalledTimes(2);
    expect(changeSpy.mock.calls[1][0].detail.address.raw).toBe('');
    expect(changeSpy.mock.calls[1][0].detail.isValid).toBe(false);
  });

  it('guarantees all 774 LGAs across Nigeria with legal 2-digit decimal codes', () => {
    // 1. Exact total count is 774
    expect(getTotalLGACount()).toBe(774);

    // 2. All 37 state entities (36 states + FCT) are represented
    const stateKeys = Object.keys(NIGERIAN_774_LGAS);
    expect(stateKeys.length).toBe(37);

    // 3. Kano has 44, Lagos has 20, Oyo has 33, FCT has 6
    expect(getLGAsForState('KN').length).toBe(44);
    expect(getLGAsForState('LA').length).toBe(20);
    expect(getLGAsForState('OY').length).toBe(33);
    expect(getLGAsForState('FC').length).toBe(6);

    // 4. Verify no illegal '00' codes (Rule 1: decimal 01-99)
    for (const [_, lgas] of Object.entries(NIGERIAN_774_LGAS)) {
      for (const lga of lgas) {
        expect(lga.code).toMatch(/^[0-9]{2}$/);
        expect(lga.code).not.toBe('00');
        expect(parseInt(lga.code, 10)).toBeGreaterThanOrEqual(1);
        expect(parseInt(lga.code, 10)).toBeLessThanOrEqual(99);
      }
    }
  });

  it('finds nearest LGA centroid using offline haversine search', () => {
    // Lekki Phase 1 centroid: ~6.4584, 3.5684 -> Eti-Osa (LA-08)
    const lekki = findNearestLGA(6.4584, 3.5684);
    expect(lekki).toBeTruthy();
    expect(lekki?.lga.state_code).toBe('LA');
    expect(lekki?.lga.code).toBe('08');
    expect(lekki?.lga.name).toBe('Eti-Osa');
    expect(lekki?.distanceM).toBeLessThan(100);

    // Ikeja Secretariat: ~6.6018, 3.3515 -> Ikeja (LA-11)
    const ikeja = findNearestLGA(6.6018, 3.3515);
    expect(ikeja).toBeTruthy();
    expect(ikeja?.lga.state_code).toBe('LA');
    expect(ikeja?.lga.code).toBe('11');
    expect(ikeja?.lga.name).toBe('Ikeja');
  });

  it('populates full LGA list when a state is selected in the UI', () => {
    const stateSelect = element.shadowRoot?.querySelector('#stateSelect') as HTMLSelectElement;
    const lgaSelect = element.shadowRoot?.querySelector('#lgaSelect') as HTMLSelectElement;

    // Select Kano State (KN)
    stateSelect.value = 'KN';
    stateSelect.dispatchEvent(new Event('change'));

    // Options: 1 placeholder + 44 LGAs = 45 options
    expect(lgaSelect.options.length).toBe(45);
    expect(lgaSelect.options[1].text).toContain('Ajingi (01)');
    expect(lgaSelect.options[44].text).toContain('Wudil (44)');

    // Select FCT (FC)
    stateSelect.value = 'FC';
    stateSelect.dispatchEvent(new Event('change'));

    // Options: 1 placeholder + 6 Area Councils = 7 options
    expect(lgaSelect.options.length).toBe(7);
  });

  it('provides interactive auto-repair button when postcode has fixable typo', () => {
    const mockEngine: PostcodeWasmModule = {
      validate: vi.fn().mockReturnValue({ valid: false }),
      diagnose: vi.fn().mockReturnValue({
        valid: false,
        input: 'EK-1-AO3-FK-1',
        actionable_tip: 'Segment length invalid.',
      }),
      repair: vi.fn().mockReturnValue({
        repaired: true,
        original: 'EK-1-AO3-FK-1',
        postcode: 'EK-01-A03-FK-01',
      }),
      extractAddress: vi.fn().mockReturnValue({
        raw: 'Fabian Hotel, EK-1-AO3-FK-1, Ado Ekiti',
        postcode: 'EK-1-AO3-FK-1',
        postcode_valid: false,
        state_code: 'EK',
        state_name: 'Ekiti',
        confidence: 'medium',
      }),
      listStates: vi.fn().mockReturnValue([{ code: 'EK', name: 'Ekiti' }]),
      stateLGAs: vi.fn().mockReturnValue([{ code: '01', name: 'Ado Ekiti' }]),
      reverseCoordinatesOffline: vi.fn().mockReturnValue({ found: false }),
      setAPIKey: vi.fn(),
    };

    element.engine = mockEngine;
    element.setAddress('Fabian Hotel, EK-1-AO3-FK-1, Ado Ekiti');

    const shadowRoot = element.shadowRoot!;
    const repairBtn = shadowRoot.querySelector('#repairBtn') as HTMLButtonElement;
    expect(repairBtn).toBeTruthy();
    expect(repairBtn.classList.contains('hidden')).toBe(false);
    expect(repairBtn.textContent).toContain('Apply EK-01-A03-FK-01');

    // Simulate clicking the auto-repair button
    repairBtn.click();
    expect(element.value).toContain('EK-01-A03-FK-01');
  });

  it('clears structured input fields that are absent when switching address', () => {
    element.setAddress('Plot 14 Admiralty Way, beside Zenith Bank, Lekki Phase 1, Eti-Osa, Lagos');

    const shadowRoot = element.shadowRoot!;
    const streetInput = shadowRoot.querySelector('#streetInput') as HTMLInputElement;
    const landmarkInput = shadowRoot.querySelector('#landmarkInput') as HTMLInputElement;

    expect(streetInput.value).toBe('Plot 14 Admiralty Way');
    expect(landmarkInput.value).toBe('beside Zenith Bank');

    // Now set an address without landmark or street
    element.setAddress('Kano, Nigeria');
    expect(streetInput.value).toBe('');
    expect(landmarkInput.value).toBe('');
  });

  it('updates API key and endpoint on engine dynamically when attributes change', () => {
    const setKeySpy = vi.fn();
    const configSpy = vi.fn();
    const mockEngine: any = {
      validate: vi.fn().mockReturnValue({ valid: true }),
      diagnose: vi.fn().mockReturnValue({ valid: true }),
      repair: vi.fn().mockReturnValue({ repaired: false }),
      extractAddress: vi.fn().mockReturnValue({ raw: '' }),
      listStates: vi.fn().mockReturnValue([]),
      stateLGAs: vi.fn().mockReturnValue([]),
      setAPIKey: setKeySpy,
      configure: configSpy,
    };

    element.engine = mockEngine;
    element.setAttribute('api-key', 'nipost_key_live_999');
    expect(setKeySpy).toHaveBeenCalledWith('nipost_key_live_999');

    element.setAttribute('endpoint', 'https://proxy.example.com/v1');
    expect(configSpy).toHaveBeenCalledWith({ baseURL: 'https://proxy.example.com/v1' });
  });

  it('clears on formResetCallback', () => {
    element.setAddress('14 Admiralty Way, Lekki, Lagos');
    expect(element.value).toBe('14 Admiralty Way, Lekki, Lagos');

    element.formResetCallback();
    expect(element.value).toBe('');
    expect(element.isValid).toBe(false);
  });
});
