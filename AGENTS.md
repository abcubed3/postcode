# Agent Guidelines: Nigerian Postcode SDK & Toolkit

This repository (`github.com/abcubed3/postcode`) is the reference implementation of Nigeria's 11-character National Digital Postcode system (NIPOST).

## Core Principles & Invariants

1. **11-Character Alphanumeric Grammar**:
   - Structure: `SS-LL-DDD-AA-UU` (State, LGA, District, Area, Unit).
   - Valid State codes: 36 states + `FC` (FCT Abuja).
   - LGA (`LL`) and Building Unit (`UU`) must be decimal `01`-`99`. **`00` is illegal**.
   - Phased out: Legacy 6-digit numeric postcodes are deprecated.

2. **Zero-Allocation Core Performance**:
   - `postcode.Parse()` and formatting functions in core Go SDK must maintain **0 heap allocations** (`0 B/op`).
   - Run `go test -bench=. -benchmem` to ensure zero escapes on parser benchmarks.

3. **Cost-Aware Escalation Protocol**:
   - NIPOST Gateway queries (`GET /v1/lookup`) are graded.
   - Always validate offline first (`postcode.Parse` or `validate_postcode`).
   - Default to **Level 1 (Free / Public validity)**.
   - Only escalate to **Level 2 / 3 (Commercial)** when street names or building use are explicitly demanded.

4. **Self-Healing Diagnostics**:
   - When encountering invalid postcodes, consult `postcode.Diagnose()` or `diagnose_postcode`.
   - Read `ActionableTip` and `FormatScore` to self-correct typos before raising errors.

5. **Skill Bundle**:
   - Refer to [Nigerian Postcode Skill](skills/nipost-postcode/SKILL.md) for full procedures, disambiguation SOPs, and references.
