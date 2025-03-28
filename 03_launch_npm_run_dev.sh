#!/usr/bin/env bash
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${__dir}"
set -xefu
cd reactweb
exec ./npm_run_dev.sh