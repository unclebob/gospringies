#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/025_original_demo_corpus.feature "$@"
