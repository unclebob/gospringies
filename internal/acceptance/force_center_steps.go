package acceptance

import (
	"fmt"
	"math"
	"strings"

	"springs/internal/app"
	"springs/internal/sim"
)

var supportedForceNames = map[string]struct{}{
	"gravity":                   {},
	"center of mass attraction": {},
	"center attraction":         {},
	"wall repulsion":            {},
}

const forceDirectionTolerance = 0.000001

func selectForce(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if !supportedForceName(force) {
		return fmt.Errorf("unsupported force %q", force)
	}
	world := ensureDomainWorld(w)
	world.Parameters.SelectForce(force)
	return nil
}

func assertForceExposesParameter(_ *world, example map[string]string) error {
	force, parameter, err := stringPair(example, "force", "parameter_one")
	if err != nil {
		return err
	}
	if !hasForceParameter(force, parameter) {
		return fmt.Errorf("%s does not expose %s", force, parameter)
	}
	second, err := stringValue(example, "parameter_two")
	if err != nil {
		return err
	}
	if !hasForceParameter(force, second) {
		return fmt.Errorf("%s does not expose %s", force, second)
	}
	return nil
}

func hasForceParameter(force, parameter string) bool {
	for _, candidate := range sim.ForceParameterNames(force) {
		if candidate == parameter {
			return true
		}
	}
	return false
}

func supportedForceName(force string) bool {
	_, ok := supportedForceNames[force]
	return ok
}

func setGravityDirection(w *world, example map[string]string) error {
	direction, err := stringValue(example, "direction_degrees")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1", "direction": direction})
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	return nil
}

func evaluateGravity(w *world, _ map[string]string) error {
	return evaluateCurrentForces(w)
}

func assertGravityDirection(w *world, example map[string]string) error {
	expected, err := stringValue(example, "expected_direction")
	if err != nil {
		return err
	}
	force := w.forceEvaluation.ByMassID[1].Force
	if !matchesExpectedDirection(force, expected) {
		return fmt.Errorf("gravity force = %#v, want %s", force, expected)
	}
	return nil
}

func matchesExpectedDirection(force sim.Vec2, expected string) bool {
	directions := map[string]sim.Vec2{
		"down":  {Y: -1},
		"right": {X: 1},
		"up":    {Y: 1},
		"left":  {X: -1},
	}
	want, ok := directions[expected]
	return ok && matchesForceDirectionComponent(force.X, want.X) && matchesForceDirectionComponent(force.Y, want.Y)
}

func matchesForceDirectionComponent(actual, expected float64) bool {
	return math.Abs(actual-expected) < forceDirectionTolerance
}

func createSelectedMasses(w *world, example map[string]string) error {
	selected, err := stringValue(example, "selected_masses")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	ids, err := selectedMassIDs(selected)
	if err != nil {
		return err
	}
	w.originalMassIDs = ids
	return nil
}

func selectedMassIDs(selected string) ([]int, error) {
	if selected == "none" {
		return nil, nil
	}
	if selected == "1" {
		return []int{1}, nil
	}
	return nil, fmt.Errorf("unsupported selected masses %q", selected)
}

func setForceCenter(w *world, _ map[string]string) error {
	ensureDomainWorld(w).SetForceCenter(w.originalMassIDs)
	return nil
}

func assertForceCenter(w *world, example map[string]string) error {
	expected, err := stringValue(example, "expected_center")
	if err != nil {
		return err
	}
	actual := "screen center"
	if ensureDomainWorld(w).CenterMassID() > 0 {
		actual = fmt.Sprintf("mass %d", ensureDomainWorld(w).CenterMassID())
	}
	if actual != expected {
		return fmt.Errorf("force center = %s, want %s", actual, expected)
	}
	return nil
}

func createForceCenterMass(w *world, example map[string]string) error {
	id, err := intValue(example, "center_mass")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	_ = world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: 50, Y: 50}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: id + 1, Position: sim.Vec2{X: 0, Y: 50}, Mass: 1})
	world.SetForceCenter([]int{id})
	return nil
}

func enableNamedForce(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if !supportedForceName(force) {
		return fmt.Errorf("unsupported force %q", force)
	}
	ensureDomainWorld(w).Parameters.EnableForce(force, map[string]string{"magnitude": "10", "exponent": "0", "damping": "1"})
	return nil
}

func evaluateCenterForces(w *world, _ map[string]string) error {
	return evaluateCurrentForces(w)
}

func evaluateCurrentForces(w *world) error {
	w.forceEvaluation = ensureDomainWorld(w).EvaluateForces()
	return nil
}

func assertCenterMassVisuallyMarked(w *world, example map[string]string) error {
	id, err := intValue(example, "center_mass")
	if err != nil {
		return err
	}
	game := app.NewGame()
	game.ReplaceWorld(ensureDomainWorld(w))
	if !game.RenderWorld().HasVisibleRepresentation("force center") || !game.World().IsCenterMass(id) {
		return fmt.Errorf("center mass %d was not visually marked", id)
	}
	return nil
}

func assertNoReciprocalCenterForce(w *world, example map[string]string) error {
	id, err := intValue(example, "center_mass")
	if err != nil {
		return err
	}
	forceName, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if w.forceEvaluation.ByMassID[id].Force != (sim.Vec2{}) {
		return fmt.Errorf("center mass received reciprocal response from %s: %#v", forceName, w.forceEvaluation.ByMassID[id].Force)
	}
	return nil
}

func enableForceForControls(w *world, example map[string]string) error {
	return enableNamedForce(w, example)
}

func assertForceControlsActive(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if !supportedForceName(force) {
		return fmt.Errorf("unsupported force %q", force)
	}
	active := strings.TrimSpace(ensureDomainWorld(w).Parameters.ActiveForce)
	if active != force {
		return fmt.Errorf("active force controls = %q, want %q", active, force)
	}
	return nil
}
