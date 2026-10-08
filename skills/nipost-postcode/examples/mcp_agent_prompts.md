# MCP Agent Interaction Prompts & Patterns

Here are reference interaction patterns between users and autonomous agents connected to the `postcode` Model Context Protocol server.

---

## Pattern 1: Cost-Aware Postal Verification

**User Prompt**:
> "Can you verify if postcode `EK-01-A03-FK-01` is real and show me where it is located?"

**Agent Reasoning & Tool Calls**:
1. *Tool 1: `validate_postcode`*
   ```json
   {
     "code": "EK-01-A03-FK-01"
   }
   ```
   *Result*: `{"valid": true, "postcode": "EK-01-A03-FK-01", "state": "EK"}`
2. *Tool 2: `lookup_postcode` (Cost-aware: Level 1 free check)*
   ```json
   {
     "code": "EK-01-A03-FK-01",
     "level": 1
   }
   ```
   *Result*: `{"valid": true, "postcode": "EK-01-A03-FK-01"}`
3. *Tool 3: `resolve_location` (Offline centroid, zero cost)*
   ```json
   {
     "code": "EK-01-A03-FK-01"
   }
   ```
   *Result*: `{"latitude": 7.6211, "longitude": 5.2214, "state_name": "Ekiti", "google_maps_url": "..."}`

**Agent Output**:
> "Postcode `EK-01-A03-FK-01` is officially valid. It corresponds to Ekiti State (Ado-Ekiti). Centroid coordinates are 7.621100°N, 5.221400°E. [View on Google Maps](...)"

---

## Pattern 2: Self-Healing Diagnostic Typo Recovery

**User Prompt**:
> "Geocode this address code for me: `ZZ-00-A03-FK-01`"

**Agent Reasoning & Tool Calls**:
1. *Tool 1: `validate_postcode`*
   *Result*: `{"valid": false, "error": "unknown state code ZZ"}`
2. *Tool 2: `diagnose_postcode`*
   ```json
   {
     "code": "ZZ-00-A03-FK-01"
   }
   ```
   *Result*:
   ```json
   {
     "format_score": 50.0,
     "actionable_tip": "Fix State: consider ZA (Zamfara); Fix LGA: 00 invalid (01-99)",
     "diagnoses": [
       {"segment": "state", "suggestions": ["ZA"]},
       {"segment": "lga", "message": "LGA cannot be 00"}
     ]
   }
   ```
3. *Self-Correction: Agent infers intended code was likely Zamfara state (`ZA-01-A03-FK-01`).*
4. *Tool 3: `validate_postcode` with repaired code `ZA-01-A03-FK-01`*
   *Result*: `{"valid": true}`
5. *Tool 4: `resolve_location` on `ZA-01-A03-FK-01`*

**Agent Output**:
> "The code `ZZ-00-A03-FK-01` was invalid (unknown state 'ZZ' and illegal LGA '00'). I diagnosed this and repaired it to `ZA-01-A03-FK-01` (Zamfara State). Its coordinates are..."

---

## Pattern 3: Informal Address Disambiguation

**User Prompt**:
> "What is the digital postcode for our office opposite Eko Hotel, Adetokunbo Ademola Street, Victoria Island, Lagos?"

**Agent Reasoning & Tool Calls**:
1. *Tool 1: `list_states`* $\to$ identifies Lagos as `LA`.
2. *Tool 2: `list_lgas` with `state: "LA"`* $\to$ identifies Victoria Island falls under Eti-Osa LGA (code `08` or `11`).
3. *Tool 3: Coordinates for Eko Hotel (approx 6.4281, 3.4219).*
4. *Tool 4: `reverse_geocode`*
   ```json
   {
     "latitude": 6.4281,
     "longitude": 3.4219,
     "max_distance_m": 150
   }
   ```
   *Result*: Snapped unit `LA-08-A86-RG-01`.

**Agent Output**:
> "Based on cadastral mapping for Eko Hotel on Adetokunbo Ademola Street (Victoria Island, Eti-Osa LGA, Lagos), the verified digital postcode is **`LA-08-A86-RG-01`**."
