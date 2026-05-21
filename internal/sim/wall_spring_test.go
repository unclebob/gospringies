package sim

import "testing"

func TestWallSpringStopsMassCrossingSegment(t *testing.T) {
	world := wallSpringCollisionWorld(false, false)

	world.Step(1)

	mass, _ := world.MassByID(3)
	if mass.Position.X > 0 {
		t.Fatalf("mass crossed wall spring: %#v", mass)
	}
	if mass.Velocity.X > 0 {
		t.Fatalf("mass velocity still penetrates wall spring: %#v", mass.Velocity)
	}
}

func TestWallSpringSharesResponseWithUnfixedEndpoints(t *testing.T) {
	world := wallSpringCollisionWorld(false, false)

	world.Step(1)

	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if a.Velocity.X <= 0 || b.Velocity.X <= 0 {
		t.Fatalf("endpoint velocities = %#v %#v, expected shared impulse", a.Velocity, b.Velocity)
	}
	if a.Velocity.X != b.Velocity.X {
		t.Fatalf("endpoint velocities = %#v %#v, expected equal shares", a.Velocity, b.Velocity)
	}
}

func TestWallSpringDoesNotMoveFixedEndpoint(t *testing.T) {
	world := wallSpringCollisionWorld(true, false)

	world.Step(1)

	fixed, _ := world.MassByID(1)
	free, _ := world.MassByID(2)
	if fixed.Velocity != (Vec2{}) {
		t.Fatalf("fixed endpoint velocity = %#v", fixed.Velocity)
	}
	if free.Velocity.X <= 0 {
		t.Fatalf("free endpoint velocity = %#v, expected impulse share", free.Velocity)
	}
}

func wallSpringCollisionWorld(fixedA bool, fixedB bool) *Simulation {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1, Fixed: fixedA})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 100}, Mass: 1, Fixed: fixedB})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -5, Y: 50}, Velocity: Vec2{X: 10}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	return world
}
