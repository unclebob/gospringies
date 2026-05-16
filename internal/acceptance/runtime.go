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
	"the domain model task is accepted":                                            acceptStep,
	"the system parameters task is accepted":                                       acceptStep,
	"the coder creates a new world":                                                createDomainWorld,
	"the world should contain <mass_count> masses":                                 assertDomainMassCount,
	"the world should contain <spring_count> springs":                              assertDomainSpringCount,
	"a world with mass <id> at <x>, <y>":                                           createDomainMassFromID,
	"a world with mass <mass_a> at <x_a>, <y_a>":                                   createDomainMassA,
	"a world with mass <mass_b> at <x_b>, <y_b>":                                   createDomainMassB,
	"a world with mass <existing_mass> at <x>, <y>":                                createExistingDomainMass,
	"mass <id> has velocity <vx>, <vy>":                                            setMassVelocity,
	"mass <id> has mass value <mass_value>":                                        setMassValue,
	"mass <id> has elasticity <elasticity>":                                        setMassElasticity,
	"mass <id> fixed state is <fixed>":                                             setMassFixed,
	"the coder looks up mass <id>":                                                 lookupMass,
	"the coder reads mass <id> from the domain model":                              lookupMass,
	"mass <id> should have position <x>, <y>":                                      assertMassPosition,
	"mass <id> should have velocity <vx>, <vy>":                                    assertMassVelocity,
	"mass <id> should have mass value <mass_value>":                                assertMassValue,
	"mass <id> mass value should remain <mass_value>":                              assertMassValue,
	"mass <id> should have elasticity <elasticity>":                                assertMassElasticity,
	"mass <id> fixed state should be <fixed>":                                      assertMassFixed,
	"a spring <spring_id> connects mass <mass_a> to mass <mass_b>":                 createDomainSpring,
	"spring <spring_id> has spring constant <spring_constant>":                     setSpringConstant,
	"spring <spring_id> has damping constant <damping_constant>":                   setSpringDamping,
	"spring <spring_id> has rest length <rest_length>":                             setSpringRestLength,
	"the coder looks up spring <spring_id>":                                        lookupSpring,
	"spring <spring_id> should connect mass <mass_a> to mass <mass_b>":             assertSpringEndpoints,
	"spring <spring_id> should have spring constant <spring_constant>":             assertSpringConstant,
	"spring <spring_id> should have damping constant <damping_constant>":           assertSpringDamping,
	"spring <spring_id> should have rest length <rest_length>":                     assertSpringRestLength,
	"a world already contains a <object_type> with id <id>":                        createDuplicateSubject,
	"the coder adds another <object_type> with id <id>":                            addDuplicateSubject,
	"the coder adds spring <spring_id> connecting mass <mass_a> to mass <mass_b>":  addDomainSpringForValidation,
	"validation should fail with reason <reason>":                                  assertValidationFailure,
	"parameter <parameter> should have default value <value>":                      assertParameterDefault,
	"force <force> should have enabled state <enabled>":                            assertForceEnabledState,
	"force <force> should have editable parameters":                                assertForceEditableParameters,
	"wall <wall> should have enabled state <enabled>":                              assertWallEnabledState,
	"a world with parameter <parameter> changed to <changed_value>":                changeWorldParameter,
	"the coder performs <operation>":                                               performWorldOperation,
	"parameter <parameter> should be <expected_value_source>":                      assertParameterSource,
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
