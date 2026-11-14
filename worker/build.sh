#!/usr/bin/env sh
# Builds the Worker: the Go API compiled to WebAssembly plus the static files from web/.
set -eu
cd "$(dirname "$0")"
rm -rf build
(
  cd ../api
  # workers-assets-gen recreates the output directory, so it runs first.
  go run github.com/syumai/workers-go/cmd/workers-assets-gen -mode=go -o ../worker/build
  GOOS=js GOARCH=wasm go build -trimpath -ldflags='-s -w' -o ../worker/build/app.wasm ./cmd/worker
)
mkdir -p build/public
cp -R ../web/. build/public/
cp _headers build/public/_headers
