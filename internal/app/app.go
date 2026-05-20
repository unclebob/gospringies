//go:build !appunit

package app

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

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
	focusedNumeric    string
	numericInputText  string
	numericInputTicks int
	numericInputFresh bool
	massMenu          massContextMenu
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
	return &Game{simulation: world, initialState: world.Clone(), simulationSpeed: 1, canvasYUp: true, editor: edit.NewEditor(world)}
}

func DefaultWindowConfig() WindowConfig {
	return WindowConfig{Width: screenWidth, Height: screenHeight, Title: "Springs", Resizable: true}
}

func Run() error {
	config := DefaultWindowConfig()
	ebiten.SetWindowSize(config.Width, config.Height)
	ebiten.SetWindowTitle(config.Title)
	if config.Resizable {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	}
	ebiten.SetWindowClosingHandled(true)
	return ebiten.RunGame(NewGame())
}

func (g *Game) Update() error {
	g.handleWindowClose(ebiten.IsWindowBeingClosed())
	if g.closed {
		return ebiten.Termination
	}
	g.inputActive = true
	g.pollMouseControls()
	g.pollSaveFilenameDialogKeyboard()
	g.pollValueDialogKeyboard()
	g.tickValueDialog()
	g.tickNumericTextField()
	if !g.valueDialog.Open && !g.saveDialog.Open {
		g.pollNumericTextFieldKeyboard()
		if g.focusedNumeric == "" {
			g.pollKeyboardControls()
		}
	}
	g.pollDemoPickerScroll()
	if g.closed {
		return ebiten.Termination
	}
	g.advanceSimulationFrame()
	return nil
}

func (g *Game) pollDemoPickerScroll() {
	if !g.demoPickerOpen {
		return
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		g.scrollDemoPicker(int(-wheelY))
	}
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

func (g *Game) pollMouseControls() {
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	x, y := ebiten.CursorPosition()
	g.lastCursor = g.screenToWorld(simVec(x, y))
	g.handleRightPointer(rightPressed, x, y)
	g.handlePointer(leftPressed, x, y)
}

func (g *Game) pollKeyboardControls() {
	g.pollEscapeShortcut()
	if controlKeyPressed() {
		g.pollControlShortcuts()
	}
}

func (g *Game) pollEscapeShortcut() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.HandleShortcut("Esc")
	}
}

func (g *Game) pollControlShortcuts() {
	g.handlePressedShortcut(pressedControlShortcut())
}

func (g *Game) handlePressedShortcut(shortcut string) {
	if shortcut != "" {
		g.HandleShortcut(shortcut)
	}
}

func pressedControlShortcut() string {
	return pressedControlShortcutFrom(controlShortcutBindings)
}

func pressedControlShortcutFrom(bindings []keyShortcutBinding) string {
	if len(bindings) == 0 {
		return ""
	}
	return firstPressedShortcut(bindings[0], bindings[1:])
}

func firstPressedShortcut(binding keyShortcutBinding, remaining []keyShortcutBinding) string {
	if inpututil.IsKeyJustPressed(binding.key) {
		return binding.shortcut
	}
	return pressedControlShortcutFrom(remaining)
}

type keyShortcutBinding struct {
	key      ebiten.Key
	shortcut string
}

var controlShortcutBindings = []keyShortcutBinding{
	{key: ebiten.KeyX, shortcut: "Ctrl+X"},
	{key: ebiten.KeyC, shortcut: "Ctrl+C"},
	{key: ebiten.KeyV, shortcut: "Ctrl+V"},
	{key: ebiten.KeyA, shortcut: "Ctrl+A"},
	{key: ebiten.KeyD, shortcut: "Ctrl+D"},
	{key: ebiten.KeyS, shortcut: "Ctrl+S"},
	{key: ebiten.KeyO, shortcut: "Ctrl+O"},
	{key: ebiten.KeyI, shortcut: "Ctrl+I"},
}

func controlKeyPressed() bool {
	return anyKeyPressed(controlShortcutKeys, ebiten.IsKeyPressed)
}

func (g *Game) shiftKeyPressed() bool {
	return g.shiftDown || anyKeyPressed(shiftKeys, ebiten.IsKeyPressed)
}

func (g *Game) controlKeyPressed() bool {
	return g.controlDown || controlKeyPressed()
}

func (g *Game) throwKeyPressed() bool {
	return g.throwDown || ebiten.IsKeyPressed(ebiten.KeyT)
}

var controlShortcutKeys = []ebiten.Key{
	ebiten.KeyControl,
	ebiten.KeyControlLeft,
	ebiten.KeyControlRight,
	ebiten.KeyMeta,
	ebiten.KeyMetaLeft,
	ebiten.KeyMetaRight,
}

