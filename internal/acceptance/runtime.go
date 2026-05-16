package acceptance

import (
	"errors"
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
	domainWorld    *sim.Simulation
	lookedMass     sim.Mass
	lookedSpring   sim.Spring
	validationErr  error
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
	case "the domain model task is accepted":
		return nil
	case "the coder creates a new world":
		w.domainWorld = sim.NewWorld()
		return nil
	case "the world should contain <mass_count> masses":
		world, err := domainWorld(w)
		if err != nil {
			return err
		}
		expected, err := intValue(example, "mass_count")
		if err != nil {
			return err
		}
		if len(world.Masses) != expected {
			return fmt.Errorf("expected %d masses, got %d", expected, len(world.Masses))
		}
		return nil
	case "the world should contain <spring_count> springs":
		world, err := domainWorld(w)
		if err != nil {
			return err
		}
		expected, err := intValue(example, "spring_count")
		if err != nil {
			return err
		}
		if len(world.Springs) != expected {
			return fmt.Errorf("expected %d springs, got %d", expected, len(world.Springs))
		}
		return nil
	case "a world with mass <id> at <x>, <y>", "a world with mass <mass_a> at <x_a>, <y_a>", "a world with mass <mass_b> at <x_b>, <y_b>", "a world with mass <existing_mass> at <x>, <y>":
		world := ensureDomainWorld(w)
		id, x, y, err := massFromStep(step.Text, example)
		if err != nil {
			return err
		}
		if _, ok := world.MassByID(id); ok {
			return nil
		}
		return world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: x, Y: y}, Mass: 1})
	case "mass <id> has velocity <vx>, <vy>":
		return updateMass(w, example, func(mass *sim.Mass) error {
			vx, err := floatValue(example, "vx")
			if err != nil {
				return err
			}
			vy, err := floatValue(example, "vy")
			if err != nil {
				return err
			}
			mass.Velocity = sim.Vec2{X: vx, Y: vy}
			return nil
		})
	case "mass <id> has mass value <mass_value>":
		return updateMass(w, example, func(mass *sim.Mass) error {
			value, err := floatValue(example, "mass_value")
			if err != nil {
				return err
			}
			mass.Mass = value
			return nil
		})
	case "mass <id> has elasticity <elasticity>":
		return updateMass(w, example, func(mass *sim.Mass) error {
			value, err := floatValue(example, "elasticity")
			if err != nil {
				return err
			}
			mass.Elasticity = value
			return nil
		})
	case "mass <id> fixed state is <fixed>":
		return updateMass(w, example, func(mass *sim.Mass) error {
			value, err := boolValue(example, "fixed")
			if err != nil {
				return err
			}
			mass.Fixed = value
			return nil
		})
	case "the coder looks up mass <id>", "the coder reads mass <id> from the domain model":
		world, err := domainWorld(w)
		if err != nil {
			return err
		}
		id, err := intValue(example, "id")
		if err != nil {
			return err
		}
		mass, ok := world.MassByID(id)
		if !ok {
			return fmt.Errorf("mass %d not found", id)
		}
		w.lookedMass = mass
		return nil
	case "mass <id> should have position <x>, <y>":
		x, err := floatValue(example, "x")
		if err != nil {
			return err
		}
		y, err := floatValue(example, "y")
		if err != nil {
			return err
		}
		return assertVec("position", w.lookedMass.Position, x, y)
	case "mass <id> should have velocity <vx>, <vy>":
		vx, err := floatValue(example, "vx")
		if err != nil {
			return err
		}
		vy, err := floatValue(example, "vy")
		if err != nil {
			return err
		}
		return assertVec("velocity", w.lookedMass.Velocity, vx, vy)
	case "mass <id> should have mass value <mass_value>", "mass <id> mass value should remain <mass_value>":
		expected, err := floatValue(example, "mass_value")
		if err != nil {
			return err
		}
		return assertFloat("mass value", w.lookedMass.Mass, expected)
	case "mass <id> should have elasticity <elasticity>":
		expected, err := floatValue(example, "elasticity")
		if err != nil {
			return err
		}
		return assertFloat("elasticity", w.lookedMass.Elasticity, expected)
	case "mass <id> fixed state should be <fixed>":
		expected, err := boolValue(example, "fixed")
		if err != nil {
			return err
		}
		if w.lookedMass.Fixed != expected {
			return fmt.Errorf("expected fixed %t, got %t", expected, w.lookedMass.Fixed)
		}
		return nil
	case "a spring <spring_id> connects mass <mass_a> to mass <mass_b>":
		world := ensureDomainWorld(w)
		spring, err := springFromExample(example)
		if err != nil {
			return err
		}
		if _, ok := world.SpringByID(spring.ID); ok {
			return nil
		}
		return world.AddSpring(spring)
	case "spring <spring_id> has spring constant <spring_constant>":
		return updateSpring(w, example, func(spring *sim.Spring) error {
			value, err := floatValue(example, "spring_constant")
			if err != nil {
				return err
			}
			spring.SpringConstant = value
			spring.Stiffness = value
			return nil
		})
	case "spring <spring_id> has damping constant <damping_constant>":
		return updateSpring(w, example, func(spring *sim.Spring) error {
			value, err := floatValue(example, "damping_constant")
			if err != nil {
				return err
			}
			spring.Damping = value
			return nil
		})
	case "spring <spring_id> has rest length <rest_length>":
		return updateSpring(w, example, func(spring *sim.Spring) error {
			value, err := floatValue(example, "rest_length")
			if err != nil {
				return err
			}
			spring.RestLength = value
			return nil
		})
	case "the coder looks up spring <spring_id>":
		world, err := domainWorld(w)
		if err != nil {
			return err
		}
		id, err := intValue(example, "spring_id")
		if err != nil {
			return err
		}
		spring, ok := world.SpringByID(id)
		if !ok {
			return fmt.Errorf("spring %d not found", id)
		}
		w.lookedSpring = spring
		return nil
	case "spring <spring_id> should connect mass <mass_a> to mass <mass_b>":
		massA, err := intValue(example, "mass_a")
		if err != nil {
			return err
		}
		massB, err := intValue(example, "mass_b")
		if err != nil {
			return err
		}
		if w.lookedSpring.MassA != massA || w.lookedSpring.MassB != massB {
			return fmt.Errorf("expected spring endpoints %d,%d got %d,%d", massA, massB, w.lookedSpring.MassA, w.lookedSpring.MassB)
		}
		return nil
	case "spring <spring_id> should have spring constant <spring_constant>":
		expected, err := floatValue(example, "spring_constant")
		if err != nil {
			return err
		}
		return assertFloat("spring constant", w.lookedSpring.SpringConstant, expected)
	case "spring <spring_id> should have damping constant <damping_constant>":
		expected, err := floatValue(example, "damping_constant")
		if err != nil {
			return err
		}
		return assertFloat("damping constant", w.lookedSpring.Damping, expected)
	case "spring <spring_id> should have rest length <rest_length>":
		expected, err := floatValue(example, "rest_length")
		if err != nil {
			return err
		}
		return assertFloat("rest length", w.lookedSpring.RestLength, expected)
	case "a world already contains a <object_type> with id <id>":
		world := ensureDomainWorld(w)
		objectType, err := stringValue(example, "object_type")
		if err != nil {
			return err
		}
		id, err := intValue(example, "id")
		if err != nil {
			return err
		}
		if objectType == "mass" {
			return world.AddMass(sim.Mass{ID: id, Mass: 1})
		}
		if err := world.AddMass(sim.Mass{ID: 1, Mass: 1}); err != nil {
			return err
		}
		if err := world.AddMass(sim.Mass{ID: 2, Mass: 1}); err != nil {
			return err
		}
		return world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
	case "the coder adds another <object_type> with id <id>":
		world := ensureDomainWorld(w)
		objectType, err := stringValue(example, "object_type")
		if err != nil {
			return err
		}
		id, err := intValue(example, "id")
		if err != nil {
			return err
		}
		if objectType == "mass" {
			w.validationErr = world.AddMass(sim.Mass{ID: id, Mass: 1})
		} else {
			w.validationErr = world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
		}
		return nil
	case "the coder adds spring <spring_id> connecting mass <mass_a> to mass <mass_b>":
		world := ensureDomainWorld(w)
		spring, err := springFromExample(example)
		if err != nil {
			return err
		}
		w.validationErr = world.AddSpring(spring)
		return nil
	case "validation should fail with reason <reason>":
		reason, err := stringValue(example, "reason")
		if err != nil {
			return err
		}
		return assertValidationReason(w.validationErr, reason)
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

