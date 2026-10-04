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

export interface AdministrativeAddress {
  state?: string;
  state_name?: string;
  lga?: string;
  lga_name?: string;
  district?: string;
  district_name?: string;
  area?: string;
  area_name?: string;
  unit?: string;
  zone?: string;
  locality_name?: string;
}

export interface RecentHouseAddress {
  address?: string;
  recent?: string;
  house_number?: string;
  street_name?: string;
  locality_name?: string;
}

export interface PointGeometry {
  type?: string;
  coordinates?: [number, number];
}

export interface LookupResponse {
  postcode: string;
  valid: boolean;
  administrative_address?: AdministrativeAddress;
  recent_house_address?: RecentHouseAddress;
  building_use_status?: string;
  other_building_info?: Record<string, any>;
  point_geometry?: PointGeometry;
}

export interface PostcodeEngineOptions {
  apiKey?: string;
  baseURL?: string;
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
  setAPIKey(key: string): boolean;
  getAPIKey(): string;
  configure(options: PostcodeEngineOptions): boolean;
  lookup(code: string, level?: number): Promise<LookupResponse>;
}

/**
 * Initializes and instantiates the Postcode Nigeria WebAssembly module.
 * @param wasmSourceOrOptions Optional custom path, buffer, fetch Response, or options object.
 * @param options Optional configuration options when wasmSource is specified.
 */
export function initPostcode(
  wasmSourceOrOptions?: string | Buffer | ArrayBuffer | Response | PostcodeEngineOptions,
  options?: PostcodeEngineOptions
): Promise<PostcodeEngine>;
