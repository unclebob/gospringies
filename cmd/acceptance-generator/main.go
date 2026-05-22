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
// {"version":1,"tested_at":"2026-05-22T10:51:49-05:00","module_hash":"e0724313bfd2a6e4aedd42f9a45ed091ac2c3967f063e3ae3c67ba0a4f2401d5","functions":[{"id":"func/main","name":"main","line":10,"end_line":12,"hash":"456fb961f0d5d132d6d1e97ed1c9a19d21495a7ede09523a5327d7ed85c0b4f4"},{"id":"func/run","name":"run","line":14,"end_line":24,"hash":"07ede0653fa9fd6438c934af6db586a56adaa704418279aaa9850b0d47f9b655"}]}
// mutate4go-manifest-end
