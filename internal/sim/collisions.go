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

func (s *Simulation) applyPostContactReconciliation() {
	s.reconcileEnabledWallContacts()
	s.reconcilePersistentWallSpringContacts()
	s.reconcileEnabledWallContacts()
}

func (s *Simulation) reconcileEnabledWallContacts() {
	for i := range s.Masses {
		s.reconcileEnabledWallContact(&s.Masses[i])
	}
}

func (s *Simulation) reconcileEnabledWallContact(mass *Mass) {
	for _, wall := range s.collisionWalls(mass) {
		enabled, _ := s.Parameters.WallEnabled(wall.name)
		if !enabled || !wall.outside(*wall.position) {
			continue
		}
		*wall.position = wall.boundary
		if wall.movingOutward(*wall.velocity) {
			s.bounceOrStick(mass, wall)
		}
		return
	}
}

func (s *Simulation) reconcilePersistentWallSpringContacts() {
	for range 4 {
		reconciled := false
		for _, spring := range s.Springs {
			aIndex, bIndex, ok := s.wallSpringEndpointIndexes(spring)
			if !ok {
				continue
			}
			for i := range s.Masses {
				if s.shouldApplyWallSpringCollision(i, aIndex, bIndex) {
					reconciled = s.reconcilePersistentWallSpringContact(&s.Masses[i], &s.Masses[aIndex], &s.Masses[bIndex]) || reconciled
				}
			}
		}
		if !reconciled {
			return
		}
	}
}

func (s *Simulation) reconcilePersistentWallSpringContact(mass, endpointA, endpointB *Mass) bool {
	contact, ok := persistentWallSpringContactFor(*mass, *endpointA, *endpointB)
	if !ok {
		return false
	}
	applyPersistentWallSpringPositionCorrection(mass, endpointA, endpointB, contact)
	resolvePersistentWallSpringVelocity(mass, endpointA, endpointB, contact.normal, contact.fraction)
	return true
}

type persistentWallSpringContact struct {
	normal      Vec2
	fraction    float64
	penetration float64
}

func persistentWallSpringContactFor(mass, endpointA, endpointB Mass) (persistentWallSpringContact, bool) {
	segment := endpointB.Position.Sub(endpointA.Position)
	lengthSquared := dot(segment, segment)
	if lengthSquared == 0 {
		return persistentWallSpringContact{}, false
	}
	projection := dot(mass.Position.Sub(endpointA.Position), segment) / lengthSquared
	if projection < 0 || projection > 1 {
		return persistentWallSpringContact{}, false
	}
	normal := Vec2{X: -segment.Y, Y: segment.X}.Normalize()
	side := dot(mass.Position.Sub(endpointA.Position), normal)
	if side == 0 {
		return persistentWallSpringContact{}, false
	}
	radius := MassRadius(mass)
	distance := math.Abs(side)
	if distance >= radius {
		return persistentWallSpringContact{}, false
	}
	return persistentWallSpringContact{
		normal:      normal.Scale(sideSign(side)),
		fraction:    projection,
		penetration: radius - distance,
	}, true
}

func applyPersistentWallSpringPositionCorrection(mass, endpointA, endpointB *Mass, contact persistentWallSpringContact) {
	shareA, shareB, inverseMass := wallSpringContactSharesAndInverseMass(*mass, *endpointA, *endpointB, contact.fraction)
	if inverseMass == 0 {
		return
	}
	correction := contact.normal.Scale(contact.penetration / inverseMass)
	shareWallSpringPositionCorrection(mass, correction)
	shareWallSpringPositionCorrection(endpointA, correction.Scale(-shareA))
	shareWallSpringPositionCorrection(endpointB, correction.Scale(-shareB))
}

func resolvePersistentWallSpringVelocity(mass, endpointA, endpointB *Mass, normal Vec2, contactFraction float64) {
	relativeVelocity := mass.Velocity.Sub(wallSpringContactVelocity(endpointA, endpointB, contactFraction))
	normalVelocity := dot(relativeVelocity, normal)
	if finiteWallSpringCollisionSeparating(normalVelocity) {
		return
	}
	shareA, shareB, inverseMass := wallSpringContactSharesAndInverseMass(*mass, *endpointA, *endpointB, contactFraction)
	if inverseMass == 0 {
		return
	}
	impulse := normal.Scale(-normalVelocity / inverseMass)
	shareWallSpringImpulse(mass, impulse)
	shareWallSpringImpulse(endpointA, impulse.Scale(-shareA))
	shareWallSpringImpulse(endpointB, impulse.Scale(-shareB))
}

