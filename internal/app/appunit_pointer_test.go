//go:build appunit

package app

import (
	"testing"

	"springs/internal/sim"
)

func TestAppUnitPointerPressBranches(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})

	game.beginPointerPress(sim.Vec2{X: 500, Y: 300}, 500, 300)
	if game.pointer.draggingMassID != 1 || !game.editing().MassSelected(1) {
		t.Fatalf("mass press state dragging=%d selected=%#v", game.pointer.draggingMassID, game.editing().SelectedMasses)
	}
	game.releasePointer(sim.Vec2{X: 500, Y: 300})

	game.beginPointerPress(sim.Vec2{X: 540, Y: 320}, 540, 320)
	game.releasePointer(sim.Vec2{X: 540, Y: 320})
	if len(game.World().Masses) != 2 || !game.editing().MassSelected(2) {
		t.Fatalf("empty canvas press masses=%#v selected=%#v", game.World().Masses, game.editing().SelectedMasses)
	}

	game.keyboard.controlDown = true
	game.beginPointerPress(sim.Vec2{X: 580, Y: 320}, 580, 320)
	if game.pointer.pendingSpringID == 0 || !game.pointer.springChainActive {
		t.Fatalf("control placement pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
	game.keyboard.controlDown = false
	game.beginPointerPress(sim.Vec2{X: 620, Y: 320}, 620, 320)
	if len(game.World().Springs) != 1 || game.pointer.pendingSpringID != 0 || game.pointer.springChainActive {
		t.Fatalf("chain finish springs=%#v pending=%d active=%t", game.World().Springs, game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
}

func TestAppUnitContinuePointerPressBranches(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})

	game.pointer.draggingMassID = 1
	game.pointer.draggingStart = sim.Vec2{X: 500, Y: 300}
	game.pointer.draggingLast = sim.Vec2{X: 500, Y: 300}
	game.continuePointerPress(sim.Vec2{X: 520, Y: 310}, 520)
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 520, Y: 310}) || !game.pointer.dragMoved {
		t.Fatalf("drag branch mass=%#v moved=%t", mass, game.pointer.dragMoved)
	}

	game.pointer.draggingMassID = 0
	game.pointer.pendingSpringID = 1
	game.continuePointerPress(sim.Vec2{X: 2000, Y: 2000}, 2000)
	if game.pointer.pendingSpringEnd != (sim.Vec2{X: 1340, Y: 1000}) {
		t.Fatalf("pending spring end = %#v", game.pointer.pendingSpringEnd)
	}

	game.pointer.pendingSpringID = 0
	game.pointer.selectionDrag = true
	game.continuePointerPress(sim.Vec2{X: 700, Y: 500}, 700)
	if game.pointer.selectionEnd != (sim.Vec2{X: 700, Y: 500}) {
		t.Fatalf("selection end = %#v", game.pointer.selectionEnd)
	}

	game.pointer.selectionDrag = false
	game.controls.activeNumericStep = "mass increment"
	game.controls.numericStepTicks = numericStepHoldDelayTicks - 1
	game.continuePointerPress(sim.Vec2{}, 0)
	if game.controls.numericStepTicks != numericStepHoldDelayTicks {
		t.Fatalf("numeric step ticks = %d", game.controls.numericStepTicks)
	}

	game.controls.activeNumericStep = ""
	game.controls.activeSlider = "speed slider"
	game.continuePointerPress(sim.Vec2{}, sliderTrack(mustVisibleControl(t, "speed slider")).Max.X)
	if game.run.simulationSpeed != maxSpeed {
		t.Fatalf("simulation speed = %f, want %f", game.run.simulationSpeed, maxSpeed)
	}
}

func TestAppUnitFinishWorldPointerCompletesGestures(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 540, Y: 300}, Mass: 1},
	)

	game.pointer.draggingMassID = 1
	game.pointer.draggingStart = sim.Vec2{X: 500, Y: 300}
	game.pointer.dragMoved = true
	game.keyboard.throwDown = true
	game.finishWorldPointer(sim.Vec2{X: 530, Y: 320})
	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 30, Y: 20}) {
		t.Fatalf("throw velocity = %#v", mass.Velocity)
	}

	game.keyboard.throwDown = false
	game.pointer.draggingMassID = 0
	game.pointer.pendingSpringID = 1
	game.finishWorldPointer(sim.Vec2{X: 540, Y: 300})
	if len(game.World().Springs) != 1 || game.pointer.pendingSpringID != 0 {
		t.Fatalf("spring finish springs=%#v pending=%d", game.World().Springs, game.pointer.pendingSpringID)
	}

	game.pointer.selectionDrag = true
	game.pointer.selectionStart = sim.Vec2{X: 700, Y: 300}
	game.finishWorldPointer(sim.Vec2{X: 700, Y: 300})
	if len(game.World().Masses) != 3 || game.pointer.selectionDrag {
		t.Fatalf("selection finish masses=%#v selectionDrag=%t", game.World().Masses, game.pointer.selectionDrag)
	}
}

