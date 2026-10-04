# Postcode Nigeria SDK — Practical Examples

This directory contains real-world, executable examples demonstrating core, advanced, and AI agent capabilities of the Nigerian Postcode SDK.

## Examples Directory

| Example | Focus Area | Description | Command |
| :--- | :--- | :--- | :--- |
| [`main.go`](./../examples/main.go) | **End-to-End Tour** | Full tour covering local parsing, local simulator, graded lookups (L1-L3), autocomplete, reverse geocoding, AI diagnostics, and agent guardrails. | `go run ./examples/main.go` |
| [`ai_agent_guardrails/`](./../examples/ai_agent_guardrails/main.go) | **AI Agent Safety** | Autonomous logistics dispatcher agent demonstrating session call ceilings, in-memory caching, auto-downgrade to free Level 1, and offline fallback. | `go run ./examples/ai_agent_guardrails/main.go` |
| [`ai_synthetic_eval/`](./../examples/ai_synthetic_eval/main.go) | **AI Benchmarking** | Synthetic address generator and evaluation harness benchmark testing extraction accuracy against noisy Nigerian addresses with landmarks, typos, and format variations. | `go run ./examples/ai_synthetic_eval/main.go` |
| [`wasm_node/`](./../examples/wasm_node/demo.js) | **Edge / WebAssembly** | Running the zero-dependency compiled WebAssembly binary inside Node.js for sub-millisecond offline validation and geocoding. | `node ./examples/wasm_node/demo.js` |

## Running All Examples

```bash
# 1. Complete SDK tour
go run ./examples/main.go

# 2. Autonomous Agent Guardrails & Caching
go run ./examples/ai_agent_guardrails/main.go

# 3. Synthetic Address Generator & AI Eval Benchmark
go run ./examples/ai_synthetic_eval/main.go

# 4. WebAssembly in Node.js
node ./examples/wasm_node/demo.js
```
