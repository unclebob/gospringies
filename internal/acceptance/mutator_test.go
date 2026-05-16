package acceptance

import (
	"testing"

	"springs/internal/gherkin"
)

func TestBuildMutationsUsesStableExampleCellPaths(t *testing.T) {
	feature := gherkin.Feature{Scenarios: []gherkin.Scenario{{
		Name: "mutate",
		Examples: []map[string]string{
			{"name": "Ada", "count": "20"},
		},
	}}}

	mutations := BuildMutations(feature)
	if len(mutations) != 2 {
		t.Fatalf("mutation count = %d", len(mutations))
	}
	if mutations[0].ID != "m1" || mutations[0].Path != "$.scenarios[0].examples[0].count" {
		t.Fatalf("first mutation = %#v", mutations[0])
	}
	if mutations[1].ID != "m2" || mutations[1].Path != "$.scenarios[0].examples[0].name" {
		t.Fatalf("second mutation = %#v", mutations[1])
	}
	if mutations[0].Mutated == mutations[0].Original {
		t.Fatal("mutation did not change value")
	}
}
