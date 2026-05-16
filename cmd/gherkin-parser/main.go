package main

import (
	"fmt"
	"os"
	"path/filepath"

	"springs/internal/gherkin"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: gherkin-parser <feature-file> <json-output>")
		return 2
	}
	if err := writeFeatureJSON(args[1], args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func writeFeatureJSON(featurePath, outputPath string) error {
	feature, err := gherkin.ReadFile(featurePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return gherkin.WriteJSON(feature, outputPath)
}
