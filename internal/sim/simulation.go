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
	StuckWall  string
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
	Masses           []Mass
	Springs          []Spring
	Damping          float64
	Parameters       Parameters
	Bounds           Bounds
	Time             float64
	LastAdvanceSteps int
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

func (s *Simulation) Clone() *Simulation {
	clone := NewSimulation()
	clone.LoadFrom(s)
	clone.Damping = s.Damping
	clone.Bounds = s.Bounds
	return clone
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
	s.Damping = other.Damping
	s.Bounds = other.Bounds
	s.LastAdvanceSteps = other.LastAdvanceSteps
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
	return byID(s.Masses, id, func(mass Mass) int { return mass.ID })
}

func (s *Simulation) SpringByID(id int) (Spring, bool) {
	return byID(s.Springs, id, func(spring Spring) int { return spring.ID })
}

func byID[T any](items []T, id int, itemID func(T) int) (T, bool) {
	for _, item := range items {
		if itemID(item) == id {
			return item, true
		}
	}
	var zero T
	return zero, false
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
	s.LastAdvanceSteps = 0
	for remaining := duration; remaining > 0.000000000001; {
		step := s.advanceStepDuration()
		if remaining < step {
			step = remaining
		}
		s.Step(step)
		s.LastAdvanceSteps++
		remaining -= step
	}
}

func (s *Simulation) Step(dt float64) {
	s.stepRK4(dt)
	s.Time += dt
}

func (s *Simulation) advanceStepDuration() float64 {
	dt := parameterFloat(s.Parameters, "timestep")
	if dt <= 0 {
		dt = 0.016
	}
	if s.Parameters.Value("adaptive timestep") != "true" {
		return dt
	}
	precision := parameterFloat(s.Parameters, "precision")
	if precision <= 0 {
		precision = 0.001
	}
	step := dt * sqrt(precision/0.001)
	if step <= 0 {
		return dt
	}
	if step > dt {
		return dt
	}
	return step
}

func (s *Simulation) stepRK4(dt float64) {
	active := s.activeMasses()
	start := append([]Mass{}, s.Masses...)
	k1 := s.derivatives(start, active)
	k2 := s.derivatives(offsetMasses(start, k1, dt/2), active)
	k3 := s.derivatives(offsetMasses(start, k2, dt/2), active)
	k4 := s.derivatives(offsetMasses(start, k3, dt), active)
	for i := range s.Masses {
		if !active[i] {
			continue
		}
		s.Masses[i].Position = start[i].Position.Add(weightedDerivative(k1[i].Velocity, k2[i].Velocity, k3[i].Velocity, k4[i].Velocity, dt))
		s.Masses[i].Velocity = start[i].Velocity.Add(weightedDerivative(k1[i].Acceleration, k2[i].Acceleration, k3[i].Acceleration, k4[i].Acceleration, dt)).Scale(s.Damping)
		s.applyWallCollision(&s.Masses[i])
	}
}

func (s *Simulation) activeMasses() []bool {
	evaluation := s.EvaluateForces()
	active := make([]bool, len(s.Masses))
	for i := range s.Masses {
		mass := &s.Masses[i]
		if mass.Fixed {
			continue
		}
		acceleration := evaluation.ByMassID[mass.ID].Acceleration
		if s.keepStuck(mass, acceleration) {
			continue
		}
		active[i] = true
	}
	return active
}

type massDerivative struct {
	Velocity     Vec2
	Acceleration Vec2
}

func (s *Simulation) derivatives(masses []Mass, active []bool) []massDerivative {
	original := s.Masses
	s.Masses = masses
	evaluation := s.EvaluateForces()
	s.Masses = original
	derivatives := make([]massDerivative, len(masses))
	for i, mass := range masses {
		if active[i] {
			derivatives[i] = massDerivative{Velocity: mass.Velocity, Acceleration: evaluation.ByMassID[mass.ID].Acceleration}
		}
	}
	return derivatives
}

func offsetMasses(masses []Mass, derivatives []massDerivative, dt float64) []Mass {
	offset := append([]Mass{}, masses...)
	for i := range offset {
		offset[i].Position = offset[i].Position.Add(derivatives[i].Velocity.Scale(dt))
		offset[i].Velocity = offset[i].Velocity.Add(derivatives[i].Acceleration.Scale(dt))
	}
	return offset
}

func weightedDerivative(k1, k2, k3, k4 Vec2, dt float64) Vec2 {
	return k1.Add(k2.Scale(2)).Add(k3.Scale(2)).Add(k4).Scale(dt / 6)
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
