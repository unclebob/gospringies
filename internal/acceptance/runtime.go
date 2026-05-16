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
	"the domain model task is accepted":                                            acceptProjectSkeleton,
	"the coder creates a new world":                                                createDomainWorld,
	"the world should contain <mass_count> masses":                                 assertDomainMassCount,
	"the world should contain <spring_count> springs":                              assertDomainSpringCount,
	"a world with mass <id> at <x>, <y>":                                           addDomainMass,
	"a world with mass <mass_a> at <x_a>, <y_a>":                                   addDomainMassA,
	"a world with mass <mass_b> at <x_b>, <y_b>":                                   addDomainMassB,
	"a world with mass <existing_mass> at <x>, <y>":                                addExistingDomainMass,
	"mass <id> has velocity <vx>, <vy>":                                            setDomainMassVelocity,
	"mass <id> has mass value <mass_value>":                                        setDomainMassValue,
	"mass <id> has elasticity <elasticity>":                                        setDomainMassElasticity,
	"mass <id> fixed state is <fixed>":                                             setDomainMassFixed,
	"the coder looks up mass <id>":                                                 lookupDomainMass,
	"the coder reads mass <id> from the domain model":                              lookupDomainMass,
	"mass <id> should have position <x>, <y>":                                      assertDomainMassPosition,
	"mass <id> should have velocity <vx>, <vy>":                                    assertDomainMassVelocity,
	"mass <id> should have mass value <mass_value>":                                assertDomainMassValue,
	"mass <id> mass value should remain <mass_value>":                              assertDomainMassValue,
	"mass <id> should have elasticity <elasticity>":                                assertDomainMassElasticity,
	"mass <id> fixed state should be <fixed>":                                      assertDomainMassFixed,
	"a spring <spring_id> connects mass <mass_a> to mass <mass_b>":                 addDomainSpring,
	"spring <spring_id> has spring constant <spring_constant>":                     setDomainSpringConstant,
	"spring <spring_id> has damping constant <damping_constant>":                   setDomainSpringDamping,
	"spring <spring_id> has rest length <rest_length>":                             setDomainSpringRestLength,
	"the coder looks up spring <spring_id>":                                        lookupDomainSpring,
	"spring <spring_id> should connect mass <mass_a> to mass <mass_b>":             assertDomainSpringEndpoints,
	"spring <spring_id> should have spring constant <spring_constant>":             assertDomainSpringConstant,
	"spring <spring_id> should have damping constant <damping_constant>":           assertDomainSpringDamping,
	"spring <spring_id> should have rest length <rest_length>":                     assertDomainSpringRestLength,
	"a world already contains a <object_type> with id <id>":                        addExistingDomainObject,
	"the coder adds another <object_type> with id <id>":                            addDuplicateDomainObject,
	"the coder adds spring <spring_id> connecting mass <mass_a> to mass <mass_b>":  addInvalidDomainSpring,
	"validation should fail with reason <reason>":                                  assertDomainValidationReason,
	"the acceptance pipeline task is accepted":                                     acceptProjectSkeleton,
	"the coder runs the acceptance test command":                                   runAcceptanceCommand,
	"the Gherkin parser should run successfully":                                   assertParserRan,
	"the acceptance test generator should run successfully":                        assertGeneratorRan,
	"the generated executable acceptance tests should run successfully":            assertGeneratedRan,
	"the coder generates acceptance tests":                                         generateAcceptanceArtifacts,
	"generated acceptance <artifact> should be written under <generated_location>": assertGeneratedArtifactExists,
	"hand-written <test_type> tests should remain outside <generated_location>":    assertHandwrittenTestsOutside,
	"the coder adds a minimal smoke feature":                                       addSmokeFeature,
	"the smoke feature should parse successfully":                                  parseSmokeFeature,
	"the smoke feature should generate an executable acceptance test":              generateSmokeAcceptanceTest,
	"the generated smoke acceptance test should pass":                              assertSmokeAcceptanceTestPasses,
	"the coder checks out the committed project":                                   acceptProjectSkeleton,
	"the acceptance test command should pass without uncommitted setup steps":      assertAcceptanceCommandPassesFromCleanCheckout,
	"acceptance smoke is ready":                                                    acceptProjectSkeleton,
	"acceptance smoke should pass":                                                 acceptProjectSkeleton,
	"the project skeleton task is accepted":                                        acceptProjectSkeleton,
	"the coder creates the initial Go package layout":                              createPackageLayout,
	"the <package> package should not import <graphics_library>":                   assertPackageDoesNotImport,
	"the coder creates the desktop application command":                            createApplicationCommand,
	"the application command should build successfully":                            assertApplicationCommandBuilds,
	"the coder creates the initial Go module":                                      createGoModule,
	"the Go test suite should pass":                                                assertGoTestsPass,
	"a demo spring simulation":                                                     createDemoSimulation,
	"I advance the simulation <steps> steps":                                       advanceSimulation,
	"mass <mass> x should be <x>":                                                  assertMassX,
}

