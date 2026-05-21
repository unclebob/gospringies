package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"springs/internal/acceptancemutation"
	"springs/internal/gherkin"
	"springs/internal/mutationstamp"
)

func TestPrintTextIncludesSurvivorDetails(t *testing.T) {
	var output bytes.Buffer
	printText(
		&output,
		acceptancemutation.MutationSummary{Total: 1, Survived: 1},
		[]acceptancemutation.MutationResult{{
			Status: acceptancemutation.MutationSurvived,
			Mutation: acceptancemutation.Mutation{
				Description: "$.path: old -> new",
			},
			Error:  "boom",
			Output: "details\n",
		}},
	)

	for _, fragment := range []string{"total=1 killed=0 survived=1 errors=0", "survived $.path: old -> new", "error: boom", "output:\ndetails"} {
		if !strings.Contains(output.String(), fragment) {
			t.Fatalf("output missing %q:\n%s", fragment, output.String())
		}
	}
}

func TestPrintTextOmitsKilledDetails(t *testing.T) {
	var output bytes.Buffer
	printText(
		&output,
		acceptancemutation.MutationSummary{Total: 1, Killed: 1},
		[]acceptancemutation.MutationResult{{
			Status: acceptancemutation.MutationKilled,
			Mutation: acceptancemutation.Mutation{
				Description: "$.path: old -> new",
			},
			Output: "hidden\n",
		}},
	)

	if strings.Contains(output.String(), "hidden") {
		t.Fatalf("killed output should be hidden:\n%s", output.String())
	}
}

func TestPrintJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	printJSON(&stdout, &stderr, acceptancemutation.MutationSummary{Total: 1}, nil)

	if !strings.Contains(stdout.String(), `"Total": 1`) {
		t.Fatalf("json output = %s", stdout.String())
	}
}

func TestPrintProgressReportsCounts(t *testing.T) {
	var stderr bytes.Buffer
	printProgress(&stderr)(acceptancemutation.MutationProgress{Completed: 20, Total: 39, Killed: 19, Survived: 1})

	if !strings.Contains(stderr.String(), "progress completed=20 total=39 killed=19 survived=1 errors=0") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestProgressWriterUsesStdoutForTextAndStderrForJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if progressWriter(options{}, &stdout, &stderr) != &stdout {
		t.Fatal("expected text progress to use stdout")
	}
	if progressWriter(options{jsonReport: true}, &stdout, &stderr) != &stderr {
		t.Fatal("expected JSON progress to use stderr")
	}
}

func TestParseOptionsAcceptsWorkers(t *testing.T) {
	var stderr bytes.Buffer
	options, err := parseOptions([]string{
		"-feature", "feature.feature",
		"-workers", "4",
		"-timeout", "3m",
		"-mutant-timeout", "5s",
	}, &stderr)
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}
	if options.workers != 4 {
		t.Fatalf("workers = %d, want 4", options.workers)
	}
	if options.timeout != 3*time.Minute {
		t.Fatalf("timeout = %v, want 3m", options.timeout)
	}
	if options.mutantTimeout != 5*time.Second {
		t.Fatalf("mutant timeout = %v, want 5s", options.mutantTimeout)
	}
}

func TestParseOptionsDefaults(t *testing.T) {
	var stderr bytes.Buffer
	options, err := parseOptions(nil, &stderr)
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}
	if options.workers <= 0 {
		t.Fatalf("workers = %d, want positive default", options.workers)
	}
	if options.timeout != 0 {
		t.Fatalf("timeout = %v, want 0", options.timeout)
	}
	if options.mutantTimeout != 30*time.Second {
		t.Fatalf("mutant timeout = %v, want 30s", options.mutantTimeout)
	}
}

func TestMutationContextUsesDeadlineOnlyForPositiveTimeout(t *testing.T) {
	zeroCtx, zeroCancel := mutationContext(0)
	defer zeroCancel()
	if _, ok := zeroCtx.Deadline(); ok {
		t.Fatal("zero timeout should not set a deadline")
	}
	oneNanoCtx, oneNanoCancel := mutationContext(time.Nanosecond)
	defer oneNanoCancel()
	if _, ok := oneNanoCtx.Deadline(); !ok {
		t.Fatal("one nanosecond timeout should set a deadline")
	}
	positiveCtx, positiveCancel := mutationContext(time.Second)
	defer positiveCancel()
	if _, ok := positiveCtx.Deadline(); !ok {
		t.Fatal("positive timeout should set a deadline")
	}
}

func TestRunFeatureMutationsReturnsReadError(t *testing.T) {
	_, _, _, _, _, err := runFeatureMutations(options{featurePath: "missing.feature"}, "impl", io.Discard)
	if err == nil {
		t.Fatal("expected missing feature error")
	}
}

