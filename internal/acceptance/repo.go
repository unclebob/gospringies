package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runCommand(name string, args ...string) error {
	return runCommandWithEnv(nil, name, args...)
}

func runCommandWithEnv(env []string, name string, args ...string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	return runCommandInDirWithEnv(root, env, name, args...)
}

func runCommandInDir(dir, name string, args ...string) error {
	return runCommandInDirWithEnv(dir, nil, name, args...)
}

func runCommandInDirWithEnv(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, output)
	}
	return nil
}

func runPipeline(feature, base string) error {
	jsonPath := "build/_acceptance-pipeline/" + base + "/feature.json"
	generatedPath := "build/_acceptance-pipeline/" + base + "/generated/feature_acceptance_test.go"
	if err := runParserAndGenerator(feature, jsonPath, generatedPath); err != nil {
		return err
	}
	return runCommand("go", "test", "./build/_acceptance-pipeline/"+base+"/generated")
}

func runParserAndGenerator(feature, jsonPath, generatedPath string) error {
	if err := runParser(feature, jsonPath); err != nil {
		return err
	}
	return runGenerator(jsonPath, generatedPath)
}

func runParser(feature, output string) error {
	return runCommand("go", "run", "./cmd/gherkin-parser", feature, output)
}

func runGenerator(jsonPath, output string) error {
	return runCommand("go", "run", "./cmd/acceptance-generator", jsonPath, output)
}

func fileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func repoPath(path string) string {
	root, err := repoRoot()
	if err != nil {
		return path
	}
	return filepath.Join(root, path)
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if hasGoMod(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate go.mod")
		}
		dir = parent
	}
}

func hasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}
