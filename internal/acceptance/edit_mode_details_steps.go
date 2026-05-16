package acceptance

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"springs/internal/edit"
	"springs/internal/sim"
)

var editInitialVelocity = sim.Vec2{X: 9, Y: 9}

func activateEditMode(w *world, _ map[string]string) error {
	editor := ensureMouseEditor(w)
	editor.Mode = edit.ModeEdit
	return nil
}

func addObjectNearPointer(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	return ensureEditMass(w, id, sim.Vec2{X: float64(id * 10), Y: 0}, false)
}

func setInitialEditSelection(w *world, example map[string]string) error {
	selection, err := editIDList(example, "initial_selection")
	if err != nil {
		return err
	}
	editor := ensureMouseEditor(w)
	editor.SelectedMasses = map[int]bool{}
	for _, id := range selection {
		if err := ensureEditMass(w, id, sim.Vec2{X: float64(id * 10), Y: 0}, false); err != nil {
			return err
		}
		editor.SelectedMasses[id] = true
	}
	return nil
}

func clickEditObject(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	action, err := stringValue(example, "click_action")
	if err != nil {
		return err
	}
	toggle, ok := editClickToggle[action]
	if !ok {
		return fmt.Errorf("unsupported click action %q", action)
	}
	return ensureMouseEditor(w).SelectNearest(editPointerPosition(id), toggle)
}

var editClickToggle = map[string]bool{
	"left clicks":       false,
	"shift left clicks": true,
}

func addObjectsInsideSelectionBox(w *world, example map[string]string) error {
	return addEditObjects(w, example, "inside_objects", insideSelectionBoxPosition)
}

func addObjectsOutsideSelectionBox(w *world, example map[string]string) error {
	return addEditObjects(w, example, "outside_objects", outsideSelectionBoxPosition)
}

func dragSelectionBox(w *world, example map[string]string) error {
	modifier, err := stringValue(example, "modifier")
	if err != nil {
		return err
	}
	switch modifier {
	case "none":
		ensureMouseEditor(w).BoxSelect(sim.Vec2{}, sim.Vec2{X: 50, Y: 50}, false)
	case "shift":
		ensureMouseEditor(w).BoxSelect(sim.Vec2{}, sim.Vec2{X: 50, Y: 50}, true)
	default:
		return fmt.Errorf("unsupported selection-box modifier %q", modifier)
	}
	return nil
}

func addSelectedObjectAtStart(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	position, err := positionValue(example, "start_position")
	if err != nil {
		return err
	}
	if err := ensureEditMass(w, id, position, false); err != nil {
		return err
	}
	ensureMouseEditor(w).SelectedMasses[id] = true
	return nil
}

func middleDragSelectedObjects(w *world, example map[string]string) error {
	return applyEditVector(w, example, "drag_delta", (*edit.Editor).MoveSelected)
}

func assertEditObjectPosition(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	expected, err := positionValue(example, "expected_position")
	if err != nil {
		return err
	}
	mass, ok := ensureDomainWorld(w).MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	if mass.Position != expected {
		return fmt.Errorf("mass %d position = %#v, want %#v", id, mass.Position, expected)
	}
	return nil
}

func addSelectedMassWithFixedState(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	fixed, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	if err := ensureEditMass(w, id, sim.Vec2{X: float64(id * 10), Y: 0}, fixed); err != nil {
		return err
	}
	setEditMassVelocity(w, id, editInitialVelocity)
	ensureMouseEditor(w).SelectedMasses[id] = true
	return nil
}

func rightDragSelectedMasses(w *world, example map[string]string) error {
	return applyEditVector(w, example, "release_velocity", (*edit.Editor).ThrowSelected)
}

func assertEditMassVelocity(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, err := editMassByID(w, id)
	if err != nil {
		return err
	}
	return assertEditMassExpectedVelocity(id, mass, example)
}

