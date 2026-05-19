package acceptance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"springs/internal/app"
	"springs/internal/edit"
	"springs/internal/gherkin"
	"springs/internal/sim"
)

func TestRunFeatureExecutesProjectSkeletonScenarios(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Project skeleton",
		Background: []gherkin.Step{
			{Keyword: "Given", Text: "the project skeleton task is accepted"},
		},
		Scenarios: []gherkin.Scenario{{
			Name: "domain independence",
			Steps: []gherkin.Step{
				{Keyword: "When", Text: "the coder creates the initial Go package layout"},
				{Keyword: "Then", Text: "the <package> package should not import <graphics_library>", Parameters: []string{"package", "graphics_library"}},
			},
			Examples: []map[string]string{
				{"package": "simulation", "graphics_library": "Ebitengine"},
				{"package": "file format", "graphics_library": "Ebitengine"},
			},
		}},
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureExecutesPipelineSmokeScenario(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Pipeline smoke",
		Scenarios: []gherkin.Scenario{{
			Name: "smoke",
			Steps: []gherkin.Step{
				{Keyword: "Given", Text: "acceptance smoke is ready"},
				{Keyword: "Then", Text: "acceptance smoke should pass"},
			},
		}},
	}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureFailsUnsupportedSteps(t *testing.T) {
	feature := gherkin.Feature{
		Name: "Unsupported",
		Scenarios: []gherkin.Scenario{{
			Name:  "bad",
			Steps: []gherkin.Step{{Keyword: "Then", Text: "something unknown happens"}},
		}},
	}

	err := RunFeature(feature)
	if err == nil || !strings.Contains(err.Error(), "unsupported step") {
		t.Fatalf("expected unsupported step error, got %v", err)
	}
}

func TestRunFeatureReportsOneBasedExampleNumber(t *testing.T) {
	feature := gherkin.Feature{
		Scenarios: []gherkin.Scenario{{
			Name:  "bad examples",
			Steps: []gherkin.Step{{Keyword: "Then", Text: "something unknown happens"}},
			Examples: []map[string]string{
				{"case": "first"},
				{"case": "second"},
			},
		}},
	}

	err := RunFeature(feature)
	if err == nil || !strings.Contains(err.Error(), "bad examples/example_1:") {
		t.Fatalf("expected example_1 error, got %v", err)
	}
}

func TestRunFeatureExecutesSimulationSteps(t *testing.T) {
	feature := gherkin.Feature{Scenarios: []gherkin.Scenario{{
		Name: "simulation",
		Steps: []gherkin.Step{
			{Keyword: "Given", Text: "a demo spring simulation"},
			{Keyword: "When", Text: "I advance the simulation <steps> steps", Parameters: []string{"steps"}},
			{Keyword: "Then", Text: "mass <mass> x should be <x>", Parameters: []string{"mass", "x"}},
		},
		Examples: []map[string]string{{"steps": "0", "mass": "0", "x": "160"}},
	}}}

	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestRunFeatureExecutesDomainModelFeature(t *testing.T) {
	runFeatureFile(t, "features/003_domain_model.feature")
}

func TestStepPrerequisitesReturnHelpfulErrors(t *testing.T) {
	cases := []gherkin.Step{
		{Text: "the <package> package should not import <graphics_library>"},
		{Text: "the application command should build successfully"},
		{Text: "the Go test suite should pass"},
		{Text: "I advance the simulation <steps> steps"},
		{Text: "mass <mass> x should be <x>"},
		{Text: "the Gherkin parser should run successfully"},
		{Text: "the acceptance test generator should run successfully"},
		{Text: "the generated executable acceptance tests should run successfully"},
		{Text: "generated acceptance <artifact> should be written under <generated_location>"},
		{Text: "the smoke feature should parse successfully"},
		{Text: "the smoke feature should generate an executable acceptance test"},
		{Text: "the generated smoke acceptance test should pass"},
	}
	example := map[string]string{
		"package":            "simulation",
		"graphics_library":   "Ebitengine",
		"steps":              "1",
		"mass":               "0",
		"x":                  "160",
		"artifact":           "test source",
		"generated_location": "acceptance/generated",
	}

	for _, step := range cases {
		if err := runStep(&world{}, step, example); err == nil {
			t.Fatalf("expected prerequisite error for %q", step.Text)
		}
	}
}

func TestExampleValueParsingReportsMissingAndInvalidValues(t *testing.T) {
	if _, err := stringValue(nil, "missing"); err == nil {
		t.Fatal("expected missing string error")
	}
	if _, err := intValue(map[string]string{"value": "NaN"}, "value"); err == nil {
		t.Fatal("expected invalid integer error")
	}
	if _, err := floatValue(map[string]string{"value": "NaN?"}, "value"); err == nil {
		t.Fatal("expected invalid float error")
	}
	if _, err := boolValue(map[string]string{"value": "maybe"}, "value"); err == nil {
		t.Fatal("expected invalid bool error")
	}
}

func TestDomainModelHelpersReportFailures(t *testing.T) {
	w := &world{}
	if _, err := domainWorld(w); err == nil {
		t.Fatal("expected missing domain world error")
	}
	if err := lookupDomainMass(&world{domainWorld: sim.NewWorld()}, map[string]string{"id": "7"}); err == nil {
		t.Fatal("expected missing mass error")
	}
	if err := lookupDomainSpring(&world{domainWorld: sim.NewWorld()}, map[string]string{"spring_id": "7"}); err == nil {
		t.Fatal("expected missing spring error")
	}
	if err := assertValidationReason(nil, "duplicate id"); err == nil {
		t.Fatal("expected validation success error")
	}
	if err := assertValidationReason(sim.ErrDuplicateID, "missing spring endpoint"); err == nil {
		t.Fatal("expected validation reason mismatch")
	}
	if objectType, id, err := objectTypeAndID(map[string]string{"id": "1"}); err == nil {
		t.Fatal("expected missing object type error")
	} else if objectType != "" || id != 0 {
		t.Fatalf("expected zero values on missing object type, got %q %d", objectType, id)
	}
	if objectType, id, err := objectTypeAndID(map[string]string{"object_type": "mass"}); err == nil {
		t.Fatal("expected missing object id error")
	} else if objectType != "" || id != 0 {
		t.Fatalf("expected zero values on missing object id, got %q %d", objectType, id)
	}
	for _, tt := range []struct {
		name   string
		fields map[string]string
	}{
		{name: "missing id", fields: map[string]string{"x": "1", "y": "2"}},
		{name: "missing x", fields: map[string]string{"id": "1", "y": "2"}},
		{name: "missing y", fields: map[string]string{"id": "1", "x": "2"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			id, x, y, err := massFields(tt.fields, "id", "x", "y")
			if err == nil {
				t.Fatal("expected missing mass fields error")
			}
			if id != 0 || x != 0 || y != 0 {
				t.Fatalf("expected zero values on missing mass fields, got %d %f %f", id, x, y)
			}
		})
	}
	if err := assertVec("point", sim.Vec2{X: 0.000001, Y: 0}, 0, 0); err != nil {
		t.Fatalf("expected exact x tolerance to pass: %v", err)
	}
	if err := assertVec("point", sim.Vec2{X: 0, Y: 0.000001}, 0, 0); err != nil {
		t.Fatalf("expected exact y tolerance to pass: %v", err)
	}
	if err := assertFloat("value", 0.000001, 0); err != nil {
		t.Fatalf("expected exact float tolerance to pass: %v", err)
	}
	if err := assertDomainCount(&world{domainWorld: sim.NewWorld()}, map[string]string{}, "masses", "mass_count", massCount); err == nil {
		t.Fatal("expected missing domain count error")
	}
	if err := assertVecExample("point", sim.Vec2{}, map[string]string{"x": "1"}, "x", "y"); err == nil {
		t.Fatal("expected missing vector example error")
	}
	if err := assertFloatExample("value", 0, map[string]string{}, "value"); err == nil {
		t.Fatal("expected missing float example error")
	}
	if err := assertDomainMassFixed(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing fixed value error")
	}
	if err := assertDomainSpringEndpoints(&world{}, map[string]string{"mass_b": "2"}); err == nil {
		t.Fatal("expected missing spring mass A error")
	}
	if err := assertDomainSpringEndpoints(&world{}, map[string]string{"mass_a": "1"}); err == nil {
		t.Fatal("expected missing spring mass B error")
	}
	if err := assertDomainSpringEndpoints(&world{lookedSpring: sim.Spring{MassA: 9, MassB: 2}}, map[string]string{"mass_a": "1", "mass_b": "2"}); err == nil {
		t.Fatal("expected spring mass A mismatch")
	}
	if err := assertDomainSpringEndpoints(&world{lookedSpring: sim.Spring{MassA: 1, MassB: 9}}, map[string]string{"mass_a": "1", "mass_b": "2"}); err == nil {
		t.Fatal("expected spring mass B mismatch")
	}
	if err := assertDomainValidationReason(&world{validationErr: sim.ErrDuplicateID}, map[string]string{}); err == nil {
		t.Fatal("expected missing validation reason error")
	} else if err.Error() != "missing example value reason" {
		t.Fatalf("expected missing validation reason error, got %v", err)
	}
}

func TestDomainModelHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := createDomainWorld(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainMassCount(w, map[string]string{"mass_count": "0"}); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainSpringCount(w, map[string]string{"spring_count": "0"}); err != nil {
		t.Fatal(err)
	}
	massExample := map[string]string{
		"id": "1", "x": "1.5", "y": "2.5",
		"vx": "3.5", "vy": "4.5",
		"mass_value": "5.5", "elasticity": "0.8", "fixed": "true",
	}
	if err := addDomainMass(w, massExample); err != nil {
		t.Fatalf("addDomainMass returned error: %v", err)
	}
	if mass, ok := w.domainWorld.MassByID(1); !ok {
		t.Fatal("expected domain mass to be added")
	} else if mass.Mass != 1 {
		t.Fatalf("expected default domain mass value 1, got %f", mass.Mass)
	}
	for _, fn := range []stepHandler{
		setDomainMassVelocity,
		setDomainMassValue,
		setDomainMassElasticity,
		setDomainMassFixed,
		lookupDomainMass,
		assertDomainMassPosition,
		assertDomainMassVelocity,
		assertDomainMassValue,
		assertDomainMassElasticity,
		assertDomainMassFixed,
	} {
		if err := fn(w, massExample); err != nil {
			t.Fatalf("mass handler returned error: %v", err)
		}
	}
	mass, ok := w.domainWorld.MassByID(1)
	if !ok {
		t.Fatal("expected domain mass to be added")
	}
	if mass.Mass != 5.5 {
		t.Fatalf("expected updated mass value 5.5, got %f", mass.Mass)
	}
}

func TestDomainSpringHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := addDomainMassA(w, map[string]string{"mass_a": "1", "x_a": "0", "y_a": "0"}); err != nil {
		t.Fatal(err)
	}
	if err := addDomainMassB(w, map[string]string{"mass_b": "2", "x_b": "10", "y_b": "0"}); err != nil {
		t.Fatal(err)
	}
	springExample := map[string]string{
		"spring_id": "7", "mass_a": "1", "mass_b": "2",
		"spring_constant": "12.5", "damping_constant": "0.7", "rest_length": "10",
	}
	for _, fn := range []stepHandler{
		addDomainSpring,
		setDomainSpringConstant,
		setDomainSpringDamping,
		setDomainSpringRestLength,
		lookupDomainSpring,
		assertDomainSpringEndpoints,
		assertDomainSpringConstant,
		assertDomainSpringDamping,
		assertDomainSpringRestLength,
	} {
		if err := fn(w, springExample); err != nil {
			t.Fatalf("spring handler returned error: %v", err)
		}
	}
}

func TestDomainValidationHandlers(t *testing.T) {
	duplicateMass := map[string]string{"object_type": "mass", "id": "1", "reason": "duplicate id"}
	w := &world{}
	if err := addExistingDomainObject(w, duplicateMass); err != nil {
		t.Fatal(err)
	}
	if mass, ok := w.domainWorld.MassByID(1); !ok {
		t.Fatal("expected existing duplicate mass")
	} else if mass.Mass != 1 {
		t.Fatalf("expected existing duplicate mass value 1, got %f", mass.Mass)
	}
	if err := addDuplicateDomainObject(w, duplicateMass); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainValidationReason(w, duplicateMass); err != nil {
		t.Fatal(err)
	}

	duplicateSpring := map[string]string{"object_type": "spring", "id": "5", "reason": "duplicate id"}
	w = &world{}
	if err := addExistingDomainObject(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}
	if mass, ok := w.domainWorld.MassByID(1); !ok {
		t.Fatal("expected first endpoint mass")
	} else if mass.Mass != 1 {
		t.Fatalf("expected first endpoint mass value 1, got %f", mass.Mass)
	}
	if mass, ok := w.domainWorld.MassByID(2); !ok {
		t.Fatal("expected second endpoint mass")
	} else if mass.Mass != 1 {
		t.Fatalf("expected second endpoint mass value 1, got %f", mass.Mass)
	}
	if err := addDuplicateDomainObject(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(w.validationErr, sim.ErrDuplicateID) {
		t.Fatalf("expected duplicate spring error, got %v", w.validationErr)
	}
	if err := assertDomainValidationReason(w, duplicateSpring); err != nil {
		t.Fatal(err)
	}

	w = &world{domainWorld: sim.NewWorld()}
	if err := w.domainWorld.AddMass(sim.Mass{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := w.domainWorld.AddMass(sim.Mass{ID: 2}); err != nil {
		t.Fatal(err)
	}
	if err := addDuplicateDomainObject(w, map[string]string{"object_type": "spring", "id": "6"}); err != nil {
		t.Fatal(err)
	}
	spring, ok := w.domainWorld.SpringByID(6)
	if !ok {
		t.Fatal("expected added spring")
	}
	if spring.MassA != 1 || spring.MassB != 2 {
		t.Fatalf("expected added spring endpoints 1,2 got %d,%d", spring.MassA, spring.MassB)
	}

	missingEndpoint := map[string]string{"existing_mass": "1", "x": "0", "y": "0", "spring_id": "2", "mass_a": "1", "mass_b": "9", "reason": "missing spring endpoint"}
	w = &world{}
	if err := addExistingDomainMass(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
	if err := addInvalidDomainSpring(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
	if err := assertDomainValidationReason(w, missingEndpoint); err != nil {
		t.Fatal(err)
	}
}

func TestRunFeatureExecutesSystemParameterFeature(t *testing.T) {
	runFeatureFile(t, "features/004_system_parameters.feature")
}

func TestRunFeatureExecutesForceEvaluationFeature(t *testing.T) {
	runFeatureFile(t, "features/005_force_evaluation.feature")
}

func TestRunFeatureExecutesSimulationStepFeature(t *testing.T) {
	runFeatureFile(t, "features/006_simulation_step.feature")
}

func TestRunFeatureExecutesXSPLoadSaveFeature(t *testing.T) {
	runFeatureFile(t, "features/007_xsp_load_save.feature")
}

func TestRunFeatureExecutesEbitengineWindowFeature(t *testing.T) {
	runFeatureFile(t, "features/008_ebitengine_window.feature")
}

func TestRunFeatureExecutesScreenControlsFeature(t *testing.T) {
	runFeatureFile(t, "features/008a_screen_and_controls.feature")
}

func TestRunFeatureExecutesRenderWorldFeature(t *testing.T) {
	runFeatureFile(t, "features/009_render_world.feature")
}

func TestRunFeatureExecutesMouseEditingFeature(t *testing.T) {
	runFeatureFile(t, "features/010_mouse_editing.feature")
}

func TestRunFeatureExecutesSelectionEditingFeature(t *testing.T) {
	runFeatureFile(t, "features/011_selection_and_editing.feature")
}

func TestRunFeatureExecutesControlsHotkeysFeature(t *testing.T) {
	runFeatureFile(t, "features/012_controls_and_hotkeys.feature")
}

func TestRunFeatureExecutesDemoFilesFeature(t *testing.T) {
	runFeatureFile(t, "features/013_demo_files.feature")
}

func TestRunFeatureExecutesEditModeDetailsFeature(t *testing.T) {
	runFeatureFile(t, "features/015_edit_mode_details.feature")
}

func TestRunFeatureExecutesSpringModeMouseSemanticsFeature(t *testing.T) {
	runFeatureFile(t, "features/016_spring_mode_mouse_semantics.feature")
}

func TestRunFeatureExecutesStateSaveRestoreFeature(t *testing.T) {
	runFeatureFile(t, "features/017_state_save_restore.feature")
}

func TestRunFeatureExecutesSelectedObjectParameterEditingFeature(t *testing.T) {
	runFeatureFile(t, "features/018_selected_object_parameter_editing.feature")
}

func TestRunFeatureExecutesWallCollisionStickinessFeature(t *testing.T) {
	runFeatureFile(t, "features/019_wall_collision_and_stickiness.feature")
}

func TestRunFeatureExecutesXSPCompleteFileFormatFeature(t *testing.T) {
	runFeatureFile(t, "features/020_xsp_complete_file_format.feature")
}

func TestRunFeatureExecutesForceCenterParametersFeature(t *testing.T) {
	runFeatureFile(t, "features/021_force_center_and_force_parameters.feature")
}

func TestRunFeatureExecutesAdaptiveRK4NumericsFeature(t *testing.T) {
	runFeatureFile(t, "features/022_adaptive_rk4_numerics.feature")
}

func TestRunFeatureExecutesNonblankStartupEditorFeature(t *testing.T) {
	runFeatureFile(t, "features/023_1_nonblank_startup_editor.feature")
}

func TestWallCollisionHelpersValidateInputs(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "reversed velocity missing mass id",
			err:  assertWallNormalVelocityReversed(&world{}, map[string]string{"wall": "left"}),
			want: "mass_id",
		},
		{
			name: "scaled velocity missing elasticity",
			err:  assertWallNormalVelocityScaled(wallCollisionWorldWithMass(1, "left", sim.Vec2{X: 5}), map[string]string{"mass_id": "1", "wall": "left"}),
			want: "elasticity",
		},
		{
			name: "passed through missing wall",
			err:  assertMassPassedThroughWall(wallCollisionWorldWithMass(1, "left", sim.Vec2{}), map[string]string{"mass_id": "1"}),
			want: "wall",
		},
		{
			name: "stuck assertion missing mass",
			err:  assertMassStuckToWall(&world{}, map[string]string{"mass_id": "1", "wall": "left"}),
			want: "mass 1 not found",
		},
		{
			name: "release assertion missing result",
			err:  assertMassReleaseResult(wallCollisionWorldWithMass(1, "left", sim.Vec2{}), map[string]string{"mass_id": "1", "wall": "left"}),
			want: "release_result",
		},
		{
			name: "disabled bounce assertion missing wall",
			err:  assertMassDidNotBounce(wallCollisionWorldWithMass(1, "left", sim.Vec2{}), map[string]string{"mass_id": "1"}),
			want: "wall",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == nil || !strings.Contains(test.err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", test.err, test.want)
			}
		})
	}
}

func TestWallCollisionHelperContracts(t *testing.T) {
	for wall, boundary := range map[string]sim.Vec2{
		"left":   {X: 0, Y: 50},
		"right":  {X: 100, Y: 50},
		"top":    {X: 50, Y: 100},
		"bottom": {X: 50, Y: 0},
	} {
		if insideWallBoundary(boundary, wall) {
			t.Fatalf("%s boundary counted as passed through", wall)
		}
		if inwardVelocity(wall) == (sim.Vec2{}) {
			t.Fatalf("%s inward velocity was zero", wall)
		}
	}
	if !insideWallBoundary(sim.Vec2{X: 50, Y: 99}, "top") {
		t.Fatal("top inside position was not counted as passed through")
	}
	if normalSignTowardInside("right") != -1 || normalSignTowardInside("top") != -1 {
		t.Fatal("right and top normal signs should point inward with -1")
	}

	w := wallCollisionWorldWithMass(1, "left", sim.Vec2{X: 0.5})
	if err := assertWallNormalVelocityReversed(w, map[string]string{"mass_id": "1", "wall": "left"}); err != nil {
		t.Fatalf("small reversed velocity rejected: %v", err)
	}
	w.domainWorld.Masses[0].Velocity = sim.Vec2{}
	if err := assertWallNormalVelocityReversed(w, map[string]string{"mass_id": "1", "wall": "left"}); err == nil {
		t.Fatal("zero normal velocity accepted as reversed")
	}

	passed := wallCollisionWorldWithMass(1, "left", sim.Vec2{})
	passed.domainWorld.Masses[0].Position = sim.Vec2{X: 1, Y: 50}
	if err := assertMassPassedThroughWall(passed, map[string]string{"mass_id": "1", "wall": "left"}); err != nil {
		t.Fatalf("passed-through position rejected: %v", err)
	}

	stuck := wallCollisionWorldWithMass(1, "left", sim.Vec2{})
	stuck.domainWorld.Masses[0].StuckWall = "left"
	if err := assertMassStuckToWall(stuck, map[string]string{"mass_id": "1", "wall": "left"}); err != nil {
		t.Fatalf("stuck mass rejected: %v", err)
	}
	if err := assertMassReleaseResult(stuck, map[string]string{"mass_id": "1", "wall": "left", "release_result": "stuck"}); err != nil {
		t.Fatalf("stuck release result rejected: %v", err)
	}
	if err := assertMassReleaseResult(&world{}, map[string]string{"mass_id": "1", "wall": "left", "release_result": "stuck"}); err == nil || !strings.Contains(err.Error(), "mass 1 not found") {
		t.Fatalf("missing mass release result error = %v", err)
	}
	if released, err := expectedReleased(map[string]string{}); err == nil || released {
		t.Fatalf("missing release result = %t, %v", released, err)
	}
	if released, err := expectedReleased(map[string]string{"release_result": "stuck"}); err != nil || released {
		t.Fatalf("stuck release result = %t, %v", released, err)
	}

	bounced := wallCollisionWorldWithMass(1, "left", sim.Vec2{X: 0.5})
	if err := assertMassDidNotBounce(bounced, map[string]string{"mass_id": "1", "wall": "left"}); err == nil {
		t.Fatal("small inward bounce accepted as no bounce")
	}
	zeroVelocity := wallCollisionWorldWithMass(1, "left", sim.Vec2{})
	if err := assertMassDidNotBounce(zeroVelocity, map[string]string{"mass_id": "1", "wall": "left"}); err != nil {
		t.Fatalf("zero normal velocity counted as bounce: %v", err)
	}
}

func TestWallCollisionSetupHelpers(t *testing.T) {
	w := &world{}
	mass := ensureCollisionMass(w, 7)
	if mass.ID != 7 || mass.Position != (sim.Vec2{X: 50, Y: 50}) || mass.Mass != 1 || mass.Elasticity != 1 {
		t.Fatalf("collision mass defaults = %#v", mass)
	}
	if w.domainWorld.Bounds.Width != 100 {
		t.Fatalf("collision world defaults = %#v", w.domainWorld)
	}

	stuck := wallCollisionWorldWithMass(1, "left", sim.Vec2{})
	stuck.domainWorld.Masses[0].StuckWall = "left"
	stuck.domainWorld.Parameters.Set("stickiness", "10")
	if err := pullMassAwayFromWall(stuck, map[string]string{"release_force": "sufficient"}); err != nil {
		t.Fatalf("pull mass away: %v", err)
	}
	if stuck.domainWorld.Time != 1 {
		t.Fatalf("pull did not step world time: %f", stuck.domainWorld.Time)
	}

	id, wall, err := collisionMassAndWall(map[string]string{})
	if err == nil || id != 0 || wall != "" {
		t.Fatalf("missing mass id parsed as id=%d wall=%q err=%v", id, wall, err)
	}
	id, wall, err = collisionMassAndWall(map[string]string{"mass_id": "1"})
	if err == nil || id != 0 || wall != "" {
		t.Fatalf("missing wall parsed as id=%d wall=%q err=%v", id, wall, err)
	}
	id, wall, err = collisionMassAndWall(map[string]string{"mass_id": "1", "wall": "left"})
	if err != nil || id != 1 || wall != "left" {
		t.Fatalf("valid mass wall parsed as id=%d wall=%q err=%v", id, wall, err)
	}
}

func wallCollisionWorldWithMass(id int, wall string, velocity sim.Vec2) *world {
	w := &world{}
	mass := ensureCollisionMass(w, id)
	mass.Position = insideCollisionPosition(wall)
	mass.Velocity = velocity
	return w
}

func TestStateSaveRestoreHelpersValidateInputs(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "change state missing example value",
			err:  changeApplicationState(&world{}, nil),
			want: "changed_state",
		},
		{
			name: "assert state missing example value",
			err:  assertApplicationStateWorld(&world{appGame: app.NewGame()}, nil),
			want: "memory_state",
		},
		{
			name: "assert state unsupported value",
			err:  assertApplicationStateWorld(&world{appGame: app.NewGame()}, map[string]string{"memory_state": "unknown"}),
			want: "unsupported state",
		},
		{
			name: "file operation missing example value",
			err:  runStateFileOperation(&world{appGame: app.NewGame()}, nil),
			want: "file_operation",
		},
		{
			name: "file operation unsupported value",
			err:  runStateFileOperation(&world{appGame: app.NewGame()}, map[string]string{"file_operation": "delete file"}),
			want: "unsupported file operation",
		},
		{
			name: "replace unsupported state",
			err:  replaceApplicationWorld(&world{appGame: app.NewGame()}, "unknown"),
			want: "unsupported state",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == nil || !strings.Contains(test.err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", test.err, test.want)
			}
		})
	}
}

