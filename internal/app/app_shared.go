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
	world     worldSession
	editState editState
	run       simulationRunState
	pointer   pointerState
	keyboard  keyboardState
	controls  controlState
	overlays  overlayState
	document  documentState
	runtime   appRuntimeState
}

type worldSession struct {
	simulation   *sim.Simulation
	initialState *sim.Simulation
	savedState   *sim.Simulation
	editor       *edit.Editor
}

type editState struct {
	selected    bool
	dirty       bool
	lastCommand string
	clipboard   editClipboard
}

type simulationRunState struct {
	paused          bool
	simulationSpeed float64
	canvasYUp       bool
}

type appRuntimeState struct {
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
	game := &Game{
		world: worldSession{
			simulation: world,
			editor:     edit.NewEditor(world),
		},
		run: simulationRunState{simulationSpeed: 1, canvasYUp: true},
	}
	game.applyCanvasWallBounds(world)
	game.world.initialState = world.Clone()
	return game
}

func DefaultWindowConfig() WindowConfig {
	return WindowConfig{Width: screenWidth, Height: screenHeight, Title: "Springs", Resizable: true}
}

func (g *Game) advanceSimulationFrame() {
	if !g.run.paused && g.run.simulationSpeed > 0 {
		g.world.simulation.AdvanceDuration(baseFrameTime * g.run.simulationSpeed)
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
		g.editState.dirty = true
		return
	}
	g.throwSingleDraggingMass(velocity)
}

func (g *Game) throwSelectedDraggingMasses(velocity sim.Vec2) {
	for i := range g.world.simulation.Masses {
		if _, ok := g.pointer.draggingOffsets[g.world.simulation.Masses[i].ID]; ok {
			g.world.simulation.Masses[i].Velocity = velocity
		}
	}
}

func (g *Game) throwSingleDraggingMass(velocity sim.Vec2) {
	for i := range g.world.simulation.Masses {
		if g.world.simulation.Masses[i].ID != g.pointer.draggingMassID {
			continue
		}
		g.world.simulation.Masses[i].Velocity = velocity
		g.editState.dirty = true
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
		g.editState.dirty = true
		return id, true
	}
	return 0, false
}

