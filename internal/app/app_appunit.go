//go:build appunit

package app

import (
	"image/color"
	"math"

	"springs/internal/edit"
	"springs/internal/sim"
)

const (
	screenWidth     = 1700
	screenHeight    = 1000
	toolbarWidth    = 72
	inspectorWidth  = 360
	topBarHeight    = 36
	statusHeight    = 24
	baseFrameTime   = 1.0 / 60.0
	maxSpeed        = 4.0
	springThickness = 2.0
)

var (
	backgroundColor = color.RGBA{R: 18, G: 20, B: 24, A: 255}
	springColor     = color.RGBA{R: 116, G: 190, B: 222, A: 255}
	massColor       = color.RGBA{R: 238, G: 212, B: 96, A: 255}
	fixedMassColor  = color.RGBA{R: 238, G: 116, B: 96, A: 255}
	wallColor       = color.RGBA{R: 180, G: 186, B: 196, A: 255}
	selectionColor  = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	chromeColor     = color.RGBA{R: 48, G: 56, B: 68, A: 255}
	panelColor      = color.RGBA{R: 34, G: 40, B: 50, A: 255}
	gridPointColor  = color.RGBA{R: 78, G: 90, B: 106, A: 255}
)

type Game struct {
	simulation        *sim.Simulation
	initialState      *sim.Simulation
	savedState        *sim.Simulation
	selected          bool
	dirty             bool
	lastCommand       string
	pathEntryCommand  string
	paused            bool
	mousePressed      bool
	rightMousePressed bool
	draggingMassID    int
	draggingStart     sim.Vec2
	draggingLast      sim.Vec2
	draggingOffsets   map[int]sim.Vec2
	dragMoved         bool
	pendingSpringID   int
	pendingSpringEnd  sim.Vec2
	springChainActive bool
	selectionDrag     bool
	selectionStart    sim.Vec2
	selectionEnd      sim.Vec2
	selectionAdd      bool
	shiftDown         bool
	controlDown       bool
	throwDown         bool
	activeSlider      string
	activeNumericStep string
	numericStepTicks  int
	activeValueStep   float64
	valueStepTicks    int
	focusedNumeric    string
	numericInputText  string
	numericInputTicks int
	numericInputFresh bool
	massMenu          massContextMenu
	springMenu        springContextMenu
	valueDialog       valueDialog
	saveDialog        saveFilenameDialog
	editMenuOpen      bool
	editClipboard     editClipboard
	lastCursor        sim.Vec2
	demoPickerOpen    bool
	demoPickerScroll  int
	demoFiles         []string
	currentFilePath   string
	lastFileError     string
	simulationSpeed   float64
	canvasYUp         bool
	editor            *edit.Editor
	inputActive       bool
	renderingActive   bool
	closed            bool
}

type WindowConfig struct {
	Width     int
	Height    int
	Title     string
	Resizable bool
}

func NewGame() *Game {
	world := newDefaultStartupWorld()
	game := &Game{simulation: world, simulationSpeed: 1, canvasYUp: true, editor: edit.NewEditor(world)}
	game.applyCanvasWallBounds(world)
	game.initialState = world.Clone()
	return game
}

func DefaultWindowConfig() WindowConfig {
	return WindowConfig{Width: screenWidth, Height: screenHeight, Title: "Springs", Resizable: true}
}

func Run() error {
	return nil
}

func (g *Game) Update() error {
	g.inputActive = true
	g.tickNumericTextField()
	g.advanceSimulationFrame()
	return nil
}

func (g *Game) advanceSimulationFrame() {
	if !g.paused && g.simulationSpeed > 0 {
		g.simulation.AdvanceDuration(baseFrameTime * g.simulationSpeed)
		g.pinDraggingMasses(g.lastCursor)
	}
}

func (g *Game) handleWindowClose(closing bool) {
	if closing {
		_ = g.Close()
	}
}

func (g *Game) shiftKeyPressed() bool {
	return g.shiftDown
}

func (g *Game) controlKeyPressed() bool {
	return g.controlDown
}

func (g *Game) throwKeyPressed() bool {
	return g.throwDown
}

func (g *Game) handleRightPointer(pressed bool, x int, y int) {
	if pressed && !g.rightMousePressed {
		g.openContextAt(x, y)
	}
	g.rightMousePressed = pressed
}

func (g *Game) handlePointer(pressed bool, x int, y int) {
	position := g.screenToWorld(simVec(x, y))
	if pressed {
		g.handlePressedPointer(position, x, y)
	} else {
		g.releasePointer(position)
	}
	g.mousePressed = pressed
}

func (g *Game) handlePressedPointer(position sim.Vec2, x int, y int) {
	if !g.mousePressed {
		g.beginPointerPress(position, x, y)
		return
	}
	g.continuePointerPress(position, x)
}

