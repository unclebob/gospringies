package sim

import "testing"

func TestEnabledWallsBounceWithElasticity(t *testing.T) {
	world := NewWorld()
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
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 1, Y: 20}, Velocity: Vec2{X: -4}, Mass: 1, Elasticity: 1})

	world.Step(1)

	mass, _ := world.MassByID(1)
	if mass.Velocity.X != -4 || mass.Position.X >= 0 {
		t.Fatalf("mass = %#v", mass)
	}
}

func TestStickinessCanHoldAndReleaseMass(t *testing.T) {
	world := NewWorld()
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
	world.Masses[0].Position = Vec2{X: 20, Y: world.Bounds.Height}
	world.Step(1)
	mass, _ = world.MassByID(1)
	if mass.StuckWall != "" {
		t.Fatalf("released mass = %#v", mass)
	}
}

func TestFixedMassesIgnoreWallCollision(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableWall("left")
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 1, Y: 20}, Velocity: Vec2{X: -4}, Mass: 1, Elasticity: 1, Fixed: true})

	world.Step(1)

	mass, _ := world.MassByID(1)
	if mass.Position.X != 1 || mass.Velocity.X != -4 {
		t.Fatalf("fixed mass = %#v", mass)
	}
}

func TestWallBounceStickinessThresholds(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("stickiness", "2")
	exact := Mass{ID: 1, Velocity: Vec2{X: -4}, Elasticity: 0.5}
	world.bounceOrStick(&exact, namedCollisionWall(t, world, &exact, "left"))
	if exact.StuckWall != "left" || exact.Velocity.X != 0 {
		t.Fatalf("exact threshold should stick: %#v", exact)
	}

	above := Mass{ID: 2, Velocity: Vec2{X: -5}, Elasticity: 0.5}
	world.bounceOrStick(&above, namedCollisionWall(t, world, &above, "left"))
	if above.StuckWall != "" || above.Velocity.X != 0.5 {
		t.Fatalf("above threshold should rebound: %#v", above)
	}

	right := Mass{ID: 3, Velocity: Vec2{X: 4}, Elasticity: 1}
	world.Parameters.Set("stickiness", "0")
	world.bounceOrStick(&right, namedCollisionWall(t, world, &right, "right"))
	if right.Velocity.X != -4 {
		t.Fatalf("right wall rebound = %#v", right)
	}

	top := Mass{ID: 4, Velocity: Vec2{Y: 3}, Elasticity: 1}
	world.bounceOrStick(&top, namedCollisionWall(t, world, &top, "top"))
	if top.Velocity.Y != -3 {
		t.Fatalf("top wall rebound = %#v", top)
	}

	bottom := Mass{ID: 5, Velocity: Vec2{Y: -3}, Elasticity: 1}
	world.bounceOrStick(&bottom, namedCollisionWall(t, world, &bottom, "bottom"))
	if bottom.Velocity.Y != 3 {
		t.Fatalf("bottom wall rebound = %#v", bottom)
	}
}

func TestWallCollisionActivationBoundaries(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableWall("left")

	atBoundary := Mass{ID: 1, Position: Vec2{X: 0}, Velocity: Vec2{X: -1}}
	if world.wallCollisionActive(namedCollisionWall(t, world, &atBoundary, "left")) {
		t.Fatal("left wall active at boundary")
	}

	inside := Mass{ID: 2, Position: Vec2{X: 0.5}, Velocity: Vec2{X: -1}}
	if world.wallCollisionActive(namedCollisionWall(t, world, &inside, "left")) {
		t.Fatal("left wall active inside boundary")
	}

	notMovingOut := Mass{ID: 3, Position: Vec2{X: -1}, Velocity: Vec2{}}
	if world.wallCollisionActive(namedCollisionWall(t, world, &notMovingOut, "left")) {
		t.Fatal("left wall active with zero normal velocity")
	}

	movingInside := Mass{ID: 4, Position: Vec2{X: -1}, Velocity: Vec2{X: 0.5}}
	if world.wallCollisionActive(namedCollisionWall(t, world, &movingInside, "left")) {
		t.Fatal("left wall active while moving inward")
	}
}