func ensureDomainWorld(w *world) *sim.Simulation {
	if w.domainWorld == nil {
		w.domainWorld = sim.NewWorld()
	}
	return w.domainWorld
}

func domainWorld(w *world) (*sim.Simulation, error) {
	if w.domainWorld == nil {
		return nil, fmt.Errorf("domain world has not been created")
	}
	return w.domainWorld, nil
}

func updateMass(w *world, example map[string]string, update func(*sim.Mass) error) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, "id")
	if err != nil {
		return err
	}
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			return update(&world.Masses[i])
		}
	}
	return fmt.Errorf("mass %d not found", id)
}

func updateSpring(w *world, example map[string]string, update func(*sim.Spring) error) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	for i := range world.Springs {
		if world.Springs[i].ID == id {
			return update(&world.Springs[i])
		}
	}
	return fmt.Errorf("spring %d not found", id)
}

func springFromExample(example map[string]string) (sim.Spring, error) {
	id, err := intValue(example, "spring_id")
	if err != nil {
		return sim.Spring{}, err
	}
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return sim.Spring{}, err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return sim.Spring{}, err
	}
	return sim.Spring{ID: id, MassA: massA, MassB: massB}, nil
}

func assertValidationReason(err error, reason string) error {
	if err == nil {
		return fmt.Errorf("validation succeeded, expected %s", reason)
	}
	switch strings.TrimSpace(reason) {
	case "duplicate id":
		if errors.Is(err, sim.ErrDuplicateID) {
			return nil
		}
	case "missing spring endpoint":
		if errors.Is(err, sim.ErrMissingSpringEndpoint) {
			return nil
		}
	}
	return fmt.Errorf("expected validation reason %q, got %v", reason, err)
}

