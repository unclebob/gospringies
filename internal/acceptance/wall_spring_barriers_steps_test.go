package acceptance

import "testing"

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

func TestWallSpringBarrierCollisionSteps(t *testing.T) {
	example := map[string]string{
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
	}
	w := &world{}
	mustWallSpringStep(t, w, example, createWallSpringByCoordinates)
	mustWallSpringStep(t, w, example, createBarrierMovingMass)
	mustWallSpringStep(t, w, example, advanceThroughWallSpringCollision)
	mustWallSpringStep(t, w, example, assertMassOnStartingWallSpringSide)
	mustWallSpringStep(t, w, example, assertMassVelocityResolvedAwayFromWallSpring)
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

func TestWallSpringBarrierXSPSteps(t *testing.T) {
	for _, example := range []map[string]string{
		{"spring_id": "1", "input_wall": "true", "loaded_wall": "true", "saved_wall": "true"},
		{"spring_id": "1", "input_wall": "absent", "loaded_wall": "false", "saved_wall": "false"},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, createWallSpringXSPInput)
		mustWallSpringStep(t, w, example, loadAndSaveXSPInput)
		mustWallSpringStep(t, w, example, assertLoadedWallSpringXSP)
		mustWallSpringStep(t, w, example, assertSavedWallSpringXSP)
	}
}

func TestWallSpringBarrierVisibleControlSteps(t *testing.T) {
	for _, example := range []map[string]string{
		{"spring_id": "1", "old_wall": "false", "new_wall": "true"},
		{"spring_id": "1", "old_wall": "true", "new_wall": "false"},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, createSelectedSpringWithWall)
		mustWallSpringStep(t, w, example, changeSpringWallControl)
		mustWallSpringStep(t, w, example, assertSpringWallValue)
	}
}

func TestWallSpringBarrierSpringContextMenuSteps(t *testing.T) {
	for _, item := range []string{"Kspring", "Kdamp", "RestLen", "Wall"} {
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
	for _, example := range []map[string]string{
		{"spring_id": "1", "wall": "false", "rendering_style": "normal"},
		{"spring_id": "1", "wall": "true", "rendering_style": "wall"},
	} {
		w := &world{}
		mustWallSpringStep(t, w, example, createRenderableWallSpring)
		mustWallSpringStep(t, w, example, renderWallSpring)
		mustWallSpringStep(t, w, example, assertWallSpringRenderingStyle)
	}
}

func mustWallSpringStep(t *testing.T, w *world, example map[string]string, step func(*world, map[string]string) error) {
	t.Helper()
	if err := step(w, example); err != nil {
		t.Fatal(err)
	}
}