func TestRestoreApplicationStateZeroCountLeavesWorldUnchanged(t *testing.T) {
	w := &world{}
	if err := setApplicationStateWorld(w, "A"); err != nil {
		t.Fatalf("set state A: %v", err)
	}
	if err := saveApplicationState(w, nil); err != nil {
		t.Fatalf("save state: %v", err)
	}
	if err := replaceApplicationWorld(w, "B"); err != nil {
		t.Fatalf("replace state B: %v", err)
	}
	if err := restoreApplicationState(w, 0); err != nil {
		t.Fatalf("restore zero times: %v", err)
	}
	if err := assertApplicationStateWorld(w, map[string]string{"memory_state": "B"}); err != nil {
		t.Fatalf("zero restores changed world: %v", err)
	}
}

func TestStateAWorldIncludesExpectedMassAndSpringDetails(t *testing.T) {
	world := stateAWorld()
	assertMasses(t, world.Masses, []sim.Mass{
		{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 2, Elasticity: 0.6, Fixed: true},
		{ID: 2, Position: sim.Vec2{X: 40, Y: 20}, Mass: 3, Elasticity: 0.7},
	})
	assertSprings(t, world.Springs, []sim.Spring{
		{ID: 3, A: 0, B: 1, MassA: 1, MassB: 2, RestLength: 30, Stiffness: 8, SpringConstant: 8, Damping: 0.4},
	})
	if got := world.Parameters.Value("current mass"); got != "state-a" {
		t.Fatalf("current mass = %q", got)
	}
	if !world.Parameters.Walls["left"] {
		t.Fatal("left wall was not enabled")
	}
}

func TestSimulationStateEqualDetectsEachComponent(t *testing.T) {
	base := stateAWorld()
	same := stateAWorld()
	if !simulationStateEqual(base, same) {
		t.Fatal("equal states reported different")
	}

	changedMasses := stateAWorld()
	changedMasses.Masses[0].Mass = 99
	if simulationStateEqual(base, changedMasses) {
		t.Fatal("mass difference reported equal")
	}

	changedSprings := stateAWorld()
	changedSprings.Springs[0].MassA = 2
	if simulationStateEqual(base, changedSprings) {
		t.Fatal("spring difference reported equal")
	}

	changedParameters := stateAWorld()
	changedParameters.Parameters.Set("current mass", "other")
	if simulationStateEqual(base, changedParameters) {
		t.Fatal("parameter difference reported equal")
	}
}

