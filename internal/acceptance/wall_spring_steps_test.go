package acceptance

import (
	"testing"

	"springs/internal/sim"
)

func TestWallSpringBarrierForceStateSteps(t *testing.T) {
	for _, example := range []map[string]string{
		{
			"spring_id":           "1",
			"mass_a":              "1",
			"mass_b":              "2",
			"wall":                "false",
			"kspring":             "10",
			"kdamp":               "0.5",
			"rest_len":            "20",
			"spring_force_state":  "enabled",
			"damping_force_state": "enabled",
		},
		{
			"spring_id":           "1",
			"mass_a":              "1",
			"mass_b":              "2",
			"wall":                "true",
			"kspring":             "10",
			"kdamp":               "0.5",
			"rest_len":            "20",
			"spring_force_state":  "disabled",
			"damping_force_state": "disabled",
		},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, addBarrierSpring)
		mustWallSpringStep(t, w, example, setBarrierSpringWall)
		mustWallSpringStep(t, w, example, setBarrierSpringParameters)
		mustWallSpringStep(t, w, example, evaluateBarrierSpringForces)
		mustWallSpringStep(t, w, example, assertBarrierSpringForceState)
		mustWallSpringStep(t, w, example, assertBarrierSpringDampingState)
	}
}

func TestWallSpringBarrierWallSetterCreatesMissingSpring(t *testing.T) {
	example := map[string]string{"spring_id": "1", "wall": "true", "new_wall": "true"}
	w := &world{}
	mustWallSpringStep(t, w, example, setBarrierSpringWall)
	mustWallSpringStep(t, w, example, assertSpringWallValue)
}

func TestWallSpringBarrierLengthConstraintSteps(t *testing.T) {
	for _, example := range []map[string]string{
		{
			"spring_id":            "1",
			"initial_length":       "120",
			"rest_len":             "100",
			"endpoint_a":           "1",
			"endpoint_b":           "2",
			"fixed_a":              "false",
			"fixed_b":              "false",
			"expected_length":      "100",
			"correction_direction": "along segment",
		},
		{
			"spring_id":            "1",
			"initial_length":       "80",
			"rest_len":             "100",
			"endpoint_a":           "1",
			"endpoint_b":           "2",
			"fixed_a":              "false",
			"fixed_b":              "false",
			"expected_length":      "100",
			"correction_direction": "along segment",
		},
		{
			"spring_id":            "1",
			"initial_length":       "120",
			"rest_len":             "100",
			"endpoint_a":           "1",
			"endpoint_b":           "2",
			"fixed_a":              "true",
			"fixed_b":              "false",
			"expected_length":      "100",
			"correction_direction": "along segment",
		},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, createWallSpringLengthConstraint)
		mustWallSpringStep(t, w, example, setWallSpringEndpointFixed)
		mustWallSpringStep(t, w, example, setWallSpringEndpointBFixed)
		mustWallSpringStep(t, w, example, advanceWallSpringLengthConstraint)
		mustWallSpringStep(t, w, example, assertWallSpringEndpointDistance)
		mustWallSpringStep(t, w, example, assertWallSpringEndpointCorrection)
	}
}

func TestWallSpringBarrierCollisionSteps(t *testing.T) {
	assertWallSpringBarrierMovingMassSteps(t, map[string]string{
		"mass_x":  "-5",
		"mass_vx": "10",
	}, createBarrierMovingMass, advanceThroughWallSpringCollision)
}

func TestWallSpringBarrierFastCollisionSteps(t *testing.T) {
	assertWallSpringBarrierMovingMassSteps(t, map[string]string{
		"mass_x":   "-50",
		"mass_vx":  "1000",
		"duration": "1 step",
	}, createFastBarrierMovingMass, advanceThroughWallSpringCollisionByDuration)
}

func assertWallSpringBarrierMovingMassSteps(t *testing.T, overrides map[string]string, createMass, advance stepHandler) {
	t.Helper()
	example := mergeWallSpringExample(map[string]string{
		"spring_id": "1",
		"wall_x1":   "0",
		"wall_y1":   "0",
		"wall_x2":   "0",
		"wall_y2":   "100",
		"mass_id":   "3",
		"mass_x":    "-5",
		"mass_y":    "50",
		"mass_vx":   "10",
		"mass_vy":   "0",
	}, overrides)
	w := &world{}
	mustWallSpringStep(t, w, example, createWallSpringByCoordinates)
	mustWallSpringStep(t, w, example, createMass)
	mustWallSpringStep(t, w, example, advance)
	mustWallSpringStep(t, w, example, assertMassOnStartingWallSpringSide)
	mustWallSpringStep(t, w, example, assertMassVelocityResolvedAwayFromWallSpring)
}

