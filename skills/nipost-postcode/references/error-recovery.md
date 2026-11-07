# Diagnostic Error Codes & Self-Healing Playbook

## 1. Overview of Diagnostic Engine

The `postcode.Diagnose()` API and `diagnose_postcode` MCP tool analyze malformed candidate postcodes, scoring their conformance from `0.0` to `100.0` (`FormatScore`) and outputting segment-by-segment errors with actionable repair tips (`ActionableTip`).

---

## 2. Common Diagnostic Errors & Repair Policy

Repairs fall into two classes:

- **Mechanical** (safe to apply automatically): the corrected code names the *same* location — case, separators, zero-padding, O/0 and I/1 confusion in numeric segments.
- **Semantic** (needs user confirmation): the corrected code could name a *different* location — a changed State, a guessed LGA or Unit, or a filled-in missing segment. Present the suggestion and ask; never substitute silently.

| Diagnosis Type | Typical Input | Root Cause | Repair Class | Action |
| :--- | :--- | :--- | :--- | :--- |
| **Lowercase Letters** | `la-11-w06-tc-10` | Non-canonical case | Mechanical | Convert to uppercase: `LA-11-W06-TC-10`. |
| **Single-Digit Segment** | `LA-1-W06-TC-1` | Missing zero-padding | Mechanical | Zero-pad to 2 digits: `LA-01-W06-TC-01`. |
| **Non-Alphanumeric Character** | `LA-11-W@6-TC-01` | Punctuation or noise in district | Semantic | Suggest likely character; confirm with user or inspect surrounding text. |
| **Invalid State Code** | `ZZ-01-A03-FK-01` | Unknown 2-letter state | Semantic | Offer `suggestions` (e.g. `ZA` for Zamfara) and confirm. |
| **Zero LGA Code** | `LA-00-A03-FK-01` | LGA segment is `00` | Semantic | Look up the official LGA list via `list_lgas(state: "LA")` and ask which LGA is meant. |
| **Zero Building Unit** | `LA-11-W06-TC-00` | Unit segment is `00` | Semantic | Ask for the unit number; only fall back to `01` if the user accepts an area-level result. |
| **Invalid Length** | `LA-11-W06-TC` | Missing unit segment | Semantic | Check if input was truncated; ask for the missing segment. |

---

## 3. The Agent Self-Healing Pattern

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
[Step 3: Classify Repairs]
All three are semantic (State, LGA, Unit change the location).
         │
         ▼
[Step 4: Confirm with User]
"ZZ-00-A03-FK-00 is invalid. Did you mean Zamfara (ZA)? Which LGA and unit?"
         │
         ▼
[Step 5: Re-Validate the Confirmed Code]
validate_postcode(code: "ZA-01-A03-FK-01") -> Valid: true
         │
         ▼
[Step 6: Report Result, Stating Original → Corrected]
```

By following this loop, agents fix formatting noise without bothering the user, and only ask when a correction would change which place the code refers to.
