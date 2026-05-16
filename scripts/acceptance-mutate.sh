#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/008a_screen_and_controls.feature "$@"
