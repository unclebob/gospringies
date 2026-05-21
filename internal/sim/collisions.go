package sim

import "math"

const fixedMassCollisionRadius = 4

func (s *Simulation) applyMassCollisions() {
	_, ok := s.enabledForce("mass collision")
	if !ok {
		return
	}
	for i := range s.Masses {
		m1 := &s.Masses[i]
		for j := firstCollisionPartnerIndex(i); j < len(s.Masses); j++ {
			m2 := &s.Masses[j]
			s.applyMassCollision(m1, m2)
		}
	}
}

func firstCollisionPartnerIndex(index int) int {
	return index + 1
}

func (s *Simulation) applyMassCollision(m1, m2 *Mass) {
	geometry, ok := collisionGeometryFor(*m1, *m2)
	if !ok {
		return
	}
	m1Velocity := m1.Velocity
	m2Velocity := m2.Velocity
	if collisionVelocitiesSeparating(m1Velocity, m2Velocity, geometry) {
		return
	}
	geometry.avoidVerticalDivision()
	applyCollisionVelocity(m1, *m2, m1Velocity, m2Velocity, geometry)
	applyCollisionVelocity(m2, *m1, m2Velocity, m1Velocity, geometry)
}

type collisionGeometry struct {
	dx     float64
	dy     float64
	dxq    float64
	dyq    float64
	sumxyq float64
}

func collisionGeometryFor(m1, m2 Mass) (collisionGeometry, bool) {
	dx := m2.Position.X - m1.Position.X
	dy := m2.Position.Y - m1.Position.Y
	dxq := dx * dx
	dyq := dy * dy
	sumxyq := dxq + dyq
	if sumxyq == 0 {
		return collisionGeometry{}, false
	}
	if math.Sqrt(sumxyq) >= MassRadius(m1)+MassRadius(m2) {
		return collisionGeometry{}, false
	}
	return collisionGeometry{dx: dx, dy: dy, dxq: dxq, dyq: dyq, sumxyq: sumxyq}, true
}

func collisionVelocitiesSeparating(m1Velocity, m2Velocity Vec2, geometry collisionGeometry) bool {
	return axisVelocitiesSeparating(m1Velocity.X-m2Velocity.X, geometry.dx) &&
		axisVelocitiesSeparating(m1Velocity.Y-m2Velocity.Y, geometry.dy)
}

func axisVelocitiesSeparating(relativeVelocity, displacement float64) bool {
	switch {
	case displacement > 0:
		return relativeVelocity <= 0
	case displacement < 0:
		return relativeVelocity >= 0
	default:
		return true
	}
}

func (g *collisionGeometry) avoidVerticalDivision() {
	if g.dx == 0 {
		g.dx = 1e-10
	}
}

func applyCollisionVelocity(moving *Mass, other Mass, movingVelocity Vec2, otherVelocity Vec2, geometry collisionGeometry) {
	if moving.Fixed {
		return
	}
	ratio := collisionRatio(*moving, other)
	moving.Velocity.X = (movingVelocity.X-(movingVelocity.X-otherVelocity.X)*ratio)*(geometry.dxq/geometry.sumxyq) +
		movingVelocity.X*(geometry.dyq/geometry.sumxyq) -
		(movingVelocity.Y-otherVelocity.Y)*ratio*(geometry.dx*geometry.dy/geometry.sumxyq)
	moving.Velocity.Y = (moving.Velocity.X-movingVelocity.X)*(geometry.dy/geometry.dx) + movingVelocity.Y
}

func collisionRatio(moving, other Mass) float64 {
	elasticity := 1 + (moving.Elasticity+other.Elasticity)/2
	if other.Fixed {
		return elasticity
	}
	return elasticity / (1 + effectiveCollisionMass(moving)/effectiveCollisionMass(other))
}

func effectiveCollisionMass(mass Mass) float64 {
	if mass.Mass == 0 {
		return 1
	}
	return mass.Mass
}

