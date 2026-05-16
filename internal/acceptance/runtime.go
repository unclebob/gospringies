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
	parserRan      bool
	generatorRan   bool
	generatedRan   bool
	generated      bool
	smokeAdded     bool
	smokeParsed    bool
	smokeGenerated bool
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
	case "the acceptance pipeline task is accepted":
		return nil
	case "the coder runs the acceptance test command":
		if err := runPipeline("features/pipeline_smoke.feature", "pipeline_command"); err != nil {
			return err
		}
		w.parserRan = true
		w.generatorRan = true
		w.generatedRan = true
		return nil
	case "the Gherkin parser should run successfully":
		if !w.parserRan {
			return fmt.Errorf("gherkin parser did not run successfully")
		}
		return nil
	case "the acceptance test generator should run successfully":
		if !w.generatorRan {
			return fmt.Errorf("acceptance generator did not run successfully")
		}
		return nil
	case "the generated executable acceptance tests should run successfully":
		if !w.generatedRan {
			return fmt.Errorf("generated executable acceptance tests did not run successfully")
		}
		return nil
	case "the coder generates acceptance tests":
		if err := runParser("features/pipeline_smoke.feature", "build/acceptance/pipeline_artifacts.json"); err != nil {
			return err
		}
		if err := runGenerator("build/acceptance/pipeline_artifacts.json", "acceptance/generated/pipeline_artifacts_acceptance_test.go"); err != nil {
			return err
		}
		w.generated = true
		return nil
	case "generated acceptance <artifact> should be written under <generated_location>":
		if !w.generated {
			return fmt.Errorf("acceptance tests have not been generated")
		}
		artifact, err := stringValue(example, "artifact")
		if err != nil {
			return err
		}
		location, err := stringValue(example, "generated_location")
		if err != nil {
			return err
		}
		return generatedArtifactExists(artifact, location)
	case "hand-written <test_type> tests should remain outside <generated_location>":
		testType, err := stringValue(example, "test_type")
		if err != nil {
			return err
		}
		if strings.TrimSpace(testType) != "unit" {
			return fmt.Errorf("unsupported hand-written test type %q", testType)
		}
		location, err := stringValue(example, "generated_location")
		if err != nil {
			return err
		}
		return handwrittenTestsOutside(location)
	case "the coder adds a minimal smoke feature":
		if _, err := os.Stat(repoPath("features/pipeline_smoke.feature")); err != nil {
			return err
		}
		w.smokeAdded = true
		return nil
	case "the smoke feature should parse successfully":
		if !w.smokeAdded {
			return fmt.Errorf("smoke feature has not been added")
		}
		if err := runParser("features/pipeline_smoke.feature", "build/_acceptance-pipeline/smoke/feature.json"); err != nil {
			return err
		}
		w.smokeParsed = true
		return nil
	case "the smoke feature should generate an executable acceptance test":
		if !w.smokeParsed {
			return fmt.Errorf("smoke feature has not been parsed")
		}
		if err := runGenerator("build/_acceptance-pipeline/smoke/feature.json", "build/_acceptance-pipeline/smoke/generated/pipeline_smoke_acceptance_test.go"); err != nil {
			return err
		}
		w.smokeGenerated = true
		return nil
	case "the generated smoke acceptance test should pass":
		if !w.smokeGenerated {
			return fmt.Errorf("smoke acceptance test has not been generated")
		}
		return runCommand("go", "test", "./build/_acceptance-pipeline/smoke/generated")
	case "the coder checks out the committed project":
		return nil
	case "the acceptance test command should pass without uncommitted setup steps":
		return runCommandWithEnv([]string{
			"ACCEPTANCE_BUILD_DIR=build/_acceptance-pipeline/clean",
			"ACCEPTANCE_GENERATED_DIR=build/_acceptance-pipeline/clean/generated",
		}, "./scripts/acceptance.sh", "features/pipeline_smoke.feature")
	case "acceptance smoke is ready":
		return nil
	case "acceptance smoke should pass":
		return nil
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
	return runCommandWithEnv(nil, name, args...)
}

func runCommandWithEnv(env []string, name string, args ...string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = root
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
	if err := runParser(feature, jsonPath); err != nil {
		return err
	}
	if err := runGenerator(jsonPath, generatedPath); err != nil {
		return err
	}
	return runCommand("go", "test", "./build/_acceptance-pipeline/"+base+"/generated")
}

func runParser(feature, output string) error {
	return runCommand("go", "run", "./cmd/gherkin-parser", feature, output)
}

func runGenerator(jsonPath, output string) error {
	return runCommand("go", "run", "./cmd/acceptance-generator", jsonPath, output)
}

func generatedArtifactExists(artifact, location string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	var path string
	switch artifact {
	case "test source":
		path = filepath.Join(root, location, "pipeline_artifacts_acceptance_test.go")
	case "parsed feature":
		path = filepath.Join(root, location, "pipeline_artifacts.json")
	default:
		return fmt.Errorf("unsupported generated artifact %q", artifact)
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func handwrittenTestsOutside(location string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	generatedLocation := filepath.Clean(filepath.Join(root, location))
	var violations []string
	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			if strings.HasPrefix(filepath.Clean(path), generatedLocation) {
				violations = append(violations, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if len(violations) > 0 {
		return fmt.Errorf("hand-written tests under generated location: %s", strings.Join(violations, ", "))
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