func acceptProjectSkeleton(*world, map[string]string) error {
	return nil
}

func createDomainWorld(w *world, _ map[string]string) error {
	w.domainWorld = sim.NewWorld()
	return nil
}

func assertDomainMassCount(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	expected, err := intValue(example, "mass_count")
	if err != nil {
		return err
	}
	return assertCount("masses", len(world.Masses), expected)
}

func assertDomainSpringCount(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	expected, err := intValue(example, "spring_count")
	if err != nil {
		return err
	}
	return assertCount("springs", len(world.Springs), expected)
}

func assertCount(name string, got, expected int) error {
	if got != expected {
		return fmt.Errorf("expected %d %s, got %d", expected, name, got)
	}
	return nil
}

func addDomainMass(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "id", "x", "y")
}

func addDomainMassA(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "mass_a", "x_a", "y_a")
}

func addDomainMassB(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "mass_b", "x_b", "y_b")
}

func addExistingDomainMass(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "existing_mass", "x", "y")
}

func addDomainMassFromKeys(w *world, example map[string]string, idKey, xKey, yKey string) error {
	world := ensureDomainWorld(w)
	id, x, y, err := massFields(example, idKey, xKey, yKey)
	if err != nil {
		return err
	}
	if _, ok := world.MassByID(id); ok {
		return nil
	}
	return world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: x, Y: y}, Mass: 1})
}

func setDomainMassVelocity(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		velocity, err := vecFromExample(example, "vx", "vy")
		if err != nil {
			return err
		}
		mass.Velocity = velocity
		return nil
	})
}

func setDomainMassValue(w *world, example map[string]string) error {
	return updateMassFloat(w, example, "mass_value", func(mass *sim.Mass, value float64) {
		mass.Mass = value
	})
}

func setDomainMassElasticity(w *world, example map[string]string) error {
	return updateMassFloat(w, example, "elasticity", func(mass *sim.Mass, value float64) {
		mass.Elasticity = value
	})
}

func setDomainMassFixed(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := boolValue(example, "fixed")
		if err != nil {
			return err
		}
		mass.Fixed = value
		return nil
	})
}

func updateMassFloat(w *world, example map[string]string, key string, assign func(*sim.Mass, float64)) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		assign(mass, value)
		return nil
	})
}

func lookupDomainMass(w *world, example map[string]string) error {
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
}

func assertDomainMassPosition(w *world, example map[string]string) error {
	expected, err := vecFromExample(example, "x", "y")
	if err != nil {
		return err
	}
	return assertVec("position", w.lookedMass.Position, expected.X, expected.Y)
}

func assertDomainMassVelocity(w *world, example map[string]string) error {
	expected, err := vecFromExample(example, "vx", "vy")
	if err != nil {
		return err
	}
	return assertVec("velocity", w.lookedMass.Velocity, expected.X, expected.Y)
}

func vecFromExample(example map[string]string, xKey, yKey string) (sim.Vec2, error) {
	x, err := floatValue(example, xKey)
	if err != nil {
		return sim.Vec2{}, err
	}
	y, err := floatValue(example, yKey)
	if err != nil {
		return sim.Vec2{}, err
	}
	return sim.Vec2{X: x, Y: y}, nil
}

func assertDomainMassValue(w *world, example map[string]string) error {
	return assertFloatExample("mass value", w.lookedMass.Mass, example, "mass_value")
}

func assertDomainMassElasticity(w *world, example map[string]string) error {
	return assertFloatExample("elasticity", w.lookedMass.Elasticity, example, "elasticity")
}