func TestRunReturnsFailureForMissingFeature(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-feature", filepath.Join(t.TempDir(), "missing.feature")}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "missing.feature") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestRunFeatureMutationsReturnsWorkDirError(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "empty.feature")
	workDir := filepath.Join(dir, "not-a-directory")
	writeFile(t, featurePath, "Feature: Empty\n\nScenario: no examples\n  Given nothing\n")
	writeFile(t, workDir, "x")

	_, _, _, _, _, err := runFeatureMutations(options{featurePath: featurePath, workDir: workDir, workers: 1}, "impl", io.Discard)
	if err == nil {
		t.Fatal("expected work dir error")
	}
}

func TestFinishRunReportsStampError(t *testing.T) {
	var stderr bytes.Buffer
	code := finishRun(filepath.Join(t.TempDir(), "missing.feature"), acceptancemutation.MutationSummary{}, acceptancemutation.ScenarioManifest{}, acceptancemutation.ScenarioSkipPlan{}, nil, gherkin.Feature{}, "impl", &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "missing.feature") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestExitCodeFromSummary(t *testing.T) {
	for _, tt := range []struct {
		name    string
		summary acceptancemutation.MutationSummary
		want    int
	}{
		{name: "all killed", summary: acceptancemutation.MutationSummary{Killed: 1}, want: 0},
		{name: "survivor", summary: acceptancemutation.MutationSummary{Survived: 1}, want: 1},
		{name: "error", summary: acceptancemutation.MutationSummary{Errors: 1}, want: 1},
		{name: "survivor and error", summary: acceptancemutation.MutationSummary{Survived: 1, Errors: 1}, want: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCodeFromSummary(tt.summary); got != tt.want {
				t.Fatalf("exit code = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMainExitsForInvalidArguments(t *testing.T) {
	if os.Getenv("SPRINGS_GHERKIN_MUTATOR_MAIN_TEST") == "1" {
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitsForInvalidArguments")
	cmd.Env = append(os.Environ(), "SPRINGS_GHERKIN_MUTATOR_MAIN_TEST=1")
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.ExitCode())
	}
}

func TestRunReturnsSuccessForFeatureWithoutMutations(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "empty.feature")
	writeFile(t, featurePath, "Feature: Empty\n\nScenario: no examples\n  Given nothing\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-feature", featurePath, "-work-dir", filepath.Join(dir, "mutations"), "-json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"Total": 0`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunStampsFeatureAfterSuccessfulMutation(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "empty.feature")
	writeFile(t, featurePath, "Feature: Empty\n\nScenario: no examples\n  Given nothing\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-feature", featurePath, "-work-dir", filepath.Join(dir, "mutations")}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	content := readFile(t, featurePath)
	if !strings.Contains(content, mutationstamp.Prefix) {
		t.Fatalf("feature was not stamped:\n%s", content)
	}
}

func TestRunSkipsStampedFeature(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "empty.feature")
	writeFile(t, featurePath, "Feature: Empty\n\nScenario: no examples\n  Given nothing\n")
	if err := mutationstamp.Stamp(featurePath); err != nil {
		t.Fatal(err)
	}
	before := readFile(t, featurePath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-feature", featurePath, "-work-dir", filepath.Join(dir, "mutations")}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mutation stamp valid; skipping") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if after := readFile(t, featurePath); after != before {
		t.Fatalf("stamped feature changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestRunSkipsScenariosFromManifest(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "manifest.feature")
	writeFile(t, featurePath, strings.Join([]string{
		"Feature: Manifest skip",
		"",
		"Scenario Outline: first",
		"  Then value <value>",
		"",
		"Examples:",
		"  | value |",
		"  | 1     |",
		"",
		"Scenario Outline: second",
		"  Then name <name>",
		"",
		"Examples:",
		"  | name |",
		"  | Ada  |",
		"",
	}, "\n"))
	feature, err := gherkin.ReadFile(featurePath)
	if err != nil {
		t.Fatal(err)
	}
	implementationHash, err := acceptancemutation.CurrentImplementationHash()
	if err != nil {
		t.Fatal(err)
	}
	manifest := acceptancemutation.BuildScenarioManifest(featurePath, feature, acceptancemutation.ScenarioManifest{}, acceptancemutation.ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, []acceptancemutation.MutationResult{
		{Mutation: acceptancemutation.Mutation{Scenario: 0}, Status: acceptancemutation.MutationKilled},
		{Mutation: acceptancemutation.Mutation{Scenario: 1}, Status: acceptancemutation.MutationKilled},
	}, implementationHash, time.Unix(1, 0).UTC())
	if err := acceptancemutation.WriteScenarioManifestFile(featurePath, feature, manifest, acceptancemutation.ScenarioSkipPlan{SkipScenarios: map[int]bool{0: true, 1: true}}, nil, implementationHash, time.Unix(1, 0).UTC()); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-feature", featurePath, "-work-dir", filepath.Join(dir, "mutations")}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "skipped_scenarios=2") || !strings.Contains(stdout.String(), "scenario manifest valid") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if !strings.Contains(readFile(t, featurePath), mutationstamp.Prefix) {
		t.Fatal("feature was not stamped after manifest skip")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
