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
	switch step.Text {
	case "the project skeleton task is accepted":
		return nil
	case "the coder creates the initial Go package layout":
		w.layoutCreated = true
		return nil
	case "the <package> package should not import <graphics_library>":
		if !w.layoutCreated {
			return fmt.Errorf("package layout has not been created")
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
	case "the coder creates the desktop application command":
		w.commandCreated = true
		return nil
	case "the application command should build successfully":
		if !w.commandCreated {
			return fmt.Errorf("application command has not been created")
		}
		return runCommand("go", "build", "-o", filepath.Join(os.TempDir(), "springs-acceptance-app"), "./cmd/springs")
	case "the coder creates the initial Go module":
		w.moduleCreated = true
		return nil
	case "the Go test suite should pass":
		if !w.moduleCreated {
			return fmt.Errorf("go module has not been created")
		}
		return runCommand("go", "test", "./internal/...", "./cmd/...")
	case "a demo spring simulation":
		w.simulation = sim.NewDemoSimulation()
		return nil
	case "I advance the simulation <steps> steps":
		steps, err := intValue(example, "steps")
		if err != nil {
			return err
		}
		if w.simulation == nil {
			return fmt.Errorf("simulation is not ready")
		}
		w.simulation.Advance(steps, 0.016)
		return nil
	case "mass <mass> x should be <x>":
		massIndex, err := intValue(example, "mass")
		if err != nil {
			return err
		}
		expected, err := floatValue(example, "x")
		if err != nil {
			return err
		}
		if w.simulation == nil {
			return fmt.Errorf("simulation is not ready")
		}
		if massIndex < 0 || massIndex >= len(w.simulation.Masses) {
			return fmt.Errorf("mass index %d out of range", massIndex)
		}
		got := w.simulation.Masses[massIndex].Position.X
		if math.Abs(got-expected) > 0.00001 {
			return fmt.Errorf("expected mass %d x %f, got %f", massIndex, expected, got)
		}
		return nil
	default:
		return fmt.Errorf("unsupported step %q", step.Text)
	}
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
	dir = filepath.Join(root, dir)
	needle := strings.ToLower(library)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(string(data)), needle) ||
			strings.Contains(strings.ToLower(string(data)), "github.com/hajimehoshi/ebiten") {
			return fmt.Errorf("%s package imports %s", packageName, library)
		}
	}
	return nil
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
	cmd := exec.Command(name, args...)
	cmd.Dir = root
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
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate go.mod")
		}
		dir = parent
	}
}

func stringValue(example map[string]string, key string) (string, error) {
	value, ok := example[key]
	if !ok {
		return "", fmt.Errorf("missing example value %s", key)
	}
	return value, nil
}

func intValue(example map[string]string, key string) (int, error) {
	value, ok := example[key]
	if !ok {
		return 0, fmt.Errorf("missing example value %s", key)
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid integer %s=%q", key, value)
	}
	return parsed, nil
}

func floatValue(example map[string]string, key string) (float64, error) {
	value, ok := example[key]
	if !ok {
		return 0, fmt.Errorf("missing example value %s", key)
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %s=%q", key, value)
	}
	return parsed, nil
}
