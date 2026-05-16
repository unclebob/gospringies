package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"springs/internal/gherkin"
	"springs/internal/sim"
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

func TestRunFeatureExecutesSimulationSteps(t *testing.T) {
	feature := gherkin.Feature{Scenarios: []gherkin.Scenario{{
		Name: "simulation",
		Steps: []gherkin.Step{
			{Keyword: "Given", Text: "a demo spring simulation"},
			{Keyword: "When", Text: "I advance the simulation <steps> steps", Parameters: []string{"steps"}},
			{Keyword: "Then", Text: "mass <mass> x should be <x>", Parameters: []string{"mass", "x"}},
		},
		Examples: []map[string]string{{"steps": "0", "mass": "0", "x": "160"}},
	}}}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureExecutesDomainModelFeature(t *testing.T) {
	feature, err := gherkin.ReadFile(repoPath("features/003_domain_model.feature"))
	if err != nil {
		t.Fatal(err)
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestStepPrerequisitesReturnHelpfulErrors(t *testing.T) {
	cases := []gherkin.Step{
		{Text: "the <package> package should not import <graphics_library>"},
		{Text: "the application command should build successfully"},
		{Text: "the Go test suite should pass"},
		{Text: "I advance the simulation <steps> steps"},
		{Text: "mass <mass> x should be <x>"},
		{Text: "the Gherkin parser should run successfully"},
		{Text: "the acceptance test generator should run successfully"},
		{Text: "the generated executable acceptance tests should run successfully"},
		{Text: "generated acceptance <artifact> should be written under <generated_location>"},
		{Text: "the smoke feature should parse successfully"},
		{Text: "the smoke feature should generate an executable acceptance test"},
		{Text: "the generated smoke acceptance test should pass"},
	}
	example := map[string]string{
		"package":            "simulation",
		"graphics_library":   "Ebitengine",
		"steps":              "1",
		"mass":               "0",
		"x":                  "160",
		"artifact":           "test source",
		"generated_location": "acceptance/generated",
	}

	for _, step := range cases {
		if err := runStep(&world{}, step, example); err == nil {
			t.Fatalf("expected prerequisite error for %q", step.Text)
		}
	}
}

func TestExampleValueParsingReportsMissingAndInvalidValues(t *testing.T) {
	if _, err := stringValue(nil, "missing"); err == nil {
		t.Fatal("expected missing string error")
	}
	if _, err := intValue(map[string]string{"value": "NaN"}, "value"); err == nil {
		t.Fatal("expected invalid integer error")
	}
	if _, err := floatValue(map[string]string{"value": "NaN?"}, "value"); err == nil {
		t.Fatal("expected invalid float error")
	}
	if _, err := boolValue(map[string]string{"value": "maybe"}, "value"); err == nil {
		t.Fatal("expected invalid bool error")
	}
}

func TestDomainModelHelpersReportFailures(t *testing.T) {
	w := &world{}
	if _, err := domainWorld(w); err == nil {
		t.Fatal("expected missing domain world error")
	}
	if err := lookupDomainMass(&world{domainWorld: sim.NewWorld()}, map[string]string{"id": "7"}); err == nil {
		t.Fatal("expected missing mass error")
	}
	if err := lookupDomainSpring(&world{domainWorld: sim.NewWorld()}, map[string]string{"spring_id": "7"}); err == nil {
		t.Fatal("expected missing spring error")
	}
	if err := assertValidationReason(nil, "duplicate id"); err == nil {
		t.Fatal("expected validation success error")
	}
	if err := assertValidationReason(sim.ErrDuplicateID, "missing spring endpoint"); err == nil {
		t.Fatal("expected validation reason mismatch")
	}
	if _, _, _, err := massFields(map[string]string{}, "id", "x", "y"); err == nil {
		t.Fatal("expected missing mass fields error")
	}
}

func TestDomainModelHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := createDomainWorld(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainMassCount(w, map[string]string{"mass_count": "0"}); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainSpringCount(w, map[string]string{"spring_count": "0"}); err != nil {
		t.Fatal(err)
	}
	massExample := map[string]string{
		"id": "1", "x": "1.5", "y": "2.5",
		"vx": "3.5", "vy": "4.5",
		"mass_value": "5.5", "elasticity": "0.8", "fixed": "true",
	}
	for _, fn := range []stepHandler{
		addDomainMass,
		setDomainMassVelocity,
		setDomainMassValue,
		setDomainMassElasticity,
		setDomainMassFixed,
		lookupDomainMass,
		assertDomainMassPosition,
		assertDomainMassVelocity,
		assertDomainMassValue,
		assertDomainMassElasticity,
		assertDomainMassFixed,
	} {
		if err := fn(w, massExample); err != nil {
			t.Fatalf("mass handler returned error: %v", err)
		}
	}
}

func TestDomainSpringHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := addDomainMassA(w, map[string]string{"mass_a": "1", "x_a": "0", "y_a": "0"}); err != nil {
		t.Fatal(err)
	}
	if err := addDomainMassB(w, map[string]string{"mass_b": "2", "x_b": "10", "y_b": "0"}); err != nil {
		t.Fatal(err)
	}
	springExample := map[string]string{
		"spring_id": "7", "mass_a": "1", "mass_b": "2",
		"spring_constant": "12.5", "damping_constant": "0.7", "rest_length": "10",
	}
	for _, fn := range []stepHandler{
		addDomainSpring,
		setDomainSpringConstant,
		setDomainSpringDamping,
		setDomainSpringRestLength,
		lookupDomainSpring,
		assertDomainSpringEndpoints,
		assertDomainSpringConstant,
		assertDomainSpringDamping,
		assertDomainSpringRestLength,
	} {
		if err := fn(w, springExample); err != nil {
			t.Fatalf("spring handler returned error: %v", err)
		}
	}
}

func TestDomainValidationHandlers(t *testing.T) {
	duplicateMass := map[string]string{"object_type": "mass", "id": "1", "reason": "duplicate id"}
	w := &world{}
	if err := addExistingDomainObject(w, duplicateMass); err != nil {
		t.Fatal(err)
	}
	if err := addDuplicateDomainObject(w, duplicateMass); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainValidationReason(w, duplicateMass); err != nil {
		t.Fatal(err)
	}

	duplicateSpring := map[string]string{"object_type": "spring", "id": "5", "reason": "duplicate id"}
	w = &world{}
	if err := addExistingDomainObject(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}
	if err := addDuplicateDomainObject(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainValidationReason(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}

	missingEndpoint := map[string]string{"existing_mass": "1", "x": "0", "y": "0", "spring_id": "2", "mass_a": "1", "mass_b": "9", "reason": "missing spring endpoint"}
	w = &world{}
	if err := addExistingDomainMass(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
	if err := addInvalidDomainSpring(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainValidationReason(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
}

func TestRunFeatureExecutesSystemParameterFeature(t *testing.T) {
	feature, err := gherkin.ReadFile(repoPath("features/004_system_parameters.feature"))
	if err != nil {
		t.Fatal(err)
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureExecutesForceEvaluationFeature(t *testing.T) {
	feature, err := gherkin.ReadFile(repoPath("features/005_force_evaluation.feature"))
	if err != nil {
		t.Fatal(err)
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureExecutesSimulationStepFeature(t *testing.T) {
	feature, err := gherkin.ReadFile(repoPath("features/006_simulation_step.feature"))
	if err != nil {
		t.Fatal(err)
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestForceEvaluationAndSimulationStepHelperBranches(t *testing.T) {
	w := &world{}
	if err := createMovableMassAffectedByForce(w, map[string]string{"force": "center of mass attraction"}); err != nil {
		t.Fatal(err)
	}
	if len(w.domainWorld.Masses) != 2 {
		t.Fatalf("masses = %#v", w.domainWorld.Masses)
	}
	if err := createMassStartPosition(&world{}, map[string]string{"mass_id": "7", "start_position": "initial"}); err != nil {
		t.Fatal(err)
	}
	if _, err := durationValue(map[string]string{"duration": "forever"}, "duration"); err == nil {
		t.Fatal("expected unsupported duration")
	}
	if _, err := frameRateValue(map[string]string{"frame_rate": "120 fps"}); err == nil {
		t.Fatal("expected unsupported frame rate")
	}
}

func TestSimulationStepHelpersReportFailures(t *testing.T) {
	if err := createMassStartPosition(&world{}, map[string]string{"mass_id": "1", "start_position": "custom"}); err == nil {
		t.Fatal("expected unsupported position marker")
	}
	if err := assertResultDeterministic(nil, map[string]string{"initial_state": "unknown", "duration": "1 second"}); err == nil {
		t.Fatal("expected unsupported initial state")
	}
	w := &world{domainWorld: sim.NewWorld()}
	if err := advanceByDurationAtFrameRate(w, map[string]string{"duration": "1 second", "frame_rate": "bad"}); err == nil {
		t.Fatal("expected unsupported frame rate")
	}
	w.resultingWorld = sim.NewWorld()
	_ = w.resultingWorld.AddMass(sim.Mass{ID: 1, Velocity: sim.Vec2{X: 1}})
	if err := assertMassVelocityRemains(w, map[string]string{"mass_id": "1", "start_velocity": "zero"}); err == nil {
		t.Fatal("expected changed velocity error")
	}
}

func TestSameWorldStateDetectsDifferences(t *testing.T) {
	first := sim.NewWorld()
	second := sim.NewWorld()
	_ = first.AddMass(sim.Mass{ID: 1})
	if sameWorldState(first, second) {
		t.Fatal("expected length mismatch")
	}
	_ = second.AddMass(sim.Mass{ID: 1})
	second.Time = 1
	if sameWorldState(first, second) {
		t.Fatal("expected time mismatch")
	}
	second.Time = 0
	second.Masses[0].Position = sim.Vec2{X: 1}
	if sameWorldState(first, second) {
		t.Fatal("expected position mismatch")
	}
}

func TestSystemParameterHandlersReportFailures(t *testing.T) {
	w := &world{domainWorld: sim.NewWorld()}
	if err := assertParameterDefault(w, map[string]string{"parameter": "viscosity", "value": "unset"}); err == nil {
		t.Fatal("expected unsupported default marker")
	}
	if err := assertParameterDefault(w, map[string]string{"parameter": "missing", "value": "set"}); err == nil {
		t.Fatal("expected missing default parameter")
	}
	if err := assertForceEnabledState(w, map[string]string{"force": "missing", "enabled": "set"}); err == nil {
		t.Fatal("expected missing force")
	}
	if err := assertWallEnabledState(w, map[string]string{"wall": "missing", "enabled": "set"}); err == nil {
		t.Fatal("expected missing wall")
	}
	if err := performWorldOperation(w, map[string]string{"operation": "delete file"}); err == nil {
		t.Fatal("expected unsupported operation")
	}
	if _, err := expectedParameterValue("viscosity", "unknown source", nil); err == nil {
		t.Fatal("expected unsupported parameter source")
	}
}

func TestSystemParameterHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := changeWorldParameter(w, map[string]string{"parameter": "viscosity", "changed_value": "custom"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterDefault(w, map[string]string{"parameter": "viscosity", "value": "set"}); err != nil {
		t.Fatal(err)
	}
	if err := assertForceEnabledState(w, map[string]string{"force": "gravity", "enabled": "set"}); err != nil {
		t.Fatal(err)
	}
	if err := assertForceEditableParameters(w, map[string]string{"force": "gravity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertWallEnabledState(w, map[string]string{"wall": "top", "enabled": "set"}); err != nil {
		t.Fatal(err)
	}
	if err := performWorldOperation(w, map[string]string{"operation": "insert file", "parameter": "viscosity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterSource(w, map[string]string{
		"parameter": "viscosity", "changed_value": "custom", "expected_value_source": "existing world value",
	}); err != nil {
		t.Fatal(err)
	}
	if err := performWorldOperation(w, map[string]string{"operation": "load file", "parameter": "viscosity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterSource(w, map[string]string{"parameter": "viscosity", "expected_value_source": "value from loaded file"}); err != nil {
		t.Fatal(err)
	}
	if err := performWorldOperation(w, map[string]string{"operation": "reset", "parameter": "viscosity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterSource(w, map[string]string{"parameter": "viscosity", "expected_value_source": "default value"}); err != nil {
		t.Fatal(err)
	}
}

func TestForceEvaluationHandlerHelpers(t *testing.T) {
	springExample := map[string]string{
		"mass_a": "1", "mass_b": "2", "rest_length": "10", "spring_constant": "12",
		"velocity_a": "moving", "velocity_b": "still", "damping_constant": "0.5",
	}
	w := &world{}
	for _, fn := range []stepHandler{
		createSpringForceWorld,
		setOnlySpringRestLength,
		setOnlySpringConstant,
		setMassAVelocity,
		setMassBVelocity,
		setOnlySpringDamping,
		evaluateForces,
		assertSpringForcesEqualOpposite,
		assertSpringDampingDirection,
	} {
		if err := fn(w, springExample); err != nil {
			t.Fatalf("force handler returned error: %v", err)
		}
	}

	for _, force := range []string{"gravity", "viscosity", "wall repulsion", "center attraction", "center of mass attraction"} {
		w = &world{}
		example := map[string]string{"force": force}
		if err := enableEnvironmentalForce(w, example); err != nil {
			t.Fatal(err)
		}
		if err := createMovableMassAffectedByForce(w, example); err != nil {
			t.Fatal(err)
		}
		if err := evaluateForces(w, nil); err != nil {
			t.Fatal(err)
		}
		if err := assertMassReceivesForce(w, example); err != nil {
			t.Fatal(err)
		}
	}

	fixedExample := map[string]string{"mass_id": "1", "fixed": "true", "force": "gravity", "acceleration": "zero"}
	w = &world{}
	if err := createMassFixedState(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := affectMassByForce(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := evaluateForces(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertMassAcceleration(w, fixedExample); err != nil {
		t.Fatal(err)
	}

	for _, wall := range []string{"top", "left", "right", "bottom"} {
		w = &world{}
		example := map[string]string{"wall": wall, "mass_id": "1"}
		if err := enableWall(w, example); err != nil {
			t.Fatal(err)
		}
		if err := createMassOutsideWall(w, example); err != nil {
			t.Fatal(err)
		}
		if err := evaluateForces(w, nil); err != nil {
			t.Fatal(err)
		}
		if err := assertWallForceTowardInside(w, example); err != nil {
			t.Fatal(err)
		}
	}
}

func TestForceEvaluationHandlersReportFailures(t *testing.T) {
	if err := updateFirstSpring(&world{domainWorld: sim.NewWorld()}, nil); err == nil {
		t.Fatal("expected missing spring error")
	}
	if _, err := namedVelocity("fast"); err == nil {
		t.Fatal("expected unsupported velocity")
	}
	w := &world{domainWorld: sim.NewWorld()}
	if err := setMassNamedVelocity(w, map[string]string{"mass_a": "9", "velocity_a": "moving"}, "mass_a", "velocity_a"); err == nil {
		t.Fatal("expected missing mass error")
	}
	if err := affectMassByForce(&world{}, map[string]string{"force": "wind"}); err == nil {
		t.Fatal("expected unsupported force")
	}
	w = &world{forceEvaluation: sim.ForceEvaluation{ByMassID: map[int]sim.MassForces{1: {Force: sim.Vec2{}}}}}
	if err := assertMassReceivesForce(w, map[string]string{"force": "gravity"}); err == nil {
		t.Fatal("expected missing force assertion")
	}
	if err := assertMassAcceleration(&world{}, map[string]string{"mass_id": "1", "acceleration": "moving"}); err == nil {
		t.Fatal("expected unsupported acceleration expectation")
	}
}

func TestPackageDirDoesNotImportDetectsGraphicsLibrary(t *testing.T) {
	dir := t.TempDir()
	writeSource(t, filepath.Join(dir, "domain.go"), "package domain\n")
	if err := packageDirDoesNotImport(dir, "domain", "Ebitengine"); err != nil {
		t.Fatalf("packageDirDoesNotImport returned error: %v", err)
	}

	writeSource(t, filepath.Join(dir, "ui.go"), "package domain\nimport \"github.com/hajimehoshi/ebiten/v2\"\n")
	if err := packageDirDoesNotImport(dir, "domain", "Ebitengine"); err == nil {
		t.Fatal("expected graphics import error")
	}
}

func TestArtifactHelpers(t *testing.T) {
	if artifact, location, err := artifactExample(map[string]string{
		"artifact":           "test source",
		"generated_location": "acceptance/generated",
	}); err != nil || artifact != "test source" || location != "acceptance/generated" {
		t.Fatalf("artifactExample = %q, %q, %v", artifact, location, err)
	}
	if _, _, err := artifactExample(map[string]string{"artifact": "test source"}); err == nil {
		t.Fatal("expected missing generated location error")
	}
	if err := generatedArtifactExists("unsupported", "acceptance/generated"); err == nil {
		t.Fatal("expected unsupported artifact error")
	}
	path, err := generatedArtifactPath("test source", "acceptance/generated")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, filepath.Join("acceptance", "generated", "pipeline_artifacts_acceptance_test.go")) {
		t.Fatalf("generated artifact path = %s", path)
	}
}

func TestHandwrittenTestsOutside(t *testing.T) {
	if err := handwrittenTestsOutside("acceptance/generated"); err != nil {
		t.Fatalf("handwrittenTestsOutside returned error: %v", err)
	}
	if err := handwrittenTestsOutside("internal"); err == nil {
		t.Fatal("expected internal tests to violate generated location")
	}
}

func TestAssertHandwrittenTestsOutside(t *testing.T) {
	example := map[string]string{
		"test_type":          "unit",
		"generated_location": "acceptance/generated",
	}
	if err := assertHandwrittenTestsOutside(nil, example); err != nil {
		t.Fatalf("assertHandwrittenTestsOutside returned error: %v", err)
	}
	example["test_type"] = "integration"
	if err := assertHandwrittenTestsOutside(nil, example); err == nil {
		t.Fatal("expected unsupported test type error")
	}
}

func TestHandwrittenViolationHelpers(t *testing.T) {
	dir := t.TempDir()
	testPath := filepath.Join(dir, "example_test.go")
	writeSource(t, testPath, "package example\n")

	violations, err := handwrittenTestViolations(dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 || violations[0] != testPath {
		t.Fatalf("violations = %#v", violations)
	}
	if !isHandwrittenTestUnder(testPath, fakeDirEntry{name: "example_test.go"}, dir) {
		t.Fatal("expected test file under generated location")
	}
	if err := reportHandwrittenViolations(violations); err == nil {
		t.Fatal("expected violation report")
	}
}

func TestRunCommandInDirReportsFailures(t *testing.T) {
	if err := runCommandInDir(t.TempDir(), "go", "version"); err != nil {
		t.Fatalf("runCommandInDir returned error: %v", err)
	}
	if err := runCommandInDir(t.TempDir(), "go", "not-a-command"); err == nil {
		t.Fatal("expected command failure")
	}
}

func writeSource(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

type fakeDirEntry struct {
	name string
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return false }
func (f fakeDirEntry) Type() os.FileMode          { return 0 }
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }
