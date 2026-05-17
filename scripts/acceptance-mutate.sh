#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/024_1_render_visible_controls.feature "$@"