func mergeWallSpringExample(base map[string]string, overrides map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}

func TestWallSpringBarrierMovingWallCollisionSteps(t *testing.T) {
	example := map[string]string{
		"spring_id": "1",
		"wall_x1":   "-5",
		"wall_y1":   "0",
		"wall_x2":   "-5",
		"wall_y2":   "100",
		"wall_vx":   "10",
		"wall_vy":   "0",
		"mass_id":   "3",
		"mass_x":    "0",
		"mass_y":    "50",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createMovingWallSpringByCoordinates)
	mustWallSpringStep(t, w, example, createBarrierStationaryMass)
	mustWallSpringStep(t, w, example, advanceThroughWallSpringCollision)
	mustWallSpringStep(t, w, example, assertMassOnStartingWallSpringSide)
	mustWallSpringStep(t, w, example, assertMovingWallSpringVelocityResolvedAwayFromMass)
}

func TestWallSpringBarrierLengthConstraintEndpointCollisionSteps(t *testing.T) {
	example := map[string]string{
		"barrier_spring": "1",
		"barrier_x1":     "0",
		"barrier_y1":     "0",
		"barrier_x2":     "0",
		"barrier_y2":     "100",
		"moving_spring":  "2",
		"endpoint_a":     "3",
		"endpoint_a_x":   "-5",
		"endpoint_a_y":   "40",
		"endpoint_b":     "4",
		"endpoint_b_x":   "-80",
		"endpoint_b_y":   "40",
		"rest_len":       "150",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createBarrierWallSpringByCoordinates)
	mustWallSpringStep(t, w, example, createConstrainedWallSpringEndpointA)
	mustWallSpringStep(t, w, example, createConstrainedWallSpringEndpointB)
	mustWallSpringStep(t, w, example, createConstrainedWallSpring)
	mustWallSpringStep(t, w, example, advanceWallSpringLengthConstraintsAndCollisions)
	mustWallSpringStep(t, w, example, assertWallSpringEndpointAOnStartingBarrierSide)
	mustWallSpringStep(t, w, example, assertWallSpringEndpointBOnStartingBarrierSide)
}

func TestWallSpringBarrierEndpointImpulseSteps(t *testing.T) {
	for _, example := range []map[string]string{
		{
			"spring_id":        "1",
			"endpoint_a":       "1",
			"endpoint_b":       "2",
			"fixed_a":          "false",
			"fixed_b":          "false",
			"mass_id":          "3",
			"contact_fraction": "0.50",
			"impulse_share_a":  "0.50",
			"impulse_share_b":  "0.50",
		},
		{
			"spring_id":        "1",
			"endpoint_a":       "1",
			"endpoint_b":       "2",
			"fixed_a":          "false",
			"fixed_b":          "false",
			"mass_id":          "3",
			"contact_fraction": "0.25",
			"impulse_share_a":  "0.75",
			"impulse_share_b":  "0.25",
		},
		{
			"spring_id":        "1",
			"endpoint_a":       "1",
			"endpoint_b":       "2",
			"fixed_a":          "true",
			"fixed_b":          "false",
			"mass_id":          "3",
			"contact_fraction": "0.25",
			"impulse_share_a":  "absorbed",
			"impulse_share_b":  "0.25",
		},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, createWallSpringByEndpointIDs)
		mustWallSpringStep(t, w, example, setWallSpringEndpointFixed)
		mustWallSpringStep(t, w, example, setWallSpringEndpointBFixed)
		mustWallSpringStep(t, w, example, createMassCollidingWithWallSpring)
		mustWallSpringStep(t, w, example, resolveWallSpringCollision)
		mustWallSpringStep(t, w, example, assertWallSpringEndpointImpulseShare)
		mustWallSpringStep(t, w, example, assertWallSpringEndpointBImpulseShare)
	}
}

func TestFloatingWallMomentumSteps(t *testing.T) {
	runWallSpringStepExamples(t,
		[]map[string]string{{
			"endpoint_a_mass": "2",
			"endpoint_b_mass": "5",
			"mass_id":         "3",
			"moving_mass":     "1",
			"mass_x":          "-5",
			"mass_y":          "50",
			"mass_vx":         "10",
			"mass_vy":         "0",
		}},
		createUnequalMassFloatingWall,
		createMassAimedAtFloatingWall,
		advanceUntilFloatingWallCollision,
		assertFloatingWallMomentumUnchanged,
	)
}

func TestFloatingWallMomentumUsesDefaultMassAndSkipsMissingIDs(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Velocity: sim.Vec2{X: 3}})
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 4, Velocity: sim.Vec2{Y: 2}})

	got := totalMassMomentum(world, 1, 2, 9)

	if got != (sim.Vec2{X: 3, Y: 8}) {
		t.Fatalf("momentum = %#v, expected default and explicit mass contributions", got)
	}
}

