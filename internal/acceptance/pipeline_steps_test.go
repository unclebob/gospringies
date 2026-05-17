package acceptance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"springs/internal/mutationstamp"
)

func TestFeatureMutationStampStepSetsAndAssertsState(t *testing.T) {
	feature := testBuildFeature(t, "Feature: Stamp\n")
	example := map[string]string{"feature_file": feature, "stamp_state": "stamped"}

	if err := setFeatureMutationStampState(&world{}, example); err != nil {
		t.Fatal(err)
	}
	if err := assertFeatureMutationStampState(&world{}, map[string]string{
		"feature_file":         feature,
		"expected_stamp_state": "stamped",
	}); err != nil {
		t.Fatal(err)
	}

	example["stamp_state"] = "unstamped"
	if err := setFeatureMutationStampState(&world{}, example); err != nil {
		t.Fatal(err)
	}
	if err := assertFeatureMutationStampState(&world{}, map[string]string{
		"feature_file":         feature,
		"expected_stamp_state": "unstamped",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestFeatureMutationStampStepsRejectUnsupportedStates(t *testing.T) {
	if err := setFeatureMutationStampState(&world{}, map[string]string{
		"feature_file": "build/_acceptance-pipeline/test.feature",
		"stamp_state":  "unknown",
	}); err == nil {
		t.Fatal("expected unsupported setter state")
	}
	if err := assertFeatureMutationStampState(&world{}, map[string]string{
		"feature_file":         "build/_acceptance-pipeline/test.feature",
		"expected_stamp_state": "unknown",
	}); err == nil {
		t.Fatal("expected unsupported assertion state")
	}
}

func TestAcceptanceMutationBehaviorAssertions(t *testing.T) {
	runWorld := &world{mutationOutput: "total=1 killed=1 survived=0 errors=0\n"}
	if err := assertAcceptanceMutationBehavior(runWorld, map[string]string{
		"mutation_behavior": "run and stamp",
		"feature_file":      "features/pipeline_smoke.feature",
	}); err != nil {
		t.Fatal(err)
	}

	skipWorld := &world{mutationOutput: "mutation stamp valid; skipping features/pipeline_smoke.feature\n"}
	if err := assertAcceptanceMutationBehavior(skipWorld, map[string]string{
		"mutation_behavior": "skip",
		"feature_file":      "features/pipeline_smoke.feature",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAcceptanceMutationBehaviorAssertionsReportMismatches(t *testing.T) {
	if err := assertAcceptanceMutationBehavior(&world{mutationOutput: "total=1\n"}, map[string]string{
		"mutation_behavior": "skip",
		"feature_file":      "features/pipeline_smoke.feature",
	}); err == nil {
		t.Fatal("expected skip mismatch")
	}
	if err := assertAcceptanceMutationBehavior(&world{mutationOutput: "mutation stamp valid; skipping x\n"}, map[string]string{
		"mutation_behavior": "run and stamp",
		"feature_file":      "features/pipeline_smoke.feature",
	}); err == nil {
		t.Fatal("expected run mismatch")
	}
	if err := assertAcceptanceMutationBehavior(&world{}, map[string]string{
		"mutation_behavior": "unknown",
		"feature_file":      "features/pipeline_smoke.feature",
	}); err == nil {
		t.Fatal("expected unsupported behavior")
	}
}

func TestRunAcceptanceMutationForFeatureCapturesOutput(t *testing.T) {
	original := runAcceptanceMutation
	defer func() { runAcceptanceMutation = original }()
	runAcceptanceMutation = func(feature string) (string, error) {
		if feature != "features/pipeline_smoke.feature" {
			t.Fatalf("feature = %q", feature)
		}
		return "mutation output", nil
	}

	w := &world{}
	err := runAcceptanceMutationForFeature(w, map[string]string{"feature_file": "features/pipeline_smoke.feature"})
	if err != nil {
		t.Fatal(err)
	}
	if w.mutationOutput != "mutation output" {
		t.Fatalf("mutationOutput = %q", w.mutationOutput)
	}
}

func TestRunAcceptanceMutationForFeatureReturnsRunnerError(t *testing.T) {
	original := runAcceptanceMutation
	defer func() { runAcceptanceMutation = original }()
	runAcceptanceMutation = func(string) (string, error) {
		return "failure output", errors.New("failed")
	}

	w := &world{}
	err := runAcceptanceMutationForFeature(w, map[string]string{"feature_file": "features/pipeline_smoke.feature"})
	if err == nil || !strings.Contains(err.Error(), "failed") {
		t.Fatalf("expected runner error, got %v", err)
	}
	if w.mutationOutput != "failure output" {
		t.Fatalf("mutationOutput = %q", w.mutationOutput)
	}
}

func TestRunAcceptanceMutationCommandStampsFeature(t *testing.T) {
	feature := testBuildFeature(t, "Feature: Empty\n\nScenario: no examples\n  Given nothing\n")

	output, err := runAcceptanceMutationCommand(feature)
	if err != nil {
		t.Fatalf("runAcceptanceMutationCommand returned error: %v\n%s", err, output)
	}
	if !strings.Contains(output, "total=0") {
		t.Fatalf("output = %s", output)
	}
	if !mutationstamp.Valid(repoPath(feature)) {
		t.Fatal("feature should be stamped after successful mutation command")
	}
}

func TestRunAcceptanceMutationCommandReportsFailure(t *testing.T) {
	output, err := runAcceptanceMutationCommand("build/_acceptance-pipeline/unit/missing.feature")
	if err == nil {
		t.Fatal("expected missing feature error")
	}
	if !strings.Contains(err.Error(), "acceptance mutation failed") {
		t.Fatalf("error = %v", err)
	}
	if output == "" {
		t.Fatal("expected command output")
	}
}

func testBuildFeature(t *testing.T, content string) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("build", "_acceptance-pipeline", "unit", t.Name()+".feature")
	absolute := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolute, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(absolute) })
	if mutationstamp.Valid(absolute) {
		t.Fatal("test feature unexpectedly started stamped")
	}
	return path
}
