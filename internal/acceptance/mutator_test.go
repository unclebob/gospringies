package acceptance

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
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

func TestBuildMutationsSkipsEquivalentDomainModelCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Domain model",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{{"mass_count": "0"}}},
			{Examples: []map[string]string{{"id": "1", "reason": "duplicate id"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 2 {
		t.Fatalf("mutation count = %d: %#v", len(mutations), mutations)
	}
	if mutations[0].Key != "mass_count" || mutations[1].Key != "reason" {
		t.Fatalf("mutations = %#v", mutations)
	}
	if !isEquivalentMutation(feature, 1, "id") {
		t.Fatal("expected domain property mutation to be equivalent")
	}
	if isEquivalentMutation(feature, 1, "reason") {
		t.Fatal("reason mutations should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentSystemParameterSetupCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "System parameters",
		Scenarios: []gherkin.Scenario{
			{},
			{},
			{},
			{Examples: []map[string]string{{"parameter": "viscosity", "changed_value": "custom", "operation": "reset"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 1 {
		t.Fatalf("mutation count = %d: %#v", len(mutations), mutations)
	}
	if mutations[0].Key != "operation" {
		t.Fatalf("mutations = %#v", mutations)
	}
	if !isEquivalentSystemParameterMutation(3, "parameter") {
		t.Fatal("expected system parameter setup mutation to be equivalent")
	}
	if isEquivalentSystemParameterMutation(3, "operation") {
		t.Fatal("operation mutation should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentForceEvaluationSetupCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Force evaluation",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{{"mass_a": "1", "mass_b": "2", "expected": "opposite"}}},
			{Examples: []map[string]string{{"damping_constant": "1", "expected": "directional"}}},
			{},
			{Examples: []map[string]string{{"mass_id": "1", "acceleration": "zero"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 3 {
		t.Fatalf("mutation count = %d: %#v", len(mutations), mutations)
	}
	if mutations[0].Key != "expected" || mutations[1].Key != "expected" || mutations[2].Key != "acceleration" {
		t.Fatalf("mutations = %#v", mutations)
	}
	if !isEquivalentForceEvaluationMutation(0, "mass_a") {
		t.Fatal("expected force setup mutation to be equivalent")
	}
	if isEquivalentForceEvaluationMutation(0, "expected") {
		t.Fatal("expected assertion mutation to remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentSimulationStepMassID(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Simulation step",
		Scenarios: []gherkin.Scenario{
			{},
			{Examples: []map[string]string{{"mass_id": "1", "fixed": "true"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 1 || mutations[0].Key != "fixed" {
		t.Fatalf("mutations = %#v", mutations)
	}
	if !isEquivalentSimulationStepMutation(1, "mass_id") {
		t.Fatal("expected fixed-mass id mutation to be equivalent")
	}
	if isEquivalentSimulationStepMutation(1, "fixed") {
		t.Fatal("fixed state mutation should remain meaningful")
	}
}

func TestBuildMutationReturnsStableMutationOrSkipsEquivalent(t *testing.T) {
	feature := gherkin.Feature{Scenarios: []gherkin.Scenario{{}}}
	mutation, ok := buildMutation(feature, 0, 0, "count", "20", 1)
	if !ok {
		t.Fatal("expected mutation")
	}
	if mutation.ID != "m1" || mutation.Path != "$.scenarios[0].examples[0].count" || mutation.Key != "count" {
		t.Fatalf("mutation = %#v", mutation)
	}

	domainFeature := gherkin.Feature{Name: "Domain model", Scenarios: []gherkin.Scenario{{}, {}}}
	if _, ok := buildMutation(domainFeature, 1, 0, "id", "1", 1); ok {
		t.Fatal("expected equivalent mutation to be skipped")
	}
}

func TestRunMutationsReturnsNoResultsWhenFeatureHasNoExamples(t *testing.T) {
	results, err := RunMutations(gherkin.Feature{Scenarios: []gherkin.Scenario{{Name: "empty"}}}, t.TempDir())
	if err != nil {
		t.Fatalf("RunMutations returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("results = %#v", results)
	}
}

func TestSummarizeCountsMutationStatuses(t *testing.T) {
	summary := Summarize([]MutationResult{
		{Status: "killed"},
		{Status: "survived"},
		{Status: "error"},
	})

	if summary.Total != 3 || summary.Killed != 1 || summary.Survived != 1 || summary.Errors != 1 {
		t.Fatalf("summary = %#v", summary)
	}
}

func TestMutateValueHandlesSupportedScalarShapes(t *testing.T) {
	cases := []string{
		"true",
		"false",
		"null",
		"10",
		"10.5",
		"2026-05-16",
		"3s",
		"alpha",
		"a, b, c",
	}
	for _, value := range cases {
		if got := mutateValue("path."+value, value); got == value {
			t.Fatalf("mutateValue(%q) did not change value", value)
		}
	}
}

func TestMutationPaths(t *testing.T) {
	generated, ir := mutationPaths("work", Mutation{ID: "m1"})

	if generated != filepath.Join("work", "m1", "generated", "feature_acceptance_test.go") {
		t.Fatalf("generated path = %s", generated)
	}
	if ir != filepath.Join("work", "m1", "feature.json") {
		t.Fatalf("ir path = %s", ir)
	}
}

func TestWriteMutationTestCreatesTaggedGeneratedTest(t *testing.T) {
	workDir := t.TempDir()
	generated := filepath.Join(workDir, "generated", "feature_acceptance_test.go")
	ir := filepath.Join(workDir, "feature.json")
	feature := gherkin.Feature{Scenarios: []gherkin.Scenario{{
		Examples: []map[string]string{{"count": "1"}},
	}}}
	mutation := Mutation{Scenario: 0, Example: 0, Key: "count", Mutated: "2"}

	if err := writeMutationTest(feature, mutation, generated, ir); err != nil {
		t.Fatalf("writeMutationTest returned error: %v", err)
	}
	data, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "//go:build acceptance_mutation") {
		t.Fatalf("generated test missing build tag:\n%s", data)
	}
}

func TestMutationStatus(t *testing.T) {
	if mutationStatus(nil) != "survived" {
		t.Fatal("nil error should survive")
	}
	if mutationStatus(os.ErrNotExist) != "killed" {
		t.Fatal("non-nil error should be killed")
	}
}

func TestMutationHelpers(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	if got := mutateList("path", "a, b", rng); got == "a, b" {
		t.Fatal("list mutation did not change value")
	}
	if got, ok := mutateKeyword("true", "true", rng); !ok || got != "false" {
		t.Fatalf("keyword mutation = %q, %v", got, ok)
	}
	if got, ok := mutateNumber("12", rng); !ok || got == "12" {
		t.Fatalf("number mutation = %q, %v", got, ok)
	}
	if got, ok := mutateDate("2026-05-16", rng); !ok || got == "2026-05-16" {
		t.Fatalf("date mutation = %q, %v", got, ok)
	}
	if got, ok := mutateDuration("2s", rng); !ok || got == "2s" {
		t.Fatalf("duration mutation = %q, %v", got, ok)
	}
}
