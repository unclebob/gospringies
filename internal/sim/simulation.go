package sim

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
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
	Wall           bool
	Temperature    float64
}

type Simulation struct {
	Masses           []Mass
	Springs          []Spring
	Parameters       Parameters
	Bounds           Bounds
	Time             float64
	LastAdvanceSteps int
	temperatureRand  *rand.Rand
}

type Bounds struct {
	Width  float64
	Height float64
	Left   float64
	Right  float64
	Bottom float64
	Top    float64
}

func (b Bounds) MinX() float64 {
	return b.Left
}

func (b Bounds) MaxX() float64 {
	return configuredBoundary(b.Right, b.Width)
}

func (b Bounds) MinY() float64 {
	return b.Bottom
}

func (b Bounds) MaxY() float64 {
	return configuredBoundary(b.Top, b.Height)
}

func configuredBoundary(value, fallback float64) float64 {
	if value != 0 {
		return value
	}
	return fallback
}

func (b Bounds) Center() Vec2 {
	return Vec2{X: (b.MinX() + b.MaxX()) / 2, Y: (b.MinY() + b.MaxY()) / 2}
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

func (s *Simulation) SetTemperatureSeed(seed int64) {
	s.temperatureRand = rand.New(rand.NewSource(seed))
}

func (s *Simulation) temperatureRandom() *rand.Rand {
	if s.temperatureRand == nil {
		s.SetTemperatureSeed(1)
	}
	return s.temperatureRand
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
	s.cleanupOffCanvasObjects()
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
	startPositions := massPositions(start)
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
	beforeLengthConstraints := massPositions(s.Masses)
	s.applyWallSpringLengthConstraints()
	s.applyWallSpringLengthConstraintCollisions(dt, beforeLengthConstraints)
	s.applyMovingWallSpringFixedEndpointCollisions(dt, startPositions)
	s.applyWallSpringCollisions(dt, startPositions)
	s.applyMassCollisions()
	s.applyPostContactReconciliation()
}

func massPositions(masses []Mass) []Vec2 {
	positions := make([]Vec2, len(masses))
	for i, mass := range masses {
		positions[i] = mass.Position
	}
	return positions
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
	return math.Sqrt(value)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-23T11:43:25-05:00","module_hash":"bd485fec83c8e462cd4dfa731e389f6deebc1d97590e4afcc36ca4f4464e1d57","functions":[{"id":"func/Vec2.Add","name":"Vec2.Add","line":26,"end_line":28,"hash":"77e2923b025f32c4e2dcde5f0fab9cfc40a52972e61d448ba4ed0209b802db32"},{"id":"func/Vec2.Sub","name":"Vec2.Sub","line":30,"end_line":32,"hash":"ac8fee4a8a1cee51daca7e5c5cef459e23de44bba38b9951da67f1a04324643e"},{"id":"func/Vec2.Scale","name":"Vec2.Scale","line":34,"end_line":36,"hash":"17415c0b629bdfbe0014a0a2ef6bf64f493f7f9fd5a1e7c6ef04a5f65d6e6ccf"},{"id":"func/Vec2.Normalize","name":"Vec2.Normalize","line":38,"end_line":44,"hash":"88c0b1a96b1a44c1f7829f00c593783855f509dceeddcdbf220c50a64944a3ff"},{"id":"func/Bounds.MinX","name":"Bounds.MinX","line":89,"end_line":91,"hash":"520d95605054ac01a91b5f597efb0c637eb21aadc4eda6b321e316163cfa79c2"},{"id":"func/Bounds.MaxX","name":"Bounds.MaxX","line":93,"end_line":95,"hash":"2aeb799da32c29a05ec6e8a0048d4ead598f5bad6ed244eb8eb867726a1a402b"},{"id":"func/Bounds.MinY","name":"Bounds.MinY","line":97,"end_line":99,"hash":"63198ffa9b273508fdaab9283e738adfcd8ec3397f7ba100f2c3f8b27721c0a3"},{"id":"func/Bounds.MaxY","name":"Bounds.MaxY","line":101,"end_line":103,"hash":"e339bb036932965cff911b1ebed867cae5f7a558ae683d28ad6bcffde1ddba5a"},{"id":"func/configuredBoundary","name":"configuredBoundary","line":105,"end_line":110,"hash":"01bda5a9c3d9ecee117dbd158ec48190ede11adb9c66209223808350e9c56473"},{"id":"func/Bounds.Center","name":"Bounds.Center","line":112,"end_line":114,"hash":"1d9fe3c4a960661cd4717bf13a9ee79963e86018aeab49eb88a47e26fb9cf9e5"},{"id":"func/NewSimulation","name":"NewSimulation","line":116,"end_line":118,"hash":"160a6d2ce890e3cd58e5ce940a9cd4befe3dc6cde7f208b51e3ca0e471f64289"},{"id":"func/NewWorld","name":"NewWorld","line":120,"end_line":122,"hash":"010cfd0223e2fa51981f618fa16e1cf849b544976a5976f7ad5e02cbb4d85c65"},{"id":"func/Simulation.Clone","name":"Simulation.Clone","line":124,"end_line":129,"hash":"721ce42906710d4d4a0daa67a737eb89e13a4d9e31b65330d4bc950f95632f87"},{"id":"func/Simulation.Reset","name":"Simulation.Reset","line":131,"end_line":136,"hash":"0466f0ed5207fb19d314eda6ea978a95ca31601395a4371c21e1fbcd2d1188d6"},{"id":"func/Simulation.LoadFrom","name":"Simulation.LoadFrom","line":138,"end_line":145,"hash":"14fddb9a2351dc9774f6a9c019078a62a9a3cbbe1a1112395d92cee56c3e3e1f"},{"id":"func/Simulation.SetTemperatureSeed","name":"Simulation.SetTemperatureSeed","line":147,"end_line":149,"hash":"2994d00632496f36186b5f6d985be7ab891d95e0ea52862b11cb26696d71f7a4"},{"id":"func/Simulation.temperatureRandom","name":"Simulation.temperatureRandom","line":151,"end_line":156,"hash":"e040fd5131abc90780822ec8b851ea3be38a21aa4b7baa189fd9edf62abcc3f4"},{"id":"func/Simulation.InsertFrom","name":"Simulation.InsertFrom","line":158,"end_line":161,"hash":"0b50b5242df069f72ce29922b4030a3283e73e4df889a9fbeaa06bfa4924468d"},{"id":"func/NewDemoSimulation","name":"NewDemoSimulation","line":163,"end_line":169,"hash":"5f5f7b649b0d181ad81cad9bcd5282a30c5126df223171b1f452eeed6ff478ab"},{"id":"func/Simulation.AddMass","name":"Simulation.AddMass","line":171,"end_line":177,"hash":"7b918d417f9d5de518ff24f7a00ceabc4d6d389819510ee31c1ca2600c93c202"},{"id":"func/Simulation.AddMassAt","name":"Simulation.AddMassAt","line":179,"end_line":183,"hash":"4893fbdaa6ec7de25d9f75dfc527566c222f0cfadb346510057202223b321318"},{"id":"func/Simulation.AddSpring","name":"Simulation.AddSpring","line":185,"end_line":204,"hash":"6153f05093a0f9125cbbe448b4fc365fa18150d0fe36fd80b275e4584881340d"},{"id":"func/Simulation.AddSpringBetween","name":"Simulation.AddSpringBetween","line":206,"end_line":217,"hash":"b1d393e03bb0031f90aecd133feff685e7b0f06cf8ee2f6f7d3ae02fd63c1c3a"},{"id":"func/Simulation.MassByID","name":"Simulation.MassByID","line":219,"end_line":221,"hash":"7b0a33b903a704f23e0559c0cece6998b30f338dae6838453ecdefac52e86cc3"},{"id":"func/Simulation.SpringByID","name":"Simulation.SpringByID","line":223,"end_line":225,"hash":"651f958229ed21ae0b0b2ea9ee6b4439d7cb70c8e27d4e036fda1adaa0c8c066"},{"id":"func/byID","name":"byID","line":227,"end_line":235,"hash":"75d7cac77ad77bda3d6d508e269a76866882db52aac9c0ec44490bb786aaf782"},{"id":"func/Simulation.massIndexByID","name":"Simulation.massIndexByID","line":237,"end_line":244,"hash":"91e427e1030505e69f9b229111959c641e9d1aafcbc02f9aaca2b8a8157b611b"},{"id":"func/Simulation.Advance","name":"Simulation.Advance","line":246,"end_line":250,"hash":"7bc5b88b861d4365f11dfb1071441c783f926ea2f04fac265d90e32e6ca9dbd3"},{"id":"func/Simulation.AdvanceDuration","name":"Simulation.AdvanceDuration","line":252,"end_line":260,"hash":"9eb7a39863e9da6d3f336bd45dfc196565182c849cb7c544449de04d550024c7"},{"id":"func/Simulation.Step","name":"Simulation.Step","line":262,"end_line":266,"hash":"e7c043cac5768a8a9cd3ed1c012d60c095deaad258d4f63165743c2d5f8b0acb"},{"id":"func/Simulation.advanceStepDuration","name":"Simulation.advanceStepDuration","line":268,"end_line":274,"hash":"38b5adafad153e1ccad81cf626b4338a3f01fab3029c447fb4c3d24cc70955c6"},{"id":"func/Simulation.configuredTimeStep","name":"Simulation.configuredTimeStep","line":276,"end_line":278,"hash":"0ed665a698c49efc36246f6542b9a5ae93e0f22ecb2ee0bac76d9b023a2046a3"},{"id":"func/Simulation.configuredPrecision","name":"Simulation.configuredPrecision","line":280,"end_line":282,"hash":"f7e72ba9eea2516b79e645a8c6dac234d75614c23af9da83f20dd6df9872b142"},{"id":"func/positiveParameterOrDefault","name":"positiveParameterOrDefault","line":284,"end_line":290,"hash":"cd081e55248c2446c9c56e700ec26835dd80988af5b8f491836da964adad2cad"},{"id":"func/adaptiveStepDuration","name":"adaptiveStepDuration","line":292,"end_line":298,"hash":"87bcbdd4d23648636cb25b9c14c30acc18d39a72eb34b3dcd41096118a2348d6"},{"id":"func/positiveAdvanceStep","name":"positiveAdvanceStep","line":300,"end_line":305,"hash":"63106042bc0d340ad63fed219d2aee3c1646db65024de54e5f0881e9bd265496"},{"id":"func/Simulation.stepRK4","name":"Simulation.stepRK4","line":307,"end_line":330,"hash":"fd24363707d744b800ec4c5e03f5c98911d8c35aadbea3f5b0d8f99d4ca6e36a"},{"id":"func/massPositions","name":"massPositions","line":332,"end_line":338,"hash":"d23e3b60be1f2db7d62043cf92f8623b7c8151fc9ef9827c812867603bbbf26a"},{"id":"func/Simulation.activeMasses","name":"Simulation.activeMasses","line":340,"end_line":355,"hash":"de9b03ced38a8c61d047b38370ab0e8926f28db905bb0c4adba3a18d5291b1ac"},{"id":"func/Simulation.derivatives","name":"Simulation.derivatives","line":362,"end_line":374,"hash":"4ada3a92560034281556e506a5251d8601c5250ac63ddcd7b5d869047841bb14"},{"id":"func/offsetMasses","name":"offsetMasses","line":376,"end_line":383,"hash":"2025e74dd2423ff69347e441d3e3f00687445013392ddf34be5954bf421c2936"},{"id":"func/weightedDerivative","name":"weightedDerivative","line":385,"end_line":387,"hash":"74000ace2dbda6f71cb2977c16b794f5638aa8c32d941d7a9b5a238879deb991"},{"id":"func/length","name":"length","line":389,"end_line":391,"hash":"29a2e4a2140a9fecae149ba85c782ae29c7be7b01788bb9f7d15c0fafea34eff"},{"id":"func/sqrt","name":"sqrt","line":393,"end_line":395,"hash":"1efa773ed0f7f72db28806559a4178b8a49ae9a59597cae90629e23c3e094b79"}]}
// mutate4go-manifest-end
