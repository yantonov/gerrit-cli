#!/usr/bin/env sh
set -o errexit -o nounset

cd "$(dirname "$0")/.."

# -s -w drop the symbol table and the DWARF sections. The binary is only ever
# run, never attached to by a debugger, and a Go panic still names its frames:
# stack traces come from the runtime's own line table, which is not stripped.
ldflags="-s -w"

# Stamped into the binary and reported by --version. Outside a checkout, or
# without git, the binary keeps the "dev" fallback instead of claiming a
# revision.
version="$(git describe --tags --always --dirty 2>/dev/null || true)"
if [ -n "$version" ]; then
  ldflags="$ldflags -X gerrit-cli/cmd/app/command.version=$version"
fi

echo "Building gerrit-cli ${version:-dev}..."

# The package, not main.go: building a file list silently ignores every other
# file in package main, so the day a second one appears the build drops it.
go build -ldflags "$ldflags" -o target/gerrit-cli ./cmd/app

echo "Build complete: target/gerrit-cli"
