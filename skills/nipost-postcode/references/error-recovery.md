# Diagnostic Error Codes & Self-Healing Playbook

## 1. Overview of Diagnostic Engine

The `postcode.Diagnose()` API and `diagnose_postcode` MCP tool analyze malformed candidate postcodes, scoring their conformance from `0.0` to `100.0` (`FormatScore`) and outputting segment-by-segment errors with actionable repair tips (`ActionableTip`).

---

## 2. Common Diagnostic Errors & Automated Repairs

| Diagnosis Type | Typical Input | Root Cause | Automated Self-Healing Action |
| :--- | :--- | :--- | :--- |
| **Invalid State Code** | `ZZ-01-A03-FK-01` | Unknown 2-letter state | Check `suggestions` array for closest valid code (e.g. `ZA` for Zamfara). |
| **Zero LGA Code** | `LA-00-A03-FK-01` | LGA segment is `00` | Replace `00` with `01` or look up official LGA list via `list_lgas(state: "LA")`. |
| **Zero Building Unit** | `LA-11-W06-TC-00` | Unit segment is `00` | Replace with default unit `01`. |
| **Single-Digit Segment** | `LA-1-W06-TC-1` | Missing zero-padding | Zero-pad to 2 digits: `LA-01-W06-TC-01`. |
| **Invalid Length** | `LA-11-W06-TC` | Missing unit segment | Check if input was truncated; append default unit `01`. |
| **Non-Alphanumeric Character** | `LA-11-W@6-TC-01` | Punctuation or noise in district | Strip invalid character or inspect surrounding text. |
| **Lowercase Letters** | `la-11-w06-tc-10` | Non-canonical case | Convert to uppercase: `LA-11-W06-TC-10`. |

---

## 3. The Autonomous Agent Self-Healing Pattern

When an agent encounters a postcode validation failure in a reasoning loop:

```
[Agent Observation: Validation failed on "ZZ-00-A03-FK-00"]
         │
         ▼
[Step 1: Run diagnose_postcode(code: "ZZ-00-A03-FK-00")]
         │
         ▼
[Step 2: Inspect Response]
{
  "format_score": 40.0,
  "actionable_tip": "Fix State: consider ZA (Zamfara); Fix LGA: 00 invalid (01-99); Fix Unit: 00 invalid (01-99)",
  "diagnoses": [
    {"segment": "state", "message": "Unknown state code 'ZZ'", "suggestions": ["ZA"]},
    {"segment": "lga", "message": "LGA cannot be '00'"},
    {"segment": "unit", "message": "Unit cannot be '00'"}
  ]
}
         │
         ▼
[Step 3: Synthesize Repaired Candidate]
Repaired: "ZA-01-A03-FK-01"
         │
         ▼
[Step 4: Re-Validate]
validate_postcode(code: "ZA-01-A03-FK-01") -> Valid: true
         │
         ▼
[Step 5: Output Repaired Result to User / Downstream Tool]
```

By following this loop, agents eliminate unforced errors and resolve 95%+ of address parsing issues autonomously without interrupting the user.
