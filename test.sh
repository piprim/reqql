#!/usr/bin/env bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

for dir in args goqu; do
    cd "${SCRIPT_DIR}/$dir" || exit 1
    echo "=> Testing ${dir}…"
    go test -v "."
    echo
done
