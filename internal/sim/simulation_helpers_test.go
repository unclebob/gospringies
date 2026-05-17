package sim

import (
	"math"
	"testing"
)

func TestVec2OperationsUseBothComponents(t *testing.T) {
	first := Vec2{X: 7, Y: 11}
	second := Vec2{X: 2, Y: 3}

	if got := first.Add(second); got != (Vec2{X: 9, Y: 14}) {
		t.Fatalf("add = %#v", got)
	}
	if got := first.Sub(second); got != (Vec2{X: 5, Y: 8}) {
		t.Fatalf("sub = %#v", got)
	}
	if got := (Vec2{}).Normalize(); got != (Vec2{}) {
		t.Fatalf("zero normalize = %#v", got)
	}
}

func TestResetClearsObjectsParametersAndTime(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Mass: 1})
	world.Parameters.Set("current mass", "custom")
	world.Time = 2.5

	world.Reset()

	if len(world.Masses) != 0 || len(world.Springs) != 0 {
		t.Fatalf("objects after reset = %#v %#v", world.Masses, world.Springs)
	}
	if world.Time != 0 {
		t.Fatalf("time after reset = %f", world.Time)
	}
	if world.Parameters.Value("current mass") != DefaultParameters().Value("current mass") {
		t.Fatalf("parameters after reset = %#v", world.Parameters)
	}
}

func TestDemoSimulationUsesFixedAndMovableEndpoints(t *testing.T) {
	world := NewDemoSimulation()

	if len(world.Masses) != 2 || len(world.Springs) != 1 {
		t.Fatalf("demo world = %#v %#v", world.Masses, world.Springs)
	}
	if !world.Masses[0].Fixed || world.Masses[1].Fixed {
		t.Fatalf("demo fixed flags = %#v", world.Masses)
	}
	if world.Masses[0].Mass != 1 || world.Masses[1].Mass != 1 {
		t.Fatalf("demo masses = %#v", world.Masses)
	}
}

func TestAddMassAtReturnsIndexAndSequentialID(t *testing.T) {
	world := NewWorld()

	first := world.AddMassAt(Vec2{X: 10, Y: 20}, 2, false)
	second := world.AddMassAt(Vec2{X: 30, Y: 40}, 3, true)

	if first != 0 || second != 1 {
		t.Fatalf("indices = %d, %d", first, second)
	}
	if world.Masses[0].ID != 1 || world.Masses[1].ID != 2 {
		t.Fatalf("ids = %#v", world.Masses)
	}
}

func TestAddSpringNormalizesEndpointIndexesAndConstants(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 10, Mass: 1})
	_ = world.AddMass(Mass{ID: 20, Mass: 1})

	if err := world.AddSpring(Spring{ID: 1, MassA: 10, MassB: 20, SpringConstant: 12}); err != nil {
		t.Fatal(err)
	}
	if err := world.AddSpring(Spring{ID: 2, MassA: 20, MassB: 10, Stiffness: 7}); err != nil {
		t.Fatal(err)
	}

	first, _ := world.SpringByID(1)
	second, _ := world.SpringByID(2)
	if first.A != 0 || first.B != 1 || first.Stiffness != 12 {
		t.Fatalf("first spring = %#v", first)
	}
	if second.A != 1 || second.B != 0 || second.SpringConstant != 7 {
		t.Fatalf("second spring = %#v", second)
	}
}

func TestAddSpringBetweenAssignsSpringIDAndEndpointIDs(t *testing.T) {
	world := NewWorld()
	a := world.AddMassAt(Vec2{X: 0}, 1, false)
	b := world.AddMassAt(Vec2{X: 10}, 1, false)

	world.AddSpringBetween(a, b, 10, 5)
	world.AddSpringBetween(b, a, 10, 6)

	if world.Springs[0].ID != 1 || world.Springs[1].ID != 2 {
		t.Fatalf("spring IDs = %#v", world.Springs)
	}
	if world.Springs[0].MassA != 1 || world.Springs[0].MassB != 2 {
		t.Fatalf("endpoint IDs = %#v", world.Springs[0])
	}
}

func TestMassIndexByIDReportsMissingAsZeroFalse(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 4, Mass: 1})

	index, ok := world.massIndexByID(9)

	if ok || index != 0 {
		t.Fatalf("missing index = %d, %t", index, ok)
	}
}

