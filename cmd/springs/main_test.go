//go:build appunit

package main

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestExitIfErrorAllowsNil(t *testing.T) {
	exitIfError(nil)
}

func TestExitIfErrorFatal(t *testing.T) {
	if os.Getenv("SPRINGS_MAIN_FATAL_TEST") == "1" {
		log.SetOutput(io.Discard)
		exitIfError(errors.New("boom"))
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExitIfErrorFatal")
	cmd.Env = append(os.Environ(), "SPRINGS_MAIN_FATAL_TEST=1")
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("exit code = %d, want 1", exitErr.ExitCode())
	}
}
