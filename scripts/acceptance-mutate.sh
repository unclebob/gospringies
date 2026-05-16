#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/008_ebitengine_window.feature "$@"
