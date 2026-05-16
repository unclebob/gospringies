#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/015_edit_mode_details.feature "$@"
