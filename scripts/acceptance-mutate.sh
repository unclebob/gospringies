#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/003_domain_model.feature "$@"