func assertFloatExample(name string, got float64, example map[string]string, key string) error {
	expected, err := floatValue(example, key)
	if err != nil {
		return err
	}
	return assertFloat(name, got, expected)
}

func assertDomainMassFixed(w *world, example map[string]string) error {
	expected, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	if w.lookedMass.Fixed != expected {
		return fmt.Errorf("expected fixed %t, got %t", expected, w.lookedMass.Fixed)
	}
	return nil
}

func addDomainSpring(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	spring, err := springFromExample(example)
	if err != nil {
		return err
	}
	if _, ok := world.SpringByID(spring.ID); ok {
		return nil
	}
	return world.AddSpring(spring)
}

func setDomainSpringConstant(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "spring_constant", func(spring *sim.Spring, value float64) {
		spring.SpringConstant = value
		spring.Stiffness = value
	})
}

func setDomainSpringDamping(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "damping_constant", func(spring *sim.Spring, value float64) {
		spring.Damping = value
	})
}

func setDomainSpringRestLength(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "rest_length", func(spring *sim.Spring, value float64) {
		spring.RestLength = value
	})
}

func updateSpringFloat(w *world, example map[string]string, key string, assign func(*sim.Spring, float64)) error {
	return updateSpring(w, example, func(spring *sim.Spring) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		assign(spring, value)
		return nil
	})
}

func lookupDomainSpring(w *world, example map[string]string) error {
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
}

func assertDomainSpringEndpoints(w *world, example map[string]string) error {
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
}

func assertDomainSpringConstant(w *world, example map[string]string) error {
	return assertFloatExample("spring constant", w.lookedSpring.SpringConstant, example, "spring_constant")
}

func assertDomainSpringDamping(w *world, example map[string]string) error {
	return assertFloatExample("damping constant", w.lookedSpring.Damping, example, "damping_constant")
}

func assertDomainSpringRestLength(w *world, example map[string]string) error {
	return assertFloatExample("rest length", w.lookedSpring.RestLength, example, "rest_length")
}

func addExistingDomainObject(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	if objectType == "mass" {
		return world.AddMass(sim.Mass{ID: id, Mass: 1})
	}
	return addExistingSpring(world, id)
}

func addDuplicateDomainObject(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	if objectType == "mass" {
		w.validationErr = world.AddMass(sim.Mass{ID: id, Mass: 1})
	} else {
		w.validationErr = world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
	}
	return nil
}

func objectTypeAndID(example map[string]string) (string, int, error) {
	objectType, err := stringValue(example, "object_type")
	if err != nil {
		return "", 0, err
	}
	id, err := intValue(example, "id")
	if err != nil {
		return "", 0, err
	}
	return objectType, id, nil
}

func addExistingSpring(world *sim.Simulation, id int) error {
	if err := world.AddMass(sim.Mass{ID: 1, Mass: 1}); err != nil {
		return err
	}
	if err := world.AddMass(sim.Mass{ID: 2, Mass: 1}); err != nil {
		return err
	}
	return world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
}

func addInvalidDomainSpring(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	spring, err := springFromExample(example)
	if err != nil {
		return err
	}
	w.validationErr = world.AddSpring(spring)
	return nil
}

func assertDomainValidationReason(w *world, example map[string]string) error {
	reason, err := stringValue(example, "reason")
	if err != nil {
		return err
	}
	return assertValidationReason(w.validationErr, reason)
}

func runAcceptanceCommand(w *world, _ map[string]string) error {
	if err := runPipeline("features/pipeline_smoke.feature", "pipeline_command"); err != nil {
		return err
	}
	w.parserRan = true
	w.generatorRan = true
	w.generatedRan = true
	return nil
}

func assertParserRan(w *world, _ map[string]string) error {
	if !w.parserRan {
		return fmt.Errorf("gherkin parser did not run successfully")
	}
	return nil
}

func assertGeneratorRan(w *world, _ map[string]string) error {
	if !w.generatorRan {
		return fmt.Errorf("acceptance generator did not run successfully")
	}
	return nil
}

func assertGeneratedRan(w *world, _ map[string]string) error {
	if !w.generatedRan {
		return fmt.Errorf("generated executable acceptance tests did not run successfully")
	}
	return nil
}

