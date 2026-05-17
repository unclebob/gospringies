package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"springs/internal/mutationstamp"
)

func runAcceptanceCommand(w *world, _ map[string]string) error {
	if err := runPipeline("features/pipeline_smoke.feature", "pipeline_command"); err != nil {
		return err
	}
	w.parserRan = true
	w.generatorRan = true
	w.generatedRan = true
	return nil
}

func assertParserRan(w *world, _ map[string]string) error {
	return requirePrerequisite(w.parserRan, "gherkin parser did not run successfully")
}

func assertGeneratorRan(w *world, _ map[string]string) error {
	return requirePrerequisite(w.generatorRan, "acceptance generator did not run successfully")
}

func assertGeneratedRan(w *world, _ map[string]string) error {
	return requirePrerequisite(w.generatedRan, "generated executable acceptance tests did not run successfully")
}

func generateAcceptanceArtifacts(w *world, _ map[string]string) error {
	if err := runParserAndGenerator(
		"features/pipeline_smoke.feature",
		"build/acceptance/pipeline_artifacts.json",
		"acceptance/generated/pipeline_artifacts_acceptance_test.go",
	); err != nil {
		return err
	}
	w.generated = true
	return nil
}

func assertGeneratedArtifactExists(w *world, example map[string]string) error {
	if err := requirePrerequisite(w.generated, "acceptance tests have not been generated"); err != nil {
		return err
	}
	artifact, location, err := artifactExample(example)
	if err != nil {
		return err
	}
	return generatedArtifactExists(artifact, location)
}

func artifactExample(example map[string]string) (string, string, error) {
	return stringPair(example, "artifact", "generated_location")
}

func assertHandwrittenTestsOutside(_ *world, example map[string]string) error {
	testType, err := stringValue(example, "test_type")
	if err != nil {
		return err
	}
	if strings.TrimSpace(testType) != "unit" {
		return fmt.Errorf("unsupported hand-written test type %q", testType)
	}
	location, err := stringValue(example, "generated_location")
	if err != nil {
		return err
	}
	return handwrittenTestsOutside(location)
}

func addSmokeFeature(w *world, _ map[string]string) error {
	if _, err := os.Stat(repoPath("features/pipeline_smoke.feature")); err != nil {
		return err
	}
	w.smokeAdded = true
	return nil
}

func parseSmokeFeature(w *world, _ map[string]string) error {
	return runSmokeStage(w.smokeAdded, "smoke feature has not been added", &w.smokeParsed, parseSmoke)
}

func generateSmokeAcceptanceTest(w *world, _ map[string]string) error {
	return runSmokeStage(w.smokeParsed, "smoke feature has not been parsed", &w.smokeGenerated, generateSmoke)
}

func parseSmoke() error {
	return runParser("features/pipeline_smoke.feature", "build/_acceptance-pipeline/smoke/feature.json")
}

func generateSmoke() error {
	return runGenerator("build/_acceptance-pipeline/smoke/feature.json", "build/_acceptance-pipeline/smoke/generated/pipeline_smoke_acceptance_test.go")
}

func runSmokeStage(ready bool, message string, done *bool, action func() error) error {
	if err := requirePrerequisite(ready, message); err != nil {
		return err
	}
	if err := action(); err != nil {
		return err
	}
	*done = true
	return nil
}

func assertSmokeAcceptanceTestPasses(w *world, _ map[string]string) error {
	if err := requirePrerequisite(w.smokeGenerated, "smoke acceptance test has not been generated"); err != nil {
		return err
	}
	return runCommand("go", "test", "./build/_acceptance-pipeline/smoke/generated")
}

func assertAcceptanceCommandPassesFromCleanCheckout(*world, map[string]string) error {
	return runCommandWithEnv([]string{
		"ACCEPTANCE_BUILD_DIR=build/_acceptance-pipeline/clean",
		"ACCEPTANCE_GENERATED_DIR=build/_acceptance-pipeline/clean/generated",
	}, "./scripts/acceptance.sh", "features/pipeline_smoke.feature")
}

func setFeatureMutationStampState(_ *world, example map[string]string) error {
	return applyFeatureMutationStampState(example, "stamp_state", mutationStampStateSetters)
}

func runAcceptanceMutationForFeature(w *world, example map[string]string) error {
	feature, err := stringValue(example, "feature_file")
	if err != nil {
		return err
	}
	output, err := runAcceptanceMutation(feature)
	w.mutationOutput = output
	return err
}

var runAcceptanceMutation = runAcceptanceMutationCommand

