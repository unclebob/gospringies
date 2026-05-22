package app

import (
	"image/color"

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

func (g *Game) markDirty() {
	g.editState.dirty = true
}

func (g *Game) clearDirty() {
	g.editState.dirty = false
}

func (g *Game) setSelected(selected bool) {
	g.editState.selected = selected
}

func (g *Game) reattachEditor() {
	if g.world.editor != nil {
		g.world.editor.World = g.world.simulation
	}
}

func (g *Game) loadWorldIntoSession(world *sim.Simulation) {
	g.world.simulation.LoadFrom(world)
	g.reattachEditor()
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
// {"version":1,"tested_at":"2026-05-22T09:43:56-05:00","module_hash":"26ecd9a96537b895bf864d7bca061efc2d5a78885d85dd58433d92b3e8015a44","functions":[{"id":"func/NewGame","name":"NewGame","line":133,"end_line":145,"hash":"47db37f71ce2ab29ddf870ff90facecdd6dd22775c6cbf35318a8e8f5cec342c"},{"id":"func/DefaultWindowConfig","name":"DefaultWindowConfig","line":147,"end_line":149,"hash":"d07a778f3e78f2a72f6ae58d2c7afa2ecd83053f575705b5fa26512178690d50"},{"id":"func/Game.advanceSimulationFrame","name":"Game.advanceSimulationFrame","line":151,"end_line":156,"hash":"f8c98b9b8bf8859017b73642bbc7ec8a1853ae067f6e5fbddc91864bc8d1d135"},{"id":"func/Game.markDirty","name":"Game.markDirty","line":158,"end_line":160,"hash":"c4dfca7e7db390641a8cc3f8262a95482d54dd0998eb217837e67fe669474d96"},{"id":"func/Game.clearDirty","name":"Game.clearDirty","line":162,"end_line":164,"hash":"502ee4bacbb36f914c8ad754f440053d0604bb470e261e3b24106854bba725ff"},{"id":"func/Game.setSelected","name":"Game.setSelected","line":166,"end_line":168,"hash":"e946d021304ebd3b7c3cc8efb4e714b01f39f63c13f30ff5f434e6c51e70b2f9"},{"id":"func/Game.reattachEditor","name":"Game.reattachEditor","line":170,"end_line":174,"hash":"5deaa40024bec2170983ca99be9ea20ff2f49d8f327fa2ae459deb911d92b433"},{"id":"func/Game.loadWorldIntoSession","name":"Game.loadWorldIntoSession","line":176,"end_line":179,"hash":"169a3633e75c21a85390478d25b2721a5e38b8c077558d39303483940565d10a"},{"id":"func/Game.scrollDemoPicker","name":"Game.scrollDemoPicker","line":181,"end_line":187,"hash":"c78d6e2d96e418e6daff83f002e782966140a5c37e40377a73af73d140b532a8"},{"id":"func/Game.demoList","name":"Game.demoList","line":189,"end_line":195,"hash":"a2b251e6bb5464680774c9189a40f4ccce99406502b8341fa65e4f80891a1a46"},{"id":"func/Game.editing","name":"Game.editing","line":197,"end_line":202,"hash":"bcf09a98a5f419f295a2fd6c7c7e85d7b1a9aaf656db9a844e7e63cdf465d673"},{"id":"func/Game.Layout","name":"Game.Layout","line":204,"end_line":206,"hash":"aab68cdc4f078367f499500c8a90603c494f128359287735d76907fe8d472cf0"},{"id":"func/Game.World","name":"Game.World","line":208,"end_line":210,"hash":"7fc326e28e4dfddaa10b2905d3e99ea6fd531ba2a6f4c96433203191da8b51c1"},{"id":"func/Game.SetPaused","name":"Game.SetPaused","line":212,"end_line":214,"hash":"a7f2a3c7ec5e2816b287d6c35c13d3e2f50cf91ff080f0402f932c4e43bf555b"},{"id":"func/Game.Paused","name":"Game.Paused","line":216,"end_line":218,"hash":"c8d2e386a87960313955f0049367542112021d80e438d778709fbb844c3a19d6"},{"id":"func/Game.InputActive","name":"Game.InputActive","line":220,"end_line":222,"hash":"f91d0fd54dc7aaff10c504b1c9bf4dbcfdfc0387add7cf29fbf873ed528ee66e"},{"id":"func/Game.RenderingActive","name":"Game.RenderingActive","line":224,"end_line":226,"hash":"988262465f3679184feef405a117d1050eb0a730ac6761f8c261e56e81a0657e"},{"id":"func/Game.RenderFrame","name":"Game.RenderFrame","line":228,"end_line":230,"hash":"6e9d3826815b3d345002e53402029fc9eec62a43124acc31742f795d57f7e7d5"},{"id":"func/Game.Close","name":"Game.Close","line":232,"end_line":235,"hash":"df8cfa513463e44f848de31c611e1efb6d0f4dfcfcba23e8cbf9c7947bf5a97e"},{"id":"func/Game.Closed","name":"Game.Closed","line":237,"end_line":239,"hash":"0fa4138145c5b4d4d164f60f1336e23b1759e704dfe312bfc890f2ede5e2e373"}]}
// mutate4go-manifest-end
