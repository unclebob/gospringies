package sim

import "testing"

func TestMassCollisionSwapsEqualMassVelocities(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableForce("mass collision", map[string]string{})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 2, Y: 0}, Velocity: Vec2{X: -1}, Mass: 1, Elasticity: 1})

	world.applyMassCollisions()

	assertVecEqual(t, world.Masses[0].Velocity, Vec2{X: -1})
	assertVecEqual(t, world.Masses[1].Velocity, Vec2{X: 1})
}

func TestMassCollisionBouncesFreeMassOffFixedMass(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableForce("mass collision", map[string]string{})
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 2, Y: 0}, Mass: 1, Elasticity: 1, Fixed: true})

	world.applyMassCollisions()

	assertVecEqual(t, world.Masses[0].Velocity, Vec2{X: -1})
	assertVecEqual(t, world.Masses[1].Velocity, Vec2{})
}

func TestMassCollisionIgnoresDisabledAndSeparatedMasses(t *testing.T) {
	disabled := NewWorld()
	_ = disabled.AddMass(Mass{ID: 1, Position: Vec2{}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = disabled.AddMass(Mass{ID: 2, Position: Vec2{X: 2}, Velocity: Vec2{X: -1}, Mass: 1, Elasticity: 1})
	disabled.applyMassCollisions()
	assertVecEqual(t, disabled.Masses[0].Velocity, Vec2{X: 1})
	assertVecEqual(t, disabled.Masses[1].Velocity, Vec2{X: -1})

	separated := NewWorld()
	separated.Parameters.EnableForce("mass collision", map[string]string{})
	_ = separated.AddMass(Mass{ID: 1, Position: Vec2{}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = separated.AddMass(Mass{ID: 2, Position: Vec2{X: 100}, Velocity: Vec2{X: -1}, Mass: 1, Elasticity: 1})
	separated.applyMassCollisions()
	assertVecEqual(t, separated.Masses[0].Velocity, Vec2{X: 1})
	assertVecEqual(t, separated.Masses[1].Velocity, Vec2{X: -1})
}

func TestMassRadiusMatchesXSpringiesMassRadius(t *testing.T) {
	if got := MassRadius(Mass{Mass: 1}); got != 3 {
		t.Fatalf("mass radius = %v", got)
	}
	if got := MassRadius(Mass{Mass: 1, Fixed: true}); got != fixedMassCollisionRadius {
		t.Fatalf("fixed radius = %v", got)
	}
	if got := MassRadius(Mass{Mass: 1e20}); got != 64 {
		t.Fatalf("large radius = %v", got)
	}
}
