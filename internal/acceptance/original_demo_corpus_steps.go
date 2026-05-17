package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	xspfmt "springs/internal/format"
)

func importOriginalDemoCorpus(*world, map[string]string) error {
	return fileExists(repoPath("demos/original/PROVENANCE.md"))
}

func assertImportedDemoExists(_ *world, example map[string]string) error {
	path, err := importedDemoPath(example)
	if err != nil {
		return err
	}
	return fileExists(path)
}

func assertImportedDemoPreservesFilename(_ *world, example map[string]string) error {
	demoFile, err := stringValue(example, "demo_file")
	if err != nil {
		return err
	}
	path, err := importedDemoPath(example)
	if err != nil {
		return err
	}
	if filepath.Base(path) != demoFile {
		return fmt.Errorf("imported filename = %q, want %q", filepath.Base(path), demoFile)
	}
	return nil
}

func assertImportedOriginalDemoExists(w *world, example map[string]string) error {
	return assertImportedDemoExists(w, map[string]string{
		"demo_directory": "demos/original",
		"demo_file":      example["demo_file"],
	})
}

func loadImportedOriginalDemo(w *world, example map[string]string) error {
	path, err := originalDemoFilePath(example)
	if err != nil {
		w.xspLoadErr = err
		return nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		w.xspLoadErr = err
		return nil
	}
	w.xspWorld, w.xspLoadErr = xspfmt.LoadXSP(string(content))
	return nil
}

func assertLoadingPassed(w *world, _ map[string]string) error {
	return w.xspLoadErr
}

func assertStarterDemoExists(_ *world, example map[string]string) error {
	directory, demoFile, err := stringPair(example, "starter_directory", "starter_demo")
	if err != nil {
		return err
	}
	return fileExists(repoPath(filepath.Join(directory, demoFile)))
}

func assertStarterDemoRemainsUnder(_ *world, example map[string]string) error {
	return assertStarterDemoExists(nil, example)
}

func assertOriginalDemosRemainUnder(_ *world, example map[string]string) error {
	directory, err := stringValue(example, "original_directory")
	if err != nil {
		return err
	}
	return fileExists(repoPath(directory))
}

func assertProvenanceFieldDocumented(_ *world, example map[string]string) error {
	field, err := stringValue(example, "field")
	if err != nil {
		return err
	}
	content, err := os.ReadFile(repoPath("demos/original/PROVENANCE.md"))
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(string(content)), strings.ToLower(field)) {
		return fmt.Errorf("provenance field %q was not documented", field)
	}
	return nil
}

func importedDemoPath(example map[string]string) (string, error) {
	directory, demoFile, err := stringPair(example, "demo_directory", "demo_file")
	if err != nil {
		return "", err
	}
	return repoPath(filepath.Join(directory, demoFile)), nil
}

func originalDemoFilePath(example map[string]string) (string, error) {
	demoFile, err := stringValue(example, "demo_file")
	if err != nil {
		return "", err
	}
	return repoPath(filepath.Join("demos", "original", demoFile)), nil
}
