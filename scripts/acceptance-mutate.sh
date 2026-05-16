#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/012_controls_and_hotkeys.feature "$@"
