package acceptance

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"springs/internal/gherkin"
	"springs/internal/sim"
)

type world struct {
	simulation     *sim.Simulation
	layoutCreated  bool
	commandCreated bool
	moduleCreated  bool
}

type stepHandler func(*world, map[string]string) error

func RunFeature(feature gherkin.Feature) error {
	for _, scenario := range feature.Scenarios {
		examples := scenario.Examples
		if len(examples) == 0 {
			examples = []map[string]string{{}}
		}
		for i, example := range examples {
			w := &world{}
			steps := append([]gherkin.Step{}, feature.Background...)
			steps = append(steps, scenario.Steps...)
			for _, step := range steps {
				if err := runStep(w, step, example); err != nil {
					return fmt.Errorf("%s/example_%d: %w", scenario.Name, i+1, err)
				}
			}
		}
	}
	return nil
}

func runStep(w *world, step gherkin.Step, example map[string]string) error {
	handler, ok := stepHandlers[step.Text]
	if !ok {
		return fmt.Errorf("unsupported step %q", step.Text)
	}
	return handler(w, example)
}

var stepHandlers = map[string]stepHandler{
	"the project skeleton task is accepted":                      acceptProjectSkeleton,
	"the coder creates the initial Go package layout":            createPackageLayout,
	"the <package> package should not import <graphics_library>": assertPackageDoesNotImport,
	"the coder creates the desktop application command":          createApplicationCommand,
	"the application command should build successfully":          assertApplicationCommandBuilds,
	"the coder creates the initial Go module":                    createGoModule,
	"the Go test suite should pass":                              assertGoTestsPass,
	"a demo spring simulation":                                   createDemoSimulation,
	"I advance the simulation <steps> steps":                     advanceSimulation,
	"mass <mass> x should be <x>":                                assertMassX,
}

func acceptProjectSkeleton(*world, map[string]string) error {
	return nil
}

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

func requirePrerequisite(ready bool, message string) error {
	if !ready {
		return fmt.Errorf("%s", message)
	}
	return nil
}

func markCreated(created *bool) error {
	*created = true
	return nil
}

func createDemoSimulation(w *world, _ map[string]string) error {
	w.simulation = sim.NewDemoSimulation()
	return nil
}

func advanceSimulation(w *world, example map[string]string) error {
	steps, err := intValue(example, "steps")
	if err != nil {
		return err
	}
	if w.simulation == nil {
		return fmt.Errorf("simulation is not ready")
	}
	w.simulation.Advance(steps, 0.016)
	return nil
}

func assertMassX(w *world, example map[string]string) error {
	massIndex, mass, err := exampleMass(w, example)
	if err != nil {
		return err
	}
	expected, err := floatValue(example, "x")
	if err != nil {
		return err
	}
	got := mass.Position.X
	if math.Abs(got-expected) > 0.00001 {
		return fmt.Errorf("expected mass %d x %f, got %f", massIndex, expected, got)
	}
	return nil
}

func exampleMass(w *world, example map[string]string) (int, sim.Mass, error) {
	massIndex, err := intValue(example, "mass")
	if err != nil {
		return 0, sim.Mass{}, err
	}
	if w.simulation == nil {
		return 0, sim.Mass{}, fmt.Errorf("simulation is not ready")
	}
	if massIndex < 0 || massIndex >= len(w.simulation.Masses) {
		return 0, sim.Mass{}, fmt.Errorf("mass index %d out of range", massIndex)
	}
	return massIndex, w.simulation.Masses[massIndex], nil
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

func runCommand(name string, args ...string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	return runCommandInDir(root, name, args...)
}

func runCommandInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, output)
	}
	return nil
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

func stringValue(example map[string]string, key string) (string, error) {
	value, ok := example[key]
	if !ok {
		return "", fmt.Errorf("missing example value %s", key)
	}
	return value, nil
}

func intValue(example map[string]string, key string) (int, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid integer %s=%q", key, value)
	}
	return parsed, nil
}

func floatValue(example map[string]string, key string) (float64, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %s=%q", key, value)
	}
	return parsed, nil
}