func wallSpringContactSharesAndInverseMass(mass, endpointA, endpointB Mass, contactFraction float64) (float64, float64, float64) {
	shareA := 1 - contactFraction
	shareB := contactFraction
	inverseMass := contactShareInverseMass(mass, 1) + contactShareInverseMass(endpointA, shareA) + contactShareInverseMass(endpointB, shareB)
	return shareA, shareB, inverseMass
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
	if endpointA.Fixed && endpointB.Fixed {
		return
	}
	if moveSingleFixedWallSpringEndpoint(endpointA, endpointB, correction) {
		return
	}
	shareWallSpringLengthCorrection(endpointA, endpointB, correction)
}

func moveSingleFixedWallSpringEndpoint(endpointA, endpointB *Mass, correction Vec2) bool {
	if endpointA.Fixed {
		moveWallSpringEndpoint(endpointB, correction.Scale(-1))
		return true
	}
	if endpointB.Fixed {
		moveWallSpringEndpoint(endpointA, correction)
		return true
	}
	return false
}

func shareWallSpringLengthCorrection(endpointA, endpointB *Mass, correction Vec2) {
	half := correction.Scale(0.5)
	moveWallSpringEndpoint(endpointA, half)
	moveWallSpringEndpoint(endpointB, half.Scale(-1))
}

func moveWallSpringEndpoint(endpoint *Mass, correction Vec2) {
	endpoint.Position = endpoint.Position.Add(correction)
}

func (s *Simulation) applyWallSpringLengthConstraintCollisions(dt float64, beforeLengthConstraints []Vec2) {
	if dt <= 0 {
		return
	}
	for sourceIndex, source := range s.Springs {
		aIndex, bIndex, ok := s.wallSpringEndpointIndexes(source)
		if !ok {
			continue
		}
		s.applyWallSpringEndpointConstraintCollisions(sourceIndex, aIndex, beforeLengthConstraints)
		s.applyWallSpringEndpointConstraintCollisions(sourceIndex, bIndex, beforeLengthConstraints)
	}
}

func (s *Simulation) applyWallSpringEndpointConstraintCollisions(sourceSpringIndex int, massIndex int, beforeLengthConstraints []Vec2) {
	if beforeLengthConstraints[massIndex] == s.Masses[massIndex].Position {
		return
	}
	for springIndex, spring := range s.Springs {
		if springIndex == sourceSpringIndex {
			continue
		}
		aIndex, bIndex, ok := s.wallSpringEndpointIndexes(spring)
		if !ok {
			continue
		}
		if !s.shouldApplyWallSpringCollision(massIndex, aIndex, bIndex) {
			continue
		}
		s.applyWallSpringCollision(spring, &s.Masses[massIndex], &s.Masses[aIndex], &s.Masses[bIndex], beforeLengthConstraints[massIndex], beforeLengthConstraints[aIndex], beforeLengthConstraints[bIndex], true)
	}
}

func (s *Simulation) applyWallSpringCollisions(dt float64, startPositions []Vec2) {
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
				previousMass := wallSpringPreviousPosition(s.Masses[i], startPositions, i, dt)
				previousEndpointA := wallSpringPreviousPosition(s.Masses[aIndex], startPositions, aIndex, dt)
				previousEndpointB := wallSpringPreviousPosition(s.Masses[bIndex], startPositions, bIndex, dt)
				allowBoundaryStart := wallSpringBoundaryStartPenetrating(s.Masses[i], s.Masses[aIndex], s.Masses[bIndex], previousMass, previousEndpointA, previousEndpointB)
				s.applyWallSpringCollision(spring, &s.Masses[i], &s.Masses[aIndex], &s.Masses[bIndex], previousMass, previousEndpointA, previousEndpointB, allowBoundaryStart)
			}
		}
	}
}

func wallSpringBoundaryStartPenetrating(mass, endpointA, endpointB Mass, previousMass, previousEndpointA, previousEndpointB Vec2) bool {
	segment := endpointB.Position.Sub(endpointA.Position)
	lengthSquared := dot(segment, segment)
	if lengthSquared == 0 {
		return false
	}
	normal := Vec2{X: -segment.Y, Y: segment.X}.Normalize()
	previousNormal, ok := wallSpringNormal(previousEndpointA, previousEndpointB)
	if !ok {
		return false
	}
	previousSide := dot(previousMass.Sub(previousEndpointA), previousNormal)
	currentSide := dot(mass.Position.Sub(endpointA.Position), normal)
	if previousSide != 0 || currentSide == 0 {
		return false
	}
	contactFraction, ok := wallSpringContactFraction(previousMass.Sub(previousEndpointA), mass.Position.Sub(endpointA.Position), segment, lengthSquared, previousSide, currentSide, true)
	if !ok {
		return false
	}
	wallVelocity := wallSpringContactVelocity(&endpointA, &endpointB, contactFraction)
	normalVelocity := dot(mass.Velocity.Sub(wallVelocity), normal)
	return !wallSpringVelocitySeparating(normalVelocity, collisionStartSide(previousSide, currentSide))
}

