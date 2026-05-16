package sim

import "testing"

func TestEnabledWallsBounceWithElasticity(t *testing.T) {
	world := NewWorld()
	world.Damping = 1
	world.Parameters.EnableWall("left")
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 1, Y: 20}, Velocity: Vec2{X: -4}, Mass: 1, Elasticity: 0.5})

	world.Step(1)

	mass, _ := world.MassByID(1)
	if mass.Position.X != 0 || mass.Velocity.X != 2 {
		t.Fatalf("mass = %#v", mass)
	}
}

func TestOneWayWallsAllowOutsideMassesToEnter(t *testing.T) {
	world := NewWorld()
	world.Damping = 1
	world.Parameters.EnableWall("left")
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: -2, Y: 20}, Velocity: Vec2{X: 4}, Mass: 1, Elasticity: 1})

	world.Step(1)

	mass, _ := world.MassByID(1)
	if mass.Velocity.X != 4 || mass.Position.X <= 0 {
		t.Fatalf("mass = %#v", mass)
	}
}

func TestDisabledWallsDoNotBounce(t *testing.T) {
	world := NewWorld()
	world.Damping = 1
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 1, Y: 20}, Velocity: Vec2{X: -4}, Mass: 1, Elasticity: 1})

	world.Step(1)

	mass, _ := world.MassByID(1)
	if mass.Velocity.X != -4 || mass.Position.X >= 0 {
		t.Fatalf("mass = %#v", mass)
	}
}

func TestStickinessCanHoldAndReleaseMass(t *testing.T) {
	world := NewWorld()
	world.Damping = 1
	world.Parameters.EnableWall("left")
	world.Parameters.Set("stickiness", "10")
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 1, Y: 20}, Velocity: Vec2{X: -4}, Mass: 1, Elasticity: 1})

	world.Step(1)
	mass, _ := world.MassByID(1)
	if mass.StuckWall != "left" || mass.Velocity.X != 0 {
		t.Fatalf("stuck mass = %#v", mass)
	}

	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "20", "direction": "0"})
	world.Masses[0].StuckWall = "top"
	world.Masses[0].Position = Vec2{X: 20, Y: 0}
	world.Step(1)
	mass, _ = world.MassByID(1)
	if mass.StuckWall != "" {
		t.Fatalf("released mass = %#v", mass)
	}
}

func TestFixedMassesIgnoreWallCollision(t *testing.T) {
	world := NewWorld()
	world.Damping = 1
	world.Parameters.EnableWall("left")
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 1, Y: 20}, Velocity: Vec2{X: -4}, Mass: 1, Elasticity: 1, Fixed: true})

	world.Step(1)

	mass, _ := world.MassByID(1)
	if mass.Position.X != 1 || mass.Velocity.X != -4 {
		t.Fatalf("fixed mass = %#v", mass)
	}
}
