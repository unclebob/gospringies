#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/010_mouse_editing.feature "$@"
