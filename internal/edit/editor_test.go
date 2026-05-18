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

	id, err := editor.Click(sim.Vec2{})
	if err == nil {
		t.Fatal("expected unsupported click mode")
	}
	if id != 0 {
		t.Fatalf("id = %d, want 0", id)
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
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 1})
	editor := NewEditor(world)

	if id, err := editor.CreateSpring(1, 2); err == nil || id != 0 {
		t.Fatal("expected unsupported spring mode")
	}

	editor.Mode = ModeAddSpring
	if id, err := editor.CreateSpring(1, 99); err == nil || id != 0 {
		t.Fatalf("expected missing second spring endpoint, id = %d err = %v", id, err)
	}
	if id, err := editor.CreateSpring(99, 2); err == nil || id != 0 {
		t.Fatalf("expected missing first spring endpoint, id = %d err = %v", id, err)
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

func TestEditorMathHelpersCoverBoundaryCases(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 1})
	if nextMassID(world) != 2 || nextSpringID(world) != 2 {
		t.Fatalf("next IDs = mass %d spring %d", nextMassID(world), nextSpringID(world))
	}
	editor := NewEditor(world)
	editor.GridSnapEnabled = true
	editor.GridSnapSize = 1
	if got := editor.snap(sim.Vec2{X: 1.4, Y: 2.6}); got != (sim.Vec2{X: 1, Y: 3}) {
		t.Fatalf("snap = %#v", got)
	}
	if got := distance(sim.Vec2{X: 3, Y: 0}, sim.Vec2{X: 1, Y: 0}); got != 2 {
		t.Fatalf("distance = %f", got)
	}
	if got := distance(sim.Vec2{X: 0, Y: 4}, sim.Vec2{X: 0, Y: 1}); got != 3 {
		t.Fatalf("vertical distance = %f", got)
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

func TestSelectReportsMissingObjects(t *testing.T) {
	editor := NewEditor(sim.NewWorld())

	if err := editor.SelectMass(1); err == nil {
		t.Fatal("expected missing mass")
	}
	if err := editor.SelectSpring(1); err == nil {
		t.Fatal("expected missing spring")
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

func TestWorldIndexByMassIDReportsMissingMass(t *testing.T) {
	editor := NewEditor(sim.NewWorld())

	index, ok := editor.worldIndexByMassID(99)

	if index != 0 || ok {
		t.Fatalf("index = %d ok = %t, want 0 false", index, ok)
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
	if !editor.MassSelected(duplicated.MassIDs[0]) || !editor.MassSelected(duplicated.MassIDs[1]) || !editor.SpringSelected(duplicated.SpringIDs[0]) {
		t.Fatalf("duplicate selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
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

func TestSelectNearestReplacesOrTogglesSelection(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 0}, Mass: 1})
	editor := NewEditor(world)
	editor.SelectedMasses[1] = true

	if err := editor.SelectNearest(sim.Vec2{X: 19, Y: 0}, false); err != nil {
		t.Fatal(err)
	}
	if editor.MassSelected(1) || !editor.MassSelected(2) {
		t.Fatalf("selection = %#v", editor.SelectedMasses)
	}

	if err := editor.SelectNearest(sim.Vec2{X: 1, Y: 0}, true); err != nil {
		t.Fatal(err)
	}
	if !editor.MassSelected(1) || !editor.MassSelected(2) {
		t.Fatalf("selection = %#v", editor.SelectedMasses)
	}
	if err := editor.SelectNearest(sim.Vec2{X: 1, Y: 0}, true); err != nil {
		t.Fatal(err)
	}
	if editor.MassSelected(1) || !editor.MassSelected(2) {
		t.Fatalf("selection = %#v", editor.SelectedMasses)
	}

	tieWorld := sim.NewWorld()
	_ = tieWorld.AddMass(sim.Mass{ID: 10, Position: sim.Vec2{X: 0}, Mass: 1})
	_ = tieWorld.AddMass(sim.Mass{ID: 20, Position: sim.Vec2{X: 2}, Mass: 1})
	tieEditor := NewEditor(tieWorld)
	if id, ok := tieEditor.nearestMassID(sim.Vec2{X: 1}); id != 10 || !ok {
		t.Fatalf("tie nearest = id %d ok %t", id, ok)
	}
	if id, ok := NewEditor(sim.NewWorld()).nearestMassID(sim.Vec2{}); id != 0 || ok {
		t.Fatalf("empty nearest = id %d ok %t", id, ok)
	}
}

func TestBoxSelectAndShiftBoxSelect(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1})
	editor := NewEditor(world)

	editor.BoxSelect(sim.Vec2{}, sim.Vec2{X: 50, Y: 50}, false)
	if !editor.MassSelected(1) || !editor.MassSelected(2) || editor.MassSelected(3) {
		t.Fatalf("selection = %#v", editor.SelectedMasses)
	}

	editor.clearSelection()
	editor.SelectedMasses[3] = true
	editor.BoxSelect(sim.Vec2{}, sim.Vec2{X: 15, Y: 15}, true)
	if !editor.MassSelected(1) || editor.MassSelected(2) || !editor.MassSelected(3) {
		t.Fatalf("selection = %#v", editor.SelectedMasses)
	}

	for _, point := range []sim.Vec2{{X: 0, Y: 0}, {X: 15, Y: 15}, {X: 0, Y: 15}, {X: 15, Y: 0}} {
		if !withinBox(point, sim.Vec2{X: 15, Y: 15}, sim.Vec2{}) {
			t.Fatalf("point %#v should be inside reversed box", point)
		}
	}
	for _, point := range []sim.Vec2{{X: -1, Y: 10}, {X: 16, Y: 10}, {X: 10, Y: -1}, {X: 10, Y: 16}} {
		if withinBox(point, sim.Vec2{}, sim.Vec2{X: 15, Y: 15}) {
			t.Fatalf("point %#v should be outside box", point)
		}
	}
}

func TestBoxSelectSelectsFullyEnclosedSpring(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 40, Y: 40}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 30, SpringConstant: 1})
	editor := NewEditor(world)

	editor.BoxSelect(sim.Vec2{}, sim.Vec2{X: 50, Y: 50}, false)

	if !editor.MassSelected(1) || !editor.MassSelected(2) || !editor.SpringSelected(3) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}
}

