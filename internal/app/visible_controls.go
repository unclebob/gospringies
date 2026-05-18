package app

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	debugGlyphWidth  = 6
	debugGlyphHeight = 16
)

var (
	controlColor       = color.RGBA{R: 74, G: 88, B: 108, A: 255}
	activeControlColor = color.RGBA{R: 96, G: 132, B: 166, A: 255}
	sectionColor       = color.RGBA{R: 58, G: 70, B: 86, A: 255}
)

type controlBox struct {
	Name   string
	Label  string
	Region string
	Rect   image.Rectangle
}

type statusField struct {
	Name  string
	Label string
	Rect  image.Rectangle
}

type DrawFrameReport struct {
	RegionPixels        map[string]int
	Controls            map[string]string
	ActiveControls      map[string]bool
	InspectorSections   map[string]bool
	StatusFields        map[string]string
	CanvasWorldPixels   int
	ControlLabelsFit    bool
	RegionControlCounts map[string]int
}

func (g *Game) drawVisibleControls(screen *ebiten.Image) {
	for _, control := range visibleControls() {
		g.drawControl(screen, control)
	}
	for _, control := range g.editMenuControls() {
		g.drawControl(screen, control)
	}
	for _, section := range inspectorSections() {
		drawLabeledRect(screen, section.Rect, sectionColor, section.Label)
	}
	for _, field := range g.statusFields() {
		drawLabeledRect(screen, field.Rect, controlColor, field.Label)
	}
}

func (g *Game) drawControl(screen *ebiten.Image, control controlBox) {
	if isSliderControl(control.Name) {
		g.drawSlider(screen, control)
		return
	}
	fill := controlColor
	if g.activeControl(control.Name) {
		fill = activeControlColor
	}
	drawLabeledRect(screen, control.Rect, fill, control.Label)
}

func (g *Game) drawSlider(screen *ebiten.Image, control controlBox) {
	drawLabeledRect(screen, control.Rect, controlColor, g.sliderLabel(control))
	track := sliderTrack(control)
	vector.DrawFilledRect(screen, float32(track.Min.X), float32(track.Min.Y), float32(track.Dx()), float32(track.Dy()), sectionColor, false)
	fill := track
	fill.Max.X = track.Min.X + int(g.sliderFraction(control.Name)*float64(track.Dx()))
	vector.DrawFilledRect(screen, float32(fill.Min.X), float32(fill.Min.Y), float32(fill.Dx()), float32(fill.Dy()), activeControlColor, false)
}

func drawLabeledRect(screen *ebiten.Image, rect image.Rectangle, fill color.RGBA, label string) {
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), fill, false)
	ebitenutil.DebugPrintAt(screen, label, rect.Min.X+4, rect.Min.Y+4)
}

func isSliderControl(name string) bool {
	return name == "gravity slider" || name == "speed slider" || name == "viscosity slider"
}

func sliderTrack(control controlBox) image.Rectangle {
	return image.Rect(control.Rect.Min.X+80, control.Rect.Min.Y+13, control.Rect.Max.X-10, control.Rect.Min.Y+17)
}

func (g *Game) sliderFraction(name string) float64 {
	switch name {
	case "gravity slider":
		force, _ := g.simulation.Parameters.Force("gravity")
		return clampFloat(forceValueFloat(force, "magnitude")/50, 0, 1)
	case "speed slider":
		return clampFloat(g.simulationSpeed/maxSpeed, 0, 1)
	case "viscosity slider":
		return clampFloat(g.parameterFloat("viscosity")/2, 0, 1)
	default:
		return 0
	}
}

func (g *Game) sliderLabel(control controlBox) string {
	switch control.Name {
	case "gravity slider":
		force, _ := g.simulation.Parameters.Force("gravity")
		return fmt.Sprintf("%s %s", control.Label, formatControlFloat(forceValueFloat(force, "magnitude")))
	case "speed slider":
		return fmt.Sprintf("%s %sx", control.Label, formatControlFloat(g.simulationSpeed))
	case "viscosity slider":
		return fmt.Sprintf("%s %s", control.Label, formatControlFloat(g.parameterFloat("viscosity")))
	default:
		return control.Label
	}
}

