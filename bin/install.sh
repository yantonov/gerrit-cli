#!/usr/bin/env sh
set -o errexit -o nounset

script_dir="$(cd "$(dirname "$0")" && pwd)"
cd "$script_dir/.."

artifact="target/gerrit-cli"
install_dir="${HOME}/.local/bin"
install_path="${install_dir}/gerrit-cli"

if [ ! -f "$artifact" ]; then
  echo "Missing binary artifact: $artifact" >&2
  echo "Building it now..." >&2
  "$script_dir/build.sh"
fi

mkdir -p "$install_dir"
cp "$artifact" "$install_path"
chmod +x "$install_path"

echo "Installed: $install_path"
