package acceptance

import (
	"fmt"
	"strconv"
	"strings"

	"springs/internal/edit"
	"springs/internal/sim"
)

func activateSpringMode(w *world, _ map[string]string) error {
	ensureMouseEditor(w).Mode = edit.ModeAddSpring
	return nil
}

func pressNearSpringMass(w *world, example map[string]string) error {
	id, err := intValue(example, "start_mass")
	if err != nil {
		return err
	}
	w.springStartMassID = id
	return ensureSpringModeMass(w, id, springModeMassPosition(id))
}

func releaseSpringPointer(w *world, example map[string]string) error {
	if err := ensureMouseEditor(w).BeginSpring(springModeMassPosition(w.springStartMassID), edit.SpringButtonLeft); err != nil {
		return err
	}
	position, err := releaseTargetPosition(w, example)
	if err != nil {
		return err
	}
	id, created, err := ensureMouseEditor(w).ReleaseSpring(position)
	w.createdSpringID = id
	w.springCreated = created
	return err
}

func assertSpringCreationResult(w *world, example map[string]string) error {
	result, err := stringValue(example, "result")
	if err != nil {
		return err
	}
	if result == "discard pending spring" {
		return assertSpringDiscarded(w)
	}
	massA, massB, ok := parseCreatedSpringResult(result)
	if !ok {
		return fmt.Errorf("unsupported spring creation result %q", result)
	}
	return assertCreatedSpringEndpoints(w, massA, massB)
}

func dragSpringWithButton(w *world, example map[string]string) error {
	button, err := stringValue(example, "button")
	if err != nil {
		return err
	}
	editor := ensureMouseEditor(w)
	if err := editor.BeginSpring(springModeMassPosition(w.springStartMassID), button); err != nil {
		return err
	}
	editor.DragSpring(sim.Vec2{X: 15, Y: 5})
	w.springBehavior = pendingSpringBehavior(editor)
	return discardTemporarySpring(w, editor, button)
}

func assertPendingSpringBehavior(w *world, example map[string]string) error {
	expected, err := stringValue(example, "behavior")
	if err != nil {
		return err
	}
	if w.springBehavior != expected {
		return fmt.Errorf("pending spring behavior = %q, want %q", w.springBehavior, expected)
	}
	return nil
}

func setCurrentKspring(w *world, example map[string]string) error {
	return setSpringParameter(w, example, "spring constant", "kspring")
}

func setCurrentKdamp(w *world, example map[string]string) error {
	return setSpringParameter(w, example, "damping", "kdamp")
}

func createSpringWithLength(w *world, example map[string]string) error {
	length, err := floatValue(example, "creation_length")
	if err != nil {
		return err
	}
	if err := ensureSpringModeMass(w, 1, sim.Vec2{}); err != nil {
		return err
	}
	if err := ensureSpringModeMass(w, 2, sim.Vec2{X: length}); err != nil {
		return err
	}
	editor := ensureMouseEditor(w)
	if err := editor.BeginSpring(springModeMassPosition(1), edit.SpringButtonLeft); err != nil {
		return err
	}
	id, created, err := editor.ReleaseSpring(sim.Vec2{X: length})
	w.createdSpringID = id
	w.springCreated = created
	return err
}

func assertCreatedSpringKspring(w *world, example map[string]string) error {
	return assertCreatedSpringFloat(w, example, "kspring", func(spring sim.Spring) float64 { return spring.SpringConstant })
}

func assertCreatedSpringKdamp(w *world, example map[string]string) error {
	return assertCreatedSpringFloat(w, example, "kdamp", func(spring sim.Spring) float64 { return spring.Damping })
}

func assertCreatedSpringRestLength(w *world, example map[string]string) error {
	return assertCreatedSpringFloat(w, example, "creation_length", func(spring sim.Spring) float64 { return spring.RestLength })
}

func ensureSpringModeMass(w *world, id int, position sim.Vec2) error {
	if _, ok := ensureDomainWorld(w).MassByID(id); ok {
		return nil
	}
	return ensureDomainWorld(w).AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
}

func springModeMassPosition(id int) sim.Vec2 {
	return sim.Vec2{X: float64((id - 1) * 30)}
}

func releaseTargetPosition(w *world, example map[string]string) (sim.Vec2, error) {
	target, err := stringValue(example, "release_target")
	if err != nil {
		return sim.Vec2{}, err
	}
	if target == "away from mass" {
		return sim.Vec2{X: 1000, Y: 1000}, nil
	}
	id, ok := parseNearMass(target)
	if !ok {
		return sim.Vec2{}, fmt.Errorf("unsupported release target %q", target)
	}
	position := springModeMassPosition(id)
	return position, ensureSpringModeMass(w, id, position)
}

func parseNearMass(value string) (int, bool) {
	parts := strings.Fields(value)
	if len(parts) != 3 || parts[0] != "near" || parts[1] != "mass" {
		return 0, false
	}
	id, err := strconv.Atoi(parts[2])
	return id, err == nil
}

func parseCreatedSpringResult(value string) (int, int, bool) {
	parts := strings.Fields(value)
	if len(parts) != 6 || strings.Join(parts[:3], " ") != "create spring between" || parts[4] != "and" {
		return 0, 0, false
	}
	massA, errA := strconv.Atoi(parts[3])
	massB, errB := strconv.Atoi(parts[5])
	return massA, massB, errA == nil && errB == nil
}

func assertSpringDiscarded(w *world) error {
	if w.springCreated || len(ensureDomainWorld(w).Springs) != 0 {
		return fmt.Errorf("spring was created")
	}
	return nil
}

func assertCreatedSpringEndpoints(w *world, massA int, massB int) error {
	spring, err := createdMouseSpring(w)
	if err != nil {
		return err
	}
	if !w.springCreated || spring.MassA != massA || spring.MassB != massB {
		return fmt.Errorf("created spring = %#v, created=%t", spring, w.springCreated)
	}
	return nil
}

func pendingSpringBehavior(editor *edit.Editor) string {
	pending, ok := editor.PendingSpring()
	if !ok {
		return "none"
	}
	if pending.Temporary {
		return "temporary cursor spring"
	}
	if pending.Active {
		return "actively affects the first mass"
	}
	return "inactive until the spring is placed"
}

func discardTemporarySpring(w *world, editor *edit.Editor, button string) error {
	if button != edit.SpringButtonMiddle {
		return nil
	}
	if _, created, err := editor.ReleaseSpring(sim.Vec2{X: 15, Y: 5}); err != nil || created {
		return fmt.Errorf("temporary spring release created=%t err=%v", created, err)
	}
	w.springBehavior = "temporary cursor spring discarded on release"
	return nil
}

func setSpringParameter(w *world, example map[string]string, parameter string, key string) error {
	value, err := stringValue(example, key)
	if err != nil {
		return err
	}
	ensureDomainWorld(w).Parameters.Set(parameter, value)
	return nil
}

func assertCreatedSpringFloat(w *world, example map[string]string, key string, field func(sim.Spring) float64) error {
	expected, err := floatValue(example, key)
	if err != nil {
		return err
	}
	spring, err := createdMouseSpring(w)
	if err != nil {
		return err
	}
	if field(spring) != expected {
		return fmt.Errorf("spring %s = %f, want %f", key, field(spring), expected)
	}
	return nil
}