func TestSelectedObjectParameterHelpersValidateInputs(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "mass assertion missing mass id",
			err:  assertMassControlValue(&world{}, map[string]string{"control": "mass", "value": "1"}),
			want: "mass_id",
		},
		{
			name: "mass assertion missing control value",
			err:  assertMassControlValue(&world{domainWorld: worldWithParameterMass(1)}, map[string]string{"mass_id": "1"}),
			want: "control",
		},
		{
			name: "spring assertion missing spring id",
			err:  assertSpringControlValue(&world{}, map[string]string{"control": "Kspring", "value": "1"}),
			want: "spring_id",
		},
		{
			name: "spring assertion missing control value",
			err:  assertSpringControlValue(&world{domainWorld: worldWithParameterSpring(1)}, map[string]string{"spring_id": "1"}),
			want: "control",
		},
		{
			name: "rest length assertion missing value",
			err:  assertSelectedSpringRestLength(&world{domainWorld: worldWithParameterSpring(1)}, map[string]string{"spring_id": "1"}),
			want: "current_length",
		},
		{
			name: "future object assertion missing object type",
			err:  assertFutureObjectUsesControlValue(&world{}, map[string]string{"control": "mass", "value": "2"}),
			want: "object_type",
		},
		{
			name: "future mass assertion unsupported control",
			err:  assertFutureMassControl(&world{}, "unsupported", "2"),
			want: "unsupported mass control",
		},
		{
			name: "future spring assertion unsupported control",
			err:  assertFutureSpringControl(&world{domainWorld: worldWithParameterSpringEndpoints()}, "unsupported", "2"),
			want: "unsupported spring control",
		},
		{
			name: "int and float invalid integer",
			err:  intAndFloatError(map[string]string{"spring_id": "bad", "current_length": "42"}),
			want: "invalid integer",
		},
		{
			name: "float assertion invalid value",
			err:  assertStringFloat("mass", 1, "bad"),
			want: "invalid float",
		},
		{
			name: "bool assertion invalid value",
			err:  assertStringBool("fixed", true, "bad"),
			want: "invalid bool",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == nil || !strings.Contains(test.err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", test.err, test.want)
			}
		})
	}

	if err := assertMassControlValue(&world{domainWorld: worldWithParameterMass(1)}, map[string]string{"mass_id": "1"}); err == nil || err.Error() != "missing example value control" {
		t.Fatalf("mass missing control error = %v", err)
	}
	if err := assertSpringControlValue(&world{domainWorld: worldWithParameterSpring(1)}, map[string]string{"spring_id": "1"}); err == nil || err.Error() != "missing example value control" {
		t.Fatalf("spring missing control error = %v", err)
	}

	id, value, err := intAndFloat(map[string]string{"spring_id": "bad", "current_length": "42"}, "spring_id", "current_length")
	if err == nil || id != 0 || value != 0 {
		t.Fatalf("invalid integer intAndFloat = %d, %f, %v", id, value, err)
	}
}

func TestSelectedObjectParameterHelpersCreateExpectedObjects(t *testing.T) {
	w := &world{}
	if err := createSelectedParameterMass(w, map[string]string{"mass_id": "5"}); err != nil {
		t.Fatalf("create selected mass: %v", err)
	}
	mass, ok := ensureDomainWorld(w).MassByID(5)
	if !ok {
		t.Fatal("mass 5 was not created")
	}
	if mass.Mass != 1 || mass.Elasticity != 0.2 {
		t.Fatalf("mass 5 defaults = %#v", mass)
	}
	if !ensureMouseEditor(w).SelectedMasses[5] {
		t.Fatal("mass 5 was not selected")
	}

	springWorld := &world{}
	if err := addParameterSpring(springWorld, 8, 42); err != nil {
		t.Fatalf("add parameter spring: %v", err)
	}
	assertSimulationMassPosition(t, ensureDomainWorld(springWorld), 1, sim.Vec2{X: 0, Y: 20})
	assertSimulationMassPosition(t, ensureDomainWorld(springWorld), 2, sim.Vec2{X: 42, Y: 20})
	assertParameterMassDefaults(t, ensureDomainWorld(springWorld), 1)
	assertParameterMassDefaults(t, ensureDomainWorld(springWorld), 2)
	spring, ok := ensureDomainWorld(springWorld).SpringByID(8)
	if !ok {
		t.Fatal("spring 8 was not created")
	}
	assertSprings(t, []sim.Spring{spring}, []sim.Spring{
		{ID: 8, A: 0, B: 1, MassA: 1, MassB: 2, RestLength: 1, Stiffness: 8, SpringConstant: 8, Damping: 0.2},
	})

	if isSpringControl("RestLength") {
		t.Fatal("RestLength should not be a directly editable spring control")
	}
	if !isSpringControl("Kdamp") {
		t.Fatal("Kdamp should be a spring control")
	}
}

func assertParameterMassDefaults(t *testing.T, world *sim.Simulation, id int) {
	t.Helper()
	mass, ok := world.MassByID(id)
	if !ok {
		t.Fatalf("mass %d not found", id)
	}
	if mass.Mass != 1 {
		t.Fatalf("mass %d default mass = %f", id, mass.Mass)
	}
}

func intAndFloatError(example map[string]string) error {
	_, _, err := intAndFloat(example, "spring_id", "current_length")
	return err
}

func worldWithParameterMass(id int) *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: id, Mass: 1})
	return world
}

func worldWithParameterSpring(id int) *sim.Simulation {
	world := worldWithParameterSpringEndpoints()
	_ = world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2, RestLength: 1, SpringConstant: 1})
	return world
}

func worldWithParameterSpringEndpoints() *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 20}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 40, Y: 20}, Mass: 1})
	return world
}

func TestRunFeatureExecutesRenderVisibleControlsFeature(t *testing.T) {
	runFeatureFile(t, "features/024_1_render_visible_controls.feature")
}

func TestRunFeatureExecutesClickableVisibleControlsFeature(t *testing.T) {
	runFeatureFile(t, "features/024_2_clickable_visible_controls.feature")
}

func TestVisibleControlsApplicationStateSetup(t *testing.T) {
	tests := []struct {
		name   string
		state  string
		before func(*app.Game)
		assert func(*testing.T, *app.Game)
	}{
		{
			name:  "running",
			state: "running",
			before: func(game *app.Game) {
				game.SetPaused(true)
			},
			assert: func(t *testing.T, game *app.Game) {
				t.Helper()
				if game.Paused() {
					t.Fatal("game remained paused")
				}
			},
		},
		{
			name:  "object counts",
			state: "object counts",
			assert: func(*testing.T, *app.Game) {
			},
		},
		{
			name:  "saved",
			state: "saved",
			before: func(game *app.Game) {
				game.SetDirty(true)
			},
			assert: func(t *testing.T, game *app.Game) {
				t.Helper()
				if got := game.DrawFrameReport().StatusFields["file state"]; !strings.Contains(got, "saved") {
					t.Fatalf("file state = %q, want saved", got)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := app.NewGame()
			if test.before != nil {
				test.before(game)
			}
			w := &world{appGame: game}
			if err := setVisibleControlsApplicationState(w, map[string]string{"state": test.state}); err != nil {
				t.Fatalf("set state returned error: %v", err)
			}
			test.assert(t, game)
		})
	}
}

func TestVisibleControlsApplicationStateRejectsUnsupportedState(t *testing.T) {
	err := setVisibleControlsApplicationState(&world{appGame: app.NewGame()}, map[string]string{"state": "unsupported"})
	if err == nil || !strings.Contains(err.Error(), "unsupported visible controls state") {
		t.Fatalf("expected unsupported state error, got %v", err)
	}
}

func TestRunFeatureExecutesOriginalDemoCorpusFeature(t *testing.T) {
	runFeatureFile(t, "features/025_original_demo_corpus.feature")
}

func TestRunFeatureExecutesMassCollisionFeature(t *testing.T) {
	runFeatureFile(t, "features/026_mass_collision.feature")
}

func TestRenderWorldHelpersValidateInputs(t *testing.T) {
	if err := createApplicationWorldState(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing world state")
	}
	if err := createApplicationWorldState(&world{}, map[string]string{"world_state": "unsupported"}); err == nil {
		t.Fatal("expected unsupported world state")
	}
	if err := assertVisibleRepresentation(&world{}, map[string]string{}); err == nil || !strings.Contains(err.Error(), "object") {
		t.Fatalf("expected missing object, got %v", err)
	}
	if err := assertSpringLineVisibility(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing spring visibility")
	}
	if err := assertSpringLineVisibility(&world{}, map[string]string{"spring_visibility": "blurred"}); err == nil {
		t.Fatal("expected unsupported spring visibility")
	}
	if visible, ok := booleanState("visible", springVisibilityStates); !ok || !visible {
		t.Fatalf("visible spring state = %t, %t", visible, ok)
	}
	if hidden, ok := booleanState("hidden", springVisibilityStates); !ok || hidden {
		t.Fatalf("hidden spring state = %t, %t", hidden, ok)
	}
}

func TestRenderableObjectSetupsCreateExpectedWorlds(t *testing.T) {
	tests := []struct {
		name      string
		masses    []sim.Mass
		springs   []sim.Spring
		wallLeft  bool
		visibleAs string
	}{
		{
			name:      "movable mass",
			masses:    []sim.Mass{{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1}},
			visibleAs: "movable mass",
		},
		{
			name:      "fixed mass",
			masses:    []sim.Mass{{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1, Fixed: true}},
			visibleAs: "fixed mass",
		},
		{
			name: "spring",
			masses: []sim.Mass{
				{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1, Fixed: true},
				{ID: 2, Position: sim.Vec2{X: 40, Y: 20}, Mass: 1},
			},
			springs:   []sim.Spring{{ID: 1, A: 0, B: 1, MassA: 1, MassB: 2, RestLength: 20, Stiffness: 12, SpringConstant: 12}},
			visibleAs: "spring",
		},
		{name: "enabled wall", wallLeft: true, visibleAs: "enabled wall"},
		{
			name:      "selection",
			masses:    []sim.Mass{{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1}},
			visibleAs: "selection",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := emptyRenderGame()
			if err := addRenderableObject(game, test.name); err != nil {
				t.Fatal(err)
			}
			assertMasses(t, game.World().Masses, test.masses)
			assertSprings(t, game.World().Springs, test.springs)
			if game.World().Parameters.Walls["left"] != test.wallLeft {
				t.Fatalf("left wall = %t, want %t", game.World().Parameters.Walls["left"], test.wallLeft)
			}
			if result := game.RenderWorld(); !result.HasVisibleRepresentation(test.visibleAs) {
				t.Fatalf("render result missing %q: %#v", test.visibleAs, result.Representations)
			}
		})
	}

	if err := addRenderableObject(emptyRenderGame(), "unsupported"); err == nil {
		t.Fatal("expected unsupported renderable object")
	}
}

func TestDemoFileHelpersReportFailures(t *testing.T) {
	if err := assertDemoFileAdded(nil, map[string]string{"demo_file": "pendulum.xsp"}); err != nil {
		t.Fatal(err)
	}
	if err := assertDemoFileAdded(nil, map[string]string{"demo_file": "unknown.xsp"}); err == nil {
		t.Fatal("expected unknown demo file error")
	}
	if err := assertDemoFileValid(nil, map[string]string{"demo_file": "missing.xsp"}); err == nil {
		t.Fatal("expected missing demo file validity error")
	}
	invalidDemoPath := repoPath(filepath.Join("demos", "invalid-test.xsp"))
	writeSource(t, invalidDemoPath, "not xsp\n")
	t.Cleanup(func() { _ = os.Remove(invalidDemoPath) })
	if err := assertDemoFileValid(nil, map[string]string{"demo_file": "invalid-test.xsp"}); err == nil {
		t.Fatal("expected invalid demo file error")
	}
	if err := assertDemoFileHumanReadable(nil, map[string]string{"demo_file": "missing.xsp"}); err == nil {
		t.Fatal("expected missing demo file readability error")
	}

	demoPath := repoPath(filepath.Join("demos", "unreadable-test.xsp"))
	writeSource(t, demoPath, " mass 1 0 0 0 0 1 1 false\n")
	t.Cleanup(func() { _ = os.Remove(demoPath) })
	if err := assertDemoFileHumanReadable(nil, map[string]string{"demo_file": "unreadable-test.xsp"}); err == nil {
		t.Fatal("expected unreadable demo file error")
	}
	if err := assertDemoLinesReadable("mass 1 0 0 0 0 1 1 false\n mass 2 0 0 0 0 1 1 false\n"); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("expected line 2 surrounding whitespace error, got %v", err)
	}
	if err := assertDemoLinesReadable("mass 1 0 0 0 0 1 1 false\n\n"); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("expected line 2 blank line error, got %v", err)
	}

	if err := assertDemoLoadedFeature(&world{}, nil); err == nil {
		t.Fatal("expected missing required feature error")
	}
	if err := assertDemoLoadedFeature(&world{xspWorld: sim.NewWorld()}, map[string]string{"required_feature": "unsupported"}); err == nil {
		t.Fatal("expected unsupported demo feature error")
	}
}

func TestDemoFeatureHelpersValidateBoundaries(t *testing.T) {
	if err := assertDemoHasMultipleSprings(&sim.Simulation{Springs: []sim.Spring{{ID: 1}, {ID: 2}}}); err != nil {
		t.Fatal(err)
	}
	if err := assertDemoHasMultipleSprings(&sim.Simulation{Springs: []sim.Spring{{ID: 1}}}); err == nil {
		t.Fatal("expected single spring to fail multiple spring assertion")
	}
	if err := assertDemoHasFixedMass(&sim.Simulation{Masses: []sim.Mass{{ID: 1, Fixed: true}}}); err != nil {
		t.Fatal(err)
	}
	if err := assertDemoHasFixedMass(&sim.Simulation{Masses: []sim.Mass{{ID: 1}}}); err == nil {
		t.Fatal("expected no fixed mass error")
	}
}

