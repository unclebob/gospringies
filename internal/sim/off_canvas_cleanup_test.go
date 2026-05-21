package sim

import "testing"

func TestStepDeletesMassesBeyondOneCanvasHeightFromCanvas(t *testing.T) {
	world := NewWorld()
	world.Bounds = Bounds{Width: 200, Height: 100}
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 50, Y: 50}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 50, Y: 201}, Mass: 1, Fixed: true})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2})

	world.Step(0)

	if _, ok := world.MassByID(2); ok {
		t.Fatal("off-canvas mass remained present")
	}
	if _, ok := world.MassByID(1); !ok {
		t.Fatal("inside mass was deleted")
	}
	if len(world.Masses) != 1 {
		t.Fatalf("mass count = %d, expected 1", len(world.Masses))
	}
	if len(world.Springs) != 0 {
		t.Fatalf("spring count = %d, expected attached spring to be deleted", len(world.Springs))
	}
}

func TestStepDeletesMassBeyondRightCleanupBoundary(t *testing.T) {
	world := NewWorld()
	world.Bounds = Bounds{Width: 200, Height: 100}
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 301, Y: 50}, Mass: 1, Fixed: true})

	world.Step(0)

	if len(world.Masses) != 0 {
		t.Fatalf("mass count = %d, expected right boundary cleanup", len(world.Masses))
	}
}

func TestStepRetainsMassesAtOffCanvasCleanupBoundary(t *testing.T) {
	world := NewWorld()
	world.Bounds = Bounds{Width: 200, Height: 100}
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: -100, Y: 50}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 300, Y: 50}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: 50, Y: -100}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 4, Position: Vec2{X: 50, Y: 200}, Mass: 1, Fixed: true})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2})

	world.Step(0)

	if len(world.Masses) != 4 {
		t.Fatalf("mass count = %d, expected boundary masses retained", len(world.Masses))
	}
	if len(world.Springs) != 1 {
		t.Fatalf("spring count = %d, expected spring retained", len(world.Springs))
	}
}

func TestStepReindexesSpringsAfterOffCanvasCleanup(t *testing.T) {
	world := NewWorld()
	world.Bounds = Bounds{Width: 200, Height: 100}
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: -101, Y: 50}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 40, Y: 50}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: 80, Y: 50}, Mass: 1, Fixed: true})
	_ = world.AddSpring(Spring{ID: 1, MassA: 2, MassB: 3})

	world.Step(0)

	spring, ok := world.SpringByID(1)
	if !ok {
		t.Fatal("spring between remaining masses was deleted")
	}
	if spring.A != 0 || spring.B != 1 {
		t.Fatalf("spring indexes = %d,%d, expected reindexed to remaining masses", spring.A, spring.B)
	}
}

func TestAssignSpringEndpointIDsBackfillsOnlyMissingValidEndpoints(t *testing.T) {
	world := NewWorld()
	world.Masses = []Mass{
		{ID: 10, Mass: 1},
		{ID: 20, Mass: 1},
	}
	world.Springs = []Spring{
		{ID: 1, A: 0, B: 1},
		{ID: 2, A: 3, B: -1},
		{ID: 3, A: 0, B: 1, MassA: 30, MassB: 40},
	}

	world.assignSpringEndpointIDs()

	if world.Springs[0].MassA != 10 || world.Springs[0].MassB != 20 {
		t.Fatalf("valid legacy endpoint indexes were not backfilled: %#v", world.Springs[0])
	}
	if world.Springs[1].MassA != 0 || world.Springs[1].MassB != 0 {
		t.Fatalf("invalid legacy endpoint indexes were backfilled: %#v", world.Springs[1])
	}
	if world.Springs[2].MassA != 30 || world.Springs[2].MassB != 40 {
		t.Fatalf("existing endpoint IDs were overwritten: %#v", world.Springs[2])
	}
}

func TestValidMassIndexRejectsLength(t *testing.T) {
	world := NewWorld()
	world.Masses = []Mass{{ID: 1, Mass: 1}}

	if world.validMassIndex(len(world.Masses)) {
		t.Fatal("length should be outside valid mass indexes")
	}
}