func (g *Game) continuePointerPress(position sim.Vec2, x int) {
	switch {
	case g.draggingMassID != 0:
		g.DragMass(g.draggingMassID, position)
	case g.pendingSpringID != 0:
		g.pendingSpringEnd = g.clampToCanvas(position)
	case g.selectionDrag:
		g.selectionEnd = position
	case g.activeNumericStep != "":
		g.continueNumericStepHold()
	case g.activeSlider != "":
		g.setSliderAt(g.activeSlider, x)
	}
}

func (g *Game) releasePointer(position sim.Vec2) {
	g.finishWorldPointer(position)
	g.draggingMassID = 0
	g.draggingOffsets = nil
	g.dragMoved = false
	g.selectionDrag = false
	g.activeSlider = ""
	g.activeNumericStep = ""
	g.numericStepTicks = 0
}

func (g *Game) beginPointerPress(position sim.Vec2, x int, y int) {
	if g.clickOpenOverlay(x, y) {
		return
	}
	if g.springChainActive {
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
	if g.valueDialog.Open {
		g.clickValueDialog(x, y)
		return true
	}
	if g.massMenu.Open {
		g.clickMassContextMenu(x, y)
		return true
	}
	if g.demoPickerOpen {
		g.clickDemoPicker(x, y)
		return true
	}
	return false
}

func (g *Game) controlPointerPress(position sim.Vec2, x int, y int) {
	if !g.ClickAt(x, y) {
		g.beginControlPlacementAt(position)
	}
}

func (g *Game) scrollDemoPicker(delta int) {
	if !g.demoPickerOpen {
		return
	}
	maxScroll := len(g.demoList()) - demoPickerVisibleRows()
	g.demoPickerScroll = clampInt(g.demoPickerScroll+delta, 0, maxScroll)
}

func (g *Game) demoList() []string {
	if g.demoFiles != nil {
		return g.demoFiles
	}
	g.demoFiles = g.buildDemoList()
	return g.demoFiles
}

func (g *Game) beginCanvasGesture(position sim.Vec2) {
	g.beginSelectGesture(position)
}

func (g *Game) finishWorldPointer(position sim.Vec2) {
	if g.draggingMassID != 0 {
		g.finishMassDrag(position)
	}
	if g.pendingSpringID != 0 && !g.springChainActive {
		g.finishSpringAt(position)
	}
	if g.selectionDrag {
		g.finishSelectGesture(position)
	}
}

func (g *Game) finishMassDrag(position sim.Vec2) {
	if g.dragMoved && g.throwKeyPressed() {
		g.throwDraggedMasses(position.Sub(g.draggingStart))
		return
	}
	if g.dragMoved || g.selectionAdd {
		return
	}
	if selectionClick(g.draggingStart, position) {
		_ = g.editing().SelectMass(g.draggingMassID)
		g.syncSelectionState()
	}
}

func (g *Game) throwDraggedMasses(velocity sim.Vec2) {
	if len(g.draggingOffsets) > 0 {
		g.throwSelectedDraggingMasses(velocity)
		g.dirty = true
		return
	}
	g.throwSingleDraggingMass(velocity)
}

func (g *Game) throwSelectedDraggingMasses(velocity sim.Vec2) {
	for i := range g.simulation.Masses {
		if _, ok := g.draggingOffsets[g.simulation.Masses[i].ID]; ok {
			g.simulation.Masses[i].Velocity = velocity
		}
	}
}

func (g *Game) throwSingleDraggingMass(velocity sim.Vec2) {
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID != g.draggingMassID {
			continue
		}
		g.simulation.Masses[i].Velocity = velocity
		g.dirty = true
		return
	}
}

func (g *Game) createMassAt(position sim.Vec2, addToSelection bool) (int, bool) {
	if !g.positionInCanvas(position) {
		return 0, false
	}
	editor := g.editing()
	editor.Mode = edit.ModeAddMass
	editor.GridSnapEnabled = g.gridSnapEnabled()
	editor.GridSnapSize = g.gridSnapSize()
	id, err := editor.Click(g.snapToCanvas(position))
	if err == nil {
		if addToSelection {
			_ = editor.AddMassSelection(id)
		} else {
			_ = editor.SelectMass(id)
		}
		g.syncSelectionState()
		g.dirty = true
		return id, true
	}
	return 0, false
}

func (g *Game) selectNearest(position sim.Vec2) {
	if err := g.editing().SelectNearest(position, false); err == nil {
		g.selected = true
	}
}

