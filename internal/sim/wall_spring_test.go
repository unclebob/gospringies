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

func TestWallSpringStopsFastMassPathCrossingSegment(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 100}, Mass: 1})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -50, Y: 50}, Velocity: Vec2{X: 1000}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})

	world.Step(1)

	mass, _ := world.MassByID(3)
	if mass.Position.X > 0 {
		t.Fatalf("fast mass crossed wall spring: %#v", mass)
	}
	if mass.Velocity.X > 0 {
		t.Fatalf("fast mass velocity still penetrates wall spring: %#v", mass.Velocity)
	}
}

func TestMovingWallSpringStopsStationaryMassCrossingSegment(t *testing.T) {
	world := movingWallSpringCollisionWorld()

	world.Step(1)

	mass, _ := world.MassByID(3)
	if mass.Position.X < 0 {
		t.Fatalf("stationary mass crossed moving wall spring: %#v", mass)
	}
	endpointA, _ := world.MassByID(1)
	endpointB, _ := world.MassByID(2)
	if endpointA.Velocity.X > 0.000001 || endpointB.Velocity.X > 0.000001 {
		t.Fatalf("wall spring endpoint velocities still penetrate mass: %#v %#v", endpointA.Velocity, endpointB.Velocity)
	}
}

func TestWallSpringCollisionPlacesMassAtRadius(t *testing.T) {
	world := wallSpringCollisionWorld(false, false)

	world.Step(1)

	mass, _ := world.MassByID(3)
	if !closeWallSpringLength(mass.Position.X, -MassRadius(mass)) {
		t.Fatalf("mass position X = %f, expected wall radius offset %f", mass.Position.X, -MassRadius(mass))
	}
}

func TestUnitLengthWallSpringStillCollides(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{Y: 1}, Mass: 1})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -0.5, Y: 0.5}, Velocity: Vec2{X: 1}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})

	world.Step(1)

	mass, _ := world.MassByID(3)
	if mass.Position.X > 0 || mass.Velocity.X > 0 {
		t.Fatalf("unit wall spring did not collide: %#v", mass)
	}
}

func TestWallSpringContactVelocityInterpolatesEndpointVelocity(t *testing.T) {
	endpointA := &Mass{Velocity: Vec2{X: 10}}
	endpointB := &Mass{Velocity: Vec2{X: 20}}

	if got := wallSpringContactVelocity(endpointA, endpointB, 0.25); got != (Vec2{X: 12.5}) {
		t.Fatalf("contact velocity = %#v, expected 12.5", got)
	}
}

func TestWallSpringContactFractionRejectsNonCrossings(t *testing.T) {
	segment := Vec2{Y: 10}
	for _, tc := range []struct {
		name         string
		previousSide float64
		currentSide  float64
	}{
		{name: "previous on wall", previousSide: 0, currentSide: -1},
		{name: "current on wall", previousSide: 1, currentSide: 0},
		{name: "same positive side", previousSide: 1, currentSide: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := wallSpringContactFraction(Vec2{X: tc.previousSide, Y: 5}, Vec2{X: tc.currentSide, Y: 5}, segment, 100, tc.previousSide, tc.currentSide)
			if ok {
				t.Fatal("contact fraction accepted non-crossing")
			}
		})
	}
}

func TestWallSpringContactFractionAcceptsEndpointContacts(t *testing.T) {
	segment := Vec2{Y: 10}
	for _, tc := range []struct {
		name     string
		previous Vec2
		current  Vec2
		want     float64
	}{
		{name: "start endpoint", previous: Vec2{X: -1}, current: Vec2{X: 1}, want: 0},
		{name: "end endpoint", previous: Vec2{X: -1, Y: 10}, current: Vec2{X: 1, Y: 10}, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := wallSpringContactFraction(tc.previous, tc.current, segment, 100, tc.previous.X, tc.current.X)
			if !ok || !closeWallSpringLength(got, tc.want) {
				t.Fatalf("contact fraction = %f, %t, expected %f", got, ok, tc.want)
			}
		})
	}
}

func TestWallSpringContactFractionRejectsOutsideSegment(t *testing.T) {
	segment := Vec2{Y: 10}
	for _, tc := range []struct {
		name     string
		previous Vec2
		current  Vec2
	}{
		{name: "before start", previous: Vec2{X: -1, Y: -0.1}, current: Vec2{X: 1, Y: -0.1}},
		{name: "after end", previous: Vec2{X: -1, Y: 10.1}, current: Vec2{X: 1, Y: 10.1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := wallSpringContactFraction(tc.previous, tc.current, segment, 100, tc.previous.X, tc.current.X)
			if ok {
				t.Fatal("contact fraction accepted outside segment")
			}
		})
	}
}

