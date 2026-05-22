package acceptance

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"springs/internal/edit"
	"springs/internal/sim"
)

const (
	mouseDefaultMass       = 2.5
	mouseDefaultElasticity = 0.6
)

var mouseGridSnapStates = map[string]bool{"enabled": true, "disabled": false}

func setMouseEditorMode(w *world, example map[string]string) error {
	return updateMouseEditorString(w, example, "mode", func(editor *edit.Editor, value string) { editor.Mode = value })
}

func setMouseEditorModeAddMass(w *world, _ map[string]string) error {
	ensureMouseEditor(w).Mode = edit.ModeAddMass
	return nil
}

func configureCurrentMassDefaults(w *world, _ map[string]string) error {
	world := ensureDomainWorld(w)
	world.Parameters.Set("current mass", "2.5")
	world.Parameters.Set("elasticity", "0.6")
	return nil
}

func clickMouseEditor(w *world, example map[string]string) error {
	position, err := positionValue(example, "pointer_position")
	if err != nil {
		return err
	}
	id, err := ensureMouseEditor(w).Click(position)
	w.createdMassID = id
	return err
}

func assertCreatedMassPosition(w *world, example map[string]string) error {
	position, err := positionValue(example, "expected_position")
	if err != nil {
		return err
	}
	mass, err := createdMouseMass(w)
	if err != nil {
		return err
	}
	return assertVec("created mass position", mass.Position, position.X, position.Y)
}

func assertCreatedMassDefaults(w *world, _ map[string]string) error {
	mass, err := createdMouseMass(w)
	if err != nil {
		return err
	}
	if mass.Mass != mouseDefaultMass || mass.Elasticity != mouseDefaultElasticity {
		return fmt.Errorf("mass defaults = %f, %f", mass.Mass, mass.Elasticity)
	}
	return nil
}

func assertMassPlacementConstrainedToGrid(w *world, example map[string]string) error {
	snap, err := stringValue(example, "grid_snap")
	if err != nil {
		return err
	}
	enabled, ok := booleanState(snap, mouseGridSnapStates)
	if !ok {
		return fmt.Errorf("unsupported grid snap %q", snap)
	}
	if !enabled {
		return nil
	}
	mass, err := createdMouseMass(w)
	if err != nil {
		return err
	}
	return assertGridPoint("mass placement", mass.Position, ensureMouseEditor(w).GridSnapSize)
}

func setMouseGridSnap(w *world, example map[string]string) error {
	snap, err := stringValue(example, "grid_snap")
	if err != nil {
		return err
	}
	enabled, ok := booleanState(snap, mouseGridSnapStates)
	if !ok {
		return fmt.Errorf("unsupported grid snap %q", snap)
	}
	ensureMouseEditor(w).GridSnapEnabled = enabled
	return nil
}

func setMouseGridSnapEnabled(w *world, _ map[string]string) error {
	ensureMouseEditor(w).GridSnapEnabled = true
	return nil
}

func setMouseGridSnapSize(w *world, example map[string]string) error {
	return updateMouseEditorFloat(w, example, "snap_size", func(editor *edit.Editor, value float64) { editor.GridSnapSize = value })
}

func addMouseMassA(w *world, example map[string]string) error {
	return addMouseMass(w, example, "mass_a")
}

func addMouseMassB(w *world, example map[string]string) error {
	return addMouseMass(w, example, "mass_b")
}

func createMouseSpring(w *world, example map[string]string) error {
	massA, massB, err := springForceMassIDs(example)
	if err != nil {
		return err
	}
	id, err := ensureMouseEditor(w).CreateSpring(massA, massB)
	w.createdSpringID = id
	return err
}

func assertMouseSpringEndpoints(w *world, example map[string]string) error {
	massA, massB, err := springForceMassIDs(example)
	if err != nil {
		return err
	}
	spring, err := createdMouseSpring(w)
	if err != nil {
		return err
	}
	if spring.MassA != massA || spring.MassB != massB {
		return fmt.Errorf("spring endpoints = %d, %d", spring.MassA, spring.MassB)
	}
	return nil
}

func assertMouseSpringDefaults(w *world, _ map[string]string) error {
	spring, err := createdMouseSpring(w)
	if err != nil {
		return err
	}
	if spring.SpringConstant != 12 || spring.Damping != 0.7 {
		return fmt.Errorf("spring defaults = %f, %f", spring.SpringConstant, spring.Damping)
	}
	return nil
}

func dragMouseMass(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	position, err := positionValue(example, "target_position")
	if err != nil {
		return err
	}
	if game, ok := dragModeGame(w); ok {
		return dragAppMass(w, game, id, position)
	}
	return ensureMouseEditor(w).DragMass(id, position)
}

func dragMouseMassThrough(w *world, example map[string]string) error {
	if err := dragMouseMassToPosition(w, example, "drag_position"); err != nil {
		return err
	}
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, ok := w.domainWorld.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	w.mouseDragPosition = mass.Position
	w.mouseDragRecorded = true
	return dragMouseMassToPosition(w, example, "target_position")
}

