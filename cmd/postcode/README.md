# Postcode Nigeria CLI (`postcode`) — User Guide & Reference

> **The unofficial, ultra-fast command-line interface for Nigeria's 11-character National Digital Postcode System.**

[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![NIPOST Compliant](https://img.shields.io/badge/NIPOST-Compliant-008751.svg)](https://postcode.gov.ng)

---

## Table of Contents

- [Overview & Architecture](#overview--architecture)
- [Installation](#installation)
  - [Download Pre-Built Binaries](#download-pre-built-binaries)
  - [Install via `go install`](#install-via-go-install)
  - [Build from Source](#build-from-source)
- [Quickstart: The 60-Second Cheat Sheet](#quickstart-the-60-second-cheat-sheet)
- [Configuration & Precedence](#configuration--precedence)
  - [Environment Variables](#environment-variables)
  - [YAML Configuration File](#yaml-configuration-file)
  - [Precedence Order](#precedence-order)
- [Output Formats (`text`, `json`, `yaml`, `csv`)](#output-formats-text-json-yaml-csv)
- [Command Deep Dives & Practical Examples](#command-deep-dives--practical-examples)
  - [1. Validation (`validate`)](#1-validation-validate)
  - [2. Parsing & Introspection (`parse`)](#2-parsing--introspection-parse)
  - [3. Formatting & Normalization (`format`)](#3-formatting--normalization-format)
  - [4. Coordinates & Geocoding (`coords`)](#4-coordinates--geocoding-coords)
  - [5. Mapping & Directions (`map`)](#5-mapping--directions-map)
  - [6. Assembly & Disassembly (`assemble` / `disassemble`)](#6-assembly--disassembly-assemble--disassemble)
  - [7. High-Throughput Batch Processing (`batch`)](#7-high-throughput-batch-processing-batch)
  - [8. Official NIPOST API Lookup (`lookup`)](#8-official-nipost-API-lookup-lookup)
  - [9. Reverse Geocoding & Coordinate Snapping (`reverse`)](#9-reverse-geocoding--coordinate-snapping-reverse)
  - [10. Radial Proximity Search (`nearby`)](#10-radial-proximity-search-nearby)
  - [11. Interactive Autocomplete (`autocomplete`)](#11-interactive-autocomplete-autocomplete)
  - [12. API Diagnostics & Rate Limits (`status`)](#12-API-diagnostics--rate-limits-status)
  - [13. Local Mock API Simulator (`serve`)](#13-local-mock-API-simulator-serve)
  - [14. AI Diagnostic Inspection & Self-Correction (`diagnose`)](#14-ai-diagnostic-inspection--self-correction-diagnose)
  - [15. Model Context Protocol Server (`mcp`)](#15-model-context-protocol-server-mcp)
  - [16. Synthetic Address Generator & AI Evaluation Harness (`eval`)](#16-synthetic-address-generator--ai-evaluation-harness-eval)
- [Production Recipes & Shell Integration](#production-recipes--shell-integration)
  - [Recipe A: Stream Processing with `jq` and `curl`](#recipe-a-stream-processing-with-jq-and-curl)
  - [Recipe B: Fast DB Data Cleaning in ETL Pipelines](#recipe-b-fast-db-data-cleaning-in-etl-pipelines)
- [Shell Auto-Completion (Bash, Zsh, Fish, PowerShell)](#shell-auto-completion-bash-zsh-fish-powershell)
- [Troubleshooting & FAQ](#troubleshooting--faq)

---

## Overview & Architecture

Nigeria's digital postal system, established by the Nigerian Postal Service (**NIPOST**), represents every physical address using a deterministic **11-character alphanumeric code** formatted as:

$$\mathbf{\underbrace{LA}_{\text{State}}-\underbrace{11}_{\text{LGA}}-\underbrace{W06}_{\text{District}}-\underbrace{TC}_{\text{Area}}-\underbrace{10}_{\text{Unit}}}$$

```text
 ┌─────────┬─────────┬──────────────┬──────────┬─────────────┐
 │  State  │   LGA   │   District   │   Area   │ Bldg / Unit │
 │  (2-char│ (2-digit│ (3-char code │ (2-char  │  (2-digit   │
 │  alpha) │ numeric)│ alpha+num)   │  alpha)  │  numeric)   │
 ├─────────┼─────────┼──────────────┼──────────┼─────────────┤
 │   LA    │   11    │     W06      │    TC    │     10      │
 └─────────┴─────────┴──────────────┴──────────┴─────────────┘
```

The `postcode` CLI operates in two modes:
1. **Offline Mode (Default, Zero Latency)**: Runs pure algorithmic validation, formatting, coordinate centroid lookup, format conversion, and CSV batch processing directly on your machine without making any network calls.
2. **Online Mode (Online)**: Integrates directly with the official NIPOST Postcode API (`api.postcode.gov.ng`) or a local mock simulator for real-time cadastral verification, street-level reverse geocoding, and multi-grade property lookup.

---

## Installation

### Download Pre-Built Binaries

Download compiled binaries for your OS and architecture from GitHub Releases:

- **Linux**: `postcode-linux-amd64`, `postcode-linux-arm64`
- **macOS (Apple Silicon & Intel)**: `postcode-darwin-arm64`, `postcode-darwin-amd64`
- **Windows**: `postcode-windows-amd64.exe`, `postcode-windows-arm64.exe`

```bash
# Example for macOS Apple Silicon
curl -sSLO https://github.com/abcubed3/postcode/releases/latest/download/postcode-darwin-arm64
chmod +x postcode-darwin-arm64
sudo mv postcode-darwin-arm64 /usr/local/bin/postcode
```

### Install via `go install`

If you have Go 1.27+ installed on your workstation:

```bash
go install github.com/abcubed3/postcode/cmd/postcode@latest
```

Verify your installation:

```bash
postcode version
```

### Build from Source

```bash
git clone https://github.com/abcubed3/postcode.git
cd postcode/cmd/postcode
go build -trimpath -ldflags="-s -w" -o postcode .
./postcode --help
```

---

## Quickstart: The 60-Second Cheat Sheet

```bash
# 1. Validate one or more postcodes
postcode validate EK-01-A03-FK-01 "LA 11 W06 TC 10"

# 2. Inspect all administrative components in JSON
postcode parse "FC 03 B06 AG 12" -o json

# 3. Normalize format to standard canonical dashes
postcode format "ek 01 a03 fk 01" --style canonical

# 4. Extract geographic coordinates and precision
postcode coords EK-01-A03-FK-01

# 5. Generate turn-by-turn navigation link
postcode map EK-01-A03-FK-01 --directions

# 6. Break down a postcode into raw segments
postcode disassemble LA-11-W06-TC-10

# 7. Assemble a postcode from its individual parts
postcode assemble --state EK --lga 01 --district A03 --area FK --unit 01

# 8. Enrich an entire CSV database with coordinates in seconds
postcode batch --input customers.csv --output enriched.csv --column shipping_postcode

# 9. Reverse geocode a GPS latitude/longitude to the nearest postcode unit
postcode reverse --lat 7.6211 --lng 5.2215

# 10. Start a local zero-latency mock NIPOST server for testing
postcode serve --port 2340
```

---

## Configuration & Precedence

The CLI seamlessly handles configuration through command-line flags, environment variables, or a YAML configuration file.

### Environment Variables

Every setting can be specified via environment variables prefixed with `POSTCODE_`:

| Environment Variable | CLI Flag | Description | Default |
| :--- | :--- | :--- | :--- |
| `POSTCODE_API_KEY` | `-k`, `--apikey` | Official NIPOST API API key | `""` |
| `POSTCODE_BASE_URL` | `-u`, `--url` | Base URL of NIPOST API | `https://api.postcode.gov.ng` |
| `POSTCODE_TIMEOUT` | `-t`, `--timeout` | HTTP timeout duration | `10s` |
| `POSTCODE_OUTPUT` | `-o`, `--output` | Default output format (`text`, `json`, `yaml`, `csv`) | `text` |
| `POSTCODE_CONFIG` | `--config` | Custom configuration file path | `$HOME/.postcode.yaml` |

```bash
# Set credentials in your current shell or CI runner
export POSTCODE_API_KEY="nipost_test_f2e9dabcdef1234567890"
export POSTCODE_OUTPUT="json"
```

### YAML Configuration File

Create a configuration file at `$HOME/.postcode.yaml` or `./.postcode.yaml`:

```yaml
# ~/.postcode.yaml
apikey: "nipost_test_f2e9ddef1234567890"
url: "https://api.postcode.gov.ng"
timeout: 15s
output: "text"
```

You can target a specific config file using the `--config` flag:

```bash
postcode lookup EK-01-A03-FK-01 --config /path/to/custom-config.yaml
```

### Precedence Order

When resolving configuration settings, `postcode` evaluates values in this strict order:
1. **Explicit CLI Flags** (highest priority)
2. **Environment Variables** (`POSTCODE_*`)
3. **YAML Config File** (`~/.postcode.yaml` or `./.postcode.yaml`)
4. **Built-in Defaults** (lowest priority)

---

## Output Formats (`text`, `json`, `yaml`, `csv`)

All queries support the global `-o` / `--output` flag:

### 1. Human-Readable Text (Default)
Optimized for terminal readability:
```bash
postcode parse EK-01-A03-FK-01
```
```text
Postcode:      EK-01-A03-FK-01
Compact:       EK01A03FK01
Spaced:        EK 01 A03 FK 01
State:         Ekiti (EK)
Capital:       Ado Ekiti
LGA:           Ado Ekiti (01)
District:      A03
Area:          FK
Building Unit: 01
Zone:          SOUTH WEST
```

### 2. JSON (`-o json`)
Structured for scripts, webhooks, and `jq`:
```bash
postcode parse EK-01-A03-FK-01 -o json
```
```json
{
  "total": 1,
  "results": [
    {
      "input": "EK-01-A03-FK-01",
      "formatted": "EK-01-A03-FK-01",
      "compact": "EK01A03FK01",
      "spaced": "EK 01 A03 FK 01",
      "state_code": "EK",
      "state_name": "Ekiti",
      "state_capital": "Ado Ekiti",
      "lga_code": "01",
      "lga_name": "Ado Ekiti",
      "district": "A03",
      "area": "FK",
      "unit": "01",
      "zone": "SOUTH WEST"
    }
  ]
}
```

### 3. YAML (`-o yaml`)
Clean serialization for humans as well:
```bash
postcode parse EK-01-A03-FK-01 -o yaml
```

### 4. CSV (`-o csv`)
Standard comma-separated format for spreadsheets:
```bash
postcode coords EK-01-A03-FK-01 "LA 11 W06 TC 10" -o csv
```
```csv
input,postcode,latitude,longitude,precision,address,state
EK-01-A03-FK-01,EK-01-A03-FK-01,7.621100,5.221500,building,"NTA Road, Back of Fabian Hotel, Ado Ekiti",Ekiti
LA 11 W06 TC 10,LA-11-W06-TC-10,6.601800,3.351500,district,"",Lagos
```

---

## Command Deep Dives & Practical Examples

### 1. Validation (`validate`)

Checks whether given codes strictly comply with Nigerian alphanumeric digital postcode structure and valid state codes.

```bash
# Validate multiple codes in one command
postcode validate EK-01-A03-FK-01 "LA 11 W06 TC 10" INVALID-CODE
```
```text
✓ EK-01-A03-FK-01  -> EK-01-A03-FK-01  (State: EK)
✓ LA 11 W06 TC 10  -> LA-11-W06-TC-10  (State: LA)
✗ INVALID-CODE     -> invalid postcode length: got 12, want 11 alphanumeric characters
```

#### Quiet Mode for CI/CD Pipelines
Use `-q` / `--quiet` to suppress output and return exit code `0` on success, or non-zero if any code is invalid:

```bash
if postcode validate -q "$USER_POSTCODE"; then
    echo "Postcode is valid!"
else
    echo "Invalid postcode detected!"
    exit 1
fi
```

#### Stdin Streaming
Validate codes piped directly from files or other CLI tools:
```bash
cat postcodes.txt | postcode validate -
```

---

### 2. Parsing & Introspection (`parse`)

Deconstructs postcodes into administrative divisions (State, Capital, LGA, District, Area, Unit, Geopolitical Zone).

```bash
postcode parse "FC 03 B06 AG 12"
```

Extract specific fields using `jq`:
```bash
postcode parse "FC 03 B06 AG 12" -o json | jq -r '.results[0].state_capital'
# Output: Abuja
```

---

### 3. Formatting & Normalization (`format`)

Normalizes messy user input into standardized formats.

Supported styles via `-s` / `--style`:
- `canonical`: `EK-01-A03-FK-01` (hyphenated standard)
- `spaced`: `EK 01 A03 FK 01` (human-readable print standard)
- `compact`: `EK01A03FK01` (database key standard)

```bash
# Convert spaced input to canonical format
postcode format "ek 01 a03 fk 01" --style canonical
# Output: EK-01-A03-FK-01

# Convert messy input to compact format
postcode format "  la-11-w06-tc-10  " --style compact
# Output: LA11W06TC10
```

---

### 4. Coordinates & Geocoding (`coords`)

Extracts GPS coordinates (Latitude & Longitude) and precision level (`building`, `district`, `lga`, `state`) using the offline centroid dataset.

```bash
postcode coords EK-01-A03-FK-01
```
```text
EK-01-A03-FK-01  -> Lat:   7.621100, Lng:   5.221500 [building] (NTA Road, Back of Fabian Hotel, Ado Ekiti - Ekiti)
```

#### Online API Enrichment
Pass `--online` to query live NIPOST boundary layers for coordinate resolution:
```bash
postcode coords EK-01-A03-FK-01 --online
```

---

### 5. Mapping & Directions (`map`)

Generates navigation and mapping URLs for coordinates derived from postcodes.

```bash
# Generate Google Maps Search URL (default)
postcode map EK-01-A03-FK-01

# Generate turn-by-turn directions link
postcode map EK-01-A03-FK-01 --directions

# Target Apple Maps or OpenStreetMap
postcode map EK-01-A03-FK-01 --provider apple
postcode map EK-01-A03-FK-01 --provider osm

# Automatically open the location in your default browser
postcode map EK-01-A03-FK-01 --open
```

---

### 6. Assembly & Disassembly (`assemble` / `disassemble`)

#### Assemble a Postcode
Construct a valid 11-character postcode from individual administrative segments:

```bash
postcode assemble \
  --state EK \
  --lga 01 \
  --district A03 \
  --area FK \
  --unit 01
```
```text
Postcode: EK-01-A03-FK-01
Display:  EK 01 A03 FK 01
Compact:  EK01A03FK01
```

#### Disassemble a Postcode
Extract every component into a structured map:
```bash
postcode disassemble EK-01-A03-FK-01
```
```text
EK-01-A03-FK-01  -> State: EK | LGA: 01 | District: A03 | Area: FK | Unit: 01
```

---

### 7. High-Throughput Batch Processing (`batch`)

Enrich CSV datasets containing thousands or millions of addresses with validity, canonical code, state, LGA, coordinates, precision, and Google Maps links.

- **Zero network requests** — runs purely in-memory
- **Processes over 150,000 rows/second** on standard multi-core laptops

```bash
postcode batch \
  --input customers.csv \
  --output customers_enriched.csv \
  --column postcode
```

#### Pipe Stdin to Stdout
Clean data seamlessly inside Unix pipes:
```bash
cat raw_orders.csv | postcode batch --input - --output - --column shipping_code > enriched_orders.csv
```

---

### 8. Official NIPOST API Lookup (`lookup`)

Performs live verification against the official NIPOST API (`api.postcode.gov.ng`).

Supports three levels of verification via `-l` / `--level`:
- **Level 1 (Default)**: Postal validation and administrative assignment.
- **Level 2**: Detailed address geocoding and street attributes.
- **Level 3**: Building use classification and parcel boundary coordinates.

```bash
# Basic lookup
postcode lookup EK-01-A03-FK-01 --level 1

# Comprehensive cadastral verification
postcode lookup EK-01-A03-FK-01 --level 3 -o json
```

> [!TIP]
> If network access is lost, `lookup` automatically falls back to local reference data by default. Disable this with `--offline-fallback=false`.

---

### 9. Reverse Geocoding & Coordinate Snapping (`reverse`)

Converts a GPS latitude and longitude coordinate into the closest official postal unit.

```bash
postcode reverse --lat 7.621100 --lng 5.221500 --max-dist 50
```
```text
Snapping (7.621100, 5.221500) -> EK-01-A03-FK-01
------------------------------------------------------------
Postcode:      EK-01-A03-FK-01
Distance:      0.0 meters
Confidence:    HIGH
State:         Ekiti
LGA:           Ado Ekiti
District:      A03
Area:          FK
Address:       NTA Road, Back of Fabian Hotel, Ado Ekiti
```

---

### 10. Radial Proximity Search (`nearby`)

Finds all registered postal units within a specific radial distance (up to 300 meters) from a GPS coordinate:

```bash
postcode nearby --lat 7.6211 --lng 5.2215 --radius 250
```

---

### 11. Interactive Autocomplete (`autocomplete`)

Provides suggestions and completions as users type a postcode:

```bash
postcode autocomplete "EK 01 A03"
```

---

### 12. API Diagnostics & Rate Limits (`status`)

Inspects NIPOST API health, connection latency, API key authentication status, and real-time rate limit quota remaining:

```bash
postcode status
```
```text
============================================================
API Endpoint: https://api.postcode.gov.ng
Status:           HEALTHY (42.15ms latency)
API Key:          CONFIGURED
Rate Quota:       948 / 1000 remaining
Reset At:         2026-10-04T00:00:00Z
============================================================
```

---

### 13. Local Mock API Simulator (`serve`)

Start a lightweight, local HTTP simulator of the NIPOST API for automated integration testing and offline development:

```bash
# Launch with embedded reference dataset (21 official test postcodes)
postcode serve --port 2340 --host localhost

# Or load a custom test dataset at runtime
postcode serve --port 2340 --data ./my_test_postcodes.json
```
```text
================================================================================
🏛️ NIPOST Digital Postcode API - Local Simulator
📖 Official Docs: https://docs.postcode.gov.ng/
================================================================================
🚀 LocalServer listening at: http://localhost:2340
🩺 Health endpoint:     http://localhost:2340/healthz
--------------------------------------------------------------------------------
📋 Loaded 21 official reference postcodes across 11 states
Press Ctrl+C to stop.
```

To point your local queries to the mock simulator:
```bash
postcode lookup EK-01-A03-FK-01 --api http://localhost:2340
```

---

### 14. AI Diagnostic Inspection & Self-Correction (`diagnose`)

Perform segment-by-segment grammar, structural, and administrative inspection of candidate postcodes. Pinpoints exact segment errors, calculates character counts, generates candidate suggestions for typos, and produces actionable tips for AI agent self-correction:

```bash
# Diagnose an invalid postcode with typos and prohibited 00 codes
postcode diagnose "ZZ 00 A03 FK 00"
```
```text
✗ ZZ 00 A03 FK 00  -> INVALID (11/11 chars)
    [State] State code "ZZ" is not a recognized Nigerian state code
      Suggestions: ZA (Zamfara), SO (Sokoto)
    [LGA] LGA code '00' is invalid; valid LGA codes range from 01 to 99
      Suggestions: 01, 02
    [BuildingUnit] Building unit code '00' is invalid; valid unit codes range from 01 to 99
      Suggestions: 01, 02
    Action: Fix State: consider ZA (Zamfara), SO (Sokoto); Fix LGA: consider 01, 02; Fix BuildingUnit: consider 01, 02
```

Export detailed JSON for automated agent error handling:
```bash
echo "LA 11 W06" | postcode diagnose -o json
```

---

### 15. Model Context Protocol Server (`mcp`)

Run an in-process Model Context Protocol (MCP) server over standard input/output (`stdio`), exposing postcode validation, diagnostics, geocoding, reverse search, and gazetteer resources directly to **Claude Desktop**, **Cursor**, and **Antigravity IDE**:

```bash
postcode mcp
```

#### Claude Desktop Setup
Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "postcode": {
      "command": "postcode",
      "args": ["mcp"],
      "env": {
        "POSTCODE_API_KEY": "YOUR_NIPOST_API_KEY"
      }
    }
  }
}
```

#### Available Agent Tools via MCP
- `validate_postcode`: Validates 11-character grammar, format, and state code.
- `diagnose_postcode`: In-depth segment-by-segment analysis with suggestions.
- `resolve_location`: Resolves coordinates, administrative metadata, and Google Maps URL.
- `reverse_geocode`: Snaps latitude/longitude to the nearest postcode unit.
- `search_nearby`: Searches for active units within a radius (default 300m).
- `autocomplete_postcode`: Segment-aware autocomplete suggestions.
- `lookup_postcode`: Graded official gateway query (L1 free, L2/L3 commercial).
- `list_states`: Returns directory of all 36 States + FCT with capitals and centroids.

---

### 16. Synthetic Address Generator & AI Evaluation Harness (`eval`)

Benchmark AI agents, LLM entity extractors, and OCR pipelines on realistic Nigerian address datasets. Generates synthetic noisy addresses containing Nigerian landmarks, colloquial descriptions, informal abbreviations, and state typos:

```bash
# Generate 100 synthetic addresses with 30% noise and benchmark extraction accuracy
postcode eval --samples 100 --noise 0.3
```

```text
Nigerian Address AI Evaluation Scorecard
========================================
Total Samples:       100
Noise Rate:          30.0%
State Accuracy:      98.00%
Postcode Accuracy:   89.00%
Valid Postcode Rate: 89.00%
Avg Format Score:    93.50 / 100
Eval Latency:        2 ms
```

#### Export Datasets for AI Eval Frameworks
Export benchmark datasets to JSON for use with LangSmith, Promptfoo, Braintrust, or DeepEval:
```bash
postcode eval --samples 500 --noise 0.4 --export benchmarks/nigerian_addresses.json
```

---

## Production Recipes & Shell Integration

### Recipe A: Stream Processing with `jq` and `curl`

Fetch coordinates from your customer service API and immediately snap them to digital postcodes:

```bash
curl -s https://api.example.com/driver/location \
  | jq -r '"--lat \(.latitude) --lng \(.longitude)"' \
  | xargs postcode reverse -o json \
  | jq '.unit.postcode'
```

### Recipe B: Fast DB Data Cleaning in ETL Pipelines

Clean and enrich an export of 500,000+ customer records from PostgreSQL:

```bash
psql -d production -c "COPY (SELECT id, raw_postcode FROM customers) TO STDOUT WITH CSV HEADER" \
  | postcode batch -i - -o - -c raw_postcode \
  | psql -d analytics -c "COPY enriched_customers FROM STDIN WITH CSV HEADER"
```

---

## Shell Auto-Completion (Bash, Zsh, Fish, PowerShell)

Generate auto-completion scripts for your shell:

### Bash
```bash
postcode completion bash > ~/.local/share/bash-completion/completions/postcode
source ~/.local/share/bash-completion/completions/postcode
```

### Zsh
```bash
postcode completion zsh > "${fpath[1]}/_postcode"
# Add to ~/.zshrc if not already present:
# autoload -U compinit && compinit
```

### Fish
```bash
postcode completion fish > ~/.config/fish/completions/postcode.fish
```

### PowerShell (Windows)
```powershell
postcode completion powershell | Out-String | Invoke-Expression
```

---

## Troubleshooting & FAQ

### Q: Why do I get `postcode column not found in CSV header` during batch processing?
**A:** Ensure the `--column` flag matches your CSV header exactly (case-insensitive). For example, if your CSV header is `Delivery_Postcode`, use `--column Delivery_Postcode`.

### Q: What is the difference between offline and online geocoding?
**A:** Offline coordinates use pre-computed centroid coordinates embedded in the library binary. Online geocoding (`--online` or `lookup`) contacts the official NIPOST registry to fetch cadastral coordinates and exact building boundary geometry.

### Q: Can I use this CLI in an air-gapped or offline environment?
**A:** Yes! All offline commands (`validate`, `parse`, `format`, `coords`, `map`, `assemble`, `disassemble`, and `batch`) execute with **zero network dependencies** and will never attempt to dial an external server.

---

*Authored by [@abcubed3](https://abcubed3.dev) with ❤️ and a few tokens for the Nigerian Open Source Community & Digital Public Infrastructure.*
