//go:build appunit

package app

import (
	"testing"

	"springs/internal/sim"
)

func TestAppUnitOverlayAndDialogRepeatHelpers(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1})

	game.overlays.value = valueDialog{Open: true, Text: "1", Min: 0, Max: 10}
	increment := game.valueDialogIncrementRect()
	if !game.clickOpenOverlay(increment.Min.X, increment.Min.Y) || game.overlays.value.Text != "1.1" {
		t.Fatalf("value overlay text = %q", game.overlays.value.Text)
	}

	game.overlays.value.Open = false
	game.overlays.massMenu = massContextMenu{Open: true, MassID: 1, X: 10, Y: 10}
	if !game.clickOpenOverlay(0, 0) || game.overlays.massMenu.Open {
		t.Fatalf("mass overlay open = %t", game.overlays.massMenu.Open)
	}

	game.controls.demoPickerOpen = true
	if !game.clickOpenOverlay(0, 0) || game.controls.demoPickerOpen {
		t.Fatalf("demo picker open = %t", game.controls.demoPickerOpen)
	}
	if game.clickOpenOverlay(0, 0) {
		t.Fatal("closed overlays should not handle click")
	}

	game.overlays.value = valueDialog{Open: true, Text: "1", Min: 0, Max: 10}
	game.controls.activeValueStep = numericStepAmount
	game.controls.valueStepTicks = numericStepHoldDelayTicks - 1
	game.continueValueDialogStepHold()
	if game.overlays.value.Text != "1.1" {
		t.Fatalf("repeated value text = %q", game.overlays.value.Text)
	}
	game.overlays.value.Open = false
	game.continueValueDialogStepHold()
	if game.controls.activeValueStep != 0 || game.controls.valueStepTicks != 0 {
		t.Fatalf("closed value repeat state step=%f ticks=%d", game.controls.activeValueStep, game.controls.valueStepTicks)
	}
}
