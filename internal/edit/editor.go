package edit

import (
	"fmt"
	"math"
	"strconv"

	"springs/internal/sim"
)

const (
	ModeAddMass   = "add mass"
	ModeAddSpring = "add spring"
)

type Editor struct {
	World           *sim.Simulation
	Mode            string
	GridSnapEnabled bool
	GridSnapSize    float64
	SelectedMasses  map[int]bool
	SelectedSprings map[int]bool
}

func NewEditor(world *sim.Simulation) *Editor {
	return &Editor{World: world, SelectedMasses: map[int]bool{}, SelectedSprings: map[int]bool{}}
}

func (e *Editor) Click(position sim.Vec2) (int, error) {
	if e.Mode != ModeAddMass {
		return 0, fmt.Errorf("unsupported click mode %q", e.Mode)
	}
	mass := sim.Mass{
		ID:         nextMassID(e.World),
		Position:   e.snap(position),
		Mass:       parameterFloat(e.World.Parameters, "current mass"),
		Elasticity: parameterFloat(e.World.Parameters, "elasticity"),
	}
	return mass.ID, e.World.AddMass(mass)
}

func (e *Editor) CreateSpring(massA int, massB int) (int, error) {
	if e.Mode != ModeAddSpring {
		return 0, fmt.Errorf("unsupported spring mode %q", e.Mode)
	}
	a, okA := e.World.MassByID(massA)
	b, okB := e.World.MassByID(massB)
	if !okA || !okB {
		return 0, fmt.Errorf("missing spring endpoint")
	}
	spring := sim.Spring{
		ID:             nextSpringID(e.World),
		MassA:          massA,
		MassB:          massB,
		RestLength:     distance(a.Position, b.Position),
		SpringConstant: parameterFloat(e.World.Parameters, "spring constant"),
		Damping:        parameterFloat(e.World.Parameters, "damping"),
	}
	return spring.ID, e.World.AddSpring(spring)
}

func (e *Editor) DragMass(id int, position sim.Vec2) error {
	for i := range e.World.Masses {
		if e.World.Masses[i].ID == id {
			if !e.World.Masses[i].Fixed {
				e.World.Masses[i].Position = e.snap(position)
			}
			return nil
		}
	}
	return fmt.Errorf("mass %d not found", id)
}

type DuplicatedObjects struct {
	MassIDs   []int
	SpringIDs []int
}

func (e *Editor) SelectMass(id int) error {
	return e.selectExisting(id, "mass", e.massExists, func() { e.SelectedMasses[id] = true })
}

func (e *Editor) SelectSpring(id int) error {
	return e.selectExisting(id, "spring", e.springExists, func() { e.SelectedSprings[id] = true })
}

func (e *Editor) selectExisting(id int, objectType string, exists func(int) bool, selectObject func()) error {
	if !exists(id) {
		return fmt.Errorf("%s %d not found", objectType, id)
	}
	e.clearSelection()
	selectObject()
	return nil
}

func (e *Editor) SelectAll() {
	e.clearSelection()
	for _, mass := range e.World.Masses {
		e.SelectedMasses[mass.ID] = true
	}
	for _, spring := range e.World.Springs {
		e.SelectedSprings[spring.ID] = true
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

func (e *Editor) snap(position sim.Vec2) sim.Vec2 {
	if !e.GridSnapEnabled || e.GridSnapSize <= 0 {
		return position
	}
	return sim.Vec2{
		X: math.Round(position.X/e.GridSnapSize) * e.GridSnapSize,
		Y: math.Round(position.Y/e.GridSnapSize) * e.GridSnapSize,
	}
}

func nextMassID(world *sim.Simulation) int {
	return nextID(world.Masses, func(mass sim.Mass) int { return mass.ID })
}

func nextSpringID(world *sim.Simulation) int {
	return nextID(world.Springs, func(spring sim.Spring) int { return spring.ID })
}

func nextID[T any](items []T, itemID func(T) int) int {
	next := 1
	for _, item := range items {
		id := itemID(item)
		if id >= next {
			next = id + 1
		}
	}
	return next
}

func distance(a sim.Vec2, b sim.Vec2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func replacementID(ids map[int]int, id int) int {
	if replacement, ok := ids[id]; ok {
		return replacement
	}
	return id
}

func parameterFloat(parameters sim.Parameters, name string) float64 {
	value, _ := strconv.ParseFloat(parameters.Value(name), 64)
	return value
}
