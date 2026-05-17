#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/023_1_nonblank_startup_editor.feature "$@"
