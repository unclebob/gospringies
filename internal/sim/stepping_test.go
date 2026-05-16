package sim

import "testing"

func TestStepMovesMassUnderGravity(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})

	world.Step(0.1)

	mass, _ := world.MassByID(1)
	if mass.Position == (Vec2{}) {
		t.Fatal("position did not change")
	}
	if mass.Velocity == (Vec2{}) {
		t.Fatal("velocity did not change")
	}
}

func TestStepKeepsFixedMassStationary(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 5, Y: 6}, Mass: 1, Fixed: true})

	world.AdvanceDuration(1.0)

	mass, _ := world.MassByID(1)
	if mass.Position != (Vec2{X: 5, Y: 6}) {
		t.Fatalf("fixed mass position = %#v", mass.Position)
	}
	if mass.Velocity != (Vec2{}) {
		t.Fatalf("fixed mass velocity = %#v", mass.Velocity)
	}
}

func TestAdvanceDurationIsDeterministic(t *testing.T) {
	first := NewDemoSimulation()
	second := NewDemoSimulation()

	first.AdvanceDuration(1.0)
	second.AdvanceDuration(1.0)

	if first.Masses[1].Position != second.Masses[1].Position {
		t.Fatalf("positions differ: %#v != %#v", first.Masses[1].Position, second.Masses[1].Position)
	}
}

func TestAdvanceDurationTracksRequestedTime(t *testing.T) {
	world := NewWorld()

	world.AdvanceDuration(1.0)

	if abs(world.Time-1.0) > 0.000001 {
		t.Fatalf("time = %f", world.Time)
	}
}