func (s *Simulation) applyMovingWallSpringFixedEndpointCollisions(dt float64, startPositions []Vec2) {
	if dt <= 0 {
		return
	}
	for sourceIndex, source := range s.Springs {
		aIndex, bIndex, ok := s.movingWallSpringEndpointIndexes(source)
		if !ok {
			continue
		}
		s.applyMovingWallSpringAgainstFixedEndpoints(sourceIndex, aIndex, bIndex, startPositions)
	}
}

func (s *Simulation) movingWallSpringEndpointIndexes(spring Spring) (int, int, bool) {
	aIndex, bIndex, ok := s.wallSpringEndpointIndexes(spring)
	return aIndex, bIndex, ok && !s.Masses[aIndex].Fixed && !s.Masses[bIndex].Fixed
}

func (s *Simulation) applyMovingWallSpringAgainstFixedEndpoints(sourceIndex, aIndex, bIndex int, startPositions []Vec2) {
	for targetIndex, target := range s.Springs {
		if targetIndex == sourceIndex {
			continue
		}
		targetA, targetB, ok := s.wallSpringEndpointIndexes(target)
		if !ok {
			continue
		}
		s.applyMovingWallSpringFixedEndpointCollision(aIndex, bIndex, targetA, startPositions)
		s.applyMovingWallSpringFixedEndpointCollision(aIndex, bIndex, targetB, startPositions)
	}
}

func (s *Simulation) applyMovingWallSpringFixedEndpointCollision(aIndex, bIndex, fixedIndex int, startPositions []Vec2) {
	if s.skipMovingWallSpringFixedEndpointCollision(aIndex, bIndex, fixedIndex) {
		return
	}
	endpointA := &s.Masses[aIndex]
	endpointB := &s.Masses[bIndex]
	fixed := s.Masses[fixedIndex]
	normal, currentFraction, currentSide, ok := s.movingWallSpringFixedEndpointContact(aIndex, bIndex, fixed, startPositions)
	if !ok || fixedEndpointContactOutside(currentSide) {
		return
	}
	oldVelocity := wallSpringContactVelocity(endpointA, endpointB, currentFraction)
	contactDelta := resolvedFixedEndpointContactVelocity(oldVelocity, normal).Sub(oldVelocity)
	if fixedEndpointContactResolved(contactDelta, currentSide) {
		return
	}
	correction := normal.Scale(fixedMassCollisionRadius - currentSide)
	moveWallSpringEndpoint(endpointA, correction)
	moveWallSpringEndpoint(endpointB, correction)
	shareMovingWallSpringContactImpulse(endpointA, endpointB, contactDelta, currentFraction)
}

func (s *Simulation) skipMovingWallSpringFixedEndpointCollision(aIndex, bIndex, fixedIndex int) bool {
	return fixedIndex == aIndex || fixedIndex == bIndex || !s.Masses[fixedIndex].Fixed
}

func fixedEndpointContactOutside(currentSide float64) bool {
	return currentSide >= fixedMassCollisionRadius
}

func fixedEndpointContactResolved(contactDelta Vec2, currentSide float64) bool {
	return contactDelta == (Vec2{}) && currentSide >= 0
}

func (s *Simulation) movingWallSpringFixedEndpointContact(aIndex, bIndex int, fixed Mass, startPositions []Vec2) (Vec2, float64, float64, bool) {
	normal, ok := s.previousFixedEndpointNormal(aIndex, bIndex, fixed, startPositions)
	if !ok {
		return Vec2{}, 0, 0, false
	}
	currentFraction, currentSide, ok := currentFixedEndpointContact(fixed.Position, s.Masses[aIndex].Position, s.Masses[bIndex].Position, normal)
	return normal, currentFraction, currentSide, ok
}

func (s *Simulation) previousFixedEndpointNormal(aIndex, bIndex int, fixed Mass, startPositions []Vec2) (Vec2, bool) {
	previousA := wallSpringPreviousPosition(s.Masses[aIndex], startPositions, aIndex, 0)
	previousB := wallSpringPreviousPosition(s.Masses[bIndex], startPositions, bIndex, 0)
	previousSegment := previousB.Sub(previousA)
	previousLengthSquared := dot(previousSegment, previousSegment)
	if previousLengthSquared == 0 {
		return Vec2{}, false
	}
	previousContact := closestPointOnSegment(fixed.Position, previousA, previousSegment, previousLengthSquared)
	normal := previousContact.Sub(fixed.Position).Normalize()
	return normal, normal != (Vec2{})
}

