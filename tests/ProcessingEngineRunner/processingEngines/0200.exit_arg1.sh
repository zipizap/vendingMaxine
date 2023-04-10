#!/usr/bin/env bash
set -xefu
arg1="${1:-0}"
echo "Exiting with code arg1 '${arg1}'"
exit ${arg1}