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
	simulation      *sim.Simulation
	initialState    *sim.Simulation
	savedState      *sim.Simulation
	selected        bool
	dirty           bool
	lastCommand     string
	paused          bool
	pointer         pointerState
	keyboard        keyboardState
	controls        controlState
	overlays        overlayState
	document        documentState
	editClipboard   editClipboard
	simulationSpeed float64
	canvasYUp       bool
	editor          *edit.Editor
	inputActive     bool
	renderingActive bool
	closed          bool
}

type pointerState struct {
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
	lastCursor        sim.Vec2
}

type keyboardState struct {
	shiftDown   bool
	controlDown bool
	throwDown   bool
}

type controlState struct {
	activeSlider      string
	activeNumericStep string
	numericStepTicks  int
	activeValueStep   float64
	valueStepTicks    int
	focusedNumeric    string
	numericInputText  string
	numericInputTicks int
	numericInputFresh bool
	editMenuOpen      bool
	demoPickerOpen    bool
	demoPickerScroll  int
	demoFiles         []string
}

type overlayState struct {
	massMenu   massContextMenu
	springMenu springContextMenu
	value      valueDialog
	save       saveFilenameDialog
}

type documentState struct {
	pathEntryCommand string
	currentFilePath  string
	lastFileError    string
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
		g.pinDraggingMasses(g.pointer.lastCursor)
	}
}

func (g *Game) handleWindowClose(closing bool) {
	if closing {
		_ = g.Close()
	}
}