func TestPackagingDocsHelpersReportFailures(t *testing.T) {
	w := &world{documentation: "go test"}
	if err := assertDocumentedCommand(w, nil); err == nil {
		t.Fatal("expected missing command documentation assertion error")
	}
	if err := assertDocumentedCommandPassed(&world{}, nil); err == nil {
		t.Fatal("expected missing command result assertion error")
	}
	if err := assertDocumentationExplains(&world{}, nil); err == nil {
		t.Fatal("expected missing topic documentation assertion error")
	}
	if err := assertDocumentationExplains(&world{documentation: ""}, map[string]string{"topic": "creating a simulation"}); err == nil {
		t.Fatal("expected missing topic terms error")
	}
	if _, err := commandFromExample(nil); err == nil {
		t.Fatal("expected missing command example error")
	}
	if command, err := commandFromExample(map[string]string{"command": "unit tests"}); err != nil || command.name != "go" {
		t.Fatalf("commandFromExample = %#v, %v", command, err)
	}
	if err := assertHandoffIncludesVerificationCommands(&world{handoffVerification: map[string]string{"go test": "passed"}}, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertHandoffIncludesVerificationResults(&world{handoffVerification: map[string]string{"go test": "passed"}}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestEditModeDetailsHelpersValidateSetupAndAssertions(t *testing.T) {
	w := &world{}
	if err := activateEditMode(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := addObjectNearPointer(w, map[string]string{"object_id": "3"}); err != nil {
		t.Fatal(err)
	}
	mass, err := editMassByID(w, 3)
	if err != nil {
		t.Fatal(err)
	}
	if mass.Position != (sim.Vec2{X: 30, Y: 0}) || mass.Mass != 1 || mass.Fixed {
		t.Fatalf("mass near pointer = %#v", mass)
	}

	if editPointerPosition(4) != (sim.Vec2{X: 40, Y: 0}) {
		t.Fatalf("pointer position = %#v", editPointerPosition(4))
	}
	if insideSelectionBoxPosition(2) != (sim.Vec2{X: 30, Y: 10}) {
		t.Fatalf("inside position = %#v", insideSelectionBoxPosition(2))
	}
	if outsideSelectionBoxPosition(2) != (sim.Vec2{X: 120, Y: 100}) {
		t.Fatalf("outside position = %#v", outsideSelectionBoxPosition(2))
	}

	if err := assertEditObjectPosition(w, nil); err == nil {
		t.Fatal("expected missing object id error")
	}
	if err := assertEditObjectPosition(w, map[string]string{"object_id": "3"}); err == nil {
		t.Fatal("expected missing expected position error")
	}
	if err := assertEditObjectPosition(w, map[string]string{"object_id": "3", "expected_position": "30,0"}); err != nil {
		t.Fatal(err)
	}
	if err := assertEditObjectPosition(w, map[string]string{"object_id": "99", "expected_position": "30,0"}); err == nil {
		t.Fatal("expected missing mass position error")
	}
	if err := assertEditSelection(w, nil); err == nil {
		t.Fatal("expected missing selection error")
	}
}

func TestEditModeDetailsSelectionHelpersValidateBranches(t *testing.T) {
	w := &world{}
	toggle, err := editClickToggle("left clicks")
	if err != nil {
		t.Fatal(err)
	}
	if toggle {
		t.Fatal("left clicks should replace selection")
	}
	toggle, err = editClickToggle("shift left clicks")
	if err != nil {
		t.Fatal(err)
	}
	if !toggle {
		t.Fatal("shift left clicks should toggle selection")
	}
	if toggle, err := editClickToggle("unsupported"); err == nil {
		t.Fatal("expected unsupported click action")
	} else if toggle {
		t.Fatal("unsupported click action should not toggle selection")
	}
	if err := setInitialEditSelection(w, map[string]string{"initial_selection": "1,2"}); err != nil {
		t.Fatal(err)
	}
	if selected := selectedEditMassIDs(ensureMouseEditor(w)); strings.Join(intStrings(selected), ",") != "1,2" {
		t.Fatalf("selection = %v", selected)
	}
	first, _ := editMassByID(w, 1)
	second, _ := editMassByID(w, 2)
	if first.Position != (sim.Vec2{X: 10, Y: 0}) || first.Fixed || second.Position != (sim.Vec2{X: 20, Y: 0}) || second.Fixed {
		t.Fatalf("selected mass positions = %#v %#v", first, second)
	}

	if err := addEditObjects(&world{}, map[string]string{"ids": "5,6"}, "ids", insideSelectionBoxPosition); err != nil {
		t.Fatal(err)
	}
	boxWorld := &world{}
	if err := addObjectsInsideSelectionBox(boxWorld, map[string]string{"inside_objects": "1,2"}); err != nil {
		t.Fatal(err)
	}
	if err := addObjectsOutsideSelectionBox(boxWorld, map[string]string{"outside_objects": "3"}); err != nil {
		t.Fatal(err)
	}
	ensureMouseEditor(boxWorld).SelectedMasses[3] = true
	if err := dragSelectionBox(boxWorld, map[string]string{"modifier": "none"}); err != nil {
		t.Fatal(err)
	}
	if err := assertEditSelection(boxWorld, map[string]string{"expected_selection": "1,2"}); err != nil {
		t.Fatal(err)
	}
	ensureMouseEditor(boxWorld).SelectedMasses[3] = true
	if err := dragSelectionBox(boxWorld, map[string]string{"modifier": "shift"}); err != nil {
		t.Fatal(err)
	}
	if err := assertEditSelection(boxWorld, map[string]string{"expected_selection": "1,2,3"}); err != nil {
		t.Fatal(err)
	}
	if err := dragSelectionBox(boxWorld, map[string]string{"modifier": "unsupported"}); err == nil {
		t.Fatal("expected unsupported selection-box modifier")
	}
	if id, err := parseEditIDPart("bad", "ids", "bad"); err == nil {
		t.Fatal("expected invalid edit id")
	} else if id != 0 {
		t.Fatalf("expected zero id on invalid edit id, got %d", id)
	}
}

func TestEditModeDetailsVelocityHelpersValidateBranches(t *testing.T) {
	w := &world{}
	if err := addSelectedMassWithFixedState(w, map[string]string{"mass_id": "4", "fixed": "true"}); err != nil {
		t.Fatal(err)
	}
	mass, err := editMassByID(w, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !mass.Fixed || mass.Position != (sim.Vec2{X: 40, Y: 0}) || mass.Velocity != editInitialVelocity {
		t.Fatalf("fixed selected mass = %#v", mass)
	}

	setEditMassVelocity(w, 4, sim.Vec2{X: 1, Y: 2})
	if err := assertEditMassVelocity(w, map[string]string{"mass_id": "4", "expected_velocity": "1,2"}); err != nil {
		t.Fatal(err)
	}
	if err := assertEditMassVelocity(w, nil); err == nil {
		t.Fatal("expected missing mass id velocity error")
	}
	if err := assertEditMassVelocity(w, map[string]string{"mass_id": "99", "expected_velocity": "1,2"}); err == nil || !strings.Contains(err.Error(), "mass 99 not found") {
		t.Fatal("expected missing mass velocity error")
	}
	if err := assertEditMassExpectedVelocity(4, mass, nil); err == nil {
		t.Fatal("expected missing expected velocity")
	}
	if err := assertEditMassExpectedVelocity(4, mass, map[string]string{"expected_velocity": "bad"}); err == nil || !strings.Contains(err.Error(), "invalid position") {
		t.Fatal("expected invalid expected velocity")
	}
	if err := assertEditMassExpectedVelocity(4, sim.Mass{Velocity: editInitialVelocity}, map[string]string{"expected_velocity": "unchanged"}); err != nil {
		t.Fatal(err)
	}
}

func TestSpringModeMouseHelpersValidateSetupAndAssertions(t *testing.T) {
	w := &world{}
	if err := activateSpringMode(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertSpringCreationResult(w, nil); err == nil || !strings.Contains(err.Error(), "missing example value") {
		t.Fatalf("expected missing spring result, got %v", err)
	}
	if err := assertPendingSpringBehavior(w, nil); err == nil || !strings.Contains(err.Error(), "missing example value") {
		t.Fatalf("expected missing pending behavior, got %v", err)
	}
	if err := ensureSpringModeMass(w, 3, springModeMassPosition(3)); err != nil {
		t.Fatal(err)
	}
	mass, err := editMassByID(w, 3)
	if err != nil {
		t.Fatal(err)
	}
	if mass.Position != (sim.Vec2{X: 60}) || mass.Mass != 1 {
		t.Fatalf("spring mode mass = %#v", mass)
	}
	if err := createSpringWithLength(w, map[string]string{"creation_length": "30"}); err != nil {
		t.Fatal(err)
	}
	if err := assertCreatedSpringEndpoints(w, 1, 2); err != nil {
		t.Fatal(err)
	}
	if err := assertCreatedSpringRestLength(w, map[string]string{"creation_length": "30"}); err != nil {
		t.Fatal(err)
	}
	if err := assertCreatedSpringFloat(w, nil, "missing", func(sim.Spring) float64 { return 0 }); err == nil || !strings.Contains(err.Error(), "missing example value") {
		t.Fatalf("expected missing spring float parameter, got %v", err)
	}
	if err := assertCreatedSpringFloat(&world{domainWorld: sim.NewWorld()}, map[string]string{"kspring": "12"}, "kspring", func(sim.Spring) float64 { return 12 }); err == nil || !strings.Contains(err.Error(), "created spring 0 not found") {
		t.Fatalf("expected missing created spring, got %v", err)
	}
}

func TestSpringModeMouseParsersRejectMalformedValues(t *testing.T) {
	if id, ok := parseNearMass("beside mass 2"); ok || id != 0 {
		t.Fatal("expected invalid near-mass prefix")
	}
	if id, ok := parseNearMass("near node 2"); ok || id != 0 {
		t.Fatal("expected invalid near-mass noun")
	}
	if _, ok := parseNearMass("near mass bad"); ok {
		t.Fatal("expected invalid near-mass id")
	}

	if massA, massB, ok := parseCreatedSpringResult("make spring between 1 and 2"); ok || massA != 0 || massB != 0 {
		t.Fatal("expected invalid created-spring prefix")
	}
	if massA, massB, ok := parseCreatedSpringResult("create spring between 1 to 2"); ok || massA != 0 || massB != 0 {
		t.Fatal("expected invalid created-spring separator")
	}
	if _, _, ok := parseCreatedSpringResult("create spring between bad and 2"); ok {
		t.Fatal("expected invalid first created-spring id")
	}
	if _, _, ok := parseCreatedSpringResult("create spring between 1 and bad"); ok {
		t.Fatal("expected invalid second created-spring id")
	}
}

func TestSpringModeMouseAssertionsDetectIndividualBranches(t *testing.T) {
	if err := assertSpringDiscarded(&world{springCreated: true, domainWorld: sim.NewWorld()}); err == nil {
		t.Fatal("expected created flag to fail discard assertion")
	}
	wWithSpring := &world{domainWorld: sim.NewWorld()}
	if err := wWithSpring.domainWorld.AddMass(sim.Mass{ID: 1, Mass: 1}); err != nil {
		t.Fatal(err)
	}
	if err := wWithSpring.domainWorld.AddMass(sim.Mass{ID: 2, Mass: 1}); err != nil {
		t.Fatal(err)
	}
	if err := wWithSpring.domainWorld.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2}); err != nil {
		t.Fatal(err)
	}
	if err := assertSpringDiscarded(wWithSpring); err == nil {
		t.Fatal("expected existing spring to fail discard assertion")
	}

	if err := assertCreatedSpringEndpoints(&world{domainWorld: sim.NewWorld()}, 1, 2); err == nil || !strings.Contains(err.Error(), "created spring 0 not found") {
		t.Fatalf("expected missing created spring endpoints, got %v", err)
	}
	w := &world{domainWorld: sim.NewWorld(), createdSpringID: 1}
	if err := w.domainWorld.AddMass(sim.Mass{ID: 1, Mass: 1}); err != nil {
		t.Fatal(err)
	}
	if err := w.domainWorld.AddMass(sim.Mass{ID: 2, Mass: 1}); err != nil {
		t.Fatal(err)
	}
	if err := w.domainWorld.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2}); err != nil {
		t.Fatal(err)
	}
	if err := assertCreatedSpringEndpoints(w, 1, 2); err == nil {
		t.Fatal("expected unset created flag to fail endpoint assertion")
	}
	w.springCreated = true
	if err := assertCreatedSpringEndpoints(w, 9, 2); err == nil {
		t.Fatal("expected mass A mismatch")
	}
	if err := assertCreatedSpringEndpoints(w, 1, 9); err == nil {
		t.Fatal("expected mass B mismatch")
	}

	editor := edit.NewEditor(w.domainWorld)
	if err := discardTemporarySpring(w, editor, edit.SpringButtonMiddle); err == nil || !strings.Contains(err.Error(), "temporary spring release") {
		t.Fatalf("expected temporary spring release error, got %v", err)
	}
}

func assertMasses(t *testing.T, actual, expected []sim.Mass) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("mass count = %d, want %d", len(actual), len(expected))
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("mass %d = %#v, want %#v", i, actual[i], expected[i])
		}
	}
}

func assertSprings(t *testing.T, actual, expected []sim.Spring) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("spring count = %d, want %d", len(actual), len(expected))
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("spring %d = %#v, want %#v", i, actual[i], expected[i])
		}
	}
}

func TestMouseEditingHelpersReportFailures(t *testing.T) {
	if err := setMouseEditorMode(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing mode")
	}
	if err := clickMouseEditor(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing pointer position")
	}
	if err := clickMouseEditor(&world{}, map[string]string{"pointer_position": "1"}); err == nil {
		t.Fatal("expected invalid pointer position")
	}
	if err := setMouseGridSnap(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing grid snap")
	}
	if err := setMouseGridSnap(&world{}, map[string]string{"grid_snap": "maybe"}); err == nil {
		t.Fatal("expected unsupported grid snap")
	}
	if enabled, ok := booleanState("enabled", mouseGridSnapStates); !ok || !enabled {
		t.Fatalf("enabled grid snap = %t, %t", enabled, ok)
	}
	if disabled, ok := booleanState("disabled", mouseGridSnapStates); !ok || disabled {
		t.Fatalf("disabled grid snap = %t, %t", disabled, ok)
	}
	if err := setMouseGridSnapSize(&world{}, map[string]string{"snap_size": "bad"}); err == nil {
		t.Fatal("expected invalid snap size")
	}
	if err := createMouseSpring(&world{}, map[string]string{"mass_a": "1"}); err == nil {
		t.Fatal("expected missing spring endpoint id")
	}
	if err := dragMouseMass(&world{}, map[string]string{"mass_id": "1"}); err == nil {
		t.Fatal("expected missing target position")
	}
	if err := assertMouseMassPosition(&world{}, map[string]string{"mass_id": "1"}); err == nil {
		t.Fatal("expected missing expected position")
	}
	if _, err := positionValue(map[string]string{"position": "bad,2"}, "position"); err == nil {
		t.Fatal("expected invalid x position")
	}
	if _, err := positionValue(map[string]string{"position": "1,bad"}, "position"); err == nil {
		t.Fatal("expected invalid y position")
	}
}

