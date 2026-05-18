package sim

import "math"

const fixedMassCollisionRadius = 4

func (s *Simulation) applyMassCollisions() {
	force, ok := s.enabledForce("mass collision")
	if !ok || force.Enabled != "true" {
		return
	}
	for i := range s.Masses {
		m1 := &s.Masses[i]
		for j := i + 1; j < len(s.Masses); j++ {
			m2 := &s.Masses[j]
			s.applyMassCollision(m1, m2)
		}
	}
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
	return (m1Velocity.X-m2Velocity.X)*geometry.dx <= 0 && (m1Velocity.Y-m2Velocity.Y)*geometry.dy <= 0
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
	if radius < 1 {
		radius = 1
	}
	if radius > 64 {
		radius = 64
	}
	return float64(radius)
}
