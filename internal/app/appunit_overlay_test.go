//go:build appunit

package app

import (
	"testing"

	"springs/internal/sim"
)

func TestAppUnitOverlayAndDialogRepeatHelpers(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1})

	game.valueDialog = valueDialog{Open: true, Text: "1", Min: 0, Max: 10}
	increment := game.valueDialogIncrementRect()
	if !game.clickOpenOverlay(increment.Min.X, increment.Min.Y) || game.valueDialog.Text != "1.1" {
		t.Fatalf("value overlay text = %q", game.valueDialog.Text)
	}

	game.valueDialog.Open = false
	game.massMenu = massContextMenu{Open: true, MassID: 1, X: 10, Y: 10}
	if !game.clickOpenOverlay(0, 0) || game.massMenu.Open {
		t.Fatalf("mass overlay open = %t", game.massMenu.Open)
	}

	game.controls.demoPickerOpen = true
	if !game.clickOpenOverlay(0, 0) || game.controls.demoPickerOpen {
		t.Fatalf("demo picker open = %t", game.controls.demoPickerOpen)
	}
	if game.clickOpenOverlay(0, 0) {
		t.Fatal("closed overlays should not handle click")
	}

	game.valueDialog = valueDialog{Open: true, Text: "1", Min: 0, Max: 10}
	game.controls.activeValueStep = numericStepAmount
	game.controls.valueStepTicks = numericStepHoldDelayTicks - 1
	game.continueValueDialogStepHold()
	if game.valueDialog.Text != "1.1" {
		t.Fatalf("repeated value text = %q", game.valueDialog.Text)
	}
	game.valueDialog.Open = false
	game.continueValueDialogStepHold()
	if game.controls.activeValueStep != 0 || game.controls.valueStepTicks != 0 {
		t.Fatalf("closed value repeat state step=%f ticks=%d", game.controls.activeValueStep, game.controls.valueStepTicks)
	}
}
