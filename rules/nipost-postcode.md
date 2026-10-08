---
trigger: always_on
description: Core principles, 11-character grammar rules, and cost-aware protocols for Nigerian Postcodes.
---

# Nigerian Postcode Agent Guidelines

When operating on Nigerian addresses, geocoding tasks, or postal codes, agents must follow these domain invariants:

## 1. 11-Character Alphanumeric Grammar
- Standard Structure: `SS-LL-DDD-AA-UU` (State, LGA, District, Area, Unit).
- Valid State codes: 36 states + `FC` (FCT Abuja).
- LGA (`LL`) and Building Unit (`UU`) must be decimal `01`-`99`. **`00` is illegal**.
- Single-digit LGA or Unit values must be zero-padded (`1` $\to$ `01`).
- Legacy 6-digit numeric postcodes are deprecated.

## 2. Cost-Aware Escalation Protocol
- NIPOST Gateway queries (`GET /v1/lookup`) are graded.
- Always validate offline first (`postcode.Parse` or `validate_postcode`).
- Default to **Level 1 (Free / Public validity check)**.
- Only escalate to **Level 2 / 3 (Commercial)** when physical street names or cadastral building use are explicitly demanded.
- For coordinate queries, resolve local administrative centroids offline (`resolve_location` / `postcode.ResolveLocation`) at $0 cost.

## 3. Self-Healing Diagnostics
- When encountering invalid postcodes, consult `postcode.Diagnose()` or `diagnose_postcode`.
- Read `actionable_tip` and `format_score` to self-heal typos before raising errors to users.

## 4. Zero-Allocation SDK Performance
- Go SDK `postcode.Parse()` and formatting functions maintain 0 heap allocations (`0 B/op`).
