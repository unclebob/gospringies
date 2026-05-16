#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/021_force_center_and_force_parameters.feature "$@"
