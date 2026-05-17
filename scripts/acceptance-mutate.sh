#!/bin/sh
set -eu

go run ./cmd/gherkin-mutator --feature features/018_selected_object_parameter_editing.feature "$@"