func TestAppUnitFinishMassDragNonThrowBranches(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Velocity: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	game.pointer.draggingMassID = 1
	game.pointer.draggingStart = sim.Vec2{X: 500, Y: 300}
	game.pointer.dragMoved = true

	game.finishMassDrag(sim.Vec2{X: 530, Y: 320})
	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 3, Y: 4}) {
		t.Fatalf("non-throw drag velocity = %#v", mass.Velocity)
	}

	game.pointer.dragMoved = false
	game.pointer.selectionAdd = true
	game.finishMassDrag(sim.Vec2{X: 500, Y: 300})
	if game.editing().MassSelected(1) {
		t.Fatal("selection-add drag release should not replace selection")
	}
}

func TestAppUnitSmallPointerHelpers(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 140, Y: 100}, Mass: 1},
	)

	game.run.paused = false
	game.run.simulationSpeed = 1
	game.pointer.lastCursor = sim.Vec2{X: 110, Y: 110}
	game.pointer.draggingMassID = 1
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {X: 1, Y: 2}}
	game.advanceSimulationFrame()
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 110, Y: 110}) {
		t.Fatalf("pinned mass after frame = %#v", mass.Position)
	}

	game.handleRightPointer(true, 110, 110)
	if !game.pointer.rightMousePressed || !game.overlays.massMenu.Open {
		t.Fatalf("right pointer state pressed=%t menu=%#v", game.pointer.rightMousePressed, game.overlays.massMenu)
	}
	game.pointer.selectionDrag = true
	game.handleRightPointer(true, 120, 120)
	if !game.pointer.selectionDrag {
		t.Fatal("held right pointer cancelled placement gesture again")
	}
	game.handleRightPointer(false, 100, 900)
	if game.pointer.rightMousePressed {
		t.Fatal("right pointer release stayed pressed")
	}

	rightClick := NewGame()
	rightClick.ReplaceWorld(sim.NewWorld())
	rightClick.pointer.selectionDrag = true
	rightClick.pointer.selectionStart = sim.Vec2{X: 300, Y: 300}
	rightClick.pointer.selectionEnd = sim.Vec2{X: 300, Y: 300}
	rightClick.handleRightPointer(true, 300, 300)
	rightClick.releasePointer(sim.Vec2{X: 300, Y: 300})
	if len(rightClick.World().Masses) != 0 || rightClick.pointer.selectionDrag {
		t.Fatalf("right click placed mass or left selection active: masses=%#v selectionDrag=%t", rightClick.World().Masses, rightClick.pointer.selectionDrag)
	}

	game.pointer.pendingSpringID = 0
	game.beginSpringAt(sim.Vec2{X: 110, Y: 110})
	if game.pointer.pendingSpringID != 1 || game.pointer.pendingSpringEnd != (sim.Vec2{X: 110, Y: 110}) {
		t.Fatalf("pending spring state id=%d end=%#v", game.pointer.pendingSpringID, game.pointer.pendingSpringEnd)
	}

	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {}, 2: {}}
	game.throwSelectedDraggingMasses(sim.Vec2{X: 7, Y: 9})
	first, _ := game.World().MassByID(1)
	second, _ := game.World().MassByID(2)
	if first.Velocity != (sim.Vec2{X: 7, Y: 9}) || second.Velocity != (sim.Vec2{X: 7, Y: 9}) {
		t.Fatalf("selected throw velocities = %#v %#v", first.Velocity, second.Velocity)
	}

	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = false
	if ids := game.selectedMassIDs(); len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("selected mass ids = %#v", ids)
	}

	game.moveSelectedMasses(sim.Vec2{X: 5, Y: -5})
	first, _ = game.World().MassByID(1)
	second, _ = game.World().MassByID(2)
	if first.Position != (sim.Vec2{X: 120, Y: 110}) || second.Position != (sim.Vec2{X: 140, Y: 100}) {
		t.Fatalf("moved selected positions = %#v %#v", first.Position, second.Position)
	}
}

func TestAppUnitReleasePointerClearsGestureAndRepeatState(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1})
	game.pointer.draggingMassID = 1
	game.pointer.dragMoved = true
	game.pointer.selectionDrag = true
	game.controls.activeSlider = "speed slider"
	game.controls.activeNumericStep = "mass increment"
	game.controls.numericStepTicks = 7
	game.controls.activeValueStep = 0.1
	game.controls.valueStepTicks = 9

	game.releasePointer(sim.Vec2{X: 120, Y: 120})

	if game.pointer.draggingMassID != 0 || game.pointer.dragMoved || game.pointer.selectionDrag {
		t.Fatalf("pointer state not cleared: %#v", game.pointer)
	}
	if game.controls.activeSlider != "" || game.controls.activeNumericStep != "" || game.controls.numericStepTicks != 0 || game.controls.activeValueStep != 0 || game.controls.valueStepTicks != 0 {
		t.Fatalf("control repeat state not cleared: %#v", game.controls)
	}
}

func TestAppUnitContinueControlPressRepeatsValueDialogStep(t *testing.T) {
	game := NewGame()
	game.overlays.value.Open = true
	game.overlays.value.Text = "1"
	game.overlays.value.Min = 0
	game.overlays.value.Max = 10
	game.controls.activeValueStep = 0.1
	game.controls.valueStepTicks = numericStepHoldDelayTicks - 1

	game.continuePointerPress(sim.Vec2{}, 0)

	if game.overlays.value.Text != "1.1" || game.controls.valueStepTicks != numericStepHoldDelayTicks {
		t.Fatalf("value dialog repeat text=%q ticks=%d", game.overlays.value.Text, game.controls.valueStepTicks)
	}
}