func assertVec(name string, got sim.Vec2, expectedX, expectedY float64) error {
	if math.Abs(got.X-expectedX) > 0.000001 || math.Abs(got.Y-expectedY) > 0.000001 {
		return fmt.Errorf("expected %s %f,%f got %f,%f", name, expectedX, expectedY, got.X, got.Y)
	}
	return nil
}

func assertFloat(name string, got, expected float64) error {
	if math.Abs(got-expected) > 0.000001 {
		return fmt.Errorf("expected %s %f got %f", name, expected, got)
	}
	return nil
}

func firstFloat(example map[string]string, keys ...string) (float64, error) {
	for _, key := range keys {
		if _, ok := example[key]; ok {
			return floatValue(example, key)
		}
	}
	return 0, fmt.Errorf("missing example value among %s", strings.Join(keys, ", "))
}

func massFromStep(stepText string, example map[string]string) (int, float64, float64, error) {
	switch stepText {
	case "a world with mass <id> at <x>, <y>":
		return massFields(example, "id", "x", "y")
	case "a world with mass <mass_a> at <x_a>, <y_a>":
		return massFields(example, "mass_a", "x_a", "y_a")
	case "a world with mass <mass_b> at <x_b>, <y_b>":
		return massFields(example, "mass_b", "x_b", "y_b")
	case "a world with mass <existing_mass> at <x>, <y>":
		return massFields(example, "existing_mass", "x", "y")
	default:
		return 0, 0, 0, fmt.Errorf("unsupported mass step %q", stepText)
	}
}

func massFields(example map[string]string, idKey, xKey, yKey string) (int, float64, float64, error) {
	id, err := intValue(example, idKey)
	if err != nil {
		return 0, 0, 0, err
	}
	x, err := floatValue(example, xKey)
	if err != nil {
		return 0, 0, 0, err
	}
	y, err := floatValue(example, yKey)
	if err != nil {
		return 0, 0, 0, err
	}
	return id, x, y, nil
}

func boolValue(example map[string]string, key string) (bool, error) {
	value, ok := example[key]
	if !ok {
		return false, fmt.Errorf("missing example value %s", key)
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %s=%q", key, value)
	}
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
