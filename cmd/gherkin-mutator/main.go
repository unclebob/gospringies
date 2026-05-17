package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"springs/internal/acceptance"
	"springs/internal/gherkin"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		return 2
	}
	if mutationStampValid(options.featurePath) {
		fmt.Fprintf(stdout, "mutation stamp valid; skipping %s\n", options.featurePath)
		return 0
	}
	summary, results, err := runFeatureMutations(options, progressWriter(options, stdout, stderr))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printReport(stdout, stderr, options.jsonReport, summary, results)
	code := exitCodeFromSummary(summary)
	if code == 0 {
		if err := stampMutationFeature(options.featurePath); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	return code
}

const mutationStampPrefix = "# mutation-stamp: "

type options struct {
	featurePath string
	workDir     string
	jsonReport  bool
	workers     int
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	flags := flag.NewFlagSet("gherkin-mutator", flag.ContinueOnError)
	flags.SetOutput(stderr)
	featurePath := flags.String("feature", "features/a-feature.feature", "Gherkin feature file to parse and mutate")
	workDir := flags.String("work-dir", "build/acceptance-mutation", "directory where mutation work files are written")
	jsonReport := flags.Bool("json", false, "emit JSON report")
	workers := flags.Int("workers", runtime.NumCPU(), "maximum mutation workers")
	_ = flags.Duration("timeout", 0, "full mutation timeout")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	return options{featurePath: *featurePath, workDir: *workDir, jsonReport: *jsonReport, workers: *workers}, nil
}

func runFeatureMutations(options options, progress io.Writer) (acceptance.MutationSummary, []acceptance.MutationResult, error) {
	feature, err := gherkin.ReadFile(options.featurePath)
	if err != nil {
		return acceptance.MutationSummary{}, nil, err
	}
	results, err := acceptance.RunMutationsWithOptions(feature, options.workDir, acceptance.RunMutationOptions{
		Workers:       options.workers,
		ProgressEvery: 20,
		Progress:      printProgress(progress),
	})
	if err != nil {
		return acceptance.MutationSummary{}, nil, err
	}
	return acceptance.Summarize(results), results, nil
}

func mutationStampValid(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	stamp, unstamped := splitMutationStamp(string(content))
	return stamp != "" && stamp == mutationContentHash(unstamped)
}

func stampMutationFeature(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, unstamped := splitMutationStamp(string(content))
	stamp := mutationContentHash(unstamped)
	return os.WriteFile(path, []byte(mutationStampPrefix+stamp+"\n"+unstamped), 0o644)
}

func splitMutationStamp(content string) (string, string) {
	lines := strings.SplitAfter(content, "\n")
	var unstamped strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r\n")
		if stamp, ok := strings.CutPrefix(trimmed, mutationStampPrefix); ok {
			return stamp, contentWithoutStampLine(lines, line)
		}
		unstamped.WriteString(line)
	}
	return "", unstamped.String()
}

func contentWithoutStampLine(lines []string, stampLine string) string {
	var unstamped strings.Builder
	removed := false
	for _, line := range lines {
		if !removed && line == stampLine {
			removed = true
			continue
		}
		unstamped.WriteString(line)
	}
	return unstamped.String()
}

func mutationContentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func progressWriter(options options, stdout, stderr io.Writer) io.Writer {
	if options.jsonReport {
		return stderr
	}
	return stdout
}

func printProgress(w io.Writer) func(acceptance.MutationProgress) {
	return func(progress acceptance.MutationProgress) {
		fmt.Fprintf(w, "progress completed=%d total=%d killed=%d survived=%d errors=%d\n",
			progress.Completed,
			progress.Total,
			progress.Killed,
			progress.Survived,
			progress.Errors,
		)
	}
}

func printReport(stdout, stderr io.Writer, jsonReport bool, summary acceptance.MutationSummary, results []acceptance.MutationResult) {
	if jsonReport {
		printJSON(stdout, stderr, summary, results)
	} else {
		printText(stdout, summary, results)
	}
}

func exitCodeFromSummary(summary acceptance.MutationSummary) int {
	if summary.Survived > 0 || summary.Errors > 0 {
		return 1
	}
	return 0
}

func printText(w io.Writer, summary acceptance.MutationSummary, results []acceptance.MutationResult) {
	fmt.Fprintf(w, "total=%d killed=%d survived=%d errors=%d\n", summary.Total, summary.Killed, summary.Survived, summary.Errors)
	for _, result := range results {
		printResult(w, result)
	}
}

func printResult(w io.Writer, result acceptance.MutationResult) {
	fmt.Fprintf(w, "%-8s %s\n", result.Status, result.Mutation.Description)
	if result.Status != "survived" && result.Status != "error" {
		return
	}
	printError(w, result.Error)
	printOutput(w, result.Output)
}

func printError(w io.Writer, message string) {
	printOptional(w, "error", message, false)
}

func printOutput(w io.Writer, output string) {
	printOptional(w, "output", output, true)
}

func printOptional(w io.Writer, label, value string, block bool) {
	if value == "" {
		return
	}
	if block {
		fmt.Fprintf(w, "  %s:\n%s", label, value)
	} else {
		fmt.Fprintf(w, "  %s: %s\n", label, value)
	}
}

func printJSON(stdout, stderr io.Writer, summary acceptance.MutationSummary, results []acceptance.MutationResult) {
	data, err := json.MarshalIndent(struct {
		Summary acceptance.MutationSummary
		Results []acceptance.MutationResult
	}{summary, results}, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return
	}
	fmt.Fprintln(stdout, string(data))
}
