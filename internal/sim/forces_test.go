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

func TestSpringForcesUseEndpointIDsWhenIndexesAreStale(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 10, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(Mass{ID: 20, Position: Vec2{X: 12, Y: 0}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, A: 0, B: 0, MassA: 10, MassB: 20, RestLength: 10, SpringConstant: 12})

	forces := world.EvaluateForces()

	if forces.ByMassID[10].Force.X <= 0 {
		t.Fatalf("mass 10 force = %#v, want pull toward mass 20", forces.ByMassID[10].Force)
	}
	if forces.ByMassID[20].Force.X >= 0 {
		t.Fatalf("mass 20 force = %#v, want pull toward mass 10", forces.ByMassID[20].Force)
	}
}

func TestSpringEndpointResolutionCoversIDAndIndexContracts(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 10, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(Mass{ID: 20, Position: Vec2{X: 12, Y: 0}, Mass: 1})

	a, b, ok := world.springEndpointMasses(Spring{A: 0, B: 1})
	if !ok || a.ID != 10 || b.ID != 20 {
		t.Fatalf("index endpoints = %#v, %#v, %v", a, b, ok)
	}
	if _, _, ok := world.springEndpointMasses(Spring{A: -1, B: 1}); ok {
		t.Fatal("negative index should not resolve")
	}
	if _, _, ok := world.springEndpointMasses(Spring{A: 0, B: len(world.Masses)}); ok {
		t.Fatal("past-end index should not resolve")
	}
	if _, _, ok := world.springEndpointMasses(Spring{A: 0, B: 1, MassA: 10}); ok {
		t.Fatal("partial ID endpoints should not fall back to indexes")
	}
	if _, _, ok := world.springEndpointMasses(Spring{A: 0, B: 1, MassA: 10, MassB: 99}); ok {
		t.Fatal("missing ID endpoint should not resolve")
	}
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

func TestSpringForceHandlesCoincidentAndCompressedMasses(t *testing.T) {
	coincident := NewWorld()
	_ = coincident.AddMass(Mass{ID: 1, Position: Vec2{X: 5, Y: 5}, Mass: 1})
	_ = coincident.AddMass(Mass{ID: 2, Position: Vec2{X: 5, Y: 5}, Mass: 1})
	_ = coincident.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 10, SpringConstant: 12})
	forces := coincident.EvaluateForces()
	if forces.ByMassID[1].Force != (Vec2{}) || forces.ByMassID[2].Force != (Vec2{}) {
		t.Fatalf("coincident spring forces = %#v", forces.ByMassID)
	}

	compressed := NewWorld()
	_ = compressed.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = compressed.AddMass(Mass{ID: 2, Position: Vec2{X: 5, Y: 0}, Mass: 1})
	_ = compressed.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 10, SpringConstant: 2})
	forces = compressed.EvaluateForces()
	assertVecEqual(t, forces.ByMassID[1].Force, Vec2{X: -10})
	assertVecEqual(t, forces.ByMassID[2].Force, Vec2{X: 10})
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

func TestWallForcePushesMassAwayFromInsideWalls(t *testing.T) {
	cases := []struct {
		wall string
		id   int
		pos  Vec2
		want Vec2
	}{
		{"bottom", 1, Vec2{X: 50, Y: 5}, Vec2{Y: 1}},
		{"left", 2, Vec2{X: 5, Y: 50}, Vec2{X: 1}},
		{"right", 3, Vec2{X: 95, Y: 50}, Vec2{X: -1}},
		{"top", 4, Vec2{X: 50, Y: 95}, Vec2{Y: -1}},
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
		{"0", Vec2{Y: -1}},
		{"90", Vec2{X: 1}},
		{"180", Vec2{Y: 1}},
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

func TestGravityScalesByMagnitudeAndMass(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "2", "direction": "45"})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 3})

	force := world.EvaluateForces().ByMassID[1].Force

	assertVecEqual(t, roundedVec(force), Vec2{X: 4, Y: -4})
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

