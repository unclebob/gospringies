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
	if canSkipStampedFeature(options, hasScenarioManifest) {
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

func canSkipStampedFeature(options options, hasScenarioManifest bool) bool {
	return options.level != acceptancemutation.ScenarioManifestFull && !hasScenarioManifest && mutationstamp.Valid(options.featurePath)
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
	level         acceptancemutation.ScenarioManifestMode
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
	level := flags.String("level", string(acceptancemutation.ScenarioManifestHard), "mutation level: hard, soft, or full")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	parsedLevel, err := parseMutationLevel(*level)
	if err != nil {
		return options{}, err
	}
	return options{
		featurePath:   *featurePath,
		workDir:       *workDir,
		jsonReport:    *jsonReport,
		workers:       *workers,
		timeout:       *timeout,
		mutantTimeout: *mutantTimeout,
		level:         parsedLevel,
	}, nil
}

func parseMutationLevel(value string) (acceptancemutation.ScenarioManifestMode, error) {
	switch acceptancemutation.ScenarioManifestMode(value) {
	case acceptancemutation.ScenarioManifestHard, acceptancemutation.ScenarioManifestSoft, acceptancemutation.ScenarioManifestFull:
		return acceptancemutation.ScenarioManifestMode(value), nil
	default:
		return "", fmt.Errorf("invalid mutation level %q: want hard, soft, or full", value)
	}
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
	skipPlan := acceptancemutation.ScenarioSkipPlanForMode(feature, options.featurePath, manifest, implementationHash, options.level)
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
// {"version":1,"tested_at":"2026-05-22T10:54:05-05:00","module_hash":"4eed3d382ce0e4edeb35517a62bdc045eef2472b6f1293258cecaf2006e846f4","functions":[{"id":"func/main","name":"main","line":18,"end_line":20,"hash":"0a99d648406cdf7162467223a8772faf9edec83f7793a817ec79fb28063a61c2"},{"id":"func/run","name":"run","line":22,"end_line":46,"hash":"c8bd4336a3d0dd681e9e8bf9ab2dd953364c9334915d3465ceaab74c891d59dd"},{"id":"func/canSkipStampedFeature","name":"canSkipStampedFeature","line":48,"end_line":50,"hash":"a0d4cb1a50bc91c3a8b688a40c69e063b1889d6e9e82297ee894fd49da90da89"},{"id":"func/finishRun","name":"finishRun","line":52,"end_line":66,"hash":"baaf2957df27198d2b491aa6f5156af094f09179affc52e263e71391c5e6047a"},{"id":"func/parseOptions","name":"parseOptions","line":78,"end_line":104,"hash":"52659ef18e36c7ddc132584de5989d8614b09daa93949bb69f75bdbbb04f98af"},{"id":"func/parseMutationLevel","name":"parseMutationLevel","line":106,"end_line":113,"hash":"6b616efe26521f782626543b11ecc1dc18713a71f17aa1645bab60740ed89b92"},{"id":"func/runFeatureMutations","name":"runFeatureMutations","line":115,"end_line":145,"hash":"811228a79a238b07c7c3adf30d6c5bdfb777514eaceff6e58e2feb85d4273c49"},{"id":"func/mutationContext","name":"mutationContext","line":147,"end_line":152,"hash":"da44888fe48052d3996da8880c37d8aadf2a0fd64d8c2436ab10c532010713ba"},{"id":"func/progressWriter","name":"progressWriter","line":154,"end_line":159,"hash":"04bd94380ab50bf06c4f6e898266b33d9f875e971f0c3effa230194c57ffa244"},{"id":"func/printProgress","name":"printProgress","line":161,"end_line":171,"hash":"c5a5599a78d574fd685a4ce319117c40f9160afcdd7ec348bb70094dd13ffd02"},{"id":"func/printReport","name":"printReport","line":173,"end_line":179,"hash":"bf8ad75c5c267c00b8e504804d51ee081ac216d50660b61b70086d3a28f04b9b"},{"id":"func/printSkipReport","name":"printSkipReport","line":181,"end_line":187,"hash":"63d461c0a6b7a4e0ab12994d8877e1bdc7f92a11503e1a043fcca65b8ac4c884"},{"id":"func/exitCodeFromSummary","name":"exitCodeFromSummary","line":189,"end_line":194,"hash":"78eccb865fcaed32819cc7b2a27f065f5ade2a88be79b9cd64aa22e5811a2a94"},{"id":"func/printText","name":"printText","line":196,"end_line":201,"hash":"56ae71fd68cb0df38dca753c45201c9b9f6c284c69359ccd68b85af4ad3e84eb"},{"id":"func/printResult","name":"printResult","line":203,"end_line":210,"hash":"c803304f2cac0f6c021c718a9c8b08214b9d2fc6767575398299461726952c8c"},{"id":"func/printError","name":"printError","line":212,"end_line":214,"hash":"3820d9a2e1a1340e727c329477a96a9165102a6a9b25a82c749947d8cb266868"},{"id":"func/printOutput","name":"printOutput","line":216,"end_line":218,"hash":"b8f7818fe0106fc0e50c53ae3c2e0d8ec8b0722b235b1484835f977263a72626"},{"id":"func/printOptional","name":"printOptional","line":220,"end_line":229,"hash":"c5d0bc4af69a1642e5f42c87ffe64eb5bfba2fc2a4f96ea1266e6218008c95f0"},{"id":"func/printJSON","name":"printJSON","line":231,"end_line":241,"hash":"5687a2159273948fa1e379398f7ad5445b8c9782c8a15dcb469e3fe4aa60b30b"}]}
// mutate4go-manifest-end
