//go:build property

package edit

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"testing/quick"

	"springs/internal/sim"
)

func TestPropertyEditorSnapIDsAndDistance(t *testing.T) {
	checkProperty(t, 1, 300, editorSnapIDsAndDistance)
}

func TestPropertyBoxSelectIsInvariantUnderBoxCornerOrder(t *testing.T) {
	checkProperty(t, 2, 300, boxSelectIsInvariantUnderBoxCornerOrder)
}

func TestPropertyBoxSelectChoosesSinglePartialSpringOnlyWhenNothingElseIsEnclosed(t *testing.T) {
	checkProperty(t, 3, 300, boxSelectChoosesSinglePartialSpringOnlyWhenNothingElseIsEnclosed)
}

func TestPropertyMoveAndThrowSelectedAffectOnlySelectedMovableMasses(t *testing.T) {
	checkProperty(t, 4, 300, moveAndThrowSelectedAffectOnlySelectedMovableMasses)
}

func TestPropertyDeleteSelectedRemovesAttachedSpringsAndReindexes(t *testing.T) {
	checkProperty(t, 5, 300, deleteSelectedRemovesAttachedSpringsAndReindexes)
}

func TestPropertyDuplicateSelectedCreatesUniqueValidIDs(t *testing.T) {
	checkProperty(t, 6, 300, duplicateSelectedCreatesUniqueValidIDs)
}

func TestPropertySelectionGeometryIsOrderAndTranslationInvariant(t *testing.T) {
	checkProperty(t, 7, 500, selectionGeometryIsOrderAndTranslationInvariant)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{
		MaxCount: maxCount,
		Rand:     rand.New(rand.NewSource(seed)),
	}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func editorSnapIDsAndDistance(xInput, yInput, snapInput float64) bool {
	position := propertyVec(xInput, yInput, 100)
	snapSize := propertyFloat(snapInput, 0.1, 25)
	editor := NewEditor(sim.NewWorld())
	if editor.snap(position) != position {
		panic("disabled snap changed position")
	}
	editor.GridSnapEnabled = true
	editor.GridSnapSize = snapSize
	snapped := editor.snap(position)
	assertClose("snapped x grid", snapped.X/snapSize, math.Round(snapped.X/snapSize), 1e-9)
	assertClose("snapped y grid", snapped.Y/snapSize, math.Round(snapped.Y/snapSize), 1e-9)

	_ = editor.World.AddMass(sim.Mass{ID: 10, Position: position, Mass: 1})
	_ = editor.World.AddMass(sim.Mass{ID: 3, Position: snapped, Mass: 1})
	_ = editor.World.AddSpring(sim.Spring{ID: 7, MassA: 10, MassB: 3})
	if nextMassID(editor.World) != 11 || nextSpringID(editor.World) != 8 {
		panic(fmt.Sprintf("next IDs wrong: mass=%d spring=%d", nextMassID(editor.World), nextSpringID(editor.World)))
	}
	if nextID([]sim.Mass{}, func(mass sim.Mass) int { return mass.ID }) != 1 {
		panic("nextID for empty slice should be 1")
	}
	assertClose("distance symmetric", distance(position, snapped), distance(snapped, position), 0)
	if distance(position, position) != 0 || distance(position, snapped) < 0 {
		panic("distance invariant failed")
	}
	return true
}

func boxSelectIsInvariantUnderBoxCornerOrder(minXInput, minYInput, maxXInput, maxYInput float64) bool {
	min := sim.Vec2{X: propertySignedFloat(minXInput, 50), Y: propertySignedFloat(minYInput, 50)}
	max := sim.Vec2{X: min.X + propertyFloat(maxXInput, 1, 50), Y: min.Y + propertyFloat(maxYInput, 1, 50)}
	world := selectionWorld(min, max)
	forward := NewEditor(world.Clone())
	reversed := NewEditor(world.Clone())
	forward.BoxSelect(min, max, false)
	reversed.BoxSelect(max, min, false)
	assertSameSelection("mass selection", forward.SelectedMasses, reversed.SelectedMasses)
	assertSameSelection("spring selection", forward.SelectedSprings, reversed.SelectedSprings)
	if !forward.SelectedMasses[1] || !forward.SelectedMasses[2] || !forward.SelectedSprings[1] {
		panic(fmt.Sprintf("expected enclosed masses and spring selected: masses=%#v springs=%#v", forward.SelectedMasses, forward.SelectedSprings))
	}
	return true
}

func boxSelectChoosesSinglePartialSpringOnlyWhenNothingElseIsEnclosed(xInput, yInput float64) bool {
	offset := sim.Vec2{X: propertySignedFloat(xInput, 20), Y: propertySignedFloat(yInput, 20)}
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: offset.Add(sim.Vec2{X: -10, Y: 5}), Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: offset.Add(sim.Vec2{X: 10, Y: 5}), Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	editor := NewEditor(world)
	editor.BoxSelect(offset.Add(sim.Vec2{X: -1, Y: 0}), offset.Add(sim.Vec2{X: 1, Y: 10}), false)
	if len(editor.SelectedMasses) != 0 || !editor.SelectedSprings[1] {
		panic(fmt.Sprintf("single partial spring not selected as expected: masses=%#v springs=%#v", editor.SelectedMasses, editor.SelectedSprings))
	}
	return true
}

func moveAndThrowSelectedAffectOnlySelectedMovableMasses(dxInput, dyInput, vxInput, vyInput float64) bool {
	delta := propertyVec(dxInput, dyInput, 25)
	velocity := propertyVec(vxInput, vyInput, 25)
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 1, Y: 2}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 3, Y: 4}, Velocity: sim.Vec2{X: 9}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 5, Y: 6}, Velocity: sim.Vec2{Y: 9}, Mass: 1, Fixed: true})
	editor := NewEditor(world)
	editor.SelectedMasses[1] = true
	editor.SelectedMasses[3] = true
	startUnselected := world.Masses[1]
	startFixed := world.Masses[2]
	editor.MoveSelected(delta)
	editor.ThrowSelected(velocity)
	assertVecClose("selected movable position", world.Masses[0].Position, (sim.Vec2{X: 1, Y: 2}).Add(delta), 0)
	assertVecClose("selected movable velocity", world.Masses[0].Velocity, velocity, 0)
	assertMassUnchanged("unselected mass", world.Masses[1], startUnselected)
	assertMassUnchanged("fixed selected mass", world.Masses[2], startFixed)
	return true
}

