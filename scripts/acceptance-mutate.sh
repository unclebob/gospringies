#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/017_state_save_restore.feature "$@"