func (g *Game) activeControl(name string) bool {
	return g.activeRunControl(name) ||
		g.activeForceControl(name) ||
		g.activeParameterControl(name) ||
		g.activeWallControl(name)
}

func (g *Game) activeRunControl(name string) bool {
	switch name {
	case "pause command":
		return g.paused
	case "run command":
		return !g.paused
	default:
		return false
	}
}

func (g *Game) activeForceControl(name string) bool {
	switch name {
	case "gravity force":
		return g.forceEnabled("gravity")
	case "center attraction force":
		return g.forceEnabled("center attraction")
	case "center mass force":
		return g.forceEnabled("center of mass attraction")
	case "wall repulsion force":
		return g.forceEnabled("wall repulsion")
	case "mass collision force":
		return g.forceEnabled("mass collision")
	default:
		return false
	}
}

func (g *Game) activeParameterControl(name string) bool {
	switch name {
	case "fixed mass toggle":
		return g.parameterEnabled("fixed mass")
	case "show springs toggle":
		return g.parameterEnabled("show springs")
	case "adaptive timestep toggle":
		return g.parameterEnabled("adaptive timestep")
	case "grid snap toggle":
		return g.gridSnapEnabled()
	default:
		return false
	}
}

func (g *Game) activeWallControl(name string) bool {
	switch name {
	case "top wall toggle":
		return g.wallEnabled("top")
	case "left wall toggle":
		return g.wallEnabled("left")
	case "right wall toggle":
		return g.wallEnabled("right")
	case "bottom wall toggle":
		return g.wallEnabled("bottom")
	default:
		return false
	}
}

func visibleControls() []controlBox {
	controls := append(menuControls(), toolbarControls()...)
	controls = append(controls, commandControls()...)
	return append(controls, inspectorControls()...)
}

func menuControls() []controlBox {
	return []controlBox{
		{Name: "edit menu", Label: "Edit", Region: "top command bar", Rect: image.Rect(8, 8, 52, 28)},
	}
}

func (g *Game) editMenuControls() []controlBox {
	if !g.editMenuOpen {
		return nil
	}
	return []controlBox{
		{Name: "cut command", Label: "Cut     Ctrl+X", Region: "top command bar", Rect: image.Rect(8, 30, 132, 54)},
		{Name: "copy command", Label: "Copy    Ctrl+C", Region: "top command bar", Rect: image.Rect(8, 54, 132, 78)},
		{Name: "paste command", Label: "Paste   Ctrl+V", Region: "top command bar", Rect: image.Rect(8, 78, 132, 102)},
	}
}

func toolbarControls() []controlBox {
	return []controlBox{
		{Name: "select all command", Label: "All", Region: "left toolbar", Rect: image.Rect(8, 48, 64, 68)},
		{Name: "duplicate command", Label: "Dup", Region: "left toolbar", Rect: image.Rect(8, 78, 64, 98)},
		{Name: "delete command", Label: "Del", Region: "left toolbar", Rect: image.Rect(8, 108, 64, 128)},
	}
}

