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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-23T11:48:53-05:00","module_hash":"f3bbdffc1419793df7f4a41070d55519a907148561581fe07ccbcee96cafae8f","functions":[{"id":"func/Game.handleWindowClose","name":"Game.handleWindowClose","line":5,"end_line":9,"hash":"1fa7c9632d93845006cbfb2e5987fe93c53cc4060e482fd661e58389d2aa0d74"},{"id":"func/Game.handleRightPointer","name":"Game.handleRightPointer","line":11,"end_line":17,"hash":"ee7b1ed7d2ec67683c2a993e33fc6798e18a0c6ff0bde842a26cf86ef8ba3a97"},{"id":"func/Game.cancelPlacementGestures","name":"Game.cancelPlacementGestures","line":19,"end_line":22,"hash":"de09242e2300de58e3daedb1b16bdaeac15dab10874bab0565c43e9395615e3e"},{"id":"func/Game.handlePointer","name":"Game.handlePointer","line":24,"end_line":32,"hash":"9ed12bc968325df1cad2f12065c9c692da4abe248edaee27b9839fb4320826a1"},{"id":"func/Game.handlePressedPointer","name":"Game.handlePressedPointer","line":34,"end_line":40,"hash":"4342ecf36fca8e3c855125afefe412b7cce4b4a49a7fb9cec9c4214963274472"},{"id":"func/Game.continuePointerPress","name":"Game.continuePointerPress","line":42,"end_line":53,"hash":"9be7038b76b1a74ed55fd60b060a01d7981c546c2c8e62fc9b61a72be0690f60"},{"id":"func/Game.continueControlPress","name":"Game.continueControlPress","line":55,"end_line":64,"hash":"5c5b48e13323a0242b43e2470de8d6e9cbf286cbec5d892d9d1d8b972db1600c"},{"id":"func/Game.releasePointer","name":"Game.releasePointer","line":66,"end_line":77,"hash":"34c0bcde2f0f1921cb3150544cff0f8fa92d90ed3842a396777e2b844e4bde65"},{"id":"func/Game.beginPointerPress","name":"Game.beginPointerPress","line":79,"end_line":94,"hash":"fcc8ec5c4ad7b661c18415865b842634a927dd95dd233fed37bcd7851ff1a02b"},{"id":"func/Game.clickOpenOverlay","name":"Game.clickOpenOverlay","line":96,"end_line":103,"hash":"bd5b6c770d43d9d0d4843d0e7ffa1ced573a125deee18503c9fe6f9fe22ff1cf"},{"id":"func/overlayClick.run","name":"overlayClick.run","line":110,"end_line":116,"hash":"6cdb80fabb3a6fcaa4d3e7d33c230899d928b552231c4137272d329e27f0ad33"},{"id":"func/Game.openOverlayClicks","name":"Game.openOverlayClicks","line":118,"end_line":126,"hash":"fae322207183fa23f0fd9a10db90682d5e2b677cfd2be1837e92912da6c7d178"},{"id":"func/Game.controlPointerPress","name":"Game.controlPointerPress","line":128,"end_line":132,"hash":"ae34ac7cc5ca9ac4547264759269c59939550a6d2fa1bd717e049dede1faf9de"}]}
// mutate4go-manifest-end
