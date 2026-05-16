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
		a := s.Masses[spring.A]
		b := s.Masses[spring.B]
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
	return Vec2{X: magnitude * math.Sin(radians) * mass.Mass, Y: magnitude * math.Cos(radians) * mass.Mass}
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
		if enabled, _ := s.Parameters.WallEnabled(wall.name); enabled && wall.outside {
			total = total.Add(wall.force)
		}
	}
	return total
}

type wallCheck struct {
	name    string
	outside bool
	force   Vec2
}

func (s *Simulation) wallChecks(mass Mass, magnitude float64) []wallCheck {
	exponent := forceExponent(s.Parameters.Forces["wall repulsion"])
	return []wallCheck{
		{name: "top", outside: mass.Position.Y < 0, force: Vec2{Y: wallMagnitude(magnitude, -mass.Position.Y, exponent)}},
		{name: "left", outside: mass.Position.X < 0, force: Vec2{X: wallMagnitude(magnitude, -mass.Position.X, exponent)}},
		{name: "right", outside: mass.Position.X > s.Bounds.Width, force: Vec2{X: -wallMagnitude(magnitude, mass.Position.X-s.Bounds.Width, exponent)}},
		{name: "bottom", outside: mass.Position.Y > s.Bounds.Height, force: Vec2{Y: -wallMagnitude(magnitude, mass.Position.Y-s.Bounds.Height, exponent)}},
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
	if distance <= 0 {
		return magnitude
	}
	return magnitude / math.Pow(distance, exponent)
}

func dot(a, b Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}
