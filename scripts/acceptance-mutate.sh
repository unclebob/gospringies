#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/013_demo_files.feature "$@"
