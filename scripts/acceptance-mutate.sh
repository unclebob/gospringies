#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/024_2_clickable_visible_controls.feature "$@"
