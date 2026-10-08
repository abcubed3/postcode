# NIPOST Lookup Levels & Credit Accounting

## 1. Overview of Graded Response Model

The official NIPOST Postcode API (`GET /v1/lookup`) structures postcode metadata into progressive tiers. Higher tiers return more enriched cadastral and structural property data, but require paid commercial credits.

```text
 Level 1: Public / Free       ───> Postcode existence & validity boolean
 Level 2: Commercial Tier A   ───> Administrative address + street name
 Level 3: Commercial Tier B   ───> Building use (residential, commercial, mixed)
 Level 4: Enterprise Cadastre ───> Structural metadata, floor count, unit count
 Level 5: Spatial Geometry    ───> Full parcel polygon boundaries & cadastral coordinates
```

---

## 2. Detailed Breakdown of Tiers

### Level 1: Free / Public Tier
- **Cost**: 0 credits (Free).
- **Rate Limit**: Typically 60-120 requests/minute on standard API keys.
- **Fields Returned**:
  - `valid`: `true` or `false`.
  - `postcode`: Canonical formatted string (`LA-11-W06-TC-10`).
- **Use When**:
  - Verifying if a postcode actually exists in the national database.
  - Form validation on checkout pages or user profile forms.
  - General presence checks where physical street name is already known.

### Level 2: Commercial Address Tier
- **Cost**: 1 Commercial Credit per successful lookup.
- **Fields Returned**:
  - All Level 1 fields.
  - `state`: Name of State (e.g., "Lagos").
  - `lga`: Local Government Area (e.g., "Eti-Osa").
  - `district`: Postal district name.
  - `area`: Neighborhood or ward name.
  - `street`: Street or road name (e.g., "Admiralty Way").
- **Use When**:
  - Address autofill (user enters postcode, system fills street, city, state).
  - Dispatching logistics drivers to a validated physical street.

### Level 3: Commercial Building Classification Tier
- **Cost**: 2 Commercial Credits per successful lookup.
- **Fields Returned**:
  - All Level 2 fields.
  - `building_use`: Classification enum (`residential`, `commercial`, `industrial`, `mixed`, `institutional`).
  - `building_name`: Recognized building or complex name (if cadastred).
- **Use When**:
  - Commercial underwriting, insurance risk scoring, or KYC property verification.
  - Distinguishing business delivery addresses from private residential units.

### Level 4 & 5: Advanced Cadastral Tiers
- **Cost**: Specialized enterprise quotas.
- **Fields Returned**:
  - Building footprint, number of floors, occupancy estimates, geospatial parcel GeoJSON.
- **Note**: Standard commercial API keys are strictly capped at Level 3. Requests specifying Level 4 or 5 without enterprise permissions will fail with an authorization error.

---

## 3. Quota Management & Rate Limit Headers

The NIPOST Gateway transmits quota telemetry in response HTTP headers:

| Header | Meaning |
| :--- | :--- |
| `X-RateLimit-Limit` | Max requests allowed in current window. |
| `X-RateLimit-Remaining` | Remaining requests in current window. |
| `X-RateLimit-Reset` | Unix timestamp when the rate limit window resets. |
| `X-Credit-Balance` | Remaining commercial credit balance for the API key. |
| `Retry-After` | Seconds to wait before retrying after a `429 Too Many Requests`. |

---

## 4. Agent Best Practices for Quota Preservation

1. **Local Centroid First**: If an agent only needs coordinates for distance calculation or Google Maps directions, use `postcode.ResolveLocation()` or CLI `postcode coords`. It resolves the centroid locally in ~115ns at **$0 cost**.
2. **Never Query Without Validation**: Always validate the 11-character grammar locally before sending a request to the API. Never spend API credits on malformed strings.
3. **Budget Guardrails**: In long-running autonomous agent loops, always configure `AgentGuardConfig` to prevent recursive retries from depleting API balances.
