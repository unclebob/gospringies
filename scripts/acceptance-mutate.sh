#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/006_simulation_step.feature "$@"
