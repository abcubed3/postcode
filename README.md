# Nigerian Postcode Go SDK (`postcode`)

A high-performance, zero-allocation Go client library for Nigeria's National Digital Alphanumeric Postcode system ([docs.postcode.gov.ng](https://docs.postcode.gov.ng/)).

[![Go Reference](https://pkg.go.dev/badge/github.com/abcubed3/postcode.svg)](https://pkg.go.dev/github.com/abcubed3/postcode)
[![Go Version](https://img.shields.io/badge/go-1.27%2B-blue)](https://go.dev/)

## Features

- **Zero-Allocation Parsing**: `postcode.Parse()` parses and validates 11-character postcodes with 0 heap escapes (~28ns/op).
- **Safe & Idiomatic Go**: Fully memory-safe accessors, Go range-over-func iterators (`postcode.ParseSeq`), standard errors (`errors.As`), and functional options.
- **Protocol Conformance**: Built to match the official NIPOST OpenAPI spec (`GET /v1/lookup`, `GET /v1/search/autocomplete`, `GET /v1/search/nearby`, `GET /v1/search/reverse`, and `POST /v1/assembly/assemble`).
- **Resilient Transport**: Production-tuned HTTP transport with exponential backoff, full jitter, `Retry-After` header support, and safe connection draining.
- **Pluggable Observability**: Zero external dependencies in the core module, with an optional OpenTelemetry module (`github.com/abcubed3/postcode/otelpostcode`).

## Installation

Requires **Go 1.27+**:

```bash
go get github.com/abcubed3/postcode
```

## Quickstart

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/abcubed3/postcode"
)

func main() {
    ctx := context.Background()

    // 1. High-throughput client with functional options
    client, err := postcode.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    // 2. Graded postcode lookup (Level 1 is free, Level 2-3 require commercial credits)
    res, err := client.Lookup(ctx, "EK 01 A03 FK 01", postcode.Level1)
    if err != nil {
        var apiErr *postcode.APIError
        if errors.As(err, &apiErr) && apiErr.IsInsufficientCredits() {
            log.Fatalf("Top up required: %v", apiErr)
        }
        log.Fatalf("Lookup failed: %v", err)
    }
    fmt.Printf("Postcode: %s, Valid: %t\n", res.Postcode, res.Valid)

    // 3. Segment-aware autocomplete
    suggs, err := client.Autocomplete(ctx, "EK 01 A")
    if err != nil {
        log.Fatalf("Autocomplete failed: %v", err)
    }
    fmt.Printf("Active segment: %s, Suggestions: %+v\n", suggs.Segment, suggs.Suggestions)

    // 4. Generate Google Maps URL & Coordinates from 11-digit postcode string
    gmapsURL, err := postcode.GoogleMapsURL("EK-01-A03-FK-01")
    if err != nil {
        log.Fatalf("Mapping failed: %v", err)
    }
    lat, lng, _ := postcode.Coordinates("EK-01-A03-FK-01")
    fmt.Printf("Google Maps URL: %s\n", gmapsURL)
    fmt.Printf("Coordinates: %.6f, %.6f\n", lat, lng)

    // 5. Batch streaming parse 
    records := []string{"EK 01 A03 FK 01", "INVALID_CODE", "FC 02 A09 DB 09"}
    for p, err := range postcode.ParseSeq(records) {
        if err != nil {
            fmt.Printf("Skipping invalid entry: %v\n", err)
            continue
        }
        fmt.Printf("Parsed: %s (State: %s, LGA: %s, Unit: %s)\n",
            p.Formatted(), p.State(), p.LGA(), p.BuildingUnit())
    }
}
```

## API Surface

| Method / Function | Endpoint / Scope | Description |
| :--- | :--- | :--- |
| `postcode.GoogleMapsURL(code)` | Offline / Local | Generates direct universal Google Maps URL for an 11-digit postcode. |
| `postcode.Coordinates(code)` | Offline / Local | Returns `(latitude, longitude)` for an 11-digit postcode string. |
| `postcode.ResolveLocation(code)` | Offline / Local | Resolves `Location` struct (lat/long, precision, Google Maps/Apple/OSM URLs). |
| `p.GoogleMapsURL()` | Method on `Postcode` | Zero-allocation / sub-microsecond Google Maps URL generation. |
| `client.ResolveLocation(ctx, code)` | Gateway + Offline | Resolves building location with Level 3 live enrichment and offline fallback. |
| `client.Lookup(ctx, code, level)` | `GET /v1/lookup` | Graded postcode lookup (Levels 1–3 cumulative per official spec). |
| `client.Autocomplete(ctx, q)` | `GET /v1/search/autocomplete` | Segment-aware suggestions for partial input. |
| `client.Nearby(ctx, params)` | `GET /v1/search/nearby` | Units within a radius (default 300m) of a coordinate. |
| `client.Reverse(ctx, params)` | `GET /v1/search/reverse` | Snaps coordinate to nearest active unit (default 25m). |
| `client.Assemble(ctx, segs)` | `POST /v1/assembly/assemble` | Assembles 5 segments into canonical format. |
| `client.Disassemble(ctx, code)` | `GET /v1/assembly/disassemble` | Decomposes code into its 5 administrative segments. |

## Command-Line Interface (`postcode`)

The repository includes a production-grade, Cobra- and Viper-powered CLI binary for terminal and pipeline workflows:

```bash
# Build and install the CLI
go install ./cmd/postcode
```

### Key CLI Capabilities

- **Offline Validation**: `postcode validate "EK 01 A03 FK 01"` (grammar, format, state codes, exit codes).
- **Segment Parsing**: `postcode parse "LA 11 W06 TC 10" -o json` (extracts State, LGA, District, Area, Unit, Zone).
- **Format Normalization**: `cat dirty.txt | postcode format --style canonical` (Unix pipe friendly).
- **Geocoding & Maps**: `postcode map EK-01-A03-FK-01` (generates Google Maps, Apple Maps, OSM URLs).
- **Batch CSV Processing**: `postcode batch --input orders.csv --output enriched.csv --column postcode` (processes >350k rows/sec offline).
- **Gateway Operations**: `postcode lookup`, `postcode autocomplete`, `postcode nearby`, `postcode reverse`, and `postcode status`.
- **Embedded Simulator**: `postcode serve --port 8080` (runs local mock NIPOST gateway directly).

See `postcode --help` or individual command help (`postcode <cmd> --help`) for full flag details.

## Live Server Simulator

For local development and automated CI testing without incurring API fees or requiring live network access, the SDK includes an in-memory mock server pre-loaded with all official test postcodes from [docs.postcode.gov.ng](https://docs.postcode.gov.ng/concepts/lookup-levels#test-postcodes):

### 1. Run as a Standalone Server

```bash
# Start on localhost:8080 (or specify -port 8080)
go run ./cmd/simulator
```

Test with `curl`:

```bash
# L1 Public validity lookup (Free, no auth)
curl "http://localhost:8080/v1/lookup?code=EK-01-A03-FK-01&level=1"

# L3 Commercial lookup (Requires X-API-Key header)
curl "http://localhost:8080/v1/lookup?code=EK-01-A03-FK-01&level=3" \
  -H "X-API-Key: nipost_live_test"
```

### 2. Embed Directly in Go Unit & Integration Tests

```go
import "github.com/abcubed3/postcode/simulator"

func TestMyService(t *testing.T) {
    srv := simulator.NewServer() // Starts httptest.Server
    defer srv.Close()

    client, err := postcode.NewClient(
        postcode.WithBaseURL(srv.URL),
        postcode.WithAPIKey("test_key"),
    )
    // Run your application tests against srv.URL
}
```

### 3. Preloaded Test Postcodes

| Postcode | State | Locality / Address |
| :--- | :--- | :--- |
| `EK-01-A03-FK-01` | Ekiti | NTA Road, Back of Fabian Hotel, Ado Ekiti |
| `AK-11-I61-ZF-12` | Akwa Ibom | 12 Oron Road, Uyo |
| `AK-11-H40-WD-11` | Akwa Ibom | 11 Wellington Bassey Way, Uyo |
| `BA-02-M67-BL-69` | Bauchi | 69 Bank Road, GRA, Bauchi |
| `BA-02-E99-NE-30` | Bauchi | 30 Ahmadu Bello Way, Bauchi |
| `EB-13-G95-FR-90` | Ebonyi | 90 Ogoja Road, Abakaliki |
| `EB-13-I97-AB-30` | Ebonyi | 30 Water Works Road, Abakaliki |
| `EN-05-V19-CD-22` | Enugu | 22 Chime Avenue, New Haven, Enugu |
| `EN-05-V19-FT-20` | Enugu | 20 Ogui Road, Enugu |
| `FC-03-B06-AG-12` | FCT | 12 Shehu Shagari Way, Garki, Abuja |
| `FC-02-B19-RT-30` | FCT | 30 Gado Nasko Way, Phase 4, Kubwa, Abuja |
| `JI-24-O18-JP-23` | Jigawa | 23 Sani Abacha Way, Dutse |
| `JI-24-N11-VM-58` | Jigawa | 58 Kano-Dutse Expressway, Dutse |
| `KN-31-F82-WJ-80` | Kano | 80 Badu Road, Bompai, Kano |
| `KN-31-D78-IQ-38` | Kano | 38 Ibrahim Taiwo Road, Kano |
| `LA-11-W06-TC-10` | Lagos | 10 Obafemi Awolowo Way, Ikeja, Lagos |
| `LA-11-U34-ZR-63` | Lagos | 63 Isaac John Street, GRA Ikeja, Lagos |
| `NI-09-J67-QC-65` | Niger | 65 Bosso Road, Minna |
| `NI-09-A75-DA-10` | Niger | 10 Paida Road, Minna |
| `OG-14-T18-BN-16` | Ogun | 16 Lalubu Street, Oke-Ilewo, Abeokuta |
| `OG-14-M82-QA-09` | Ogun | 9 Quarry Road, Abeokuta |

## Rate Limiting & Resilient Retries

As specified in the [NIPOST Authentication Documentation](https://docs.postcode.gov.ng/authentication), all authenticated gateway requests are rate-limited per key. The gateway returns quota headers on every response:

```http
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 597
```

When quota is exceeded, the gateway returns `429 Too Many Requests` with a cooldown duration.

### Intelligent 429 Handling

- **Automatic Cooldown Synchronization**: On receiving HTTP 429, the transport extracts cooldown intervals from `Retry-After` (integer seconds or RFC1123 HTTP dates) and `X-RateLimit-Reset`, adding randomized jitter to avoid thundering herds when the window resets.
- **Fail-Fast Defense (`MaxRateLimitDelay`)**: If the gateway requests a cooldown longer than the configured threshold (default: 10s), the client aborts immediately, preventing goroutine hangs and wasted retries.
- **Live Quota Introspection**: Call `client.RateLimit()` at any time to inspect `Remaining`, `Limit`, and `ResetAt` in a thread-safe manner without making extra network calls.
- **Rich Error Context**: `apiErr.IsRateLimit()` returns `true` on 429 errors, exposing `apiErr.RetryAfter` and `apiErr.RateLimit`.

```go
// 1. Fine-tune rate limit retries
client, err := postcode.NewClient(
    postcode.WithAPIKey("nipost_live_..."),
    postcode.WithRateLimitRetry(true, 5*time.Second), // Max wait threshold: 5s
)

// 2. Inspect rate limit quota
if rl := client.RateLimit(); rl != nil {
    fmt.Printf("Quota remaining: %d/%d (resets at %v)\n", rl.Remaining, rl.Limit, rl.ResetAt)
}

// 3. Structured error handling
res, err := client.Lookup(ctx, "EK-01-A03-FK-01", postcode.Level2)
if err != nil {
    var apiErr *postcode.APIError
    if errors.As(err, &apiErr) && apiErr.IsRateLimit() {
        log.Printf("Rate limit exceeded. Try again in %v", apiErr.RetryAfter)
    }
}
```

## OpenTelemetry Integration

Install the telemetry adapter submodule:

```bash
go get github.com/abcubed3/postcode/otelpostcode
```

Hook it into your client:

```go
package main

import (
    "context"
    "log"

    "github.com/abcubed3/postcode"
    "github.com/abcubed3/postcode/otelpostcode"
)

func main() {
    // Initialize tracing and metrics from global OTel providers
    tel, err := otelpostcode.NewTelemetry()
    if err != nil {
        log.Fatal(err)
    }

    client, err := postcode.NewClient(
        postcode.WithAPIKey("YOUR_API_KEY"),
        postcode.WithTelemetry(tel),
    )
    if err != nil {
        log.Fatal(err)
    }

    _, _ = client.Lookup(context.Background(), "EK 01 A03 FK 01", postcode.Level2)
}
```

## Benchmarks

Microbenchmarks measured on `darwin/arm64` (Apple M2 Max) using `go test -bench=. -benchmem`:

```text
BenchmarkParse-12             42424563        28.25 ns/op          0 B/op        0 allocs/op
BenchmarkFormatted-12        195061753         6.15 ns/op          0 B/op        0 allocs/op
BenchmarkResolveLocation-12   10245876       115.70 ns/op         32 B/op        4 allocs/op
BenchmarkGoogleMapsURL-12      3721462       325.20 ns/op        128 B/op        7 allocs/op
```

- **`postcode.Parse`**: Fully validates, normalizes, and extracts segments with **`0 B/op` and `0 allocs/op`** in **~28 ns/op**.
- **`postcode.Formatted`**: Formats canonical hyphenated postcodes (`AA-99-H77-BB-55`) with **`0 B/op` and `0 allocs/op`** in **~6 ns/op**.
- **`postcode.ResolveLocation`**: Resolves building and administrative coordinates offline in **~115 ns/op**.
- **`postcode.GoogleMapsURL`**: Formats and generates universal Google Maps search URLs in **~325 ns/op**.
