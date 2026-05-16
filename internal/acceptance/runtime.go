package acceptance

import (
	"fmt"

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
	"the acceptance pipeline task is accepted":                                     acceptStep,
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
	"the coder checks out the committed project":                                   acceptStep,
	"the acceptance test command should pass without uncommitted setup steps":      assertAcceptanceCommandPassesFromCleanCheckout,
	"acceptance smoke is ready":                                                    acceptStep,
	"acceptance smoke should pass":                                                 acceptStep,
	"the project skeleton task is accepted":                                        acceptStep,
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

func acceptStep(*world, map[string]string) error {
	return nil
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
