#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/005_force_evaluation.feature "$@"