func MassRadius(mass Mass) float64 {
	if mass.Fixed {
		return fixedMassCollisionRadius
	}
	radius := int(2 * math.Log(4*effectiveCollisionMass(mass)+1))
	return math.Min(64, math.Max(1, float64(radius)))
}

func (s *Simulation) applyWallSpringLengthConstraints() {
	for i := range s.Springs {
		aIndex, bIndex, ok := s.wallSpringEndpointIndexes(s.Springs[i])
		if !ok {
			continue
		}
		s.applyWallSpringLengthConstraint(&s.Springs[i], &s.Masses[aIndex], &s.Masses[bIndex])
	}
}

func (s *Simulation) applyWallSpringLengthConstraint(spring *Spring, endpointA, endpointB *Mass) {
	segment := endpointB.Position.Sub(endpointA.Position)
	distance := length(segment)
	if distance == 0 {
		return
	}
	if spring.RestLength <= 0 {
		spring.RestLength = distance
		return
	}
	correction := segment.Normalize().Scale(distance - spring.RestLength)
	applyWallSpringLengthCorrection(endpointA, endpointB, correction)
}

func applyWallSpringLengthCorrection(endpointA, endpointB *Mass, correction Vec2) {
	switch {
	case endpointA.Fixed && endpointB.Fixed:
		return
	case endpointA.Fixed:
		endpointB.Position = endpointB.Position.Sub(correction)
	case endpointB.Fixed:
		endpointA.Position = endpointA.Position.Add(correction)
	default:
		endpointA.Position = endpointA.Position.Add(correction.Scale(0.5))
		endpointB.Position = endpointB.Position.Sub(correction.Scale(0.5))
	}
}

func (s *Simulation) applyWallSpringCollisions(dt float64) {
	if dt <= 0 {
		return
	}
	for _, spring := range s.Springs {
		aIndex, bIndex, ok := s.wallSpringEndpointIndexes(spring)
		if !ok {
			continue
		}
		for i := range s.Masses {
			if s.shouldApplyWallSpringCollision(i, aIndex, bIndex) {
				s.applyWallSpringCollision(spring, &s.Masses[i], &s.Masses[aIndex], &s.Masses[bIndex], dt)
			}
		}
	}
}

func (s *Simulation) wallSpringEndpointIndexes(spring Spring) (int, int, bool) {
	if !spring.Wall {
		return 0, 0, false
	}
	return s.springEndpointIndexes(spring)
}

func (s *Simulation) shouldApplyWallSpringCollision(massIndex int, aIndex int, bIndex int) bool {
	return massIndex != aIndex && massIndex != bIndex && !s.Masses[massIndex].Fixed
}

func (s *Simulation) springEndpointIndexes(spring Spring) (int, int, bool) {
	if spring.MassA != 0 || spring.MassB != 0 {
		a, okA := s.massIndexByID(spring.MassA)
		b, okB := s.massIndexByID(spring.MassB)
		return a, b, okA && okB
	}
	return spring.A, spring.B, s.validSpringMassIndexes(spring)
}

func (s *Simulation) applyWallSpringCollision(spring Spring, mass, endpointA, endpointB *Mass, dt float64) {
	segment := endpointB.Position.Sub(endpointA.Position)
	lengthSquared := dot(segment, segment)
	if lengthSquared == 0 {
		return
	}
	normal := Vec2{X: -segment.Y, Y: segment.X}.Normalize()
	previous := mass.Position.Sub(mass.Velocity.Scale(dt))
	currentSide := dot(mass.Position.Sub(endpointA.Position), normal)
	previousSide := dot(previous.Sub(endpointA.Position), normal)
	contactFraction, ok := wallSpringContactFraction(previous, mass.Position, endpointA.Position, segment, lengthSquared, previousSide, currentSide)
	if !ok {
		return
	}
	side := sideSign(previousSide)
	oldVelocity := mass.Velocity
	contact := closestPointOnSegment(mass.Position, endpointA.Position, segment, lengthSquared)
	mass.Position = contact.Add(normal.Scale(side * MassRadius(*mass)))
	resolveWallSpringVelocity(mass, normal, side)
	s.applyWallSpringTemperatureKick(spring, mass)
	shareWallSpringImpulse(endpointA, oldVelocity.Sub(mass.Velocity).Scale(1-contactFraction))
	shareWallSpringImpulse(endpointB, oldVelocity.Sub(mass.Velocity).Scale(contactFraction))
}

