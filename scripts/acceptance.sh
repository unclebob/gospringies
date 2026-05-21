#!/bin/sh
set -eu

FEATURE="${1:-features/024_2_clickable_visible_controls.feature}"
BASENAME=$(basename "$FEATURE" .feature)
BUILD_DIR="${ACCEPTANCE_BUILD_DIR:-build/acceptance}"
GENERATED_DIR="${ACCEPTANCE_GENERATED_DIR:-acceptance/generated}"
JSON_OUTPUT="${BUILD_DIR}/${BASENAME}.json"
GENERATED_OUTPUT="${GENERATED_DIR}/${BASENAME}_acceptance_test.go"
GO_TEST_TAGS="${ACCEPTANCE_GO_TEST_TAGS:-appunit}"

rm -rf "$GENERATED_DIR"
mkdir -p "$BUILD_DIR" "$GENERATED_DIR"

go run ./cmd/gherkin-parser \
  "$FEATURE" \
  "$JSON_OUTPUT"

go run ./cmd/acceptance-generator \
  "$JSON_OUTPUT" \
  "$GENERATED_OUTPUT"

if [ -n "$GO_TEST_TAGS" ]; then
  go test -tags "$GO_TEST_TAGS" "./$GENERATED_DIR"
else
  go test "./$GENERATED_DIR"
fi