func generateAcceptanceArtifacts(w *world, _ map[string]string) error {
	if err := runParserAndGenerator(
		"features/pipeline_smoke.feature",
		"build/acceptance/pipeline_artifacts.json",
		"acceptance/generated/pipeline_artifacts_acceptance_test.go",
	); err != nil {
		return err
	}
	w.generated = true
	return nil
}

func assertGeneratedArtifactExists(w *world, example map[string]string) error {
	if !w.generated {
		return fmt.Errorf("acceptance tests have not been generated")
	}
	artifact, location, err := artifactExample(example)
	if err != nil {
		return err
	}
	return generatedArtifactExists(artifact, location)
}

func artifactExample(example map[string]string) (string, string, error) {
	artifact, err := stringValue(example, "artifact")
	if err != nil {
		return "", "", err
	}
	location, err := stringValue(example, "generated_location")
	if err != nil {
		return "", "", err
	}
	return artifact, location, nil
}

func assertHandwrittenTestsOutside(_ *world, example map[string]string) error {
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
}

func addSmokeFeature(w *world, _ map[string]string) error {
	if _, err := os.Stat(repoPath("features/pipeline_smoke.feature")); err != nil {
		return err
	}
	w.smokeAdded = true
	return nil
}

func parseSmokeFeature(w *world, _ map[string]string) error {
	if !w.smokeAdded {
		return fmt.Errorf("smoke feature has not been added")
	}
	return markSmokeParsed(w, runParser("features/pipeline_smoke.feature", "build/_acceptance-pipeline/smoke/feature.json"))
}

func markSmokeParsed(w *world, err error) error {
	if err != nil {
		return err
	}
	w.smokeParsed = true
	return nil
}

func generateSmokeAcceptanceTest(w *world, _ map[string]string) error {
	if !w.smokeParsed {
		return fmt.Errorf("smoke feature has not been parsed")
	}
	return markSmokeGenerated(w, runGenerator("build/_acceptance-pipeline/smoke/feature.json", "build/_acceptance-pipeline/smoke/generated/pipeline_smoke_acceptance_test.go"))
}

func markSmokeGenerated(w *world, err error) error {
	if err != nil {
		return err
	}
	w.smokeGenerated = true
	return nil
}

func assertSmokeAcceptanceTestPasses(w *world, _ map[string]string) error {
	if !w.smokeGenerated {
		return fmt.Errorf("smoke acceptance test has not been generated")
	}
	return runCommand("go", "test", "./build/_acceptance-pipeline/smoke/generated")
}

func assertAcceptanceCommandPassesFromCleanCheckout(*world, map[string]string) error {
	return runCommandWithEnv([]string{
		"ACCEPTANCE_BUILD_DIR=build/_acceptance-pipeline/clean",
		"ACCEPTANCE_GENERATED_DIR=build/_acceptance-pipeline/clean/generated",
	}, "./scripts/acceptance.sh", "features/pipeline_smoke.feature")
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

func generatedArtifactExists(artifact, location string) error {
	path, err := generatedArtifactPath(artifact, location)
	if err != nil {
		return err
	}
	return fileExists(path)
}

func generatedArtifactPath(artifact, location string) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	name, ok := generatedArtifactNames[artifact]
	if !ok {
		return "", fmt.Errorf("unsupported generated artifact %q", artifact)
	}
	return filepath.Join(root, location, name), nil
}

var generatedArtifactNames = map[string]string{
	"test source":    "pipeline_artifacts_acceptance_test.go",
	"parsed feature": "pipeline_artifacts.json",
}

func fileExists(path string) error {
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
		dirViolations, err := handwrittenTestViolations(filepath.Join(root, dir), generatedLocation)
		if err != nil {
			return err
		}
		violations = append(violations, dirViolations...)
	}
	return reportHandwrittenViolations(violations)
}

func handwrittenTestViolations(root, generatedLocation string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if isHandwrittenTestUnder(path, entry, generatedLocation) {
			violations = append(violations, path)
		}
		return nil
	})
	return violations, err
}

func isHandwrittenTestUnder(path string, entry os.DirEntry, generatedLocation string) bool {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
		return false
	}
	return strings.HasPrefix(filepath.Clean(path), generatedLocation)
}

func reportHandwrittenViolations(violations []string) error {
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
