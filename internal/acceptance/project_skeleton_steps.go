package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func createPackageLayout(w *world, _ map[string]string) error {
	return markCreated(&w.layoutCreated)
}

func assertPackageDoesNotImport(w *world, example map[string]string) error {
	if err := requirePrerequisite(w.layoutCreated, "package layout has not been created"); err != nil {
		return err
	}
	packageName, err := stringValue(example, "package")
	if err != nil {
		return err
	}
	library, err := stringValue(example, "graphics_library")
	if err != nil {
		return err
	}
	return packageDoesNotImport(packageName, library)
}

func createApplicationCommand(w *world, _ map[string]string) error {
	return markCreated(&w.commandCreated)
}

func assertApplicationCommandBuilds(w *world, _ map[string]string) error {
	if err := requirePrerequisite(w.commandCreated, "application command has not been created"); err != nil {
		return err
	}
	return runCommand("go", "build", "-o", filepath.Join(os.TempDir(), "springs-acceptance-app"), "./cmd/springs")
}

func createGoModule(w *world, _ map[string]string) error {
	return markCreated(&w.moduleCreated)
}

func assertGoTestsPass(w *world, _ map[string]string) error {
	if err := requirePrerequisite(w.moduleCreated, "go module has not been created"); err != nil {
		return err
	}
	return runCommand("go", "test", "./internal/...", "./cmd/...")
}

func packageDoesNotImport(packageName, library string) error {
	dir, err := domainPackageDir(packageName)
	if err != nil {
		return err
	}
	if strings.ToLower(strings.TrimSpace(library)) != "ebitengine" {
		return fmt.Errorf("unsupported graphics library %q", library)
	}
	root, err := repoRoot()
	if err != nil {
		return err
	}
	return packageDirDoesNotImport(filepath.Join(root, dir), packageName, library)
}

func packageDirDoesNotImport(dir, packageName, library string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		imports, err := fileImportsLibrary(dir, entry, library)
		if err != nil {
			return err
		}
		if imports {
			return fmt.Errorf("%s package imports %s", packageName, library)
		}
	}
	return nil
}

func fileImportsLibrary(dir string, entry os.DirEntry, library string) (bool, error) {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
		return false, nil
	}
	data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
	if err != nil {
		return false, err
	}
	return mentionsGraphicsLibrary(string(data), library), nil
}

func mentionsGraphicsLibrary(source, library string) bool {
	source = strings.ToLower(source)
	for _, needle := range []string{strings.ToLower(library), "github.com/hajimehoshi/ebiten"} {
		if strings.Contains(source, needle) {
			return true
		}
	}
	return false
}

func domainPackageDir(packageName string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(packageName)) {
	case "simulation":
		return "internal/sim", nil
	case "file format":
		return "internal/format", nil
	default:
		return "", fmt.Errorf("unknown package %q", packageName)
	}
}
