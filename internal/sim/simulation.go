package sim

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrDuplicateID           = errors.New("duplicate id")
	ErrMissingSpringEndpoint = errors.New("missing spring endpoint")
)

const (
	defaultStepDuration = 0.016
	defaultPrecision    = 0.001
	advanceEpsilon      = 0.000000000001
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
	return &Simulation{Parameters: DefaultParameters(), Bounds: Bounds{Width: 640, Height: 480}}
}

func NewWorld() *Simulation {
	return NewSimulation()
}

func (s *Simulation) Clone() *Simulation {
	clone := NewSimulation()
	clone.LoadFrom(s)
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
	for remaining := duration; remaining > advanceEpsilon; {
		step := positiveAdvanceStep(math.Min(remaining, s.advanceStepDuration()), remaining)
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
	dt := s.configuredTimeStep()
	if s.Parameters.Value("adaptive timestep") != "true" {
		return dt
	}
	return adaptiveStepDuration(dt, s.configuredPrecision())
}

func (s *Simulation) configuredTimeStep() float64 {
	return positiveParameterOrDefault(s.Parameters, "timestep", defaultStepDuration)
}

func (s *Simulation) configuredPrecision() float64 {
	return positiveParameterOrDefault(s.Parameters, "precision", defaultPrecision)
}

func positiveParameterOrDefault(parameters Parameters, name string, defaultValue float64) float64 {
	value := parameterFloat(parameters, name)
	if value <= 0 {
		return defaultValue
	}
	return value
}

func adaptiveStepDuration(dt, precision float64) float64 {
	step := dt * sqrt(precision/defaultPrecision)
	if step <= 0 {
		return dt
	}
	return math.Min(step, dt)
}

func positiveAdvanceStep(step, remaining float64) float64 {
	if step <= 0 {
		return math.Min(remaining, defaultStepDuration)
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
		s.Masses[i].Velocity = start[i].Velocity.Add(weightedDerivative(k1[i].Acceleration, k2[i].Acceleration, k3[i].Acceleration, k4[i].Acceleration, dt))
		s.applyWallCollision(&s.Masses[i])
	}
	s.applyMassCollisions()
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T09:58:40-05:00","module_hash":"17e9a8254790ca4dbf0e0ff47a21ab120802a9bbdd1fa4f9b28bd6743f903746","functions":[{"id":"func/Vec2.Add","name":"Vec2.Add","line":25,"end_line":27,"hash":"77e2923b025f32c4e2dcde5f0fab9cfc40a52972e61d448ba4ed0209b802db32"},{"id":"func/Vec2.Sub","name":"Vec2.Sub","line":29,"end_line":31,"hash":"ac8fee4a8a1cee51daca7e5c5cef459e23de44bba38b9951da67f1a04324643e"},{"id":"func/Vec2.Scale","name":"Vec2.Scale","line":33,"end_line":35,"hash":"17415c0b629bdfbe0014a0a2ef6bf64f493f7f9fd5a1e7c6ef04a5f65d6e6ccf"},{"id":"func/Vec2.Normalize","name":"Vec2.Normalize","line":37,"end_line":43,"hash":"88c0b1a96b1a44c1f7829f00c593783855f509dceeddcdbf220c50a64944a3ff"},{"id":"func/NewSimulation","name":"NewSimulation","line":81,"end_line":83,"hash":"160a6d2ce890e3cd58e5ce940a9cd4befe3dc6cde7f208b51e3ca0e471f64289"},{"id":"func/NewWorld","name":"NewWorld","line":85,"end_line":87,"hash":"010cfd0223e2fa51981f618fa16e1cf849b544976a5976f7ad5e02cbb4d85c65"},{"id":"func/Simulation.Clone","name":"Simulation.Clone","line":89,"end_line":94,"hash":"721ce42906710d4d4a0daa67a737eb89e13a4d9e31b65330d4bc950f95632f87"},{"id":"func/Simulation.Reset","name":"Simulation.Reset","line":96,"end_line":101,"hash":"0466f0ed5207fb19d314eda6ea978a95ca31601395a4371c21e1fbcd2d1188d6"},{"id":"func/Simulation.LoadFrom","name":"Simulation.LoadFrom","line":103,"end_line":110,"hash":"14fddb9a2351dc9774f6a9c019078a62a9a3cbbe1a1112395d92cee56c3e3e1f"},{"id":"func/Simulation.InsertFrom","name":"Simulation.InsertFrom","line":112,"end_line":115,"hash":"0b50b5242df069f72ce29922b4030a3283e73e4df889a9fbeaa06bfa4924468d"},{"id":"func/NewDemoSimulation","name":"NewDemoSimulation","line":117,"end_line":123,"hash":"5f5f7b649b0d181ad81cad9bcd5282a30c5126df223171b1f452eeed6ff478ab"},{"id":"func/Simulation.AddMass","name":"Simulation.AddMass","line":125,"end_line":131,"hash":"7b918d417f9d5de518ff24f7a00ceabc4d6d389819510ee31c1ca2600c93c202"},{"id":"func/Simulation.AddMassAt","name":"Simulation.AddMassAt","line":133,"end_line":137,"hash":"4893fbdaa6ec7de25d9f75dfc527566c222f0cfadb346510057202223b321318"},{"id":"func/Simulation.AddSpring","name":"Simulation.AddSpring","line":139,"end_line":158,"hash":"6153f05093a0f9125cbbe448b4fc365fa18150d0fe36fd80b275e4584881340d"},{"id":"func/Simulation.AddSpringBetween","name":"Simulation.AddSpringBetween","line":160,"end_line":171,"hash":"b1d393e03bb0031f90aecd133feff685e7b0f06cf8ee2f6f7d3ae02fd63c1c3a"},{"id":"func/Simulation.MassByID","name":"Simulation.MassByID","line":173,"end_line":175,"hash":"7b0a33b903a704f23e0559c0cece6998b30f338dae6838453ecdefac52e86cc3"},{"id":"func/Simulation.SpringByID","name":"Simulation.SpringByID","line":177,"end_line":179,"hash":"651f958229ed21ae0b0b2ea9ee6b4439d7cb70c8e27d4e036fda1adaa0c8c066"},{"id":"func/byID","name":"byID","line":181,"end_line":189,"hash":"75d7cac77ad77bda3d6d508e269a76866882db52aac9c0ec44490bb786aaf782"},{"id":"func/Simulation.massIndexByID","name":"Simulation.massIndexByID","line":191,"end_line":198,"hash":"91e427e1030505e69f9b229111959c641e9d1aafcbc02f9aaca2b8a8157b611b"},{"id":"func/Simulation.Advance","name":"Simulation.Advance","line":200,"end_line":204,"hash":"7bc5b88b861d4365f11dfb1071441c783f926ea2f04fac265d90e32e6ca9dbd3"},{"id":"func/Simulation.AdvanceDuration","name":"Simulation.AdvanceDuration","line":206,"end_line":214,"hash":"9eb7a39863e9da6d3f336bd45dfc196565182c849cb7c544449de04d550024c7"},{"id":"func/Simulation.Step","name":"Simulation.Step","line":216,"end_line":219,"hash":"c0a831eb967708192b2a093af6e10a27a1245913c4ce183b025fbd96400e6c94"},{"id":"func/Simulation.advanceStepDuration","name":"Simulation.advanceStepDuration","line":221,"end_line":227,"hash":"38b5adafad153e1ccad81cf626b4338a3f01fab3029c447fb4c3d24cc70955c6"},{"id":"func/Simulation.configuredTimeStep","name":"Simulation.configuredTimeStep","line":229,"end_line":231,"hash":"0ed665a698c49efc36246f6542b9a5ae93e0f22ecb2ee0bac76d9b023a2046a3"},{"id":"func/Simulation.configuredPrecision","name":"Simulation.configuredPrecision","line":233,"end_line":235,"hash":"f7e72ba9eea2516b79e645a8c6dac234d75614c23af9da83f20dd6df9872b142"},{"id":"func/positiveParameterOrDefault","name":"positiveParameterOrDefault","line":237,"end_line":243,"hash":"cd081e55248c2446c9c56e700ec26835dd80988af5b8f491836da964adad2cad"},{"id":"func/adaptiveStepDuration","name":"adaptiveStepDuration","line":245,"end_line":251,"hash":"87bcbdd4d23648636cb25b9c14c30acc18d39a72eb34b3dcd41096118a2348d6"},{"id":"func/positiveAdvanceStep","name":"positiveAdvanceStep","line":253,"end_line":258,"hash":"63106042bc0d340ad63fed219d2aee3c1646db65024de54e5f0881e9bd265496"},{"id":"func/Simulation.stepRK4","name":"Simulation.stepRK4","line":260,"end_line":276,"hash":"65bc66b9df90d366887e579b6836d7a44bafaae8a494e224896a171000cdf7b4"},{"id":"func/Simulation.activeMasses","name":"Simulation.activeMasses","line":278,"end_line":293,"hash":"de9b03ced38a8c61d047b38370ab0e8926f28db905bb0c4adba3a18d5291b1ac"},{"id":"func/Simulation.derivatives","name":"Simulation.derivatives","line":300,"end_line":312,"hash":"4ada3a92560034281556e506a5251d8601c5250ac63ddcd7b5d869047841bb14"},{"id":"func/offsetMasses","name":"offsetMasses","line":314,"end_line":321,"hash":"2025e74dd2423ff69347e441d3e3f00687445013392ddf34be5954bf421c2936"},{"id":"func/weightedDerivative","name":"weightedDerivative","line":323,"end_line":325,"hash":"74000ace2dbda6f71cb2977c16b794f5638aa8c32d941d7a9b5a238879deb991"},{"id":"func/length","name":"length","line":327,"end_line":329,"hash":"29a2e4a2140a9fecae149ba85c782ae29c7be7b01788bb9f7d15c0fafea34eff"},{"id":"func/sqrt","name":"sqrt","line":331,"end_line":340,"hash":"4057d98eae74cbedfa6692b8182a71b07d6ac1f9550b749415ecc5a5a4ac4569"}]}
// mutate4go-manifest-end
