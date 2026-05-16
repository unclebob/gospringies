#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/001_project_skeleton.feature "$@"