func runAcceptanceMutationCommand(feature string) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("go", "run", "./cmd/gherkin-mutator", "--feature", feature, "--work-dir", "build/_acceptance-pipeline/mutation-stamps")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("acceptance mutation failed: %w\n%s", err, output)
	}
	return string(output), nil
}

func assertAcceptanceMutationBehavior(w *world, example map[string]string) error {
	behavior, feature, err := stringPair(example, "mutation_behavior", "feature_file")
	if err != nil {
		return err
	}
	assertion, ok := mutationBehaviorAssertions[behavior]
	if !ok {
		return fmt.Errorf("unsupported mutation behavior %q", behavior)
	}
	return assertion(w.mutationOutput, feature)
}

func assertFeatureMutationStampState(_ *world, example map[string]string) error {
	return applyFeatureMutationStampState(example, "expected_stamp_state", mutationStampStateAssertions)
}

func applyFeatureMutationStampState(example map[string]string, stateKey string, actions map[string]func(string) error) error {
	feature, state, err := stringPair(example, "feature_file", stateKey)
	if err != nil {
		return err
	}
	action, ok := actions[state]
	if !ok {
		return fmt.Errorf("unsupported mutation stamp state %q", state)
	}
	return action(repoPath(feature))
}

var mutationStampStateSetters = map[string]func(string) error{
	"unstamped": mutationstamp.Remove,
	"stamped":   writeCurrentFeatureMutationStamp,
}

func writeCurrentFeatureMutationStamp(path string) error {
	if err := mutationstamp.Remove(path); err != nil {
		return err
	}
	return mutationstamp.Stamp(path)
}

var mutationBehaviorAssertions = map[string]func(string, string) error{
	"run and stamp": assertMutationRanAndStamped,
	"skip":          assertMutationSkipped,
}

func assertMutationRanAndStamped(output, feature string) error {
	if strings.Contains(output, "mutation stamp valid; skipping") {
		return fmt.Errorf("acceptance mutation skipped %s", feature)
	}
	if !strings.Contains(output, "total=") {
		return fmt.Errorf("acceptance mutation did not report run for %s:\n%s", feature, output)
	}
	return nil
}

func assertMutationSkipped(output, feature string) error {
	if !strings.Contains(output, "mutation stamp valid; skipping "+feature) {
		return fmt.Errorf("acceptance mutation did not skip %s:\n%s", feature, output)
	}
	return nil
}

var mutationStampStateAssertions = map[string]func(string) error{
	"stamped":   assertFeatureStamped,
	"unstamped": assertFeatureUnstamped,
}

func assertFeatureStamped(path string) error {
	return requirePrerequisite(mutationstamp.Valid(path), "feature file is not stamped")
}

func assertFeatureUnstamped(path string) error {
	if mutationstamp.Valid(path) {
		return fmt.Errorf("feature file is stamped")
	}
	return nil
}

func generatedArtifactExists(artifact, location string) error {
	path, err := generatedArtifactPath(artifact, location)
	if err != nil {
		return err
	}
	return fileExists(path)
}

func generatedArtifactPath(artifact, location string) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	name, ok := generatedArtifactNames[artifact]
	if !ok {
		return "", fmt.Errorf("unsupported generated artifact %q", artifact)
	}
	return filepath.Join(root, location, name), nil
}

var generatedArtifactNames = map[string]string{
	"test source":    "pipeline_artifacts_acceptance_test.go",
	"parsed feature": "pipeline_artifacts.json",
}

func handwrittenTestsOutside(location string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	generatedLocation := filepath.Clean(filepath.Join(root, location))
	var violations []string
	for _, dir := range []string{"internal", "cmd"} {
		dirViolations, err := handwrittenTestViolations(filepath.Join(root, dir), generatedLocation)
		if err != nil {
			return err
		}
		violations = append(violations, dirViolations...)
	}
	return reportHandwrittenViolations(violations)
}

func handwrittenTestViolations(root, generatedLocation string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if isHandwrittenTestUnder(path, entry, generatedLocation) {
			violations = append(violations, path)
		}
		return nil
	})
	return violations, err
}

func isHandwrittenTestUnder(path string, entry os.DirEntry, generatedLocation string) bool {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
		return false
	}
	return strings.HasPrefix(filepath.Clean(path), generatedLocation)
}

func reportHandwrittenViolations(violations []string) error {
	if len(violations) > 0 {
		return fmt.Errorf("hand-written tests under generated location: %s", strings.Join(violations, ", "))
	}
	return nil
}
