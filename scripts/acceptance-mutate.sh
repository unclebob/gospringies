#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/014_packaging_and_docs.feature "$@"
