#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/019_wall_collision_and_stickiness.feature "$@"
