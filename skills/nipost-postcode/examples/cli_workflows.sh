#!/usr/bin/env bash
# Nigerian Postcode CLI Workflow Recipes
#
# Prerequisite: the postcode CLI is already installed by the user, pinned to a
# reviewed release, e.g.:
#   go install github.com/abcubed3/postcode/cmd/postcode@v0.2.6
# This script does not install anything. It only calls the local `postcode`
# binary, and the only files it writes go to a private temp directory
# (override with OUT_DIR=...).

set -euo pipefail

command -v postcode >/dev/null 2>&1 || {
  echo "postcode CLI not found on PATH; install a pinned release first." >&2
  exit 1
}

OUT_DIR="${OUT_DIR:-$(mktemp -d)}"
echo "Writing generated artifacts to: ${OUT_DIR}"

echo "=== 1. Offline Validation & Formatting ==="
# Fast local validation (exit code 0 if valid, non-zero if invalid)
postcode validate "LA-11-W06-TC-10"

# Convert compact or spaced code to canonical hyphenated format
postcode format "LA 11 W06 TC 10" --style canonical

# Parse into JSON for downstream shell processing
postcode parse "EK01A03FK01" --format json

echo "=== 2. Offline Diagnostics & Typo Recovery ==="
# Diagnose malformed postcode with actionable suggestion tips.
# Exits non-zero for invalid input, so don't abort the script here.
postcode diagnose "ZZ-00-A03-FK-00" --format json || true

echo "=== 3. Offline Centroid Coordinates & Maps ==="
# Get latitude/longitude without any network calls
postcode coords "LA-11-W06-TC-10" --format json

# Generate Google Maps directions URL
postcode map "LA-11-W06-TC-10" --directions

echo "=== 4. Reverse Geocoding & Proximity Search ==="
# Snap coordinates in Victoria Island, Lagos to nearest postcode unit
postcode reverse --lat 6.4281 --lng 3.4219 --format json

# Discover active delivery points within a 200m radius
postcode nearby --code "LA-08-A86-RG-01" --radius 200 --format json

echo "=== 5. NIPOST Gateway Live Lookup (Cost-Aware) ==="
# Level 1: Free / Public validity check (0 credits).
# Reads POSTCODE_API_KEY from the environment; never pass keys on the command line.
postcode lookup "EK-01-A03-FK-01" --level 1 --format json

# Level 2: Commercial address retrieval (1 credit) - use only when street name needed
# postcode lookup "LA-11-W06-TC-10" --level 2 --format json

echo "=== 6. High-Throughput Batch Processing ==="
# Clean and enrich CSV records using concurrent workers
# postcode batch raw_customers.csv --column postcode --out "${OUT_DIR}/verified_customers.csv" --workers 8

echo "=== 7. Starting the MCP Server ==="
# Launch stdio MCP server for Claude Desktop / Cursor / Antigravity.
# Export POSTCODE_API_KEY from your secret manager beforehand rather than
# typing it inline (inline keys end up in shell history).
# postcode mcp

echo "=== 8. Shareable & Printable QR Codes ==="
# Terminal preview (no files written)
postcode qr "LA-11-W06-TC-10" --payload directions
# Printable NIPOST delivery badge SVG, written only inside OUT_DIR
postcode qr "EK-01-A03-FK-01" --label --out "${OUT_DIR}/ekiti-label.svg"
