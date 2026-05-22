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

	"springs/internal/sim"
)

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
	if g.runtime.closed {
		return ebiten.Termination
	}
	g.runtime.inputActive = true
	g.pollMouseControls()
	g.pollSaveFilenameDialogKeyboard()
	g.pollValueDialogKeyboard()
	g.tickValueDialog()
	g.tickNumericTextField()
	if !g.overlays.value.Open && !g.overlays.save.Open {
		g.pollNumericTextFieldKeyboard()
		if g.controls.focusedNumeric == "" {
			g.pollKeyboardControls()
		}
	}
	g.pollDemoPickerScroll()
	if g.runtime.closed {
		return ebiten.Termination
	}
	g.advanceSimulationFrame()
	return nil
}

func (g *Game) pollDemoPickerScroll() {
	if !g.controls.demoPickerOpen {
		return
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		g.scrollDemoPicker(int(-wheelY))
	}
}

func (g *Game) pollMouseControls() {
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	x, y := ebiten.CursorPosition()
	g.pointer.lastCursor = g.screenToWorld(simVec(x, y))
	g.updateSpringChainEnd(g.pointer.lastCursor)
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
	return g.keyboard.shiftDown || anyKeyPressed(shiftKeys, ebiten.IsKeyPressed)
}

func (g *Game) controlKeyPressed() bool {
	return g.keyboard.controlDown || controlKeyPressed()
}