func TestAssertMomentumReportsMismatch(t *testing.T) {
	if err := assertMomentum("momentum", sim.Vec2{X: 1}, sim.Vec2{X: 1}); err != nil {
		t.Fatal(err)
	}
	if err := assertMomentum("momentum", sim.Vec2{X: 1}, sim.Vec2{X: 2}); err == nil {
		t.Fatal("mismatched momentum should fail")
	}
}

func TestWallSpringBarrierTemperatureKickSteps(t *testing.T) {
	for _, example := range []map[string]string{
		{
			"spring_id":        "1",
			"temperature":      "0",
			"seed":             "11",
			"mass_id":          "3",
			"contact_fraction": "0.50",
			"kick_behavior":    "none",
		},
		{
			"spring_id":        "1",
			"temperature":      "10",
			"seed":             "11",
			"mass_id":          "3",
			"contact_fraction": "0.50",
			"kick_behavior":    "full screen height against gravity 10",
		},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, createWallSpringWithTemperature)
		mustWallSpringStep(t, w, example, setTemperatureRandomSeed)
		mustWallSpringStep(t, w, example, createMassCollidingWithWallSpring)
		mustWallSpringStep(t, w, example, resolveWallSpringCollision)
		mustWallSpringStep(t, w, example, assertMassTemperatureKick)
	}
}

func TestWallSpringBarrierNonWallTemperatureSteps(t *testing.T) {
	example := map[string]string{
		"spring_id":     "1",
		"wall":          "false",
		"temperature":   "10",
		"seed":          "11",
		"mass_id":       "3",
		"kick_behavior": "none",
		"new_wall":      "false",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, setBarrierSpringWallFalse)
	mustWallSpringStep(t, w, example, setSpringTemperature)
	mustWallSpringStep(t, w, example, setTemperatureRandomSeed)
	mustWallSpringStep(t, w, example, createMassCollidingWithSpring)
	mustWallSpringStep(t, w, example, resolveSpringCollision)
	mustWallSpringStep(t, w, example, assertMassTemperatureKick)
}

func TestWallSpringBarrierTemperatureKickStepReportsMissingMass(t *testing.T) {
	example := map[string]string{
		"mass_id":       "9",
		"kick_behavior": "none",
	}
	w := &world{}
	ensureDomainWorld(w)
	expectWallSpringStepError(t, w, example, assertMassTemperatureKick, "missing mass temperature kick assertion should fail")
}

func TestWallSpringBarrierTemperatureKickStepReportsUnsupportedBehavior(t *testing.T) {
	example := map[string]string{
		"spring_id":        "1",
		"temperature":      "10",
		"seed":             "11",
		"mass_id":          "3",
		"contact_fraction": "0.50",
		"kick_behavior":    "sideways",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createWallSpringWithTemperature)
	mustWallSpringStep(t, w, example, setTemperatureRandomSeed)
	mustWallSpringStep(t, w, example, createMassCollidingWithWallSpring)
	mustWallSpringStep(t, w, example, resolveWallSpringCollision)
	expectWallSpringStepError(t, w, example, assertMassTemperatureKick, "unsupported temperature kick behavior should fail")
}

func TestWallSpringBarrierRejectsUnsupportedTemperatureStepExample(t *testing.T) {
	example := map[string]string{"spring_id": "1", "temperature": "5"}
	expectWallSpringStepError(t, &world{}, example, setSpringTemperature, "unsupported temperature example should fail")
}

func TestWallSpringBarrierXSPPersistenceSteps(t *testing.T) {
	for _, test := range []struct {
		name     string
		examples []map[string]string
		steps    []func(*world, map[string]string) error
	}{
		{
			name: "wall",
			examples: []map[string]string{
				{"spring_id": "1", "input_wall": "true", "loaded_wall": "true", "saved_wall": "true"},
				{"spring_id": "1", "input_wall": "absent", "loaded_wall": "false", "saved_wall": "false"},
			},
			steps: []func(*world, map[string]string) error{createWallSpringXSPInput, loadAndSaveXSPInput, assertLoadedWallSpringXSP, assertSavedWallSpringXSP},
		},
		{
			name: "temperature",
			examples: []map[string]string{
				{"spring_id": "1", "input_temperature": "7.5", "loaded_temperature": "7.5", "saved_temperature": "7.5"},
				{"spring_id": "1", "input_temperature": "absent", "loaded_temperature": "0", "saved_temperature": "0"},
			},
			steps: []func(*world, map[string]string) error{createTemperatureSpringXSPInput, loadAndSaveXSPInput, assertLoadedSpringTemperatureXSP, assertSavedSpringTemperatureXSP},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runWallSpringStepExamples(t, test.examples, test.steps...)
		})
	}
}

