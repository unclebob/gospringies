package main

import (
	"fmt"
	"os"

	"springs/internal/acceptancegen"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: acceptance-generator <json-ir> <generated-test-output>")
		return 2
	}
	if err := acceptancegen.GenerateGoTest(args[1], args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-18T21:14:20-05:00","module_hash":"3bce3be60cc81c27ab59617b80e3c5b0ca66219e48f5da2f1a83a21552c7a107","functions":[{"id":"func/main","name":"main","line":10,"end_line":12,"hash":"456fb961f0d5d132d6d1e97ed1c9a19d21495a7ede09523a5327d7ed85c0b4f4"},{"id":"func/run","name":"run","line":14,"end_line":24,"hash":"e576d2f4ba4d48652eab7801d0613bf95e6307bf48076b7ca7d6bd8206ff9fff"}]}
// mutate4go-manifest-end
