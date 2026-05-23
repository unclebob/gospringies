//go:build appunit

package app

import (
	"math"
	"testing"

	"springs/internal/sim"
)

func TestAppUnitSelectionAndRenderHelpers(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 140, Y: 100}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})

	if err := game.SelectSprings(1); err != nil || !game.editing().SelectedSprings[1] {
		t.Fatalf("select springs err=%v selected=%#v", err, game.editing().SelectedSprings)
	}
	if err := game.SelectSprings(99); err == nil {
		t.Fatal("missing spring selection should fail")
	}

	representations := map[string]string{}
	game.springRepresentation(representations)
	if representations["spring"] != "cyan line" || representations["wall spring"] != "heavy orange line" {
		t.Fatalf("spring representations = %#v", representations)
	}

	game.World().Parameters.Set("show springs", "false")
	representations = map[string]string{}
	game.springRepresentation(representations)
	if len(representations) != 0 {
		t.Fatalf("hidden spring representations = %#v", representations)
	}
}

func TestAppUnitGridPointsFollowGridSnap(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("grid snap", "10")

	points := game.gridPoints()
	if len(points) == 0 {
		t.Fatal("expected grid points")
	}

	canvas := visibleRegionRects()["canvas"]
	_, _, minY, maxY := game.canvasWorldBounds()
	for _, point := range points {
		if math.Mod(point.X, 10) != 0 || math.Mod(point.Y, 10) != 0 {
			t.Fatalf("grid point = %#v, want multiples of 10", point)
		}
		if point.X < float64(canvas.Min.X) || point.X >= float64(canvas.Max.X) {
			t.Fatalf("grid point x = %f outside canvas %#v", point.X, canvas)
		}
		if point.Y < minY || point.Y > maxY {
			t.Fatalf("grid point y = %f outside bounds %f..%f", point.Y, minY, maxY)
		}
	}

	game.World().Parameters.Set("grid snap", "0")
	if points := game.gridPoints(); len(points) != 0 {
		t.Fatalf("grid point count = %d, want none when grid snap is disabled", len(points))
	}
}

func TestAppUnitPendingSpringLineStates(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 120}})

	if line, ok := game.pendingSpringLine(); ok || line != (selectionLine{}) {
		t.Fatalf("pending spring line without pending id = %#v, %t", line, ok)
	}

	game.pointer.pendingSpringID = 99
	if line, ok := game.pendingSpringLine(); ok || line != (selectionLine{}) {
		t.Fatalf("pending spring line with missing start = %#v, %t", line, ok)
	}

	game.pointer.pendingSpringID = 1
	game.pointer.pendingSpringEnd = sim.Vec2{X: 160, Y: 180}
	line, ok := game.pendingSpringLine()
	if !ok || line != (selectionLine{x1: 100, y1: 120, x2: 160, y2: 180}) {
		t.Fatalf("pending spring line = %#v, %t", line, ok)
	}
}

func TestAppUnitSelectionGeometryStates(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 40}},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	_ = game.World().AddSpring(sim.Spring{ID: 2, MassA: 1, MassB: 99})
	_ = game.World().AddSpring(sim.Spring{ID: 3, MassA: 99, MassB: 1})

	game.editing().SelectedMasses[1] = true
	explicit := game.explicitSelectedMasses()
	if len(explicit) != 1 || explicit[0].ID != 1 {
		t.Fatalf("explicit selected masses = %#v, want mass 1", explicit)
	}
	if selected := game.selectedMasses(); len(selected) != 1 || selected[0].ID != 1 {
		t.Fatalf("selected masses = %#v, want explicit mass 1", selected)
	}

	game.editing().SelectedMasses = map[int]bool{}
	game.editState.selected = true
	if selected := game.selectedMasses(); len(selected) != 2 {
		t.Fatalf("implicit selected masses = %#v, want both masses", selected)
	}

	game.editing().SelectedSprings = map[int]bool{1: true, 2: true, 3: true}
	if selected := game.selectedMasses(); len(selected) != 0 {
		t.Fatalf("spring-only selected masses = %#v, want none", selected)
	}
	lines := game.selectedSpringLines()
	if len(lines) != 1 || lines[0] != (selectionLine{x1: 10, y1: 20, x2: 30, y2: 40}) {
		t.Fatalf("selected spring lines = %#v, want one complete endpoint line", lines)
	}
}