func TestMouseEditingHelpersReportMissingCreatedObjects(t *testing.T) {
	w := &world{domainWorld: sim.NewWorld(), createdMassID: 99, createdSpringID: 77}

	if _, err := createdMouseMass(w); err == nil {
		t.Fatal("expected missing created mass")
	}
	if err := assertCreatedMassPosition(w, map[string]string{"expected_position": "1,2"}); err == nil {
		t.Fatal("expected missing created mass position")
	}
	if err := assertCreatedMassDefaults(w, nil); err == nil {
		t.Fatal("expected missing created mass defaults")
	}
	if _, err := createdMouseSpring(w); err == nil {
		t.Fatal("expected missing created spring")
	}
	if err := assertMouseSpringEndpoints(w, map[string]string{"mass_a": "1", "mass_b": "2"}); err == nil {
		t.Fatal("expected missing created spring endpoints")
	}
	if err := assertMouseSpringDefaults(w, nil); err == nil {
		t.Fatal("expected missing created spring defaults")
	}
	if err := assertMouseMassID(w, map[string]string{"mass_id": "99"}); err == nil {
		t.Fatal("expected missing mouse mass")
	}
}

func TestMouseEditingAssertionsDetectIndividualFieldMismatches(t *testing.T) {
	w := &world{domainWorld: sim.NewWorld(), createdMassID: 1, createdSpringID: 1}
	_ = w.domainWorld.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 12, Y: 8}, Mass: mouseDefaultMass, Elasticity: mouseDefaultElasticity})
	_ = w.domainWorld.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 8}, Mass: 1})
	_ = w.domainWorld.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, SpringConstant: 12, Damping: 0.7})

	if err := assertCreatedMassPosition(w, map[string]string{"expected_position": "12,8"}); err != nil {
		t.Fatal(err)
	}
	if err := assertCreatedMassPosition(w, map[string]string{"expected_position": "13,8"}); err == nil {
		t.Fatal("expected created mass x mismatch")
	}
	if err := assertCreatedMassDefaults(w, nil); err != nil {
		t.Fatal(err)
	}
	setCreatedMass(t, w, sim.Mass{ID: 1, Position: sim.Vec2{X: 12, Y: 8}, Mass: 9, Elasticity: mouseDefaultElasticity})
	if err := assertCreatedMassDefaults(w, nil); err == nil {
		t.Fatal("expected created mass mass mismatch")
	}
	setCreatedMass(t, w, sim.Mass{ID: 1, Position: sim.Vec2{X: 12, Y: 8}, Mass: mouseDefaultMass, Elasticity: 9})
	if err := assertCreatedMassDefaults(w, nil); err == nil {
		t.Fatal("expected created mass elasticity mismatch")
	}

	if err := assertMouseSpringEndpoints(w, map[string]string{"mass_a": "1", "mass_b": "2"}); err != nil {
		t.Fatal(err)
	}
	if err := assertMouseSpringEndpoints(w, map[string]string{"mass_a": "2", "mass_b": "2"}); err == nil {
		t.Fatal("expected spring mass_a mismatch")
	}
	if err := assertMouseSpringEndpoints(w, map[string]string{"mass_a": "1", "mass_b": "1"}); err == nil {
		t.Fatal("expected spring mass_b mismatch")
	}
	if err := assertMouseSpringDefaults(w, nil); err != nil {
		t.Fatal(err)
	}
	setCreatedSpring(t, w, sim.Spring{ID: 1, MassA: 1, MassB: 2, SpringConstant: 9, Damping: 0.7})
	if err := assertMouseSpringDefaults(w, nil); err == nil {
		t.Fatal("expected spring constant mismatch")
	}
	setCreatedSpring(t, w, sim.Spring{ID: 1, MassA: 1, MassB: 2, SpringConstant: 12, Damping: 9})
	if err := assertMouseSpringDefaults(w, nil); err == nil {
		t.Fatal("expected spring damping mismatch")
	}
}

func TestMouseMassHelpersCreatePositionFromID(t *testing.T) {
	w := &world{}

	if err := addMouseMassA(w, map[string]string{"mass_a": "3"}); err != nil {
		t.Fatal(err)
	}
	if err := addMouseMassB(w, map[string]string{"mass_b": "4"}); err != nil {
		t.Fatal(err)
	}
	assertMouseMass(t, w, 3, sim.Vec2{X: 60, Y: 20})
	assertMouseMass(t, w, 4, sim.Vec2{X: 80, Y: 20})
}

func setCreatedMass(t *testing.T, w *world, mass sim.Mass) {
	t.Helper()
	for i := range w.domainWorld.Masses {
		if w.domainWorld.Masses[i].ID == mass.ID {
			w.domainWorld.Masses[i] = mass
			return
		}
	}
	t.Fatalf("mass %d not found", mass.ID)
}

func setCreatedSpring(t *testing.T, w *world, spring sim.Spring) {
	t.Helper()
	for i := range w.domainWorld.Springs {
		if w.domainWorld.Springs[i].ID == spring.ID {
			w.domainWorld.Springs[i] = spring
			return
		}
	}
	t.Fatalf("spring %d not found", spring.ID)
}

func assertMouseMass(t *testing.T, w *world, id int, position sim.Vec2) {
	t.Helper()
	mass, ok := w.domainWorld.MassByID(id)
	if !ok {
		t.Fatalf("mass %d not found", id)
	}
	if mass.Position != position || mass.Mass != 1 {
		t.Fatalf("mass %d = %#v", id, mass)
	}
}

func TestSelectionEditingHelpersReportFailures(t *testing.T) {
	w := &world{}
	if err := assertObjectSelected(w, map[string]string{"object_type": "mass", "id": "1"}); err == nil {
		t.Fatal("expected unselected mass error")
	}
	if err := addSelectionObject(w, "mass", 1); err != nil {
		t.Fatal(err)
	}
	if err := assertObjectDeleted(w, map[string]string{"object_type": "mass", "id": "1"}); err == nil {
		t.Fatal("expected existing mass error")
	}
	if err := assertMassOneDeleted(w, nil); err == nil {
		t.Fatal("expected mass one to still exist")
	}
	if objectSelected(w, "unsupported", 1) {
		t.Fatal("unsupported object type reported selected")
	}
	if objectExists(w, "unsupported", 1) {
		t.Fatal("unsupported object type reported existing")
	}
}

func TestSelectionEditingHelpersCreateExpectedObjects(t *testing.T) {
	w := &world{}
	if err := addSelectionObject(w, "mass", 5); err != nil {
		t.Fatal(err)
	}
	mass, _ := w.domainWorld.MassByID(5)
	if mass.Position != (sim.Vec2{X: 5, Y: 1}) || mass.Mass != 1 {
		t.Fatalf("mass = %#v", mass)
	}

	w = &world{}
	if err := addSelectionObject(w, "spring", 7); err != nil {
		t.Fatal(err)
	}
	spring, _ := w.domainWorld.SpringByID(7)
	massOne, _ := w.domainWorld.MassByID(1)
	massTwo, _ := w.domainWorld.MassByID(2)
	if spring.MassA != 1 || spring.MassB != 2 || massOne.Position.X != 10 || massTwo.Position.X != 20 {
		t.Fatalf("spring = %#v massOne = %#v massTwo = %#v", spring, massOne, massTwo)
	}
}

func TestSelectionEditingSelectedSetHelpersRecordState(t *testing.T) {
	w := &world{}
	if err := createSelectedObjectSet(w, map[string]string{"object_set": "one mass"}); err != nil {
		t.Fatal(err)
	}
	if len(w.originalMassIDs) != 1 || w.originalMassIDs[0] != 1 || !w.mouseEditor.MassSelected(1) {
		t.Fatalf("mass selection state = %#v selected=%t", w.originalMassIDs, w.mouseEditor.MassSelected(1))
	}

	w = &world{}
	if err := createSelectedObjectSet(w, map[string]string{"object_set": "two masses and a spring"}); err != nil {
		t.Fatal(err)
	}
	if len(w.originalMassIDs) != 2 || len(w.originalSpringIDs) != 1 || !w.mouseEditor.SpringSelected(3) {
		t.Fatalf("spring selection state = masses %#v springs %#v", w.originalMassIDs, w.originalSpringIDs)
	}
}

func TestSelectionEditingAllSelectedRequiresEveryID(t *testing.T) {
	w := &world{}
	if err := addSelectionObject(w, "mass", 1); err != nil {
		t.Fatal(err)
	}
	if err := addSelectionObject(w, "mass", 2); err != nil {
		t.Fatal(err)
	}
	if err := ensureMouseEditor(w).SelectMass(1); err != nil {
		t.Fatal(err)
	}
	if err := assertEveryMassSelected(w, nil); err == nil {
		t.Fatal("expected mass 2 not selected")
	}
}

func TestSelectionEditingDuplicateIDHelpersReportFailures(t *testing.T) {
	if !repeatedID([]int{4, 4}) {
		t.Fatal("expected repeated id")
	}
	if !anySharedID([]int{4}, []int{4}) {
		t.Fatal("expected shared id")
	}
	if !idSet([]int{9})[9] {
		t.Fatal("expected id in set")
	}
	if err := assertUniqueNewIDs("mass", []int{2, 2}, nil); err == nil {
		t.Fatal("expected repeated id error")
	}
	if err := assertUniqueNewIDs("mass", []int{2}, []int{2}); err == nil {
		t.Fatal("expected shared id error")
	}

	w := &world{duplicated: edit.DuplicatedObjects{SpringIDs: []int{3, 3}}}
	if err := assertDuplicatedUniqueIDs(w, nil); err == nil {
		t.Fatal("expected duplicated spring id error")
	}
}

func TestSelectionEditingDuplicateIndependenceReportsFailures(t *testing.T) {
	w := &world{domainWorld: sim.NewWorld(), originalMassIDs: []int{1}, duplicated: edit.DuplicatedObjects{MassIDs: []int{2}}}
	_ = w.domainWorld.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 5}, Mass: 1})
	if err := assertDuplicatedMassesIndependent(w); err == nil {
		t.Fatal("expected missing original mass error")
	}

	w = &world{domainWorld: sim.NewWorld(), originalMassIDs: []int{1}, duplicated: edit.DuplicatedObjects{SpringIDs: []int{3}}}
	_ = w.domainWorld.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = w.domainWorld.AddMass(sim.Mass{ID: 2, Mass: 1})
	_ = w.domainWorld.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2})
	if err := assertDuplicatedIndependent(w, nil); err == nil {
		t.Fatal("expected duplicate spring endpoint error")
	}
}

