# Postcode Nigeria CLI (`postcode`)

> Comprehensive command-line toolkit for parsing, validating, formatting, geocoding, and querying Nigeria's 11-character digital postcodes.

Built for high-performance offline address processing and full integration with the official NIPOST Postcode API (`api.postcode.gov.ng`).

## Quick Start

```bash
# Install
go install github.com/abcubed3/postcode/cmd/postcode@latest

# Validate postcodes
postcode validate EK-01-A03-FK-01 "LA 11 W06 TC 10"

# Parse into JSON
postcode parse "FC 03 B06 AG 12" -o json

# Extract coordinates
postcode coords EK-01-A03-FK-01

# Enrich an entire CSV database with coordinates in seconds
postcode batch --input customers.csv --output enriched.csv --column postcode
```

## Detailed Documentation

For full command references, flags, architecture details, environment variables, production recipes, and shell integration, see the **[CLI User Guide & Reference](./docs/README.md)**.
