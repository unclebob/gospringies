package gherkin

import (
	"strings"
	"testing"
)

func TestParseFeatureWithBackgroundScenarioOutlineAndExamples(t *testing.T) {
	source := `
Feature: Spring greeting

Background:
  Given the spring greeter is ready

Scenario Outline: greet a person
  When I greet <name>
  Then the greeting should be <greeting>

Examples:
  | name | greeting |
  | Ada  | Hello, Ada |
  | Bob  | Hello, Bob |
`

	feature, err := Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if feature.Name != "Spring greeting" {
		t.Fatalf("feature name = %q", feature.Name)
	}
	if len(feature.Background) != 1 || feature.Background[0].Text != "the spring greeter is ready" {
		t.Fatalf("background not parsed: %#v", feature.Background)
	}
	if len(feature.Scenarios) != 1 {
		t.Fatalf("scenario count = %d", len(feature.Scenarios))
	}
	scenario := feature.Scenarios[0]
	if len(scenario.Steps) != 2 {
		t.Fatalf("step count = %d", len(scenario.Steps))
	}
	if got := scenario.Steps[0].Parameters; len(got) != 1 || got[0] != "name" {
		t.Fatalf("parameters = %#v", got)
	}
	if got := scenario.Examples[1]["greeting"]; got != "Hello, Bob" {
		t.Fatalf("second greeting = %q", got)
	}
}

func TestParseRejectsMissingFeature(t *testing.T) {
	_, err := Parse(strings.NewReader("Scenario: missing feature\n"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRejectsMalformedExamples(t *testing.T) {
	source := `
Feature: Broken
Scenario Outline: bad examples
  Given a value <value>
Examples:
  | value | other |
  | one |
`
	_, err := Parse(strings.NewReader(source))
	if err == nil {
		t.Fatal("expected error")
	}
}