func TestControlsHotkeysHelpersReportFailures(t *testing.T) {
	w := &world{}
	if err := createControlWorldState(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertControlWorldState(w, map[string]string{"expected_state": "written to XSP file"}); err == nil {
		t.Fatal("expected unsaved world state error")
	}
	if err := assertControlParameterResult(w, map[string]string{"parameter_result": "replaced by XSP file"}); err == nil {
		t.Fatal("expected unchanged parameter error")
	}
	if err := runNamedFileCommand(w, w.appGame.(*app.Game), "export"); err == nil {
		t.Fatal("expected unsupported file command")
	}
	if _, err := concreteGame(&world{}); err == nil {
		t.Fatal("expected missing concrete application")
	}
}

func TestApplicationWindowHelpersReportFailures(t *testing.T) {
	openErr := errors.New("open failed")
	if err := assertApplicationWindowOpened(&world{appErr: openErr}, nil); err != openErr {
		t.Fatal("expected application error")
	}
	if err := assertApplicationWorldEmpty(&world{}, nil); err == nil {
		t.Fatal("expected missing application")
	}

	worldWithSpring := appWorldWithMassAndSpring(false, true)
	if err := assertApplicationWorldEmpty(worldWithSpring, nil); err == nil {
		t.Fatal("expected invalid spring-only world error")
	}

	if err := resizeApplicationWindow(&world{}, map[string]string{"window_size": "small"}); err != nil {
		t.Fatal(err)
	}
	if err := resizeApplicationWindow(&world{}, map[string]string{"window_size": "large"}); err != nil {
		t.Fatal(err)
	}
	if err := resizeApplicationWindow(&world{}, map[string]string{"window_size": "medium"}); err == nil {
		t.Fatal("expected unsupported window size")
	}
}

func TestApplicationSteppingHelpersReportFailures(t *testing.T) {
	steppingGame := newSteppingGame()
	if len(steppingGame.World().Masses) != 1 || steppingGame.World().Masses[0].ID != 1 || steppingGame.World().Masses[0].Mass != 1 {
		t.Fatalf("stepping mass = %#v", steppingGame.World().Masses)
	}

	if err := assertApplicationStepping(&world{}, map[string]string{"stepping": "active"}); err == nil {
		t.Fatal("expected missing application")
	}
	if err := assertApplicationStepping(&world{appGame: steppingGame}, map[string]string{}); err == nil {
		t.Fatal("expected missing stepping")
	}
	if err := assertApplicationStepping(&world{appGame: steppingGame}, map[string]string{"stepping": "paused"}); err == nil {
		t.Fatal("expected unsupported stepping")
	}
	if expected, ok := expectedSteppingState("paused"); ok || expected {
		t.Fatalf("unsupported stepping state = %t, %t", expected, ok)
	}

	activeWorld := &world{appGame: steppingGame, appBeforeTime: steppingGame.World().Time - 1}
	if err := assertApplicationStepping(activeWorld, map[string]string{"stepping": "active"}); err != nil {
		t.Fatal(err)
	}
	stoppedWorld := &world{appGame: steppingGame, appBeforeTime: steppingGame.World().Time}
	if err := assertApplicationStepping(stoppedWorld, map[string]string{"stepping": "stopped"}); err != nil {
		t.Fatal(err)
	}
	if err := assertApplicationStepping(activeWorld, map[string]string{"stepping": "stopped"}); err == nil {
		t.Fatal("expected stepping mismatch")
	}
}

func TestApplicationActivityAndExitHelpersReportFailures(t *testing.T) {
	if err := assertApplicationActive(&world{}, "input handling", appGame.InputActive); err == nil {
		t.Fatal("expected missing application")
	}

	game := app.NewGame()
	if err := assertApplicationActive(&world{appGame: game}, "input handling", appGame.InputActive); err == nil {
		t.Fatal("expected inactive input")
	}

	if err := assertApplicationExitClean(&world{}, nil); err == nil {
		t.Fatal("expected missing application")
	}
	if err := assertApplicationExitClean(&world{appGame: game, appErr: errors.New("close failed")}, nil); err == nil {
		t.Fatal("expected close error")
	}
	if err := assertApplicationExitClean(&world{appGame: game}, nil); err == nil {
		t.Fatal("expected unclosed application")
	}
}

func TestScreenControlHelpersReportFailures(t *testing.T) {
	screen := app.NewGame().EditorScreen()
	w := &world{appGame: app.NewGame(), editorScreen: screen}

	if err := assertScreenRegionVisible(w, map[string]string{"region": "footer"}); err == nil {
		t.Fatal("expected missing region")
	}
	if err := assertScreenRegionPurpose(w, map[string]string{"purpose": "anything"}); err == nil {
		t.Fatal("expected missing region value")
	}
	if err := assertScreenRegionPurpose(w, map[string]string{"region": "canvas"}); err == nil {
		t.Fatal("expected missing purpose value")
	}
	if err := assertScreenRegionPurpose(w, map[string]string{"region": "canvas", "purpose": "wrong"}); err == nil {
		t.Fatal("expected purpose mismatch")
	}
	if err := assertScreenRegionPurpose(w, map[string]string{"region": "footer", "purpose": ""}); err == nil {
		t.Fatal("expected absent region mismatch")
	}

	if err := assertVisibleIndicator(w, map[string]string{"state": "running"}); err == nil {
		t.Fatal("expected missing indicator value")
	}
	if err := assertVisibleIndicator(w, map[string]string{"indicator": "simulation state"}); err == nil {
		t.Fatal("expected missing state value")
	}
	if err := assertVisibleIndicator(w, map[string]string{"indicator": "simulation state", "state": "running"}); err != nil {
		t.Fatal(err)
	}
	if err := assertVisibleIndicator(w, map[string]string{"indicator": "simulation state", "state": "wrong"}); err == nil {
		t.Fatal("expected indicator mismatch")
	}

	if err := assertCurrentScreen(&world{}, func(editorScreen) bool { return true }, "missing"); err == nil {
		t.Fatal("expected missing screen")
	}
	if err := assertCurrentScreen(w, func(editorScreen) bool { return false }, "mismatch"); err == nil {
		t.Fatal("expected screen mismatch")
	}
	if err := assertVisibleControl(w, map[string]string{}, "command", "command", editorScreen.HasCommandControl); err == nil {
		t.Fatal("expected missing control value")
	}
	if err := assertVisibleControl(w, map[string]string{"command": "export"}, "command", "command", editorScreen.HasCommandControl); err == nil {
		t.Fatal("expected invisible control")
	}
}

func TestScreenCommandAndStateHelpersReportFailures(t *testing.T) {
	if err := setSimulationState(&world{}, map[string]string{}); err == nil {
		t.Fatal("expected missing simulation state")
	}
	if err := setSimulationState(&world{}, map[string]string{"simulation_state": "waiting"}); err == nil {
		t.Fatal("expected unsupported simulation state")
	}
	if err := setSimulationState(&world{}, map[string]string{"simulation_state": "paused"}); err != nil {
		t.Fatal(err)
	}
	if paused, ok := simulationPausedState("paused"); !ok || !paused {
		t.Fatalf("paused state = %t, %t", paused, ok)
	}
	if paused, ok := simulationPausedState("running"); !ok || paused {
		t.Fatalf("running paused state = %t, %t", paused, ok)
	}

	game := app.NewGame()
	if err := assertCommandRan(&world{}, map[string]string{"command": "pause"}); err == nil {
		t.Fatal("expected missing application")
	}
	if err := assertCommandRan(&world{appGame: game}, map[string]string{}); err == nil {
		t.Fatal("expected missing command")
	}
	game.RunCommand("pause")
	if err := assertCommandRan(&world{appGame: game, appCommand: ""}, map[string]string{"command": "pause"}); err == nil {
		t.Fatal("expected queued command mismatch")
	}
	if err := assertCommandRan(&world{appGame: app.NewGame(), appCommand: "pause"}, map[string]string{"command": "pause"}); err == nil {
		t.Fatal("expected executed command mismatch")
	}
	if err := assertCommandRan(&world{appGame: game, appCommand: "pause"}, map[string]string{"command": "pause"}); err != nil {
		t.Fatal(err)
	}
	resetGame := app.NewGame()
	resetGame.RunCommand("reset")
	if err := assertCommandRan(&world{appGame: resetGame, appCommand: "reset"}, map[string]string{"command": "pause toggle"}); err == nil {
		t.Fatal("expected non-pause command mismatch")
	}
}

func appWorldWithMassAndSpring(includeMass, includeSpring bool) *world {
	game := app.NewGame()
	game.World().Reset()
	if includeMass || includeSpring {
		_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})
	}
	if includeSpring {
		game.World().Springs = append(game.World().Springs, sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 1, SpringConstant: 1})
		if !includeMass {
			game.World().Masses = nil
		}
	}
	return &world{appGame: game}
}

func runFeatureFile(t *testing.T, path string) {
	t.Helper()
	feature, err := gherkin.ReadFile(repoPath(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := RunFeature(feature); err != nil {
		t.Fatalf("RunFeature returned error: %v", err)
	}
}

func TestXSPLoadedStateChecksSuccessfulLoadState(t *testing.T) {
	w := &world{xspWorld: sim.NewWorld()}

	err := assertXSPLoadedState(w, map[string]string{"loaded_state": "current mass"})

	if err == nil {
		t.Fatal("expected loaded state mismatch")
	}
}

func TestForceCenterHelpersValidateInputsAndBranches(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "select force missing force", err: selectForce(&world{}, nil), want: "missing example value force"},
		{name: "exposes parameter missing force", err: assertForceExposesParameter(&world{}, map[string]string{"parameter_one": "Magnitude", "parameter_two": "Direction"}), want: "missing example value force"},
		{name: "exposes parameter missing second", err: assertForceExposesParameter(&world{}, map[string]string{"force": "gravity", "parameter_one": "Magnitude"}), want: "missing example value parameter_two"},
		{name: "gravity direction missing expected", err: assertGravityDirection(&world{}, nil), want: "missing example value expected_direction"},
		{name: "force center missing expected", err: assertForceCenter(&world{}, nil), want: "missing example value expected_center"},
		{name: "visual marker missing center", err: assertCenterMassVisuallyMarked(&world{}, nil), want: "missing example value center_mass"},
		{name: "reciprocal check missing center", err: assertNoReciprocalCenterForce(&world{}, nil), want: "missing example value center_mass"},
		{name: "controls active missing force", err: assertForceControlsActive(&world{}, nil), want: "missing example value force"},
	}
	for _, test := range tests {
		assertErrorContains(t, test.name, test.err, test.want)
	}

	if supportedForceName("missing force") {
		t.Fatal("expected missing force to be unsupported")
	}
	if !hasForceParameter("gravity", "Magnitude") {
		t.Fatal("expected gravity Magnitude parameter")
	}
	if hasForceParameter("gravity", "Exponent") {
		t.Fatal("expected gravity Exponent parameter to be absent")
	}
	if matchesExpectedDirection(sim.Vec2{X: 0.000001, Y: -1}, "down") {
		t.Fatal("expected tolerance boundary to be outside direction match")
	}
	if matchesExpectedDirection(sim.Vec2{X: 0, Y: -0.999999}, "down") {
		t.Fatal("expected Y tolerance boundary to be outside direction match")
	}
	if matchesExpectedDirection(sim.Vec2{X: 0, Y: 1}, "down") {
		t.Fatal("expected mismatched component to reject direction")
	}

	w := &world{}
	if err := createSelectedMasses(w, map[string]string{"selected_masses": "1"}); err != nil {
		t.Fatal(err)
	}
	assertSimulationMassPosition(t, w.domainWorld, 1, sim.Vec2{X: 10, Y: 20})
	assertSimulationMassValue(t, w.domainWorld, 1, 1)
	if len(w.originalMassIDs) != 1 || w.originalMassIDs[0] != 1 {
		t.Fatalf("selected mass ids = %#v", w.originalMassIDs)
	}
	w.domainWorld.Parameters.Set("center mass", "0")
	if err := assertForceCenter(w, map[string]string{"expected_center": "screen center"}); err != nil {
		t.Fatal(err)
	}

	w = &world{}
	if err := createForceCenterMass(w, map[string]string{"center_mass": "4"}); err != nil {
		t.Fatal(err)
	}
	assertSimulationMassPosition(t, w.domainWorld, 4, sim.Vec2{X: 50, Y: 50})
	assertSimulationMassPosition(t, w.domainWorld, 5, sim.Vec2{X: 0, Y: 50})
	assertSimulationMassValue(t, w.domainWorld, 4, 1)
	assertSimulationMassValue(t, w.domainWorld, 5, 1)
	if !w.domainWorld.IsCenterMass(4) {
		t.Fatalf("center mass = %d, want 4", w.domainWorld.CenterMassID())
	}
	if err := assertCenterMassVisuallyMarked(w, map[string]string{"center_mass": "5"}); err == nil {
		t.Fatal("expected center marker mismatch")
	}

	w.forceEvaluation = sim.ForceEvaluation{ByMassID: map[int]sim.MassForces{4: {}}}
	if err := assertNoReciprocalCenterForce(w, map[string]string{"center_mass": "4"}); err == nil {
		t.Fatal("expected missing force name")
	}
	if err := assertForceControlsActive(w, map[string]string{"force": "unsupported"}); err == nil {
		t.Fatal("expected unsupported force")
	}
}

func assertErrorContains(t *testing.T, name string, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error", name)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("%s: error = %q, want substring %q", name, err.Error(), want)
	}
}

func TestXSPHelpersRejectMissingAndMismatchedState(t *testing.T) {
	if err := assertXSPLoadResult(&world{}, nil); err == nil {
		t.Fatal("expected missing load result error")
	}
	if err := assertXSPLoadResult(&world{xspLoadErr: errors.New("load failed")}, map[string]string{"result": "pass"}); err == nil {
		t.Fatal("expected load pass mismatch")
	}

	loadedWorld := sim.NewWorld()
	loadedWorld.Parameters.EnableForce("gravity", map[string]string{"magnitude": "5", "direction": "90"})
	if err := assertForceLoaded(loadedWorld); err == nil {
		t.Fatal("expected force mismatch")
	}
	loadedWorld.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	if err := assertForceLoaded(loadedWorld); err != nil {
		t.Fatal(err)
	}

	_ = loadedWorld.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 9, Y: 20}, Mass: 1})
	if err := assertMassLoaded(loadedWorld); err == nil {
		t.Fatal("expected mass mismatch")
	}
	loadedWorld.Masses[0].Position = sim.Vec2{X: 10, Y: 20}
	if err := assertMassLoaded(loadedWorld); err != nil {
		t.Fatal(err)
	}

	_ = loadedWorld.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	_ = loadedWorld.AddSpring(sim.Spring{ID: 1, MassA: 2, MassB: 1, RestLength: 1, SpringConstant: 1})
	if err := assertSpringLoaded(loadedWorld); err == nil {
		t.Fatal("expected spring mismatch")
	}
	loadedWorld.Springs[0].MassA = 1
	loadedWorld.Springs[0].MassB = 2
	if err := assertSpringLoaded(loadedWorld); err != nil {
		t.Fatal(err)
	}

	worldWithBadSpringA := sim.NewWorld()
	_ = worldWithBadSpringA.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = worldWithBadSpringA.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 1}, Mass: 1})
	_ = worldWithBadSpringA.AddSpring(sim.Spring{ID: 1, MassA: 2, MassB: 2, RestLength: 1, SpringConstant: 1})
	if err := assertSpringLoaded(worldWithBadSpringA); err == nil {
		t.Fatal("expected spring mass A mismatch")
	}

	worldWithBadSpringB := sim.NewWorld()
	_ = worldWithBadSpringB.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = worldWithBadSpringB.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 1}, Mass: 1})
	_ = worldWithBadSpringB.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 1, RestLength: 1, SpringConstant: 1})
	if err := assertSpringLoaded(worldWithBadSpringB); err == nil {
		t.Fatal("expected spring mass B mismatch")
	}

	if err := assertXSPLoadErrorReason(&world{}, map[string]string{"reason": "duplicate id"}); err == nil {
		t.Fatal("expected missing load error")
	}

	fixedWorld := sim.NewWorld()
	_ = fixedWorld.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1, Fixed: false})
	if err := assertXSPMassFixedState(&world{xspWorld: fixedWorld}, map[string]string{"mass_id": "1", "fixed": "true"}); err == nil {
		t.Fatal("expected fixed state mismatch")
	}

	if err := assertSavedMassSign(&world{}, map[string]string{"mass_id": "1", "file_mass_sign": "negative"}); err == nil {
		t.Fatal("expected missing saved mass")
	}
	if err := assertFileMassSign("mass 1 10 20 -3", "negative"); err != nil {
		t.Fatal(err)
	}
	if err := assertFileMassSign("mass 1 10 20 3 0.8", "positive"); err != nil {
		t.Fatal(err)
	}
}

func TestForceEvaluationAndSimulationStepHelperBranches(t *testing.T) {
	w := &world{}
	if err := createMovableMassAffectedByForce(w, map[string]string{"force": "center of mass attraction"}); err != nil {
		t.Fatal(err)
	}
	if len(w.domainWorld.Masses) != 2 {
		t.Fatalf("masses = %#v", w.domainWorld.Masses)
	}
	assertSimulationMassPosition(t, w.domainWorld, 1, sim.Vec2{X: 0, Y: 0})
	assertSimulationMassPosition(t, w.domainWorld, 2, sim.Vec2{X: 50, Y: 50})
	assertSimulationMassValue(t, w.domainWorld, 1, 1)
	assertSimulationMassValue(t, w.domainWorld, 2, 1)

	w = &world{}
	if err := createMovableMassAffectedByForce(w, map[string]string{"force": "wall repulsion"}); err != nil {
		t.Fatal(err)
	}
	assertSimulationMassPosition(t, w.domainWorld, 1, sim.Vec2{X: 1, Y: 50})
	if velocity := w.domainWorld.Masses[0].Velocity; velocity != (sim.Vec2{X: 2, Y: 0}) {
		t.Fatalf("movable mass velocity = %#v", velocity)
	}
	assertSimulationMassPosition(t, &sim.Simulation{Masses: []sim.Mass{{ID: 1, Position: forceMassPosition("gravity")}}}, 1, sim.Vec2{X: -1, Y: 50})

	if err := createMassStartPosition(&world{}, map[string]string{"mass_id": "7", "start_position": "initial"}); err != nil {
		t.Fatal(err)
	}
	w = &world{}
	if err := createMassStartPosition(w, map[string]string{"mass_id": "7", "start_position": "12,13"}); err != nil {
		t.Fatal(err)
	}
	assertSimulationMassPosition(t, w.domainWorld, 7, sim.Vec2{X: 12, Y: 13})
	if updated := setMassStartPosition(w.domainWorld, 7, sim.Vec2{X: 14, Y: 15}); !updated {
		t.Fatal("expected existing mass update")
	}
	assertSimulationMassPosition(t, w.domainWorld, 7, sim.Vec2{X: 14, Y: 15})
	if updated := setMassStartPosition(w.domainWorld, 8, sim.Vec2{X: 1, Y: 1}); updated {
		t.Fatal("expected missing mass update to report false")
	}
	if _, err := durationValue(map[string]string{"duration": "forever"}, "duration"); err == nil {
		t.Fatal("expected unsupported duration")
	}
	if _, err := frameRateValue(map[string]string{"frame_rate": "120 fps"}); err == nil {
		t.Fatal("expected unsupported frame rate")
	}
}

