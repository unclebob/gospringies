package acceptance

import (
	"fmt"

	"springs/internal/edit"
	"springs/internal/sim"
)

func createSelectedParameterMass(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	if err := ensureDomainWorld(w).AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: 10}, Mass: 1, Elasticity: 0.2}); err != nil {
		return err
	}
	return ensureMouseEditor(w).SelectMass(id)
}

func changeMassControl(w *world, example map[string]string) error {
	return changeControlFromExample(w, example, "control", "value")
}

func assertMassControlValue(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, ok := ensureDomainWorld(w).MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	control, value, err := stringPair(example, "control", "value")
	if err != nil {
		return err
	}
	return assertMassControlField(mass, control, value)
}

func createSelectedParameterSpring(w *world, example map[string]string) error {
	id, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	if err := addParameterSpring(w, id, 30); err != nil {
		return err
	}
	return ensureMouseEditor(w).SelectSpring(id)
}

func changeSpringControl(w *world, example map[string]string) error {
	return changeControlFromExample(w, example, "control", "value")
}

func assertSpringControlValue(w *world, example map[string]string) error {
	id, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	spring, ok := ensureDomainWorld(w).SpringByID(id)
	if !ok {
		return fmt.Errorf("spring %d not found", id)
	}
	control, value, err := stringPair(example, "control", "value")
	if err != nil {
		return err
	}
	return assertSpringControlField(spring, control, value)
}

func createSelectedSpringWithCurrentLength(w *world, example map[string]string) error {
	id, length, err := intAndFloat(example, "spring_id", "current_length")
	if err != nil {
		return err
	}
	if err := addParameterSpring(w, id, length); err != nil {
		return err
	}
	return ensureMouseEditor(w).SelectSpring(id)
}

func setSelectedRestLength(w *world, _ map[string]string) error {
	return ensureMouseEditor(w).SetRestLength()
}

func assertSelectedSpringRestLength(w *world, example map[string]string) error {
	id, expected, err := intAndFloat(example, "spring_id", "current_length")
	if err != nil {
		return err
	}
	spring, ok := ensureDomainWorld(w).SpringByID(id)
	if !ok {
		return fmt.Errorf("spring %d not found", id)
	}
	return assertFloat("spring rest length", spring.RestLength, expected)
}

func createNoCompatibleSelection(w *world, example map[string]string) error {
	control, err := stringValue(example, "control")
	if err != nil {
		return err
	}
	if isMassControl(control) {
		return createSelectedParameterSpring(w, map[string]string{"spring_id": "1"})
	}
	if isSpringControl(control) {
		if err := createSelectedParameterMass(w, map[string]string{"mass_id": "1"}); err != nil {
			return err
		}
		return ensureSelectionMass(ensureDomainWorld(w), 2)
	}
	return fmt.Errorf("unsupported control %q", control)
}

func changeGenericControl(w *world, example map[string]string) error {
	return changeControlFromExample(w, example, "control", "value")
}

func assertFutureObjectUsesControlValue(w *world, example map[string]string) error {
	objectType, control, value, err := futureObjectExpectation(example)
	if err != nil {
		return err
	}
	switch objectType {
	case "mass":
		return assertFutureMassControl(w, control, value)
	case "spring":
		return assertFutureSpringControl(w, control, value)
	default:
		return fmt.Errorf("unsupported object type %q", objectType)
	}
}

func changeControlFromExample(w *world, example map[string]string, controlKey string, valueKey string) error {
	control, value, err := stringPair(example, controlKey, valueKey)
	if err != nil {
		return err
	}
	return ensureMouseEditor(w).ChangeControl(control, value)
}

func addParameterSpring(w *world, id int, length float64) error {
	world := ensureDomainWorld(w)
	if err := ensureParameterMass(world, 1, sim.Vec2{X: 0, Y: 20}); err != nil {
		return err
	}
	if err := ensureParameterMass(world, 2, sim.Vec2{X: length, Y: 20}); err != nil {
		return err
	}
	return world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2, RestLength: 1, SpringConstant: 8, Damping: 0.2})
}

func ensureParameterMass(world *sim.Simulation, id int, position sim.Vec2) error {
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			world.Masses[i].Position = position
			return nil
		}
	}
	return world.AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
}

func assertMassControlField(mass sim.Mass, control string, value string) error {
	switch control {
	case "mass":
		return assertStringFloat("mass", mass.Mass, value)
	case "elasticity":
		return assertStringFloat("elasticity", mass.Elasticity, value)
	case "fixed":
		return assertStringBool("fixed", mass.Fixed, value)
	default:
		return fmt.Errorf("unsupported mass control %q", control)
	}
}

func assertSpringControlField(spring sim.Spring, control string, value string) error {
	switch control {
	case "Kspring":
		return assertStringFloat("Kspring", spring.SpringConstant, value)
	case "Kdamp":
		return assertStringFloat("Kdamp", spring.Damping, value)
	default:
		return fmt.Errorf("unsupported spring control %q", control)
	}
}

func assertFutureMassControl(w *world, control string, value string) error {
	editor := ensureMouseEditor(w)
	editor.Mode = edit.ModeAddMass
	id, err := editor.Click(sim.Vec2{X: 100, Y: 20})
	if err != nil {
		return err
	}
	mass, _ := ensureDomainWorld(w).MassByID(id)
	return assertMassControlField(mass, control, value)
}

func assertFutureSpringControl(w *world, control string, value string) error {
	editor := ensureMouseEditor(w)
	editor.Mode = edit.ModeAddSpring
	id, err := editor.CreateSpring(1, 2)
	if err != nil {
		return err
	}
	spring, _ := ensureDomainWorld(w).SpringByID(id)
	return assertSpringControlField(spring, control, value)
}

func futureObjectExpectation(example map[string]string) (string, string, string, error) {
	control, value, err := stringPair(example, "control", "value")
	if err != nil {
		return "", "", "", err
	}
	objectType, err := stringValue(example, "object_type")
	return objectType, control, value, err
}

func isMassControl(control string) bool {
	return control == "mass" || control == "elasticity" || control == "fixed"
}

func isSpringControl(control string) bool {
	return control == "Kspring" || control == "Kdamp"
}

func intAndFloat(example map[string]string, intKey string, floatKey string) (int, float64, error) {
	id, err := intValue(example, intKey)
	if err != nil {
		return 0, 0, err
	}
	value, err := floatValue(example, floatKey)
	return id, value, err
}

func assertStringFloat(name string, actual float64, expected string) error {
	value, err := floatValue(map[string]string{name: expected}, name)
	if err != nil {
		return err
	}
	return assertFloat(name, actual, value)
}

func assertStringBool(name string, actual bool, expected string) error {
	value, err := boolValue(map[string]string{name: expected}, name)
	if err != nil {
		return err
	}
	if actual != value {
		return fmt.Errorf("%s = %t, want %t", name, actual, value)
	}
	return nil
}
