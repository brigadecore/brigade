#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

if [[ "${1:-}" == "--watch" ]]; then
  shift

  if ! hash entr 2>/dev/null; then
    echo "entr is required" 1>&2
    exit 1
  fi

  echo bin/brigade | entr -cr bin/brigade "$@"
else
  bin/brigade "$@"
fi

