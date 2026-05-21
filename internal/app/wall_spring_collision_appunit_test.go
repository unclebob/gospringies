//go:build appunit

package app

import (
	"testing"

	"springs/internal/sim"
)

func TestAppUnitWallSpringCollisionStopsPenetration(t *testing.T) {
	world := appUnitWallSpringCollisionWorld(false, false, -5, 10, 50)

	world.Step(1)

	mass, _ := world.MassByID(3)
	if mass.Position.X > 0 {
		t.Fatalf("mass crossed wall spring: %#v", mass)
	}
	if mass.Velocity.X > 0 {
		t.Fatalf("mass velocity still penetrates wall spring: %#v", mass.Velocity)
	}
}

func TestAppUnitWallSpringCollisionPlacesMassOnStartingSide(t *testing.T) {
	for _, test := range []struct {
		name     string
		startX   float64
		velocity float64
		wantX    float64
	}{
		{name: "left to right", startX: -5, velocity: 10, wantX: -sim.MassRadius(sim.Mass{Mass: 1})},
		{name: "right to left", startX: 5, velocity: -10, wantX: sim.MassRadius(sim.Mass{Mass: 1})},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := appUnitWallSpringCollisionWorld(false, false, test.startX, test.velocity, 50)

			world.Step(1)

			mass, _ := world.MassByID(3)
			if mass.Position.X < test.wantX-0.000001 || mass.Position.X > test.wantX+0.000001 {
				t.Fatalf("mass x = %f, want %f; mass=%#v", mass.Position.X, test.wantX, mass)
			}
		})
	}
}

func TestAppUnitWallSpringCollisionSharesImpulseByContactFraction(t *testing.T) {
	for _, test := range []struct {
		name     string
		contactY float64
		wantA    float64
		wantB    float64
	}{
		{name: "endpoint a", contactY: 0, wantA: 1, wantB: 0},
		{name: "quarter", contactY: 25, wantA: 0.75, wantB: 0.25},
		{name: "middle", contactY: 50, wantA: 0.50, wantB: 0.50},
		{name: "three quarters", contactY: 75, wantA: 0.25, wantB: 0.75},
		{name: "endpoint b", contactY: 100, wantA: 0, wantB: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := appUnitWallSpringCollisionWorld(false, false, -5, 10, test.contactY)

			world.Step(1)

			assertAppUnitWallSpringEndpointShare(t, world, 1, test.wantA)
			assertAppUnitWallSpringEndpointShare(t, world, 2, test.wantB)
		})
	}
}

func TestAppUnitWallSpringCollisionSupportsIndexBackedSprings(t *testing.T) {
	world := appUnitWallSpringCollisionWorld(false, false, -5, 10, 50)
	world.Springs[0] = sim.Spring{ID: 1, A: 0, B: 1, Wall: true}

	world.Step(1)

	assertAppUnitWallSpringEndpointShare(t, world, 1, 0.50)
	assertAppUnitWallSpringEndpointShare(t, world, 2, 0.50)
}

func TestAppUnitWallSpringCollisionSupportsUnitLengthSprings(t *testing.T) {
	world := appUnitWallSpringCollisionWorld(false, false, -5, 10, 0.5)
	world.Masses[1].Position = sim.Vec2{X: 0, Y: 1}

	world.Step(1)

	assertAppUnitWallSpringEndpointShare(t, world, 1, 0.50)
	assertAppUnitWallSpringEndpointShare(t, world, 2, 0.50)
}

func TestAppUnitWallSpringCollisionDoesNotMoveFixedEndpoint(t *testing.T) {
	world := appUnitWallSpringCollisionWorld(true, false, -5, 10, 25)

	world.Step(1)

	fixed, _ := world.MassByID(1)
	free, _ := world.MassByID(2)
	if fixed.Velocity != (sim.Vec2{}) {
		t.Fatalf("fixed endpoint velocity = %#v", fixed.Velocity)
	}
	if free.Velocity.X <= 0 || free.Velocity.X >= 10 {
		t.Fatalf("free endpoint velocity = %#v, expected weighted impulse share", free.Velocity)
	}
}

func TestAppUnitWallSpringCollisionIgnoresInvalidCases(t *testing.T) {
	for _, test := range []struct {
		name  string
		world *sim.Simulation
	}{
		{name: "zero dt", world: appUnitWallSpringCollisionWorld(false, false, -5, 10, 50)},
		{name: "non wall spring", world: appUnitNonWallSpringCollisionWorld()},
		{name: "fixed moving mass", world: appUnitFixedMovingMassWallSpringWorld()},
		{name: "zero length spring", world: appUnitZeroLengthWallSpringWorld()},
		{name: "outside segment", world: appUnitWallSpringCollisionWorld(false, false, -5, 10, 150)},
		{name: "starts on wall", world: appUnitWallSpringCollisionWorld(false, false, 0, 10, 50)},
		{name: "ends on wall", world: appUnitWallSpringCollisionWorld(false, false, -5, 5, 50)},
	} {
		t.Run(test.name, func(t *testing.T) {
			before, _ := test.world.MassByID(3)
			if test.name == "zero dt" {
				test.world.Step(0)
			} else {
				test.world.Step(1)
			}
			after, _ := test.world.MassByID(3)
			if after.Velocity != before.Velocity {
				t.Fatalf("moving mass velocity changed from %#v to %#v", before.Velocity, after.Velocity)
			}
		})
	}
}

func appUnitWallSpringCollisionWorld(fixedA bool, fixedB bool, x float64, velocityX float64, contactY float64) *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1, Fixed: fixedA})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 0, Y: 100}, Mass: 1, Fixed: fixedB})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: x, Y: contactY}, Velocity: sim.Vec2{X: velocityX}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	return world
}

func appUnitNonWallSpringCollisionWorld() *sim.Simulation {
	world := appUnitWallSpringCollisionWorld(false, false, -5, 10, 50)
	world.Springs[0].Wall = false
	return world
}

func appUnitFixedMovingMassWallSpringWorld() *sim.Simulation {
	world := appUnitWallSpringCollisionWorld(false, false, -5, 10, 50)
	world.Masses[2].Fixed = true
	return world
}

func appUnitZeroLengthWallSpringWorld() *sim.Simulation {
	world := appUnitWallSpringCollisionWorld(false, false, -5, 10, 50)
	world.Masses[1].Position = world.Masses[0].Position
	return world
}

func assertAppUnitWallSpringEndpointShare(t *testing.T, world *sim.Simulation, id int, want float64) {
	t.Helper()
	endpoint, _ := world.MassByID(id)
	moving, _ := world.MassByID(3)
	impulse := 10 - moving.Velocity.X
	if impulse == 0 {
		t.Fatal("moving mass received no collision impulse")
	}
	got := endpoint.Velocity.X / impulse
	if got < want-0.000001 || got > want+0.000001 {
		t.Fatalf("endpoint %d impulse share = %f, want %f; endpoint=%#v moving=%#v", id, got, want, endpoint, moving)
	}
}
