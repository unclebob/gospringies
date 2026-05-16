package main

import (
	"fmt"
	"os"

	"springs/internal/acceptance"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: acceptance-generator <json-ir> <generated-test-output>")
		return 2
	}
	if err := acceptance.GenerateGoTest(args[1], args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
