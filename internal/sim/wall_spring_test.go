package sim

import (
	"math"
	"math/rand"
	"testing"
)

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

func TestWallSpringSharesResponseByContactFraction(t *testing.T) {
	world := wallSpringCollisionWorld(false, false, 25)

	world.Step(1)

	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if a.Velocity.X <= 0 || b.Velocity.X <= 0 {
		t.Fatalf("endpoint velocities = %#v %#v, expected shared impulse", a.Velocity, b.Velocity)
	}
	if a.Velocity.X <= b.Velocity.X {
		t.Fatalf("endpoint velocities = %#v %#v, expected endpoint A to receive larger share", a.Velocity, b.Velocity)
	}
	ratio := b.Velocity.X / a.Velocity.X
	if ratio < 0.32 || ratio > 0.35 {
		t.Fatalf("endpoint velocity ratio = %f, expected 0.25/0.75", ratio)
	}
}

func TestWallSpringDoesNotMoveFixedEndpoint(t *testing.T) {
	world := wallSpringCollisionWorld(true, false, 25)

	world.Step(1)

	fixed, _ := world.MassByID(1)
	free, _ := world.MassByID(2)
	if fixed.Velocity != (Vec2{}) {
		t.Fatalf("fixed endpoint velocity = %#v", fixed.Velocity)
	}
	if free.Velocity.X <= 0 || free.Velocity.X >= 10 {
		t.Fatalf("free endpoint velocity = %#v, expected impulse share", free.Velocity)
	}
}

func TestWallSpringTemperatureKicksCollidingMass(t *testing.T) {
	for _, temperature := range []float64{10, 1} {
		world := wallSpringCollisionWorld(false, false, 50)
		world.Springs[0].Temperature = temperature
		seed := int64(11)
		world.SetTemperatureSeed(seed)

		world.Step(1)

		assertWallSpringTemperatureKick(t, world, temperature, seed, "temperature kick")
	}
}

func expectedTemperatureKick(height float64, temperature float64, seed int64) Vec2 {
	angle := rand.New(rand.NewSource(seed)).Float64() * 2 * math.Pi
	kick := math.Sqrt(2*10*height) * temperature / 10
	return Vec2{X: math.Cos(angle) * kick, Y: math.Sin(angle) * kick}
}

func assertWallSpringTemperatureKick(t *testing.T, world *Simulation, temperature float64, seed int64, description string) {
	t.Helper()
	mass, _ := world.MassByID(3)
	kick := mass.Velocity.Sub(Vec2{X: -10})
	expected := expectedTemperatureKick(world.Bounds.Height, temperature, seed)
	if !closeWallSpringLength(kick.X, expected.X) || !closeWallSpringLength(kick.Y, expected.Y) {
		t.Fatalf("%s = %#v, expected %#v", description, kick, expected)
	}
}

func TestWallSpringTemperatureKickDoesNotChangeEndpointImpulseShare(t *testing.T) {
	world := wallSpringCollisionWorld(false, false, 50)
	world.Springs[0].Temperature = 10
	world.SetTemperatureSeed(11)

	world.Step(1)

	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if !closeWallSpringLength(a.Velocity.X, 10) || !closeWallSpringLength(a.Velocity.Y, 0) {
		t.Fatalf("endpoint A velocity = %#v, expected collision impulse only", a.Velocity)
	}
	if !closeWallSpringLength(b.Velocity.X, 10) || !closeWallSpringLength(b.Velocity.Y, 0) {
		t.Fatalf("endpoint B velocity = %#v, expected collision impulse only", b.Velocity)
	}
}

func TestWallSpringTemperatureZeroAndNonWallApplyNoKick(t *testing.T) {
	for _, wall := range []bool{true, false} {
		world := wallSpringCollisionWorld(false, false, 50)
		world.Springs[0].Wall = wall
		world.Springs[0].Temperature = 0
		if !wall {
			world.Springs[0].Temperature = 10
		}
		world.SetTemperatureSeed(11)

		world.Step(1)

		mass, _ := world.MassByID(3)
		expectedX := 10.0
		if wall {
			expectedX = -10
		}
		if !closeWallSpringLength(mass.Velocity.X, expectedX) || !closeWallSpringLength(mass.Velocity.Y, 0) {
			t.Fatalf("wall=%t velocity=%#v, expected no temperature kick", wall, mass.Velocity)
		}
	}
}

