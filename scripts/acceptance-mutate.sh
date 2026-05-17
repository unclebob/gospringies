#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/023_nonblank_startup_editor.feature "$@"
