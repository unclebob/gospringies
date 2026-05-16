#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/016_spring_mode_mouse_semantics.feature "$@"
