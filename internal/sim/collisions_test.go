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

	configuredOff := NewWorld()
	configuredOff.Parameters.EnableForce("mass collision", map[string]string{})
	configuredOff.Parameters.Forces["mass collision"] = ForceConfig{Enabled: "false"}
	_ = configuredOff.AddMass(Mass{ID: 1, Position: Vec2{}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = configuredOff.AddMass(Mass{ID: 2, Position: Vec2{X: 2}, Velocity: Vec2{X: -1}, Mass: 1, Elasticity: 1})
	configuredOff.applyMassCollisions()
	assertVecEqual(t, configuredOff.Masses[0].Velocity, Vec2{X: 1})
	assertVecEqual(t, configuredOff.Masses[1].Velocity, Vec2{X: -1})

	missingForce := NewWorld()
	delete(missingForce.Parameters.Forces, "mass collision")
	_ = missingForce.AddMass(Mass{ID: 1, Position: Vec2{}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = missingForce.AddMass(Mass{ID: 2, Position: Vec2{X: 2}, Velocity: Vec2{X: -1}, Mass: 1, Elasticity: 1})
	missingForce.applyMassCollisions()
	assertVecEqual(t, missingForce.Masses[0].Velocity, Vec2{X: 1})
	assertVecEqual(t, missingForce.Masses[1].Velocity, Vec2{X: -1})

	separated := NewWorld()
	separated.Parameters.EnableForce("mass collision", map[string]string{})
	_ = separated.AddMass(Mass{ID: 1, Position: Vec2{}, Velocity: Vec2{X: 1}, Mass: 1, Elasticity: 1})
	_ = separated.AddMass(Mass{ID: 2, Position: Vec2{X: 100}, Velocity: Vec2{X: -1}, Mass: 1, Elasticity: 1})
	separated.applyMassCollisions()
	assertVecEqual(t, separated.Masses[0].Velocity, Vec2{X: 1})
	assertVecEqual(t, separated.Masses[1].Velocity, Vec2{X: -1})
}

func TestCollisionHelperContracts(t *testing.T) {
	if firstCollisionPartnerIndex(3) != 4 {
		t.Fatal("first partner should skip self")
	}

	geometry, ok := collisionGeometryFor(Mass{Position: Vec2{X: 1, Y: 2}, Mass: 4}, Mass{Position: Vec2{X: 3, Y: 5}, Mass: 4})
	if !ok {
		t.Fatal("expected overlapping collision geometry")
	}
	if geometry.dx != 2 || geometry.dy != 3 || geometry.dxq != 4 || geometry.dyq != 9 || geometry.sumxyq != 13 {
		t.Fatalf("geometry = %#v", geometry)
	}
	if _, ok := collisionGeometryFor(Mass{Position: Vec2{X: 1, Y: 1}}, Mass{Position: Vec2{X: 1, Y: 1}}); ok {
		t.Fatal("coincident masses should not produce geometry")
	}
	if _, ok := collisionGeometryFor(Mass{Position: Vec2{}, Mass: 1}, Mass{Position: Vec2{X: 6}, Mass: 1}); ok {
		t.Fatal("tangent masses should not collide")
	}
	if _, ok := collisionGeometryFor(Mass{Position: Vec2{}, Mass: 1}, Mass{Position: Vec2{X: 7}, Mass: 1}); ok {
		t.Fatal("separated masses should not collide")
	}

	if collisionVelocitiesSeparating(Vec2{X: 3, Y: 4}, Vec2{X: 1, Y: 1}, geometry) {
		t.Fatal("approaching velocities should not be separating")
	}
	if !collisionVelocitiesSeparating(Vec2{X: 1, Y: 1}, Vec2{X: 3, Y: 4}, geometry) {
		t.Fatal("receding velocities should be separating")
	}
	if !collisionVelocitiesSeparating(Vec2{X: 1, Y: 4}, Vec2{X: 1, Y: 4}, geometry) {
		t.Fatal("equal velocities should be separating")
	}
	if collisionVelocitiesSeparating(Vec2{X: 1.25, Y: 1}, Vec2{X: 1, Y: 3}, geometry) {
		t.Fatal("positive x approach should not be separating")
	}
	if collisionVelocitiesSeparating(Vec2{X: 3, Y: 2.75}, Vec2{X: 1, Y: 3}, collisionGeometry{dx: -2, dy: -3}) {
		t.Fatal("negative y approach should not be separating")
	}
	if !collisionVelocitiesSeparating(Vec2{X: 2, Y: 9}, Vec2{X: 1, Y: 3}, collisionGeometry{dx: 0, dy: -3}) {
		t.Fatal("zero displacement axis should be separating")
	}
	if axisVelocitiesSeparating(0.25, 0.5) {
		t.Fatal("small positive displacement should detect approach")
	}
	if !axisVelocitiesSeparating(-1, 0) {
		t.Fatal("zero displacement should ignore negative relative velocity")
	}
	if !axisVelocitiesSeparating(0, -2) {
		t.Fatal("zero relative velocity should separate on negative displacement")
	}
	if !axisVelocitiesSeparating(0.5, -2) {
		t.Fatal("positive relative velocity should separate on negative displacement")
	}

	vertical := collisionGeometry{dx: 0, dy: 2, dxq: 0, dyq: 4, sumxyq: 4}
	vertical.avoidVerticalDivision()
	if vertical.dx != 1e-10 {
		t.Fatalf("vertical dx = %v", vertical.dx)
	}
	nonVertical := collisionGeometry{dx: 2}
	nonVertical.avoidVerticalDivision()
	if nonVertical.dx != 2 {
		t.Fatalf("non-vertical dx = %v", nonVertical.dx)
	}
}

func TestCollisionVelocityUsesObliqueGeometryAndMassRatio(t *testing.T) {
	geometry := collisionGeometry{dx: 2, dy: 3, dxq: 4, dyq: 9, sumxyq: 13}
	moving := Mass{ID: 1, Velocity: Vec2{X: 3, Y: 4}, Mass: 2, Elasticity: 0.5}
	other := Mass{ID: 2, Velocity: Vec2{X: -1, Y: 2}, Mass: 6, Elasticity: 0.25}

	if ratio := collisionRatio(moving, other); ratio != 1.03125 {
		t.Fatalf("collision ratio = %v", ratio)
	}
	applyCollisionVelocity(&moving, other, moving.Velocity, other.Velocity, geometry)
	assertVecEqual(t, moving.Velocity, Vec2{X: 0.7788461538461537, Y: 0.6682692307692304})

	fixed := Mass{ID: 3, Velocity: Vec2{X: 5, Y: 6}, Fixed: true}
	applyCollisionVelocity(&fixed, other, fixed.Velocity, other.Velocity, geometry)
	assertVecEqual(t, fixed.Velocity, Vec2{X: 5, Y: 6})
}

func TestMassRadiusMatchesXSpringiesMassRadius(t *testing.T) {
	if got := effectiveCollisionMass(Mass{}); got != 1 {
		t.Fatalf("default collision mass = %v", got)
	}
	if got := MassRadius(Mass{Mass: 1}); got != 3 {
		t.Fatalf("mass radius = %v", got)
	}
	if got := MassRadius(Mass{}); got != 3 {
		t.Fatalf("default mass radius = %v", got)
	}
	if got := MassRadius(Mass{Mass: -0.2}); got != 1 {
		t.Fatalf("minimum radius = %v", got)
	}
	if got := MassRadius(Mass{Mass: -0.1}); got != 1 {
		t.Fatalf("zero radius clamps to minimum = %v", got)
	}
	if got := MassRadius(Mass{Mass: 1, Fixed: true}); got != fixedMassCollisionRadius {
		t.Fatalf("fixed radius = %v", got)
	}
	if got := MassRadius(Mass{Mass: 2.7762217705032126e13}); got != 64 {
		t.Fatalf("radius 64 stays in range = %v", got)
	}
	if got := MassRadius(Mass{Mass: 1e20}); got != 64 {
		t.Fatalf("large radius = %v", got)
	}
}
