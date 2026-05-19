package sim

import (
	"math"
	"strconv"
)

type ForceEvaluation struct {
	ByMassID map[int]MassForces
}

type MassForces struct {
	Force        Vec2
	Acceleration Vec2
}

var forceParameterNames = map[string][]string{
	"gravity":                   {"Magnitude", "Direction"},
	"center of mass attraction": {"Magnitude", "Damping"},
	"center attraction":         {"Magnitude", "Exponent"},
	"wall repulsion":            {"Magnitude", "Exponent"},
	"mass collision":            {},
}

func (s *Simulation) EvaluateForces() ForceEvaluation {
	evaluation := ForceEvaluation{ByMassID: map[int]MassForces{}}
	for _, mass := range s.Masses {
		evaluation.ByMassID[mass.ID] = MassForces{}
	}
	s.addSpringForces(evaluation.ByMassID)
	s.addEnvironmentalForces(evaluation.ByMassID)
	s.computeAccelerations(evaluation.ByMassID)
	return evaluation
}

func (s *Simulation) addSpringForces(forces map[int]MassForces) {
	for _, spring := range s.Springs {
		a, b, ok := s.springEndpointMasses(spring)
		if !ok {
			continue
		}
		delta := b.Position.Sub(a.Position)
		distance := length(delta)
		if distance == 0 {
			continue
		}
		direction := delta.Scale(1 / distance)
		magnitude := spring.SpringConstant * (distance - spring.RestLength)
		relativeVelocity := b.Velocity.Sub(a.Velocity)
		magnitude += spring.Damping * dot(relativeVelocity, direction)
		force := direction.Scale(magnitude)
		addForce(forces, a.ID, force)
		addForce(forces, b.ID, force.Scale(-1))
	}
}

func (s *Simulation) springEndpointMasses(spring Spring) (Mass, Mass, bool) {
	if spring.MassA != 0 || spring.MassB != 0 {
		return s.springEndpointMassesByID(spring)
	}
	return s.springEndpointMassesByIndex(spring)
}

func (s *Simulation) springEndpointMassesByID(spring Spring) (Mass, Mass, bool) {
	a, okA := s.MassByID(spring.MassA)
	b, okB := s.MassByID(spring.MassB)
	return a, b, okA && okB
}

func (s *Simulation) springEndpointMassesByIndex(spring Spring) (Mass, Mass, bool) {
	if !s.validSpringMassIndexes(spring) {
		return Mass{}, Mass{}, false
	}
	return s.Masses[spring.A], s.Masses[spring.B], true
}

func (s *Simulation) validSpringMassIndexes(spring Spring) bool {
	return validMassIndex(spring.A, len(s.Masses)) && validMassIndex(spring.B, len(s.Masses))
}

func validMassIndex(index int, massCount int) bool {
	return index >= 0 && index < massCount
}

func (s *Simulation) addEnvironmentalForces(forces map[int]MassForces) {
	for _, mass := range s.Masses {
		addForce(forces, mass.ID, s.gravityForce(mass))
		addForce(forces, mass.ID, s.viscosityForce(mass))
		addForce(forces, mass.ID, s.centerForce(mass, "center attraction", s.forceCenter()))
		addForce(forces, mass.ID, s.centerForce(mass, "center of mass attraction", s.centerOfMass()))
		addForce(forces, mass.ID, s.wallForce(mass))
	}
}

func (s *Simulation) computeAccelerations(forces map[int]MassForces) {
	for _, mass := range s.Masses {
		entry := forces[mass.ID]
		if !mass.Fixed && mass.Mass != 0 {
			entry.Acceleration = entry.Force.Scale(1 / mass.Mass)
		}
		forces[mass.ID] = entry
	}
}