func (g *Game) selectNearest(position sim.Vec2) {
	if err := g.editing().SelectNearest(position, false); err == nil {
		g.editState.selected = true
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
	if mass, ok := g.world.simulation.MassByID(g.pointer.draggingMassID); ok {
		g.pointer.draggingOffsets[g.pointer.draggingMassID] = mass.Position.Sub(cursor)
	}
}

func (g *Game) captureSelectedDraggingOffsets(cursor sim.Vec2) bool {
	if len(g.editing().SelectedMasses) == 0 || !g.editing().MassSelected(g.pointer.draggingMassID) {
		return false
	}
	for _, mass := range g.world.simulation.Masses {
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
	for _, mass := range g.world.simulation.Masses {
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
	if !g.run.canvasYUp {
		return position
	}
	return g.flipCanvasY(position)
}

func (g *Game) flipCanvasY(position sim.Vec2) sim.Vec2 {
	return sim.Vec2{X: position.X, Y: g.world.simulation.Bounds.Height - position.Y}
}

func (g *Game) editing() *edit.Editor {
	if g.world.editor == nil || g.world.editor.World != g.world.simulation {
		g.world.editor = edit.NewEditor(g.world.simulation)
	}
	return g.world.editor
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) World() *sim.Simulation {
	return g.world.simulation
}

func (g *Game) SetPaused(paused bool) {
	g.run.paused = paused
}

func (g *Game) Paused() bool {
	return g.run.paused
}

func (g *Game) InputActive() bool {
	return g.runtime.inputActive
}

func (g *Game) RenderingActive() bool {
	return g.runtime.renderingActive
}

func (g *Game) RenderFrame() {
	g.runtime.renderingActive = true
}

func (g *Game) Close() error {
	g.runtime.closed = true
	return nil
}

func (g *Game) Closed() bool {
	return g.runtime.closed
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"9c821804610f87381db06069a8aa0f9f087a44e0359988a375067f70a2b32a68","functions":[{"id":"func/NewGame","name":"NewGame","line":134,"end_line":146,"hash":"47db37f71ce2ab29ddf870ff90facecdd6dd22775c6cbf35318a8e8f5cec342c"},{"id":"func/DefaultWindowConfig","name":"DefaultWindowConfig","line":148,"end_line":150,"hash":"d07a778f3e78f2a72f6ae58d2c7afa2ecd83053f575705b5fa26512178690d50"},{"id":"func/Game.advanceSimulationFrame","name":"Game.advanceSimulationFrame","line":152,"end_line":157,"hash":"f8c98b9b8bf8859017b73642bbc7ec8a1853ae067f6e5fbddc91864bc8d1d135"},{"id":"func/Game.handleWindowClose","name":"Game.handleWindowClose","line":159,"end_line":163,"hash":"1fa7c9632d93845006cbfb2e5987fe93c53cc4060e482fd661e58389d2aa0d74"},{"id":"func/Game.handleRightPointer","name":"Game.handleRightPointer","line":165,"end_line":170,"hash":"a3fc0947248f134d340cbbf8800f8295391fce2d62b35e92cf1a9219bb47d06a"},{"id":"func/Game.handlePointer","name":"Game.handlePointer","line":172,"end_line":180,"hash":"9ed12bc968325df1cad2f12065c9c692da4abe248edaee27b9839fb4320826a1"},{"id":"func/Game.handlePressedPointer","name":"Game.handlePressedPointer","line":182,"end_line":188,"hash":"4342ecf36fca8e3c855125afefe412b7cce4b4a49a7fb9cec9c4214963274472"},{"id":"func/Game.continuePointerPress","name":"Game.continuePointerPress","line":190,"end_line":201,"hash":"9be7038b76b1a74ed55fd60b060a01d7981c546c2c8e62fc9b61a72be0690f60"},{"id":"func/Game.continueControlPress","name":"Game.continueControlPress","line":203,"end_line":212,"hash":"5c5b48e13323a0242b43e2470de8d6e9cbf286cbec5d892d9d1d8b972db1600c"},{"id":"func/Game.releasePointer","name":"Game.releasePointer","line":214,"end_line":225,"hash":"34c0bcde2f0f1921cb3150544cff0f8fa92d90ed3842a396777e2b844e4bde65"},{"id":"func/Game.beginPointerPress","name":"Game.beginPointerPress","line":227,"end_line":242,"hash":"fcc8ec5c4ad7b661c18415865b842634a927dd95dd233fed37bcd7851ff1a02b"},{"id":"func/Game.clickOpenOverlay","name":"Game.clickOpenOverlay","line":244,"end_line":251,"hash":"bd5b6c770d43d9d0d4843d0e7ffa1ced573a125deee18503c9fe6f9fe22ff1cf"},{"id":"func/overlayClick.run","name":"overlayClick.run","line":258,"end_line":264,"hash":"6cdb80fabb3a6fcaa4d3e7d33c230899d928b552231c4137272d329e27f0ad33"},{"id":"func/Game.openOverlayClicks","name":"Game.openOverlayClicks","line":266,"end_line":274,"hash":"fae322207183fa23f0fd9a10db90682d5e2b677cfd2be1837e92912da6c7d178"},{"id":"func/Game.controlPointerPress","name":"Game.controlPointerPress","line":276,"end_line":280,"hash":"ae34ac7cc5ca9ac4547264759269c59939550a6d2fa1bd717e049dede1faf9de"},{"id":"func/Game.scrollDemoPicker","name":"Game.scrollDemoPicker","line":282,"end_line":288,"hash":"c78d6e2d96e418e6daff83f002e782966140a5c37e40377a73af73d140b532a8"},{"id":"func/Game.demoList","name":"Game.demoList","line":290,"end_line":296,"hash":"a2b251e6bb5464680774c9189a40f4ccce99406502b8341fa65e4f80891a1a46"},{"id":"func/Game.beginCanvasGesture","name":"Game.beginCanvasGesture","line":298,"end_line":300,"hash":"197ca42d03064c4be68a351f807238cf8abd0777b55b3cf6e830680482c8003d"},{"id":"func/Game.finishWorldPointer","name":"Game.finishWorldPointer","line":302,"end_line":312,"hash":"dd9b5731c3eac78cb42665e05cbf06fd46650e79abed992c1aec7f00e58acc23"},{"id":"func/Game.finishMassDrag","name":"Game.finishMassDrag","line":314,"end_line":326,"hash":"d122c857bfcadc2fef1d0d79913252932f42f186822a863005e14dd7567c14d4"},{"id":"func/Game.throwDraggedMasses","name":"Game.throwDraggedMasses","line":328,"end_line":335,"hash":"324af5dae0a04bbaf882c4e4cb4c8f59074f9072609810ebe5cf623253098479"},{"id":"func/Game.throwSelectedDraggingMasses","name":"Game.throwSelectedDraggingMasses","line":337,"end_line":343,"hash":"fbf3d78192c8d57825af0cba8ee19c25aae5185dc5794c5511409976a9e792ff"},{"id":"func/Game.throwSingleDraggingMass","name":"Game.throwSingleDraggingMass","line":345,"end_line":354,"hash":"bf7e32dc72ebe0f7a5be261d52144ed838ad07c5effab974ae39e465e6f12575"},{"id":"func/Game.createMassAt","name":"Game.createMassAt","line":356,"end_line":376,"hash":"f36ce45bc786657b451c10eee31db647adcfb45a7bfa536333ecb926cec4b45f"},{"id":"func/Game.selectNearest","name":"Game.selectNearest","line":378,"end_line":382,"hash":"9ab9d8b0a35b3835e1e75067cd70de1bcab571140dc0b30398970e65d21259f8"},{"id":"func/Game.beginSelectGesture","name":"Game.beginSelectGesture","line":384,"end_line":405,"hash":"45ad3e7fea3d28bb752bb27b89d1343c4d4ee1c22f5419512e8134c6036e3bc2"},{"id":"func/Game.finishSelectGesture","name":"Game.finishSelectGesture","line":407,"end_line":420,"hash":"0c40acff2337b4f2ff3d6b02ce099e224795749865190bc4214804202b73301c"},{"id":"func/selectionClick","name":"selectionClick","line":422,"end_line":424,"hash":"f62199bd3dc9a322b7d12c1c28e74fcc9cb030a3565813897710ba8c44a0a3c9"},{"id":"func/Game.beginSpringAt","name":"Game.beginSpringAt","line":426,"end_line":435,"hash":"7a9047db3306e9d1de75fa42aef1d88a8a65ce5bf48f7915fc518e5b1ca0b50a"},{"id":"func/Game.finishSpringAt","name":"Game.finishSpringAt","line":437,"end_line":447,"hash":"f8b5091082d2eee0006b96ade69cc1d29538aff4689bd3b667676bd57390fca7"},{"id":"func/Game.beginMassDrag","name":"Game.beginMassDrag","line":449,"end_line":460,"hash":"64fedf1ba7b7bd1564ba7c4890821810075631d2d3ed99eb4f6326aa95d282e0"},{"id":"func/Game.captureDraggingOffsets","name":"Game.captureDraggingOffsets","line":462,"end_line":470,"hash":"e95c3b367d00da8939c3a741020cecf45f7d14a04350147363edffe72e161f2b"},{"id":"func/Game.captureSelectedDraggingOffsets","name":"Game.captureSelectedDraggingOffsets","line":472,"end_line":482,"hash":"0f0daf6c57e4f765045bdde34f75bf3952ab775e997da8e6b5a0239b319b2d74"},{"id":"func/Game.pinDraggingMasses","name":"Game.pinDraggingMasses","line":484,"end_line":489,"hash":"e77e762693460c32f2edf1c072c72665017b8726b65480cc97f3c3c863b4e27c"},{"id":"func/Game.massAt","name":"Game.massAt","line":491,"end_line":499,"hash":"3ff80823eb103b2b67167abb1f6f1c560442cfac60cc1790b7a3d506f09d7ac6"},{"id":"func/massDrawCircle","name":"massDrawCircle","line":501,"end_line":503,"hash":"141b72a0cb028af6c4bf4c0bf2d03b72b5298f6cfa2c5e0a19537d66c7de9cdb"},{"id":"func/Game.screenToWorld","name":"Game.screenToWorld","line":505,"end_line":507,"hash":"20d063171fc096c5763e0a5add37194d9b46437547eea38be243c0595489ca4e"},{"id":"func/Game.worldToScreen","name":"Game.worldToScreen","line":509,"end_line":511,"hash":"e02c541b189ecfde32bc8b1576f8e472ce03671dd138c3445907b27b19eb2b71"},{"id":"func/Game.canvasCoordinate","name":"Game.canvasCoordinate","line":513,"end_line":518,"hash":"35a00aa3fe34efe7bf6dcfcc8e214822d61bb6a715c366d38b844342c079ab8b"},{"id":"func/Game.flipCanvasY","name":"Game.flipCanvasY","line":520,"end_line":522,"hash":"d83a41b60cd87dee4025d8245fd79519e489c0ebcb8016b38fbb9e56affd0f17"},{"id":"func/Game.editing","name":"Game.editing","line":524,"end_line":529,"hash":"bcf09a98a5f419f295a2fd6c7c7e85d7b1a9aaf656db9a844e7e63cdf465d673"},{"id":"func/Game.Layout","name":"Game.Layout","line":531,"end_line":533,"hash":"aab68cdc4f078367f499500c8a90603c494f128359287735d76907fe8d472cf0"},{"id":"func/Game.World","name":"Game.World","line":535,"end_line":537,"hash":"7fc326e28e4dfddaa10b2905d3e99ea6fd531ba2a6f4c96433203191da8b51c1"},{"id":"func/Game.SetPaused","name":"Game.SetPaused","line":539,"end_line":541,"hash":"a7f2a3c7ec5e2816b287d6c35c13d3e2f50cf91ff080f0402f932c4e43bf555b"},{"id":"func/Game.Paused","name":"Game.Paused","line":543,"end_line":545,"hash":"c8d2e386a87960313955f0049367542112021d80e438d778709fbb844c3a19d6"},{"id":"func/Game.InputActive","name":"Game.InputActive","line":547,"end_line":549,"hash":"f91d0fd54dc7aaff10c504b1c9bf4dbcfdfc0387add7cf29fbf873ed528ee66e"},{"id":"func/Game.RenderingActive","name":"Game.RenderingActive","line":551,"end_line":553,"hash":"988262465f3679184feef405a117d1050eb0a730ac6761f8c261e56e81a0657e"},{"id":"func/Game.RenderFrame","name":"Game.RenderFrame","line":555,"end_line":557,"hash":"6e9d3826815b3d345002e53402029fc9eec62a43124acc31742f795d57f7e7d5"},{"id":"func/Game.Close","name":"Game.Close","line":559,"end_line":562,"hash":"df8cfa513463e44f848de31c611e1efb6d0f4dfcfcba23e8cbf9c7947bf5a97e"},{"id":"func/Game.Closed","name":"Game.Closed","line":564,"end_line":566,"hash":"0fa4138145c5b4d4d164f60f1336e23b1759e704dfe312bfc890f2ede5e2e373"}]}
// mutate4go-manifest-end
