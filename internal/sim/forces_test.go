package sim

import (
	"math"
	"testing"
)

func TestSpringForceIsEqualAndOpposite(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 12, Y: 0}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 10, SpringConstant: 12})

	forces := world.EvaluateForces()

	assertVecEqual(t, forces.ByMassID[1].Force, forces.ByMassID[2].Force.Scale(-1))
}

func TestSpringDampingActsAlongSpringDirection(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Velocity: Vec2{X: 1, Y: 5}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 10, Y: 0}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 10, Damping: 0.5})

	forces := world.EvaluateForces()

	if forces.ByMassID[1].Force.Y != 0 || forces.ByMassID[2].Force.Y != 0 {
		t.Fatalf("damping affected non-spring direction: %#v", forces.ByMassID)
	}
	if forces.ByMassID[1].Force.X == 0 {
		t.Fatal("damping did not affect spring direction")
	}
}

func TestEnvironmentalForcesCanBeEvaluatedIndependently(t *testing.T) {
	for _, forceName := range []string{"gravity", "viscosity", "wall repulsion", "center attraction", "center of mass attraction"} {
		world := worldWithEnvironmentalForce(forceName)

		forces := world.EvaluateForces()

		if forces.ByMassID[1].Force == (Vec2{}) {
			t.Fatalf("%s produced no force", forceName)
		}
	}
}

func TestFixedMassesDoNotAccumulateAcceleration(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 10, Y: 10}, Mass: 1, Fixed: true})
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})

	forces := world.EvaluateForces()

	if forces.ByMassID[1].Acceleration != (Vec2{}) {
		t.Fatalf("fixed acceleration = %#v", forces.ByMassID[1].Acceleration)
	}
}

func TestWallForcePushesMassBackInside(t *testing.T) {
	cases := []struct {
		wall string
		id   int
		pos  Vec2
		want Vec2
	}{
		{"top", 1, Vec2{X: 50, Y: -5}, Vec2{Y: 1}},
		{"left", 2, Vec2{X: -5, Y: 50}, Vec2{X: 1}},
		{"right", 3, Vec2{X: 105, Y: 50}, Vec2{X: -1}},
		{"bottom", 4, Vec2{X: 50, Y: 105}, Vec2{Y: -1}},
	}
	for _, tc := range cases {
		world := NewWorld()
		world.Bounds = Bounds{Width: 100, Height: 100}
		world.Parameters.EnableWall(tc.wall)
		world.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "10", "exponent": "1"})
		_ = world.AddMass(Mass{ID: tc.id, Position: tc.pos, Mass: 1})

		forces := world.EvaluateForces()

		got := forces.ByMassID[tc.id].Force
		if dot(got, tc.want) <= 0 {
			t.Fatalf("%s force = %#v, expected toward %#v", tc.wall, got, tc.want)
		}
	}
}

func TestGravityDirectionUsesXSpringiesDegrees(t *testing.T) {
	cases := []struct {
		degrees string
		want    Vec2
	}{
		{"0", Vec2{Y: 1}},
		{"90", Vec2{X: 1}},
		{"180", Vec2{Y: -1}},
		{"270", Vec2{X: -1}},
	}
	for _, tc := range cases {
		world := NewWorld()
		world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1", "direction": tc.degrees})
		_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1})

		force := world.EvaluateForces().ByMassID[1].Force

		assertVecEqual(t, roundedVec(force), tc.want)
	}
}

func TestCenterMassDoesNotReceiveCenterForceResponse(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableForce("center attraction", map[string]string{"magnitude": "10", "exponent": "0"})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 50, Y: 50}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 0, Y: 50}, Mass: 1})
	world.SetForceCenter([]int{1})

	forces := world.EvaluateForces()

	if forces.ByMassID[1].Force != (Vec2{}) {
		t.Fatalf("center mass force = %#v", forces.ByMassID[1].Force)
	}
	if forces.ByMassID[2].Force == (Vec2{}) {
		t.Fatal("non-center mass did not receive center force")
	}
}

func TestForceParametersAndCenterSelection(t *testing.T) {
	if got := ForceParameterNames("gravity"); len(got) != 2 || got[0] != "Magnitude" || got[1] != "Direction" {
		t.Fatalf("gravity parameters = %#v", got)
	}
	world := NewWorld()
	world.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "1"})
	if world.Parameters.ActiveForce != "wall repulsion" {
		t.Fatalf("active force = %q", world.Parameters.ActiveForce)
	}
	world.SetForceCenter([]int{7})
	if world.CenterMassID() != 7 {
		t.Fatalf("center mass = %d", world.CenterMassID())
	}
	world.SetForceCenter(nil)
	if world.CenterMassID() != -1 {
		t.Fatalf("screen center id = %d", world.CenterMassID())
	}
}

func worldWithEnvironmentalForce(forceName string) *Simulation {
	world := NewWorld()
	world.Bounds = Bounds{Width: 100, Height: 100}
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: -1, Y: 50}, Velocity: Vec2{X: 2, Y: 0}, Mass: 1})
	if forceName == "center of mass attraction" {
		_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 50, Y: 50}, Mass: 1})
	}
	if forceName == "viscosity" {
		world.Parameters.Set("viscosity", "1")
		return world
	}
	if forceName == "wall repulsion" {
		world.Parameters.EnableWall("left")
	}
	world.Parameters.EnableForce(forceName, map[string]string{"magnitude": "10", "direction": "90", "exponent": "1", "damping": "1"})
	return world
}

func assertVecEqual(t *testing.T, got, want Vec2) {
	t.Helper()
	if math.Abs(got.X-want.X) > 0.000001 || math.Abs(got.Y-want.Y) > 0.000001 {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func roundedVec(v Vec2) Vec2 {
	return Vec2{X: math.Round(v.X), Y: math.Round(v.Y)}
}
