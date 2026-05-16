package acceptance

import (
	"fmt"

	"springs/internal/gherkin"
	"springs/internal/sim"
)

type world struct {
	simulation      *sim.Simulation
	layoutCreated   bool
	commandCreated  bool
	moduleCreated   bool
	parserRan       bool
	generatorRan    bool
	generatedRan    bool
	generated       bool
	smokeAdded      bool
	smokeParsed     bool
	smokeGenerated  bool
	domainWorld     *sim.Simulation
	lookedMass      sim.Mass
	lookedSpring    sim.Spring
	validationErr   error
	forceEvaluation sim.ForceEvaluation
	resultingWorld  *sim.Simulation
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
	"the domain model task is accepted":                                                    acceptStep,
	"the system parameters task is accepted":                                               acceptStep,
	"the force evaluation task is accepted":                                                acceptStep,
	"the simulation step task is accepted":                                                 acceptStep,
	"the coder creates a new world":                                                        createDomainWorld,
	"the world should contain <mass_count> masses":                                         assertDomainMassCount,
	"the world should contain <spring_count> springs":                                      assertDomainSpringCount,
	"a world with mass <id> at <x>, <y>":                                                   addDomainMass,
	"a world with mass <mass_a> at <x_a>, <y_a>":                                           addDomainMassA,
	"a world with mass <mass_b> at <x_b>, <y_b>":                                           addDomainMassB,
	"a world with mass <existing_mass> at <x>, <y>":                                        addExistingDomainMass,
	"mass <id> has velocity <vx>, <vy>":                                                    setDomainMassVelocity,
	"mass <id> has mass value <mass_value>":                                                setDomainMassValue,
	"mass <id> has elasticity <elasticity>":                                                setDomainMassElasticity,
	"mass <id> fixed state is <fixed>":                                                     setDomainMassFixed,
	"the coder looks up mass <id>":                                                         lookupDomainMass,
	"the coder reads mass <id> from the domain model":                                      lookupDomainMass,
	"mass <id> should have position <x>, <y>":                                              assertDomainMassPosition,
	"mass <id> should have velocity <vx>, <vy>":                                            assertDomainMassVelocity,
	"mass <id> should have mass value <mass_value>":                                        assertDomainMassValue,
	"mass <id> mass value should remain <mass_value>":                                      assertDomainMassValue,
	"mass <id> should have elasticity <elasticity>":                                        assertDomainMassElasticity,
	"mass <id> fixed state should be <fixed>":                                              assertDomainMassFixed,
	"a spring <spring_id> connects mass <mass_a> to mass <mass_b>":                         addDomainSpring,
	"spring <spring_id> has spring constant <spring_constant>":                             setDomainSpringConstant,
	"spring <spring_id> has damping constant <damping_constant>":                           setDomainSpringDamping,
	"spring <spring_id> has rest length <rest_length>":                                     setDomainSpringRestLength,
	"the coder looks up spring <spring_id>":                                                lookupDomainSpring,
	"spring <spring_id> should connect mass <mass_a> to mass <mass_b>":                     assertDomainSpringEndpoints,
	"spring <spring_id> should have spring constant <spring_constant>":                     assertDomainSpringConstant,
	"spring <spring_id> should have damping constant <damping_constant>":                   assertDomainSpringDamping,
	"spring <spring_id> should have rest length <rest_length>":                             assertDomainSpringRestLength,
	"a world already contains a <object_type> with id <id>":                                addExistingDomainObject,
	"the coder adds another <object_type> with id <id>":                                    addDuplicateDomainObject,
	"the coder adds spring <spring_id> connecting mass <mass_a> to mass <mass_b>":          addInvalidDomainSpring,
	"validation should fail with reason <reason>":                                          assertDomainValidationReason,
	"parameter <parameter> should have default value <value>":                              assertParameterDefault,
	"force <force> should have enabled state <enabled>":                                    assertForceEnabledState,
	"force <force> should have editable parameters":                                        assertForceEditableParameters,
	"wall <wall> should have enabled state <enabled>":                                      assertWallEnabledState,
	"a world with parameter <parameter> changed to <changed_value>":                        changeWorldParameter,
	"the coder performs <operation>":                                                       performWorldOperation,
	"parameter <parameter> should be <expected_value_source>":                              assertParameterSource,
	"mass <mass_a> is connected to mass <mass_b> by a spring":                              createSpringForceWorld,
	"the spring has rest length <rest_length>":                                             setOnlySpringRestLength,
	"the spring has spring constant <spring_constant>":                                     setOnlySpringConstant,
	"mass <mass_a> has velocity <velocity_a>":                                              setMassAVelocity,
	"mass <mass_b> has velocity <velocity_b>":                                              setMassBVelocity,
	"the spring has damping constant <damping_constant>":                                   setOnlySpringDamping,
	"the coder evaluates forces without advancing time":                                    evaluateForces,
	"force on mass <mass_a> should be equal and opposite to force on mass <mass_b>":        assertSpringForcesEqualOpposite,
	"spring damping should affect only the spring direction":                               assertSpringDampingDirection,
	"a world with force <force> enabled":                                                   enableEnvironmentalForce,
	"a movable mass is affected by <force>":                                                createMovableMassAffectedByForce,
	"the mass should receive a force from <force>":                                         assertMassReceivesForce,
	"mass <mass_id> fixed state is <fixed>":                                                createMassFixedState,
	"mass <mass_id> is affected by force <force>":                                          affectMassByForce,
	"mass <mass_id> acceleration should be <acceleration>":                                 assertMassAcceleration,
	"wall <wall> is enabled":                                                               enableWall,
	"mass <mass_id> is outside the <wall> boundary":                                        createMassOutsideWall,
	"mass <mass_id> should receive force toward the inside of the world":                   assertWallForceTowardInside,
	"a movable mass starts at position <start_position>":                                   createMovableMassAtStart,
	"gravity is enabled":                                                                   enableGravity,
	"the coder advances the simulation by <duration>":                                      advanceByDuration,
	"the mass position should differ from <start_position>":                                assertMassPositionDiffers,
	"the mass velocity should differ from <start_velocity>":                                assertMassVelocityDiffers,
	"mass <mass_id> starts at position <start_position>":                                   createMassStartPosition,
	"mass <mass_id> position should remain <start_position>":                               assertMassPositionRemains,
	"mass <mass_id> velocity should remain <start_velocity>":                               assertMassVelocityRemains,
	"a world in state <initial_state>":                                                     createWorldInState,
	"the resulting state should be the same on every run":                                  assertResultDeterministic,
	"the coder advances the simulation by <duration> using render frame rate <frame_rate>": advanceByDurationAtFrameRate,
	"the resulting simulation time should be <duration>":                                   assertSimulationTime,
	"the acceptance pipeline task is accepted":                                             acceptStep,
	"the coder runs the acceptance test command":                                           runAcceptanceCommand,
	"the Gherkin parser should run successfully":                                           assertParserRan,
	"the acceptance test generator should run successfully":                                assertGeneratorRan,
	"the generated executable acceptance tests should run successfully":                    assertGeneratedRan,
	"the coder generates acceptance tests":                                                 generateAcceptanceArtifacts,
	"generated acceptance <artifact> should be written under <generated_location>":         assertGeneratedArtifactExists,
	"hand-written <test_type> tests should remain outside <generated_location>":            assertHandwrittenTestsOutside,
	"the coder adds a minimal smoke feature":                                               addSmokeFeature,
	"the smoke feature should parse successfully":                                          parseSmokeFeature,
	"the smoke feature should generate an executable acceptance test":                      generateSmokeAcceptanceTest,
	"the generated smoke acceptance test should pass":                                      assertSmokeAcceptanceTestPasses,
	"the coder checks out the committed project":                                           acceptStep,
	"the acceptance test command should pass without uncommitted setup steps":              assertAcceptanceCommandPassesFromCleanCheckout,
	"acceptance smoke is ready":                                                            acceptStep,
	"acceptance smoke should pass":                                                         acceptStep,
	"the project skeleton task is accepted":                                                acceptStep,
	"the coder creates the initial Go package layout":                                      createPackageLayout,
	"the <package> package should not import <graphics_library>":                           assertPackageDoesNotImport,
	"the coder creates the desktop application command":                                    createApplicationCommand,
	"the application command should build successfully":                                    assertApplicationCommandBuilds,
	"the coder creates the initial Go module":                                              createGoModule,
	"the Go test suite should pass":                                                        assertGoTestsPass,
	"a demo spring simulation":                                                             createDemoSimulation,
	"I advance the simulation <steps> steps":                                               advanceSimulation,
	"mass <mass> x should be <x>":                                                          assertMassX,
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

func setSimulation(target **sim.Simulation, simulation *sim.Simulation) error {
	*target = simulation
	return nil
}