func currentFixedEndpointContact(fixedPosition, endpointA, endpointB, normal Vec2) (float64, float64, bool) {
	currentSegment := endpointB.Sub(endpointA)
	currentLengthSquared := dot(currentSegment, currentSegment)
	if currentLengthSquared == 0 {
		return 0, 0, false
	}
	currentFraction := closestFractionOnSegment(fixedPosition, endpointA, currentSegment, currentLengthSquared)
	currentContact := endpointA.Add(currentSegment.Scale(currentFraction))
	return currentFraction, dot(currentContact.Sub(fixedPosition), normal), true
}

func closestFractionOnSegment(point, start, segment Vec2, lengthSquared float64) float64 {
	projection := dot(point.Sub(start), segment) / lengthSquared
	return math.Min(1, math.Max(0, projection))
}

func resolvedFixedEndpointContactVelocity(velocity Vec2, normal Vec2) Vec2 {
	normalVelocity := math.Min(0, dot(velocity, normal))
	return velocity.Sub(normal.Scale(2 * normalVelocity))
}

func shareMovingWallSpringContactImpulse(endpointA, endpointB *Mass, contactDelta Vec2, contactFraction float64) {
	shareA := 1 - contactFraction
	shareB := contactFraction
	inverseMass := contactShareInverseMass(*endpointA, shareA) + contactShareInverseMass(*endpointB, shareB)
	if inverseMass == 0 {
		return
	}
	impulse := contactDelta.Scale(1 / inverseMass)
	shareWallSpringImpulse(endpointA, impulse.Scale(shareA))
	shareWallSpringImpulse(endpointB, impulse.Scale(shareB))
}

func contactShareInverseMass(endpoint Mass, share float64) float64 {
	if endpoint.Fixed {
		return 0
	}
	return share * share / effectiveCollisionMass(endpoint)
}