func (g *Game) beginSelectGesture(position sim.Vec2) {
	g.selectionAdd = g.shiftKeyPressed()
	id, ok := g.massAt(position)
	if ok {
		alreadySelected := g.editing().MassSelected(id)
		if g.selectionAdd {
			_ = g.editing().AddMassSelection(id)
		} else if !alreadySelected {
			_ = g.editing().SelectMass(id)
		}
		g.syncSelectionState()
		g.draggingMassID = id
		g.draggingStart = position
		g.draggingLast = position
		g.captureDraggingOffsets(position)
		g.dragMoved = false
		return
	}
	g.selectionDrag = true
	g.selectionStart = position
	g.selectionEnd = position
}

func (g *Game) finishSelectGesture(position sim.Vec2) {
	start := g.selectionStart
	g.selectionEnd = position
	g.selectionDrag = false
	if selectionClick(start, position) {
		if !g.selectionAdd {
			g.clearSelection()
		}
		g.createMassAt(position, g.selectionAdd)
		return
	}
	g.editing().BoxSelect(start, position, g.selectionAdd)
	g.syncSelectionState()
}

func selectionClick(start sim.Vec2, end sim.Vec2) bool {
	return math.Hypot(start.X-end.X, start.Y-end.Y) < 3
}

func (g *Game) beginSpringAt(position sim.Vec2) {
	if !g.positionInCanvas(position) {
		return
	}
	id, ok := g.massAt(position)
	if ok {
		g.pendingSpringID = id
		g.pendingSpringEnd = position
	}
}

func (g *Game) finishSpringAt(position sim.Vec2) {
	defer g.clearPendingSpring()
	if !g.positionInCanvas(position) {
		return
	}
	endID, ok := g.massAt(position)
	if !ok || endID == g.pendingSpringID {
		return
	}
	g.createSpringBetween(g.pendingSpringID, endID)
}

func (g *Game) beginMassDrag(position sim.Vec2) {
	id, ok := g.massAt(position)
	if !ok {
		return
	}
	g.draggingMassID = id
	g.draggingStart = position
	g.draggingLast = position
	g.captureDraggingOffsets(position)
	g.dragMoved = false
	g.DragMass(id, position)
}

func (g *Game) captureDraggingOffsets(cursor sim.Vec2) {
	g.draggingOffsets = map[int]sim.Vec2{}
	if g.captureSelectedDraggingOffsets(cursor) {
		return
	}
	if mass, ok := g.simulation.MassByID(g.draggingMassID); ok {
		g.draggingOffsets[g.draggingMassID] = mass.Position.Sub(cursor)
	}
}

func (g *Game) captureSelectedDraggingOffsets(cursor sim.Vec2) bool {
	if len(g.editing().SelectedMasses) == 0 || !g.editing().MassSelected(g.draggingMassID) {
		return false
	}
	for _, mass := range g.simulation.Masses {
		if g.editing().MassSelected(mass.ID) {
			g.draggingOffsets[mass.ID] = mass.Position.Sub(cursor)
		}
	}
	return true
}

func (g *Game) pinDraggingMasses(cursor sim.Vec2) {
	if g.draggingMassID == 0 || len(g.draggingOffsets) == 0 {
		return
	}
	g.applyDraggingOffsets(cursor)
}

func (g *Game) massAt(position sim.Vec2) (int, bool) {
	for _, mass := range g.simulation.Masses {
		_, _, radius := massDrawCircle(mass)
		if math.Hypot(mass.Position.X-position.X, mass.Position.Y-position.Y) <= float64(radius) {
			return mass.ID, true
		}
	}
	return 0, false
}

func massDrawCircle(mass sim.Mass) (float32, float32, float32) {
	return float32(mass.Position.X), float32(mass.Position.Y), float32(sim.MassRadius(mass))
}

func (g *Game) screenToWorld(position sim.Vec2) sim.Vec2 {
	if !g.canvasYUp {
		return position
	}
	return sim.Vec2{X: position.X, Y: g.simulation.Bounds.Height - position.Y}
}

func (g *Game) worldToScreen(position sim.Vec2) sim.Vec2 {
	if !g.canvasYUp {
		return position
	}
	return sim.Vec2{X: position.X, Y: g.simulation.Bounds.Height - position.Y}
}

func (g *Game) editing() *edit.Editor {
	if g.editor == nil || g.editor.World != g.simulation {
		g.editor = edit.NewEditor(g.simulation)
	}
	return g.editor
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) World() *sim.Simulation {
	return g.simulation
}

func (g *Game) SetPaused(paused bool) {
	g.paused = paused
}

func (g *Game) Paused() bool {
	return g.paused
}

func (g *Game) InputActive() bool {
	return g.inputActive
}

func (g *Game) RenderingActive() bool {
	return g.renderingActive
}

func (g *Game) RenderFrame() {
	g.renderingActive = true
}

func (g *Game) Close() error {
	g.closed = true
	return nil
}

func (g *Game) Closed() bool {
	return g.closed
}