func TestBoxSelectSelectsPartOfSpringWhenNothingElseIsEnclosed(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 100, Y: 10}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 90, SpringConstant: 1})
	editor := NewEditor(world)

	editor.BoxSelect(sim.Vec2{X: 40, Y: 0}, sim.Vec2{X: 60, Y: 20}, false)

	if editor.MassSelected(1) || editor.MassSelected(2) || !editor.SpringSelected(3) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}
}

func TestBoxSelectDoesNotSelectPartialSpringWhenAnythingElseIsEnclosed(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 100, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 4, Position: sim.Vec2{X: 50, Y: 15}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 90, SpringConstant: 1})
	editor := NewEditor(world)

	editor.BoxSelect(sim.Vec2{X: 40, Y: 0}, sim.Vec2{X: 60, Y: 20}, false)

	if !editor.MassSelected(4) || editor.SpringSelected(3) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}
}

func TestBoxSelectDoesNotSelectMultiplePartialSprings(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 100, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 4, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 5, Position: sim.Vec2{X: 100, Y: 20}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 90, SpringConstant: 1})
	_ = world.AddSpring(sim.Spring{ID: 6, MassA: 4, MassB: 5, RestLength: 90, SpringConstant: 1})
	editor := NewEditor(world)

	editor.BoxSelect(sim.Vec2{X: 40, Y: 0}, sim.Vec2{X: 60, Y: 30}, false)

	if editor.SpringSelected(3) || editor.SpringSelected(6) {
		t.Fatalf("selection = %#v %#v", editor.SelectedMasses, editor.SelectedSprings)
	}
}

func TestReindexSpringsRequiresBothEndpoints(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 10, Mass: 1})
	world.Springs = append(world.Springs, sim.Spring{ID: 1, A: 7, B: 8, MassA: 10, MassB: 99})
	editor := NewEditor(world)

	editor.reindexSprings()

	spring, _ := world.SpringByID(1)
	if spring.A != 7 || spring.B != 8 {
		t.Fatalf("spring indexes = %#v", spring)
	}
}

func TestMoveAndThrowSelectedSkipFixedMasses(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1, Fixed: true})
	editor := NewEditor(world)
	editor.SelectedMasses[1] = true
	editor.SelectedMasses[2] = true

	editor.MoveSelected(sim.Vec2{X: 5, Y: -3})
	editor.ThrowSelected(sim.Vec2{X: 4, Y: -2})

	movable, _ := world.MassByID(1)
	fixed, _ := world.MassByID(2)
	if movable.Position != (sim.Vec2{X: 15, Y: 7}) || movable.Velocity != (sim.Vec2{X: 4, Y: -2}) {
		t.Fatalf("movable = %#v", movable)
	}
	if fixed.Position != (sim.Vec2{X: 20, Y: 20}) || fixed.Velocity != (sim.Vec2{}) {
		t.Fatalf("fixed = %#v", fixed)
	}
}