func assertEditMassExpectedVelocity(id int, mass sim.Mass, example map[string]string) error {
	expectedText, err := stringValue(example, "expected_velocity")
	if err != nil {
		return err
	}
	if expectedText == "unchanged" {
		return assertEditVelocityUnchanged(id, mass)
	}
	expected, err := positionValue(example, "expected_velocity")
	if err != nil {
		return err
	}
	return assertEditVelocityEquals(id, mass, expected)
}

func assertEditSelection(w *world, example map[string]string) error {
	expected, err := editIDList(example, "expected_selection")
	if err != nil {
		return err
	}
	actual := selectedEditMassIDs(ensureMouseEditor(w))
	if strings.Join(intStrings(actual), ",") != strings.Join(intStrings(expected), ",") {
		return fmt.Errorf("selection = %v, want %v", actual, expected)
	}
	return nil
}

func addEditObjects(w *world, example map[string]string, key string, position func(int) sim.Vec2) error {
	ids, err := editIDList(example, key)
	if err != nil {
		return err
	}
	for index, id := range ids {
		if err := ensureEditMass(w, id, position(index)); err != nil {
			return err
		}
	}
	return nil
}

func editPointerPosition(id int) sim.Vec2 {
	return sim.Vec2{X: float64(id * 10), Y: 0}
}

func insideSelectionBoxPosition(index int) sim.Vec2 {
	return sim.Vec2{X: float64(10 + index*10), Y: 10}
}

func outsideSelectionBoxPosition(index int) sim.Vec2 {
	return sim.Vec2{X: float64(100 + index*10), Y: 100}
}

func applyEditVector(w *world, example map[string]string, key string, action func(*edit.Editor, sim.Vec2)) error {
	vector, err := positionValue(example, key)
	if err != nil {
		return err
	}
	action(ensureMouseEditor(w), vector)
	return nil
}

func ensureEditMass(w *world, id int, position sim.Vec2, fixed ...bool) error {
	if _, ok := ensureDomainWorld(w).MassByID(id); ok {
		return nil
	}
	isFixed := len(fixed) > 0 && fixed[0]
	return ensureDomainWorld(w).AddMass(sim.Mass{ID: id, Position: position, Mass: 1, Fixed: isFixed})
}

func setEditMassVelocity(w *world, id int, velocity sim.Vec2) {
	for index := range ensureDomainWorld(w).Masses {
		if ensureDomainWorld(w).Masses[index].ID == id {
			ensureDomainWorld(w).Masses[index].Velocity = velocity
			return
		}
	}
}

func editMassByID(w *world, id int) (sim.Mass, error) {
	mass, ok := ensureDomainWorld(w).MassByID(id)
	if !ok {
		return sim.Mass{}, fmt.Errorf("mass %d not found", id)
	}
	return mass, nil
}

func editIDList(example map[string]string, key string) ([]int, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return nil, err
	}
	return parseEditIDList(key, value)
}

func parseEditIDList(key string, value string) ([]int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "none" || trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		id, err := parseEditIDPart(part, key, value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids, nil
}

func parseEditIDPart(part string, key string, value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(part))
	if err != nil {
		return 0, fmt.Errorf("invalid id list %s=%q", key, value)
	}
	return id, nil
}

func assertEditVelocityUnchanged(id int, mass sim.Mass) error {
	if mass.Velocity != editInitialVelocity {
		return fmt.Errorf("mass %d velocity changed to %#v", id, mass.Velocity)
	}
	return nil
}

func assertEditVelocityEquals(id int, mass sim.Mass, expected sim.Vec2) error {
	if mass.Velocity != expected {
		return fmt.Errorf("mass %d velocity = %#v, want %#v", id, mass.Velocity, expected)
	}
	return nil
}

func selectedEditMassIDs(editor *edit.Editor) []int {
	ids := make([]int, 0, len(editor.SelectedMasses))
	for id, selected := range editor.SelectedMasses {
		if selected {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids
}

func intStrings(ids []int) []string {
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = strconv.Itoa(id)
	}
	return values
}
