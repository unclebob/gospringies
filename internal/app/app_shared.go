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
	wallSpringColor = color.RGBA{R: 235, G: 176, B: 78, A: 255}
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
	default:
		g.continueControlPress(x)
	}
}

func (g *Game) continueControlPress(x int) {
	switch {
	case g.activeNumericStep != "":
		g.continueNumericStepHold()
	case g.activeValueStep != 0:
		g.continueValueDialogStepHold()
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
	g.activeValueStep = 0
	g.valueStepTicks = 0
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
		{open: func() bool { return g.saveDialog.Open }, click: g.clickSaveFilenameDialog},
		{open: func() bool { return g.valueDialog.Open }, click: g.clickValueDialog},
		{open: func() bool { return g.massMenu.Open }, click: g.clickMassContextMenu},
		{open: func() bool { return g.springMenu.Open }, click: g.clickSpringContextMenu},
		{open: func() bool { return g.demoPickerOpen }, click: g.clickDemoPicker},
	}
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
	return g.canvasCoordinate(position)
}

func (g *Game) worldToScreen(position sim.Vec2) sim.Vec2 {
	return g.canvasCoordinate(position)
}

func (g *Game) canvasCoordinate(position sim.Vec2) sim.Vec2 {
	if !g.canvasYUp {
		return position
	}
	return g.flipCanvasY(position)
}