func TestCenterForceBoundaryAndDamping(t *testing.T) {
	disabled := NewWorld()
	_ = disabled.AddMass(Mass{ID: 1, Position: Vec2{X: 10, Y: 0}, Mass: 1})
	if force := disabled.EvaluateForces().ByMassID[1].Force; force != (Vec2{}) {
		t.Fatalf("disabled center force = %#v", force)
	}

	atCenter := NewWorld()
	atCenter.Bounds = Bounds{Width: 100, Height: 100}
	atCenter.Parameters.EnableForce("center attraction", map[string]string{"magnitude": "10", "exponent": "1"})
	_ = atCenter.AddMass(Mass{ID: 1, Position: Vec2{X: 50, Y: 50}, Mass: 1})
	if force := atCenter.EvaluateForces().ByMassID[1].Force; force != (Vec2{}) {
		t.Fatalf("mass at center force = %#v", force)
	}

	damped := NewWorld()
	damped.Parameters.EnableForce("center of mass attraction", map[string]string{"magnitude": "10", "damping": "2"})
	_ = damped.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Velocity: Vec2{X: 2}, Mass: 1})
	_ = damped.AddMass(Mass{ID: 2, Position: Vec2{X: 10, Y: 0}, Mass: 1})
	force := damped.EvaluateForces().ByMassID[1].Force
	assertVecEqual(t, force, Vec2{X: -2})

	centerMass := NewWorld()
	centerMass.Parameters.EnableForce("center of mass attraction", map[string]string{"magnitude": "10", "damping": "0"})
	_ = centerMass.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = centerMass.AddMass(Mass{ID: 2, Position: Vec2{X: 10, Y: 0}, Mass: 1})
	centerMass.SetForceCenter([]int{1})
	if force := centerMass.EvaluateForces().ByMassID[1].Force; force != (Vec2{}) {
		t.Fatalf("center mass center-of-mass force = %#v", force)
	}
}

func TestForceCenterAndCenterOfMassFallbacks(t *testing.T) {
	world := NewWorld()
	world.Bounds = Bounds{Width: 80, Height: 60}
	world.Parameters.Set("center mass", "0")
	if got := world.forceCenter(); got != (Vec2{X: 40, Y: 30}) {
		t.Fatalf("zero center id force center = %#v", got)
	}
	world.Masses = []Mass{{ID: 0, Position: Vec2{X: 1, Y: 2}}}
	if got := world.forceCenter(); got != (Vec2{X: 40, Y: 30}) {
		t.Fatalf("zero center id with invalid mass = %#v", got)
	}
	world.Parameters.Set("center mass", "1")
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 10, Y: 20}, Mass: 1})
	if got := world.forceCenter(); got != (Vec2{X: 10, Y: 20}) {
		t.Fatalf("selected force center = %#v", got)
	}
	world.Masses = nil
	if got := world.centerOfMass(); got != (Vec2{X: 40, Y: 30}) {
		t.Fatalf("empty center of mass = %#v", got)
	}

	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 10, Y: 20}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 30, Y: 40}, Mass: 1})
	if got := world.centerOfMass(); got != (Vec2{X: 20, Y: 30}) {
		t.Fatalf("center of mass = %#v", got)
	}
}

