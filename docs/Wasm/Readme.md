# Nigerian Postcode Go SDK (`postcode`) — WebAssembly (Wasm) Build Guide


## Building the WASM Binary

```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o wasm/postcode.wasm ./wasm
```
