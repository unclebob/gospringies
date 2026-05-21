//go:build appunit

package app

import (
	"testing"

	"springs/internal/sim"
)

func appUnitGameWithMasses(masses ...sim.Mass) *Game {
	game := NewGame()
	world := sim.NewWorld()
	for _, mass := range masses {
		if mass.Mass == 0 {
			mass.Mass = 1
		}
		_ = world.AddMass(mass)
	}
	game.ReplaceWorld(world)
	return game
}

func mustVisibleControl(t *testing.T, name string) controlBox {
	t.Helper()
	control, ok := visibleControlWithName(name)
	if !ok {
		t.Fatalf("missing visible control %q", name)
	}
	return control
}
