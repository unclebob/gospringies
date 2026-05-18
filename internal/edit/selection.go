package edit

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

type DuplicatedObjects struct {
	MassIDs   []int
	SpringIDs []int
}

func (e *Editor) SelectMass(id int) error {
	return e.selectExisting(id, "mass", e.massExists, func() { e.SelectedMasses[id] = true })
}

func (e *Editor) AddMassSelection(id int) error {
	return e.selectExisting(id, "mass", e.massExists, func() { e.SelectedMasses[id] = true }, keepSelection)
}

func (e *Editor) SelectSpring(id int) error {
	return e.selectExisting(id, "spring", e.springExists, func() { e.SelectedSprings[id] = true })
}

func (e *Editor) SelectNearest(position sim.Vec2, toggle bool) error {
	id, ok := e.nearestMassID(position)
	if !ok {
		return fmt.Errorf("no object near pointer")
	}
	if toggle {
		e.toggleMassSelection(id)
		return nil
	}
	e.clearSelection()
	e.SelectedMasses[id] = true
	return nil
}

func (e *Editor) selectExisting(id int, objectType string, exists func(int) bool, selectObject func(), options ...func(*Editor)) error {
	if !exists(id) {
		return fmt.Errorf("%s %d not found", objectType, id)
	}
	if len(options) == 0 {
		e.clearSelection()
	}
	for _, option := range options {
		option(e)
	}
	selectObject()
	return nil
}

func keepSelection(*Editor) {}

func (e *Editor) SelectAll() {
	e.clearSelection()
	for _, mass := range e.World.Masses {
		e.SelectedMasses[mass.ID] = true
	}
	for _, spring := range e.World.Springs {
		e.SelectedSprings[spring.ID] = true
	}
}

func (e *Editor) BoxSelect(min sim.Vec2, max sim.Vec2, add bool) {
	if !add {
		e.clearSelection()
	}
	massesInBox := 0
	for _, mass := range e.World.Masses {
		if withinBox(mass.Position, min, max) {
			e.SelectedMasses[mass.ID] = true
			massesInBox++
		}
	}
	fullyEnclosedSprings := e.selectFullyEnclosedSprings(min, max)
	if massesInBox == 0 && fullyEnclosedSprings == 0 {
		e.selectSinglePartiallyEnclosedSpring(min, max)
	}
}

func (e *Editor) selectFullyEnclosedSprings(min sim.Vec2, max sim.Vec2) int {
	count := 0
	for _, spring := range e.World.Springs {
		a, okA := e.World.MassByID(spring.MassA)
		b, okB := e.World.MassByID(spring.MassB)
		if okA && okB && withinBox(a.Position, min, max) && withinBox(b.Position, min, max) {
			e.SelectedSprings[spring.ID] = true
			count++
		}
	}
	return count
}

func (e *Editor) selectSinglePartiallyEnclosedSpring(min sim.Vec2, max sim.Vec2) {
	selectedID := 0
	for _, spring := range e.World.Springs {
		a, okA := e.World.MassByID(spring.MassA)
		b, okB := e.World.MassByID(spring.MassB)
		if !okA || !okB || !segmentIntersectsBox(a.Position, b.Position, min, max) {
			continue
		}
		if selectedID != 0 {
			return
		}
		selectedID = spring.ID
	}
	if selectedID != 0 {
		e.SelectedSprings[selectedID] = true
	}
}

func (e *Editor) MoveSelected(delta sim.Vec2) {
	for i := range e.World.Masses {
		if e.SelectedMasses[e.World.Masses[i].ID] && !e.World.Masses[i].Fixed {
			e.World.Masses[i].Position = e.World.Masses[i].Position.Add(delta)
		}
	}
}

func (e *Editor) ThrowSelected(velocity sim.Vec2) {
	for i := range e.World.Masses {
		if e.SelectedMasses[e.World.Masses[i].ID] && !e.World.Masses[i].Fixed {
			e.World.Masses[i].Velocity = velocity
		}
	}
}

func (e *Editor) MassSelected(id int) bool {
	return e.SelectedMasses[id]
}

func (e *Editor) SpringSelected(id int) bool {
	return e.SelectedSprings[id]
}

func (e *Editor) DeleteSelected() {
	e.deleteSelectedMasses()
	e.deleteSelectedSprings()
	e.reindexSprings()
	e.clearSelection()
}

func (e *Editor) DuplicateSelected() (DuplicatedObjects, error) {
	duplicated := DuplicatedObjects{}
	massIDs := e.duplicateMasses(&duplicated)
	if err := e.duplicateSprings(massIDs, &duplicated); err != nil {
		return DuplicatedObjects{}, err
	}
	e.clearSelection()
	for _, id := range duplicated.MassIDs {
		e.SelectedMasses[id] = true
	}
	for _, id := range duplicated.SpringIDs {
		e.SelectedSprings[id] = true
	}
	return duplicated, nil
}

func (e *Editor) clearSelection() {
	e.SelectedMasses = map[int]bool{}
	e.SelectedSprings = map[int]bool{}
}

func (e *Editor) ClearSelection() {
	e.clearSelection()
}

func (e *Editor) toggleMassSelection(id int) {
	if e.SelectedMasses[id] {
		delete(e.SelectedMasses, id)
		return
	}
	e.SelectedMasses[id] = true
}

func (e *Editor) deleteSelectedMasses() {
	masses := e.World.Masses[:0]
	for _, mass := range e.World.Masses {
		if !e.SelectedMasses[mass.ID] {
			masses = append(masses, mass)
		}
	}
	e.World.Masses = masses
}

