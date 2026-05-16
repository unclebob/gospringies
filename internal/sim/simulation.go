package sim

import (
	"errors"
	"fmt"
)

var (
	ErrDuplicateID           = errors.New("duplicate id")
	ErrMissingSpringEndpoint = errors.New("missing spring endpoint")
)

type Vec2 struct {
	X float64
	Y float64
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vec2) Scale(factor float64) Vec2 {
	return Vec2{X: v.X * factor, Y: v.Y * factor}
}

func (v Vec2) Normalize() Vec2 {
	l := length(v)
	if l == 0 {
		return Vec2{}
	}
	return v.Scale(1 / l)
}

type Mass struct {
	ID         int
	Position   Vec2
	Velocity   Vec2
	Mass       float64
	Elasticity float64
	Fixed      bool
}

type Spring struct {
	ID             int
	A              int
	B              int
	MassA          int
	MassB          int
	RestLength     float64
	Stiffness      float64
	SpringConstant float64
	Damping        float64
}

type Simulation struct {
	Masses     []Mass
	Springs    []Spring
	Damping    float64
	Parameters Parameters
	Bounds     Bounds
	Time       float64
}

type Bounds struct {
	Width  float64
	Height float64
}

func NewSimulation() *Simulation {
	return &Simulation{Damping: 0.98, Parameters: DefaultParameters(), Bounds: Bounds{Width: 640, Height: 480}}
}

func NewWorld() *Simulation {
	return NewSimulation()
}

func (s *Simulation) Reset() {
	s.Masses = nil
	s.Springs = nil
	s.Parameters = DefaultParameters()
	s.Time = 0
}

func (s *Simulation) LoadFrom(other *Simulation) {
	s.Masses = append([]Mass{}, other.Masses...)
	s.Springs = append([]Spring{}, other.Springs...)
	s.Parameters = other.Parameters.Clone()
	s.Time = other.Time
}

func (s *Simulation) InsertFrom(other *Simulation) {
	s.Masses = append(s.Masses, other.Masses...)
	s.Springs = append(s.Springs, other.Springs...)
}

func NewDemoSimulation() *Simulation {
	s := NewSimulation()
	left := s.AddMassAt(Vec2{X: 160, Y: 240}, 1, true)
	right := s.AddMassAt(Vec2{X: 320, Y: 240}, 1, false)
	s.AddSpringBetween(left, right, 100, 12)
	return s
}

func (s *Simulation) AddMass(mass Mass) error {
	if _, ok := s.MassByID(mass.ID); ok {
		return fmt.Errorf("%w: mass %d", ErrDuplicateID, mass.ID)
	}
	s.Masses = append(s.Masses, mass)
	return nil
}

func (s *Simulation) AddMassAt(position Vec2, mass float64, fixed bool) int {
	id := len(s.Masses) + 1
	s.Masses = append(s.Masses, Mass{ID: id, Position: position, Mass: mass, Fixed: fixed})
	return len(s.Masses) - 1
}

func (s *Simulation) AddSpring(spring Spring) error {
	if _, ok := s.SpringByID(spring.ID); ok {
		return fmt.Errorf("%w: spring %d", ErrDuplicateID, spring.ID)
	}
	aIndex, okA := s.massIndexByID(spring.MassA)
	bIndex, okB := s.massIndexByID(spring.MassB)
	if !okA || !okB {
		return fmt.Errorf("%w: spring %d", ErrMissingSpringEndpoint, spring.ID)
	}
	spring.A = aIndex
	spring.B = bIndex
	if spring.Stiffness == 0 {
		spring.Stiffness = spring.SpringConstant
	}
	if spring.SpringConstant == 0 {
		spring.SpringConstant = spring.Stiffness
	}
	s.Springs = append(s.Springs, spring)
	return nil
}

func (s *Simulation) AddSpringBetween(a, b int, restLength, stiffness float64) {
	s.Springs = append(s.Springs, Spring{
		ID:             len(s.Springs) + 1,
		A:              a,
		B:              b,
		MassA:          s.Masses[a].ID,
		MassB:          s.Masses[b].ID,
		RestLength:     restLength,
		Stiffness:      stiffness,
		SpringConstant: stiffness,
	})
}

func (s *Simulation) MassByID(id int) (Mass, bool) {
	for _, mass := range s.Masses {
		if mass.ID == id {
			return mass, true
		}
	}
	return Mass{}, false
}

func (s *Simulation) SpringByID(id int) (Spring, bool) {
	for _, spring := range s.Springs {
		if spring.ID == id {
			return spring, true
		}
	}
	return Spring{}, false
}

func (s *Simulation) massIndexByID(id int) (int, bool) {
	for i, mass := range s.Masses {
		if mass.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *Simulation) Advance(steps int, dt float64) {
	for i := 0; i < steps; i++ {
		s.Step(dt)
	}
}

func (s *Simulation) AdvanceDuration(duration float64) {
	dt := parameterFloat(s.Parameters, "timestep")
	if dt <= 0 {
		dt = 0.016
	}
	for remaining := duration; remaining > 0; {
		step := dt
		if remaining < step {
			step = remaining
		}
		s.Step(step)
		remaining -= step
	}
}

func (s *Simulation) Step(dt float64) {
	evaluation := s.EvaluateForces()
	for i := range s.Masses {
		mass := &s.Masses[i]
		if mass.Fixed {
			continue
		}
		acceleration := evaluation.ByMassID[mass.ID].Acceleration
		mass.Velocity = mass.Velocity.Add(acceleration.Scale(dt)).Scale(s.Damping)
		mass.Position = mass.Position.Add(mass.Velocity.Scale(dt))
	}
	s.Time += dt
}

func length(v Vec2) float64 {
	return sqrt(v.X*v.X + v.Y*v.Y)
}

func sqrt(value float64) float64 {
	z := value
	if z == 0 {
		return 0
	}
	for i := 0; i < 10; i++ {
		z -= (z*z - value) / (2 * z)
	}
	return z
}