func commandControls() []controlBox {
	return []controlBox{
		{Name: "run command", Label: "Run", Region: "top command bar", Rect: image.Rect(76, 8, 110, 28)},
		{Name: "pause command", Label: "Pause", Region: "top command bar", Rect: image.Rect(112, 8, 158, 28)},
		{Name: "reset command", Label: "Reset", Region: "top command bar", Rect: image.Rect(160, 8, 206, 28)},
		{Name: "save state command", Label: "State+", Region: "top command bar", Rect: image.Rect(208, 8, 260, 28)},
		{Name: "restore state command", Label: "State", Region: "top command bar", Rect: image.Rect(262, 8, 310, 28)},
		{Name: "load command", Label: "Load", Region: "top command bar", Rect: image.Rect(312, 8, 354, 28)},
		{Name: "insert command", Label: "Insert", Region: "top command bar", Rect: image.Rect(356, 8, 408, 28)},
		{Name: "save command", Label: "Save", Region: "top command bar", Rect: image.Rect(410, 8, 452, 28)},
		{Name: "quit command", Label: "Quit", Region: "top command bar", Rect: image.Rect(454, 8, 496, 28)},
	}
}

func inspectorControls() []controlBox {
	x := inspectorLeft() + 16
	right := screenWidth - 16
	half := (right - x - 8) / 2
	return []controlBox{
		{Name: "mass parameter", Label: "Mass", Region: "right inspector", Rect: image.Rect(x, 68, x+half, 88)},
		{Name: "elasticity parameter", Label: "Elas", Region: "right inspector", Rect: image.Rect(x+half+8, 68, right, 88)},
		{Name: "fixed mass toggle", Label: "Fixed", Region: "right inspector", Rect: image.Rect(x, 94, right, 114)},
		{Name: "kspring parameter", Label: "Kspr", Region: "right inspector", Rect: image.Rect(x, 148, x+half, 168)},
		{Name: "kdamp parameter", Label: "Kdmp", Region: "right inspector", Rect: image.Rect(x+half+8, 148, right, 168)},
		{Name: "set rest length command", Label: "RestLen", Region: "right inspector", Rect: image.Rect(x, 174, right, 194)},
		{Name: "gravity force", Label: "Gravity", Region: "right inspector", Rect: image.Rect(x, 230, right, 250)},
		{Name: "gravity slider", Label: "Gravity", Region: "right inspector", Rect: image.Rect(x, 256, right, 286)},
		{Name: "center attraction force", Label: "Center", Region: "right inspector", Rect: image.Rect(x, 292, x+half, 312)},
		{Name: "center mass force", Label: "CMass", Region: "right inspector", Rect: image.Rect(x+half+8, 292, right, 312)},
		{Name: "wall repulsion force", Label: "WallRep", Region: "right inspector", Rect: image.Rect(x, 318, x+half, 338)},
		{Name: "mass collision force", Label: "Collide", Region: "right inspector", Rect: image.Rect(x+half+8, 318, right, 338)},
		{Name: "set center command", Label: "SetCtr", Region: "right inspector", Rect: image.Rect(x, 344, right, 364)},
		{Name: "top wall toggle", Label: "Top", Region: "right inspector", Rect: image.Rect(x, 414, x+half, 434)},
		{Name: "bottom wall toggle", Label: "Bot", Region: "right inspector", Rect: image.Rect(x+half+8, 414, right, 434)},
		{Name: "left wall toggle", Label: "Left", Region: "right inspector", Rect: image.Rect(x, 440, x+half, 460)},
		{Name: "right wall toggle", Label: "Right", Region: "right inspector", Rect: image.Rect(x+half+8, 440, right, 460)},
		{Name: "grid snap toggle", Label: "Grid", Region: "right inspector", Rect: image.Rect(x, 514, x+half, 534)},
		{Name: "show springs toggle", Label: "Springs", Region: "right inspector", Rect: image.Rect(x+half+8, 514, right, 534)},
		{Name: "viscosity slider", Label: "Viscosity", Region: "right inspector", Rect: image.Rect(x, 540, right, 570)},
		{Name: "stickiness parameter", Label: "Stick", Region: "right inspector", Rect: image.Rect(x, 576, right, 596)},
		{Name: "speed slider", Label: "Speed", Region: "right inspector", Rect: image.Rect(x, 602, right, 632)},
		{Name: "timestep parameter", Label: "Step", Region: "right inspector", Rect: image.Rect(x, 638, x+half, 658)},
		{Name: "precision parameter", Label: "Prec", Region: "right inspector", Rect: image.Rect(x+half+8, 638, right, 658)},
		{Name: "adaptive timestep toggle", Label: "Adapt", Region: "right inspector", Rect: image.Rect(x, 664, right, 684)},
	}
}

