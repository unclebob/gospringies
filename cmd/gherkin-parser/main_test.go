package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunParsesFeatureToJSON(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "project.feature")
	outputPath := filepath.Join(dir, "out", "project.json")
	writeFile(t, featurePath, "Feature: Project\n\nScenario: one\n  Given a thing\n")

	if code := run([]string{"gherkin-parser", featurePath, outputPath}); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsWrongArgumentCount(t *testing.T) {
	if code := run([]string{"gherkin-parser"}); code != 2 {
		t.Fatalf("exit code = %d", code)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
