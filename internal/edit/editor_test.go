package edit

import (
	"testing"

	"springs/internal/sim"
)

func TestClickAddMassUsesDefaults(t *testing.T) {
	world := sim.NewWorld()
	world.Parameters.Set("current mass", "2.5")
	world.Parameters.Set("elasticity", "0.6")
	editor := NewEditor(world)
	editor.Mode = ModeAddMass

	id, err := editor.Click(sim.Vec2{X: 120, Y: 80})

	if err != nil {
		t.Fatal(err)
	}
	mass, ok := world.MassByID(id)
	if !ok || mass.Position != (sim.Vec2{X: 120, Y: 80}) || mass.Mass != 2.5 || mass.Elasticity != 0.6 {
		t.Fatalf("mass = %#v, ok = %t", mass, ok)
	}
}

func TestClickRejectsUnsupportedMode(t *testing.T) {
	editor := NewEditor(sim.NewWorld())
	editor.Mode = ModeAddSpring

	if _, err := editor.Click(sim.Vec2{}); err == nil {
		t.Fatal("expected unsupported click mode")
	}
}

func TestGridSnapAffectsPlacement(t *testing.T) {
	world := sim.NewWorld()
	editor := NewEditor(world)
	editor.Mode = ModeAddMass
	editor.GridSnapEnabled = true
	editor.GridSnapSize = 10

	id, err := editor.Click(sim.Vec2{X: 123, Y: 87})

	if err != nil {
		t.Fatal(err)
	}
	mass, _ := world.MassByID(id)
	if mass.Position != (sim.Vec2{X: 120, Y: 90}) {
		t.Fatalf("position = %#v", mass.Position)
	}
}

func TestGridSnapIgnoresInvalidSize(t *testing.T) {
	world := sim.NewWorld()
	editor := NewEditor(world)
	editor.Mode = ModeAddMass
	editor.GridSnapEnabled = true

	id, err := editor.Click(sim.Vec2{X: 123, Y: 87})

	if err != nil {
		t.Fatal(err)
	}
	mass, _ := world.MassByID(id)
	if mass.Position != (sim.Vec2{X: 123, Y: 87}) {
		t.Fatalf("position = %#v", mass.Position)
	}
}

func TestCreateSpringUsesDefaults(t *testing.T) {
	world := sim.NewWorld()
	world.Parameters.Set("spring constant", "14")
	world.Parameters.Set("damping", "0.3")
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	editor := NewEditor(world)
	editor.Mode = ModeAddSpring

	id, err := editor.CreateSpring(1, 2)

	if err != nil {
		t.Fatal(err)
	}
	spring, ok := world.SpringByID(id)
	if !ok || spring.MassA != 1 || spring.MassB != 2 || spring.SpringConstant != 14 || spring.Damping != 0.3 || spring.RestLength != 5 {
		t.Fatalf("spring = %#v, ok = %t", spring, ok)
	}
}

func TestCreateSpringReportsUnsupportedModeAndMissingEndpoint(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	editor := NewEditor(world)

	if _, err := editor.CreateSpring(1, 2); err == nil {
		t.Fatal("expected unsupported spring mode")
	}

	editor.Mode = ModeAddSpring
	if _, err := editor.CreateSpring(1, 2); err == nil {
		t.Fatal("expected missing spring endpoint")
	}
}

func TestDragMovesOnlyMovableMasses(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1, Fixed: true})
	editor := NewEditor(world)

	if err := editor.DragMass(1, sim.Vec2{X: 40, Y: 50}); err != nil {
		t.Fatal(err)
	}
	if err := editor.DragMass(2, sim.Vec2{X: 40, Y: 50}); err != nil {
		t.Fatal(err)
	}

	movable, _ := world.MassByID(1)
	fixed, _ := world.MassByID(2)
	if movable.Position != (sim.Vec2{X: 40, Y: 50}) || fixed.Position != (sim.Vec2{X: 10, Y: 10}) {
		t.Fatalf("movable = %#v fixed = %#v", movable, fixed)
	}
}

func TestDragReportsMissingMass(t *testing.T) {
	editor := NewEditor(sim.NewWorld())

	if err := editor.DragMass(99, sim.Vec2{}); err == nil {
		t.Fatal("expected missing mass")
	}
}