func inspectorSections() []controlBox {
	x := inspectorLeft() + 16
	right := screenWidth - 16
	return []controlBox{
		{Label: "Mass", Region: "right inspector", Rect: image.Rect(x, 44, right, 64)},
		{Label: "Spring", Region: "right inspector", Rect: image.Rect(x, 122, right, 142)},
		{Label: "Forces", Region: "right inspector", Rect: image.Rect(x, 204, right, 224)},
		{Label: "Walls", Region: "right inspector", Rect: image.Rect(x, 386, right, 408)},
		{Label: "Simulation", Region: "right inspector", Rect: image.Rect(x, 486, right, 508)},
	}
}

func inspectorLeft() int {
	return screenWidth - inspectorWidth
}

func (g *Game) forceEnabled(name string) bool {
	force, ok := g.simulation.Parameters.Force(name)
	return ok && force.Enabled == "true"
}

func (g *Game) parameterEnabled(name string) bool {
	return g.simulation.Parameters.Value(name) == "true"
}

func (g *Game) wallEnabled(name string) bool {
	enabled, _ := g.simulation.Parameters.WallEnabled(name)
	return enabled
}

func (g *Game) gridSnapEnabled() bool {
	return g.gridSnapSize() > 0
}

func (g *Game) statusFields() []statusField {
	return []statusField{
		{Name: "run state", Label: g.simulationState(), Rect: image.Rect(8, screenHeight-22, 104, screenHeight-2)},
		{Name: "object counts", Label: "object counts", Rect: image.Rect(112, screenHeight-22, 212, screenHeight-2)},
		{Name: "selected object count", Label: fmt.Sprintf("%d sel", g.selectedObjectCount()), Rect: image.Rect(220, screenHeight-22, 290, screenHeight-2)},
		{Name: "current file", Label: g.pathEntryCommand, Rect: image.Rect(298, screenHeight-22, 412, screenHeight-2)},
		{Name: "dirty state", Label: g.fileState(), Rect: image.Rect(420, screenHeight-22, 512, screenHeight-2)},
		{Name: "file state", Label: g.fileState(), Rect: image.Rect(520, screenHeight-22, 612, screenHeight-2)},
		{Name: "last error", Label: "", Rect: image.Rect(620, screenHeight-22, 872, screenHeight-2)},
	}
}

func (g *Game) selectedObjectCount() int {
	count := 0
	for _, selected := range g.editing().SelectedMasses {
		if selected {
			count++
		}
	}
	for _, selected := range g.editing().SelectedSprings {
		if selected {
			count++
		}
	}
	return count
}

func (g *Game) DrawFrameReport() DrawFrameReport {
	return analyzeDrawnFrame(g)
}

func analyzeDrawnFrame(game *Game) DrawFrameReport {
	report := DrawFrameReport{
		RegionPixels:        map[string]int{},
		Controls:            visibleControlLabels(),
		ActiveControls:      game.visibleActiveControls(),
		InspectorSections:   visibleInspectorSections(),
		StatusFields:        game.visibleStatusFields(),
		RegionControlCounts: game.visibleRegionControlCounts(),
		ControlLabelsFit:    visibleLabelsFit(game),
	}
	for name := range visibleRegionRects() {
		report.RegionPixels[name] = game.visibleRegionPixels(name)
	}
	report.CanvasWorldPixels = visibleWorldPixels(game)
	return report
}

func (g *Game) visibleActiveControls() map[string]bool {
	active := map[string]bool{}
	for _, control := range visibleControls() {
		isActive := g.activeControl(control.Name)
		active[control.Name] = isActive
		active[control.Label] = active[control.Label] || isActive
	}
	return active
}

