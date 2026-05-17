package sim

import "math"

type wallCollision struct {
	name           string
	position       *float64
	velocity       *float64
	boundary       float64
	outside        func(float64) bool
	movingOutward  func(float64) bool
	releaseForce   func(Vec2) float64
	keepTangential func(*Mass)
}

func (s *Simulation) applyWallCollision(mass *Mass) {
	for _, wall := range s.collisionWalls(mass) {
		if !s.wallCollisionActive(wall) {
			continue
		}
		*wall.position = wall.boundary
		s.bounceOrStick(mass, wall)
		return
	}
}

func (s *Simulation) wallCollisionActive(wall wallCollision) bool {
	enabled, _ := s.Parameters.WallEnabled(wall.name)
	return enabled && wall.outside(*wall.position) && wall.movingOutward(*wall.velocity)
}

func (s *Simulation) bounceOrStick(mass *Mass, wall wallCollision) {
	rebound := math.Abs(*wall.velocity)*mass.Elasticity - parameterFloat(s.Parameters, "stickiness")
	if rebound <= 0 {
		*wall.velocity = 0
		mass.StuckWall = wall.name
		return
	}
	*wall.velocity = signedRebound(rebound, wall)
	mass.StuckWall = ""
}

func signedRebound(rebound float64, wall wallCollision) float64 {
	switch wall.name {
	case "right", "bottom":
		return -rebound
	default:
		return rebound
	}
}

func (s *Simulation) keepStuck(mass *Mass, acceleration Vec2) bool {
	if mass.StuckWall == "" {
		return false
	}
	wall, ok := s.stuckWall(mass)
	if !ok || s.wallReleasedBy(wall, acceleration) {
		mass.StuckWall = ""
		return false
	}
	*wall.position = wall.boundary
	wall.keepTangential(mass)
	return true
}

func (s *Simulation) wallReleasedBy(wall wallCollision, acceleration Vec2) bool {
	return wall.releaseForce(acceleration) > parameterFloat(s.Parameters, "stickiness")
}

func (s *Simulation) stuckWall(mass *Mass) (wallCollision, bool) {
	for _, wall := range s.collisionWalls(mass) {
		if wall.name == mass.StuckWall {
			return wall, true
		}
	}
	return wallCollision{}, false
}

func (s *Simulation) collisionWalls(mass *Mass) []wallCollision {
	return []wallCollision{
		{
			name: "left", position: &mass.Position.X, velocity: &mass.Velocity.X, boundary: 0,
			outside: func(position float64) bool { return position < 0 }, movingOutward: func(velocity float64) bool { return velocity < 0 },
			releaseForce:   func(force Vec2) float64 { return force.X },
			keepTangential: func(mass *Mass) { mass.Velocity.X = 0 },
		},
		{
			name: "right", position: &mass.Position.X, velocity: &mass.Velocity.X, boundary: s.Bounds.Width,
			outside: func(position float64) bool { return position > s.Bounds.Width }, movingOutward: func(velocity float64) bool { return velocity > 0 },
			releaseForce:   func(force Vec2) float64 { return -force.X },
			keepTangential: func(mass *Mass) { mass.Velocity.X = 0 },
		},
		{
			name: "top", position: &mass.Position.Y, velocity: &mass.Velocity.Y, boundary: 0,
			outside: func(position float64) bool { return position < 0 }, movingOutward: func(velocity float64) bool { return velocity < 0 },
			releaseForce:   func(force Vec2) float64 { return force.Y },
			keepTangential: func(mass *Mass) { mass.Velocity.Y = 0 },
		},
		{
			name: "bottom", position: &mass.Position.Y, velocity: &mass.Velocity.Y, boundary: s.Bounds.Height,
			outside: func(position float64) bool { return position > s.Bounds.Height }, movingOutward: func(velocity float64) bool { return velocity > 0 },
			releaseForce:   func(force Vec2) float64 { return -force.Y },
			keepTangential: func(mass *Mass) { mass.Velocity.Y = 0 },
		},
	}
}
