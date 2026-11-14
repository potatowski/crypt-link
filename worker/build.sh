#!/usr/bin/env sh
# Builds the Worker: the Go API compiled to WebAssembly with TinyGo, plus the
# static files from web/. TinyGo keeps the module small and fast to start,
# which keeps CPU time per request within the Workers free plan.
set -eu
cd "$(dirname "$0")"
rm -rf build
(
  cd ../api
  # workers-assets-gen recreates the output directory, so it runs first.
  go run github.com/syumai/workers-go/cmd/workers-assets-gen -mode=tinygo -o ../worker/build
  tinygo build -o ../worker/build/app.wasm -target wasm -no-debug ./cmd/worker
)
mkdir -p build/public
cp -R ../web/. build/public/
cp _headers build/public/_headers
# Replace the generated entry point with one that reuses the Go instance.
cp src/worker.mjs build/worker.mjs
