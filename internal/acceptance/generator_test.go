package acceptance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"springs/internal/gherkin"
)

func TestGenerateWritesDeterministicGoTestThatEmbedsIR(t *testing.T) {
	feature := gherkin.Feature{Name: "Generated", Scenarios: []gherkin.Scenario{{Name: "empty"}}}
	dir := t.TempDir()
	ir := filepath.Join(dir, "feature.json")
	output := filepath.Join(dir, "feature_acceptance_test.go")
	data, err := json.MarshalIndent(feature, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ir, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateGoTest(ir, output); err != nil {
		t.Fatalf("GenerateGoTest returned error: %v", err)
	}
	first, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if err := GenerateGoTest(ir, output); err != nil {
		t.Fatalf("GenerateGoTest returned error: %v", err)
	}
	second, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}

	if string(first) != string(second) {
		t.Fatal("generated output is not deterministic")
	}
	if !strings.Contains(string(first), "RunFeature") || !strings.Contains(string(first), `"Generated"`) {
		t.Fatalf("generated test does not embed executable IR:\n%s", first)
	}
}

func TestGenerateGoTestRejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "feature.json")
	output := filepath.Join(dir, "generated", "feature_acceptance_test.go")
	if err := os.WriteFile(input, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateGoTest(input, output); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
