//go:build appunit

package app

import (
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
