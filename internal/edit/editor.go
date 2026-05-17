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
	ModeEdit      = "edit"
)

type Editor struct {
	World           *sim.Simulation
	Mode            string
	GridSnapEnabled bool
	GridSnapSize    float64
	SelectedMasses  map[int]bool
	SelectedSprings map[int]bool
	pendingSpring   *PendingSpring
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
		Fixed:      parameterBool(e.World.Parameters, "fixed mass"),
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

func parameterFloat(parameters sim.Parameters, name string) float64 {
	value, _ := strconv.ParseFloat(parameters.Value(name), 64)
	return value
}

func parameterBool(parameters sim.Parameters, name string) bool {
	value, _ := strconv.ParseBool(parameters.Value(name))
	return value
}