func visibleControlLabels() map[string]string {
	labels := map[string]string{}
	for _, control := range visibleControls() {
		labels[control.Name] = control.Label
	}
	return labels
}

func visibleInspectorSections() map[string]bool {
	sections := map[string]bool{}
	for _, section := range inspectorSections() {
		sections[section.Label] = true
	}
	return sections
}

func (g *Game) visibleStatusFields() map[string]string {
	fields := map[string]string{}
	for _, field := range g.statusFields() {
		fields[field.Name] = field.Label
	}
	return fields
}

func (g *Game) visibleRegionControlCounts() map[string]int {
	counts := map[string]int{}
	for _, control := range visibleControls() {
		counts[control.Region]++
	}
	for _, section := range inspectorSections() {
		counts[section.Region]++
	}
	counts["status line"] = len(g.statusFields())
	return counts
}

func visibleLabelsFit(game *Game) bool {
	return controlLabelsFit(visibleControls()) &&
		controlLabelsFit(inspectorSections()) &&
		statusLabelsFit(game.statusFields())
}

func controlLabelsFit(boxes []controlBox) bool {
	return labelsFitItems(boxes, func(box controlBox) (string, image.Rectangle) { return box.Label, box.Rect })
}

func statusLabelsFit(fields []statusField) bool {
	return labelsFitItems(fields, func(field statusField) (string, image.Rectangle) { return field.Label, field.Rect })
}

func labelsFitItems[T any](items []T, labelAndRect func(T) (string, image.Rectangle)) bool {
	for _, item := range items {
		label, rect := labelAndRect(item)
		if !labelFits(label, rect) {
			return false
		}
	}
	return true
}

func labelFits(label string, rect image.Rectangle) bool {
	return len(label)*debugGlyphWidth <= rect.Dx()-8 && debugGlyphHeight <= rect.Dy()
}

func visibleRegionRects() map[string]image.Rectangle {
	return map[string]image.Rectangle{
		"canvas":          image.Rect(toolbarWidth, topBarHeight, screenWidth-inspectorWidth, screenHeight-statusHeight),
		"left toolbar":    image.Rect(0, topBarHeight, toolbarWidth, screenHeight-statusHeight),
		"top command bar": image.Rect(toolbarWidth, 0, screenWidth-inspectorWidth, topBarHeight),
		"right inspector": image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight-statusHeight),
		"status line":     image.Rect(0, screenHeight-statusHeight, screenWidth, screenHeight),
	}
}

func (g *Game) visibleRegionPixels(region string) int {
	count := regionControlPixels(visibleControls(), region) +
		regionControlPixels(inspectorSections(), region) +
		g.regionStatusPixels(region)
	if count == 0 && region == "canvas" {
		return rectPixels(visibleRegionRects()[region])
	}
	return count
}

func regionControlPixels(controls []controlBox, region string) int {
	count := 0
	for _, control := range controls {
		if control.Region == region {
			count += rectPixels(control.Rect)
		}
	}
	return count
}

func (g *Game) regionStatusPixels(region string) int {
	if region != "status line" {
		return 0
	}
	count := 0
	for _, field := range g.statusFields() {
		count += rectPixels(field.Rect)
	}
	return count
}

func rectPixels(rect image.Rectangle) int {
	return rect.Dx() * rect.Dy()
}

func visibleWorldPixels(game *Game) int {
	canvas := visibleRegionRects()["canvas"]
	count := 0
	for _, mass := range game.simulation.Masses {
		screenPosition := game.worldToScreen(mass.Position)
		point := image.Pt(int(screenPosition.X), int(screenPosition.Y))
		if point.In(canvas) {
			count += 25
		}
	}
	for _, spring := range game.simulation.Springs {
		if game.validSpring(spring) {
			count += 50
		}
	}
	return count
}