func TestWallSpringTemperatureZeroDoesNotAdvanceRandomSource(t *testing.T) {
	world := wallSpringCollisionWorld(false, false, 50)
	seed := int64(11)
	world.SetTemperatureSeed(seed)

	world.Step(1)
	resetWallSpringCollisionWorld(world)
	world.Springs[0].Temperature = 10
	world.Step(1)

	assertWallSpringTemperatureKick(t, world, 10, seed, "temperature kick after zero-temperature collision")
}

func resetWallSpringCollisionWorld(world *Simulation) {
	world.Masses[0].Position = Vec2{X: 0, Y: 0}
	world.Masses[0].Velocity = Vec2{}
	world.Masses[1].Position = Vec2{X: 0, Y: 100}
	world.Masses[1].Velocity = Vec2{}
	world.Masses[2].Position = Vec2{X: -5, Y: 50}
	world.Masses[2].Velocity = Vec2{X: 10}
}

func TestWallSpringRestoresEndpointDistanceToRestLength(t *testing.T) {
	world := wallSpringLengthWorld(120, 100, false, false)

	world.Step(1)

	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if got := length(b.Position.Sub(a.Position)); !closeWallSpringLength(got, 100) {
		t.Fatalf("endpoint distance = %f, expected 100", got)
	}
	if a.Position.Y != 0 || b.Position.Y != 0 {
		t.Fatalf("length correction should stay along segment: a=%#v b=%#v", a.Position, b.Position)
	}
	if a.Position.X >= b.Position.X {
		t.Fatalf("length correction reversed endpoints: a=%#v b=%#v", a.Position, b.Position)
	}
}

func TestWallSpringLengthCorrectionAbsorbsFixedEndpointShare(t *testing.T) {
	world := wallSpringLengthWorld(120, 100, true, false)

	world.Step(1)

	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if a.Position != (Vec2{}) {
		t.Fatalf("fixed endpoint moved to %#v", a.Position)
	}
	if got := length(b.Position.Sub(a.Position)); !closeWallSpringLength(got, 100) {
		t.Fatalf("endpoint distance = %f, expected 100", got)
	}
}

func TestWallSpringRestoresEndpointDistanceToUnitRestLength(t *testing.T) {
	world := wallSpringLengthWorld(120, 1, false, false)

	world.Step(1)

	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if got := length(b.Position.Sub(a.Position)); !closeWallSpringLength(got, 1) {
		t.Fatalf("endpoint distance = %f, expected 1", got)
	}
}

func TestWallSpringZeroRestLengthCapturesCurrentEndpointDistance(t *testing.T) {
	world := wallSpringLengthWorld(120, 0, false, false)

	world.Step(1)

	spring, _ := world.SpringByID(1)
	a, _ := world.MassByID(1)
	b, _ := world.MassByID(2)
	if !closeWallSpringLength(spring.RestLength, 120) {
		t.Fatalf("captured rest length = %f, expected 120", spring.RestLength)
	}
	if got := length(b.Position.Sub(a.Position)); !closeWallSpringLength(got, 120) {
		t.Fatalf("endpoint distance = %f, expected 120", got)
	}
}

func TestWallSpringUnitLengthCapturesZeroRestLength(t *testing.T) {
	world := wallSpringLengthWorld(1, 0, false, false)

	world.Step(1)

	spring, _ := world.SpringByID(1)
	if !closeWallSpringLength(spring.RestLength, 1) {
		t.Fatalf("captured rest length = %f, expected 1", spring.RestLength)
	}
}

func wallSpringCollisionWorld(fixedA bool, fixedB bool, contactY ...float64) *Simulation {
	y := 50.0
	if len(contactY) > 0 {
		y = contactY[0]
	}
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1, Fixed: fixedA})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 100}, Mass: 1, Fixed: fixedB})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -5, Y: y}, Velocity: Vec2{X: 10}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	return world
}

func wallSpringLengthWorld(initialLength, restLength float64, fixedA, fixedB bool) *Simulation {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1, Fixed: fixedA})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: initialLength}, Mass: 1, Fixed: fixedB})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: restLength, Wall: true})
	return world
}

func closeWallSpringLength(got, want float64) bool {
	return math.Abs(got-want) <= 0.00001
}