func (s *Simulation) applyWallSpringTemperatureKick(spring Spring, mass *Mass) {
	if spring.Temperature <= 0 {
		return
	}
	temperature := math.Min(10, spring.Temperature)
	angle := s.temperatureRandom().Float64() * 2 * math.Pi
	kick := fullScreenGravityKick(s) * temperature / 10
	mass.Velocity = mass.Velocity.Add(Vec2{X: math.Cos(angle) * kick, Y: math.Sin(angle) * kick})
}

func fullScreenGravityKick(s *Simulation) float64 {
	return math.Sqrt(2 * 10 * s.Bounds.Height)
}

func wallSpringContactFraction(previous, current, start, segment Vec2, lengthSquared float64, previousSide, currentSide float64) (float64, bool) {
	if previousSide == 0 || currentSide == 0 || sameSign(previousSide, currentSide) {
		return 0, false
	}
	intersectionFraction := previousSide / (previousSide - currentSide)
	crossing := previous.Add(current.Sub(previous).Scale(intersectionFraction))
	projection := dot(crossing.Sub(start), segment) / lengthSquared
	return projection, projection >= 0 && projection <= 1
}

func sameSign(a, b float64) bool {
	return (a > 0 && b > 0) || (a < 0 && b < 0)
}

func sideSign(value float64) float64 {
	if value < 0 {
		return -1
	}
	return 1
}

func closestPointOnSegment(point, start, segment Vec2, lengthSquared float64) Vec2 {
	projection := dot(point.Sub(start), segment) / lengthSquared
	return start.Add(segment.Scale(math.Min(1, math.Max(0, projection))))
}

func resolveWallSpringVelocity(mass *Mass, normal Vec2, startingSide float64) {
	normalVelocity := dot(mass.Velocity, normal)
	if normalVelocity*startingSide >= 0 {
		return
	}
	elasticity := 1 + math.Max(1, mass.Elasticity)
	mass.Velocity = mass.Velocity.Sub(normal.Scale(elasticity * normalVelocity))
}

