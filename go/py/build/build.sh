#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/../.." && pwd)"
output_dir="${script_dir}/dist/gopy"

if ! command -v gopy >/dev/null 2>&1; then
  echo "Tier 2 build requires gopy + CPython 3.13+ -- see README.md"
  exit 0
fi

python_bin="${PYTHON:-python3.13}"
if ! command -v "${python_bin}" >/dev/null 2>&1; then
  echo "Tier 2 build requires gopy + CPython 3.13+ -- see README.md"
  exit 0
fi

mkdir -p "${output_dir}"
cd "${repo_root}"

gopy build \
  -output="${output_dir}" \
  -name=core \
  -vm="${python_bin}" \
  ./bindings/...
