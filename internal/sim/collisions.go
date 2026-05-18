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
	dx := m2.Position.X - m1.Position.X
	dy := m2.Position.Y - m1.Position.Y
	dxq := dx * dx
	dyq := dy * dy
	sumxyq := dxq + dyq
	if sumxyq == 0 {
		return
	}
	if math.Sqrt(sumxyq) >= MassRadius(*m1)+MassRadius(*m2) {
		return
	}
	m1vx := m1.Velocity.X
	m1vy := m1.Velocity.Y
	m2vx := m2.Velocity.X
	m2vy := m2.Velocity.Y
	if (m1vx-m2vx)*dx <= 0 && (m1vy-m2vy)*dy <= 0 {
		return
	}
	if dx == 0 {
		dx = 1e-10
	}
	if !m1.Fixed {
		ratio := collisionRatio(*m1, *m2)
		m1.Velocity.X = (m1vx-(m1vx-m2vx)*ratio)*(dxq/sumxyq) +
			m1vx*(dyq/sumxyq) -
			(m1vy-m2vy)*ratio*(dx*dy/sumxyq)
		m1.Velocity.Y = (m1.Velocity.X-m1vx)*(dy/dx) + m1vy
	}
	if !m2.Fixed {
		ratio := collisionRatio(*m2, *m1)
		m2.Velocity.X = (m2vx-(m2vx-m1vx)*ratio)*(dxq/sumxyq) +
			m2vx*(dyq/sumxyq) -
			(m2vy-m1vy)*ratio*(dx*dy/sumxyq)
		m2.Velocity.Y = (m2.Velocity.X-m2vx)*(dy/dx) + m2vy
	}
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