var shiftKeys = []ebiten.Key{
	ebiten.KeyShift,
	ebiten.KeyShiftLeft,
	ebiten.KeyShiftRight,
}

func anyKeyPressed(keys []ebiten.Key, pressed func(ebiten.Key) bool) bool {
	for _, key := range keys {
		if pressed(key) {
			return true
		}
	}
	return false
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
		g.pendingSpringEnd = position
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
	if g.controlKeyPressed() {
		g.controlPointerPress(position, x, y)
		return
	}
	if !g.ClickAt(x, y) {
		g.beginCanvasGesture(position)
	}
}

func (g *Game) clickOpenOverlay(x int, y int) bool {
	if g.saveDialog.Open {
		g.clickSaveFilenameDialog(x, y)
		return true
	}
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
		g.beginSpringAt(position)
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
	if g.pendingSpringID != 0 {
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
	editor := g.editing()
	editor.Mode = edit.ModeAddMass
	editor.GridSnapEnabled = g.gridSnapEnabled()
	editor.GridSnapSize = g.gridSnapSize()
	id, err := editor.Click(position)
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
	id, ok := g.massAt(position)
	if ok {
		g.pendingSpringID = id
		g.pendingSpringEnd = position
	}
}

func (g *Game) finishSpringAt(position sim.Vec2) {
	defer func() { g.pendingSpringID = 0 }()
	endID, ok := g.massAt(position)
	if !ok || endID == g.pendingSpringID {
		return
	}
	editor := g.editing()
	editor.Mode = edit.ModeAddSpring
	if _, err := editor.CreateSpring(g.pendingSpringID, endID); err == nil {
		g.dirty = true
	}
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

func (g *Game) Draw(screen *ebiten.Image) {
	result := g.RenderWorld()
	screen.Fill(backgroundColor)
	g.drawEditorChrome(screen)
	g.drawVisibleControls(screen)
	g.drawGridPoints(screen)
	if result.SpringLinesVisible {
		g.drawSprings(screen)
	}
	g.drawPendingSpring(screen)
	g.drawSelectionDrag(screen)
	g.drawMasses(screen)
	g.drawWalls(screen)
	if g.selected {
		g.drawSelection(screen)
	}
	if g.demoPickerOpen {
		g.drawDemoPicker(screen)
	}
	if g.massMenu.Open {
		g.drawMassContextMenu(screen)
	}
	if g.valueDialog.Open {
		g.drawValueDialog(screen)
	}
	if g.saveDialog.Open {
		g.drawSaveFilenameDialog(screen)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS %.0f", ebiten.ActualTPS()))
}

func (g *Game) drawEditorChrome(screen *ebiten.Image) {
	for _, rect := range editorChromeRects() {
		vector.DrawFilledRect(screen, rect.x, rect.y, rect.width, rect.height, rect.color, editorChromeAntiAlias())
	}
}

func editorChromeAntiAlias() bool {
	return false
}

type chromeRect struct {
	x      float32
	y      float32
	width  float32
	height float32
	color  color.RGBA
}

func editorChromeRects() []chromeRect {
	return []chromeRect{
		{x: 0, y: 0, width: screenWidth, height: topBarHeight, color: chromeColor},
		{x: 0, y: topBarHeight, width: toolbarWidth, height: screenHeight - topBarHeight - statusHeight, color: panelColor},
		{x: screenWidth - inspectorWidth, y: topBarHeight, width: inspectorWidth, height: screenHeight - topBarHeight - statusHeight, color: panelColor},
		{x: 0, y: screenHeight - statusHeight, width: screenWidth, height: statusHeight, color: chromeColor},
	}
}

func (g *Game) drawGridPoints(screen *ebiten.Image) {
	for _, rect := range g.gridPointRects() {
		vector.DrawFilledRect(screen, rect.x, rect.y, rect.width, rect.height, rect.color, rect.antiAlias)
	}
}

type fillRectDraw struct {
	x, y, width, height float32
	color               color.RGBA
	antiAlias           bool
}

func (g *Game) gridPointRects() []fillRectDraw {
	var rects []fillRectDraw
	for _, point := range g.gridPoints() {
		screenPoint := g.worldToScreen(point)
		rects = append(rects, fillRectDraw{
			x:         float32(screenPoint.X),
			y:         float32(screenPoint.Y),
			width:     gridPointPixelSize(),
			height:    gridPointPixelSize(),
			color:     gridPointColor,
			antiAlias: gridPointAntiAlias(),
		})
	}
	return rects
}

func gridPointPixelSize() float32 {
	return 1
}

func gridPointAntiAlias() bool {
	return false
}

func (g *Game) gridPoints() []sim.Vec2 {
	size := g.gridSnapSize()
	if !validGridSnapSize(size) {
		return nil
	}
	canvas := visibleRegionRects()["canvas"]
	left := firstGridCoordinateAtOrAfter(float64(canvas.Min.X), size)
	top := firstGridCoordinateAtOrAfter(0, size)
	points := []sim.Vec2{}
	for y := top; y <= g.simulation.Bounds.Height; y += size {
		for x := left; x < float64(canvas.Max.X); x += size {
			points = append(points, sim.Vec2{X: x, Y: y})
		}
	}
	return points
}

func validGridSnapSize(size float64) bool {
	return size > 0
}

func firstGridCoordinateAtOrAfter(min float64, size float64) float64 {
	return math.Ceil(min/size) * size
}

func (g *Game) drawSprings(screen *ebiten.Image) {
	for _, spring := range g.simulation.Springs {
		a, b, ok := g.springEndpoints(spring)
		if !ok {
			continue
		}
		drawSpringLine(screen, g.worldToScreen(a.Position), g.worldToScreen(b.Position), springColor)
	}
}

func (g *Game) drawPendingSpring(screen *ebiten.Image) {
	line, ok := g.pendingSpringLine()
	if !ok {
		return
	}
	drawSpringLine(screen, g.worldToScreen(sim.Vec2{X: line.x1, Y: line.y1}), g.worldToScreen(sim.Vec2{X: line.x2, Y: line.y2}), springColor)
}

func drawSpringLine(screen *ebiten.Image, a sim.Vec2, b sim.Vec2, color color.RGBA) {
	vector.StrokeLine(screen, float32(a.X), float32(a.Y), float32(b.X), float32(b.Y), springThickness, color, springLineAntiAlias())
}

func springLineAntiAlias() bool {
	return false
}

func (g *Game) pendingSpringLine() (selectionLine, bool) {
	if g.pendingSpringID == 0 {
		return selectionLine{}, false
	}
	start, ok := g.simulation.MassByID(g.pendingSpringID)
	if !ok {
		return selectionLine{}, false
	}
	return selectionLine{x1: start.Position.X, y1: start.Position.Y, x2: g.pendingSpringEnd.X, y2: g.pendingSpringEnd.Y}, true
}

func (g *Game) drawSelectionDrag(screen *ebiten.Image) {
	if !g.selectionDrag {
		return
	}
	for _, line := range selectionRectangleLines(g.selectionStart, g.selectionEnd) {
		g.drawSelectionLine(screen, line, selectionColor)
	}
}

func selectionRectangleLines(start sim.Vec2, end sim.Vec2) []selectionLine {
	left := math.Min(start.X, end.X)
	right := math.Max(start.X, end.X)
	top := math.Min(start.Y, end.Y)
	bottom := math.Max(start.Y, end.Y)
	return []selectionLine{
		{x1: left, y1: top, x2: right, y2: top},
		{x1: right, y1: top, x2: right, y2: bottom},
		{x1: right, y1: bottom, x2: left, y2: bottom},
		{x1: left, y1: bottom, x2: left, y2: top},
	}
}

func (g *Game) drawMasses(screen *ebiten.Image) {
	for _, mass := range g.simulation.Masses {
		screenPosition := g.worldToScreen(mass.Position)
		screenMass := mass
		screenMass.Position = screenPosition
		x, y, radius := massDrawCircle(screenMass)
		vector.DrawFilledCircle(screen, x, y, radius, massDrawColor(mass), massDrawAntiAlias())
	}
}

func massDrawAntiAlias() bool {
	return true
}

func massDrawCircle(mass sim.Mass) (float32, float32, float32) {
	return float32(mass.Position.X), float32(mass.Position.Y), float32(sim.MassRadius(mass))
}

func massDrawColor(mass sim.Mass) color.RGBA {
	if mass.Fixed {
		return fixedMassColor
	}
	return massColor
}

func (g *Game) drawWalls(screen *ebiten.Image) {
	drawWallLine := func(name string, x1, y1, x2, y2 float64) {
		if enabled, _ := g.simulation.Parameters.WallEnabled(name); enabled {
			ebitenutil.DrawLine(screen, x1, y1, x2, y2, wallColor)
		}
	}
	for _, line := range wallDrawLines(g.simulation.Bounds) {
		start := g.worldToScreen(sim.Vec2{X: line.x1, Y: line.y1})
		end := g.worldToScreen(sim.Vec2{X: line.x2, Y: line.y2})
		drawWallLine(line.name, start.X, start.Y, end.X, end.Y)
	}
}

type wallDrawLine struct {
	name           string
	x1, y1, x2, y2 float64
}

func wallDrawLines(bounds sim.Bounds) []wallDrawLine {
	return []wallDrawLine{
		{name: "top", x1: 0, y1: bounds.Height - 1, x2: bounds.Width, y2: bounds.Height - 1},
		{name: "bottom", x1: 0, y1: 0, x2: bounds.Width, y2: 0},
		{name: "left", x1: 0, y1: 0, x2: 0, y2: bounds.Height},
		{name: "right", x1: bounds.Width - 1, y1: 0, x2: bounds.Width - 1, y2: bounds.Height},
	}
}

func (g *Game) drawSelection(screen *ebiten.Image) {
	for _, line := range selectedMassOutline(g.selectedMasses()) {
		g.drawSelectionLine(screen, line, selectionColor)
	}
	for _, line := range g.selectedSpringLines() {
		g.drawSelectionLine(screen, line, selectionColor)
	}
}

func (g *Game) drawSelectionLine(screen *ebiten.Image, line selectionLine, color color.RGBA) {
	start := g.worldToScreen(sim.Vec2{X: line.x1, Y: line.y1})
	end := g.worldToScreen(sim.Vec2{X: line.x2, Y: line.y2})
	ebitenutil.DrawLine(screen, start.X, start.Y, end.X, end.Y, color)
}

func (g *Game) selectedMasses() []sim.Mass {
	selected := g.explicitSelectedMasses()
	if len(selected) == 0 && g.allMassesImplicitlySelected() {
		return g.simulation.Masses
	}
	return selected
}

func (g *Game) explicitSelectedMasses() []sim.Mass {
	var selected []sim.Mass
	for _, mass := range g.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selected = append(selected, mass)
		}
	}
	return selected
}

func (g *Game) allMassesImplicitlySelected() bool {
	return g.selected && len(g.selectedSpringLines()) == 0
}

func (g *Game) selectedSpringLines() []selectionLine {
	var lines []selectionLine
	for _, spring := range g.simulation.Springs {
		if !g.editing().SelectedSprings[spring.ID] {
			continue
		}
		a, okA := g.simulation.MassByID(spring.MassA)
		b, okB := g.simulation.MassByID(spring.MassB)
		if okA && okB {
			lines = append(lines, selectionLine{x1: a.Position.X, y1: a.Position.Y, x2: b.Position.X, y2: b.Position.Y})
		}
	}
	return lines
}

type selectionLine struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

func selectedMassOutline(masses []sim.Mass) []selectionLine {
	var lines []selectionLine
	for _, mass := range masses {
		lines = append(lines, selectionOutline(mass)...)
	}
	return lines
}

func selectionOutline(mass sim.Mass) []selectionLine {
	x := mass.Position.X
	y := mass.Position.Y
	radius := sim.MassRadius(mass) + 3
	return []selectionLine{
		{x - radius, y - radius, x + radius, y - radius},
		{x + radius, y - radius, x + radius, y + radius},
		{x + radius, y + radius, x - radius, y + radius},
		{x - radius, y + radius, x - radius, y - radius},
	}
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
// {"version":1,"tested_at":"2026-05-19T11:18:55-05:00","module_hash":"f18581571824587c512fa7b1a60ab0eba5f5ad962ebf07feb5e59217cfabc7a1","functions":[{"id":"func/NewGame","name":"NewGame","line":92,"end_line":95,"hash":"3b7fcae20e81266dc069f835423e617ef0bd524615b449b217304bc9b2c83e07"},{"id":"func/DefaultWindowConfig","name":"DefaultWindowConfig","line":97,"end_line":99,"hash":"d07a778f3e78f2a72f6ae58d2c7afa2ecd83053f575705b5fa26512178690d50"},{"id":"func/Run","name":"Run","line":101,"end_line":110,"hash":"0f29353f1e39f4004a508c545ca3f9cb53a79f5e92f6650d01ba41b79b91eacd"},{"id":"func/Game.Update","name":"Game.Update","line":112,"end_line":130,"hash":"df7b78f833b8e8be5f99d35a8d1e45b3cc4602e65d7bfa35ddc4b533e42dc848"},{"id":"func/Game.pollDemoPickerScroll","name":"Game.pollDemoPickerScroll","line":132,"end_line":140,"hash":"b61abfadca5c8c9e3de5137aed8d34d8bc7e2255ac0fb59ab080dbc756613bab"},{"id":"func/Game.advanceSimulationFrame","name":"Game.advanceSimulationFrame","line":142,"end_line":147,"hash":"0ce9cc0ed29e1bce0f1027b9d69abb5fdc6152fa91fc19d386447c8983f616a0"},{"id":"func/Game.handleWindowClose","name":"Game.handleWindowClose","line":149,"end_line":153,"hash":"1fa7c9632d93845006cbfb2e5987fe93c53cc4060e482fd661e58389d2aa0d74"},{"id":"func/Game.pollMouseControls","name":"Game.pollMouseControls","line":155,"end_line":162,"hash":"c9e7904ca942b11dec765dfe999114acaede29891e05271b9347c96031186d96"},{"id":"func/Game.pollKeyboardControls","name":"Game.pollKeyboardControls","line":164,"end_line":169,"hash":"974e7bd1c08ce7c0ff18eaa2303532542899782abc429edc7a23533a29eee794"},{"id":"func/Game.pollEscapeShortcut","name":"Game.pollEscapeShortcut","line":171,"end_line":175,"hash":"5e86654ef5f99460de37fc787add2d7f43966d7c9f4a4df0412a31ade98cdf3d"},{"id":"func/Game.pollControlShortcuts","name":"Game.pollControlShortcuts","line":177,"end_line":179,"hash":"2b58128a3e1003d965513b0da3e231a3783dbe1118ab1cb907dfbced3f534678"},{"id":"func/Game.handlePressedShortcut","name":"Game.handlePressedShortcut","line":181,"end_line":185,"hash":"6ad069ec21c2d4e4496b5a863c312f2882e0a6da6d67606c8c4b4f1d241ba05c"},{"id":"func/pressedControlShortcut","name":"pressedControlShortcut","line":187,"end_line":189,"hash":"c5c29917c6dfe8a97ccc7393c4e23c99392e60e240475f60f3aa9c0fb03c71ec"},{"id":"func/pressedControlShortcutFrom","name":"pressedControlShortcutFrom","line":191,"end_line":196,"hash":"2e9517d5ce38178ce08c01ffb56d86111d420144cd8f683b72f2e25e8401be66"},{"id":"func/firstPressedShortcut","name":"firstPressedShortcut","line":198,"end_line":203,"hash":"45ca55d96c80e333dfef91bdcb0c7475cc8ace90c20baca29774a17ad3e5ca18"},{"id":"func/controlKeyPressed","name":"controlKeyPressed","line":221,"end_line":223,"hash":"c8e3be2e7367aa320eeda237c2acda15d16eb3ea104cdd3453f372bfc829cd29"},{"id":"func/Game.shiftKeyPressed","name":"Game.shiftKeyPressed","line":225,"end_line":227,"hash":"c077d9b2a4c0195ba46e94fbb45b0107ede221696aa166735486693ff68c7c62"},{"id":"func/Game.controlKeyPressed","name":"Game.controlKeyPressed","line":229,"end_line":231,"hash":"e321d6cff56269f4f261bac368f4cb7481d02492fb5bdc42aad60abed79972d5"},{"id":"func/Game.throwKeyPressed","name":"Game.throwKeyPressed","line":233,"end_line":235,"hash":"60316b36601a1fbdecf3ba353c923595519b43efaa8e1d15e2bb78643f103189"},{"id":"func/anyKeyPressed","name":"anyKeyPressed","line":252,"end_line":259,"hash":"ff2843aefeb8b3b09a8cf5ee32470ec8d0eb3879cb8daeb672430cf8085933b7"},{"id":"func/Game.handleRightPointer","name":"Game.handleRightPointer","line":261,"end_line":266,"hash":"87077ecef7b95720a1a172ce0f030917d018efc34e3e23c46eba23b4344496d5"},{"id":"func/Game.handlePointer","name":"Game.handlePointer","line":268,"end_line":276,"hash":"fe19fadb0f52cc482d787ef61edbeff3cb591e665491cb2e9724e957a53efc50"},{"id":"func/Game.handlePressedPointer","name":"Game.handlePressedPointer","line":278,"end_line":284,"hash":"db03eeacbdc454bbb95e8b15369100910a380ddda7d13ad4e2a004356af49743"},{"id":"func/Game.continuePointerPress","name":"Game.continuePointerPress","line":286,"end_line":297,"hash":"e5fe1a43c9cef588763f78f4a53e14df71e7df9da2c00804c0fe881c12496df2"},{"id":"func/Game.releasePointer","name":"Game.releasePointer","line":299,"end_line":306,"hash":"4691b9f28270ef9fe9aa4f8517a9f29a49ddb567a15fbf0f3a0d11f049803a8b"},{"id":"func/Game.beginPointerPress","name":"Game.beginPointerPress","line":308,"end_line":319,"hash":"1bc26b1dda0f18e72711eb44ff8b23a78ea7c655b4a0e5442fcf4192fb185db5"},{"id":"func/Game.clickOpenOverlay","name":"Game.clickOpenOverlay","line":321,"end_line":335,"hash":"4b1405a9c229651c3de4ccdea17a5c1ad5afa0afb7113944468d29a81bb16e40"},{"id":"func/Game.controlPointerPress","name":"Game.controlPointerPress","line":337,"end_line":341,"hash":"eef0d16e331e46a6dbfde825fd181ba7b78e460ecb317ddf0c16d1f9560b0fe9"},{"id":"func/Game.scrollDemoPicker","name":"Game.scrollDemoPicker","line":343,"end_line":349,"hash":"4ce9154527c8e87cb048f53ab5497a40d1008b9a2b8c99914a702416aa14dccf"},{"id":"func/Game.demoList","name":"Game.demoList","line":351,"end_line":365,"hash":"8182ed61297a149f6ceddd95fa158919a1cc3d42a2cd1f086d2852490289a67e"},{"id":"func/Game.beginCanvasGesture","name":"Game.beginCanvasGesture","line":367,"end_line":369,"hash":"197ca42d03064c4be68a351f807238cf8abd0777b55b3cf6e830680482c8003d"},{"id":"func/Game.finishWorldPointer","name":"Game.finishWorldPointer","line":371,"end_line":381,"hash":"76cbef9e5f40b08c24da2a95495f1e35a44789b6f3d0ad15b51ff25ef0bf31d6"},{"id":"func/Game.finishMassDrag","name":"Game.finishMassDrag","line":383,"end_line":395,"hash":"f2e00273da1f0b377ecd693516c34b59924897c79c5de9de569a8bbc980db20f"},{"id":"func/Game.throwDraggedMasses","name":"Game.throwDraggedMasses","line":397,"end_line":404,"hash":"8c693251aceac50f8d148444b4a3608a340685a1e10faf501576870fae456fcd"},{"id":"func/Game.throwSelectedDraggingMasses","name":"Game.throwSelectedDraggingMasses","line":406,"end_line":412,"hash":"b58525c194991999c567664e4acf44b52d99ddb1635418e1f9995340e6411e32"},{"id":"func/Game.throwSingleDraggingMass","name":"Game.throwSingleDraggingMass","line":414,"end_line":423,"hash":"d7a6b79eeadb7d27602ff2e0b02de42aaccaa148953b5072912ac2efe99ed143"},{"id":"func/Game.createMassAt","name":"Game.createMassAt","line":425,"end_line":442,"hash":"6afcf8ab8411ebaad284c4e871e8cb984d2208f53f1fb4e5641c49a11516e573"},{"id":"func/Game.selectNearest","name":"Game.selectNearest","line":444,"end_line":448,"hash":"c64e5503f6398dce55003b362d508180e35e2b0c3d6ab50d0be690c49a26789b"},{"id":"func/Game.beginSelectGesture","name":"Game.beginSelectGesture","line":450,"end_line":471,"hash":"17786f92e1a8ce760a79bcbdb7e2409b9030bf3b35a6f7d9fa88b68c0ae4b0f8"},{"id":"func/Game.finishSelectGesture","name":"Game.finishSelectGesture","line":473,"end_line":486,"hash":"9b835315af20eddf845eb906462642fcf4454aed35909cefbb2a2405ec03d8cb"},{"id":"func/selectionClick","name":"selectionClick","line":488,"end_line":490,"hash":"f62199bd3dc9a322b7d12c1c28e74fcc9cb030a3565813897710ba8c44a0a3c9"},{"id":"func/Game.beginSpringAt","name":"Game.beginSpringAt","line":492,"end_line":498,"hash":"178ab75f0518ae4be59519263fd8c3be3d7e0eb914645b7a6d28166457897cf9"},{"id":"func/Game.finishSpringAt","name":"Game.finishSpringAt","line":500,"end_line":511,"hash":"77b80c5f818bcbbbffd30875b3db42f7ea9a0de0b06ad0df7a16cd1a6068c4ed"},{"id":"func/Game.beginMassDrag","name":"Game.beginMassDrag","line":513,"end_line":524,"hash":"b358e1665e4648ca7ea0b1c1a5e700f2749a723cfc0bd08bb79b9cc9af6574a9"},{"id":"func/Game.captureDraggingOffsets","name":"Game.captureDraggingOffsets","line":526,"end_line":534,"hash":"15f5291a75104a9d5cc437c5461ad4d7cec8ee3f366120613d6dc4cf29c81bc4"},{"id":"func/Game.captureSelectedDraggingOffsets","name":"Game.captureSelectedDraggingOffsets","line":536,"end_line":546,"hash":"a4d6564b62d2b3b4b517dd14511c53cd7cc3398ca91573180a46b61c356fd134"},{"id":"func/Game.pinDraggingMasses","name":"Game.pinDraggingMasses","line":548,"end_line":553,"hash":"5470b2d59632dd7656bf5e6e380a3836631891d8008696c89fadcbe55623775d"},{"id":"func/Game.massAt","name":"Game.massAt","line":555,"end_line":563,"hash":"473e6fecdf13bba51406f9b8e6111187d2cc4033dd3924d4490a27a32a94a25c"},{"id":"func/Game.screenToWorld","name":"Game.screenToWorld","line":565,"end_line":570,"hash":"f0df287eb9fc0424e076cdbca4c42d25d14930557aecb5913067dd412ad04538"},{"id":"func/Game.worldToScreen","name":"Game.worldToScreen","line":572,"end_line":577,"hash":"d97f25c8515c0087204eda57d49caf8b353e6e97956cf40bd4b256b8e12ac8b3"},{"id":"func/Game.editing","name":"Game.editing","line":579,"end_line":584,"hash":"d8b1ccdd85577daec4370de59fa0271f4eade8301137406cf879ffdc86f5d2e5"},{"id":"func/Game.Draw","name":"Game.Draw","line":586,"end_line":612,"hash":"d1c165b20977526205865590e5404ee292a1964c37adf5738d2bd373dc5e8ce4"},{"id":"func/Game.drawEditorChrome","name":"Game.drawEditorChrome","line":614,"end_line":618,"hash":"472fbea691e81ccd3a0f783e20fcbd32801d87ee3f99e89f0ca941bc28f56659"},{"id":"func/editorChromeAntiAlias","name":"editorChromeAntiAlias","line":620,"end_line":622,"hash":"660ffdb4c15b9f1e1b71b357db91cc1e3ccf02d18a5b93bd19da971f648614cf"},{"id":"func/editorChromeRects","name":"editorChromeRects","line":632,"end_line":639,"hash":"fea953ccb9204486399eb57207b36cd08bcf1c936372c271886a721f417a6630"},{"id":"func/Game.drawGridPoints","name":"Game.drawGridPoints","line":641,"end_line":645,"hash":"62ca4d64fc25bc322d533f258709357d5a112e79edb94e8d0332836a21cedcf7"},{"id":"func/Game.gridPointRects","name":"Game.gridPointRects","line":653,"end_line":667,"hash":"600750169ed58b07bd434dafed0f8a6fd143d81bade263f108fed60c984f43ee"},{"id":"func/gridPointPixelSize","name":"gridPointPixelSize","line":669,"end_line":671,"hash":"6c0a1cf159cdd1cb2dcab7eaf4f8d631777564f781ac420f6d26dab7a0fe6730"},{"id":"func/gridPointAntiAlias","name":"gridPointAntiAlias","line":673,"end_line":675,"hash":"9da19824153c650376ff119c27f1fb0c1582d8b0411321824a1d13b1e594d627"},{"id":"func/Game.gridPoints","name":"Game.gridPoints","line":677,"end_line":692,"hash":"3ed63f3893e22209194c76edc65efcc440d7606975675693c8e8e902c506a26b"},{"id":"func/validGridSnapSize","name":"validGridSnapSize","line":694,"end_line":696,"hash":"cf2cc0632f5d59e877c981dca062e63dac4c535b77c3065878c5c3a34b22b37e"},{"id":"func/firstGridCoordinateAtOrAfter","name":"firstGridCoordinateAtOrAfter","line":698,"end_line":700,"hash":"8a7c004c8f82076f9188bb1776b097331b802ef718b63f68bdfd2eae4acf9ec7"},{"id":"func/Game.drawSprings","name":"Game.drawSprings","line":702,"end_line":710,"hash":"c8b65e9056c613b98a4aaa516d58baccbb7bf4d6a31ec7c989d8758e8992204e"},{"id":"func/Game.drawPendingSpring","name":"Game.drawPendingSpring","line":712,"end_line":718,"hash":"8b2f10d92cea34baa15527fdf72126c1b61468c391cf7a633c81b6d273286aa9"},{"id":"func/drawSpringLine","name":"drawSpringLine","line":720,"end_line":722,"hash":"bd344626244ed0ae0211b6aadf761f458422111204fb50d9f061d4380b2d8113"},{"id":"func/springLineAntiAlias","name":"springLineAntiAlias","line":724,"end_line":726,"hash":"c16077efe2092cbf3ba89b04223ef9f600017a1311b6169e39543d7b88ac6d16"},{"id":"func/Game.pendingSpringLine","name":"Game.pendingSpringLine","line":728,"end_line":737,"hash":"975522976ed7a5264ac3efe69cd0d6af5d4e4d855b1e42f73e72e3b7f1dbd7b9"},{"id":"func/Game.drawSelectionDrag","name":"Game.drawSelectionDrag","line":739,"end_line":746,"hash":"30b96b9448f285745d3b893436841177d4db28606ef2444993030aa9ea2be8e2"},{"id":"func/selectionRectangleLines","name":"selectionRectangleLines","line":748,"end_line":759,"hash":"ac6138efe5e031f4de8e884ce3557eafe1b6b8ae5f95fba79c0fd3846af14d8e"},{"id":"func/Game.drawMasses","name":"Game.drawMasses","line":761,"end_line":769,"hash":"77cd2b63c3782a94dec70dc2d987cd4bf79367988522a777f5e1ad804d6edf92"},{"id":"func/massDrawAntiAlias","name":"massDrawAntiAlias","line":771,"end_line":773,"hash":"948e921ea933de29f25b79483885ab252d33194f86069a946d34f1508816a3c9"},{"id":"func/massDrawCircle","name":"massDrawCircle","line":775,"end_line":777,"hash":"141b72a0cb028af6c4bf4c0bf2d03b72b5298f6cfa2c5e0a19537d66c7de9cdb"},{"id":"func/massDrawColor","name":"massDrawColor","line":779,"end_line":784,"hash":"679289860205986b8f962cdaf4ac58fa39a55f4ad0dd5362f016ab7fbc3f523e"},{"id":"func/Game.drawWalls","name":"Game.drawWalls","line":786,"end_line":797,"hash":"6dcdcfc131281c68a6f5eb276a9e7d3fed0a95e700a5819612b1216134217b62"},{"id":"func/wallDrawLines","name":"wallDrawLines","line":804,"end_line":811,"hash":"6fde68dc5cf62acf85f8df7d95af2bb110081541285e6ba6438da786b9bdd056"},{"id":"func/Game.drawSelection","name":"Game.drawSelection","line":813,"end_line":820,"hash":"837fdf71a9d781c78495e22db390fe530af2786a8feacb4628b61fd157af5d53"},{"id":"func/Game.drawSelectionLine","name":"Game.drawSelectionLine","line":822,"end_line":826,"hash":"a0630296c8ec7d7e6289543381e43546b4c63a7cc91a720d920bde7e25b14c92"},{"id":"func/Game.selectedMasses","name":"Game.selectedMasses","line":828,"end_line":834,"hash":"69460d8bb53c240bc6e007a675ef4c0ae04cfabb3881d27f7d8b1a46f4c23685"},{"id":"func/Game.explicitSelectedMasses","name":"Game.explicitSelectedMasses","line":836,"end_line":844,"hash":"827934703f8bb40618ad6aa5cea693c0b401951d329c653a3bddbefdb74dfac7"},{"id":"func/Game.allMassesImplicitlySelected","name":"Game.allMassesImplicitlySelected","line":846,"end_line":848,"hash":"029656de5843af2d55357d06ec6b68a6054bf70cedee173deb702c2a913af876"},{"id":"func/Game.selectedSpringLines","name":"Game.selectedSpringLines","line":850,"end_line":863,"hash":"fe57ce268bb404f83a6cce3e75b4591baa3d466fd8f2f9a73bf655f11fe523a3"},{"id":"func/selectedMassOutline","name":"selectedMassOutline","line":872,"end_line":878,"hash":"09e522ba39856beab87cf554a1a6ce1af7c4466412f407118cc223529ae81da3"},{"id":"func/selectionOutline","name":"selectionOutline","line":880,"end_line":890,"hash":"d4ec3f8743e56bd8e3bed0d58f60210a5195c24b29e99c152f2e5b09b8806bdf"},{"id":"func/Game.Layout","name":"Game.Layout","line":892,"end_line":894,"hash":"aab68cdc4f078367f499500c8a90603c494f128359287735d76907fe8d472cf0"},{"id":"func/Game.World","name":"Game.World","line":896,"end_line":898,"hash":"a9270d11e4300269b96797a052fe66c1e96ff42aac546663ea66d24b7f48c6ab"},{"id":"func/Game.SetPaused","name":"Game.SetPaused","line":900,"end_line":902,"hash":"022f698a4f568a12c8b3a2fe5723316d1c01b16b9e1f722d69c28b666b683e27"},{"id":"func/Game.Paused","name":"Game.Paused","line":904,"end_line":906,"hash":"423f86d8b651a7bdfc4f3430ac2d8e0369da3177056a0b8d6d2ee6c26b45aab9"},{"id":"func/Game.InputActive","name":"Game.InputActive","line":908,"end_line":910,"hash":"8ba126aecb56afb862b985e28fdf1144ebb879230d2e659cca1c16327817aea3"},{"id":"func/Game.RenderingActive","name":"Game.RenderingActive","line":912,"end_line":914,"hash":"1f935194ba70ba6592f1e82561e383760fc4b5907912de3ec32c3611503570f8"},{"id":"func/Game.RenderFrame","name":"Game.RenderFrame","line":916,"end_line":918,"hash":"93435d752bdaeafc1e6f52105b61156503e4a1acdc8f0f7ad4ddbafb0dfef512"},{"id":"func/Game.Close","name":"Game.Close","line":920,"end_line":923,"hash":"05d19586766eb286b92af2eead38992b7cfcdcdc358efc607ef6a2f1504cb980"},{"id":"func/Game.Closed","name":"Game.Closed","line":925,"end_line":927,"hash":"cd43811f993dc6eb808f5aeec63512934868aa0e2fa85aa73710dc4aef48861f"}]}
// mutate4go-manifest-end
