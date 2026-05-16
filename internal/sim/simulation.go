package sim

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

type Mass struct {
	Position Vec2
	Velocity Vec2
	Mass     float64
	Fixed    bool
}

type Spring struct {
	A          int
	B          int
	RestLength float64
	Stiffness  float64
}

type Simulation struct {
	Masses  []Mass
	Springs []Spring
	Damping float64
}

func NewSimulation() *Simulation {
	return &Simulation{Damping: 0.98}
}

func NewDemoSimulation() *Simulation {
	s := NewSimulation()
	left := s.AddMass(Vec2{X: 160, Y: 240}, 1, true)
	right := s.AddMass(Vec2{X: 320, Y: 240}, 1, false)
	s.AddSpring(left, right, 100, 12)
	return s
}

func (s *Simulation) AddMass(position Vec2, mass float64, fixed bool) int {
	s.Masses = append(s.Masses, Mass{Position: position, Mass: mass, Fixed: fixed})
	return len(s.Masses) - 1
}

func (s *Simulation) AddSpring(a, b int, restLength, stiffness float64) {
	s.Springs = append(s.Springs, Spring{A: a, B: b, RestLength: restLength, Stiffness: stiffness})
}

func (s *Simulation) Advance(steps int, dt float64) {
	for i := 0; i < steps; i++ {
		s.Step(dt)
	}
}

func (s *Simulation) Step(dt float64) {
	forces := make([]Vec2, len(s.Masses))
	for _, spring := range s.Springs {
		a := s.Masses[spring.A]
		b := s.Masses[spring.B]
		delta := b.Position.Sub(a.Position)
		distance := length(delta)
		if distance == 0 {
			continue
		}
		direction := delta.Scale(1 / distance)
		magnitude := spring.Stiffness * (distance - spring.RestLength)
		force := direction.Scale(magnitude)
		forces[spring.A] = forces[spring.A].Add(force)
		forces[spring.B] = forces[spring.B].Add(force.Scale(-1))
	}
	for i := range s.Masses {
		mass := &s.Masses[i]
		if mass.Fixed {
			continue
		}
		acceleration := forces[i].Scale(1 / mass.Mass)
		mass.Velocity = mass.Velocity.Add(acceleration.Scale(dt)).Scale(s.Damping)
		mass.Position = mass.Position.Add(mass.Velocity.Scale(dt))
	}
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