func TestAdvanceBoundaryCounts(t *testing.T) {
	world := NewWorld()
	world.Advance(0, 0.25)
	if world.Time != 0 {
		t.Fatalf("zero advance time = %f", world.Time)
	}

	world.Advance(1, 0.25)
	if world.Time != 0.25 {
		t.Fatalf("one advance time = %f", world.Time)
	}
}

func TestAdvanceDurationBoundarySteps(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("timestep", "0.25")

	world.AdvanceDuration(0)
	if world.Time != 0 {
		t.Fatalf("zero duration time = %f", world.Time)
	}

	world.AdvanceDuration(0.10)
	if math.Abs(world.Time-0.10) > 0.000001 {
		t.Fatalf("partial duration time = %f", world.Time)
	}

	world.AdvanceDuration(advanceEpsilon)
	if world.LastAdvanceSteps != 0 {
		t.Fatalf("epsilon duration steps = %d", world.LastAdvanceSteps)
	}
}

func TestAdvanceDurationUsesConfiguredPositiveTimestep(t *testing.T) {
	direct := gravityWorld()
	advanced := gravityWorld()
	advanced.Parameters.Set("timestep", "0.25")

	direct.Step(0.10)
	advanced.AdvanceDuration(0.10)

	if advanced.Masses[0].Position != direct.Masses[0].Position {
		t.Fatalf("partial configured step position = %#v, want %#v", advanced.Masses[0].Position, direct.Masses[0].Position)
	}
}

func TestAdvanceDurationUsesOneStepWhenDurationEqualsTimestep(t *testing.T) {
	oneStep := gravityWorld()
	advanced := gravityWorld()
	advanced.Parameters.Set("timestep", "0.1")

	oneStep.Step(0.1)
	advanced.AdvanceDuration(0.1)

	if advanced.Masses[0].Position != oneStep.Masses[0].Position {
		t.Fatalf("equal-duration step position = %#v, want %#v", advanced.Masses[0].Position, oneStep.Masses[0].Position)
	}
	if advanced.LastAdvanceSteps != 1 {
		t.Fatalf("equal-duration steps = %d", advanced.LastAdvanceSteps)
	}
}

func TestAdaptiveStepBoundaryHelpers(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("precision", "0")
	if got := world.configuredPrecision(); got != defaultPrecision {
		t.Fatalf("zero configured precision = %f", got)
	}
	world.Parameters.Set("precision", "0.004")
	if got := adaptiveStepDuration(0.25, world.configuredPrecision()); got != 0.25 {
		t.Fatalf("clamped adaptive step = %f", got)
	}
	if got := adaptiveStepDuration(0.25, 0); got != 0.25 {
		t.Fatalf("zero adaptive step = %f", got)
	}
	if got := adaptiveStepDuration(0.25, defaultPrecision); got != 0.25 {
		t.Fatalf("equal adaptive step = %f", got)
	}
}

func TestAdvanceDurationFallsBackWhenTimestepIsZero(t *testing.T) {
	world := gravityWorld()
	world.Parameters.Set("timestep", "0")

	world.AdvanceDuration(0.10)

	if math.Abs(world.Time-0.10) > 0.000001 {
		t.Fatalf("zero-timestep fallback time = %f", world.Time)
	}
	if world.Masses[0].Position == (Vec2{}) {
		t.Fatal("zero-timestep fallback did not advance the world")
	}
}

func gravityWorld() *Simulation {
	world := NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1})
	return world
}

func TestLengthAndSqrtHelpers(t *testing.T) {
	if got := length(Vec2{X: 3, Y: 4}); math.Abs(got-5) > 0.000001 {
		t.Fatalf("length = %f", got)
	}
	if got := length(Vec2{X: 4, Y: 3}); math.Abs(got-5) > 0.000001 {
		t.Fatalf("length swapped = %f", got)
	}
	if got := sqrt(0); got != 0 {
		t.Fatalf("sqrt zero = %f", got)
	}
	if got := sqrt(2); math.Abs(got-math.Sqrt2) > 0.000001 {
		t.Fatalf("sqrt two = %f", got)
	}
	if got := sqrt(2); got != 1.4142135623730949 {
		t.Fatalf("sqrt two exact iteration result = %.17g", got)
	}
}
