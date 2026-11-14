#!/usr/bin/env sh
# Installs the pinned TinyGo release (checksum-verified) and prints its bin dir.
# Usage: worker/install-tinygo.sh <install-dir>
#
# Supports Linux x86-64 only (CI runners); elsewhere install TinyGo $VERSION by hand.
# Keep VERSION in step with the Go toolchain in api/go.mod and the workers-go
# version: each TinyGo release supports a bounded range of Go versions, and
# workers-go ships the wasm_exec glue for it.
set -eu
VERSION=0.44.0
SHA256=ed0f45543a49ff2a99120a5a1b8814a0750e1fda1a66f8e2a0a172576134aaf9
dir=${1:?usage: install-tinygo.sh <install-dir>}
bin="$dir/tinygo/bin"

if [ "$(uname -sm)" != "Linux x86_64" ]; then
  echo "install-tinygo.sh supports Linux x86-64 only; install TinyGo $VERSION from https://tinygo.org/getting-started/install/" >&2
  exit 1
fi

if [ -x "$bin/tinygo" ] && "$bin/tinygo" version | grep -q "tinygo version $VERSION "; then
  echo "$bin"
  exit 0
fi

archive="tinygo${VERSION}.linux-amd64.tar.gz"
mkdir -p "$dir"
curl -fsSL -o "$dir/$archive" "https://github.com/tinygo-org/tinygo/releases/download/v${VERSION}/${archive}"
echo "${SHA256}  $dir/$archive" | sha256sum -c - >&2
rm -rf "$dir/tinygo"
tar -xzf "$dir/$archive" -C "$dir"
rm "$dir/$archive"
echo "$bin"