func wallSpringPreviousPosition(mass Mass, startPositions []Vec2, index int, dt float64) Vec2 {
	if index >= 0 && index < len(startPositions) {
		return startPositions[index]
	}
	return mass.Position.Sub(mass.Velocity.Scale(dt))
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

func (s *Simulation) applyWallSpringCollision(spring Spring, mass, endpointA, endpointB *Mass, previousMass, previousEndpointA, previousEndpointB Vec2, allowBoundaryStart bool) {
	segment := endpointB.Position.Sub(endpointA.Position)
	lengthSquared := dot(segment, segment)
	if lengthSquared == 0 {
		return
	}
	normal := Vec2{X: -segment.Y, Y: segment.X}.Normalize()
	previousNormal, ok := wallSpringNormal(previousEndpointA, previousEndpointB)
	if !ok {
		return
	}
	previousSide := dot(previousMass.Sub(previousEndpointA), previousNormal)
	currentSide := dot(mass.Position.Sub(endpointA.Position), normal)
	contactFraction, side, ok := wallSpringCollisionContact(previousMass.Sub(previousEndpointA), mass.Position.Sub(endpointA.Position), segment, lengthSquared, previousSide, currentSide, MassRadius(*mass), allowBoundaryStart)
	if !ok {
		return
	}
	contact := closestPointOnSegment(mass.Position, endpointA.Position, segment, lengthSquared)
	mass.Position = contact.Add(normal.Scale(side * MassRadius(*mass)))
	resolveFiniteWallSpringCollision(mass, endpointA, endpointB, normal, side, contactFraction)
	s.applyWallSpringTemperatureKick(spring, mass)
}

func wallSpringCollisionContact(previous, current, segment Vec2, lengthSquared float64, previousSide, currentSide, radius float64, allowBoundaryStart bool) (float64, float64, bool) {
	contactFraction, ok := wallSpringContactFraction(previous, current, segment, lengthSquared, previousSide, currentSide, allowBoundaryStart)
	if ok {
		return contactFraction, collisionStartSide(previousSide, currentSide), true
	}
	if currentSide == 0 {
		return 0, 0, false
	}
	side := collisionStartSide(previousSide, currentSide)
	previousBoundarySide := previousSide - radius
	currentBoundarySide := currentSide - radius
	if side < 0 {
		previousBoundarySide = previousSide + radius
		currentBoundarySide = currentSide + radius
	}
	contactFraction, ok = wallSpringContactFraction(previous, current, segment, lengthSquared, previousBoundarySide, currentBoundarySide, allowBoundaryStart)
	return contactFraction, side, ok
}

func wallSpringNormal(endpointA, endpointB Vec2) (Vec2, bool) {
	segment := endpointB.Sub(endpointA)
	if dot(segment, segment) == 0 {
		return Vec2{}, false
	}
	return Vec2{X: -segment.Y, Y: segment.X}.Normalize(), true
}

func wallSpringContactVelocity(endpointA, endpointB *Mass, contactFraction float64) Vec2 {
	return endpointA.Velocity.Scale(1 - contactFraction).Add(endpointB.Velocity.Scale(contactFraction))
}

func resolveFiniteWallSpringCollision(mass, endpointA, endpointB *Mass, normal Vec2, startingSide float64, contactFraction float64) {
	contactNormal := normal.Scale(startingSide)
	relativeVelocity := mass.Velocity.Sub(wallSpringContactVelocity(endpointA, endpointB, contactFraction))
	normalVelocity := dot(relativeVelocity, contactNormal)
	if finiteWallSpringCollisionSeparating(normalVelocity) {
		return
	}
	shareA, shareB, inverseMass := wallSpringContactSharesAndInverseMass(*mass, *endpointA, *endpointB, contactFraction)
	if inverseMass == 0 {
		return
	}
	impulse := contactNormal.Scale(-(1 + wallSpringCollisionElasticity(*mass)) * normalVelocity / inverseMass)
	shareWallSpringImpulse(mass, impulse)
	shareWallSpringImpulse(endpointA, impulse.Scale(-shareA))
	shareWallSpringImpulse(endpointB, impulse.Scale(-shareB))
}

func finiteWallSpringCollisionSeparating(normalVelocity float64) bool {
	return normalVelocity >= 0
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

func wallSpringContactFraction(previous, current, segment Vec2, lengthSquared float64, previousSide, currentSide float64, allowBoundaryStart bool) (float64, bool) {
	if wallSpringCrossingRejected(previousSide, currentSide, allowBoundaryStart) {
		return 0, false
	}
	intersectionFraction := wallSpringIntersectionFraction(previousSide, currentSide)
	crossing := previous.Add(current.Sub(previous).Scale(intersectionFraction))
	projection := dot(crossing, segment) / lengthSquared
	return projection, projection >= 0 && projection <= 1
}

func wallSpringCrossingRejected(previousSide, currentSide float64, allowBoundaryStart bool) bool {
	if currentSide == 0 {
		return true
	}
	if sameSign(previousSide, currentSide) {
		return true
	}
	return previousSide == 0 && !allowBoundaryStart
}

func wallSpringIntersectionFraction(previousSide, currentSide float64) float64 {
	if previousSide == 0 {
		return 0
	}
	return previousSide / (previousSide - currentSide)
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

func collisionStartSide(previousSide, currentSide float64) float64 {
	if previousSide != 0 {
		return sideSign(previousSide)
	}
	return -sideSign(currentSide)
}

func closestPointOnSegment(point, start, segment Vec2, lengthSquared float64) Vec2 {
	projection := dot(point.Sub(start), segment) / lengthSquared
	return start.Add(segment.Scale(math.Min(1, math.Max(0, projection))))
}

func resolveWallSpringVelocity(mass *Mass, wallVelocity Vec2, normal Vec2, startingSide float64) {
	relativeVelocity := mass.Velocity.Sub(wallVelocity)
	normalVelocity := dot(relativeVelocity, normal)
	if wallSpringVelocitySeparating(normalVelocity, startingSide) {
		return
	}
	elasticity := 1 + wallSpringCollisionElasticity(*mass)
	mass.Velocity = wallVelocity.Add(relativeVelocity.Sub(normal.Scale(elasticity * normalVelocity)))
}

func wallSpringCollisionElasticity(mass Mass) float64 {
	if mass.Elasticity > 0 {
		return mass.Elasticity
	}
	return 1
}

func wallSpringVelocitySeparating(normalVelocity float64, startingSide float64) bool {
	return normalVelocity == 0 || sameSign(normalVelocity, startingSide)
}

func shareWallSpringImpulse(endpoint *Mass, impulse Vec2) {
	if !endpoint.Fixed {
		endpoint.Velocity = endpoint.Velocity.Add(impulse.Scale(1 / effectiveCollisionMass(*endpoint)))
	}
}

func shareWallSpringPositionCorrection(endpoint *Mass, correction Vec2) {
	if !endpoint.Fixed {
		endpoint.Position = endpoint.Position.Add(correction.Scale(1 / effectiveCollisionMass(*endpoint)))
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-23T10:47:30-05:00","module_hash":"97d867d3a92d264c579522c38388f148bcc00dc2a1f8adc2681e0006389b1f02","functions":[{"id":"func/Simulation.applyMassCollisions","name":"Simulation.applyMassCollisions","line":7,"end_line":19,"hash":"5379009637bed15470b5620c7ac9404b7f9365f20d5f79bb612186bc72112cff"},{"id":"func/firstCollisionPartnerIndex","name":"firstCollisionPartnerIndex","line":21,"end_line":23,"hash":"c1b7d5bed0f8810a1fc6b5eff3c2c8e2fe0c00728efb59331a2a798357d75cc7"},{"id":"func/Simulation.applyMassCollision","name":"Simulation.applyMassCollision","line":25,"end_line":38,"hash":"586096d3e011ee19bee0d952e16ac54bdd38a02321ff92dc20c7ab567a5db1c4"},{"id":"func/collisionGeometryFor","name":"collisionGeometryFor","line":48,"end_line":61,"hash":"4aede77bbffb3a3ab973a50e3fd867499c9ddb09ccc28a0622f0061ea6381f72"},{"id":"func/collisionVelocitiesSeparating","name":"collisionVelocitiesSeparating","line":63,"end_line":66,"hash":"f52eae3df0f4825de2a2a3752b1426b6cd90fd6013b578e79d862ac13258adc8"},{"id":"func/axisVelocitiesSeparating","name":"axisVelocitiesSeparating","line":68,"end_line":77,"hash":"b97093dba711234b832aed6df9635b9bfe47361bd9463f57ad1defd61ceb89c8"},{"id":"func/collisionGeometry.avoidVerticalDivision","name":"collisionGeometry.avoidVerticalDivision","line":79,"end_line":83,"hash":"88cccca7591ad7afa29e256d09bfd500d8b8c686cb0ffa3ce77e5ac229a59223"},{"id":"func/applyCollisionVelocity","name":"applyCollisionVelocity","line":85,"end_line":94,"hash":"cf26b30421af198e20ff169771094969463810af774fc1f6d74b3a488f503d7a"},{"id":"func/collisionRatio","name":"collisionRatio","line":96,"end_line":102,"hash":"95d5ae0e55e9f5b6c3190e99f44d99369e80ac845a32f82a654401b73a2e5249"},{"id":"func/effectiveCollisionMass","name":"effectiveCollisionMass","line":104,"end_line":109,"hash":"7e5e2a521ed0f604789ebf8219cc8ea5c730429b3fe678868b30bba111ed9684"},{"id":"func/MassRadius","name":"MassRadius","line":111,"end_line":117,"hash":"3c4415b5dbb666c2192df8b4cd5b580f0565990797c35d6d435ab4f5ccc12bf5"},{"id":"func/Simulation.applyWallSpringLengthConstraints","name":"Simulation.applyWallSpringLengthConstraints","line":119,"end_line":127,"hash":"cd0994a247581d6d0df5eb492a8ee281c663a852dfe022a5afe6abdb9b306047"},{"id":"func/Simulation.applyWallSpringLengthConstraint","name":"Simulation.applyWallSpringLengthConstraint","line":129,"end_line":141,"hash":"69e9c99ad19cb2f2119459db1a2def881885636deca5f1f77768038aa11241d1"},{"id":"func/applyWallSpringLengthCorrection","name":"applyWallSpringLengthCorrection","line":143,"end_line":151,"hash":"c81a1c26d3cd3e42494e73d9d4538961ebc4d07e5f6902cd65b270a89334c720"},{"id":"func/moveSingleFixedWallSpringEndpoint","name":"moveSingleFixedWallSpringEndpoint","line":153,"end_line":163,"hash":"cf307e561279f4f910053f4208f243b1a6d908804a569a8d859b94a06e2f28f3"},{"id":"func/shareWallSpringLengthCorrection","name":"shareWallSpringLengthCorrection","line":165,"end_line":169,"hash":"5cbb079415ee5f840a5981b686cfae5c6a33d8399b519e4284cdaff5121d44df"},{"id":"func/moveWallSpringEndpoint","name":"moveWallSpringEndpoint","line":171,"end_line":173,"hash":"5e8f1b7531c4f9654853aa109b55300084da6b08a3b5a1cf4933baca94066b93"},{"id":"func/Simulation.applyWallSpringLengthConstraintCollisions","name":"Simulation.applyWallSpringLengthConstraintCollisions","line":175,"end_line":187,"hash":"29ed684c7db0d608d5651c5c6af6707841c7f00618ff05e8a8753533729f876f"},{"id":"func/Simulation.applyWallSpringEndpointConstraintCollisions","name":"Simulation.applyWallSpringEndpointConstraintCollisions","line":189,"end_line":206,"hash":"a1cbb464d5e83903c68ee16c0a3fb52cfb437f3cda01078609bc9a3a530f3256"},{"id":"func/Simulation.applyWallSpringCollisions","name":"Simulation.applyWallSpringCollisions","line":208,"end_line":227,"hash":"4939345a937b1f8ea555e60beeec9ba010e7b360ec188af95f57e7dd5a661044"},{"id":"func/wallSpringBoundaryStartPenetrating","name":"wallSpringBoundaryStartPenetrating","line":229,"end_line":252,"hash":"143505257f7ece79e82099e481d6d0fea85714569943483a355ed568b1228ca1"},{"id":"func/Simulation.applyMovingWallSpringFixedEndpointCollisions","name":"Simulation.applyMovingWallSpringFixedEndpointCollisions","line":254,"end_line":265,"hash":"c9d68269daafeb253406edbc9a03eca2f22483a47694bde50d30708144845017"},{"id":"func/Simulation.movingWallSpringEndpointIndexes","name":"Simulation.movingWallSpringEndpointIndexes","line":267,"end_line":270,"hash":"b76c648217919a1005485e6258d3bae16006545b064aafb61ba02d5e4952d3f2"},{"id":"func/Simulation.applyMovingWallSpringAgainstFixedEndpoints","name":"Simulation.applyMovingWallSpringAgainstFixedEndpoints","line":272,"end_line":284,"hash":"20dd250ec3e631f61e3ee0647a66560dff90fae27dd890982075a97f474a1c15"},{"id":"func/Simulation.applyMovingWallSpringFixedEndpointCollision","name":"Simulation.applyMovingWallSpringFixedEndpointCollision","line":286,"end_line":306,"hash":"8c14e4cbc53a198ec902c7695619607b8f4b942ed7c5e6c196b5e9f9716cb073"},{"id":"func/Simulation.skipMovingWallSpringFixedEndpointCollision","name":"Simulation.skipMovingWallSpringFixedEndpointCollision","line":308,"end_line":310,"hash":"649d071e2e14b1e073a7f0a1bb9fe60f89e50a61e718904fe5ccf53a183976bc"},{"id":"func/fixedEndpointContactOutside","name":"fixedEndpointContactOutside","line":312,"end_line":314,"hash":"35d44904cfcafb83c4cbcf1c057a76cfcb7aed7d515f4bbccd566c3cc23263ba"},{"id":"func/fixedEndpointContactResolved","name":"fixedEndpointContactResolved","line":316,"end_line":318,"hash":"3862639c14152150c5a830cbbd1d32173547c107e239aee4aa2b6db630aeb24b"},{"id":"func/Simulation.movingWallSpringFixedEndpointContact","name":"Simulation.movingWallSpringFixedEndpointContact","line":320,"end_line":327,"hash":"98218cf33adaf0bba14b1e47a31abafe52fa2648e2116304b99f13ef4d7143af"},{"id":"func/Simulation.previousFixedEndpointNormal","name":"Simulation.previousFixedEndpointNormal","line":329,"end_line":340,"hash":"2ad30017dadb2a6f75b8f1bf0dd2c37ff1a2014125565d151299e70999326fe2"},{"id":"func/currentFixedEndpointContact","name":"currentFixedEndpointContact","line":342,"end_line":351,"hash":"ec63a745ef481f81dc9ab214f6e79dee2f6ccff66ac9c903a2e32f21661d3b00"},{"id":"func/closestFractionOnSegment","name":"closestFractionOnSegment","line":353,"end_line":356,"hash":"d1edeb9d58e93dce9aa32398ae0bc9c30686a024027c87e29d16cb9f0b01eb69"},{"id":"func/resolvedFixedEndpointContactVelocity","name":"resolvedFixedEndpointContactVelocity","line":358,"end_line":361,"hash":"9599dc47df7f4e9233d8ed61207791ac9912134ee8f31464bc81a759a073c29f"},{"id":"func/shareMovingWallSpringContactImpulse","name":"shareMovingWallSpringContactImpulse","line":363,"end_line":373,"hash":"253914751bba2bea264fe4171bd9b6b6755a4469e06a116c7f223d9c2e699c62"},{"id":"func/contactShareInverseMass","name":"contactShareInverseMass","line":375,"end_line":380,"hash":"d98368a0621e77fbfd24b471bb70bf770769a4d1607d7dac51cde0709a556fea"},{"id":"func/wallSpringPreviousPosition","name":"wallSpringPreviousPosition","line":382,"end_line":387,"hash":"6a4a9c259c5c91ba2f9d51348efd5f84ce96e1b3b4600e82e4a9da5e1da884fd"},{"id":"func/Simulation.wallSpringEndpointIndexes","name":"Simulation.wallSpringEndpointIndexes","line":389,"end_line":394,"hash":"307c5660a27649ce403eb5e9d332f81c5f8b5cae817b8b743b9150bcfc573d14"},{"id":"func/Simulation.shouldApplyWallSpringCollision","name":"Simulation.shouldApplyWallSpringCollision","line":396,"end_line":398,"hash":"fc1ed28fe790afac15224ddfe7c361fd41b4627bb2be61cf2ffd4c60eb5a5702"},{"id":"func/Simulation.springEndpointIndexes","name":"Simulation.springEndpointIndexes","line":400,"end_line":407,"hash":"28dc42d1c5041984b51daa62a36bef51c4a6008e2cee3fa5ea814a294348deae"},{"id":"func/Simulation.applyWallSpringCollision","name":"Simulation.applyWallSpringCollision","line":409,"end_line":430,"hash":"879e2b3bf1fc96651ab0e0f35e1b50d31e797acdd0f606fe213857fcb23ab679"},{"id":"func/wallSpringCollisionContact","name":"wallSpringCollisionContact","line":432,"end_line":449,"hash":"9327a22d12499ac39a884685316caec484325cbe60e3c63a1badf8d7af9c17f8"},{"id":"func/wallSpringNormal","name":"wallSpringNormal","line":451,"end_line":457,"hash":"ad94ad425b5345ea5f130fc76c38ad15088db9790b3bc7e696919005d1d489ae"},{"id":"func/wallSpringContactVelocity","name":"wallSpringContactVelocity","line":459,"end_line":461,"hash":"9e4131292f82d473684bafeb206784e8b8de66b5671c8bc3b390e0a7e04efc3c"},{"id":"func/resolveFiniteWallSpringCollision","name":"resolveFiniteWallSpringCollision","line":463,"end_line":480,"hash":"3442aeacce6a73d4240d2ff5756c01ade360d2ce3a65a6b9ef2b082a7517dbc2"},{"id":"func/finiteWallSpringCollisionSeparating","name":"finiteWallSpringCollisionSeparating","line":482,"end_line":484,"hash":"00562ff41992ace044b29e0ca5c80ac38f522d75815e2f554dca6a60fae938fc"},{"id":"func/Simulation.applyWallSpringTemperatureKick","name":"Simulation.applyWallSpringTemperatureKick","line":486,"end_line":494,"hash":"ee8e22bc2fa3f4adce4a27ae706b5a56f31bfa17e29955aee48c40ca77a35cf2"},{"id":"func/fullScreenGravityKick","name":"fullScreenGravityKick","line":496,"end_line":498,"hash":"b759a6d17df5f312ddc455452e72952a610374c7eec99bac660cb989e900836e"},{"id":"func/wallSpringContactFraction","name":"wallSpringContactFraction","line":500,"end_line":508,"hash":"dc34dee4a0a55dccc7aafe471d69d6079a4eee4fa1da663d72de54f085151058"},{"id":"func/wallSpringCrossingRejected","name":"wallSpringCrossingRejected","line":510,"end_line":518,"hash":"29034c13e7de7e7d99b37d0002e074bac90270988d356188c6cf1254ae72974f"},{"id":"func/wallSpringIntersectionFraction","name":"wallSpringIntersectionFraction","line":520,"end_line":525,"hash":"93e86e1f22a4fe3febfac13940e372715c087ff01970e3ef8d76e626bad43eba"},{"id":"func/sameSign","name":"sameSign","line":527,"end_line":529,"hash":"a4155fc319954816bbce383fa5fa6271ca2370c06ce7d19922bff31573a78cda"},{"id":"func/sideSign","name":"sideSign","line":531,"end_line":536,"hash":"298538937752c7335e7161193cda4a5bd2e05511e0081d1f293bd8376e12c0a2"},{"id":"func/collisionStartSide","name":"collisionStartSide","line":538,"end_line":543,"hash":"8230beac044869aaf2a79e8522c9570cc0a1a079da871d193b86d6bf94bbee14"},{"id":"func/closestPointOnSegment","name":"closestPointOnSegment","line":545,"end_line":548,"hash":"0715cf2350f10592aa42cd7fec2c8d37d91e62e3f5f323d31dbb004a4bec942c"},{"id":"func/resolveWallSpringVelocity","name":"resolveWallSpringVelocity","line":550,"end_line":558,"hash":"468a38cefa849c2f9741e124df5299b9f3f579e640c21b0f68e3bea165b7e85a"},{"id":"func/wallSpringCollisionElasticity","name":"wallSpringCollisionElasticity","line":560,"end_line":565,"hash":"3e14c8f95c00a57b977e4ce3d1ae590122b7ac8dcb5e4fa36f12a41ef123181e"},{"id":"func/wallSpringVelocitySeparating","name":"wallSpringVelocitySeparating","line":567,"end_line":569,"hash":"b289c63f8350c06a397e29480f78e5acf6b505e5ba3e8dad054f8898992e88ff"},{"id":"func/shareWallSpringImpulse","name":"shareWallSpringImpulse","line":571,"end_line":575,"hash":"cef5c2e6d22bfa3a5b5709f70e75cead82fa01b9cc622a4ee7d933e5434fec5c"}]}
// mutate4go-manifest-end