func (e *Editor) deleteSelectedSprings() {
	springs := e.World.Springs[:0]
	for _, spring := range e.World.Springs {
		if e.keepSpring(spring) {
			springs = append(springs, spring)
		}
	}
	e.World.Springs = springs
}

func (e *Editor) keepSpring(spring sim.Spring) bool {
	return !e.SelectedSprings[spring.ID] && !e.SelectedMasses[spring.MassA] && !e.SelectedMasses[spring.MassB]
}

func (e *Editor) duplicateMasses(duplicated *DuplicatedObjects) map[int]int {
	next := nextMassID(e.World)
	massIDs := map[int]int{}
	for _, mass := range e.World.Masses {
		if !e.SelectedMasses[mass.ID] {
			continue
		}
		originalID := mass.ID
		mass.ID = next
		next++
		e.World.Masses = append(e.World.Masses, mass)
		massIDs[originalID] = mass.ID
		duplicated.MassIDs = append(duplicated.MassIDs, mass.ID)
	}
	return massIDs
}

func (e *Editor) duplicateSprings(massIDs map[int]int, duplicated *DuplicatedObjects) error {
	next := nextSpringID(e.World)
	for _, spring := range e.World.Springs {
		if !e.SelectedSprings[spring.ID] {
			continue
		}
		spring.ID = next
		next++
		spring.MassA = replacementID(massIDs, spring.MassA)
		spring.MassB = replacementID(massIDs, spring.MassB)
		if err := e.World.AddSpring(spring); err != nil {
			return err
		}
		duplicated.SpringIDs = append(duplicated.SpringIDs, spring.ID)
	}
	return nil
}

func (e *Editor) reindexSprings() {
	for i := range e.World.Springs {
		a, okA := e.worldIndexByMassID(e.World.Springs[i].MassA)
		b, okB := e.worldIndexByMassID(e.World.Springs[i].MassB)
		if okA && okB {
			e.World.Springs[i].A = a
			e.World.Springs[i].B = b
		}
	}
}

func (e *Editor) massExists(id int) bool {
	return objectExists(func() (sim.Mass, bool) { return e.World.MassByID(id) })
}

func (e *Editor) springExists(id int) bool {
	return objectExists(func() (sim.Spring, bool) { return e.World.SpringByID(id) })
}

func objectExists[T any](lookup func() (T, bool)) bool {
	_, ok := lookup()
	return ok
}

func (e *Editor) worldIndexByMassID(id int) (int, bool) {
	for i, mass := range e.World.Masses {
		if mass.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (e *Editor) nearestMassID(position sim.Vec2) (int, bool) {
	if len(e.World.Masses) == 0 {
		return 0, false
	}
	nearestID := e.World.Masses[0].ID
	nearestDistance := math.MaxFloat64
	for _, mass := range e.World.Masses {
		if d := distance(mass.Position, position); d < nearestDistance {
			nearestID = mass.ID
			nearestDistance = d
		}
	}
	return nearestID, true
}

func replacementID(ids map[int]int, id int) int {
	if replacement, ok := ids[id]; ok {
		return replacement
	}
	return id
}

func withinBox(position sim.Vec2, min sim.Vec2, max sim.Vec2) bool {
	lowX, highX := ordered(min.X, max.X)
	lowY, highY := ordered(min.Y, max.Y)
	return position.X >= lowX && position.X <= highX && position.Y >= lowY && position.Y <= highY
}

func segmentIntersectsBox(a sim.Vec2, b sim.Vec2, min sim.Vec2, max sim.Vec2) bool {
	if withinBox(a, min, max) || withinBox(b, min, max) {
		return true
	}
	lowX, highX := ordered(min.X, max.X)
	lowY, highY := ordered(min.Y, max.Y)
	corners := []sim.Vec2{
		{X: lowX, Y: lowY},
		{X: highX, Y: lowY},
		{X: highX, Y: highY},
		{X: lowX, Y: highY},
	}
	for i := range corners {
		if segmentsIntersect(a, b, corners[i], corners[(i+1)%len(corners)]) {
			return true
		}
	}
	return false
}

func segmentsIntersect(a sim.Vec2, b sim.Vec2, c sim.Vec2, d sim.Vec2) bool {
	o1 := orientation(a, b, c)
	o2 := orientation(a, b, d)
	o3 := orientation(c, d, a)
	o4 := orientation(c, d, b)
	if o1 == 0 && onSegment(a, c, b) {
		return true
	}
	if o2 == 0 && onSegment(a, d, b) {
		return true
	}
	if o3 == 0 && onSegment(c, a, d) {
		return true
	}
	if o4 == 0 && onSegment(c, b, d) {
		return true
	}
	return (o1 > 0) != (o2 > 0) && (o3 > 0) != (o4 > 0)
}

func orientation(a sim.Vec2, b sim.Vec2, c sim.Vec2) float64 {
	value := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	if math.Abs(value) < 1e-9 {
		return 0
	}
	return value
}

func onSegment(a sim.Vec2, b sim.Vec2, c sim.Vec2) bool {
	lowX, highX := ordered(a.X, c.X)
	lowY, highY := ordered(a.Y, c.Y)
	return b.X >= lowX && b.X <= highX && b.Y >= lowY && b.Y <= highY
}

func ordered(a float64, b float64) (float64, float64) {
	return math.Min(a, b), math.Max(a, b)
}
