#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/009_render_world.feature "$@"
