/**
 * TypeScript definitions for Postcode Nigeria WebAssembly module.
 */

export interface ValidationResult {
  valid: boolean;
  postcode?: string;
  compact?: string;
  state_code?: string;
  state_name?: string;
  lga_code?: string;
  district?: string;
  area?: string;
  unit?: string;
  error?: string;
  actionable_tip?: string;
  clean_length?: number;
}

export interface SegmentDiagnosis {
  segment: 'Length' | 'State' | 'LGA' | 'District' | 'Area' | 'BuildingUnit';
  input: string;
  expected: string;
  message: string;
  suggestions?: string[];
}

export interface DiagnosticReport {
  input: string;
  valid: boolean;
  normalized?: string;
  clean_length: number;
  format_score: number;
  diagnoses?: SegmentDiagnosis[];
  actionable_tip?: string;
}

export interface LocationResult {
  postcode: string;
  compact: string;
  latitude: number;
  longitude: number;
  address?: string;
  state_code: string;
  state_name: string;
  lga_code?: string;
  lga_name?: string;
  zone?: string;
  precision: 'building' | 'area' | 'district' | 'lga' | 'state' | number;
  google_maps_url: string;
  apple_maps_url: string;
  osm_url: string;
  error?: string;
}

export interface StateRecord {
  code: string;
  name: string;
  capital: string;
  latitude: number;
  longitude: number;
  zone: string;
}

export interface PostcodeEngine {
  validate(code: string): ValidationResult;
  diagnose(code: string): DiagnosticReport;
  parse(code: string): any;
  resolveLocation(code: string): LocationResult;
  listStates(): Record<string, StateRecord>;
}

/**
 * Initializes and instantiates the Postcode Nigeria WebAssembly module.
 * @param wasmSource Optional custom path, buffer, or fetch Response.
 */
export function initPostcode(
  wasmSource?: string | Buffer | ArrayBuffer | Response
): Promise<PostcodeEngine>;
