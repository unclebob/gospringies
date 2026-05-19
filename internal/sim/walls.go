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
	case "right", "top":
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
			name: "bottom", position: &mass.Position.Y, velocity: &mass.Velocity.Y, boundary: bottomWallBoundary(),
			outside: func(position float64) bool { return position < 0 }, movingOutward: func(velocity float64) bool { return velocity < 0 },
			releaseForce:   func(force Vec2) float64 { return force.Y },
			keepTangential: func(mass *Mass) { mass.Velocity.Y = 0 },
		},
		{
			name: "top", position: &mass.Position.Y, velocity: &mass.Velocity.Y, boundary: s.Bounds.Height,
			outside: func(position float64) bool { return position > s.Bounds.Height }, movingOutward: func(velocity float64) bool { return velocity > 0 },
			releaseForce:   func(force Vec2) float64 { return -force.Y },
			keepTangential: func(mass *Mass) { mass.Velocity.Y = 0 },
		},
	}
}

func bottomWallBoundary() float64 {
	return 0
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T09:37:02-05:00","module_hash":"19a447728d671933907216492c43db9000e8c534afa1b3b58e6932073cba17de","functions":[{"id":"func/Simulation.applyWallCollision","name":"Simulation.applyWallCollision","line":16,"end_line":25,"hash":"4f7eb03b8d0658e370963334da089e344873888adce86c59be7526e2f6e09e69"},{"id":"func/Simulation.wallCollisionActive","name":"Simulation.wallCollisionActive","line":27,"end_line":30,"hash":"f7cd7c9555e1c64105ed74815f54eb3097ec636d6002698f184b8385922e9c45"},{"id":"func/Simulation.bounceOrStick","name":"Simulation.bounceOrStick","line":32,"end_line":41,"hash":"cb2b00bfce27afc8b7272231ad4fdf48afe0c6f17ae0076a349fce61011af65e"},{"id":"func/signedRebound","name":"signedRebound","line":43,"end_line":50,"hash":"d2acc0d1e3f652248aa91cd60186dc10969dbfe6acb39b0aee2d57e4977677c8"},{"id":"func/Simulation.keepStuck","name":"Simulation.keepStuck","line":52,"end_line":64,"hash":"238f27e1220fcbda0f592a14f397d99762079ef8f2b25867790bbbf2e1a939d2"},{"id":"func/Simulation.wallReleasedBy","name":"Simulation.wallReleasedBy","line":66,"end_line":68,"hash":"0d0a5d4dc4fd1c5e0845a01c55872b456db84dc44b5cf0e7de8f20d2c7a7dd93"},{"id":"func/Simulation.stuckWall","name":"Simulation.stuckWall","line":70,"end_line":77,"hash":"208a0cf7c4d11d5c23f0d1b74017679a4aa79c36593b8ee8e80d712d33fad178"},{"id":"func/Simulation.collisionWalls","name":"Simulation.collisionWalls","line":79,"end_line":106,"hash":"23c3716d58be907ffccc617668549a556d4819a1cc4f4967b47349674bf8c0b3"},{"id":"func/bottomWallBoundary","name":"bottomWallBoundary","line":108,"end_line":110,"hash":"eeb11f45cd98bec717ca587855ac7e094c082281315d331dc59a4d183de94153"}]}
// mutate4go-manifest-end
