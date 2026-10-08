#!/usr/bin/env bash
# Nigerian Postcode CLI Workflow Recipes
# Prerequisite: postcode CLI installed (`go install github.com/abcubed3/postcode/cmd/postcode@latest`)

set -euo pipefail

echo "=== 1. Offline Validation & Formatting ==="
# Fast local validation (exit code 0 if valid, non-zero if invalid)
postcode validate "LA-11-W06-TC-10"

# Convert compact or spaced code to canonical hyphenated format
postcode format "LA 11 W06 TC 10" --canonical

# Parse into JSON for downstream shell processing
postcode parse "EK01A03FK01" --format json

echo "=== 2. Offline Diagnostics & Typo Recovery ==="
# Diagnose malformed postcode with actionable suggestion tips
postcode diagnose "ZZ-00-A03-FK-00" --format json

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
# Level 1: Free / Public validity check (0 credits)
postcode lookup "EK-01-A03-FK-01" --level 1 --format json

# Level 2: Commercial address retrieval (1 credit) - use only when street name needed
# postcode lookup "LA-11-W06-TC-10" --level 2 --format json

echo "=== 6. High-Throughput Batch Processing ==="
# Clean and enrich 50,000+ CSV records using multi-threaded worker pools
# postcode batch raw_customers.csv --column postcode --out verified_customers.csv --workers 8

echo "=== 7. Starting the MCP Server ==="
# Launch stdio MCP server for Claude Desktop / Cursor / Antigravity
# POSTCODE_API_KEY="your-key" postcode mcp