func TestWallForceBoundariesAndHelpers(t *testing.T) {
	outsideCases := []struct {
		wall string
		pos  Vec2
	}{
		{"bottom", Vec2{X: 10, Y: -1}},
		{"left", Vec2{X: -1, Y: 10}},
		{"right", Vec2{X: 101, Y: 10}},
		{"top", Vec2{X: 10, Y: 101}},
	}
	for _, tc := range outsideCases {
		world := NewWorld()
		world.Bounds = Bounds{Width: 100, Height: 100}
		world.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
		world.Parameters.EnableWall(tc.wall)
		_ = world.AddMass(Mass{ID: 1, Position: tc.pos, Mass: 1})
		if force := world.EvaluateForces().ByMassID[1].Force; force != (Vec2{}) {
			t.Fatalf("outside %s force = %#v", tc.wall, force)
		}
	}
	if got := wallMagnitude(8, 0, 2); got != 8 {
		t.Fatalf("zero-distance wall magnitude = %v", got)
	}
	if got := wallMagnitude(8, 1, 2); got != 8 {
		t.Fatalf("unit-distance wall magnitude = %v", got)
	}
	if got := wallMagnitude(8, 0.5, 1); got != 8 {
		t.Fatalf("fractional-distance wall magnitude = %v", got)
	}
	if got := wallMagnitude(8, 2, 1); got != 4 {
		t.Fatalf("positive-distance wall magnitude = %v", got)
	}
	if got := dot(Vec2{X: 2, Y: 3}, Vec2{X: 5, Y: 7}); got != 31 {
		t.Fatalf("dot = %v", got)
	}

	disabled := NewWorld()
	disabled.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	_ = disabled.AddMass(Mass{ID: 1, Position: Vec2{X: -1, Y: 50}, Mass: 1})
	if force := disabled.EvaluateForces().ByMassID[1].Force; force != (Vec2{}) {
		t.Fatalf("disabled wall force = %#v", force)
	}

	nearRight := NewWorld()
	nearRight.Bounds = Bounds{Width: 100, Height: 100}
	nearRight.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	nearRight.Parameters.EnableWall("right")
	_ = nearRight.AddMass(Mass{ID: 1, Position: Vec2{X: 98, Y: 50}, Mass: 1})
	assertVecEqual(t, nearRight.EvaluateForces().ByMassID[1].Force, Vec2{X: -4})

	nearBottom := NewWorld()
	nearBottom.Bounds = Bounds{Width: 100, Height: 100}
	nearBottom.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	nearBottom.Parameters.EnableWall("bottom")
	_ = nearBottom.AddMass(Mass{ID: 2, Position: Vec2{X: 50, Y: 2}, Mass: 1})
	assertVecEqual(t, nearBottom.EvaluateForces().ByMassID[2].Force, Vec2{Y: 4})

	nearLeft := NewWorld()
	nearLeft.Bounds = Bounds{Width: 100, Height: 100}
	nearLeft.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	nearLeft.Parameters.EnableWall("left")
	_ = nearLeft.AddMass(Mass{ID: 3, Position: Vec2{X: 2, Y: 50}, Mass: 1})
	assertVecEqual(t, nearLeft.EvaluateForces().ByMassID[3].Force, Vec2{X: 4})

	nearTop := NewWorld()
	nearTop.Bounds = Bounds{Width: 100, Height: 100}
	nearTop.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	nearTop.Parameters.EnableWall("top")
	_ = nearTop.AddMass(Mass{ID: 4, Position: Vec2{X: 50, Y: 98}, Mass: 1})
	assertVecEqual(t, nearTop.EvaluateForces().ByMassID[4].Force, Vec2{Y: -4})

	outsideRight := NewWorld()
	outsideRight.Bounds = Bounds{Width: 100, Height: 100}
	outsideRight.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	outsideRight.Parameters.EnableWall("right")
	_ = outsideRight.AddMass(Mass{ID: 1, Position: Vec2{X: 101, Y: 50}, Mass: 1})
	assertVecEqual(t, outsideRight.EvaluateForces().ByMassID[1].Force, Vec2{})

	outsideBottom := NewWorld()
	outsideBottom.Bounds = Bounds{Width: 100, Height: 100}
	outsideBottom.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "8", "exponent": "1"})
	outsideBottom.Parameters.EnableWall("bottom")
	_ = outsideBottom.AddMass(Mass{ID: 2, Position: Vec2{X: 50, Y: -1}, Mass: 1})
	assertVecEqual(t, outsideBottom.EvaluateForces().ByMassID[2].Force, Vec2{})
}

func TestWallChecksIncludeExactSimulationBoundaries(t *testing.T) {
	world := NewWorld()
	world.Bounds = Bounds{Width: 100, Height: 80}

	for _, check := range world.wallChecks(Mass{Position: Vec2{X: 0, Y: 0}}, 8) {
		if (check.name == "bottom" || check.name == "left") && !check.inside {
			t.Fatalf("%s boundary should be inside", check.name)
		}
	}
	for _, check := range world.wallChecks(Mass{Position: Vec2{X: 100, Y: 80}}, 8) {
		if (check.name == "right" || check.name == "top") && !check.inside {
			t.Fatalf("%s boundary should be inside", check.name)
		}
	}
}

func TestCenterMassAndEnabledForceHelpers(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("center mass", "not-an-id")
	if world.CenterMassID() != -1 {
		t.Fatalf("invalid center id = %d", world.CenterMassID())
	}
	if world.IsCenterMass(0) {
		t.Fatal("zero id reported as center mass")
	}
	world.Parameters.Set("center mass", "0")
	if world.IsCenterMass(0) {
		t.Fatal("zero center id reported as center mass")
	}
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1"})
	world.Parameters.Forces["gravity"] = ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "1"}}
	if _, ok := world.enabledForce("gravity"); ok {
		t.Fatal("disabled configured force reported enabled")
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
	position := Vec2{X: -1, Y: 50}
	if forceName == "wall repulsion" {
		position = Vec2{X: 1, Y: 50}
	}
	_ = world.AddMass(Mass{ID: 1, Position: position, Velocity: Vec2{X: 2, Y: 0}, Mass: 1})
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