func TestWallSpringBarrierVisibleControlSteps(t *testing.T) {
	runWallSpringStepExamples(t, wallSpringTwoStateExamples("old_wall", "new_wall", "false", "true", "true", "false"), createSelectedSpringWithWall, changeSpringWallControl, assertSpringWallValue)
}

func TestWallSpringBarrierSelectedSpringsWallSteps(t *testing.T) {
	example := map[string]string{
		"spring_ids": "1, 2, 3",
		"old_walls":  "false, false, true",
		"new_wall":   "true",
		"new_walls":  "true, true, true",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createSelectedSpringsWithWalls)
	mustWallSpringStep(t, w, example, changeSpringWallControl)
	mustWallSpringStep(t, w, example, assertSelectedSpringsWallValues)
}

func TestWallSpringBarrierSpringContextMenuSteps(t *testing.T) {
	for _, item := range []string{"Kspring", "Kdamp", "RestLen", "Wall", "Temperature"} {
		example := map[string]string{
			"spring_id": "1",
			"old_wall":  "false",
			"menu_item": item,
			"new_wall":  "true",
		}
		w := &world{}
		mustWallSpringStep(t, w, example, createMenuSpringWithWall)
		mustWallSpringStep(t, w, example, assertSpringMenuIncludesItem)
		mustWallSpringStep(t, w, example, selectSpringMenuWallItem)
		mustWallSpringStep(t, w, example, assertSpringWallValue)
	}
}

func TestWallSpringBarrierSpringTemperatureContextMenuSteps(t *testing.T) {
	example := map[string]string{
		"spring_id":         "1",
		"old_temperature":   "0",
		"minimum":           "0",
		"maximum":           "10",
		"new_temperature":   "7.5",
		"temperature":       "10",
		"kick_behavior":     "none",
		"input_temperature": "7.5",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createMenuSpringWithTemperature)
	mustWallSpringStep(t, w, example, selectSpringMenuTemperatureItem)
	mustWallSpringStep(t, w, example, assertSpringTemperatureDialogRange)
	mustWallSpringStep(t, w, example, changeSpringTemperatureDialogValue)
	mustWallSpringStep(t, w, example, assertSpringTemperatureValue)
}

func TestWallSpringBarrierTemperatureDialogRangeRequiresOpenDialog(t *testing.T) {
	example := map[string]string{"minimum": "0", "maximum": "10"}
	w := &world{}
	expectWallSpringStepError(t, w, example, assertSpringTemperatureDialogRange, "missing app game should fail")
	mustWallSpringStep(t, w, map[string]string{"spring_id": "1", "old_temperature": "0"}, createMenuSpringWithTemperature)
	expectWallSpringStepError(t, w, example, assertSpringTemperatureDialogRange, "closed temperature dialog should fail")
}

func TestWallSpringBarrierSpringContextMenuReportsMissingItem(t *testing.T) {
	example := map[string]string{
		"spring_id": "1",
		"old_wall":  "false",
		"menu_item": "Missing",
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createMenuSpringWithWall)
	if err := assertSpringMenuIncludesItem(w, example); err == nil {
		t.Fatal("missing spring menu item should fail")
	}
}

func TestWallSpringBarrierRenderingSteps(t *testing.T) {
	runWallSpringStepExamples(t, wallSpringTwoStateExamples("wall", "rendering_style", "false", "normal", "true", "wall"), createRenderableWallSpring, renderWallSpring, assertWallSpringRenderingStyle)
}

func wallSpringTwoStateExamples(firstKey, secondKey, firstA, secondA, firstB, secondB string) []map[string]string {
	return []map[string]string{
		{"spring_id": "1", firstKey: firstA, secondKey: secondA},
		{"spring_id": "1", firstKey: firstB, secondKey: secondB},
	}
}

func runWallSpringStepExamples(t *testing.T, examples []map[string]string, steps ...func(*world, map[string]string) error) {
	t.Helper()
	for _, example := range examples {
		w := &world{}
		for _, step := range steps {
			mustWallSpringStep(t, w, example, step)
		}
	}
}

func expectWallSpringStepError(t *testing.T, w *world, example map[string]string, step func(*world, map[string]string) error, failure string) {
	t.Helper()
	if err := step(w, example); err == nil {
		t.Fatal(failure)
	}
}

func mustWallSpringStep(t *testing.T, w *world, example map[string]string, step func(*world, map[string]string) error) {
	t.Helper()
	if err := step(w, example); err != nil {
		t.Fatal(err)
	}
}
