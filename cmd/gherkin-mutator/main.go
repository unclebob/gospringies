package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"springs/internal/acceptancemutation"
	"springs/internal/gherkin"
	"springs/internal/mutationstamp"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		return 2
	}
	if mutationstamp.Valid(options.featurePath) {
		fmt.Fprintf(stdout, "mutation stamp valid; skipping %s\n", options.featurePath)
		return 0
	}
	summary, results, err := runFeatureMutations(options, progressWriter(options, stdout, stderr))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printReport(stdout, stderr, options.jsonReport, summary, results)
	return finishRun(options.featurePath, summary, stderr)
}

func finishRun(featurePath string, summary acceptancemutation.MutationSummary, stderr io.Writer) int {
	code := exitCodeFromSummary(summary)
	if code != 0 {
		return code
	}
	if err := mutationstamp.Stamp(featurePath); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

type options struct {
	featurePath   string
	workDir       string
	jsonReport    bool
	workers       int
	timeout       time.Duration
	mutantTimeout time.Duration
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	flags := flag.NewFlagSet("gherkin-mutator", flag.ContinueOnError)
	flags.SetOutput(stderr)
	featurePath := flags.String("feature", "features/a-feature.feature", "Gherkin feature file to parse and mutate")
	workDir := flags.String("work-dir", "build/acceptance-mutation", "directory where mutation work files are written")
	jsonReport := flags.Bool("json", false, "emit JSON report")
	workers := flags.Int("workers", runtime.NumCPU(), "maximum mutation workers")
	timeout := flags.Duration("timeout", 0, "full mutation timeout")
	mutantTimeout := flags.Duration("mutant-timeout", 30*time.Second, "timeout for one generated mutation test")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	return options{
		featurePath:   *featurePath,
		workDir:       *workDir,
		jsonReport:    *jsonReport,
		workers:       *workers,
		timeout:       *timeout,
		mutantTimeout: *mutantTimeout,
	}, nil
}

func runFeatureMutations(options options, progress io.Writer) (acceptancemutation.MutationSummary, []acceptancemutation.MutationResult, error) {
	feature, err := gherkin.ReadFile(options.featurePath)
	if err != nil {
		return acceptancemutation.MutationSummary{}, nil, err
	}
	ctx, cancel := mutationContext(options.timeout)
	defer cancel()
	results, err := acceptancemutation.RunMutationsWithOptions(feature, options.workDir, acceptancemutation.RunMutationOptions{
		Context:       ctx,
		Workers:       options.workers,
		MutantTimeout: options.mutantTimeout,
		ProgressEvery: 20,
		Progress:      printProgress(progress),
	})
	if err != nil {
		return acceptancemutation.MutationSummary{}, nil, err
	}
	return acceptancemutation.Summarize(results), results, nil
}

func mutationContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), timeout)
}

func progressWriter(options options, stdout, stderr io.Writer) io.Writer {
	if options.jsonReport {
		return stderr
	}
	return stdout
}

func printProgress(w io.Writer) func(acceptancemutation.MutationProgress) {
	return func(progress acceptancemutation.MutationProgress) {
		fmt.Fprintf(w, "progress completed=%d total=%d killed=%d survived=%d errors=%d\n",
			progress.Completed,
			progress.Total,
			progress.Killed,
			progress.Survived,
			progress.Errors,
		)
	}
}

func printReport(stdout, stderr io.Writer, jsonReport bool, summary acceptancemutation.MutationSummary, results []acceptancemutation.MutationResult) {
	if jsonReport {
		printJSON(stdout, stderr, summary, results)
	} else {
		printText(stdout, summary, results)
	}
}

func exitCodeFromSummary(summary acceptancemutation.MutationSummary) int {
	if summary.Survived > 0 || summary.Errors > 0 {
		return 1
	}
	return 0
}

func printText(w io.Writer, summary acceptancemutation.MutationSummary, results []acceptancemutation.MutationResult) {
	fmt.Fprintf(w, "total=%d killed=%d survived=%d errors=%d\n", summary.Total, summary.Killed, summary.Survived, summary.Errors)
	for _, result := range results {
		printResult(w, result)
	}
}

func printResult(w io.Writer, result acceptancemutation.MutationResult) {
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

func printJSON(stdout, stderr io.Writer, summary acceptancemutation.MutationSummary, results []acceptancemutation.MutationResult) {
	data, err := json.MarshalIndent(struct {
		Summary acceptancemutation.MutationSummary
		Results []acceptancemutation.MutationResult
	}{summary, results}, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return
	}
	fmt.Fprintln(stdout, string(data))
}
