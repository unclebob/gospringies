#!/bin/sh
set -eu

FEATURE="${1:-features/011_selection_and_editing.feature}"
BASENAME=$(basename "$FEATURE" .feature)
BUILD_DIR="${ACCEPTANCE_BUILD_DIR:-build/acceptance}"
GENERATED_DIR="${ACCEPTANCE_GENERATED_DIR:-acceptance/generated}"
JSON_OUTPUT="${BUILD_DIR}/${BASENAME}.json"
GENERATED_OUTPUT="${GENERATED_DIR}/${BASENAME}_acceptance_test.go"

rm -rf "$GENERATED_DIR"
mkdir -p "$BUILD_DIR" "$GENERATED_DIR"

go run ./cmd/gherkin-parser \
  "$FEATURE" \
  "$JSON_OUTPUT"

go run ./cmd/acceptance-generator \
  "$JSON_OUTPUT" \
  "$GENERATED_OUTPUT"

go test "./$GENERATED_DIR"
