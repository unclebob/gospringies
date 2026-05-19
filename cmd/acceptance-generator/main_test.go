package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesAcceptanceTest(t *testing.T) {
	dir := t.TempDir()
	irPath := filepath.Join(dir, "feature.json")
	outputPath := filepath.Join(dir, "generated", "feature_acceptance_test.go")
	writeFile(t, irPath, `{"name":"Project","scenarios":[]}`)

	if code := run([]string{"acceptance-generator", irPath, outputPath}); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRunRejectsWrongArgumentCount(t *testing.T) {
	if code := run([]string{"acceptance-generator"}); code != 2 {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRunReturnsFailureForInvalidInput(t *testing.T) {
	dir := t.TempDir()
	irPath := filepath.Join(dir, "feature.json")
	outputPath := filepath.Join(dir, "generated", "feature_acceptance_test.go")
	writeFile(t, irPath, "{")

	if code := run([]string{"acceptance-generator", irPath, outputPath}); code != 1 {
		t.Fatalf("exit code = %d", code)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
