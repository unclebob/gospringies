#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/022_adaptive_rk4_numerics.feature "$@"
