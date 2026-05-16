package main

import (
	"fmt"
	"os"

	"springs/internal/acceptance"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: acceptance-generator <json-ir> <generated-test-output>")
		os.Exit(2)
	}
	if err := acceptance.GenerateGoTest(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