func (g *Game) handleRightPointer(pressed bool, x int, y int) {
	if pressed && !g.pointer.rightMousePressed {
		g.openContextAt(x, y)
	}
	g.pointer.rightMousePressed = pressed
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

func (g *Game) scrollDemoPicker(delta int) {
	if !g.controls.demoPickerOpen {
		return
	}
	maxScroll := len(g.demoList()) - demoPickerVisibleRows()
	g.controls.demoPickerScroll = clampInt(g.controls.demoPickerScroll+delta, 0, maxScroll)
}

func (g *Game) demoList() []string {
	if g.controls.demoFiles != nil {
		return g.controls.demoFiles
	}
	g.controls.demoFiles = g.buildDemoList()
	return g.controls.demoFiles
}

func (g *Game) beginCanvasGesture(position sim.Vec2) {
	g.beginSelectGesture(position)
}

func (g *Game) finishWorldPointer(position sim.Vec2) {
	if g.pointer.draggingMassID != 0 {
		g.finishMassDrag(position)
	}
	if g.pointer.pendingSpringID != 0 && !g.pointer.springChainActive {
		g.finishSpringAt(position)
	}
	if g.pointer.selectionDrag {
		g.finishSelectGesture(position)
	}
}

func (g *Game) finishMassDrag(position sim.Vec2) {
	if g.pointer.dragMoved && g.throwKeyPressed() {
		g.throwDraggedMasses(position.Sub(g.pointer.draggingStart))
		return
	}
	if g.pointer.dragMoved || g.pointer.selectionAdd {
		return
	}
	if selectionClick(g.pointer.draggingStart, position) {
		_ = g.editing().SelectMass(g.pointer.draggingMassID)
		g.syncSelectionState()
	}
}

func (g *Game) throwDraggedMasses(velocity sim.Vec2) {
	if len(g.pointer.draggingOffsets) > 0 {
		g.throwSelectedDraggingMasses(velocity)
		g.dirty = true
		return
	}
	g.throwSingleDraggingMass(velocity)
}

func (g *Game) throwSelectedDraggingMasses(velocity sim.Vec2) {
	for i := range g.simulation.Masses {
		if _, ok := g.pointer.draggingOffsets[g.simulation.Masses[i].ID]; ok {
			g.simulation.Masses[i].Velocity = velocity
		}
	}
}

func (g *Game) throwSingleDraggingMass(velocity sim.Vec2) {
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID != g.pointer.draggingMassID {
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
	g.pointer.selectionAdd = g.shiftKeyPressed()
	id, ok := g.massAt(position)
	if ok {
		alreadySelected := g.editing().MassSelected(id)
		if g.pointer.selectionAdd {
			_ = g.editing().AddMassSelection(id)
		} else if !alreadySelected {
			_ = g.editing().SelectMass(id)
		}
		g.syncSelectionState()
		g.pointer.draggingMassID = id
		g.pointer.draggingStart = position
		g.pointer.draggingLast = position
		g.captureDraggingOffsets(position)
		g.pointer.dragMoved = false
		return
	}
	g.pointer.selectionDrag = true
	g.pointer.selectionStart = position
	g.pointer.selectionEnd = position
}

func (g *Game) finishSelectGesture(position sim.Vec2) {
	start := g.pointer.selectionStart
	g.pointer.selectionEnd = position
	g.pointer.selectionDrag = false
	if selectionClick(start, position) {
		if !g.pointer.selectionAdd {
			g.clearSelection()
		}
		g.createMassAt(position, g.pointer.selectionAdd)
		return
	}
	g.editing().BoxSelect(start, position, g.pointer.selectionAdd)
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
		g.pointer.pendingSpringID = id
		g.pointer.pendingSpringEnd = position
	}
}

func (g *Game) finishSpringAt(position sim.Vec2) {
	defer g.clearPendingSpring()
	if !g.positionInCanvas(position) {
		return
	}
	endID, ok := g.massAt(position)
	if !ok || endID == g.pointer.pendingSpringID {
		return
	}
	g.createSpringBetween(g.pointer.pendingSpringID, endID)
}

func (g *Game) beginMassDrag(position sim.Vec2) {
	id, ok := g.massAt(position)
	if !ok {
		return
	}
	g.pointer.draggingMassID = id
	g.pointer.draggingStart = position
	g.pointer.draggingLast = position
	g.captureDraggingOffsets(position)
	g.pointer.dragMoved = false
	g.DragMass(id, position)
}

func (g *Game) captureDraggingOffsets(cursor sim.Vec2) {
	g.pointer.draggingOffsets = map[int]sim.Vec2{}
	if g.captureSelectedDraggingOffsets(cursor) {
		return
	}
	if mass, ok := g.simulation.MassByID(g.pointer.draggingMassID); ok {
		g.pointer.draggingOffsets[g.pointer.draggingMassID] = mass.Position.Sub(cursor)
	}
}

func (g *Game) captureSelectedDraggingOffsets(cursor sim.Vec2) bool {
	if len(g.editing().SelectedMasses) == 0 || !g.editing().MassSelected(g.pointer.draggingMassID) {
		return false
	}
	for _, mass := range g.simulation.Masses {
		if g.editing().MassSelected(mass.ID) {
			g.pointer.draggingOffsets[mass.ID] = mass.Position.Sub(cursor)
		}
	}
	return true
}

func (g *Game) pinDraggingMasses(cursor sim.Vec2) {
	if g.pointer.draggingMassID == 0 || len(g.pointer.draggingOffsets) == 0 {
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
// {"version":1,"tested_at":"2026-05-22T08:31:42-05:00","module_hash":"a73cb2af4862a579e80722724d4cab948dbe344340f47518ee3d43383dd1e4e8","functions":[{"id":"func/NewGame","name":"NewGame","line":118,"end_line":124,"hash":"687ee2a48042134921fd4acc355bebfefb1e005457b829c027c35e06a9ff57aa"},{"id":"func/DefaultWindowConfig","name":"DefaultWindowConfig","line":126,"end_line":128,"hash":"d07a778f3e78f2a72f6ae58d2c7afa2ecd83053f575705b5fa26512178690d50"},{"id":"func/Game.advanceSimulationFrame","name":"Game.advanceSimulationFrame","line":130,"end_line":135,"hash":"b69c280c98ef22e69395f6b86d0e51785b4275e2f1373f2079972549573bf9ae"},{"id":"func/Game.handleWindowClose","name":"Game.handleWindowClose","line":137,"end_line":141,"hash":"1fa7c9632d93845006cbfb2e5987fe93c53cc4060e482fd661e58389d2aa0d74"},{"id":"func/Game.handleRightPointer","name":"Game.handleRightPointer","line":143,"end_line":148,"hash":"a3fc0947248f134d340cbbf8800f8295391fce2d62b35e92cf1a9219bb47d06a"},{"id":"func/Game.handlePointer","name":"Game.handlePointer","line":150,"end_line":158,"hash":"9ed12bc968325df1cad2f12065c9c692da4abe248edaee27b9839fb4320826a1"},{"id":"func/Game.handlePressedPointer","name":"Game.handlePressedPointer","line":160,"end_line":166,"hash":"4342ecf36fca8e3c855125afefe412b7cce4b4a49a7fb9cec9c4214963274472"},{"id":"func/Game.continuePointerPress","name":"Game.continuePointerPress","line":168,"end_line":179,"hash":"9be7038b76b1a74ed55fd60b060a01d7981c546c2c8e62fc9b61a72be0690f60"},{"id":"func/Game.continueControlPress","name":"Game.continueControlPress","line":181,"end_line":190,"hash":"5c5b48e13323a0242b43e2470de8d6e9cbf286cbec5d892d9d1d8b972db1600c"},{"id":"func/Game.releasePointer","name":"Game.releasePointer","line":192,"end_line":203,"hash":"34c0bcde2f0f1921cb3150544cff0f8fa92d90ed3842a396777e2b844e4bde65"},{"id":"func/Game.beginPointerPress","name":"Game.beginPointerPress","line":205,"end_line":220,"hash":"fcc8ec5c4ad7b661c18415865b842634a927dd95dd233fed37bcd7851ff1a02b"},{"id":"func/Game.clickOpenOverlay","name":"Game.clickOpenOverlay","line":222,"end_line":229,"hash":"bd5b6c770d43d9d0d4843d0e7ffa1ced573a125deee18503c9fe6f9fe22ff1cf"},{"id":"func/overlayClick.run","name":"overlayClick.run","line":236,"end_line":242,"hash":"6cdb80fabb3a6fcaa4d3e7d33c230899d928b552231c4137272d329e27f0ad33"},{"id":"func/Game.openOverlayClicks","name":"Game.openOverlayClicks","line":244,"end_line":252,"hash":"fae322207183fa23f0fd9a10db90682d5e2b677cfd2be1837e92912da6c7d178"},{"id":"func/Game.controlPointerPress","name":"Game.controlPointerPress","line":254,"end_line":258,"hash":"ae34ac7cc5ca9ac4547264759269c59939550a6d2fa1bd717e049dede1faf9de"},{"id":"func/Game.scrollDemoPicker","name":"Game.scrollDemoPicker","line":260,"end_line":266,"hash":"c78d6e2d96e418e6daff83f002e782966140a5c37e40377a73af73d140b532a8"},{"id":"func/Game.demoList","name":"Game.demoList","line":268,"end_line":274,"hash":"a2b251e6bb5464680774c9189a40f4ccce99406502b8341fa65e4f80891a1a46"},{"id":"func/Game.beginCanvasGesture","name":"Game.beginCanvasGesture","line":276,"end_line":278,"hash":"197ca42d03064c4be68a351f807238cf8abd0777b55b3cf6e830680482c8003d"},{"id":"func/Game.finishWorldPointer","name":"Game.finishWorldPointer","line":280,"end_line":290,"hash":"dd9b5731c3eac78cb42665e05cbf06fd46650e79abed992c1aec7f00e58acc23"},{"id":"func/Game.finishMassDrag","name":"Game.finishMassDrag","line":292,"end_line":304,"hash":"d122c857bfcadc2fef1d0d79913252932f42f186822a863005e14dd7567c14d4"},{"id":"func/Game.throwDraggedMasses","name":"Game.throwDraggedMasses","line":306,"end_line":313,"hash":"2ba208771b3f0625bb4d6a438eaefc5475ff9b12128b29c7fda64fb4e3db8809"},{"id":"func/Game.throwSelectedDraggingMasses","name":"Game.throwSelectedDraggingMasses","line":315,"end_line":321,"hash":"a632b2e5ea2825212dec79df2f2c06b32a16d64d0127fac221c79c68b785995f"},{"id":"func/Game.throwSingleDraggingMass","name":"Game.throwSingleDraggingMass","line":323,"end_line":332,"hash":"2c420662bfed4ec391c367a3e57bf25ae4590f038378c4aa63c4dd14a99d649b"},{"id":"func/Game.createMassAt","name":"Game.createMassAt","line":334,"end_line":354,"hash":"0db0bbba5b2b4523d82ffcdc4af3e636361ded07afb26181a7181e033ee65aa0"},{"id":"func/Game.selectNearest","name":"Game.selectNearest","line":356,"end_line":360,"hash":"c64e5503f6398dce55003b362d508180e35e2b0c3d6ab50d0be690c49a26789b"},{"id":"func/Game.beginSelectGesture","name":"Game.beginSelectGesture","line":362,"end_line":383,"hash":"45ad3e7fea3d28bb752bb27b89d1343c4d4ee1c22f5419512e8134c6036e3bc2"},{"id":"func/Game.finishSelectGesture","name":"Game.finishSelectGesture","line":385,"end_line":398,"hash":"0c40acff2337b4f2ff3d6b02ce099e224795749865190bc4214804202b73301c"},{"id":"func/selectionClick","name":"selectionClick","line":400,"end_line":402,"hash":"f62199bd3dc9a322b7d12c1c28e74fcc9cb030a3565813897710ba8c44a0a3c9"},{"id":"func/Game.beginSpringAt","name":"Game.beginSpringAt","line":404,"end_line":413,"hash":"7a9047db3306e9d1de75fa42aef1d88a8a65ce5bf48f7915fc518e5b1ca0b50a"},{"id":"func/Game.finishSpringAt","name":"Game.finishSpringAt","line":415,"end_line":425,"hash":"f8b5091082d2eee0006b96ade69cc1d29538aff4689bd3b667676bd57390fca7"},{"id":"func/Game.beginMassDrag","name":"Game.beginMassDrag","line":427,"end_line":438,"hash":"64fedf1ba7b7bd1564ba7c4890821810075631d2d3ed99eb4f6326aa95d282e0"},{"id":"func/Game.captureDraggingOffsets","name":"Game.captureDraggingOffsets","line":440,"end_line":448,"hash":"01ca8dc66c60827d8c51e7480cc31abb9f4a6d012533892fcddc87e18738421e"},{"id":"func/Game.captureSelectedDraggingOffsets","name":"Game.captureSelectedDraggingOffsets","line":450,"end_line":460,"hash":"00960a8a793c7695bd4be5da7838ed652c751435d7b122548e7b1903adaf1e8d"},{"id":"func/Game.pinDraggingMasses","name":"Game.pinDraggingMasses","line":462,"end_line":467,"hash":"e77e762693460c32f2edf1c072c72665017b8726b65480cc97f3c3c863b4e27c"},{"id":"func/Game.massAt","name":"Game.massAt","line":469,"end_line":477,"hash":"473e6fecdf13bba51406f9b8e6111187d2cc4033dd3924d4490a27a32a94a25c"},{"id":"func/massDrawCircle","name":"massDrawCircle","line":479,"end_line":481,"hash":"141b72a0cb028af6c4bf4c0bf2d03b72b5298f6cfa2c5e0a19537d66c7de9cdb"},{"id":"func/Game.screenToWorld","name":"Game.screenToWorld","line":483,"end_line":485,"hash":"20d063171fc096c5763e0a5add37194d9b46437547eea38be243c0595489ca4e"},{"id":"func/Game.worldToScreen","name":"Game.worldToScreen","line":487,"end_line":489,"hash":"e02c541b189ecfde32bc8b1576f8e472ce03671dd138c3445907b27b19eb2b71"},{"id":"func/Game.canvasCoordinate","name":"Game.canvasCoordinate","line":491,"end_line":496,"hash":"207330f117cdea953807380885f4a5d60d9e7d766a52f67c9280dfb03e61c0e1"},{"id":"func/Game.flipCanvasY","name":"Game.flipCanvasY","line":498,"end_line":500,"hash":"723a020c9fdefca6b79ca834d8ed5f970e16e58f21ff6c68348ec069045a098c"},{"id":"func/Game.editing","name":"Game.editing","line":502,"end_line":507,"hash":"d8b1ccdd85577daec4370de59fa0271f4eade8301137406cf879ffdc86f5d2e5"},{"id":"func/Game.Layout","name":"Game.Layout","line":509,"end_line":511,"hash":"aab68cdc4f078367f499500c8a90603c494f128359287735d76907fe8d472cf0"},{"id":"func/Game.World","name":"Game.World","line":513,"end_line":515,"hash":"a9270d11e4300269b96797a052fe66c1e96ff42aac546663ea66d24b7f48c6ab"},{"id":"func/Game.SetPaused","name":"Game.SetPaused","line":517,"end_line":519,"hash":"022f698a4f568a12c8b3a2fe5723316d1c01b16b9e1f722d69c28b666b683e27"},{"id":"func/Game.Paused","name":"Game.Paused","line":521,"end_line":523,"hash":"423f86d8b651a7bdfc4f3430ac2d8e0369da3177056a0b8d6d2ee6c26b45aab9"},{"id":"func/Game.InputActive","name":"Game.InputActive","line":525,"end_line":527,"hash":"8ba126aecb56afb862b985e28fdf1144ebb879230d2e659cca1c16327817aea3"},{"id":"func/Game.RenderingActive","name":"Game.RenderingActive","line":529,"end_line":531,"hash":"1f935194ba70ba6592f1e82561e383760fc4b5907912de3ec32c3611503570f8"},{"id":"func/Game.RenderFrame","name":"Game.RenderFrame","line":533,"end_line":535,"hash":"93435d752bdaeafc1e6f52105b61156503e4a1acdc8f0f7ad4ddbafb0dfef512"},{"id":"func/Game.Close","name":"Game.Close","line":537,"end_line":540,"hash":"05d19586766eb286b92af2eead38992b7cfcdcdc358efc607ef6a2f1504cb980"},{"id":"func/Game.Closed","name":"Game.Closed","line":542,"end_line":544,"hash":"cd43811f993dc6eb808f5aeec63512934868aa0e2fa85aa73710dc4aef48861f"}]}
// mutate4go-manifest-end
