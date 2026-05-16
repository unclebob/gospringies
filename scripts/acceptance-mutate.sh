#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/007_xsp_load_save.feature "$@"
