#!/usr/bin/env sh
set -o errexit -o nounset

cd "$(dirname "$0")/.."

artifact="target/gerrit-cli"
install_dir="${HOME}/bin"
install_path="${install_dir}/gerrit-cli"

if [ ! -f "$artifact" ]; then
  echo "Missing binary artifact: $artifact" >&2
  echo "Run bin/build.sh first." >&2
  exit 1
fi

mkdir -p "$install_dir"
cp "$artifact" "$install_path"
chmod +x "$install_path"

echo "Installed: $install_path"
