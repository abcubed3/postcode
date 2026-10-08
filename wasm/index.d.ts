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

export interface BuildingRecord {
  postcode: string;
  latitude: number;
  longitude: number;
  address?: string;
  state_code?: string;
  state_name?: string;
  lga_code?: string;
  lga_name?: string;
  zone?: string;
}

export interface Segments {
  valid?: boolean;
  error?: string;
  state: string;
  lga: string;
  district: string;
  area: string;
  unit: string;
}

export interface AssembledPostcode {
  valid: boolean;
  error?: string;
  postcode?: string;
  display?: string;
  compact?: string;
}

export interface ParsedPostcode {
  valid: boolean;
  input?: string;
  postcode?: string;
  formatted?: string;
  compact?: string;
  spaced?: string;
  state_code?: string;
  state_name?: string;
  lga_code?: string;
  lga_name?: string;
  district?: string;
  area?: string;
  unit?: string;
  zone?: string;
  state_capital?: string;
  error?: string;
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
  state?: string;
  state_code: string;
  state_name: string;
  lga?: string;
  lga_code?: string;
  lga_name?: string;
  zone?: string;
  precision: 'building' | 'area' | 'district' | 'lga' | 'state' | number;
  google_maps_url: string;
  google_maps_directions_url?: string;
  apple_maps_url: string;
  osm_url: string;
  search_query?: string;
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

export interface AutocompleteSuggestion {
  postcode: string;
  description: string;
  score?: number;
}

export interface AutocompleteResponse {
  query: string;
  suggestions: AutocompleteSuggestion[];
}

export interface NearbyUnit {
  postcode: string;
  display: string;
  distance_m: number;
  confidence: string;
  state_name?: string;
  lga_name?: string;
  address?: string;
}

export interface NearbyResponse {
  results: NearbyUnit[];
}

export interface NearbyOptions {
  postcode?: string;
  code?: string;
  latitude?: number;
  longitude?: number;
  radius_m?: number;
  radiusKm?: number;
  radius_km?: number;
  radius?: number;
  limit?: number;
}

export interface ReverseResponse {
  found: boolean;
  coordinate?: [number, number];
  unit?: NearbyUnit;
  area?: string;
  district?: string;
  state?: string;
  message?: string;
  radius_m?: number;
  latitude?: number;
  longitude?: number;
  postcode?: string;
}

export interface ReverseOptions {
  latitude: number;
  longitude: number;
  max_distance_m?: number;
  maxDistanceKm?: number;
  maxDistanceM?: number;
  max_dist?: number;
}

export interface AgentGuardMetrics {
  total_calls: number;
  commercial_calls: number;
  downgraded_calls: number;
  offline_fallbacks: number;
}

export interface SyntheticAddress {
  raw_text: string;
  expected_state: string;
  expected_state_code: string;
  expected_lga?: string;
  expected_postcode?: string;
  expected_lat?: number;
  expected_lng?: number;
  noise_type?: string;
}

export interface GeneratorOptions {
  count?: number;
  noiseRate?: number;
  seed?: number;
}

export interface NamedCode {
  code: string;
  name?: string;
}

export interface PostcodeEngineOptions {
  apiKey?: string;
  baseURL?: string;
  googleMapsApiKey?: string;
  googleMapsKey?: string;
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
  // --- 1. Core Offline (Sync) ---
  validate(code: string): ValidationResult;
  validateBatch(codes: string[]): ValidationResult[];
  diagnose(code: string): DiagnosticReport;
  parse(code: string): ParsedPostcode;
  format(code: string, style?: 'canonical' | 'compact' | 'spaced' | 'hyphenated' | 'slug'): string;
  assemble(segments: { state: string; lga: string; district: string; area: string; unit: string }): AssembledPostcode;
  normalizeSegments(segments: { state: string; lga: string; district: string; area: string; unit: string }): Segments;
  disassemble(code: string): Segments;
  resolveLocation(code: string): LocationResult;
  registerBuilding(record: BuildingRecord): boolean;
  registerBuildings(records: BuildingRecord[]): number;
  listStates(): Record<string, StateRecord>;
  referenceStatesOffline(): NamedCode[];
  referenceLGAsOffline(state: string): NamedCode[];
  stateLGAs(state: string): NamedCode[];
  searchNearbyBuildingsOffline(latitude: number, longitude: number, radiusM?: number): NearbyUnit[];
  reverseCoordinatesOffline(latitude: number, longitude: number, maxDistanceM?: number): ReverseResponse;

  // --- 2. Configuration & State (Sync) ---
  setAPIKey(key: string): boolean;
  getAPIKey(): string;
  setGoogleMapsAPIKey(key: string): boolean;
  getGoogleMapsAPIKey(): string;
  configure(options: PostcodeEngineOptions): boolean;

  // --- 3. Live Gateway Operations & Catalogs (Async) ---
  resolveLocationOnline(code: string): Promise<LocationResult>;
  lookup(code: string, level?: number): Promise<LookupResponse>;
  autocomplete(query: string): Promise<AutocompleteResponse>;
  nearby(postcodeOrParams: string | NearbyOptions, radiusM?: number): Promise<NearbyResponse>;
  nearby(latitude: number, longitude: number, radiusM?: number): Promise<NearbyResponse>;
  reverseGeocode(latOrParams: number | ReverseOptions, lng?: number, maxDistanceM?: number): Promise<ReverseResponse>;
  assembleOnline(segments: { state: string; lga: string; district: string; area: string; unit: string }): Promise<AssembledPostcode>;
  disassembleOnline(code: string): Promise<Segments>;
  referenceStates(online?: boolean): Promise<NamedCode[]>;
  referenceLGAs(state: string): Promise<NamedCode[]>;
  referenceDistricts(state: string, lga: string): Promise<NamedCode[]>;
  referenceAreas(state: string, lga: string, district: string): Promise<NamedCode[]>;
  health(): Promise<{ status: string }>;

  // --- 4. AI Agent Tooling & Guardrails ---
  getAgentTools(format?: 'openai' | 'anthropic' | 'gemini'): any[];
  executeTool(name: string, args?: Record<string, any> | string): Promise<any>;
  getAgentMetrics(): AgentGuardMetrics;

  // --- 5. Synthetic Evaluation Benchmark ---
  generateSyntheticAddresses(options?: GeneratorOptions): SyntheticAddress[];
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