func (s *Simulation) gravityForce(mass Mass) Vec2 {
	force, ok := s.enabledForce("gravity")
	if !ok {
		return Vec2{}
	}
	magnitude := forceFloat(force, "magnitude")
	radians := forceFloat(force, "direction") * math.Pi / 180
	return Vec2{X: magnitude * math.Sin(radians) * mass.Mass, Y: -magnitude * math.Cos(radians) * mass.Mass}
}

func (s *Simulation) viscosityForce(mass Mass) Vec2 {
	viscosity := parameterFloat(s.Parameters, "viscosity")
	return mass.Velocity.Scale(-viscosity)
}

func (s *Simulation) centerForce(mass Mass, name string, center Vec2) Vec2 {
	force, ok := s.enabledForce(name)
	if !ok || s.IsCenterMass(mass.ID) {
		return Vec2{}
	}
	delta := center.Sub(mass.Position)
	distance := length(delta)
	if distance == 0 {
		return Vec2{}
	}
	direction := delta.Scale(1 / distance)
	magnitude := forceFloat(force, "magnitude") / math.Pow(distance, forceExponent(force))
	if name == "center of mass attraction" {
		magnitude -= forceFloat(force, "damping") * dot(mass.Velocity, direction)
	}
	return direction.Scale(magnitude)
}

func (s *Simulation) wallForce(mass Mass) Vec2 {
	force, ok := s.enabledForce("wall repulsion")
	if !ok {
		return Vec2{}
	}
	magnitude := forceFloat(force, "magnitude")
	var total Vec2
	for _, wall := range s.wallChecks(mass, magnitude) {
		if enabled, _ := s.Parameters.WallEnabled(wall.name); enabled && wall.inside {
			total = total.Add(wall.force)
		}
	}
	return total
}

type wallCheck struct {
	name   string
	inside bool
	force  Vec2
}

func (s *Simulation) wallChecks(mass Mass, magnitude float64) []wallCheck {
	exponent := forceExponent(s.Parameters.Forces["wall repulsion"])
	return []wallCheck{
		{name: "bottom", inside: mass.Position.Y >= 0, force: Vec2{Y: wallMagnitude(magnitude, mass.Position.Y, exponent)}},
		{name: "left", inside: mass.Position.X >= 0, force: Vec2{X: wallMagnitude(magnitude, mass.Position.X, exponent)}},
		{name: "right", inside: mass.Position.X <= s.Bounds.Width, force: Vec2{X: -wallMagnitude(magnitude, s.Bounds.Width-mass.Position.X, exponent)}},
		{name: "top", inside: mass.Position.Y <= s.Bounds.Height, force: Vec2{Y: -wallMagnitude(magnitude, s.Bounds.Height-mass.Position.Y, exponent)}},
	}
}

func (s *Simulation) centerOfMass() Vec2 {
	var total Vec2
	var count float64
	for _, mass := range s.Masses {
		total = total.Add(mass.Position)
		count++
	}
	if count == 0 {
		return s.screenCenter()
	}
	return total.Scale(1 / count)
}

func (s *Simulation) forceCenter() Vec2 {
	id := s.CenterMassID()
	if id <= 0 {
		return s.screenCenter()
	}
	mass, ok := s.MassByID(id)
	if !ok {
		return s.screenCenter()
	}
	return mass.Position
}

func (s *Simulation) screenCenter() Vec2 {
	return Vec2{X: s.Bounds.Width / 2, Y: s.Bounds.Height / 2}
}

func (s *Simulation) SetForceCenter(selectedMassIDs []int) {
	centerID := -1
	if len(selectedMassIDs) == 1 {
		centerID = selectedMassIDs[0]
	}
	s.Parameters.Set("center mass", strconv.Itoa(centerID))
}

func (s *Simulation) CenterMassID() int {
	id, err := strconv.Atoi(s.Parameters.Value("center mass"))
	if err != nil {
		return -1
	}
	return id
}

func (s *Simulation) IsCenterMass(id int) bool {
	return id > 0 && s.CenterMassID() == id
}

