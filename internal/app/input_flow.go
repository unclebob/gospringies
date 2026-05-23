package app

import "springs/internal/sim"

func (g *Game) handleWindowClose(closing bool) {
	if closing {
		_ = g.Close()
	}
}

func (g *Game) handleRightPointer(pressed bool, x int, y int) {
	if pressed && !g.pointer.rightMousePressed {
		g.cancelPlacementGestures()
		g.openContextAt(x, y)
	}
	g.pointer.rightMousePressed = pressed
}

func (g *Game) cancelPlacementGestures() {
	g.pointer.selectionDrag = false
	g.clearPendingSpring()
}

func (g *Game) handlePointer(pressed bool, x int, y int) {
	position := g.screenToWorld(simVec(x, y))
	if pressed {
		g.handlePressedPointer(position, x, y)
	} else {
		g.releasePointer(position)
	}
	g.pointer.mousePressed = pressed
}

func (g *Game) handlePressedPointer(position sim.Vec2, x int, y int) {
	if !g.pointer.mousePressed {
		g.beginPointerPress(position, x, y)
		return
	}
	g.continuePointerPress(position, x)
}

func (g *Game) continuePointerPress(position sim.Vec2, x int) {
	switch {
	case g.pointer.draggingMassID != 0:
		g.DragMass(g.pointer.draggingMassID, position)
	case g.pointer.pendingSpringID != 0:
		g.pointer.pendingSpringEnd = g.clampToCanvas(position)
	case g.pointer.selectionDrag:
		g.pointer.selectionEnd = position
	default:
		g.continueControlPress(x)
	}
}

func (g *Game) continueControlPress(x int) {
	switch {
	case g.controls.activeNumericStep != "":
		g.continueNumericStepHold()
	case g.controls.activeValueStep != 0:
		g.continueValueDialogStepHold()
	case g.controls.activeSlider != "":
		g.setSliderAt(g.controls.activeSlider, x)
	}
}

func (g *Game) releasePointer(position sim.Vec2) {
	g.finishWorldPointer(position)
	g.pointer.draggingMassID = 0
	g.pointer.draggingOffsets = nil
	g.pointer.dragMoved = false
	g.pointer.selectionDrag = false
	g.controls.activeSlider = ""
	g.controls.activeNumericStep = ""
	g.controls.numericStepTicks = 0
	g.controls.activeValueStep = 0
	g.controls.valueStepTicks = 0
}

func (g *Game) beginPointerPress(position sim.Vec2, x int, y int) {
	if g.clickOpenOverlay(x, y) {
		return
	}
	if g.pointer.springChainActive {
		g.continueSpringChainAt(position, g.controlKeyPressed())
		return
	}
	if g.controlKeyPressed() {
		g.controlPointerPress(position, x, y)
		return
	}
	if !g.ClickAt(x, y) {
		g.beginCanvasGesture(position)
	}
}

func (g *Game) clickOpenOverlay(x int, y int) bool {
	for _, click := range g.openOverlayClicks() {
		if click.run(x, y) {
			return true
		}
	}
	return false
}

type overlayClick struct {
	open  func() bool
	click func(int, int)
}

func (click overlayClick) run(x int, y int) bool {
	if !click.open() {
		return false
	}
	click.click(x, y)
	return true
}

func (g *Game) openOverlayClicks() []overlayClick {
	return []overlayClick{
		{open: func() bool { return g.overlays.save.Open }, click: g.clickSaveFilenameDialog},
		{open: func() bool { return g.overlays.value.Open }, click: g.clickValueDialog},
		{open: func() bool { return g.overlays.massMenu.Open }, click: g.clickMassContextMenu},
		{open: func() bool { return g.overlays.springMenu.Open }, click: g.clickSpringContextMenu},
		{open: func() bool { return g.controls.demoPickerOpen }, click: g.clickDemoPicker},
	}
}

func (g *Game) controlPointerPress(position sim.Vec2, x int, y int) {
	if !g.ClickAt(x, y) {
		g.beginControlPlacementAt(position)
	}
}
