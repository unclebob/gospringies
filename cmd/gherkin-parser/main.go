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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:51:41-05:00","module_hash":"17c9e96fb5ce29fec2889f28ab396b4b54fc1bbf2b58586aca9a6e39cbc58102","functions":[{"id":"func/main","name":"main","line":11,"end_line":13,"hash":"456fb961f0d5d132d6d1e97ed1c9a19d21495a7ede09523a5327d7ed85c0b4f4"},{"id":"func/run","name":"run","line":15,"end_line":25,"hash":"94f47f526ac0c1d36ed7ebb1800e4bb46300d2d40d8f776bca7cb5badf788644"},{"id":"func/writeFeatureJSON","name":"writeFeatureJSON","line":27,"end_line":36,"hash":"2c8e3a2405089e8f03c4676a0998a86a0c030a9785cf8ac59d81e1e1475b0899"}]}
// mutate4go-manifest-end