func TestAllWallCollisionActivationBoundaries(t *testing.T) {
	world := NewWorld()
	for _, wall := range []string{"right", "bottom", "top"} {
		world.Parameters.EnableWall(wall)
	}

	cases := []struct {
		name         string
		wall         string
		position     Vec2
		velocity     Vec2
		active       bool
		boundaryAxis float64
	}{
		{name: "right at boundary", wall: "right", position: Vec2{X: world.Bounds.Width}, velocity: Vec2{X: 1}},
		{name: "right outside moving outward", wall: "right", position: Vec2{X: world.Bounds.Width + 1}, velocity: Vec2{X: 1}, active: true, boundaryAxis: world.Bounds.Width},
		{name: "right outside with zero normal velocity", wall: "right", position: Vec2{X: world.Bounds.Width + 1}, velocity: Vec2{}},
		{name: "right outside moving inward", wall: "right", position: Vec2{X: world.Bounds.Width + 1}, velocity: Vec2{X: -1}},
		{name: "bottom at boundary", wall: "bottom", position: Vec2{Y: 0}, velocity: Vec2{Y: -1}},
		{name: "bottom outside moving outward", wall: "bottom", position: Vec2{Y: -1}, velocity: Vec2{Y: -1}, active: true},
		{name: "bottom outside with zero normal velocity", wall: "bottom", position: Vec2{Y: -1}, velocity: Vec2{}},
		{name: "bottom outside moving inward", wall: "bottom", position: Vec2{Y: -1}, velocity: Vec2{Y: 1}},
		{name: "top at boundary", wall: "top", position: Vec2{Y: world.Bounds.Height}, velocity: Vec2{Y: 1}},
		{name: "top outside moving outward", wall: "top", position: Vec2{Y: world.Bounds.Height + 1}, velocity: Vec2{Y: 1}, active: true, boundaryAxis: world.Bounds.Height},
		{name: "top outside with zero normal velocity", wall: "top", position: Vec2{Y: world.Bounds.Height + 1}, velocity: Vec2{}},
		{name: "top outside moving inward", wall: "top", position: Vec2{Y: world.Bounds.Height + 1}, velocity: Vec2{Y: -1}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mass := Mass{ID: 1, Position: tt.position, Velocity: tt.velocity}
			wall := namedCollisionWall(t, world, &mass, tt.wall)
			if active := world.wallCollisionActive(wall); active != tt.active {
				t.Fatalf("active = %t, want %t", active, tt.active)
			}
			if tt.active && wall.boundary != tt.boundaryAxis {
				t.Fatalf("boundary = %v, want %v", wall.boundary, tt.boundaryAxis)
			}
		})
	}
}

func TestBottomWallCollisionContract(t *testing.T) {
	world := NewWorld()
	world.Parameters.EnableWall("bottom")
	mass := Mass{ID: 1, Position: Vec2{Y: -1}, Velocity: Vec2{Y: -2}}

	wall := namedCollisionWall(t, world, &mass, "bottom")
	if wall.name != "bottom" || wall.boundary != 0 {
		t.Fatalf("bottom wall = %#v", wall)
	}
	if !wall.outside(mass.Position.Y) || !wall.movingOutward(mass.Velocity.Y) {
		t.Fatal("bottom wall should be active for mass below the boundary and moving downward")
	}
	if wall.releaseForce(Vec2{Y: 3}) != 3 {
		t.Fatal("bottom wall release force should use positive Y force")
	}
}

func TestStuckWallContracts(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("stickiness", "10")
	mass := Mass{ID: 1, Position: Vec2{X: 0, Y: 2}, Velocity: Vec2{Y: 3}, StuckWall: "left"}

	if _, ok := world.stuckWall(&mass); !ok {
		t.Fatal("expected left stuck wall")
	}
	if world.wallReleasedBy(namedCollisionWall(t, world, &mass, "left"), Vec2{X: 10}) {
		t.Fatal("release force equal to stickiness should not release")
	}
	if !world.wallReleasedBy(namedCollisionWall(t, world, &mass, "right"), Vec2{X: -11}) {
		t.Fatal("right wall should release from leftward force")
	}
	if !world.wallReleasedBy(namedCollisionWall(t, world, &mass, "bottom"), Vec2{Y: 11}) {
		t.Fatal("bottom wall should release from upward force")
	}
	if !world.keepStuck(&mass, Vec2{X: 5}) {
		t.Fatal("mass should remain stuck")
	}
	if mass.Position.X != 0 || mass.Velocity.X != 0 || mass.Velocity.Y != 3 {
		t.Fatalf("kept stuck mass = %#v", mass)
	}

	invalid := Mass{ID: 2, StuckWall: "missing"}
	if world.keepStuck(&invalid, Vec2{}) {
		t.Fatal("invalid stuck wall should not remain stuck")
	}
	if invalid.StuckWall != "" {
		t.Fatalf("invalid stuck wall was not cleared: %#v", invalid)
	}
}

func TestStuckWallKeepsOnlyNormalVelocityZero(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("stickiness", "10")
	cases := []struct {
		wall     string
		position Vec2
		velocity Vec2
		want     Vec2
	}{
		{wall: "right", position: Vec2{X: world.Bounds.Width, Y: 2}, velocity: Vec2{X: 3, Y: 4}, want: Vec2{Y: 4}},
		{wall: "bottom", position: Vec2{X: 2, Y: 0}, velocity: Vec2{X: 3, Y: -4}, want: Vec2{X: 3}},
		{wall: "top", position: Vec2{X: 2, Y: world.Bounds.Height}, velocity: Vec2{X: 3, Y: 4}, want: Vec2{X: 3}},
	}
	for _, tt := range cases {
		t.Run(tt.wall, func(t *testing.T) {
			mass := Mass{ID: 1, Position: tt.position, Velocity: tt.velocity, StuckWall: tt.wall}
			if !world.keepStuck(&mass, Vec2{}) {
				t.Fatal("mass should remain stuck")
			}
			if mass.Velocity != tt.want {
				t.Fatalf("velocity = %#v, want %#v", mass.Velocity, tt.want)
			}
		})
	}
}

func namedCollisionWall(t *testing.T, world *Simulation, mass *Mass, name string) wallCollision {
	t.Helper()
	for _, wall := range world.collisionWalls(mass) {
		if wall.name == name {
			return wall
		}
	}
	t.Fatalf("wall %q not found", name)
	return wallCollision{}
}