func TestSpringPointerCreatesOrDiscardsSpringByReleaseTarget(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 0}, Mass: 1})
	editor := NewEditor(world)
	editor.Mode = ModeAddSpring

	if err := editor.BeginSpring(sim.Vec2{X: 1, Y: 0}, SpringButtonLeft); err != nil {
		t.Fatal(err)
	}
	id, created, err := editor.ReleaseSpring(sim.Vec2{X: 30, Y: 1})

	if err != nil {
		t.Fatal(err)
	}
	if !created || id != 1 {
		t.Fatalf("created = %t id = %d", created, id)
	}
	spring, ok := world.SpringByID(id)
	if !ok || spring.MassA != 1 || spring.MassB != 2 {
		t.Fatalf("spring = %#v, ok = %t", spring, ok)
	}

	if err := editor.BeginSpring(sim.Vec2{X: 0, Y: 0}, SpringButtonLeft); err != nil {
		t.Fatal(err)
	}
	if _, created, err := editor.ReleaseSpring(sim.Vec2{X: 100, Y: 100}); err != nil || created {
		t.Fatalf("created = %t err = %v", created, err)
	}
}

func TestSpringPointerButtonBehavior(t *testing.T) {
	tests := []struct {
		button    string
		active    bool
		temporary bool
	}{
		{button: SpringButtonLeft, active: true},
		{button: SpringButtonMiddle, active: true, temporary: true},
		{button: SpringButtonRight},
	}

	for _, test := range tests {
		t.Run(test.button, func(t *testing.T) {
			world := sim.NewWorld()
			_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
			editor := NewEditor(world)
			editor.Mode = ModeAddSpring

			if err := editor.BeginSpring(sim.Vec2{}, test.button); err != nil {
				t.Fatal(err)
			}
			editor.DragSpring(sim.Vec2{X: 10})
			pending, ok := editor.PendingSpring()

			if !ok || pending.StartMassID != 1 || pending.Active != test.active || pending.Temporary != test.temporary || pending.Cursor != (sim.Vec2{X: 10}) {
				t.Fatalf("pending = %#v, ok = %t", pending, ok)
			}
		})
	}
}

func TestSpringPointerUsesDefaultsAndReleaseLength(t *testing.T) {
	world := sim.NewWorld()
	world.Parameters.Set("spring constant", "12")
	world.Parameters.Set("damping", "0.5")
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 30}, Mass: 1})
	editor := NewEditor(world)
	editor.Mode = ModeAddSpring

	if err := editor.BeginSpring(sim.Vec2{}, SpringButtonLeft); err != nil {
		t.Fatal(err)
	}
	id, created, err := editor.ReleaseSpring(sim.Vec2{X: 30})

	if err != nil {
		t.Fatal(err)
	}
	spring, ok := world.SpringByID(id)
	if !created || !ok || spring.SpringConstant != 12 || spring.Damping != 0.5 || spring.RestLength != 30 {
		t.Fatalf("created = %t spring = %#v ok = %t", created, spring, ok)
	}
}

func TestSelectedMassParameterControlsUpdateSelectedMasses(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1, Elasticity: 0.2})
	_ = world.AddMass(sim.Mass{ID: 2, Mass: 1, Elasticity: 0.2})
	editor := NewEditor(world)
	if err := editor.SelectMass(1); err != nil {
		t.Fatal(err)
	}

	if err := editor.ChangeControl("mass", "2.5"); err != nil {
		t.Fatal(err)
	}
	if err := editor.ChangeControl("elasticity", "0.6"); err != nil {
		t.Fatal(err)
	}
	if err := editor.ChangeControl("fixed", "true"); err != nil {
		t.Fatal(err)
	}

	selected, _ := world.MassByID(1)
	unselected, _ := world.MassByID(2)
	if selected.Mass != 2.5 || selected.Elasticity != 0.6 || !selected.Fixed {
		t.Fatalf("selected mass = %#v", selected)
	}
	if unselected.Mass != 1 || unselected.Elasticity != 0.2 || unselected.Fixed {
		t.Fatalf("unselected mass = %#v", unselected)
	}
}

