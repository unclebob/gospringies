package acceptance

import (
	"strings"
	"testing"

	"springs/internal/gherkin"
)

func TestRunFeatureExecutesProjectSkeletonScenarios(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Project skeleton",
		Background: []gherkin.Step{
			{Keyword: "Given", Text: "the project skeleton task is accepted"},
		},
		Scenarios: []gherkin.Scenario{{
			Name: "domain independence",
			Steps: []gherkin.Step{
				{Keyword: "When", Text: "the coder creates the initial Go package layout"},
				{Keyword: "Then", Text: "the <package> package should not import <graphics_library>", Parameters: []string{"package", "graphics_library"}},
			},
			Examples: []map[string]string{
				{"package": "simulation", "graphics_library": "Ebitengine"},
				{"package": "file format", "graphics_library": "Ebitengine"},
			},
		}},
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureExecutesPipelineSmokeScenario(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Pipeline smoke",
		Scenarios: []gherkin.Scenario{{
			Name: "smoke",
			Steps: []gherkin.Step{
				{Keyword: "Given", Text: "acceptance smoke is ready"},
				{Keyword: "Then", Text: "acceptance smoke should pass"},
			},
		}},
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureFailsUnsupportedSteps(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Unsupported",
		Scenarios: []gherkin.Scenario{{
			Name:  "bad",
			Steps: []gherkin.Step{{Keyword: "Then", Text: "something unknown happens"}},
		}},
	}

	err := RunFeature(feature)
	if err == nil || !strings.Contains(err.Error(), "unsupported step") {
		t.Fatalf("expected unsupported step error, got %v", err)
	}
}
