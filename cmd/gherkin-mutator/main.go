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
	content, _ := os.ReadFile(options.featurePath)
	_, hasScenarioManifest, _ := acceptancemutation.ParseScenarioManifest(string(content))
	if !hasScenarioManifest && mutationstamp.Valid(options.featurePath) {
		fmt.Fprintf(stdout, "mutation stamp valid; skipping %s\n", options.featurePath)
		return 0
	}
	implementationHash, err := acceptancemutation.CurrentImplementationHash()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	summary, results, skipPlan, manifest, feature, err := runFeatureMutations(options, implementationHash, progressWriter(options, stdout, stderr))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printSkipReport(stdout, skipPlan)
	printReport(stdout, stderr, options.jsonReport, summary, results)
	return finishRun(options.featurePath, summary, manifest, skipPlan, results, feature, implementationHash, stderr)
}

func finishRun(featurePath string, summary acceptancemutation.MutationSummary, manifest acceptancemutation.ScenarioManifest, skipPlan acceptancemutation.ScenarioSkipPlan, results []acceptancemutation.MutationResult, feature gherkin.Feature, implementationHash string, stderr io.Writer) int {
	code := exitCodeFromSummary(summary)
	if code != 0 {
		return code
	}
	if err := acceptancemutation.WriteScenarioManifestFile(featurePath, feature, manifest, skipPlan, results, implementationHash, time.Now()); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
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

func runFeatureMutations(options options, implementationHash string, progress io.Writer) (acceptancemutation.MutationSummary, []acceptancemutation.MutationResult, acceptancemutation.ScenarioSkipPlan, acceptancemutation.ScenarioManifest, gherkin.Feature, error) {
	feature, err := gherkin.ReadFile(options.featurePath)
	if err != nil {
		return acceptancemutation.MutationSummary{}, nil, acceptancemutation.ScenarioSkipPlan{}, acceptancemutation.ScenarioManifest{}, gherkin.Feature{}, err
	}
	content, err := os.ReadFile(options.featurePath)
	if err != nil {
		return acceptancemutation.MutationSummary{}, nil, acceptancemutation.ScenarioSkipPlan{}, acceptancemutation.ScenarioManifest{}, gherkin.Feature{}, err
	}
	manifest, _, err := acceptancemutation.ParseScenarioManifest(string(content))
	if err != nil {
		return acceptancemutation.MutationSummary{}, nil, acceptancemutation.ScenarioSkipPlan{}, acceptancemutation.ScenarioManifest{}, gherkin.Feature{}, err
	}
	skipPlan := acceptancemutation.ScenarioSkipPlanFor(feature, options.featurePath, manifest, implementationHash)
	ctx, cancel := mutationContext(options.timeout)
	defer cancel()
	results, err := acceptancemutation.RunMutationsWithOptions(feature, options.workDir, acceptancemutation.RunMutationOptions{
		Context:       ctx,
		Workers:       options.workers,
		MutantTimeout: options.mutantTimeout,
		ProgressEvery: 20,
		Progress:      printProgress(progress),
		MutationFilter: func(mutation acceptancemutation.Mutation) bool {
			return !skipPlan.SkipScenarios[mutation.Scenario]
		},
	})
	if err != nil {
		return acceptancemutation.MutationSummary{}, nil, acceptancemutation.ScenarioSkipPlan{}, acceptancemutation.ScenarioManifest{}, gherkin.Feature{}, err
	}
	return acceptancemutation.Summarize(results), results, skipPlan, manifest, feature, nil
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

func printSkipReport(stdout io.Writer, skipPlan acceptancemutation.ScenarioSkipPlan) {
	if skipPlan.SkippedScenarios == 0 {
		return
	}
	fmt.Fprintf(stdout, "skipped_scenarios=%d skipped_mutations=%d\n", skipPlan.SkippedScenarios, skipPlan.SkippedMutations)
	fmt.Fprintln(stdout, "scenario manifest valid for skipped scenarios")
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
	if result.Status != acceptancemutation.MutationSurvived && result.Status != acceptancemutation.MutationError {
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-18T21:14:29-05:00","module_hash":"8844b1326e393ab0ce3c6cf4259606414a6e2fb89d61ee3e8b32bbdbb527b9b4","functions":[{"id":"func/main","name":"main","line":18,"end_line":20,"hash":"0a99d648406cdf7162467223a8772faf9edec83f7793a817ec79fb28063a61c2"},{"id":"func/run","name":"run","line":22,"end_line":38,"hash":"6d4b014b46171c6945753b189391337870d060d8fc3221acaf209b9e1dcb6ae1"},{"id":"func/finishRun","name":"finishRun","line":40,"end_line":50,"hash":"197ca584d817fda024420b46059881e7ab1d22763b16a05dffef67b33aed771c"},{"id":"func/parseOptions","name":"parseOptions","line":61,"end_line":81,"hash":"23108dbe77d6b34319bbc03a50835fddc8b7db03960787de05a8b201e802b4df"},{"id":"func/runFeatureMutations","name":"runFeatureMutations","line":83,"end_line":101,"hash":"c777096703a6acc73ebf2f92664af6e3bda702358f5580c1abfa2bc1bfc7c503"},{"id":"func/mutationContext","name":"mutationContext","line":103,"end_line":108,"hash":"da44888fe48052d3996da8880c37d8aadf2a0fd64d8c2436ab10c532010713ba"},{"id":"func/progressWriter","name":"progressWriter","line":110,"end_line":115,"hash":"04bd94380ab50bf06c4f6e898266b33d9f875e971f0c3effa230194c57ffa244"},{"id":"func/printProgress","name":"printProgress","line":117,"end_line":127,"hash":"c5a5599a78d574fd685a4ce319117c40f9160afcdd7ec348bb70094dd13ffd02"},{"id":"func/printReport","name":"printReport","line":129,"end_line":135,"hash":"bf8ad75c5c267c00b8e504804d51ee081ac216d50660b61b70086d3a28f04b9b"},{"id":"func/exitCodeFromSummary","name":"exitCodeFromSummary","line":137,"end_line":142,"hash":"78eccb865fcaed32819cc7b2a27f065f5ade2a88be79b9cd64aa22e5811a2a94"},{"id":"func/printText","name":"printText","line":144,"end_line":149,"hash":"56ae71fd68cb0df38dca753c45201c9b9f6c284c69359ccd68b85af4ad3e84eb"},{"id":"func/printResult","name":"printResult","line":151,"end_line":158,"hash":"ac80ed18ad2f0ff1b73597d41343c7b96c86bebb39b87a4cbc2c20f11bdaf484"},{"id":"func/printError","name":"printError","line":160,"end_line":162,"hash":"3820d9a2e1a1340e727c329477a96a9165102a6a9b25a82c749947d8cb266868"},{"id":"func/printOutput","name":"printOutput","line":164,"end_line":166,"hash":"b8f7818fe0106fc0e50c53ae3c2e0d8ec8b0722b235b1484835f977263a72626"},{"id":"func/printOptional","name":"printOptional","line":168,"end_line":177,"hash":"c5d0bc4af69a1642e5f42c87ffe64eb5bfba2fc2a4f96ea1266e6218008c95f0"},{"id":"func/printJSON","name":"printJSON","line":179,"end_line":189,"hash":"5687a2159273948fa1e379398f7ad5445b8c9782c8a15dcb469e3fe4aa60b30b"}]}
// mutate4go-manifest-end
