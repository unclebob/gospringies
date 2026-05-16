package edit

import (
	"testing"

	"springs/internal/sim"
)

func TestClickAddMassUsesDefaults(t *testing.T) {
	world := sim.NewWorld()
	world.Parameters.Set("current mass", "2.5")
	world.Parameters.Set("elasticity", "0.6")
	_ = world.AddMass(sim.Mass{ID: 7, Mass: 1})
	editor := NewEditor(world)
	editor.Mode = ModeAddMass

	id, err := editor.Click(sim.Vec2{X: 120, Y: 80})

	if err != nil {
		t.Fatal(err)
	}
	if id != 8 {
		t.Fatalf("id = %d, want 8", id)
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

func TestGridSnapPlacement(t *testing.T) {
	tests := []struct {
		name string
		size float64
		want sim.Vec2
	}{
		{name: "enabled", size: 10, want: sim.Vec2{X: 120, Y: 90}},
		{name: "invalid size", size: 0, want: sim.Vec2{X: 123, Y: 87}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			world := sim.NewWorld()
			editor := NewEditor(world)
			editor.Mode = ModeAddMass
			editor.GridSnapEnabled = true
			editor.GridSnapSize = test.size

			id, err := editor.Click(sim.Vec2{X: 123, Y: 87})

			if err != nil {
				t.Fatal(err)
			}
			mass, _ := world.MassByID(id)
			if mass.Position != test.want {
				t.Fatalf("position = %#v, want %#v", mass.Position, test.want)
			}
		})
	}
}

func TestCreateSpringUsesDefaults(t *testing.T) {
	world := sim.NewWorld()
	world.Parameters.Set("spring constant", "14")
	world.Parameters.Set("damping", "0.3")
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 4, MassA: 1, MassB: 2})
	editor := NewEditor(world)
	editor.Mode = ModeAddSpring

	id, err := editor.CreateSpring(1, 2)

	if err != nil {
		t.Fatal(err)
	}
	if id != 5 {
		t.Fatalf("id = %d, want 5", id)
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
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 90, Y: 90}, Mass: 1})
	editor := NewEditor(world)

	if err := editor.DragMass(1, sim.Vec2{X: 40, Y: 50}); err != nil {
		t.Fatal(err)
	}
	if err := editor.DragMass(2, sim.Vec2{X: 40, Y: 50}); err != nil {
		t.Fatal(err)
	}

	movable, _ := world.MassByID(1)
	fixed, _ := world.MassByID(2)
	untargeted, _ := world.MassByID(3)
	if movable.Position != (sim.Vec2{X: 40, Y: 50}) || fixed.Position != (sim.Vec2{X: 10, Y: 10}) {
		t.Fatalf("movable = %#v fixed = %#v", movable, fixed)
	}
	if untargeted.Position != (sim.Vec2{X: 90, Y: 90}) {
		t.Fatalf("untargeted = %#v", untargeted)
	}
}

func TestDragReportsMissingMass(t *testing.T) {
	editor := NewEditor(sim.NewWorld())

	if err := editor.DragMass(99, sim.Vec2{}); err == nil {
		t.Fatal("expected missing mass")
	}
}

func TestSelectIndividualObjects(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2})
	editor := NewEditor(world)

	if err := editor.SelectMass(1); err != nil {
		t.Fatal(err)
	}
	if !editor.MassSelected(1) || editor.SpringSelected(3) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}

	if err := editor.SelectSpring(3); err != nil {
		t.Fatal(err)
	}
	if editor.MassSelected(1) || !editor.SpringSelected(3) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}
}

func TestSelectAllObjects(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2})
	editor := NewEditor(world)

	editor.SelectAll()

	if !editor.MassSelected(1) || !editor.MassSelected(2) || !editor.SpringSelected(3) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}
}

func TestDeleteSelectedObjectsAndAttachedSprings(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2})
	editor := NewEditor(world)
	if err := editor.SelectMass(1); err != nil {
		t.Fatal(err)
	}

	editor.DeleteSelected()

	if _, ok := world.MassByID(1); ok {
		t.Fatal("mass 1 still exists")
	}
	if _, ok := world.SpringByID(3); ok {
		t.Fatal("spring 3 still exists")
	}
	if _, ok := world.MassByID(2); !ok {
		t.Fatal("mass 2 was deleted")
	}
}

func TestDeleteSelectedMassReindexesRemainingSprings(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 4, MassA: 2, MassB: 3})
	editor := NewEditor(world)
	if err := editor.SelectMass(1); err != nil {
		t.Fatal(err)
	}

	editor.DeleteSelected()

	spring, ok := world.SpringByID(4)
	if !ok || spring.A != 0 || spring.B != 1 {
		t.Fatalf("spring = %#v, ok = %t", spring, ok)
	}
}

func TestDuplicateSelectedObjectsCreatesIndependentIDs(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Velocity: sim.Vec2{X: 1}, Mass: 2, Elasticity: 0.5})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 20}, Mass: 3, Fixed: true})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 20, SpringConstant: 8, Damping: 0.4})
	editor := NewEditor(world)
	editor.SelectAll()

	duplicated, err := editor.DuplicateSelected()

	if err != nil {
		t.Fatal(err)
	}
	if len(duplicated.MassIDs) != 2 || len(duplicated.SpringIDs) != 1 {
		t.Fatalf("duplicated = %#v", duplicated)
	}
	dupMass, ok := world.MassByID(duplicated.MassIDs[0])
	if !ok || dupMass.ID == 1 || dupMass.Position != (sim.Vec2{X: 10, Y: 20}) || dupMass.Mass != 2 {
		t.Fatalf("duplicate mass = %#v, ok = %t", dupMass, ok)
	}
	dupSpring, ok := world.SpringByID(duplicated.SpringIDs[0])
	if !ok || dupSpring.ID == 3 || dupSpring.MassA == 1 || dupSpring.MassB == 2 {
		t.Fatalf("duplicate spring = %#v, ok = %t", dupSpring, ok)
	}
}
