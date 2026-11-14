#!/usr/bin/env sh
# Installs the pinned TinyGo release (checksum-verified) and prints its bin dir.
# Usage: worker/install-tinygo.sh <install-dir>
set -eu
VERSION=0.44.0
SHA256=ed0f45543a49ff2a99120a5a1b8814a0750e1fda1a66f8e2a0a172576134aaf9
dir=${1:?usage: install-tinygo.sh <install-dir>}
archive="tinygo${VERSION}.linux-amd64.tar.gz"

mkdir -p "$dir"
curl -fsSL -o "$dir/$archive" "https://github.com/tinygo-org/tinygo/releases/download/v${VERSION}/${archive}"
echo "${SHA256}  $dir/$archive" | sha256sum -c - >&2
tar -xzf "$dir/$archive" -C "$dir"
rm "$dir/$archive"
echo "$dir/tinygo/bin"