func ForceParameterNames(force string) []string {
	return append([]string{}, forceParameterNames[force]...)
}

func (s *Simulation) enabledForce(name string) (ForceConfig, bool) {
	force, ok := s.Parameters.Force(name)
	return force, ok && force.Enabled == "true"
}

func addForce(forces map[int]MassForces, id int, force Vec2) {
	entry := forces[id]
	entry.Force = entry.Force.Add(force)
	forces[id] = entry
}

func parameterFloat(parameters Parameters, key string) float64 {
	value, _ := strconv.ParseFloat(parameters.Value(key), 64)
	return value
}

func forceFloat(force ForceConfig, key string) float64 {
	value, _ := strconv.ParseFloat(force.Values[key], 64)
	return value
}

func forceExponent(force ForceConfig) float64 {
	value, ok := force.Values["exponent"]
	if !ok {
		return 1
	}
	exponent, _ := strconv.ParseFloat(value, 64)
	return exponent
}

func wallMagnitude(magnitude, distance, exponent float64) float64 {
	return magnitude / math.Pow(math.Max(1, distance), exponent)
}

func dot(a, b Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T09:55:17-05:00","module_hash":"39a6c9c80dca7446bd4f0fbe42b5ce671dfe54e59b7690965f0cd4833d083489","functions":[{"id":"func/Simulation.EvaluateForces","name":"Simulation.EvaluateForces","line":25,"end_line":34,"hash":"9a97513e4bf8b564213c8a1e8c8a18861fb6a94b4c48d68615b70f7a8ad5fc63"},{"id":"func/Simulation.addSpringForces","name":"Simulation.addSpringForces","line":36,"end_line":55,"hash":"1db61f41364e492b32333082caf8353be8238a0b8d7769fa0670aeaed6983899"},{"id":"func/Simulation.springEndpointMasses","name":"Simulation.springEndpointMasses","line":57,"end_line":62,"hash":"f254408933e355b20c7e915d94675084b71a93f85dfd2073d82f46c481465590"},{"id":"func/Simulation.springEndpointMassesByID","name":"Simulation.springEndpointMassesByID","line":64,"end_line":68,"hash":"b99d2c4579fca6e79d76c1cf2e2c363a8684415aa8043d8e26c50f0fa7cb6791"},{"id":"func/Simulation.springEndpointMassesByIndex","name":"Simulation.springEndpointMassesByIndex","line":70,"end_line":75,"hash":"6168baf875422397cd9abc4272431b554139c2836bf3ece8e1418f1777b7e36e"},{"id":"func/Simulation.validSpringMassIndexes","name":"Simulation.validSpringMassIndexes","line":77,"end_line":79,"hash":"8876739cc04156954fa7ee0a2d3b140c6aba067a4a7d9fac1805ab1ede7e0f76"},{"id":"func/validMassIndex","name":"validMassIndex","line":81,"end_line":83,"hash":"656e45e0268548c36a02ec6a2e708b46176aea5cb9c5349aadc281d9362078a9"},{"id":"func/Simulation.addEnvironmentalForces","name":"Simulation.addEnvironmentalForces","line":85,"end_line":93,"hash":"d1b0eb9e1cea25219a5a2a02295a153619e5e94805bc52d1d2025d200fa56cf5"},{"id":"func/Simulation.computeAccelerations","name":"Simulation.computeAccelerations","line":95,"end_line":103,"hash":"ac6d87956a2ce3daff4989d4e521c1e23e49e10c7e7ac768d42d00c85c51109b"},{"id":"func/Simulation.gravityForce","name":"Simulation.gravityForce","line":105,"end_line":113,"hash":"eddcfb117420d1166da752ae7aa48dfc30046912790b2e25d5918a220db3681f"},{"id":"func/Simulation.viscosityForce","name":"Simulation.viscosityForce","line":115,"end_line":118,"hash":"d045d51359e0437b53475db405cd820faa24e571b5302b0442e936e63c608d78"},{"id":"func/Simulation.centerForce","name":"Simulation.centerForce","line":120,"end_line":136,"hash":"01994d8a974d76d4830458c936758e105c63ed1afd128a906c15371053e2b892"},{"id":"func/Simulation.wallForce","name":"Simulation.wallForce","line":138,"end_line":151,"hash":"96c2352bbbc11ce093925d6a9e0a20c3ec5e5556e013d29b29c958546b8e7f82"},{"id":"func/Simulation.wallChecks","name":"Simulation.wallChecks","line":159,"end_line":167,"hash":"df7b72659bb22e8cafa137a288d5679743821cabb79d21dbd59e94759ffadf63"},{"id":"func/Simulation.centerOfMass","name":"Simulation.centerOfMass","line":169,"end_line":180,"hash":"9fbc400d170e3fdb4bcb3dbbcc9c53afdc28a8c1996acf043f9f413a916c9a97"},{"id":"func/Simulation.forceCenter","name":"Simulation.forceCenter","line":182,"end_line":192,"hash":"67cd28ae111407a13dab28df84c513da41e4e87ebfb210374e887bccc4815618"},{"id":"func/Simulation.screenCenter","name":"Simulation.screenCenter","line":194,"end_line":196,"hash":"f13a1c9c99afbce0d9b131cfd76947f1a5e7b4a6ae9bd960a911e1668ef9d779"},{"id":"func/Simulation.SetForceCenter","name":"Simulation.SetForceCenter","line":198,"end_line":204,"hash":"86d6d60eaf1fee37213c0a3658d7e14040dce1a1c6a3eee2820c40b741a5d56d"},{"id":"func/Simulation.CenterMassID","name":"Simulation.CenterMassID","line":206,"end_line":212,"hash":"bf5f8948745bf7c2cb9c0a94cc795b0f2386cc4707f4f9d962a487e6ecb6c284"},{"id":"func/Simulation.IsCenterMass","name":"Simulation.IsCenterMass","line":214,"end_line":216,"hash":"5e7a5f367d4dbb975080eab13a434decceee80d1888eb1b6f04f461cd2396f38"},{"id":"func/ForceParameterNames","name":"ForceParameterNames","line":218,"end_line":220,"hash":"a63d436ad8f5355264962d7a774340c86f6651d2738fe93edcac2d837f94aaed"},{"id":"func/Simulation.enabledForce","name":"Simulation.enabledForce","line":222,"end_line":225,"hash":"660818389c8a87c6c696514ba72b4fd6ada16d39f53f53e06a5798def9b35849"},{"id":"func/addForce","name":"addForce","line":227,"end_line":231,"hash":"6fa9626420a10c6961877f9007f3f43fd2a948fa17adc10c7916fc2ea93524f1"},{"id":"func/parameterFloat","name":"parameterFloat","line":233,"end_line":236,"hash":"29393cb91071a38a61514342a47d9a72ce8b2a8dbcaa72c34b6c0878dbf69ffd"},{"id":"func/forceFloat","name":"forceFloat","line":238,"end_line":241,"hash":"44938573269172ed830611d51e98c06e7c3c0beb9cea7bb5da78a1dc97d2df5c"},{"id":"func/forceExponent","name":"forceExponent","line":243,"end_line":250,"hash":"c88bd5abd77a8533edce4e8166e77b6eaddd19caedbdf2647118aac5b8a59bb5"},{"id":"func/wallMagnitude","name":"wallMagnitude","line":252,"end_line":254,"hash":"ddae81c4f837e0900b8e7d243ddb53290e4201b63b7f52cc25d60f53f7580860"},{"id":"func/dot","name":"dot","line":256,"end_line":258,"hash":"a98c76b211f97df5c55aa8d4f5f4fe48487b39b54e51515787beec71edaf7979"}]}
// mutate4go-manifest-end