func deleteSelectedRemovesAttachedSpringsAndReindexes(selectInput float64) bool {
	world := connectedWorld()
	editor := NewEditor(world)
	editor.SelectedMasses[2] = true
	selectedSpring := false
	if propertyFloat(selectInput, 0, 1) > 0.5 {
		editor.SelectedSprings[3] = true
		selectedSpring = true
	}
	editor.DeleteSelected()
	if _, ok := world.MassByID(2); ok {
		panic("selected mass survived delete")
	}
	for _, spring := range world.Springs {
		if spring.MassA == 2 || spring.MassB == 2 || (selectedSpring && spring.ID == 3) {
			panic(fmt.Sprintf("deleted/attached spring survived: %#v", spring))
		}
		if spring.A < 0 || spring.A >= len(world.Masses) || spring.B < 0 || spring.B >= len(world.Masses) {
			panic(fmt.Sprintf("spring not reindexed: %#v masses=%#v", spring, world.Masses))
		}
	}
	if len(editor.SelectedMasses) != 0 || len(editor.SelectedSprings) != 0 {
		panic("selection not cleared after delete")
	}
	return true
}

func duplicateSelectedCreatesUniqueValidIDs(xInput, yInput float64) bool {
	world := connectedWorld()
	editor := NewEditor(world)
	editor.SelectedMasses[1] = true
	editor.SelectedMasses[2] = true
	editor.SelectedSprings[1] = true
	duplicated, err := editor.DuplicateSelected()
	if err != nil {
		panic(err)
	}
	if len(duplicated.MassIDs) != 2 || len(duplicated.SpringIDs) != 1 {
		panic(fmt.Sprintf("unexpected duplicated objects: %#v", duplicated))
	}
	seenMassIDs := map[int]bool{}
	for _, mass := range world.Masses {
		if seenMassIDs[mass.ID] {
			panic(fmt.Sprintf("duplicate mass ID after duplicate: %#v", world.Masses))
		}
		seenMassIDs[mass.ID] = true
	}
	for _, spring := range world.Springs {
		if _, ok := world.MassByID(spring.MassA); !ok {
			panic(fmt.Sprintf("spring MassA missing after duplicate: %#v", spring))
		}
		if _, ok := world.MassByID(spring.MassB); !ok {
			panic(fmt.Sprintf("spring MassB missing after duplicate: %#v", spring))
		}
	}
	for _, id := range duplicated.MassIDs {
		if !editor.SelectedMasses[id] {
			panic(fmt.Sprintf("duplicated mass not selected: %d selection=%#v", id, editor.SelectedMasses))
		}
	}
	for _, id := range duplicated.SpringIDs {
		if !editor.SelectedSprings[id] {
			panic(fmt.Sprintf("duplicated spring not selected: %d selection=%#v", id, editor.SelectedSprings))
		}
	}
	_ = xInput
	_ = yInput
	return true
}