func TestForceEvaluationPureHelpers(t *testing.T) {
	w := &world{}
	if err := createSpringForceWorld(w, map[string]string{"mass_a": "3", "mass_b": "4"}); err != nil {
		t.Fatal(err)
	}
	assertSimulationMassPosition(t, w.domainWorld, 3, sim.Vec2{X: 0, Y: 0})
	assertSimulationMassPosition(t, w.domainWorld, 4, sim.Vec2{X: 12, Y: 0})
	assertSimulationMassValue(t, w.domainWorld, 3, 1)
	assertSimulationMassValue(t, w.domainWorld, 4, 1)
	if spring := w.domainWorld.Springs[0]; spring.ID != 1 || spring.MassA != 3 || spring.MassB != 4 || spring.RestLength != 10 || spring.Stiffness != 1 || spring.SpringConstant != 1 {
		t.Fatalf("default spring = %#v", spring)
	}

	massA, massB, err := springForceMassIDs(map[string]string{"mass_a": "3", "mass_b": "4"})
	if err != nil || massA != 3 || massB != 4 {
		t.Fatalf("springForceMassIDs = %d, %d, %v", massA, massB, err)
	}
	if massA, massB, err := springForceMassIDs(map[string]string{"mass_a": "bad", "mass_b": "4"}); err == nil || massA != 0 || massB != 0 {
		t.Fatalf("expected bad mass_a zeros, got %d, %d, %v", massA, massB, err)
	}
	if massA, massB, err := springForceMassIDs(map[string]string{"mass_a": "3", "mass_b": "bad"}); err == nil || massA != 0 || massB != 0 {
		t.Fatalf("expected bad mass_b zeros, got %d, %d, %v", massA, massB, err)
	}

	massID, velocity, err := massVelocityFromExample(map[string]string{"mass": "9", "velocity": "moving"}, "mass", "velocity")
	if err != nil || massID != 9 || velocity != (sim.Vec2{X: 1, Y: 5}) {
		t.Fatalf("massVelocityFromExample = %d, %#v, %v", massID, velocity, err)
	}
	for _, example := range []map[string]string{
		{"mass": "bad", "velocity": "moving"},
		{"mass": "9"},
		{"mass": "9", "velocity": "fast"},
	} {
		if massID, velocity, err := massVelocityFromExample(example, "mass", "velocity"); err == nil || massID != 0 || velocity != (sim.Vec2{}) {
			t.Fatalf("expected bad velocity example zeros, got %d, %#v, %v", massID, velocity, err)
		}
	}

	if velocity, err := namedVelocity("still"); err != nil || velocity != (sim.Vec2{}) {
		t.Fatalf("still velocity = %#v, %v", velocity, err)
	}
	if !vecClose(sim.Vec2{}, sim.Vec2{X: 0.000001, Y: 0.000001}) {
		t.Fatal("expected vecClose boundary match")
	}
	if !vecClose(sim.Vec2{X: 1, Y: 2}, sim.Vec2{X: 0.9999995, Y: 2}) {
		t.Fatal("expected vecClose lower X tolerance match")
	}
	if !vecClose(sim.Vec2{X: 1, Y: 2}, sim.Vec2{X: 1, Y: 1.9999995}) {
		t.Fatal("expected vecClose lower Y tolerance match")
	}
	if vecClose(sim.Vec2{X: 1, Y: 2}, sim.Vec2{X: -0.9999995, Y: 2}) {
		t.Fatal("expected mirrored X values to differ")
	}
	if vecClose(sim.Vec2{X: 1, Y: 2}, sim.Vec2{X: 1, Y: -1.9999995}) {
		t.Fatal("expected mirrored Y values to differ")
	}
	if vecClose(sim.Vec2{X: 3, Y: 2}, sim.Vec2{X: -3, Y: 2}) {
		t.Fatal("expected opposite X values to differ")
	}
	if simDot(sim.Vec2{X: 2, Y: 3}, sim.Vec2{X: 5, Y: 7}) != 31 {
		t.Fatal("unexpected dot product")
	}
}

func assertSimulationMassPosition(t *testing.T, world *sim.Simulation, id int, position sim.Vec2) {
	t.Helper()
	mass, ok := world.MassByID(id)
	if !ok {
		t.Fatalf("mass %d not found", id)
	}
	if mass.Position != position {
		t.Fatalf("mass %d position = %#v, want %#v", id, mass.Position, position)
	}
}

func assertSimulationMassValue(t *testing.T, world *sim.Simulation, id int, value float64) {
	t.Helper()
	mass, ok := world.MassByID(id)
	if !ok {
		t.Fatalf("mass %d not found", id)
	}
	if mass.Mass != value {
		t.Fatalf("mass %d value = %v, want %v", id, mass.Mass, value)
	}
}

func TestSimulationStepHelpersReportFailures(t *testing.T) {
	if err := createMassStartPosition(&world{}, map[string]string{"mass_id": "1", "start_position": "custom"}); err == nil {
		t.Fatal("expected unsupported position marker")
	}
	if err := assertResultDeterministic(nil, map[string]string{"initial_state": "unknown", "duration": "1 second"}); err == nil {
		t.Fatal("expected unsupported initial state")
	}
	w := &world{domainWorld: sim.NewWorld()}
	if err := advanceByDurationAtFrameRate(w, map[string]string{"duration": "1 second", "frame_rate": "bad"}); err == nil {
		t.Fatal("expected unsupported frame rate")
	}
	w.resultingWorld = sim.NewWorld()
	_ = w.resultingWorld.AddMass(sim.Mass{ID: 1, Velocity: sim.Vec2{X: 1}})
	if err := assertMassVelocityRemains(w, map[string]string{"mass_id": "1", "start_velocity": "zero"}); err == nil {
		t.Fatal("expected changed velocity error")
	}
}

func TestSameWorldStateDetectsDifferences(t *testing.T) {
	first := sim.NewWorld()
	second := sim.NewWorld()
	_ = first.AddMass(sim.Mass{ID: 1})
	if sameWorldState(first, second) {
		t.Fatal("expected length mismatch")
	}
	_ = second.AddMass(sim.Mass{ID: 1})
	second.Time = 1
	if sameWorldState(first, second) {
		t.Fatal("expected time mismatch")
	}
	second.Time = 0
	second.Masses[0].Position = sim.Vec2{X: 1}
	if sameWorldState(first, second) {
		t.Fatal("expected position mismatch")
	}
	second.Masses[0].Position = sim.Vec2{}
	second.Masses[0].Velocity = sim.Vec2{X: 1}
	if sameWorldState(first, second) {
		t.Fatal("expected velocity mismatch")
	}
	if sameFloat(1, 1.000002) {
		t.Fatal("expected float outside tolerance")
	}
	if sameFloat(1, 0.999999) {
		t.Fatal("expected float boundary to be outside tolerance")
	}
	if !sameFloat(1, 1.000001) {
		t.Fatal("expected positive float boundary to be inside tolerance")
	}
	if !sameFloat(0, 0.000001) {
		t.Fatal("expected literal float boundary to be inside tolerance")
	}
	if _, _, _, err := deterministicWorlds(map[string]string{"duration": "1 second"}); err == nil || !strings.Contains(err.Error(), "initial_state") {
		t.Fatalf("expected missing initial state, got %v", err)
	}
	if firstWorld, secondWorld, badDuration, err := deterministicWorlds(map[string]string{"initial_state": "unknown", "duration": "1 second"}); err == nil || firstWorld != nil || secondWorld != nil || badDuration != 0 {
		t.Fatal("expected unsupported deterministic world state")
	}
	if duration, err := durationValue(map[string]string{"duration": "10 steps"}, "duration"); err != nil || duration != sim.DefaultParameters().StepDuration()*10 {
		t.Fatalf("10 steps duration = %v, %v", duration, err)
	}
	if duration, err := durationValue(map[string]string{"duration": "1 second"}, "duration"); err != nil || duration != 1 {
		t.Fatalf("1 second duration = %v, %v", duration, err)
	}
	if duration, err := durationValue(map[string]string{}, "duration"); err == nil || duration != 0 || !strings.Contains(err.Error(), "duration") {
		t.Fatalf("expected missing duration, got %v, %v", duration, err)
	}
	if rate, err := frameRateValue(map[string]string{"frame_rate": "30 fps"}); err != nil || rate != 30 {
		t.Fatalf("30 fps = %v, %v", rate, err)
	}
	if rate, err := frameRateValue(map[string]string{"frame_rate": "bogus"}); err == nil || rate != 0 {
		t.Fatalf("expected unsupported frame rate, got %v, %v", rate, err)
	}
	if err := requireMarker(map[string]string{}, "state", "ready"); err == nil || !strings.Contains(err.Error(), "state") {
		t.Fatalf("expected missing marker, got %v", err)
	}
	if err := requireMarker(map[string]string{"state": "waiting"}, "state", "ready"); err == nil {
		t.Fatal("expected marker mismatch")
	}
}

