#!/usr/bin/env sh
set -o errexit -o nounset

cd "$(dirname "$0")/.."

echo "Building gerrit-cli..."
go build -o target/gerrit-cli cmd/app/main.go

echo "Build complete: target/gerrit-cli"
