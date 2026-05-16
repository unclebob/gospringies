package acceptance

import (
	"fmt"

	"springs/internal/edit"
	"springs/internal/sim"
)

func createSelectableObject(w *world, example map[string]string) error {
	return withSelectionObject(example, func(objectType string, id int) error { return addSelectionObject(w, objectType, id) })
}

func selectObject(w *world, example map[string]string) error {
	return withSelectionObject(example, func(objectType string, id int) error { return selectObjectByType(w, objectType, id) })
}

func assertObjectSelected(w *world, example map[string]string) error {
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	if !objectSelected(w, objectType, id) {
		return fmt.Errorf("%s %d not selected", objectType, id)
	}
	return nil
}

func createSelectionWorld(w *world, _ map[string]string) error {
	return addSelectionPair(w, 1, 2, 3)
}

func selectAllObjects(w *world, _ map[string]string) error {
	return updateSelectionEditor(w, (*edit.Editor).SelectAll)
}

func assertEveryMassSelected(w *world, _ map[string]string) error {
	return assertAllSelected("mass", massSelectionIDs(w), ensureMouseEditor(w).MassSelected)
}

func assertEverySpringSelected(w *world, _ map[string]string) error {
	return assertAllSelected("spring", springSelectionIDs(w), ensureMouseEditor(w).SpringSelected)
}

func deleteSelectedObjects(w *world, _ map[string]string) error {
	return updateSelectionEditor(w, (*edit.Editor).DeleteSelected)
}

func assertObjectDeleted(w *world, example map[string]string) error {
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	if objectExists(w, objectType, id) {
		return fmt.Errorf("%s %d still exists", objectType, id)
	}
	return nil
}

func createSelectionConnectedMasses(w *world, _ map[string]string) error {
	return createSelectionWorld(w, nil)
}

func selectMassOne(w *world, _ map[string]string) error {
	return ensureMouseEditor(w).SelectMass(1)
}

func assertMassOneDeleted(w *world, _ map[string]string) error {
	return assertSelectionMassExists(w, 1, false)
}

func assertSpringThreeDeleted(w *world, _ map[string]string) error {
	return assertSelectionSpringExists(w, 3, false)
}

func assertMassTwoExists(w *world, _ map[string]string) error {
	return assertSelectionMassExists(w, 2, true)
}

func createSelectedObjectSet(w *world, example map[string]string) error {
	objectSet, err := stringValue(example, "object_set")
	if err != nil {
		return err
	}
	return createNamedSelectedObjectSet(w, objectSet)
}

func duplicateSelectedObjects(w *world, _ map[string]string) error {
	duplicated, err := ensureMouseEditor(w).DuplicateSelected()
	w.duplicated = duplicated
	return err
}

func assertDuplicatedUniqueIDs(w *world, _ map[string]string) error {
	for _, ids := range []struct {
		name      string
		duplicate []int
		original  []int
	}{
		{"mass", w.duplicated.MassIDs, w.originalMassIDs},
		{"spring", w.duplicated.SpringIDs, w.originalSpringIDs},
	} {
		if err := assertUniqueNewIDs(ids.name, ids.duplicate, ids.original); err != nil {
			return err
		}
	}
	return nil
}

func assertDuplicatedIndependent(w *world, _ map[string]string) error {
	if err := assertDuplicatedMassesIndependent(w); err != nil {
		return err
	}
	return assertDuplicatedSpringsIndependent(w)
}

func updateSelectionEditor(w *world, update func(*edit.Editor)) error {
	update(ensureMouseEditor(w))
	return nil
}

func withSelectionObject(example map[string]string, action func(string, int) error) error {
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	return action(objectType, id)
}

func assertSelectedID(objectType string, id int, selected func(int) bool) error {
	if !selected(id) {
		return fmt.Errorf("%s %d not selected", objectType, id)
	}
	return nil
}

func assertAllSelected(objectType string, ids []int, selected func(int) bool) error {
	for _, id := range ids {
		if err := assertSelectedID(objectType, id, selected); err != nil {
			return err
		}
	}
	return nil
}

func massSelectionIDs(w *world) []int {
	return selectionIDs(ensureDomainWorld(w).Masses, func(mass sim.Mass) int { return mass.ID })
}

func springSelectionIDs(w *world) []int {
	return selectionIDs(ensureDomainWorld(w).Springs, func(spring sim.Spring) int { return spring.ID })
}

func selectionIDs[T any](items []T, itemID func(T) int) []int {
	ids := make([]int, 0, len(items))
	for _, item := range items {
		ids = append(ids, itemID(item))
	}
	return ids
}

func addSelectionObject(w *world, objectType string, id int) error {
	switch objectType {
	case "mass":
		return ensureDomainWorld(w).AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: float64(id), Y: 1}, Mass: 1})
	case "spring":
		return addSelectionPair(w, 1, 2, id)
	default:
		return fmt.Errorf("unsupported object type %q", objectType)
	}
}

