package app

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"springs/internal/edit"
	"springs/internal/sim"
)

const (
	screenWidth    = 1920
	screenHeight   = 1080
	toolbarWidth   = 72
	inspectorWidth = 360
	topBarHeight   = 36
	statusHeight   = 24
	baseFrameTime  = 1.0 / 60.0
	maxSpeed       = 4.0
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
)

type Game struct {
	simulation       *sim.Simulation
	initialState     *sim.Simulation
	savedState       *sim.Simulation
	mode             string
	selected         bool
	dirty            bool
	lastCommand      string
	pathEntryCommand string
	paused           bool
	mousePressed     bool
	draggingMassID   int
	pendingSpringID  int
	pendingSpringEnd sim.Vec2
	activeSlider     string
	demoPickerOpen   bool
	demoPickerScroll int
	demoFiles        []string
	simulationSpeed  float64
	editor           *edit.Editor
	inputActive      bool
	renderingActive  bool
	closed           bool
}

type WindowConfig struct {
	Width     int
	Height    int
	Title     string
	Resizable bool
}

func NewGame() *Game {
	world := newDefaultStartupWorld()
	return &Game{simulation: world, initialState: world.Clone(), mode: "select", simulationSpeed: 1, editor: edit.NewEditor(world)}
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
	return ebiten.RunGame(NewGame())
}

func (g *Game) Update() error {
	if g.closed {
		return ebiten.Termination
	}
	g.inputActive = true
	g.pollMouseControls()
	if g.demoPickerOpen {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 {
			g.scrollDemoPicker(int(-wheelY))
		}
	}
	if g.closed {
		return ebiten.Termination
	}
	if !g.paused && g.simulationSpeed > 0 {
		g.simulation.AdvanceDuration(baseFrameTime * g.simulationSpeed)
	}
	return nil
}

func (g *Game) pollMouseControls() {
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	x, y := ebiten.CursorPosition()
	g.handlePointer(pressed, x, y)
}

func (g *Game) handlePointer(pressed bool, x int, y int) {
	position := sim.Vec2{X: float64(x), Y: float64(y)}
	if pressed && !g.mousePressed {
		if g.demoPickerOpen {
			g.clickDemoPicker(x, y)
			g.mousePressed = pressed
			return
		}
		if !g.ClickAt(x, y) {
			g.beginWorldPointer(position)
		}
	} else if pressed && g.draggingMassID != 0 {
		g.DragMass(g.draggingMassID, position)
	} else if pressed && g.pendingSpringID != 0 {
		g.pendingSpringEnd = position
	} else if pressed && g.activeSlider != "" {
		g.setSliderAt(g.activeSlider, x)
	} else if !pressed {
		g.finishWorldPointer(position)
		g.draggingMassID = 0
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

func (g *Game) beginWorldPointer(position sim.Vec2) {
	switch g.mode {
	case "add mass":
		g.createMassAt(position)
	case "select":
		g.selectNearest(position)
	case "add spring":
		g.beginSpringAt(position)
	case "drag":
		g.beginMassDrag(position)
	}
}

func (g *Game) finishWorldPointer(position sim.Vec2) {
	if g.mode == "add spring" && g.pendingSpringID != 0 {
		g.finishSpringAt(position)
	}
}

func (g *Game) createMassAt(position sim.Vec2) {
	editor := g.editing()
	editor.Mode = edit.ModeAddMass
	editor.GridSnapEnabled = g.gridSnapEnabled()
	editor.GridSnapSize = g.gridSnapSize()
	if _, err := editor.Click(position); err == nil {
		g.dirty = true
	}
}

func (g *Game) selectNearest(position sim.Vec2) {
	if err := g.editing().SelectNearest(position, false); err == nil {
		g.selected = true
	}
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
	if g.mode != "drag" {
		return
	}
	id, ok := g.massAt(position)
	if !ok {
		return
	}
	g.draggingMassID = id
	g.DragMass(id, position)
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
	if result.SpringLinesVisible {
		g.drawSprings(screen)
	}
	g.drawPendingSpring(screen)
	g.drawMasses(screen)
	g.drawWalls(screen)
	if g.selected {
		g.drawSelection(screen)
	}
	if g.demoPickerOpen {
		g.drawDemoPicker(screen)
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

func (g *Game) drawSprings(screen *ebiten.Image) {
	for _, spring := range g.simulation.Springs {
		if !g.validSpring(spring) {
			continue
		}
		a := g.simulation.Masses[spring.A].Position
		b := g.simulation.Masses[spring.B].Position
		ebitenutil.DrawLine(screen, a.X, a.Y, b.X, b.Y, springColor)
	}
}

func (g *Game) drawPendingSpring(screen *ebiten.Image) {
	line, ok := g.pendingSpringLine()
	if !ok {
		return
	}
	ebitenutil.DrawLine(screen, line.x1, line.y1, line.x2, line.y2, springColor)
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

func (g *Game) drawMasses(screen *ebiten.Image) {
	for _, mass := range g.simulation.Masses {
		x, y, radius := massDrawCircle(mass)
		vector.DrawFilledCircle(screen, x, y, radius, massDrawColor(mass), massDrawAntiAlias())
	}
}

func massDrawAntiAlias() bool {
	return true
}

func massDrawCircle(mass sim.Mass) (float32, float32, float32) {
	return float32(mass.Position.X), float32(mass.Position.Y), 5
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
		drawWallLine(line.name, line.x1, line.y1, line.x2, line.y2)
	}
}

type wallDrawLine struct {
	name           string
	x1, y1, x2, y2 float64
}

func wallDrawLines(bounds sim.Bounds) []wallDrawLine {
	return []wallDrawLine{
		{name: "top", x1: 0, y1: 0, x2: bounds.Width, y2: 0},
		{name: "bottom", x1: 0, y1: bounds.Height - 1, x2: bounds.Width, y2: bounds.Height - 1},
		{name: "left", x1: 0, y1: 0, x2: 0, y2: bounds.Height},
		{name: "right", x1: bounds.Width - 1, y1: 0, x2: bounds.Width - 1, y2: bounds.Height},
	}
}

func (g *Game) drawSelection(screen *ebiten.Image) {
	for _, line := range selectedMassOutline(g.selectedMasses()) {
		ebitenutil.DrawLine(screen, line.x1, line.y1, line.x2, line.y2, selectionColor)
	}
}

func (g *Game) selectedMasses() []sim.Mass {
	var selected []sim.Mass
	for _, mass := range g.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selected = append(selected, mass)
		}
	}
	if len(selected) == 0 && g.selected {
		return g.simulation.Masses
	}
	return selected
}

type selectionLine struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

func selectedMassOutline(masses []sim.Mass) []selectionLine {
	if len(masses) == 0 {
		return nil
	}
	return selectionOutline(masses[0])
}

func selectionOutline(mass sim.Mass) []selectionLine {
	x := mass.Position.X
	y := mass.Position.Y
	return []selectionLine{
		{x - 8, y - 8, x + 8, y - 8},
		{x + 8, y - 8, x + 8, y + 8},
		{x + 8, y + 8, x - 8, y + 8},
		{x - 8, y + 8, x - 8, y - 8},
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
