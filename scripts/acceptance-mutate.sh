#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/011_selection_and_editing.feature "$@"
