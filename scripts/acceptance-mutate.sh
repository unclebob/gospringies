#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/020_xsp_complete_file_format.feature "$@"