func (g *Game) flipCanvasY(position sim.Vec2) sim.Vec2 {
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T07:36:37-05:00","module_hash":"e19af041141970a998881b85f3a007d1525ff180c18463e327a0f21ac98ec8f7","functions":[{"id":"func/NewGame","name":"NewGame","line":98,"end_line":104,"hash":"687ee2a48042134921fd4acc355bebfefb1e005457b829c027c35e06a9ff57aa"},{"id":"func/DefaultWindowConfig","name":"DefaultWindowConfig","line":106,"end_line":108,"hash":"d07a778f3e78f2a72f6ae58d2c7afa2ecd83053f575705b5fa26512178690d50"},{"id":"func/Game.advanceSimulationFrame","name":"Game.advanceSimulationFrame","line":110,"end_line":115,"hash":"0ce9cc0ed29e1bce0f1027b9d69abb5fdc6152fa91fc19d386447c8983f616a0"},{"id":"func/Game.handleWindowClose","name":"Game.handleWindowClose","line":117,"end_line":121,"hash":"1fa7c9632d93845006cbfb2e5987fe93c53cc4060e482fd661e58389d2aa0d74"},{"id":"func/Game.handleRightPointer","name":"Game.handleRightPointer","line":123,"end_line":128,"hash":"87077ecef7b95720a1a172ce0f030917d018efc34e3e23c46eba23b4344496d5"},{"id":"func/Game.handlePointer","name":"Game.handlePointer","line":130,"end_line":138,"hash":"fe19fadb0f52cc482d787ef61edbeff3cb591e665491cb2e9724e957a53efc50"},{"id":"func/Game.handlePressedPointer","name":"Game.handlePressedPointer","line":140,"end_line":146,"hash":"db03eeacbdc454bbb95e8b15369100910a380ddda7d13ad4e2a004356af49743"},{"id":"func/Game.continuePointerPress","name":"Game.continuePointerPress","line":148,"end_line":159,"hash":"1810a0769a26aa3733c56ad0cec02d58b740e226c4e9a47774b0487ebd92bc4e"},{"id":"func/Game.continueControlPress","name":"Game.continueControlPress","line":161,"end_line":170,"hash":"e00eeab2e059a3ba1975288f1a793a9d9482f05a358599981112791f93a5778b"},{"id":"func/Game.releasePointer","name":"Game.releasePointer","line":172,"end_line":183,"hash":"e419a510badf84445c5743baa04c01dc20521aa8307bf71e2e6e3fac0608777e"},{"id":"func/Game.beginPointerPress","name":"Game.beginPointerPress","line":185,"end_line":200,"hash":"a07cce7514560ae15a043c5789a73c41b4762fecb1347e973f1d30542bed7385"},{"id":"func/Game.clickOpenOverlay","name":"Game.clickOpenOverlay","line":202,"end_line":209,"hash":"bd5b6c770d43d9d0d4843d0e7ffa1ced573a125deee18503c9fe6f9fe22ff1cf"},{"id":"func/overlayClick.run","name":"overlayClick.run","line":216,"end_line":222,"hash":"6cdb80fabb3a6fcaa4d3e7d33c230899d928b552231c4137272d329e27f0ad33"},{"id":"func/Game.openOverlayClicks","name":"Game.openOverlayClicks","line":224,"end_line":232,"hash":"34435d2840a82b532cf2d9eb29e2254563d257c7475f590ead67da8535c567e9"},{"id":"func/Game.controlPointerPress","name":"Game.controlPointerPress","line":234,"end_line":238,"hash":"ae34ac7cc5ca9ac4547264759269c59939550a6d2fa1bd717e049dede1faf9de"},{"id":"func/Game.scrollDemoPicker","name":"Game.scrollDemoPicker","line":240,"end_line":246,"hash":"4ce9154527c8e87cb048f53ab5497a40d1008b9a2b8c99914a702416aa14dccf"},{"id":"func/Game.demoList","name":"Game.demoList","line":248,"end_line":254,"hash":"e2fdf7cbb2fc4d9665e67d52fe19360620ee910129a4444ab436da06652e7fb9"},{"id":"func/Game.beginCanvasGesture","name":"Game.beginCanvasGesture","line":256,"end_line":258,"hash":"197ca42d03064c4be68a351f807238cf8abd0777b55b3cf6e830680482c8003d"},{"id":"func/Game.finishWorldPointer","name":"Game.finishWorldPointer","line":260,"end_line":270,"hash":"ca6c3aaf8a7d68e8628d78864033031a175c2bd225c7df711a05db9d8130d16a"},{"id":"func/Game.finishMassDrag","name":"Game.finishMassDrag","line":272,"end_line":284,"hash":"f2e00273da1f0b377ecd693516c34b59924897c79c5de9de569a8bbc980db20f"},{"id":"func/Game.throwDraggedMasses","name":"Game.throwDraggedMasses","line":286,"end_line":293,"hash":"8c693251aceac50f8d148444b4a3608a340685a1e10faf501576870fae456fcd"},{"id":"func/Game.throwSelectedDraggingMasses","name":"Game.throwSelectedDraggingMasses","line":295,"end_line":301,"hash":"b58525c194991999c567664e4acf44b52d99ddb1635418e1f9995340e6411e32"},{"id":"func/Game.throwSingleDraggingMass","name":"Game.throwSingleDraggingMass","line":303,"end_line":312,"hash":"d7a6b79eeadb7d27602ff2e0b02de42aaccaa148953b5072912ac2efe99ed143"},{"id":"func/Game.createMassAt","name":"Game.createMassAt","line":314,"end_line":334,"hash":"0db0bbba5b2b4523d82ffcdc4af3e636361ded07afb26181a7181e033ee65aa0"},{"id":"func/Game.selectNearest","name":"Game.selectNearest","line":336,"end_line":340,"hash":"c64e5503f6398dce55003b362d508180e35e2b0c3d6ab50d0be690c49a26789b"},{"id":"func/Game.beginSelectGesture","name":"Game.beginSelectGesture","line":342,"end_line":363,"hash":"17786f92e1a8ce760a79bcbdb7e2409b9030bf3b35a6f7d9fa88b68c0ae4b0f8"},{"id":"func/Game.finishSelectGesture","name":"Game.finishSelectGesture","line":365,"end_line":378,"hash":"9b835315af20eddf845eb906462642fcf4454aed35909cefbb2a2405ec03d8cb"},{"id":"func/selectionClick","name":"selectionClick","line":380,"end_line":382,"hash":"f62199bd3dc9a322b7d12c1c28e74fcc9cb030a3565813897710ba8c44a0a3c9"},{"id":"func/Game.beginSpringAt","name":"Game.beginSpringAt","line":384,"end_line":393,"hash":"a92e7604d3f7702c7ea7168a0aa049863938237129fb9eb5b8cf6aa2af766bb7"},{"id":"func/Game.finishSpringAt","name":"Game.finishSpringAt","line":395,"end_line":405,"hash":"da805469b7b70d5f8e21b4c6c581a201a891832dee0ef0eab445691b541c1dc6"},{"id":"func/Game.beginMassDrag","name":"Game.beginMassDrag","line":407,"end_line":418,"hash":"b358e1665e4648ca7ea0b1c1a5e700f2749a723cfc0bd08bb79b9cc9af6574a9"},{"id":"func/Game.captureDraggingOffsets","name":"Game.captureDraggingOffsets","line":420,"end_line":428,"hash":"15f5291a75104a9d5cc437c5461ad4d7cec8ee3f366120613d6dc4cf29c81bc4"},{"id":"func/Game.captureSelectedDraggingOffsets","name":"Game.captureSelectedDraggingOffsets","line":430,"end_line":440,"hash":"a4d6564b62d2b3b4b517dd14511c53cd7cc3398ca91573180a46b61c356fd134"},{"id":"func/Game.pinDraggingMasses","name":"Game.pinDraggingMasses","line":442,"end_line":447,"hash":"5470b2d59632dd7656bf5e6e380a3836631891d8008696c89fadcbe55623775d"},{"id":"func/Game.massAt","name":"Game.massAt","line":449,"end_line":457,"hash":"473e6fecdf13bba51406f9b8e6111187d2cc4033dd3924d4490a27a32a94a25c"},{"id":"func/massDrawCircle","name":"massDrawCircle","line":459,"end_line":461,"hash":"141b72a0cb028af6c4bf4c0bf2d03b72b5298f6cfa2c5e0a19537d66c7de9cdb"},{"id":"func/Game.screenToWorld","name":"Game.screenToWorld","line":463,"end_line":465,"hash":"20d063171fc096c5763e0a5add37194d9b46437547eea38be243c0595489ca4e"},{"id":"func/Game.worldToScreen","name":"Game.worldToScreen","line":467,"end_line":469,"hash":"e02c541b189ecfde32bc8b1576f8e472ce03671dd138c3445907b27b19eb2b71"},{"id":"func/Game.canvasCoordinate","name":"Game.canvasCoordinate","line":471,"end_line":476,"hash":"207330f117cdea953807380885f4a5d60d9e7d766a52f67c9280dfb03e61c0e1"},{"id":"func/Game.flipCanvasY","name":"Game.flipCanvasY","line":478,"end_line":480,"hash":"723a020c9fdefca6b79ca834d8ed5f970e16e58f21ff6c68348ec069045a098c"},{"id":"func/Game.editing","name":"Game.editing","line":482,"end_line":487,"hash":"d8b1ccdd85577daec4370de59fa0271f4eade8301137406cf879ffdc86f5d2e5"},{"id":"func/Game.Layout","name":"Game.Layout","line":489,"end_line":491,"hash":"aab68cdc4f078367f499500c8a90603c494f128359287735d76907fe8d472cf0"},{"id":"func/Game.World","name":"Game.World","line":493,"end_line":495,"hash":"a9270d11e4300269b96797a052fe66c1e96ff42aac546663ea66d24b7f48c6ab"},{"id":"func/Game.SetPaused","name":"Game.SetPaused","line":497,"end_line":499,"hash":"022f698a4f568a12c8b3a2fe5723316d1c01b16b9e1f722d69c28b666b683e27"},{"id":"func/Game.Paused","name":"Game.Paused","line":501,"end_line":503,"hash":"423f86d8b651a7bdfc4f3430ac2d8e0369da3177056a0b8d6d2ee6c26b45aab9"},{"id":"func/Game.InputActive","name":"Game.InputActive","line":505,"end_line":507,"hash":"8ba126aecb56afb862b985e28fdf1144ebb879230d2e659cca1c16327817aea3"},{"id":"func/Game.RenderingActive","name":"Game.RenderingActive","line":509,"end_line":511,"hash":"1f935194ba70ba6592f1e82561e383760fc4b5907912de3ec32c3611503570f8"},{"id":"func/Game.RenderFrame","name":"Game.RenderFrame","line":513,"end_line":515,"hash":"93435d752bdaeafc1e6f52105b61156503e4a1acdc8f0f7ad4ddbafb0dfef512"},{"id":"func/Game.Close","name":"Game.Close","line":517,"end_line":520,"hash":"05d19586766eb286b92af2eead38992b7cfcdcdc358efc607ef6a2f1504cb980"},{"id":"func/Game.Closed","name":"Game.Closed","line":522,"end_line":524,"hash":"cd43811f993dc6eb808f5aeec63512934868aa0e2fa85aa73710dc4aef48861f"}]}
// mutate4go-manifest-end
