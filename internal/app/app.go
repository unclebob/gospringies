package app

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"sort"

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
	massMenu          massContextMenu
	valueDialog       valueDialog
	editMenuOpen      bool
	editClipboard     editClipboard
	lastCursor        sim.Vec2
	demoPickerOpen    bool
	demoPickerScroll  int
	demoFiles         []string
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
	g.pollValueDialogKeyboard()
	g.tickValueDialog()
	if !g.valueDialog.Open {
		g.pollKeyboardControls()
	}
	if g.demoPickerOpen {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 {
			g.scrollDemoPicker(int(-wheelY))
		}
	}
	if g.closed {
		return ebiten.Termination
	}
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

func (g *Game) pollMouseControls() {
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	x, y := ebiten.CursorPosition()
	g.lastCursor = g.screenToWorld(simVec(x, y))
	g.handleRightPointer(rightPressed, x, y)
	g.handlePointer(leftPressed, x, y)
}

func (g *Game) pollKeyboardControls() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.HandleShortcut("Esc")
	}
	if !controlKeyPressed() {
		return
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyX):
		g.HandleShortcut("Ctrl+X")
	case inpututil.IsKeyJustPressed(ebiten.KeyC):
		g.HandleShortcut("Ctrl+C")
	case inpututil.IsKeyJustPressed(ebiten.KeyV):
		g.HandleShortcut("Ctrl+V")
	case inpututil.IsKeyJustPressed(ebiten.KeyA):
		g.HandleShortcut("Ctrl+A")
	case inpututil.IsKeyJustPressed(ebiten.KeyD):
		g.HandleShortcut("Ctrl+D")
	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		g.HandleShortcut("Ctrl+S")
	case inpututil.IsKeyJustPressed(ebiten.KeyO):
		g.HandleShortcut("Ctrl+O")
	case inpututil.IsKeyJustPressed(ebiten.KeyI):
		g.HandleShortcut("Ctrl+I")
	}
}

func controlKeyPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight) ||
		ebiten.IsKeyPressed(ebiten.KeyMeta) ||
		ebiten.IsKeyPressed(ebiten.KeyMetaLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyMetaRight)
}

func (g *Game) shiftKeyPressed() bool {
	return g.shiftDown ||
		ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
}

func (g *Game) controlKeyPressed() bool {
	return g.controlDown || controlKeyPressed()
}

func (g *Game) throwKeyPressed() bool {
	return g.throwDown || ebiten.IsKeyPressed(ebiten.KeyT)
}

func (g *Game) handleRightPointer(pressed bool, x int, y int) {
	if pressed && !g.rightMousePressed {
		g.openContextAt(x, y)
	}
	g.rightMousePressed = pressed
}

func (g *Game) handlePointer(pressed bool, x int, y int) {
	position := g.screenToWorld(simVec(x, y))
	if pressed && !g.mousePressed {
		if g.valueDialog.Open {
			g.clickValueDialog(x, y)
			g.mousePressed = pressed
			return
		}
		if g.massMenu.Open {
			g.clickMassContextMenu(x, y)
			g.mousePressed = pressed
			return
		}
		if g.demoPickerOpen {
			g.clickDemoPicker(x, y)
			g.mousePressed = pressed
			return
		}
		if g.controlKeyPressed() {
			if !g.ClickAt(x, y) {
				g.beginSpringAt(position)
			}
			g.mousePressed = pressed
			return
		}
		if !g.ClickAt(x, y) {
			g.beginCanvasGesture(position)
		}
	} else if pressed && g.draggingMassID != 0 {
		g.DragMass(g.draggingMassID, position)
	} else if pressed && g.pendingSpringID != 0 {
		g.pendingSpringEnd = position
	} else if pressed && g.selectionDrag {
		g.selectionEnd = position
	} else if pressed && g.activeSlider != "" {
		g.setSliderAt(g.activeSlider, x)
	} else if !pressed {
		g.finishWorldPointer(position)
		g.draggingMassID = 0
		g.draggingOffsets = nil
		g.dragMoved = false
		g.selectionDrag = false
		g.activeSlider = ""
	}
	g.mousePressed = pressed
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
	var matches []string
	for _, root := range []string{"demos", filepath.Join("..", "..", "demos")} {
		original, _ := filepath.Glob(filepath.Join(root, "original", "*.xsp"))
		starter, _ := filepath.Glob(filepath.Join(root, "*.xsp"))
		matches = append(matches, original...)
		matches = append(matches, starter...)
	}
	sort.Strings(matches)
	g.demoFiles = matches
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
		for i := range g.simulation.Masses {
			if _, ok := g.draggingOffsets[g.simulation.Masses[i].ID]; ok {
				g.simulation.Masses[i].Velocity = velocity
			}
		}
		g.dirty = true
		return
	}
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == g.draggingMassID {
			g.simulation.Masses[i].Velocity = velocity
			g.dirty = true
			return
		}
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
	if len(g.editing().SelectedMasses) > 0 && g.editing().MassSelected(g.draggingMassID) {
		for _, mass := range g.simulation.Masses {
			if g.editing().MassSelected(mass.ID) {
				g.draggingOffsets[mass.ID] = mass.Position.Sub(cursor)
			}
		}
		return
	}
	if mass, ok := g.simulation.MassByID(g.draggingMassID); ok {
		g.draggingOffsets[g.draggingMassID] = mass.Position.Sub(cursor)
	}
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
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS %.0f", ebiten.ActualTPS()))
}

func (g *Game) drawEditorChrome(screen *ebiten.Image) {
	for _, rect := range editorChromeRects() {
		vector.DrawFilledRect(screen, rect.x, rect.y, rect.width, rect.height, rect.color, false)
	}
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
	for _, point := range g.gridPoints() {
		screenPoint := g.worldToScreen(point)
		vector.DrawFilledRect(screen, float32(screenPoint.X), float32(screenPoint.Y), 1, 1, gridPointColor, false)
	}
}

func (g *Game) gridPoints() []sim.Vec2 {
	size := g.gridSnapSize()
	if size <= 0 {
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
	vector.StrokeLine(screen, float32(a.X), float32(a.Y), float32(b.X), float32(b.Y), springThickness, color, false)
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
	var selected []sim.Mass
	for _, mass := range g.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selected = append(selected, mass)
		}
	}
	if len(selected) == 0 && g.selected && len(g.selectedSpringLines()) == 0 {
		return g.simulation.Masses
	}
	return selected
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