func (g *Game) throwKeyPressed() bool {
	return g.keyboard.throwDown || ebiten.IsKeyPressed(ebiten.KeyT)
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
	if g.editState.selected {
		g.drawSelection(screen)
	}
	g.drawOpenOverlays(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS %.0f", ebiten.ActualTPS()))
}

func (g *Game) drawOpenOverlays(screen *ebiten.Image) {
	if g.controls.demoPickerOpen {
		g.drawDemoPicker(screen)
	}
	if g.overlays.massMenu.Open {
		g.drawMassContextMenu(screen)
	}
	if g.overlays.springMenu.Open {
		g.drawSpringContextMenu(screen)
	}
	if g.overlays.value.Open {
		g.drawValueDialog(screen)
	}
	if g.overlays.save.Open {
		g.drawSaveFilenameDialog(screen)
	}
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
		{x: 0, y: topBarHeight, width: toolbarWidth, height: screenHeight - topBarHeight, color: panelColor},
		{x: screenWidth - inspectorWidth, y: topBarHeight, width: inspectorWidth, height: screenHeight - topBarHeight, color: panelColor},
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
	_, _, minY, maxY := g.canvasWorldBounds()
	top := firstGridCoordinateAtOrAfter(minY, size)
	points := []sim.Vec2{}
	for y := top; y <= maxY; y += size {
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
	for _, spring := range g.world.simulation.Springs {
		a, b, ok := g.springEndpoints(spring)
		if !ok {
			continue
		}
		drawSpringLine(screen, g.worldToScreen(a.Position), g.worldToScreen(b.Position), springDrawColor(spring))
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

func springDrawColor(spring sim.Spring) color.RGBA {
	return drawColorFor(spring.Wall, wallSpringColor, springColor)
}

func springLineAntiAlias() bool {
	return false
}

func (g *Game) pendingSpringLine() (selectionLine, bool) {
	if g.pointer.pendingSpringID == 0 {
		return selectionLine{}, false
	}
	start, ok := g.world.simulation.MassByID(g.pointer.pendingSpringID)
	if !ok {
		return selectionLine{}, false
	}
	return selectionLine{x1: start.Position.X, y1: start.Position.Y, x2: g.pointer.pendingSpringEnd.X, y2: g.pointer.pendingSpringEnd.Y}, true
}

func (g *Game) drawSelectionDrag(screen *ebiten.Image) {
	if !g.pointer.selectionDrag {
		return
	}
	for _, line := range selectionRectangleLines(g.pointer.selectionStart, g.pointer.selectionEnd) {
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
	for _, mass := range g.world.simulation.Masses {
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

func massDrawColor(mass sim.Mass) color.RGBA {
	return drawColorFor(mass.Fixed, fixedMassColor, massColor)
}

func drawColorFor(useAlternate bool, alternate color.RGBA, fallback color.RGBA) color.RGBA {
	if useAlternate {
		return alternate
	}
	return fallback
}

func (g *Game) drawWalls(screen *ebiten.Image) {
	drawWallLine := func(name string, x1, y1, x2, y2 float64) {
		if enabled, _ := g.world.simulation.Parameters.WallEnabled(name); enabled {
			ebitenutil.DrawLine(screen, x1, y1, x2, y2, wallColor)
		}
	}
	for _, line := range wallDrawLines(g.world.simulation.Bounds) {
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
		{name: "top", x1: bounds.MinX(), y1: bounds.MaxY() - 1, x2: bounds.MaxX(), y2: bounds.MaxY() - 1},
		{name: "bottom", x1: bounds.MinX(), y1: bounds.MinY(), x2: bounds.MaxX(), y2: bounds.MinY()},
		{name: "left", x1: bounds.MinX(), y1: bounds.MinY(), x2: bounds.MinX(), y2: bounds.MaxY()},
		{name: "right", x1: bounds.MaxX() - 1, y1: bounds.MinY(), x2: bounds.MaxX() - 1, y2: bounds.MaxY()},
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
		return g.world.simulation.Masses
	}
	return selected
}

func (g *Game) explicitSelectedMasses() []sim.Mass {
	var selected []sim.Mass
	for _, mass := range g.world.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selected = append(selected, mass)
		}
	}
	return selected
}

func (g *Game) allMassesImplicitlySelected() bool {
	return g.editState.selected && len(g.selectedSpringLines()) == 0
}

func (g *Game) selectedSpringLines() []selectionLine {
	var lines []selectionLine
	for _, spring := range g.world.simulation.Springs {
		if !g.editing().SelectedSprings[spring.ID] {
			continue
		}
		a, okA := g.world.simulation.MassByID(spring.MassA)
		b, okB := g.world.simulation.MassByID(spring.MassB)
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:44:11-05:00","module_hash":"a9d5c6d2055bf8e1342bf9b5963a2befddc8a9902e3203d067287ec077bc020a","functions":[{"id":"func/Run","name":"Run","line":18,"end_line":27,"hash":"0f29353f1e39f4004a508c545ca3f9cb53a79f5e92f6650d01ba41b79b91eacd"},{"id":"func/Game.Update","name":"Game.Update","line":29,"end_line":52,"hash":"58127d9ff943570e871db98a359ea20f54db158db255d8cda0cbe3969a094040"},{"id":"func/Game.pollDemoPickerScroll","name":"Game.pollDemoPickerScroll","line":54,"end_line":62,"hash":"c5d2ab4cd58397f0ad21541a06e5a7ba1fc05d7704f91a57ebd1b1ecb47a4761"},{"id":"func/Game.pollMouseControls","name":"Game.pollMouseControls","line":64,"end_line":72,"hash":"b4d05ab96df885ae085425be719ea95354ecf4e8bd333e264f6587b2daf8d049"},{"id":"func/Game.pollKeyboardControls","name":"Game.pollKeyboardControls","line":74,"end_line":79,"hash":"974e7bd1c08ce7c0ff18eaa2303532542899782abc429edc7a23533a29eee794"},{"id":"func/Game.pollEscapeShortcut","name":"Game.pollEscapeShortcut","line":81,"end_line":85,"hash":"5e86654ef5f99460de37fc787add2d7f43966d7c9f4a4df0412a31ade98cdf3d"},{"id":"func/Game.pollControlShortcuts","name":"Game.pollControlShortcuts","line":87,"end_line":89,"hash":"2b58128a3e1003d965513b0da3e231a3783dbe1118ab1cb907dfbced3f534678"},{"id":"func/Game.handlePressedShortcut","name":"Game.handlePressedShortcut","line":91,"end_line":95,"hash":"6ad069ec21c2d4e4496b5a863c312f2882e0a6da6d67606c8c4b4f1d241ba05c"},{"id":"func/pressedControlShortcut","name":"pressedControlShortcut","line":97,"end_line":99,"hash":"c5c29917c6dfe8a97ccc7393c4e23c99392e60e240475f60f3aa9c0fb03c71ec"},{"id":"func/pressedControlShortcutFrom","name":"pressedControlShortcutFrom","line":101,"end_line":106,"hash":"2e9517d5ce38178ce08c01ffb56d86111d420144cd8f683b72f2e25e8401be66"},{"id":"func/firstPressedShortcut","name":"firstPressedShortcut","line":108,"end_line":113,"hash":"45ca55d96c80e333dfef91bdcb0c7475cc8ace90c20baca29774a17ad3e5ca18"},{"id":"func/controlKeyPressed","name":"controlKeyPressed","line":131,"end_line":133,"hash":"c8e3be2e7367aa320eeda237c2acda15d16eb3ea104cdd3453f372bfc829cd29"},{"id":"func/Game.shiftKeyPressed","name":"Game.shiftKeyPressed","line":135,"end_line":137,"hash":"f08288a482f4861405b100bcf78510fde734ae74f8b2843e3bc4a42647a0c510"},{"id":"func/Game.controlKeyPressed","name":"Game.controlKeyPressed","line":139,"end_line":141,"hash":"7926b248863b18906223aa972bc25b76120f608f70794720d95071db157a8c49"},{"id":"func/Game.throwKeyPressed","name":"Game.throwKeyPressed","line":143,"end_line":145,"hash":"6f3f5ea95b3fe27702ef8c3002822bd717099a847f2453b6019d145cd7c32e38"},{"id":"func/anyKeyPressed","name":"anyKeyPressed","line":162,"end_line":169,"hash":"ff2843aefeb8b3b09a8cf5ee32470ec8d0eb3879cb8daeb672430cf8085933b7"},{"id":"func/Game.Draw","name":"Game.Draw","line":171,"end_line":189,"hash":"ca5642e3461b830d8cf9dfc9cf22120dbef9a2226c5a82a63f577dab4fedb4d5"},{"id":"func/Game.drawOpenOverlays","name":"Game.drawOpenOverlays","line":191,"end_line":207,"hash":"e433f2399089b9db87ca66497c6e70c5184bd6f8b34a231702abc295a895ac15"},{"id":"func/Game.drawEditorChrome","name":"Game.drawEditorChrome","line":209,"end_line":213,"hash":"472fbea691e81ccd3a0f783e20fcbd32801d87ee3f99e89f0ca941bc28f56659"},{"id":"func/editorChromeAntiAlias","name":"editorChromeAntiAlias","line":215,"end_line":217,"hash":"660ffdb4c15b9f1e1b71b357db91cc1e3ccf02d18a5b93bd19da971f648614cf"},{"id":"func/editorChromeRects","name":"editorChromeRects","line":227,"end_line":233,"hash":"a0fb22bfa94b0edd57048c739dfa85cdff479df1596114efbe85faf9fb512f60"},{"id":"func/Game.drawGridPoints","name":"Game.drawGridPoints","line":235,"end_line":239,"hash":"62ca4d64fc25bc322d533f258709357d5a112e79edb94e8d0332836a21cedcf7"},{"id":"func/Game.gridPointRects","name":"Game.gridPointRects","line":247,"end_line":261,"hash":"600750169ed58b07bd434dafed0f8a6fd143d81bade263f108fed60c984f43ee"},{"id":"func/gridPointPixelSize","name":"gridPointPixelSize","line":263,"end_line":265,"hash":"6c0a1cf159cdd1cb2dcab7eaf4f8d631777564f781ac420f6d26dab7a0fe6730"},{"id":"func/gridPointAntiAlias","name":"gridPointAntiAlias","line":267,"end_line":269,"hash":"9da19824153c650376ff119c27f1fb0c1582d8b0411321824a1d13b1e594d627"},{"id":"func/Game.gridPoints","name":"Game.gridPoints","line":271,"end_line":287,"hash":"cbd43d3678d2dc7775644dbb6ae63946ed4ef3662f70a838663d69ab3e68fc79"},{"id":"func/validGridSnapSize","name":"validGridSnapSize","line":289,"end_line":291,"hash":"cf2cc0632f5d59e877c981dca062e63dac4c535b77c3065878c5c3a34b22b37e"},{"id":"func/firstGridCoordinateAtOrAfter","name":"firstGridCoordinateAtOrAfter","line":293,"end_line":295,"hash":"8a7c004c8f82076f9188bb1776b097331b802ef718b63f68bdfd2eae4acf9ec7"},{"id":"func/Game.drawSprings","name":"Game.drawSprings","line":297,"end_line":305,"hash":"0f30062b5c677199ecd71d4bf8bf34081fe7747f4ddf51caa2c60ddc88183b8f"},{"id":"func/Game.drawPendingSpring","name":"Game.drawPendingSpring","line":307,"end_line":313,"hash":"8b2f10d92cea34baa15527fdf72126c1b61468c391cf7a633c81b6d273286aa9"},{"id":"func/drawSpringLine","name":"drawSpringLine","line":315,"end_line":317,"hash":"bd344626244ed0ae0211b6aadf761f458422111204fb50d9f061d4380b2d8113"},{"id":"func/springDrawColor","name":"springDrawColor","line":319,"end_line":321,"hash":"c798ea19eb55f4df10bec6662ff88b167e98c1ab0c1cda2f55aab678a8060bc3"},{"id":"func/springLineAntiAlias","name":"springLineAntiAlias","line":323,"end_line":325,"hash":"c16077efe2092cbf3ba89b04223ef9f600017a1311b6169e39543d7b88ac6d16"},{"id":"func/Game.pendingSpringLine","name":"Game.pendingSpringLine","line":327,"end_line":336,"hash":"35e4cb796d7894e574bf1abef437282035d185710dc5de2fb2c700011ce01b12"},{"id":"func/Game.drawSelectionDrag","name":"Game.drawSelectionDrag","line":338,"end_line":345,"hash":"494c4d595213ad71e22bb4f1c642635f31b2b9953a3f5fe26134494c5b0beb87"},{"id":"func/selectionRectangleLines","name":"selectionRectangleLines","line":347,"end_line":358,"hash":"ac6138efe5e031f4de8e884ce3557eafe1b6b8ae5f95fba79c0fd3846af14d8e"},{"id":"func/Game.drawMasses","name":"Game.drawMasses","line":360,"end_line":368,"hash":"c4906e163871286f299c404c15d97c6b4bc1f73a62d16693b9c7f12c00d44698"},{"id":"func/massDrawAntiAlias","name":"massDrawAntiAlias","line":370,"end_line":372,"hash":"948e921ea933de29f25b79483885ab252d33194f86069a946d34f1508816a3c9"},{"id":"func/massDrawColor","name":"massDrawColor","line":374,"end_line":376,"hash":"890b5d5de224ff327a7f9612820334a87d1d55c98a87946eedb0c9a12f73ee25"},{"id":"func/drawColorFor","name":"drawColorFor","line":378,"end_line":383,"hash":"deb424640c409f469cf9100cd980a426225e594d77029431530fc96e63cb5655"},{"id":"func/Game.drawWalls","name":"Game.drawWalls","line":385,"end_line":396,"hash":"d5ea45a7d1f229782533e6766282fd1bde54580ce6d1a28ea6d0cd7ab8b385bd"},{"id":"func/wallDrawLines","name":"wallDrawLines","line":403,"end_line":410,"hash":"3f62a9ee6f8fb680fb031e79c6de06472479fab01a2b74a8a18731ff5fd99564"},{"id":"func/Game.drawSelection","name":"Game.drawSelection","line":412,"end_line":419,"hash":"837fdf71a9d781c78495e22db390fe530af2786a8feacb4628b61fd157af5d53"},{"id":"func/Game.drawSelectionLine","name":"Game.drawSelectionLine","line":421,"end_line":425,"hash":"a0630296c8ec7d7e6289543381e43546b4c63a7cc91a720d920bde7e25b14c92"},{"id":"func/Game.selectedMasses","name":"Game.selectedMasses","line":427,"end_line":433,"hash":"425a668feb687968b0a698c5a1cb4b9212c6e747cc061cb6e2765985b487b04e"},{"id":"func/Game.explicitSelectedMasses","name":"Game.explicitSelectedMasses","line":435,"end_line":443,"hash":"ce61a71692909056f347ee9ca4e118c7dc813219d96d29060735f2cc717bba35"},{"id":"func/Game.allMassesImplicitlySelected","name":"Game.allMassesImplicitlySelected","line":445,"end_line":447,"hash":"6d552ca010268d6627068ed4fd689846bdbce4a7b74ca907429d953ec7a7db9a"},{"id":"func/Game.selectedSpringLines","name":"Game.selectedSpringLines","line":449,"end_line":462,"hash":"0e630d07b96362560040d393c284e777cbe0d79edb7e05fa929db378c463fd1d"},{"id":"func/selectedMassOutline","name":"selectedMassOutline","line":471,"end_line":477,"hash":"09e522ba39856beab87cf554a1a6ce1af7c4466412f407118cc223529ae81da3"},{"id":"func/selectionOutline","name":"selectionOutline","line":479,"end_line":489,"hash":"d4ec3f8743e56bd8e3bed0d58f60210a5195c24b29e99c152f2e5b09b8806bdf"}]}
// mutate4go-manifest-end