func selectionGeometryIsOrderAndTranslationInvariant(axInput, ayInput, bxInput, byInput, cxInput, cyInput, dxInput, dyInput, txInput, tyInput float64) bool {
	a := propertyVec(axInput, ayInput, 100)
	b := propertyVec(bxInput, byInput, 100)
	c := propertyVec(cxInput, cyInput, 100)
	d := propertyVec(dxInput, dyInput, 100)
	translation := propertyVec(txInput, tyInput, 100)
	min := sim.Vec2{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y)}
	max := sim.Vec2{X: math.Max(a.X, b.X) + 1, Y: math.Max(a.Y, b.Y) + 1}
	point := sim.Vec2{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}

	if !withinBox(point, min, max) || !withinBox(point, max, min) {
		panic("withinBox is not invariant under corner order")
	}
	if segmentFullyWithinBox(point, point, min, max) != segmentFullyWithinBox(point.Add(translation), point.Add(translation), min.Add(translation), max.Add(translation)) {
		panic("segmentFullyWithinBox is not translation invariant")
	}
	intersects := segmentsIntersect(a, b, c, d)
	if intersects != segmentsIntersect(b, a, c, d) || intersects != segmentsIntersect(a, b, d, c) {
		panic("segmentsIntersect is not endpoint-order invariant")
	}
	if intersects != segmentsIntersect(a.Add(translation), b.Add(translation), c.Add(translation), d.Add(translation)) {
		panic("segmentsIntersect is not translation invariant")
	}
	if segmentIntersectsBox(a, b, min, max) != segmentIntersectsBox(a.Add(translation), b.Add(translation), min.Add(translation), max.Add(translation)) {
		panic("segmentIntersectsBox is not translation invariant")
	}
	o := orientation(a, b, c)
	if orientation(a.Add(translation), b.Add(translation), c.Add(translation)) != o {
		panic("orientation is not translation invariant")
	}
	if oppositeSides(1, -1) != true || oppositeSides(1, 0) != false {
		panic("oppositeSides invariant failed")
	}
	if hasCollinearEndpoint(a, b, a, d, 0, orientation(a, b, d), 0, orientation(a, d, b)) != collinearEndpointOnSegment(a, b, a, 0) {
		panic("collinear endpoint helpers disagree")
	}
	low, high := ordered(a.X, b.X)
	if low > high || !between((low+high)/2, low, high) {
		panic("ordered/between invariant failed")
	}
	if onSegment(a, a, b) != true || onSegment(a, b, b) != true {
		panic("segment endpoints should be on segment")
	}
	return true
}

func selectionWorld(min sim.Vec2, max sim.Vec2) *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: min.X, Y: min.Y}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: max.X + 20, Y: max.Y + 20}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	_ = world.AddSpring(sim.Spring{ID: 2, MassA: 2, MassB: 3})
	return world
}

func connectedWorld() *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 10, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 20, Y: 0}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	_ = world.AddSpring(sim.Spring{ID: 2, MassA: 2, MassB: 3})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 3})
	return world
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	value := math.Abs(input)
	return minimum + math.Mod(value, maximum-minimum)
}

func propertySignedFloat(input float64, magnitude float64) float64 {
	return propertyFloat(input, 0, magnitude*2) - magnitude
}

func propertyVec(xInput, yInput float64, magnitude float64) sim.Vec2 {
	return sim.Vec2{X: propertySignedFloat(xInput, magnitude), Y: propertySignedFloat(yInput, magnitude)}
}

func assertSameSelection(label string, actual, expected map[int]bool) {
	if len(actual) != len(expected) {
		panic(fmt.Sprintf("%s length differs: %#v != %#v", label, actual, expected))
	}
	for key, actualValue := range actual {
		if expected[key] != actualValue {
			panic(fmt.Sprintf("%s differs: %#v != %#v", label, actual, expected))
		}
	}
}

func assertMassUnchanged(label string, actual, expected sim.Mass) {
	assertVecClose(label+" position", actual.Position, expected.Position, 0)
	assertVecClose(label+" velocity", actual.Velocity, expected.Velocity, 0)
}

func assertVecClose(label string, actual, expected sim.Vec2, tolerance float64) {
	assertClose(label+" x", actual.X, expected.X, tolerance)
	assertClose(label+" y", actual.Y, expected.Y, tolerance)
}

func assertClose(label string, actual, expected, tolerance float64) {
	if math.Abs(actual-expected) > tolerance {
		panic(fmt.Sprintf("%s: got %f, want %f +/- %f", label, actual, expected, tolerance))
	}
}
