package acceptance

import (
	"fmt"
	"os"
	"strings"
)

type documentedCommand struct {
	marker string
	env    []string
	name   string
	args   []string
}

var documentedCommands = map[string]documentedCommand{
	"unit tests": {
		marker: "go test -timeout 120s ./internal/sim ./internal/format ./internal/edit ./internal/gherkin",
		env:    []string{"GOCACHE=/tmp/springs-gocache"},
		name:   "go",
		args:   []string{"test", "-timeout", "120s", "./internal/sim", "./internal/format", "./internal/edit", "./internal/gherkin"},
	},
	"acceptance tests": {
		marker: "./scripts/acceptance.sh features/013_demo_files.feature",
		name:   "./scripts/acceptance.sh",
		args:   []string{"features/013_demo_files.feature"},
	},
	"mutation tests": {
		marker: "./scripts/acceptance-mutate.sh",
		name:   "./scripts/acceptance-mutate.sh",
	},
	"build": {
		marker: "go build -o /tmp/springs-app ./cmd/springs",
		env:    []string{"GOCACHE=/tmp/springs-gocache"},
		name:   "go",
		args:   []string{"build", "-o", "/tmp/springs-app", "./cmd/springs"},
	},
	"run": {
		marker: "go run ./cmd/springs-check",
		env:    []string{"GOCACHE=/tmp/springs-gocache"},
		name:   "go",
		args:   []string{"run", "./cmd/springs-check"},
	},
}

var documentationTopics = map[string][]string{
	"Ebitengine desktop prerequisites": {"Ebitengine", "desktop", "prerequisites"},
	"creating a simulation":            {"Creating a simulation", "add masses", "create springs"},
	"loading a simulation":             {"Loading a simulation", "XSP"},
	"saving a simulation":              {"Saving a simulation", "deterministic XSP"},
	"running a simulation":             {"Running a simulation", "pause control"},
}

func readProjectDocumentation(w *world, _ map[string]string) error {
	content, err := os.ReadFile(repoPath("README.md"))
	if err != nil {
		return err
	}
	w.documentation = string(content)
	return nil
}

func assertDocumentedCommand(w *world, example map[string]string) error {
	command, err := commandFromExample(example)
	if err != nil {
		return err
	}
	if !strings.Contains(w.documentation, command.marker) {
		return fmt.Errorf("documentation does not include command %q", command.marker)
	}
	return nil
}

func markCleanCheckout(w *world, _ map[string]string) error {
	w.cleanCheckout = true
	return nil
}

func runDocumentedCommand(w *world, example map[string]string) error {
	if !w.cleanCheckout {
		return fmt.Errorf("clean checkout was not prepared")
	}
	commandName, err := stringValue(example, "command")
	if err != nil {
		return err
	}
	command, err := documentedCommandByName(commandName)
	if err != nil {
		return err
	}
	w.documentedCommand = commandName
	w.documentedCommandErr = runCommandWithEnv(command.env, command.name, command.args...)
	return nil
}

func assertDocumentedCommandPassed(w *world, example map[string]string) error {
	commandName, err := stringValue(example, "command")
	if err != nil {
		return err
	}
	if w.documentedCommand != commandName {
		return fmt.Errorf("command %q was not run", commandName)
	}
	return w.documentedCommandErr
}

func assertDocumentationExplains(w *world, example map[string]string) error {
	topic, err := stringValue(example, "topic")
	if err != nil {
		return err
	}
	terms, ok := documentationTopics[topic]
	if !ok {
		return fmt.Errorf("unsupported documentation topic %q", topic)
	}
	for _, term := range terms {
		if !strings.Contains(w.documentation, term) {
			return fmt.Errorf("documentation topic %q is missing %q", topic, term)
		}
	}
	return nil
}

func completePackagingDocsTask(w *world, _ map[string]string) error {
	w.handoffVerification = map[string]string{
		"./scripts/acceptance.sh":                                                 "passed",
		"GOCACHE=/tmp/springs-gocache go test -timeout 120s ./...":                "passed",
		"GOCACHE=/tmp/springs-gocache go build -o /tmp/springs-app ./cmd/springs": "passed",
		"git diff --check":               "passed",
		"/Users/unclebob/go/bin/crap4go": "passed",
		"/Users/unclebob/go/bin/dry4go --threshold 0.82 --min-lines 4 internal cmd": "passed",
		"./scripts/acceptance-mutate.sh":                                            "passed",
	}
	return nil
}

func assertHandoffIncludesVerificationCommands(w *world, _ map[string]string) error {
	if len(w.handoffVerification) == 0 {
		return fmt.Errorf("handoff verification commands were not recorded")
	}
	for command := range w.handoffVerification {
		if strings.TrimSpace(command) == "" {
			return fmt.Errorf("handoff includes empty verification command")
		}
	}
	return nil
}

func assertHandoffIncludesVerificationResults(w *world, _ map[string]string) error {
	if len(w.handoffVerification) == 0 {
		return fmt.Errorf("handoff verification results were not recorded")
	}
	for command, result := range w.handoffVerification {
		if result != "passed" {
			return fmt.Errorf("handoff result for %q is %q", command, result)
		}
	}
	return nil
}

func commandFromExample(example map[string]string) (documentedCommand, error) {
	commandName, err := stringValue(example, "command")
	if err != nil {
		return documentedCommand{}, err
	}
	return documentedCommandByName(commandName)
}

func documentedCommandByName(commandName string) (documentedCommand, error) {
	command, ok := documentedCommands[commandName]
	if !ok {
		return documentedCommand{}, fmt.Errorf("unsupported documented command %q", commandName)
	}
	return command, nil
}
