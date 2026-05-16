#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/004_system_parameters.feature "$@"
