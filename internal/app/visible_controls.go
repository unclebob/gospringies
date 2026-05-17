package app

import (
	"fmt"
	"image"
	"image/color"
	"strings"

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
	for _, section := range inspectorSections() {
		drawLabeledRect(screen, section.Rect, sectionColor, section.Label)
	}
	for _, field := range g.statusFields() {
		drawLabeledRect(screen, field.Rect, controlColor, field.Label)
	}
}

func (g *Game) drawControl(screen *ebiten.Image, control controlBox) {
	fill := controlColor
	if g.activeControl(control.Name) {
		fill = activeControlColor
	}
	drawLabeledRect(screen, control.Rect, fill, control.Label)
}

func drawLabeledRect(screen *ebiten.Image, rect image.Rectangle, fill color.RGBA, label string) {
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), fill, false)
	ebitenutil.DebugPrintAt(screen, label, rect.Min.X+4, rect.Min.Y+4)
}

func (g *Game) activeControl(name string) bool {
	return (name == "select mode" && g.mode == "select") ||
		(name == "mass mode" && g.mode == "add mass") ||
		(name == "spring mode" && g.mode == "add spring") ||
		(name == "drag mode" && g.mode == "drag") ||
		(name == "pause command" && g.paused) ||
		(name == "run command" && !g.paused)
}

func visibleControls() []controlBox {
	return append(toolbarControls(), commandControls()...)
}

func toolbarControls() []controlBox {
	return []controlBox{
		{Name: "select mode", Label: "Select", Region: "left toolbar", Rect: image.Rect(8, 48, 64, 68)},
		{Name: "mass mode", Label: "Mass", Region: "left toolbar", Rect: image.Rect(8, 78, 64, 98)},
		{Name: "spring mode", Label: "Spring", Region: "left toolbar", Rect: image.Rect(8, 108, 64, 128)},
		{Name: "drag mode", Label: "Drag", Region: "left toolbar", Rect: image.Rect(8, 138, 64, 158)},
	}
}

func commandControls() []controlBox {
	return []controlBox{
		{Name: "run command", Label: "Run", Region: "top command bar", Rect: image.Rect(82, 8, 122, 28)},
		{Name: "pause command", Label: "Pause", Region: "top command bar", Rect: image.Rect(128, 8, 180, 28)},
		{Name: "reset command", Label: "Reset", Region: "top command bar", Rect: image.Rect(186, 8, 238, 28)},
		{Name: "load command", Label: "Load", Region: "top command bar", Rect: image.Rect(244, 8, 292, 28)},
		{Name: "insert command", Label: "Insert", Region: "top command bar", Rect: image.Rect(298, 8, 362, 28)},
		{Name: "save command", Label: "Save", Region: "top command bar", Rect: image.Rect(368, 8, 416, 28)},
		{Name: "quit command", Label: "Quit", Region: "top command bar", Rect: image.Rect(422, 8, 470, 28)},
	}
}

func inspectorSections() []controlBox {
	return []controlBox{
		{Label: "Mass", Region: "right inspector", Rect: image.Rect(522, 48, 632, 68)},
		{Label: "Spring", Region: "right inspector", Rect: image.Rect(522, 82, 632, 102)},
		{Label: "Forces", Region: "right inspector", Rect: image.Rect(522, 116, 632, 136)},
		{Label: "Walls", Region: "right inspector", Rect: image.Rect(522, 150, 632, 170)},
		{Label: "Simulation", Region: "right inspector", Rect: image.Rect(522, 184, 632, 204)},
	}
}

func (g *Game) statusFields() []statusField {
	return []statusField{
		{Name: "mode", Label: strings.Title(g.mode) + " mode", Rect: image.Rect(8, 458, 100, 478)},
		{Name: "run state", Label: g.simulationState(), Rect: image.Rect(108, 458, 196, 478)},
		{Name: "object counts", Label: fmt.Sprintf("object counts: %dM %dS", len(g.simulation.Masses), len(g.simulation.Springs)), Rect: image.Rect(204, 458, 360, 478)},
		{Name: "file state", Label: g.fileState(), Rect: image.Rect(368, 458, 462, 478)},
	}
}

func (g *Game) DrawFrameReport() DrawFrameReport {
	return analyzeDrawnFrame(g)
}

func analyzeDrawnFrame(game *Game) DrawFrameReport {
	report := DrawFrameReport{
		RegionPixels:        map[string]int{},
		Controls:            visibleControlLabels(),
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
	for _, control := range visibleControls() {
		if !labelFits(control.Label, control.Rect) {
			return false
		}
	}
	for _, section := range inspectorSections() {
		if !labelFits(section.Label, section.Rect) {
			return false
		}
	}
	for _, field := range game.statusFields() {
		if !labelFits(field.Label, field.Rect) {
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
		"canvas":          image.Rect(72, 36, screenWidth-128, screenHeight-24),
		"left toolbar":    image.Rect(0, 36, 72, screenHeight-24),
		"top command bar": image.Rect(72, 0, screenWidth-128, 36),
		"right inspector": image.Rect(screenWidth-128, 36, screenWidth, screenHeight-24),
		"status line":     image.Rect(0, screenHeight-24, screenWidth, screenHeight),
	}
}

func (g *Game) visibleRegionPixels(region string) int {
	count := 0
	for _, control := range visibleControls() {
		if control.Region == region {
			count += control.Rect.Dx() * control.Rect.Dy()
		}
	}
	for _, section := range inspectorSections() {
		if section.Region == region {
			count += section.Rect.Dx() * section.Rect.Dy()
		}
	}
	if region == "status line" {
		for _, field := range g.statusFields() {
			count += field.Rect.Dx() * field.Rect.Dy()
		}
	}
	if count == 0 && region == "canvas" {
		return visibleRegionRects()[region].Dx() * visibleRegionRects()[region].Dy()
	}
	return count
}

func visibleWorldPixels(game *Game) int {
	canvas := visibleRegionRects()["canvas"]
	count := 0
	for _, mass := range game.simulation.Masses {
		point := image.Pt(int(mass.Position.X), int(mass.Position.Y))
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