func selectObjectByType(w *world, objectType string, id int) error {
	switch objectType {
	case "mass":
		return ensureMouseEditor(w).SelectMass(id)
	case "spring":
		return ensureMouseEditor(w).SelectSpring(id)
	default:
		return fmt.Errorf("unsupported object type %q", objectType)
	}
}

func objectSelected(w *world, objectType string, id int) bool {
	switch objectType {
	case "mass":
		return ensureMouseEditor(w).MassSelected(id)
	case "spring":
		return ensureMouseEditor(w).SpringSelected(id)
	default:
		return false
	}
}

func objectExists(w *world, objectType string, id int) bool {
	switch objectType {
	case "mass":
		_, ok := ensureDomainWorld(w).MassByID(id)
		return ok
	case "spring":
		_, ok := ensureDomainWorld(w).SpringByID(id)
		return ok
	default:
		return false
	}
}

func addSelectionPair(w *world, massA, massB, springID int) error {
	world := ensureDomainWorld(w)
	if err := ensureSelectionMass(world, massA); err != nil {
		return err
	}
	if err := ensureSelectionMass(world, massB); err != nil {
		return err
	}
	return world.AddSpring(sim.Spring{ID: springID, MassA: massA, MassB: massB, RestLength: 20, SpringConstant: 8})
}

func ensureSelectionMass(world *sim.Simulation, id int) error {
	if _, ok := world.MassByID(id); ok {
		return nil
	}
	return world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: float64(id * 10), Y: 20}, Mass: 1})
}

func assertSelectionMassExists(w *world, id int, expected bool) error {
	return assertSelectionObjectExists(w, "mass", id, expected)
}

func assertSelectionSpringExists(w *world, id int, expected bool) error {
	return assertSelectionObjectExists(w, "spring", id, expected)
}

func assertSelectionObjectExists(w *world, objectType string, id int, expected bool) error {
	return assertSelectionExists(objectType, id, objectExists(w, objectType, id), expected)
}

func assertSelectionExists(objectType string, id int, exists bool, expected bool) error {
	if exists != expected {
		return fmt.Errorf("%s %d exists = %t", objectType, id, exists)
	}
	return nil
}

func createNamedSelectedObjectSet(w *world, objectSet string) error {
	switch objectSet {
	case "one mass":
		return createSelectedMassSet(w)
	case "two masses and a spring":
		return createSelectedSpringSet(w)
	default:
		return fmt.Errorf("unsupported object set %q", objectSet)
	}
}

func createSelectedMassSet(w *world) error {
	if err := addSelectionObject(w, "mass", 1); err != nil {
		return err
	}
	w.originalMassIDs = []int{1}
	return ensureMouseEditor(w).SelectMass(1)
}

func createSelectedSpringSet(w *world) error {
	if err := addSelectionPair(w, 1, 2, 3); err != nil {
		return err
	}
	w.originalMassIDs = []int{1, 2}
	w.originalSpringIDs = []int{3}
	ensureMouseEditor(w).SelectAll()
	return nil
}

func repeatedID(ids []int) bool {
	seen := map[int]bool{}
	for _, id := range ids {
		if seen[id] {
			return true
		}
		seen[id] = true
	}
	return false
}

func assertUniqueNewIDs(objectType string, duplicate []int, original []int) error {
	if repeatedID(duplicate) {
		return fmt.Errorf("duplicated %s ids are not unique: %v", objectType, duplicate)
	}
	if anySharedID(duplicate, original) {
		return fmt.Errorf("duplicated %s ids overlap originals: %v", objectType, duplicate)
	}
	return nil
}

func anySharedID(first []int, second []int) bool {
	seen := idSet(second)
	for _, id := range first {
		if seen[id] {
			return true
		}
	}
	return false
}

func idSet(ids []int) map[int]bool {
	seen := map[int]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	return seen
}

func assertDuplicatedMassesIndependent(w *world) error {
	for i, duplicateID := range w.duplicated.MassIDs {
		duplicate, ok := ensureDomainWorld(w).MassByID(duplicateID)
		if !ok {
			return fmt.Errorf("duplicate mass %d missing", duplicateID)
		}
		originalID := w.originalMassIDs[i]
		if err := ensureMouseEditor(w).DragMass(originalID, sim.Vec2{X: 99, Y: 99}); err != nil {
			return err
		}
		after, _ := ensureDomainWorld(w).MassByID(duplicateID)
		if after.Position != duplicate.Position {
			return fmt.Errorf("duplicate mass %d changed with original", duplicateID)
		}
	}
	return nil
}

func assertDuplicatedSpringsIndependent(w *world) error {
	originalMasses := idSet(w.originalMassIDs)
	for _, duplicateID := range w.duplicated.SpringIDs {
		duplicate, ok := ensureDomainWorld(w).SpringByID(duplicateID)
		if !ok {
			return fmt.Errorf("duplicate spring %d missing", duplicateID)
		}
		if originalMasses[duplicate.MassA] || originalMasses[duplicate.MassB] {
			return fmt.Errorf("duplicate spring %d still uses original endpoints", duplicateID)
		}
	}
	return nil
}
