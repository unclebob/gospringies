#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/002_acceptance_pipeline.feature "$@"
