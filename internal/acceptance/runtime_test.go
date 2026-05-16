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
	if err := lookupMass(&world{domainWorld: sim.NewWorld()}, map[string]string{"id": "7"}); err == nil {
		t.Fatal("expected missing mass error")
	}
	if err := lookupSpring(&world{domainWorld: sim.NewWorld()}, map[string]string{"spring_id": "7"}); err == nil {
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
		createDomainMassFromID,
		setMassVelocity,
		setMassValue,
		setMassElasticity,
		setMassFixed,
		lookupMass,
		assertMassPosition,
		assertMassVelocity,
		assertMassValue,
		assertMassElasticity,
		assertMassFixed,
	} {
		if err := fn(w, massExample); err != nil {
			t.Fatalf("mass handler returned error: %v", err)
		}
	}
}

func TestDomainSpringHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := createDomainMassA(w, map[string]string{"mass_a": "1", "x_a": "0", "y_a": "0"}); err != nil {
		t.Fatal(err)
	}
	if err := createDomainMassB(w, map[string]string{"mass_b": "2", "x_b": "10", "y_b": "0"}); err != nil {
		t.Fatal(err)
	}
	springExample := map[string]string{
		"spring_id": "7", "mass_a": "1", "mass_b": "2",
		"spring_constant": "12.5", "damping_constant": "0.7", "rest_length": "10",
	}
	for _, fn := range []stepHandler{
		createDomainSpring,
		setSpringConstant,
		setSpringDamping,
		setSpringRestLength,
		lookupSpring,
		assertSpringEndpoints,
		assertSpringConstant,
		assertSpringDamping,
		assertSpringRestLength,
	} {
		if err := fn(w, springExample); err != nil {
			t.Fatalf("spring handler returned error: %v", err)
		}
	}
}

func TestDomainValidationHandlers(t *testing.T) {
	duplicateMass := map[string]string{"object_type": "mass", "id": "1", "reason": "duplicate id"}
	w := &world{}
	if err := createDuplicateSubject(w, duplicateMass); err != nil {
		t.Fatal(err)
	}
	if err := addDuplicateSubject(w, duplicateMass); err != nil {
		t.Fatal(err)
	}
	if err := assertValidationFailure(w, duplicateMass); err != nil {
		t.Fatal(err)
	}

	duplicateSpring := map[string]string{"object_type": "spring", "id": "5", "reason": "duplicate id"}
	w = &world{}
	if err := createDuplicateSubject(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}
	if err := addDuplicateSubject(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}
	if err := assertValidationFailure(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}

	missingEndpoint := map[string]string{"existing_mass": "1", "x": "0", "y": "0", "spring_id": "2", "mass_a": "1", "mass_b": "9", "reason": "missing spring endpoint"}
	w = &world{}
	if err := createExistingDomainMass(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
	if err := addDomainSpringForValidation(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
	if err := assertValidationFailure(w, missingEndpoint); err != nil {
		t.Fatal(err)
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
