#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
export GOCACHE="${GOCACHE:-$ROOT/.gocache}"

cd "$ROOT"

./scripts/acceptance.sh
go test -tags appunit ./internal/app ./internal/acceptance ./internal/sim