func TestSelectedSpringParameterControlsUpdateSelectedSprings(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 10}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, SpringConstant: 8, Damping: 0.2})
	_ = world.AddSpring(sim.Spring{ID: 2, MassA: 1, MassB: 2, SpringConstant: 8, Damping: 0.2})
	editor := NewEditor(world)
	if err := editor.SelectSpring(1); err != nil {
		t.Fatal(err)
	}

	if err := editor.ChangeControl("Kspring", "15"); err != nil {
		t.Fatal(err)
	}
	if err := editor.ChangeControl("Kdamp", "0.8"); err != nil {
		t.Fatal(err)
	}

	selected, _ := world.SpringByID(1)
	unselected, _ := world.SpringByID(2)
	if selected.SpringConstant != 15 || selected.Stiffness != 15 || selected.Damping != 0.8 {
		t.Fatalf("selected spring = %#v", selected)
	}
	if unselected.SpringConstant != 8 || unselected.Damping != 0.2 {
		t.Fatalf("unselected spring = %#v", unselected)
	}
}

func TestSetRestLengthUsesCurrentSelectedSpringGeometry(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 1})
	editor := NewEditor(world)
	if err := editor.SelectSpring(1); err != nil {
		t.Fatal(err)
	}

	if err := editor.SetRestLength(); err != nil {
		t.Fatal(err)
	}

	spring, _ := world.SpringByID(1)
	if spring.RestLength != 5 {
		t.Fatalf("rest length = %f", spring.RestLength)
	}
}

func TestParameterControlsIgnoreIncompatibleSelectionsAndUpdateDefaults(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, SpringConstant: 8})
	editor := NewEditor(world)
	if err := editor.SelectSpring(1); err != nil {
		t.Fatal(err)
	}

	if err := editor.ChangeControl("mass", "3"); err != nil {
		t.Fatal(err)
	}
	editor.Mode = ModeAddMass
	massID, err := editor.Click(sim.Vec2{X: 30})
	if err != nil {
		t.Fatal(err)
	}
	createdMass, _ := world.MassByID(massID)
	if createdMass.Mass != 3 {
		t.Fatalf("created mass = %#v", createdMass)
	}

	if err := editor.SelectMass(1); err != nil {
		t.Fatal(err)
	}
	if err := editor.ChangeControl("Kspring", "20"); err != nil {
		t.Fatal(err)
	}
	editor.Mode = ModeAddSpring
	springID, err := editor.CreateSpring(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	createdSpring, _ := world.SpringByID(springID)
	if createdSpring.SpringConstant != 20 {
		t.Fatalf("created spring = %#v", createdSpring)
	}
}

func TestSpringPointerReleaseAndHitBoundaries(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	editor := NewEditor(world)
	editor.Mode = ModeAddSpring

	if id, created, err := editor.ReleaseSpring(sim.Vec2{}); err == nil || created || id != 0 {
		t.Fatalf("release without pending = id %d created %t err %v", id, created, err)
	}
	if pending, ok := editor.PendingSpring(); ok || pending != (PendingSpring{}) {
		t.Fatalf("empty pending = %#v ok=%t", pending, ok)
	}
	if id, ok := editor.massNear(sim.Vec2{}); id != 1 || !ok {
		t.Fatalf("near mass = id %d ok %t", id, ok)
	}
	if id, ok := editor.massNear(sim.Vec2{X: springHitRadius}); id != 1 || !ok {
		t.Fatalf("boundary mass = id %d ok %t", id, ok)
	}
	if id, ok := editor.massNear(sim.Vec2{X: springHitRadius + 1}); id != 0 || ok {
		t.Fatalf("outside mass = id %d ok %t", id, ok)
	}
	if id, ok := NewEditor(sim.NewWorld()).massNear(sim.Vec2{}); id != 0 || ok {
		t.Fatalf("empty-world mass = id %d ok %t", id, ok)
	}

	if err := editor.BeginSpring(sim.Vec2{}, SpringButtonMiddle); err != nil {
		t.Fatal(err)
	}
	if id, created, err := editor.ReleaseSpring(sim.Vec2{}); err != nil || created || id != 0 {
		t.Fatalf("temporary release = id %d created %t err %v", id, created, err)
	}
	if err := editor.BeginSpring(sim.Vec2{}, SpringButtonLeft); err != nil {
		t.Fatal(err)
	}
	if id, created, err := editor.ReleaseSpring(sim.Vec2{X: springHitRadius + 1}); err != nil || created || id != 0 {
		t.Fatalf("missed release = id %d created %t err %v", id, created, err)
	}
}
