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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T10:00:37-05:00","module_hash":"4c58a62ae1e17a1740c8d96114cd5a250b480880186773b5c7c1e0090338c9b4","functions":[{"id":"func/NewEditor","name":"NewEditor","line":27,"end_line":29,"hash":"b7f6863b64660fc5c51dba39f64a9219f20633153ac389d286f5a66227cac195"},{"id":"func/Editor.Click","name":"Editor.Click","line":31,"end_line":43,"hash":"2a8b140bca07fca6200481a5da808a469560281c16ee3bdc8cb242613f794428"},{"id":"func/Editor.CreateSpring","name":"Editor.CreateSpring","line":45,"end_line":63,"hash":"716c54198a0039e44328bfb91fcf99342df73110182a8c4ade49e5bdf152376e"},{"id":"func/Editor.DragMass","name":"Editor.DragMass","line":65,"end_line":75,"hash":"ef8db0580005175203554feaf7475de6cfa5ef33d19eb7f53a19644901766964"},{"id":"func/Editor.snap","name":"Editor.snap","line":77,"end_line":85,"hash":"e70b276e05feb94d9df1384a288b2de240b9583429553f173c4201d64d5591e3"},{"id":"func/nextMassID","name":"nextMassID","line":87,"end_line":89,"hash":"218110953901f2e93a10ad4e74dc3a488b77673ac005fc1608b9b72abc1b407c"},{"id":"func/nextSpringID","name":"nextSpringID","line":91,"end_line":93,"hash":"6f9e650722f3bed08e125d76e60001273d51ac6e1ca1c9d673cad0ef218f731d"},{"id":"func/nextID","name":"nextID","line":95,"end_line":104,"hash":"54eac88e91da9a5ad36542a2d58264f86328a5a1f53979314978e01fafacfe04"},{"id":"func/distance","name":"distance","line":106,"end_line":108,"hash":"ced4e5b049dd6ed54176b1185e8765f02ce417f06e7edd98dc29ac33b29c3661"},{"id":"func/parameterFloat","name":"parameterFloat","line":110,"end_line":113,"hash":"e3394cb46aa5eb644acf054a128677baf02b7755c2319ed95388866e6dd12b84"},{"id":"func/parameterBool","name":"parameterBool","line":115,"end_line":118,"hash":"4a04dd1285210fd96608c8f88f9ced49ee6ac7f761c4a5fa2f4f8dd89d99c3cd"}]}
// mutate4go-manifest-end
