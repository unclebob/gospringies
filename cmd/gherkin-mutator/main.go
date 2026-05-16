package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"springs/internal/acceptance"
	"springs/internal/gherkin"
)

func main() {
	featurePath := flag.String("feature", "features/a-feature.feature", "Gherkin feature file to parse and mutate")
	workDir := flag.String("work-dir", "build/acceptance-mutation", "directory where mutation work files are written")
	jsonReport := flag.Bool("json", false, "emit JSON report")
	_ = flag.Int("workers", 1, "maximum mutation workers")
	_ = flag.Duration("timeout", 0, "full mutation timeout")
	flag.Parse()

	feature, err := gherkin.ReadFile(*featurePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	results, err := acceptance.RunMutations(feature, *workDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	summary := acceptance.Summarize(results)
	if *jsonReport {
		printJSON(summary, results)
	} else {
		printText(summary, results)
	}
	if summary.Survived > 0 || summary.Errors > 0 {
		os.Exit(1)
	}
}

func printText(summary acceptance.MutationSummary, results []acceptance.MutationResult) {
	fmt.Printf("total=%d killed=%d survived=%d errors=%d\n", summary.Total, summary.Killed, summary.Survived, summary.Errors)
	for _, result := range results {
		fmt.Printf("%-8s %s\n", result.Status, result.Mutation.Description)
		if result.Status == "survived" || result.Status == "error" {
			if result.Error != "" {
				fmt.Printf("  error: %s\n", result.Error)
			}
			if result.Output != "" {
				fmt.Printf("  output:\n%s", result.Output)
			}
		}
	}
}

func printJSON(summary acceptance.MutationSummary, results []acceptance.MutationResult) {
	data, err := json.MarshalIndent(struct {
		Summary acceptance.MutationSummary
		Results []acceptance.MutationResult
	}{summary, results}, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