func shareWallSpringImpulse(endpoint *Mass, impulse Vec2) {
	if !endpoint.Fixed {
		endpoint.Velocity = endpoint.Velocity.Add(impulse)
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-21T12:08:30-05:00","module_hash":"da0ac5892f1336c8a3cac511c5cdd32a673c080782db0c1031a8482a388eca98","functions":[{"id":"func/Simulation.applyMassCollisions","name":"Simulation.applyMassCollisions","line":7,"end_line":19,"hash":"5379009637bed15470b5620c7ac9404b7f9365f20d5f79bb612186bc72112cff"},{"id":"func/firstCollisionPartnerIndex","name":"firstCollisionPartnerIndex","line":21,"end_line":23,"hash":"c1b7d5bed0f8810a1fc6b5eff3c2c8e2fe0c00728efb59331a2a798357d75cc7"},{"id":"func/Simulation.applyMassCollision","name":"Simulation.applyMassCollision","line":25,"end_line":38,"hash":"586096d3e011ee19bee0d952e16ac54bdd38a02321ff92dc20c7ab567a5db1c4"},{"id":"func/collisionGeometryFor","name":"collisionGeometryFor","line":48,"end_line":61,"hash":"4aede77bbffb3a3ab973a50e3fd867499c9ddb09ccc28a0622f0061ea6381f72"},{"id":"func/collisionVelocitiesSeparating","name":"collisionVelocitiesSeparating","line":63,"end_line":66,"hash":"f52eae3df0f4825de2a2a3752b1426b6cd90fd6013b578e79d862ac13258adc8"},{"id":"func/axisVelocitiesSeparating","name":"axisVelocitiesSeparating","line":68,"end_line":77,"hash":"b97093dba711234b832aed6df9635b9bfe47361bd9463f57ad1defd61ceb89c8"},{"id":"func/collisionGeometry.avoidVerticalDivision","name":"collisionGeometry.avoidVerticalDivision","line":79,"end_line":83,"hash":"88cccca7591ad7afa29e256d09bfd500d8b8c686cb0ffa3ce77e5ac229a59223"},{"id":"func/applyCollisionVelocity","name":"applyCollisionVelocity","line":85,"end_line":94,"hash":"cf26b30421af198e20ff169771094969463810af774fc1f6d74b3a488f503d7a"},{"id":"func/collisionRatio","name":"collisionRatio","line":96,"end_line":102,"hash":"95d5ae0e55e9f5b6c3190e99f44d99369e80ac845a32f82a654401b73a2e5249"},{"id":"func/effectiveCollisionMass","name":"effectiveCollisionMass","line":104,"end_line":109,"hash":"7e5e2a521ed0f604789ebf8219cc8ea5c730429b3fe678868b30bba111ed9684"},{"id":"func/MassRadius","name":"MassRadius","line":111,"end_line":117,"hash":"3c4415b5dbb666c2192df8b4cd5b580f0565990797c35d6d435ab4f5ccc12bf5"},{"id":"func/Simulation.applyWallSpringCollisions","name":"Simulation.applyWallSpringCollisions","line":119,"end_line":134,"hash":"fce45d2482b3e71404c7d72f0c1d2945d3a4bee12d232e0eb495073a68faff04"},{"id":"func/Simulation.wallSpringEndpointIndexes","name":"Simulation.wallSpringEndpointIndexes","line":136,"end_line":141,"hash":"307c5660a27649ce403eb5e9d332f81c5f8b5cae817b8b743b9150bcfc573d14"},{"id":"func/Simulation.shouldApplyWallSpringCollision","name":"Simulation.shouldApplyWallSpringCollision","line":143,"end_line":145,"hash":"fc1ed28fe790afac15224ddfe7c361fd41b4627bb2be61cf2ffd4c60eb5a5702"},{"id":"func/Simulation.springEndpointIndexes","name":"Simulation.springEndpointIndexes","line":147,"end_line":154,"hash":"28dc42d1c5041984b51daa62a36bef51c4a6008e2cee3fa5ea814a294348deae"},{"id":"func/Simulation.applyWallSpringCollision","name":"Simulation.applyWallSpringCollision","line":156,"end_line":177,"hash":"73b4df9ba95b232658a5383f0782a84585ae10663151989a289dec17874ddd2d"},{"id":"func/wallSpringContactFraction","name":"wallSpringContactFraction","line":179,"end_line":187,"hash":"282d885f85bf1ba65d1dfed0e4d0dde8e362386c2811870ed1628e42854c72c2"},{"id":"func/sameSign","name":"sameSign","line":189,"end_line":191,"hash":"a4155fc319954816bbce383fa5fa6271ca2370c06ce7d19922bff31573a78cda"},{"id":"func/sideSign","name":"sideSign","line":193,"end_line":198,"hash":"298538937752c7335e7161193cda4a5bd2e05511e0081d1f293bd8376e12c0a2"},{"id":"func/closestPointOnSegment","name":"closestPointOnSegment","line":200,"end_line":203,"hash":"0715cf2350f10592aa42cd7fec2c8d37d91e62e3f5f323d31dbb004a4bec942c"},{"id":"func/resolveWallSpringVelocity","name":"resolveWallSpringVelocity","line":205,"end_line":212,"hash":"159c78a4c56c08e9553701a3440e75f19d004718ba3ae69d6913889777fd27c8"},{"id":"func/shareWallSpringImpulse","name":"shareWallSpringImpulse","line":214,"end_line":218,"hash":"83af4feedcb694ce1e08e201c09423bdb359575d8c249808cc6c174dee9927fb"}]}
// mutate4go-manifest-end