func TestResolveWallSpringVelocityLeavesSeparatingRelativeVelocity(t *testing.T) {
	mass := Mass{Velocity: Vec2{X: 0.5}}

	resolveWallSpringVelocity(&mass, Vec2{}, Vec2{X: 1}, 1)

	if mass.Velocity != (Vec2{X: 0.5}) {
		t.Fatalf("separating velocity changed to %#v", mass.Velocity)
	}
}

func TestResolveWallSpringVelocityLeavesTangentRelativeVelocity(t *testing.T) {
	mass := Mass{Velocity: Vec2{Y: 1}}

	resolveWallSpringVelocity(&mass, Vec2{}, Vec2{X: 1}, 1)

	if mass.Velocity != (Vec2{Y: 1}) {
		t.Fatalf("tangent velocity changed to %#v", mass.Velocity)
	}
}

func TestResolveWallSpringVelocityReflectsUnitPenetratingVelocity(t *testing.T) {
	mass := Mass{Velocity: Vec2{X: 1}}

	resolveWallSpringVelocity(&mass, Vec2{}, Vec2{X: 1}, -1)

	if mass.Velocity.X >= 0 {
		t.Fatalf("penetrating unit velocity was not reflected: %#v", mass.Velocity)
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

func TestWallSpringLengthCorrectionDoesNotMoveEndpointThroughOtherWallSpring(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 100}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -5, Y: 40}, Mass: 1})
	_ = world.AddMass(Mass{ID: 4, Position: Vec2{X: -80, Y: 40}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	_ = world.AddSpring(Spring{ID: 2, MassA: 3, MassB: 4, RestLength: 150, Wall: true})

	world.Step(1)

	endpointA, _ := world.MassByID(3)
	if endpointA.Position.X > 0 {
		t.Fatalf("wall spring endpoint crossed barrier after length correction: %#v", endpointA)
	}
}

func TestWallSpringEndpointUsesFullTimestepPathForOtherWallSpringCollision(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 100}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -5, Y: 40}, Velocity: Vec2{X: 10}, Mass: 1})
	_ = world.AddMass(Mass{ID: 4, Position: Vec2{X: -100, Y: 40}, Mass: 1, Fixed: true})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	_ = world.AddSpring(Spring{ID: 2, MassA: 3, MassB: 4, RestLength: 150, Wall: true})

	world.Step(1)

	endpoint, _ := world.MassByID(3)
	if endpoint.Position.X > 0 {
		t.Fatalf("wall spring endpoint crossed barrier over full timestep path: %#v", endpoint)
	}
	if endpoint.Velocity.X > 0 {
		t.Fatalf("wall spring endpoint velocity still penetrates barrier: %#v", endpoint.Velocity)
	}
}

func TestWallSpringCollisionsIgnoreZeroTimeStepAfterLengthCorrection(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 100}, Mass: 1})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: 5, Y: 40}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	beforeLengthConstraints := []Vec2{
		{X: 0, Y: 0},
		{X: 0, Y: 100},
		{X: -5, Y: 40},
	}

	world.applyWallSpringCollisions(0, beforeLengthConstraints)

	mass, _ := world.MassByID(3)
	if mass.Position != (Vec2{X: 5, Y: 40}) || mass.Velocity != (Vec2{}) {
		t.Fatalf("zero timestep changed wall spring collision state: %#v", mass)
	}
}

func TestWallSpringPreviousPositionUsesCorrectBoundarySource(t *testing.T) {
	for _, tc := range []struct {
		name           string
		mass           Mass
		startPositions []Vec2
		index          int
		dt             float64
		want           Vec2
	}{
		{
			name:           "first index uses timestep start position",
			mass:           Mass{Position: Vec2{X: 2}, Velocity: Vec2{X: 10}},
			startPositions: []Vec2{{X: -8}},
			index:          0,
			dt:             1,
			want:           Vec2{X: -8},
		},
		{
			name:           "index at length falls back to velocity",
			mass:           Mass{Position: Vec2{X: 2}, Velocity: Vec2{X: 10}},
			startPositions: []Vec2{{X: 1}},
			index:          1,
			dt:             1,
			want:           Vec2{X: -8},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := wallSpringPreviousPosition(tc.mass, tc.startPositions, tc.index, tc.dt); got != tc.want {
				t.Fatalf("previous position = %#v, expected %#v", got, tc.want)
			}
		})
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

func movingWallSpringCollisionWorld() *Simulation {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: -5, Y: 0}, Velocity: Vec2{X: 10}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: -5, Y: 100}, Velocity: Vec2{X: 10}, Mass: 1})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: 0, Y: 50}, Mass: 1})
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
