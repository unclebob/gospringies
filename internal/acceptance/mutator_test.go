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
	if !isEquivalentMutation(feature, 1, 0, "id") {
		t.Fatal("expected domain property mutation to be equivalent")
	}
	if isEquivalentMutation(feature, 1, 0, "reason") {
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

func TestBuildMutationsSkipsEquivalentSimulationStepSetupCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Simulation step",
		Scenarios: []gherkin.Scenario{
			{},
			{Examples: []map[string]string{{"mass_id": "1", "fixed": "true", "duration": "1 step"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 2 {
		t.Fatalf("mutation count = %d: %#v", len(mutations), mutations)
	}
	if !isEquivalentSimulationStepMutation(1, "mass_id") {
		t.Fatal("expected mass id setup mutation to be equivalent")
	}
	if isEquivalentSimulationStepMutation(1, "duration") {
		t.Fatal("duration mutation should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentXSPFixedMassSetupCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "XSP load and save",
		Scenarios: []gherkin.Scenario{
			{},
			{},
			{},
			{Examples: []map[string]string{{"mass_id": "1", "file_mass_value": "-3.0", "fixed": "true", "file_mass_sign": "negative"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 2 {
		t.Fatalf("mutations = %#v", mutations)
	}
	if mutations[0].Key != "file_mass_sign" || mutations[1].Key != "fixed" {
		t.Fatalf("mutations = %#v", mutations)
	}
	if !isEquivalentXSPMutation(3, "file_mass_value") {
		t.Fatal("expected file mass value mutation to be equivalent")
	}
}

func TestBuildMutationsSkipsEquivalentSelectionEditingCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Selection and editing",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{{"object_type": "mass", "id": "1"}}},
			{},
			{Examples: []map[string]string{{"object_type": "spring", "id": "2"}}},
		},
	}

	mutations := BuildMutations(feature)
	if len(mutations) != 2 {
		t.Fatalf("mutations = %#v", mutations)
	}
	if !isEquivalentSelectionEditingMutation(0, "id") || !isEquivalentSelectionEditingMutation(2, "id") {
		t.Fatal("expected selection id setup mutations to be equivalent")
	}
}

func TestBuildMutationsSkipsEquivalentMouseEditingCells(t *testing.T) {
	if !isEquivalentMouseEditingMutation(1, "snap_size") {
		t.Fatal("expected snap size setup mutation to be equivalent")
	}
	if !isEquivalentMouseEditingMutation(2, "mass_a") {
		t.Fatal("expected spring endpoint setup mutation to be equivalent")
	}
	if !isEquivalentMouseEditingMutation(2, "mass_b") {
		t.Fatal("expected second spring endpoint setup mutation to be equivalent")
	}
	if !isEquivalentMouseEditingMutation(3, "mass_id") {
		t.Fatal("expected drag mass id setup mutation to be equivalent")
	}
	if !isEquivalentMouseEditingMutation(3, "start_position") {
		t.Fatal("expected drag start position setup mutation to be equivalent")
	}
	if !isEquivalentMouseEditingMutation(3, "target_position") {
		t.Fatal("expected fixed drag target setup mutation to be equivalent")
	}
	if isEquivalentMouseEditingMutation(1, "grid_snap") {
		t.Fatal("expected grid snap state mutation to be meaningful")
	}
	if isEquivalentMouseEditingMutation(2, "mode") {
		t.Fatal("expected spring mode mutation to be meaningful")
	}
	if isEquivalentMouseEditingMutation(3, "fixed") {
		t.Fatal("expected fixed state mutation to be meaningful")
	}
	if isEquivalentMouseEditingMutation(0, "expected_position") {
		t.Fatal("expected mass placement assertion mutation to be meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentEditModeDetailsCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Edit mode details",
		Scenarios: []gherkin.Scenario{
			{},
			{Examples: []map[string]string{{"inside_objects": "1,2", "outside_objects": "3", "expected_selection": "1,2"}}},
			{Examples: []map[string]string{{"object_id": "1", "drag_delta": "5,-3", "expected_position": "15,7"}}},
			{Examples: []map[string]string{
				{"mass_id": "1", "fixed": "false", "release_velocity": "4,-2", "expected_velocity": "4,-2"},
				{"mass_id": "2", "fixed": "false", "release_velocity": "0,0", "expected_velocity": "0,0"},
				{"mass_id": "3", "fixed": "true", "release_velocity": "4,-2", "expected_velocity": "unchanged"},
			}},
		},
	}

	mutations := BuildMutations(feature)
	for _, mutation := range mutations {
		if mutation.Key == "outside_objects" || mutation.Key == "object_id" || mutation.Key == "mass_id" {
			t.Fatalf("mutation should be filtered: %#v", mutation)
		}
		if mutation.Scenario == 3 && mutation.Example == 2 && mutation.Key == "release_velocity" {
			t.Fatalf("fixed-mass release velocity should be filtered: %#v", mutation)
		}
	}
	if !isEquivalentEditModeDetailsMutation(1, 0, "outside_objects") {
		t.Fatal("expected outside object setup mutation to be equivalent")
	}
	if !isEquivalentEditModeDetailsMutation(3, 2, "release_velocity") {
		t.Fatal("expected fixed-mass release velocity mutation to be equivalent")
	}
	if isEquivalentEditModeDetailsMutation(3, 0, "release_velocity") {
		t.Fatal("movable release velocity mutation should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentSpringModeMouseCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Spring mode mouse semantics",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{{"start_mass": "1", "release_target": "near mass 2", "result": "create spring between 1 and 2"}}},
			{Examples: []map[string]string{{"start_mass": "1", "button": "left", "behavior": "actively affects the first mass"}}},
			{Examples: []map[string]string{{"kspring": "12.0", "kdamp": "0.5", "creation_length": "30.0"}}},
		},
	}

	mutations := BuildMutations(feature)
	for _, mutation := range mutations {
		if mutation.Scenario == 1 && mutation.Key == "start_mass" {
			t.Fatalf("button behavior start mass should be filtered: %#v", mutation)
		}
		if mutation.Scenario == 2 {
			t.Fatalf("shared default/length cells should be filtered: %#v", mutation)
		}
	}
	if !isEquivalentSpringModeMouseMutation(1, 0, "start_mass") {
		t.Fatal("expected button scenario start mass to be equivalent")
	}
	if !isEquivalentSpringModeMouseMutation(2, 0, "creation_length") {
		t.Fatal("expected shared creation length to be equivalent")
	}
	if isEquivalentSpringModeMouseMutation(0, 0, "release_target") {
		t.Fatal("release target mutation should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentStateSaveRestoreRepeatedRestoreCount(t *testing.T) {
	feature := gherkin.Feature{
		Name: "State save restore",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{
				{"saved_state": "A", "changed_state": "B", "restore_count": "1"},
				{"saved_state": "A", "changed_state": "B", "restore_count": "2"},
			}},
		},
	}

	mutations := BuildMutations(feature)
	for _, mutation := range mutations {
		if mutation.Example == 1 && mutation.Key == "restore_count" {
			t.Fatalf("repeated restore count should be filtered: %#v", mutation)
		}
	}
	if !isEquivalentStateSaveRestoreMutation(0, 1, "restore_count") {
		t.Fatal("expected repeated restore count mutation to be equivalent")
	}
	if isEquivalentStateSaveRestoreMutation(0, 0, "restore_count") {
		t.Fatal("single restore count mutation should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentSelectedObjectParameterCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Selected object parameter editing",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{{"mass_id": "1", "control": "mass", "value": "2.0"}}},
			{Examples: []map[string]string{{"spring_id": "1", "control": "Kspring", "value": "15.0"}}},
			{Examples: []map[string]string{{"spring_id": "1", "current_length": "42.0"}}},
			{Examples: []map[string]string{{"control": "mass", "value": "3.0", "object_type": "mass"}}},
		},
	}

	mutations := BuildMutations(feature)
	for _, mutation := range mutations {
		if mutation.Key == "value" || mutation.Key == "mass_id" || mutation.Key == "spring_id" || mutation.Key == "current_length" {
			t.Fatalf("mutation should be filtered: %#v", mutation)
		}
	}
	if !isEquivalentSelectedObjectParameterMutation(0, "mass_id") || !isEquivalentSelectedObjectParameterMutation(1, "value") {
		t.Fatal("expected setup/assertion cells to be equivalent")
	}
	if isEquivalentSelectedObjectParameterMutation(3, "control") || isEquivalentSelectedObjectParameterMutation(3, "object_type") {
		t.Fatal("control and object type mutations should remain meaningful")
	}
}

func TestBuildMutationsSkipsEquivalentWallCollisionCells(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Wall collision and stickiness",
		Scenarios: []gherkin.Scenario{
			{Examples: []map[string]string{{"wall": "left", "mass_id": "1", "elasticity": "0.5"}}},
			{Examples: []map[string]string{{"wall": "right", "mass_id": "2"}}},
			{Examples: []map[string]string{{"stickiness": "high", "mass_id": "1", "wall": "left", "release_force": "sufficient", "release_result": "released"}}},
			{Examples: []map[string]string{{"wall": "bottom", "mass_id": "1"}}},
		},
	}

	mutations := BuildMutations(feature)
	for _, mutation := range mutations {
		if mutation.Key == "mass_id" || mutation.Key == "elasticity" {
			t.Fatalf("mutation should be filtered: %#v", mutation)
		}
	}
	if !isEquivalentWallCollisionMutation(0, "elasticity") || !isEquivalentWallCollisionMutation(3, "mass_id") {
		t.Fatal("expected setup/assertion wall cells to be equivalent")
	}
	if isEquivalentWallCollisionMutation(2, "release_result") || isEquivalentWallCollisionMutation(1, "wall") {
		t.Fatal("release result and wall mutations should remain meaningful")
	}
}

func TestEquivalentMutationPredicates(t *testing.T) {
	for _, check := range []struct {
		equivalent func(int, string) bool
		equiv      []mutationCell
		meaningful mutationCell
	}{
		{isEquivalentMouseEditingMutation, []mutationCell{{2, "mass_a"}, {3, "target_position"}}, mutationCell{0, "expected_position"}},
		{isEquivalentControlsHotkeysMutation, []mutationCell{{1, "initial_state"}, {3, "parameter"}}, mutationCell{0, "shortcut"}},
	} {
		for _, cell := range check.equiv {
			assertEquivalentMutation(t, check.equivalent, cell.scenario, cell.key)
		}
		assertMeaningfulMutation(t, check.equivalent, check.meaningful.scenario, check.meaningful.key)
	}
	assertMeaningfulMutation(t, isEquivalentControlsHotkeysMutation, 1, "command")
	assertMeaningfulMutation(t, isEquivalentControlsHotkeysMutation, 0, "parameter")
	assertMeaningfulMutation(t, isEquivalentControlsHotkeysMutation, 3, "parameter_result")
}

type mutationCell struct {
	scenario int
	key      string
}

func assertEquivalentMutation(t *testing.T, equivalent func(int, string) bool, scenario int, key string) {
	t.Helper()
	if !equivalent(scenario, key) {
		t.Fatalf("expected %s in scenario %d to be equivalent", key, scenario)
	}
}

func assertMeaningfulMutation(t *testing.T, equivalent func(int, string) bool, scenario int, key string) {
	t.Helper()
	if equivalent(scenario, key) {
		t.Fatalf("expected %s in scenario %d to be meaningful", key, scenario)
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

func TestRunMutationsReturnsWorkDirCreationError(t *testing.T) {
	workDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(workDir, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RunMutations(gherkin.Feature{}, workDir); err == nil {
		t.Fatal("expected work dir creation error")
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

func TestMutationScalarHelpersUseExpectedMutations(t *testing.T) {
	if got, ok := mutateKeyword("false", "false", rand.New(rand.NewSource(1))); !ok || got != "true" {
		t.Fatalf("false keyword mutation = %q, %t", got, ok)
	}
	if got, ok := mutateKeyword("nil", "nil", rand.New(rand.NewSource(1))); !ok || got == "nil" || got == "" {
		t.Fatalf("nil keyword mutation = %q, %t", got, ok)
	}
	if got, ok := mutateKeyword("other", "other", rand.New(rand.NewSource(1))); ok || got != "" {
		t.Fatalf("unsupported keyword mutation = %q, %t", got, ok)
	}

	if got, ok := mutateNumber("12", rand.New(rand.NewSource(1))); !ok || got != "18" {
		t.Fatalf("integer mutation = %q, %t", got, ok)
	}
	if got, ok := mutateNumber("10.5", rand.New(rand.NewSource(1))); !ok || got != "17.31" {
		t.Fatalf("float mutation = %q, %t", got, ok)
	}
	if got, ok := mutateNumber("word", rand.New(rand.NewSource(1))); ok || got != "" {
		t.Fatalf("unsupported number mutation = %q, %t", got, ok)
	}
	if got, ok := mutateNumber("NaN", rand.New(rand.NewSource(1))); ok || got != "" {
		t.Fatalf("non-decimal float mutation = %q, %t", got, ok)
	}

	if got := signedIntDelta(rand.New(rand.NewSource(1))); got != 6 {
		t.Fatalf("positive int delta = %d", got)
	}
	if got := signedIntDelta(rand.New(rand.NewSource(2))); got != -8 {
		t.Fatalf("negative int delta = %d", got)
	}
	if got := signedFloatDelta(rand.New(rand.NewSource(1))); got != 6.81 {
		t.Fatalf("positive float delta = %v", got)
	}
	if got := signedFloatDelta(rand.New(rand.NewSource(2))); got != -3.86 {
		t.Fatalf("negative float delta = %v", got)
	}

	if got, ok := mutateDate("not-a-date", rand.New(rand.NewSource(1))); ok || got != "" {
		t.Fatalf("unsupported date mutation = %q, %t", got, ok)
	}
	if got, ok := mutateDate("2026-05-16", rand.New(rand.NewSource(1))); !ok || got != "2026-05-22" {
		t.Fatalf("date mutation = %q, %t", got, ok)
	}
	if got, ok := mutateDuration("not-a-duration", rand.New(rand.NewSource(1))); ok || got != "" {
		t.Fatalf("unsupported duration mutation = %q, %t", got, ok)
	}
	if got, ok := mutateDuration("2s", rand.New(rand.NewSource(1))); !ok || got != "8s" {
		t.Fatalf("duration mutation = %q, %t", got, ok)
	}

	if got := dither("", rand.New(rand.NewSource(1))); got != "x" {
		t.Fatalf("empty dither = %q", got)
	}
	if got := dither("x", rand.New(rand.NewSource(1))); got != "y" {
		t.Fatalf("x dither = %q", got)
	}
}
