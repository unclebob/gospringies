package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"springs/internal/acceptance"
)

func TestPrintTextIncludesSurvivorDetails(t *testing.T) {
	var output bytes.Buffer
	printText(
		&output,
		acceptance.MutationSummary{Total: 1, Survived: 1},
		[]acceptance.MutationResult{{
			Status: "survived",
			Mutation: acceptance.Mutation{
				Description: "$.path: old -> new",
			},
			Output: "details\n",
		}},
	)

	for _, fragment := range []string{"total=1 killed=0 survived=1 errors=0", "survived $.path: old -> new", "details"} {
		if !strings.Contains(output.String(), fragment) {
			t.Fatalf("output missing %q:\n%s", fragment, output.String())
		}
	}
}

func TestPrintTextOmitsKilledDetails(t *testing.T) {
	var output bytes.Buffer
	printText(
		&output,
		acceptance.MutationSummary{Total: 1, Killed: 1},
		[]acceptance.MutationResult{{
			Status: "killed",
			Mutation: acceptance.Mutation{
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
	printJSON(&stdout, &stderr, acceptance.MutationSummary{Total: 1}, nil)

	if !strings.Contains(stdout.String(), `"Total": 1`) {
		t.Fatalf("json output = %s", stdout.String())
	}
}

func TestPrintProgressReportsCounts(t *testing.T) {
	var stderr bytes.Buffer
	printProgress(&stderr)(acceptance.MutationProgress{Completed: 20, Total: 39, Killed: 19, Survived: 1})

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
	options, err := parseOptions([]string{"-feature", "feature.feature", "-workers", "4"}, &stderr)
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}
	if options.workers != 4 {
		t.Fatalf("workers = %d, want 4", options.workers)
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
	if !strings.Contains(content, mutationStampPrefix) {
		t.Fatalf("feature was not stamped:\n%s", content)
	}
}

func TestRunSkipsStampedFeature(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "empty.feature")
	writeFile(t, featurePath, "Feature: Empty\n\nScenario: no examples\n  Given nothing\n")
	if err := stampMutationFeature(featurePath); err != nil {
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