func dragMouseMassToPosition(w *world, example map[string]string, key string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	position, err := positionValue(example, key)
	if err != nil {
		return err
	}
	if game, ok := dragModeGame(w); ok {
		return dragAppMass(w, game, id, position)
	}
	return ensureMouseEditor(w).DragMass(id, position)
}

func dragModeGame(w *world) (*driverGame, bool) {
	return optionalConcreteApplicationDriver(w)
}

func dragAppMass(w *world, game *driverGame, id int, position sim.Vec2) error {
	if w.domainWorld != nil {
		game.ReplaceWorld(w.domainWorld)
	}
	if !game.DragMass(id, position) {
		return fmt.Errorf("mass %d was not draggable", id)
	}
	w.domainWorld = game.World().Clone()
	return nil
}

func assertMouseMassPosition(w *world, example map[string]string) error {
	return assertMouseMassPositionValue(w, example, "expected_position", "mass position")
}

func assertMouseMassDragPosition(w *world, example map[string]string) error {
	if _, err := intValue(example, "mass_id"); err != nil {
		return err
	}
	position, err := positionValue(example, "snapped_drag_position")
	if err != nil {
		return err
	}
	if !w.mouseDragRecorded {
		return fmt.Errorf("mass drag position was not recorded")
	}
	return assertVec("mass drag position", w.mouseDragPosition, position.X, position.Y)
}

func assertMouseMassInitialPosition(w *world, example map[string]string) error {
	return assertMouseMassPositionValue(w, example, "expected_start_position", "mass initial position")
}

func assertMouseMassPositionValue(w *world, example map[string]string, key string, name string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	position, err := positionValue(example, key)
	if err != nil {
		return err
	}
	mass, ok := w.domainWorld.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	return assertVec(name, mass.Position, position.X, position.Y)
}

func assertMouseMassID(w *world, example map[string]string) error {
	return withMouseMass(w, example, func(_ sim.Mass) error { return nil })
}

func assertMouseMassExpectedID(w *world, example map[string]string) error {
	expected, err := intValue(example, "expected_mass_id")
	if err != nil {
		return err
	}
	return withMouseMass(w, example, func(mass sim.Mass) error {
		if mass.ID != expected {
			return fmt.Errorf("mass id = %d, want %d", mass.ID, expected)
		}
		return nil
	})
}

func ensureMouseEditor(w *world) *edit.Editor {
	if w.mouseEditor == nil {
		w.mouseEditor = edit.NewEditor(ensureDomainWorld(w))
	}
	return w.mouseEditor
}

func createdMouseMass(w *world) (sim.Mass, error) {
	return createdMouseObject(w.createdMassID, "mass", w.domainWorld.MassByID)
}

func createdMouseSpring(w *world) (sim.Spring, error) {
	return createdMouseObject(w.createdSpringID, "spring", w.domainWorld.SpringByID)
}

func createdMouseObject[T any](id int, name string, lookup func(int) (T, bool)) (T, error) {
	object, ok := lookup(id)
	if !ok {
		var zero T
		return zero, fmt.Errorf("created %s %d not found", name, id)
	}
	return object, nil
}

func addMouseMass(w *world, example map[string]string, key string) error {
	id, err := intValue(example, key)
	if err != nil {
		return err
	}
	position := sim.Vec2{X: float64(id * 20), Y: 20}
	return ensureDomainWorld(w).AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
}

func positionValue(example map[string]string, key string) (sim.Vec2, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return sim.Vec2{}, err
	}
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return sim.Vec2{}, fmt.Errorf("invalid position %s=%q", key, value)
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return sim.Vec2{}, fmt.Errorf("invalid position x %s=%q", key, value)
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return sim.Vec2{}, fmt.Errorf("invalid position y %s=%q", key, value)
	}
	return sim.Vec2{X: x, Y: y}, nil
}

func updateMouseEditorString(w *world, example map[string]string, key string, update func(*edit.Editor, string)) error {
	return updateMouseEditorValue(w, example, key, stringValue, update)
}

func updateMouseEditorFloat(w *world, example map[string]string, key string, update func(*edit.Editor, float64)) error {
	return updateMouseEditorValue(w, example, key, floatValue, update)
}

func updateMouseEditorValue[T any](
	w *world,
	example map[string]string,
	key string,
	read func(map[string]string, string) (T, error),
	update func(*edit.Editor, T),
) error {
	value, err := read(example, key)
	if err != nil {
		return err
	}
	update(ensureMouseEditor(w), value)
	return nil
}

func withMouseMass(w *world, example map[string]string, check func(sim.Mass) error) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, ok := w.domainWorld.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	return check(mass)
}

func assertGridPoint(name string, position sim.Vec2, size float64) error {
	if size <= 0 {
		return fmt.Errorf("%s grid snap size = %f", name, size)
	}
	if !onGrid(position.X, size) || !onGrid(position.Y, size) {
		return fmt.Errorf("%s = %f,%f is not constrained to grid size %f", name, position.X, position.Y, size)
	}
	return nil
}

func onGrid(value float64, size float64) bool {
	return math.Abs(value/size-math.Round(value/size)) < 0.000001
}
