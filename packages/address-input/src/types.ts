/**
 * Type definitions for @abcubed3/nipost-address-input
 */

export interface NigerianAddress {
  raw: string;
  postcode?: string;
  postcodeValid: boolean;
  legacyPostcode?: string;
  suggestedPrefix?: string;
  stateCode?: string;
  stateName?: string;
  lgaCode?: string;
  lgaName?: string;
  locality?: string;
  street?: string;
  landmark?: string;
  unit?: string;
  latitude?: number;
  longitude?: number;
  confidence: 'high' | 'medium' | 'low';
}

export interface StateOption {
  code: string;
  name: string;
  capital?: string;
  zone?: string;
}

export interface LGAOption {
  code: string;
  name: string;
  state_code?: string;
  latitude?: number;
  longitude?: number;
}

export interface DiagnosticResult {
  valid: boolean;
  input: string;
  actionable_tip?: string;
  error?: string;
  clean_length?: number;
  is_temporary?: boolean;
}

export interface PostcodeWasmModule {
  validate(code: string): { valid: boolean; state_name?: string; lga_name?: string; error?: string };
  diagnose(code: string): DiagnosticResult;
  repair(code: string): { repaired: boolean; original: string; postcode: string; method?: string };
  extractAddress(raw: string): {
    raw: string;
    postcode?: string;
    postcode_valid?: boolean;
    legacy_postcode?: string;
    suggested_prefix?: string;
    state_code?: string;
    state_name?: string;
    lga_code?: string;
    lga_name?: string;
    locality?: string;
    street?: string;
    landmark?: string;
    latitude?: number;
    longitude?: number;
    confidence?: string;
  };
  listStates(): Record<string, { name: string; capital?: string; zone?: string }> | StateOption[];
  referenceStatesOffline?(): StateOption[];
  stateLGAs(stateCode: string): LGAOption[];
  nearestLGA?(lat: number, lng: number): {
    found: boolean;
    state_code?: string;
    state_name?: string;
    lga_code?: string;
    lga_name?: string;
    latitude?: number;
    longitude?: number;
    distance_meters?: number;
    postcode_prefix?: string;
    error?: string;
  };
  reverseCoordinatesOffline(lat: number, lng: number, maxDist?: number): {
    found: boolean;
    coordinate?: [number, number];
    unit?: { postcode: string; display?: string; address?: string; state_name?: string; lga_name?: string; distance_m?: number };
    state_code?: string;
    state_name?: string;
    lga_code?: string;
    lga_name?: string;
    postcode?: string;
    distance_meters?: number;
    postcode_prefix?: string;
    message?: string;
    error?: string;
  };
  setAPIKey(key: string): void;
  autocomplete?(query: string): Promise<{ results: any[] }>;
  lookup?(code: string, level?: number): Promise<any>;
}

export interface AddressChangeEventDetail {
  address: NigerianAddress;
  isValid: boolean;
  mode: 'smart' | 'structured';
}

export interface AddressResolveEventDetail {
  address: NigerianAddress;
  postcode: string;
  formattedDisplay: string;
}

export interface AddressDiagnoseEventDetail {
  raw: string;
  postcodeCandidate?: string;
  tip?: string;
  isValid: boolean;
}

export interface AddressErrorEventDetail {
  message: string;
  context: 'engine-load' | 'geolocation' | 'network' | 'validation';
  error?: unknown;
}