func TestStartupEditorPureHelpers(t *testing.T) {
	simWorld := sim.NewWorld()
	_ = simWorld.AddMass(sim.Mass{ID: 1, Fixed: true})
	_ = simWorld.AddMass(sim.Mass{ID: 2})
	_ = simWorld.AddMass(sim.Mass{ID: 3})
	_ = simWorld.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	if count := startupObjectCount(simWorld, "fixed mass"); count != 1 {
		t.Fatalf("fixed mass count = %d", count)
	}
	if count := startupObjectCount(simWorld, "movable mass"); count != 2 {
		t.Fatalf("movable mass count = %d", count)
	}
	if count := startupObjectCount(simWorld, "spring"); count != 1 {
		t.Fatalf("spring count = %d", count)
	}
	if count := startupObjectCount(simWorld, "unknown"); count != 0 {
		t.Fatalf("unknown count = %d", count)
	}
	if sameStringSlices([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("expected length mismatch")
	}
	if sameStringSlices([]string{"a"}, []string{"b"}) {
		t.Fatal("expected value mismatch")
	}
	if !sameStringSlices([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatal("expected matching slices")
	}
	if len(startupRegions()) != 5 {
		t.Fatalf("startup regions = %#v", startupRegions())
	}
	if _, err := concreteStartupGame(&world{}); err == nil {
		t.Fatal("expected missing startup game")
	}
	if err := assertStartupObjectCount(&world{}, map[string]string{"object_count": "none", "object_type": "spring"}); err == nil || !strings.Contains(err.Error(), "unsupported object count") {
		t.Fatalf("expected unsupported object count, got %v", err)
	}
}

func TestSystemParameterHandlersReportFailures(t *testing.T) {
	w := &world{domainWorld: sim.NewWorld()}
	if err := assertParameterDefault(w, map[string]string{"parameter": "viscosity", "value": "unset"}); err == nil {
		t.Fatal("expected unsupported default marker")
	}
	if err := assertParameterDefault(w, map[string]string{"parameter": "missing", "value": "set"}); err == nil {
		t.Fatal("expected missing default parameter")
	}
	if err := assertForceEnabledState(w, map[string]string{"force": "missing", "enabled": "set"}); err == nil {
		t.Fatal("expected missing force")
	}
	if err := assertWallEnabledState(w, map[string]string{"wall": "missing", "enabled": "set"}); err == nil {
		t.Fatal("expected missing wall")
	}
	if err := performWorldOperation(w, map[string]string{"operation": "delete file"}); err == nil {
		t.Fatal("expected unsupported operation")
	}
	if _, err := expectedParameterValue("viscosity", "unknown source", nil); err == nil {
		t.Fatal("expected unsupported parameter source")
	}
}

func TestSystemParameterHandlerHelpers(t *testing.T) {
	w := &world{}
	if err := changeWorldParameter(w, map[string]string{"parameter": "viscosity", "changed_value": "custom"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterDefault(w, map[string]string{"parameter": "viscosity", "value": "set"}); err != nil {
		t.Fatal(err)
	}
	if err := assertForceEnabledState(w, map[string]string{"force": "gravity", "enabled": "set"}); err != nil {
		t.Fatal(err)
	}
	if err := assertForceEditableParameters(w, map[string]string{"force": "gravity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertWallEnabledState(w, map[string]string{"wall": "top", "enabled": "set"}); err != nil {
		t.Fatal(err)
	}
	if err := performWorldOperation(w, map[string]string{"operation": "insert file", "parameter": "viscosity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterSource(w, map[string]string{
		"parameter": "viscosity", "changed_value": "custom", "expected_value_source": "existing world value",
	}); err != nil {
		t.Fatal(err)
	}
	if err := performWorldOperation(w, map[string]string{"operation": "load file", "parameter": "viscosity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterSource(w, map[string]string{"parameter": "viscosity", "expected_value_source": "value from loaded file"}); err != nil {
		t.Fatal(err)
	}
	if err := performWorldOperation(w, map[string]string{"operation": "reset", "parameter": "viscosity"}); err != nil {
		t.Fatal(err)
	}
	if err := assertParameterSource(w, map[string]string{"parameter": "viscosity", "expected_value_source": "default value"}); err != nil {
		t.Fatal(err)
	}
}

func TestForceEvaluationHandlerHelpers(t *testing.T) {
	springExample := map[string]string{
		"mass_a": "1", "mass_b": "2", "rest_length": "10", "spring_constant": "12",
		"velocity_a": "moving", "velocity_b": "still", "damping_constant": "0.5",
	}
	w := &world{}
	for _, fn := range []stepHandler{
		createSpringForceWorld,
		setOnlySpringRestLength,
		setOnlySpringConstant,
		setMassAVelocity,
		setMassBVelocity,
		setOnlySpringDamping,
		evaluateForces,
		assertSpringForcesEqualOpposite,
		assertSpringDampingDirection,
	} {
		if err := fn(w, springExample); err != nil {
			t.Fatalf("force handler returned error: %v", err)
		}
	}
	assertSimulationMassPosition(t, w.domainWorld, 1, sim.Vec2{X: 0, Y: 0})
	assertSimulationMassPosition(t, w.domainWorld, 2, sim.Vec2{X: 12, Y: 0})
	assertSimulationMassValue(t, w.domainWorld, 1, 1)
	assertSimulationMassValue(t, w.domainWorld, 2, 1)
	if spring := w.domainWorld.Springs[0]; spring.ID != 1 || spring.MassA != 1 || spring.MassB != 2 || spring.RestLength != 10 || spring.SpringConstant != 12 || spring.Damping != 0.5 {
		t.Fatalf("spring = %#v", spring)
	}
	if velocity := w.domainWorld.Masses[0].Velocity; velocity != (sim.Vec2{X: 1, Y: 5}) {
		t.Fatalf("mass A velocity = %#v", velocity)
	}
	if velocity := w.domainWorld.Masses[1].Velocity; velocity != (sim.Vec2{}) {
		t.Fatalf("mass B velocity = %#v", velocity)
	}

	for _, force := range []string{"gravity", "viscosity", "wall repulsion", "center attraction", "center of mass attraction"} {
		w = &world{}
		example := map[string]string{"force": force}
		if err := enableEnvironmentalForce(w, example); err != nil {
			t.Fatal(err)
		}
		if err := createMovableMassAffectedByForce(w, example); err != nil {
			t.Fatal(err)
		}
		if err := evaluateForces(w, nil); err != nil {
			t.Fatal(err)
		}
		if err := assertMassReceivesForce(w, example); err != nil {
			t.Fatal(err)
		}
	}

	fixedExample := map[string]string{"mass_id": "1", "fixed": "true", "force": "gravity", "acceleration": "zero"}
	w = &world{}
	if err := createMassFixedState(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	assertSimulationMassPosition(t, w.domainWorld, 1, sim.Vec2{X: 10, Y: 10})
	assertSimulationMassValue(t, w.domainWorld, 1, 1)
	if !w.domainWorld.Masses[0].Fixed {
		t.Fatal("expected fixed mass")
	}
	if err := affectMassByForce(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := evaluateForces(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := assertMassAcceleration(w, fixedExample); err != nil {
		t.Fatal(err)
	}

	for _, wall := range []string{"top", "left", "right", "bottom"} {
		w = &world{}
		example := map[string]string{"wall": wall, "mass_id": "1"}
		if err := enableWall(w, example); err != nil {
			t.Fatal(err)
		}
		if err := createMassNearInsideWall(w, example); err != nil {
			t.Fatal(err)
		}
		assertSimulationMassPosition(t, w.domainWorld, 1, insideWallPosition(wall))
		assertSimulationMassValue(t, w.domainWorld, 1, 1)
		if err := evaluateForces(w, nil); err != nil {
			t.Fatal(err)
		}
		if err := assertWallForceTowardInside(w, example); err != nil {
			t.Fatal(err)
		}
	}
}

func TestForceEvaluationHandlersReportFailures(t *testing.T) {
	if err := updateFirstSpring(&world{domainWorld: sim.NewWorld()}, nil); err == nil {
		t.Fatal("expected missing spring error")
	}
	if _, err := namedVelocity("fast"); err == nil {
		t.Fatal("expected unsupported velocity")
	}
	w := &world{domainWorld: sim.NewWorld()}
	if err := setMassNamedVelocity(w, map[string]string{"mass_a": "9", "velocity_a": "moving"}, "mass_a", "velocity_a"); err == nil {
		t.Fatal("expected missing mass error")
	}
	_ = w.domainWorld.AddMass(sim.Mass{ID: 1})
	_ = w.domainWorld.AddMass(sim.Mass{ID: 2})
	if err := setMassVelocityByID(w.domainWorld, 2, sim.Vec2{X: 6, Y: 7}); err != nil {
		t.Fatal(err)
	}
	if w.domainWorld.Masses[0].Velocity != (sim.Vec2{}) || w.domainWorld.Masses[1].Velocity != (sim.Vec2{X: 6, Y: 7}) {
		t.Fatalf("velocities = %#v", w.domainWorld.Masses)
	}
	if err := setSpringFloat(&world{domainWorld: sim.NewWorld()}, map[string]string{"rest_length": "bad"}, "rest_length", setSpringRestLength); err == nil {
		t.Fatal("expected missing spring before float parse")
	}
	springWorld := sim.NewWorld()
	_ = springWorld.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = springWorld.AddMass(sim.Mass{ID: 2, Mass: 1})
	_ = springWorld.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	if err := setSpringFloat(&world{domainWorld: springWorld}, map[string]string{"rest_length": "bad"}, "rest_length", setSpringRestLength); err == nil {
		t.Fatal("expected bad spring float")
	}
	if err := assertSpringForcesEqualOpposite(&world{}, map[string]string{"mass_a": "bad", "mass_b": "2"}); err == nil {
		t.Fatal("expected bad mass_a assertion")
	}
	if err := assertSpringForcesEqualOpposite(&world{}, map[string]string{"mass_a": "1", "mass_b": "bad"}); err == nil {
		t.Fatal("expected bad mass_b assertion")
	}
	for _, evaluation := range []sim.ForceEvaluation{
		{ByMassID: map[int]sim.MassForces{1: {Force: sim.Vec2{X: 0, Y: 0}}}},
		{ByMassID: map[int]sim.MassForces{1: {Force: sim.Vec2{X: 1, Y: 1}}}},
	} {
		if err := assertSpringDampingDirection(&world{forceEvaluation: evaluation}, nil); err == nil {
			t.Fatal("expected bad damping direction")
		}
	}
	if err := affectMassByForce(&world{}, map[string]string{"force": "wind"}); err == nil {
		t.Fatal("expected unsupported force")
	}
	w = &world{forceEvaluation: sim.ForceEvaluation{ByMassID: map[int]sim.MassForces{1: {Force: sim.Vec2{}}}}}
	if err := assertMassReceivesForce(w, map[string]string{"force": "gravity"}); err == nil {
		t.Fatal("expected missing force assertion")
	}
	if err := assertMassAcceleration(&world{}, map[string]string{"mass_id": "1", "acceleration": "moving"}); err == nil {
		t.Fatal("expected unsupported acceleration expectation")
	}
	if err := assertWallForceTowardInside(&world{}, map[string]string{"mass_id": "bad", "wall": "top"}); err == nil {
		t.Fatal("expected bad wall mass id")
	}
	if err := assertWallForceTowardInside(&world{}, map[string]string{"mass_id": "1"}); err == nil {
		t.Fatal("expected missing wall")
	}
	w = &world{forceEvaluation: sim.ForceEvaluation{ByMassID: map[int]sim.MassForces{1: {Force: sim.Vec2{X: 1, Y: 0}}}}}
	if err := assertWallForceTowardInside(w, map[string]string{"mass_id": "1", "wall": "top"}); err == nil {
		t.Fatal("expected sideways wall force")
	}
	w = &world{forceEvaluation: sim.ForceEvaluation{ByMassID: map[int]sim.MassForces{1: {Force: sim.Vec2{Y: -0.5}}}}}
	if err := assertWallForceTowardInside(w, map[string]string{"mass_id": "1", "wall": "top"}); err != nil {
		t.Fatal(err)
	}
}

func TestSimulationStepHandlerHelpers(t *testing.T) {
	w := &world{}
	moveExample := map[string]string{"start_position": "initial", "start_velocity": "zero", "duration": "1 step"}
	for _, fn := range []stepHandler{
		createMovableMassAtStart,
		enableGravity,
		advanceByDuration,
		assertMassPositionDiffers,
		assertMassVelocityDiffers,
	} {
		if err := fn(w, moveExample); err != nil {
			t.Fatalf("simulation step handler returned error: %v", err)
		}
	}

	fixedExample := map[string]string{"mass_id": "2", "start_position": "initial", "start_velocity": "zero", "fixed": "true", "force": "gravity", "duration": "10 steps"}
	w = &world{}
	if err := createMassFixedState(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := createMassStartPosition(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := enableGravity(w, nil); err != nil {
		t.Fatal(err)
	}
	if err := advanceByDuration(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := assertMassPositionRemains(w, fixedExample); err != nil {
		t.Fatal(err)
	}
	if err := assertMassVelocityRemains(w, fixedExample); err != nil {
		t.Fatal(err)
	}

	for _, state := range []string{"simple spring", "gravity only"} {
		example := map[string]string{"initial_state": state, "duration": "1 second", "frame_rate": "30 fps"}
		w = &world{}
		if err := createWorldInState(w, example); err != nil {
			t.Fatal(err)
		}
		if err := assertResultDeterministic(w, example); err != nil {
			t.Fatal(err)
		}
		if err := advanceByDurationAtFrameRate(w, example); err != nil {
			t.Fatal(err)
		}
		if err := assertSimulationTime(w, example); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSimulationStepHandlersReportFailures(t *testing.T) {
	if err := requireMarker(map[string]string{"value": "wrong"}, "value", "expected"); err == nil {
		t.Fatal("expected marker mismatch")
	}
	if _, err := durationValue(map[string]string{"duration": "forever"}, "duration"); err == nil {
		t.Fatal("expected unsupported duration")
	}
	if _, err := frameRateValue(map[string]string{"frame_rate": "100 fps"}); err == nil {
		t.Fatal("expected unsupported frame rate")
	}
	if _, err := worldForState("unknown"); err == nil {
		t.Fatal("expected unsupported state")
	}
	if sameWorldState(sim.NewWorld(), &sim.Simulation{Time: 1}) {
		t.Fatal("expected world states to differ")
	}
	if err := assertMassPositionDiffers(&world{resultingWorld: sim.NewWorld()}, map[string]string{"start_position": "initial"}); err == nil {
		t.Fatal("expected missing mass error")
	}
	if err := assertMassPositionRemains(&world{resultingWorld: sim.NewWorld()}, map[string]string{"mass_id": "1"}); err == nil {
		t.Fatal("expected missing mass error")
	}
	if err := assertSimulationTime(&world{resultingWorld: &sim.Simulation{Time: 2}}, map[string]string{"duration": "1 second"}); err == nil {
		t.Fatal("expected simulation time mismatch")
	}
}

func TestPackageDirDoesNotImportDetectsGraphicsLibrary(t *testing.T) {
	dir := t.TempDir()
	writeSource(t, filepath.Join(dir, "domain.go"), "package domain\n")
	if err := packageDirDoesNotImport(dir, "domain", "Ebitengine"); err != nil {
		t.Fatalf("packageDirDoesNotImport returned error: %v", err)
	}

	writeSource(t, filepath.Join(dir, "ui.go"), "package domain\nimport \"github.com/hajimehoshi/ebiten/v2\"\n")
	if err := packageDirDoesNotImport(dir, "domain", "Ebitengine"); err == nil {
		t.Fatal("expected graphics import error")
	}
}

func TestArtifactHelpers(t *testing.T) {
	if artifact, location, err := artifactExample(map[string]string{
		"artifact":           "test source",
		"generated_location": "acceptance/generated",
	}); err != nil || artifact != "test source" || location != "acceptance/generated" {
		t.Fatalf("artifactExample = %q, %q, %v", artifact, location, err)
	}
	if _, _, err := artifactExample(map[string]string{"artifact": "test source"}); err == nil {
		t.Fatal("expected missing generated location error")
	}
	if err := generatedArtifactExists("unsupported", "acceptance/generated"); err == nil {
		t.Fatal("expected unsupported artifact error")
	}
	path, err := generatedArtifactPath("test source", "acceptance/generated")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, filepath.Join("acceptance", "generated", "pipeline_artifacts_acceptance_test.go")) {
		t.Fatalf("generated artifact path = %s", path)
	}
}

func TestHandwrittenTestsOutside(t *testing.T) {
	if err := handwrittenTestsOutside("acceptance/generated"); err != nil {
		t.Fatalf("handwrittenTestsOutside returned error: %v", err)
	}
	if err := handwrittenTestsOutside("internal"); err == nil {
		t.Fatal("expected internal tests to violate generated location")
	}
}

func TestAssertHandwrittenTestsOutside(t *testing.T) {
	example := map[string]string{
		"test_type":          "unit",
		"generated_location": "acceptance/generated",
	}
	if err := assertHandwrittenTestsOutside(nil, example); err != nil {
		t.Fatalf("assertHandwrittenTestsOutside returned error: %v", err)
	}
	example["test_type"] = "integration"
	if err := assertHandwrittenTestsOutside(nil, example); err == nil {
		t.Fatal("expected unsupported test type error")
	}
}

func TestHandwrittenViolationHelpers(t *testing.T) {
	dir := t.TempDir()
	testPath := filepath.Join(dir, "example_test.go")
	writeSource(t, testPath, "package example\n")

	violations, err := handwrittenTestViolations(dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 || violations[0] != testPath {
		t.Fatalf("violations = %#v", violations)
	}
	if !isHandwrittenTestUnder(testPath, fakeDirEntry{name: "example_test.go"}, dir) {
		t.Fatal("expected test file under generated location")
	}
	if err := reportHandwrittenViolations(violations); err == nil {
		t.Fatal("expected violation report")
	}
}

func TestRunCommandInDirReportsFailures(t *testing.T) {
	if err := runCommandInDir(t.TempDir(), "go", "version"); err != nil {
		t.Fatalf("runCommandInDir returned error: %v", err)
	}
	if err := runCommandInDir(t.TempDir(), "go", "not-a-command"); err == nil {
		t.Fatal("expected command failure")
	}
}

func writeSource(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

type fakeDirEntry struct {
	name string
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return false }
func (f fakeDirEntry) Type() os.FileMode          { return 0 }
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }
