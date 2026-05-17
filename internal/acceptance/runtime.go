package acceptance

import (
	"fmt"

	"springs/internal/edit"
	"springs/internal/gherkin"
	"springs/internal/sim"
)

type world struct {
	simulation           *sim.Simulation
	layoutCreated        bool
	commandCreated       bool
	moduleCreated        bool
	parserRan            bool
	generatorRan         bool
	generatedRan         bool
	generated            bool
	smokeAdded           bool
	smokeParsed          bool
	smokeGenerated       bool
	domainWorld          *sim.Simulation
	lookedMass           sim.Mass
	lookedSpring         sim.Spring
	validationErr        error
	forceEvaluation      sim.ForceEvaluation
	resultingWorld       *sim.Simulation
	xspInput             string
	xspWorld             *sim.Simulation
	xspLoadErr           error
	xspSavedFirst        string
	xspSavedSecond       string
	xspResolvedFilename  string
	xspSpringDir         string
	appGame              appGame
	appErr               error
	appBeforeTime        float64
	appWindowSize        string
	editorScreen         editorScreen
	startupSecondScreen  editorScreen
	visibleControlsFrame drawFrameReport
	renderResult         renderResult
	mouseEditor          *edit.Editor
	createdMassID        int
	createdSpringID      int
	springStartMassID    int
	springCreated        bool
	springBehavior       string
	duplicated           edit.DuplicatedObjects
	originalMassIDs      []int
	originalSpringIDs    []int
	appCommand           string
	clickShortcut        string
	recordedAppState     string
	documentation        string
	cleanCheckout        bool
	documentedCommand    string
	documentedCommandErr error
	handoffVerification  map[string]string
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
	"the force evaluation task is accepted":                                                acceptStep,
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
	"mass <mass_id> fixed state should be <fixed>":                                         assertXSPMassFixedState,
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
	"the system parameters task is accepted":                                               acceptStep,
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
	"the simulation step task is accepted":                                                 acceptStep,
	"a movable mass starts at position <start_position>":                                   createMovableMassAtStart,
	"gravity is enabled":                                                                   enableGravity,
	"the coder advances the simulation by <duration>":                                      advanceByDuration,
	"the mass position should differ from <start_position>":                                assertMassPositionDiffers,
	"the mass velocity should differ from <start_velocity>":                                assertMassVelocityDiffers,
	"mass <mass_id> starts at position <start_position>":                                   createMassStartPosition,
	"mass <mass_id> starts at <start_position>":                                            createMassStartPosition,
	"mass <mass_id> position should remain <start_position>":                               assertMassPositionRemains,
	"mass <mass_id> velocity should remain <start_velocity>":                               assertMassVelocityRemains,
	"a world in state <initial_state>":                                                     createWorldInState,
	"the resulting state should be the same on every run":                                  assertResultDeterministic,
	"the coder advances the simulation by <duration> using render frame rate <frame_rate>": advanceByDurationAtFrameRate,
	"the resulting simulation time should be <duration>":                                   assertSimulationTime,
	"the adaptive RK4 numerics task is accepted":                                           acceptStep,
	"adaptive timestep is <adaptive>":                                                      setAdaptiveTimestep,
	"time step is <time_step>":                                                             setTimeStep,
	"RK4 integration should advance deterministically by <duration>":                       assertRK4DeterministicAdvance,
	"precision is <precision>":                                                             setPrecision,
	"the coder advances an unstable simulation by <duration>":                              advanceUnstableSimulation,
	"adaptive RK4 should choose <step_behavior>":                                           assertAdaptiveStepBehavior,
	"simulation time should advance by <duration>":                                         assertSimulationTimeAdvanced,
	"render frame rate is <frame_rate>":                                                    setRenderFrameRate,
	"simulation state should not depend on render frame rate <frame_rate>":                 assertStateIndependentOfFrameRate,
	"the XSP load and save task is accepted":                                               acceptStep,
	"XSP input starts with <marker>":                                                       createXSPInputWithMarker,
	"the coder loads the XSP input":                                                        loadXSPInput,
	"loading should <result>":                                                              assertXSPLoadResult,
	"XSP input contains command <command>":                                                 createXSPInputWithCommand,
	"the loaded world should include <loaded_state>":                                       assertXSPLoadedState,
	"a world loaded from <input_file>":                                                     createWorldLoadedFromFile,
	"the coder saves the world twice":                                                      saveXSPWorldTwice,
	"both saved outputs should be identical":                                               assertXSPSavesIdentical,
	"each saved output should end with a newline":                                          assertXSPSaveEndsWithNewline,
	"XSP input contains mass <mass_id> with file mass value <file_mass_value>":             createXSPInputWithFileMass,
	"the coder loads and saves the XSP input":                                              loadAndSaveXSPInput,
	"saved XSP output should include command <command>":                                    assertSavedXSPIncludesCommand,
	"saved mass <mass_id> should use file mass sign <file_mass_sign>":                      assertSavedMassSign,
	"XSP input has problem <problem>":                                                      createMalformedXSPInput,
	"loading should fail with reason <reason>":                                             assertXSPLoadErrorReason,
	"the XSP complete file format task is accepted":                                        acceptStep,
	"filename input is <filename>":                                                         createFilenameInput,
	"environment variable SPRINGDIR is <springdir>":                                        setSpringDirEnvironment,
	"the coder resolves an XSP filename":                                                   resolveXSPFilename,
	"resolved filename should be <resolved_filename>":                                      assertResolvedXSPFilename,
	"current parameters are <current_parameters>":                                          createCurrentParameters,
	"the coder inserts XSP file <input_file>":                                              insertXSPFile,
	"inserted masses and springs should be added":                                          assertInsertedObjectsAdded,
	"parameters should remain <current_parameters>":                                        assertParametersRemain,
	"the force center and force parameters task is accepted":                               acceptStep,
	"the coder selects force <force>":                                                      selectForce,
	"force <force> should expose parameter <parameter_one>":                                assertForceExposesParameter,
	"force <force> should expose parameter <parameter_two>":                                assertForceExposesParameter,
	"gravity direction is <direction_degrees>":                                             setGravityDirection,
	"the coder evaluates gravity":                                                          evaluateGravity,
	"gravity should point <expected_direction>":                                            assertGravityDirection,
	"selected masses are <selected_masses>":                                                createSelectedMasses,
	"the coder sets the force center":                                                      setForceCenter,
	"force center should be <expected_center>":                                             assertForceCenter,
	"force center is mass <center_mass>":                                                   createForceCenterMass,
	"force <force> is enabled":                                                             enableNamedForce,
	"the coder evaluates center forces":                                                    evaluateCenterForces,
	"mass <center_mass> should be visually marked as the center":                           assertCenterMassVisuallyMarked,
	"mass <center_mass> should not receive reciprocal force response from <force>":         assertNoReciprocalCenterForce,
	"the coder enables force <force>":                                                      enableForceForControls,
	"parameter controls for force <force> should be active":                                assertForceControlsActive,
	"the Ebitengine window task is accepted":                                               acceptStep,
	"the coder starts the desktop application":                                             startDesktopApplication,
	"the application window should open successfully":                                      assertApplicationWindowOpened,
	"the world should be empty":                                                            assertApplicationWorldEmpty,
	"the coder resizes the application window to <window_size>":                            resizeApplicationWindow,
	"the application should continue running":                                              assertApplicationContinuesRunning,
	"the application simulation pause state is <paused>":                                   setApplicationPauseState,
	"the coder updates the application loop":                                               updateApplicationLoop,
	"simulation stepping should be <stepping>":                                             assertApplicationStepping,
	"input handling should remain active":                                                  assertApplicationInputActive,
	"rendering should remain active":                                                       assertApplicationRenderingActive,
	"the coder closes the application window":                                              closeApplicationWindow,
	"the application should exit without error":                                            assertApplicationExitClean,
	"the screen and controls task is accepted":                                             acceptStep,
	"the first screen should show the simulation editor":                                   assertFirstScreenEditor,
	"the first screen should not show a landing page":                                      assertNoLandingPage,
	"the coder lays out the editor screen":                                                 layoutEditorScreen,
	"screen region <region> should be visible":                                             assertScreenRegionVisible,
	"screen region <region> should have purpose <purpose>":                                 assertScreenRegionPurpose,
	"the coder shows the left toolbar":                                                     layoutEditorScreen,
	"editing mode <mode> should have a visible control":                                    assertModeVisible,
	"the coder shows the top command bar":                                                  layoutEditorScreen,
	"command <command> should have a visible control":                                      assertCommandVisible,
	"application state <state> is active":                                                  setApplicationState,
	"the coder renders the editor controls":                                                layoutEditorScreen,
	"visible indicator <indicator> should reflect <state>":                                 assertVisibleIndicator,
	"command <command> has visible control <control>":                                      setVisibleCommandControl,
	"the coder presses keyboard shortcut <shortcut>":                                       pressKeyboardShortcut,
	"command <command> should run":                                                         assertCommandRan,
	"simulation state is <simulation_state>":                                               setSimulationState,
	"the coder renders the editor screen":                                                  layoutEditorScreen,
	"the canvas should remain visible":                                                     assertCanvasVisible,
	"the visible controls should remain usable":                                            assertControlsUsable,
	"the mouse editing task is accepted":                                                   acceptStep,
	"the editor mode is <mode>":                                                            setMouseEditorMode,
	"the editor mode is add mass":                                                          setMouseEditorModeAddMass,
	"the current mass defaults are configured":                                             configureCurrentMassDefaults,
	"the coder clicks at <pointer_position>":                                               clickMouseEditor,
	"a mass should be created at <expected_position>":                                      assertCreatedMassPosition,
	"the mass should use the current mass defaults":                                        assertCreatedMassDefaults,
	"grid snap is <grid_snap>":                                                             setMouseGridSnap,
	"the grid snap size is <snap_size>":                                                    setMouseGridSnapSize,
	"mass <mass_a> exists":                                                                 addMouseMassA,
	"mass <mass_b> exists":                                                                 addMouseMassB,
	"the coder creates a spring from mass <mass_a> to mass <mass_b>":                       createMouseSpring,
	"a spring should connect mass <mass_a> to mass <mass_b>":                               assertMouseSpringEndpoints,
	"the spring should use the current spring defaults":                                    assertMouseSpringDefaults,
	"the coder drags mass <mass_id> to <target_position>":                                  dragMouseMass,
	"mass <mass_id> position should be <expected_position>":                                assertMouseMassPosition,
	"mass <mass_id> id should remain <mass_id>":                                            assertMouseMassID,
	"the selection and editing task is accepted":                                           acceptStep,
	"the world contains a <object_type> with id <id>":                                      createSelectableObject,
	"the coder selects <object_type> <id>":                                                 selectObject,
	"<object_type> <id> should be selected":                                                assertObjectSelected,
	"the world contains masses and springs":                                                createSelectionWorld,
	"the coder selects all objects":                                                        selectAllObjects,
	"every mass should be selected":                                                        assertEveryMassSelected,
	"every spring should be selected":                                                      assertEverySpringSelected,
	"<object_type> <id> is selected":                                                       selectObject,
	"the coder deletes selected objects":                                                   deleteSelectedObjects,
	"<object_type> <id> should not exist":                                                  assertObjectDeleted,
	"mass 1 is connected to mass 2 by spring 3":                                            createSelectionConnectedMasses,
	"mass 1 is selected":                                                                   selectMassOne,
	"mass 1 should not exist":                                                              assertMassOneDeleted,
	"spring 3 should not exist":                                                            assertSpringThreeDeleted,
	"mass 2 should still exist":                                                            assertMassTwoExists,
	"selected <object_set> exists":                                                         createSelectedObjectSet,
	"the coder duplicates selected objects":                                                duplicateSelectedObjects,
	"duplicated objects should have unique ids":                                            assertDuplicatedUniqueIDs,
	"duplicated objects should be independent from the original objects":                   assertDuplicatedIndependent,
	"the controls and hotkeys task is accepted":                                            acceptStep,
	"the application is running":                                                           createRunningApplication,
	"the coder presses shortcut <shortcut>":                                                pressShortcut,
	"the world is in state <initial_state>":                                                createControlWorldState,
	"the coder runs file command <command>":                                                runFileCommand,
	"the world state should be <expected_state>":                                           assertControlWorldState,
	"system parameters should be <parameter_result>":                                       assertControlParameterResult,
	"the world contains objects":                                                           createWorldObjects,
	"system parameters have custom values":                                                 setCustomSystemParameters,
	"the coder runs the reset command":                                                     runResetCommand,
	"the world should contain zero masses":                                                 assertControlMassCountZero,
	"the world should contain zero springs":                                                assertControlSpringCountZero,
	"system parameters should equal defaults":                                              assertControlParametersDefault,
	"parameter <parameter> has value <old_value>":                                          setControlParameterValue,
	"the coder changes parameter <parameter> to <new_value>":                               changeControlParameterValue,
	"parameter <parameter> should have value <new_value>":                                  assertControlParameterValue,
	"the render world task is accepted":                                                    acceptStep,
	"the application has <world_state>":                                                    createApplicationWorldState,
	"the coder renders the world":                                                          renderApplicationWorld,
	"rendering should complete successfully":                                               assertRenderingComplete,
	"the world contains <object>":                                                          createRenderableObject,
	"the world contains a spring":                                                          createRenderableSpring,
	"<object> should have a visible representation":                                        assertVisibleRepresentation,
	"show springs is <show_springs>":                                                       setShowSprings,
	"spring lines should be <spring_visibility>":                                           assertSpringLineVisibility,
	"masses should remain visible":                                                         assertMassesVisible,
	"the world contains a fixed mass and a movable mass":                                   createFixedAndMovableMasses,
	"the fixed mass should be visually distinguishable from the movable mass":              assertFixedMassDistinguishable,
	"the demo files task is accepted":                                                      acceptStep,
	"the coder adds demo file <demo_file>":                                                 assertDemoFileAdded,
	"demo file <demo_file> should be valid XSP":                                            assertDemoFileValid,
	"demo file <demo_file> should be human readable":                                       assertDemoFileHumanReadable,
	"demo file <demo_file> exists":                                                         assertDemoFileExists,
	"the coder loads demo file <demo_file>":                                                loadDemoFile,
	"the loaded world should include <required_feature>":                                   assertDemoLoadedFeature,
	"a demo spring simulation":                                                             createDemoSimulation,
	"I advance the simulation <steps> steps":                                               advanceSimulation,
	"mass <mass> x should be <x>":                                                          assertMassX,
	"the packaging and docs task is accepted":                                              acceptStep,
	"a developer reads the project documentation":                                          readProjectDocumentation,
	"command <command> should be documented":                                               assertDocumentedCommand,
	"a clean checkout":                                                                     markCleanCheckout,
	"a developer runs documented command <command>":                                        runDocumentedCommand,
	"command <command> should pass":                                                        assertDocumentedCommandPassed,
	"the documentation should explain <topic>":                                             assertDocumentationExplains,
	"the coder completes the packaging and docs task":                                      completePackagingDocsTask,
	"the handoff should include the local verification commands that were run":             assertHandoffIncludesVerificationCommands,
	"the handoff should include the result of each verification command":                   assertHandoffIncludesVerificationResults,
	"the edit mode details task is accepted":                                               acceptStep,
	"edit mode is active":                                                                  activateEditMode,
	"object <object_id> is near the pointer":                                               addObjectNearPointer,
	"selection initially contains <initial_selection>":                                     setInitialEditSelection,
	"the coder <click_action> object <object_id>":                                          clickEditObject,
	"selection should contain <expected_selection>":                                        assertEditSelection,
	"objects <inside_objects> are inside the selection box":                                addObjectsInsideSelectionBox,
	"objects <outside_objects> are outside the selection box":                              addObjectsOutsideSelectionBox,
	"the coder drags an empty-space selection box with <modifier>":                         dragSelectionBox,
	"selected object <object_id> starts at <start_position>":                               addSelectedObjectAtStart,
	"the coder middle-drags selected objects by <drag_delta>":                              middleDragSelectedObjects,
	"object <object_id> position should be <expected_position>":                            assertEditObjectPosition,
	"selected mass <mass_id> fixed state is <fixed>":                                       addSelectedMassWithFixedState,
	"the coder right-drags selected masses with release velocity <release_velocity>":       rightDragSelectedMasses,
	"mass <mass_id> velocity should be <expected_velocity>":                                assertEditMassVelocity,
	"the spring mode mouse semantics task is accepted":                                     acceptStep,
	"spring mode is active":                                                                activateSpringMode,
	"pointer press is near mass <start_mass>":                                              pressNearSpringMass,
	"the coder releases the pointer <release_target>":                                      releaseSpringPointer,
	"spring creation should <result>":                                                      assertSpringCreationResult,
	"the coder drags with mouse button <button>":                                           dragSpringWithButton,
	"the pending spring behavior should be <behavior>":                                     assertPendingSpringBehavior,
	"current Kspring is <kspring>":                                                         setCurrentKspring,
	"current Kdamp is <kdamp>":                                                             setCurrentKdamp,
	"the coder creates a spring with length <creation_length>":                             createSpringWithLength,
	"the spring Kspring should be <kspring>":                                               assertCreatedSpringKspring,
	"the spring Kdamp should be <kdamp>":                                                   assertCreatedSpringKdamp,
	"the spring rest length should be <creation_length>":                                   assertCreatedSpringRestLength,
	"the state save restore task is accepted":                                              acceptStep,
	"the world is in state <saved_state>":                                                  createMemoryWorldState,
	"the coder saves state":                                                                saveApplicationState,
	"the world changes to state <changed_state>":                                           changeApplicationState,
	"the coder restores state <restore_count> times":                                       restoreApplicationStateTimes,
	"the world should be in state <saved_state>":                                           assertApplicationStateWorld,
	"no state has been saved":                                                              createNoSavedApplicationState,
	"the world has changed from the initial state":                                         changeFromInitialApplicationState,
	"the coder restores state":                                                             restoreApplicationStateOnce,
	"the world should be in the initial state":                                             assertInitialApplicationState,
	"the world is in state <memory_state>":                                                 createMemoryWorldState,
	"the coder performs file operation <file_operation>":                                   runStateFileOperation,
	"the world should be in state <memory_state>":                                          assertApplicationStateWorld,
	"the selected object parameter editing task is accepted":                               acceptStep,
	"selected mass <mass_id> exists":                                                       createSelectedParameterMass,
	"the coder changes mass control <control> to <value>":                                  changeMassControl,
	"mass <mass_id> should have <control> value <value>":                                   assertMassControlValue,
	"selected spring <spring_id> exists":                                                   createSelectedParameterSpring,
	"the coder changes spring control <control> to <value>":                                changeSpringControl,
	"spring <spring_id> should have <control> value <value>":                               assertSpringControlValue,
	"selected spring <spring_id> has current length <current_length>":                      createSelectedSpringWithCurrentLength,
	"the coder sets rest length":                                                           setSelectedRestLength,
	"spring <spring_id> rest length should be <current_length>":                            assertSelectedSpringRestLength,
	"no selected object is compatible with control <control>":                              createNoCompatibleSelection,
	"the coder changes control <control> to <value>":                                       changeGenericControl,
	"future <object_type> objects should use <control> value <value>":                      assertFutureObjectUsesControlValue,
	"the wall collision and stickiness task is accepted":                                   acceptStep,
	"mass <mass_id> has elasticity <elasticity>":                                           setCollisionMassElasticity,
	"mass <mass_id> moves from inside the screen toward wall <wall>":                       moveMassFromInsideTowardWall,
	"the coder advances through the wall collision":                                        advanceThroughWallCollision,
	"mass <mass_id> wall-normal velocity should be reversed":                               assertWallNormalVelocityReversed,
	"mass <mass_id> wall-normal velocity magnitude should be scaled by <elasticity>":       assertWallNormalVelocityScaled,
	"mass <mass_id> moves from off-screen toward the screen through wall <wall>":           moveMassFromOutsideThroughWall,
	"the coder advances through the wall boundary":                                         advanceThroughWallBoundary,
	"mass <mass_id> should pass through wall <wall>":                                       assertMassPassedThroughWall,
	"stickiness is <stickiness>":                                                           setCollisionStickiness,
	"mass <mass_id> collides with wall <wall>":                                             collideMassWithWall,
	"the wall collision removes all wall-normal velocity":                                  removeWallNormalVelocity,
	"mass <mass_id> should stick to wall <wall>":                                           assertMassStuckToWall,
	"force <release_force> pulls mass <mass_id> away from wall <wall>":                     pullMassAwayFromWall,
	"mass <mass_id> should be <release_result>":                                            assertMassReleaseResult,
	"wall <wall> is disabled":                                                              disableWall,
	"mass <mass_id> moves toward wall <wall>":                                              moveMassTowardWall,
	"mass <mass_id> should not bounce from wall <wall>":                                    assertMassDidNotBounce,
	"the nonblank startup editor task is accepted":                                         acceptStep,
	"the nonblank startup editor 23.1 task is accepted":                                    acceptStep,
	"the first screen should show visible editor chrome":                                   assertStartupEditorChrome,
	"the first screen should show visible world content":                                   assertStartupWorldContent,
	"debug text should not be the only visible content":                                    assertDebugTextNotOnlyContent,
	"the startup world should be loaded from <default_demo>":                               assertStartupWorldLoadedFromDemo,
	"startup screen region <region> should be visible":                                     assertStartupRegionVisible,
	"the startup world should contain <object_count> <object_type>":                        assertStartupObjectCount,
	"the coder starts the desktop application twice":                                       startDesktopApplicationTwice,
	"both startup worlds should be equivalent":                                             assertStartupWorldsEquivalent,
	"both startup screens should show the same editor chrome":                              assertStartupScreensEquivalent,
	"the startup world should match demo file <default_demo>":                              assertStartupWorldMatchesDemo,
	"the render visible controls task is accepted":                                         acceptStep,
	"the coder draws the application frame":                                                drawApplicationFrame,
	"screen region <region> should contain non-background pixels":                          assertFrameRegionHasPixels,
	"screen region <region> should not contain only debug text":                            assertFrameRegionNotOnlyDebugText,
	"visible control <control> should have readable label <label>":                         assertVisibleControlReadableLabel,
	"inspector section <section> should be visible":                                        assertInspectorSectionVisible,
	"status field <field> should be visible":                                               assertStatusFieldVisible,
	"status field <field> should show <state>":                                             assertStatusFieldShows,
	"the canvas should contain visible world content":                                      assertCanvasWorldContentVisible,
	"editor chrome should not cover all world content":                                     assertChromeDoesNotCoverAllWorldContent,
	"the coder draws the application frame at the default window size":                     drawApplicationFrameDefaultSize,
	"visible control labels should fit inside their regions":                               assertVisibleControlLabelsFit,
	"the clickable visible controls task is accepted":                                      acceptStep,
	"the editor mode is <old_mode>":                                                        setClickableEditorMode,
	"the coder clicks inside rendered bounds of visible control <control>":                 clickInsideRenderedVisibleControlBounds,
	"the editor mode should be <new_mode>":                                                 assertClickableEditorMode,
	"visible control <control> should show active state":                                   assertVisibleControlActive,
	"keyboard path entry should open for <command>":                                        assertKeyboardPathEntryOpen,
	"visible control <control> maps to shortcut <shortcut>":                                recordVisibleControlShortcut,
	"the result should match pressing shortcut <shortcut>":                                 assertClickMatchesShortcut,
	"the application state is recorded":                                                    recordClickableApplicationState,
	"the coder clicks outside all visible controls":                                        clickOutsideVisibleControls,
	"the application state should remain unchanged":                                        assertClickableApplicationStateUnchanged,
	"simulation state is <old_state>":                                                      setClickableSimulationState,
	"simulation state should be <new_state>":                                               assertClickableSimulationState,
	"the original demo corpus task is accepted":                                            acceptStep,
	"the coder imports the original XSpringies demo corpus":                                importOriginalDemoCorpus,
	"imported demo file <demo_file> should exist under <demo_directory>":                   assertImportedDemoExists,
	"imported demo file <demo_file> should preserve its original filename":                 assertImportedDemoPreservesFilename,
	"imported original demo file <demo_file> exists":                                       assertImportedOriginalDemoExists,
	"the coder loads imported original demo file <demo_file>":                              loadImportedOriginalDemo,
	"loading should pass":                                                                  assertLoadingPassed,
	"starter demo file <starter_demo> exists":                                              assertStarterDemoExists,
	"starter demo file <starter_demo> should remain under <starter_directory>":             assertStarterDemoRemainsUnder,
	"original demos should remain under <original_directory>":                              assertOriginalDemosRemainUnder,
	"provenance field <field> should be documented":                                        assertProvenanceFieldDocumented,
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
