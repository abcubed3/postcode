# Contributing to Nigerian Postcode (`postcode`)

Thank you for your interest in contributing to `postcode`! We welcome contributions of all kinds: bug reports, documentation enhancements, feature proposals, performance optimizations, and code contributions.

This document outlines the architecture, development setup, coding standards, and workflows to ensure a smooth collaboration experience.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Repository Architecture & Workspace Layout](#repository-architecture--workspace-layout)
- [Prerequisites & Tooling](#prerequisites--tooling)
- [Setting Up Your Local Environment](#setting-up-your-local-environment)
- [Development Workflow with Taskfile](#development-workflow-with-taskfile)
- [Architectural Rules & Coding Standards](#architectural-rules--coding-standards)
  - [1. Zero-Allocation Philosophy (Hot Paths)](#1-zero-allocation-philosophy-hot-paths)
  - [2. Zero Dependencies in the Core Module](#2-zero-dependencies-in-the-core-module)
  - [3. Idiomatic Go Practices](#3-idiomatic-go-practices)
  - [4. WebAssembly (WASM) Parity](#4-webassembly-wasm-parity)
  - [5. Concurrency & Thread Safety](#5-concurrency--thread-safety)
- [Testing & Mock Simulation](#testing--mock-simulation)
- [Git & Commit Message Guidelines](#git--commit-message-guidelines)
- [Multi-Module Tagging & Releases](#multi-module-tagging--releases)
- [Pull Request Checklist](#pull-request-checklist)

---

## Code of Conduct

We are committed to providing a welcoming, inclusive, and harassment-free experience for everyone. Please be respectful, considerate, and constructive in all interactions, issues, and pull requests.

---

## Repository Architecture & Workspace Layout

This repository is structured as a **Go Multi-Module Workspace** (`go.work`) containing three distinct Go modules alongside a WebAssembly edge runtime package:

```text
/
├── go.work                   # Go workspace binding root, CLI, and otel modules
├── client.go, postcode.go    # Root Module: github.com/abcubed3/postcode
├── location*.go, mcp.go      #   - Zero-allocation parser & validator
├── geocoder.go, cache.go     #   - Offline geocoding & reference data
├── simulator/                #   - In-memory NIPOST mock server
│
├── cmd/postcode/             # CLI Module: github.com/abcubed3/postcode/cmd/postcode
│   ├── main.go               #   - Cobra & Viper CLI binary
│   └── internal/commands/    #   - Validate, Parse, Coords, Batch, MCP, Serve, Eval
│
├── otelpostcode/             # Observability Module: github.com/abcubed3/postcode/otelpostcode
│   └── provider.go           #   - OpenTelemetry metrics and tracing adapter
│
├── wasm/                     # WebAssembly Module (@abcubed3/postcode-wasm)
│   ├── main.go               #   - Go WebAssembly bridge
│   ├── index.js, index.d.ts  #   - TypeScript declarations & Node/browser loader
│   └── test.js               #   - Parity verification suite
│
├── .github/workflows/        # CI/CD pipelines (testing, cross-compile, release, npm)
├── Taskfile.yml              # Automated developer tasks (go-task)
└── CONTRIBUTING.md           # This contributor guide
```

### Module Responsibilities

1. **Root Module (`github.com/abcubed3/postcode`)**:
   - Contains the core postcode parser, validator, location resolution engine, NIPOST HTTP client, MCP server protocols, LLM agent tool schemas, and simulator.
   - **Strict Policy**: Must have **zero external runtime dependencies** (pure Go standard library only).
2. **CLI Module (`cmd/postcode`)**:
   - Houses the command-line interface powered by Cobra and Viper.
   - Provides Unix pipe support, format conversions (`json`, `yaml`, `csv`), batch processing, MCP server mode, and developer utilities.
3. **Observability Module (`otelpostcode`)**:
   - Provides OpenTelemetry instrumentation without polluting the core SDK with external dependencies for users who do not need tracing.
4. **WebAssembly Module (`wasm/`)**:
   - Compiles the offline core engine to WebAssembly (`postcode.wasm`) for edge workers (Cloudflare, Vercel), Node.js, Bun, Deno, and web browsers.

---

## Prerequisites & Tooling

To build, test, and contribute to all parts of the project, ensure you have:

- **Go**: Version **1.27+** installed ([go.dev](https://go.dev/dl/)).
- **Node.js & npm**: Node.js **v24+** (required for running WebAssembly smoke and parity tests).
- **Task (go-task)**: A fast, modern task runner used across this repo ([taskfile.dev](https://taskfile.dev)):
  ```bash
  # macOS (Homebrew)
  brew install go-task

  # Linux / Unix
  sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

  # Via Go directly
  go install github.com/go-task/task/v3/cmd/task@latest
  ```
- **Git**: Configured with your developer identity.

---

## Setting Up Your Local Environment

1. **Fork and Clone the Repository**:
   ```bash
   git clone https://github.com/<your-username>/postcode.git
   cd postcode
   ```

2. **Configure Environment Variables**:
   ```bash
   cp .env.example .env
   ```
   *(The default `.env.example` includes a mock API key suitable for local simulator and offline testing.)*

3. **Synchronize Dependencies Across All Modules**:
   ```bash
   task deps
   ```
   This synchronizes `go.work`, tidies `go.mod` across all three Go modules, and verifies npm dependencies in `wasm/`.

4. **Verify Your Setup**:
   ```bash
   task lint
   task test:cli
   ```

---

## Development Workflow with Taskfile

We use `Taskfile.yml` to automate and standardize common developer commands. Run `task` or `task --list` to view all available commands:

| Command | Description |
| :--- | :--- |
| `task deps` | Syncs `go.work` and runs `go mod tidy` across all modules + `npm install` in `wasm/`. |
| `task fmt` | Formats all Go files using `gofmt -s -w .`. |
| `task lint` | Runs `go vet` across all Go modules (matching CI standards). |
| `task lint:full` | Runs `golangci-lint` for extended static analysis (if installed). |
| `task test` | Runs standard unit tests across all Go modules. |
| `task test:race` | Runs all tests with the Go race detector (`-race`) and generates `coverage.out`. |
| `task test:cli` | Runs the CLI command test suite. |
| `task test:wasm` | Compiles WebAssembly target and runs the Node.js test suite. |
| `task test:all` | Runs both race-detected unit tests and the WebAssembly test suite. |
| `task bench` | Runs all Go microbenchmarks with memory allocation profiling (`-benchmem`). |
| `task bench:hotpaths` | Microbenchmarks hot paths (`Parse`, `Formatted`, `ResolveLocation`, `GoogleMapsURL`). |
| `task build` | Builds the `postcode` CLI binary into `bin/postcode` with git version ldflags. |
| `task build:wasm` | Compiles the WebAssembly binary into `wasm/postcode.wasm`. |
| `task build:all` | Builds both the CLI binary and the WebAssembly target. |
| `task install` | Installs the `postcode` binary to `$GOPATH/bin/postcode`. |
| `task sim` | Starts the in-memory NIPOST mock simulator on `localhost:2340`. |
| `task coverage` | Generates and opens an HTML visual coverage report in your browser. |
| `task clean` | Removes build binaries, coverage files, and compiled artifacts. |
| `task ci` | Runs the complete local verification pipeline (`fmt`, `lint`, `test:race`, `test:wasm`, `build:all`). |
| `task tag` | Prompts for version number, tags synchronized multi-module git releases, and pushes to origin. |

---

## Architectural Rules & Coding Standards

### 1. Zero-Allocation Philosophy (Hot Paths)

The core parser and formatting functions are designed to operate with **zero heap allocations**:

- `postcode.Parse(raw)` must remain **`0 B/op` and `0 allocs/op`** (~20–30 ns/op).
- `p.Formatted()` must remain **`0 B/op` and `0 allocs/op`** (~4–6 ns/op).

> ⚠️ **Allocation Guardrail**: If you modify `postcode.go`, `location.go`, or parsing logic, you **must run `task bench:hotpaths`** before submitting your PR to ensure no new heap escapes were introduced.

```bash
task bench:hotpaths
```

### 2. Zero Dependencies in the Core Module

The root module (`github.com/abcubed3/postcode`) must remain completely **free of third-party dependencies**. 

- Do not add external packages to the root `go.mod`.
- If a feature requires third-party libraries:
  - CLI features go into `cmd/postcode` (which uses Cobra and Viper).
  - Tracing/metrics features go into `otelpostcode` (which uses OpenTelemetry).

### 3. Idiomatic Go Practices

- **Keep the Happy Path Left**: Handle errors and edge cases early and return. Avoid nested `if/else` ladders.
- **Wrap Errors Meaningfully**: Use `fmt.Errorf("...: %w", err)` to preserve error chains, and use `errors.As` or `errors.Is` for inspection.
- **Context Propagation**: Always accept `context.Context` as the first argument for any I/O, network, or long-running operation.
- **Make Zero Values Useful**: Design structs so their zero-value state is valid and ready to use whenever possible.
- **Accept Interfaces, Return Structs**: Functions should return concrete types while accepting the narrowest practical interface (e.g., `io.Reader`, `http.RoundTripper`).

### 4. WebAssembly (WASM) Parity

If you introduce a new feature or API method to the Go core library:
1. Consider whether it belongs in the offline WASM runtime.
2. If applicable, expose the method in `wasm/main.go`.
3. Update TypeScript definitions in `wasm/index.d.ts`.
4. Add verification tests in `wasm/test.js` and verify with `task test:wasm`.

### 5. Concurrency & Thread Safety

All exported clients, caches, and lookup engines must be safe for concurrent use by multiple goroutines.
- Use `sync.RWMutex` or atomic operations where appropriate.
- Never introduce package-level mutable global state in the SDK library.

---

## Testing & Mock Simulation

### Never Call Live NIPOST Production Gateways in Automated Tests

All automated unit and integration tests must run against the in-memory simulator or mock HTTP handlers:

```go
import "github.com/abcubed3/postcode/simulator"

func TestMyFeature(t *testing.T) {
    srv := simulator.NewServer()
    defer srv.Close()

    client, err := postcode.NewClient(
        postcode.WithBaseURL(srv.URL),
        postcode.WithAPIKey("test_key"),
    )
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Run tests against srv.URL
}
```

To run the simulator as a standalone server during local manual testing:
```bash
task sim
# In another terminal:
curl "http://localhost:2340/v1/lookup?code=EK-01-A03-FK-01&level=1"
```

---

## Git & Commit Message Guidelines

We follow the **Conventional Commits** specification (`<type>(<scope>): <short description>`):

- `feat`: A new feature (e.g., `feat(cli): add parquet export to batch command`)
- `fix`: A bug fix (e.g., `fix(parser): handle trailing whitespace in postcode.Parse`)
- `perf`: A performance optimization (e.g., `perf(geocoder): cache spatial index lookups`)
- `docs`: Documentation updates (e.g., `docs: update quickstart in README`)
- `test`: Adding or updating test cases
- `refactor`: Code refactoring without behavior changes
- `chore`: Maintenance tasks, dependencies, or CI updates

### Branch Naming

- `feature/<short-description>`
- `fix/<short-description>`
- `docs/<short-description>`
- `perf/<short-description>`

---

## Multi-Module Tagging & Releases

Because this repository contains multiple Go modules in one Git repository, releases follow Go's standard multi-module versioning rules:

1. **Canonical Root Release (`vX.Y.Z`)**:
   - The root tag `vX.Y.Z` triggers the GitHub Actions CI release workflow, compiling multi-platform binaries, packaging archives, creating the GitHub Release, and publishing the WebAssembly npm package.
2. **Submodule Proxy Tags (`<path>/vX.Y.Z`)**:
   - Submodule tags (`cmd/postcode/vX.Y.Z` and `otelpostcode/vX.Y.Z`) are tagged to allow the Go module proxy (`go get`, `go install`) to resolve submodules at specific versions.

When releasing a synchronized version across the repository:
```bash
# Using Taskfile (prompts for version, or pass via CLI arguments)
task tag
# or: task tag -- 0.2.0

# Or manually:
VERSION="v0.2.0"
git tag "${VERSION}"
git tag "cmd/postcode/${VERSION}"
git tag "otelpostcode/${VERSION}"
git push origin "${VERSION}" "cmd/postcode/${VERSION}" "otelpostcode/${VERSION}"
```

---

## Pull Request Checklist

Before opening your Pull Request, please ensure you have completed the following:

- [ ] Ran `task fmt` to ensure standard Go formatting.
- [ ] Ran `task lint` to verify zero static analysis or `go vet` issues.
- [ ] Ran `task test:all` (or `task test:race` and `task test:wasm`) to verify all tests pass.
- [ ] Ran `task bench:hotpaths` to verify zero allocations in parser/formatting hot paths.
- [ ] Added unit tests covering any new code or bug fixes.
- [ ] Updated relevant documentation (`README.md`, `cmd/postcode/README.md`, or `wasm/README.md`) if public behavior changed.
- [ ] Followed Conventional Commit guidelines for your commit messages.

Thank you for helping make Nigerian digital postcodes fast, accessible, and reliable! 🇳🇬🚀
