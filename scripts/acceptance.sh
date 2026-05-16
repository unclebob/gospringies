#!/bin/sh
set -eu

rm -rf acceptance/generated
mkdir -p build/acceptance acceptance/generated

go run ./cmd/gherkin-parser \
  features/001_project_skeleton.feature \
  build/acceptance/001_project_skeleton.json

go run ./cmd/acceptance-generator \
  build/acceptance/001_project_skeleton.json \
  acceptance/generated/001_project_skeleton_acceptance_test.go

go test ./acceptance/generated
