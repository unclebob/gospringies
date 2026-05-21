//go:build appunit

package app

import (
	"testing"

	"springs/internal/sim"
)

func TestAppUnitPointerPressBranches(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})

	game.beginPointerPress(sim.Vec2{X: 500, Y: 300}, 500, 300)
	if game.draggingMassID != 1 || !game.editing().MassSelected(1) {
		t.Fatalf("mass press state dragging=%d selected=%#v", game.draggingMassID, game.editing().SelectedMasses)
	}
	game.releasePointer(sim.Vec2{X: 500, Y: 300})

	game.beginPointerPress(sim.Vec2{X: 540, Y: 320}, 540, 320)
	game.releasePointer(sim.Vec2{X: 540, Y: 320})
	if len(game.World().Masses) != 2 || !game.editing().MassSelected(2) {
		t.Fatalf("empty canvas press masses=%#v selected=%#v", game.World().Masses, game.editing().SelectedMasses)
	}

	game.controlDown = true
	game.beginPointerPress(sim.Vec2{X: 580, Y: 320}, 580, 320)
	if game.pendingSpringID == 0 || !game.springChainActive {
		t.Fatalf("control placement pending=%d active=%t", game.pendingSpringID, game.springChainActive)
	}
	game.controlDown = false
	game.beginPointerPress(sim.Vec2{X: 620, Y: 320}, 620, 320)
	if len(game.World().Springs) != 1 || game.pendingSpringID != 0 || game.springChainActive {
		t.Fatalf("chain finish springs=%#v pending=%d active=%t", game.World().Springs, game.pendingSpringID, game.springChainActive)
	}
}

func TestAppUnitContinuePointerPressBranches(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})

	game.draggingMassID = 1
	game.draggingStart = sim.Vec2{X: 500, Y: 300}
	game.draggingLast = sim.Vec2{X: 500, Y: 300}
	game.continuePointerPress(sim.Vec2{X: 520, Y: 310}, 520)
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 520, Y: 310}) || !game.dragMoved {
		t.Fatalf("drag branch mass=%#v moved=%t", mass, game.dragMoved)
	}

	game.draggingMassID = 0
	game.pendingSpringID = 1
	game.continuePointerPress(sim.Vec2{X: 2000, Y: 2000}, 2000)
	if game.pendingSpringEnd != (sim.Vec2{X: 1340, Y: 1000}) {
		t.Fatalf("pending spring end = %#v", game.pendingSpringEnd)
	}

	game.pendingSpringID = 0
	game.selectionDrag = true
	game.continuePointerPress(sim.Vec2{X: 700, Y: 500}, 700)
	if game.selectionEnd != (sim.Vec2{X: 700, Y: 500}) {
		t.Fatalf("selection end = %#v", game.selectionEnd)
	}

	game.selectionDrag = false
	game.activeNumericStep = "mass increment"
	game.numericStepTicks = numericStepHoldDelayTicks - 1
	game.continuePointerPress(sim.Vec2{}, 0)
	if game.numericStepTicks != numericStepHoldDelayTicks {
		t.Fatalf("numeric step ticks = %d", game.numericStepTicks)
	}

	game.activeNumericStep = ""
	game.activeSlider = "speed slider"
	game.continuePointerPress(sim.Vec2{}, sliderTrack(mustVisibleControl(t, "speed slider")).Max.X)
	if game.simulationSpeed != maxSpeed {
		t.Fatalf("simulation speed = %f, want %f", game.simulationSpeed, maxSpeed)
	}
}

func TestAppUnitFinishWorldPointerCompletesGestures(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 540, Y: 300}, Mass: 1},
	)

	game.draggingMassID = 1
	game.draggingStart = sim.Vec2{X: 500, Y: 300}
	game.dragMoved = true
	game.throwDown = true
	game.finishWorldPointer(sim.Vec2{X: 530, Y: 320})
	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 30, Y: 20}) {
		t.Fatalf("throw velocity = %#v", mass.Velocity)
	}

	game.throwDown = false
	game.draggingMassID = 0
	game.pendingSpringID = 1
	game.finishWorldPointer(sim.Vec2{X: 540, Y: 300})
	if len(game.World().Springs) != 1 || game.pendingSpringID != 0 {
		t.Fatalf("spring finish springs=%#v pending=%d", game.World().Springs, game.pendingSpringID)
	}

	game.selectionDrag = true
	game.selectionStart = sim.Vec2{X: 700, Y: 300}
	game.finishWorldPointer(sim.Vec2{X: 700, Y: 300})
	if len(game.World().Masses) != 3 || game.selectionDrag {
		t.Fatalf("selection finish masses=%#v selectionDrag=%t", game.World().Masses, game.selectionDrag)
	}
}

func TestAppUnitFinishMassDragNonThrowBranches(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Velocity: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	game.draggingMassID = 1
	game.draggingStart = sim.Vec2{X: 500, Y: 300}
	game.dragMoved = true

	game.finishMassDrag(sim.Vec2{X: 530, Y: 320})
	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 3, Y: 4}) {
		t.Fatalf("non-throw drag velocity = %#v", mass.Velocity)
	}

	game.dragMoved = false
	game.selectionAdd = true
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

	game.paused = false
	game.simulationSpeed = 1
	game.lastCursor = sim.Vec2{X: 110, Y: 110}
	game.draggingMassID = 1
	game.draggingOffsets = map[int]sim.Vec2{1: {X: 1, Y: 2}}
	game.advanceSimulationFrame()
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 110, Y: 110}) {
		t.Fatalf("pinned mass after frame = %#v", mass.Position)
	}

	game.handleRightPointer(true, 110, 110)
	if !game.rightMousePressed || !game.massMenu.Open {
		t.Fatalf("right pointer state pressed=%t menu=%#v", game.rightMousePressed, game.massMenu)
	}
	game.handleRightPointer(false, 100, 900)
	if game.rightMousePressed {
		t.Fatal("right pointer release stayed pressed")
	}

	game.pendingSpringID = 0
	game.beginSpringAt(sim.Vec2{X: 110, Y: 110})
	if game.pendingSpringID != 1 || game.pendingSpringEnd != (sim.Vec2{X: 110, Y: 110}) {
		t.Fatalf("pending spring state id=%d end=%#v", game.pendingSpringID, game.pendingSpringEnd)
	}

	game.draggingOffsets = map[int]sim.Vec2{1: {}, 2: {}}
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
