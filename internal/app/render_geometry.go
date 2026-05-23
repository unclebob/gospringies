package app

import (
	"image"
	"image/color"
	"math"

	"springs/internal/sim"
)

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

func springDrawColor(spring sim.Spring) color.RGBA {
	if spring.Wall && spring.Temperature > 0 {
		return hotWallColor
	}
	return drawColorFor(spring.Wall, wallSpringColor, springColor)
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

func massDrawColor(mass sim.Mass) color.RGBA {
	return drawColorFor(mass.Fixed, fixedMassColor, massColor)
}

func drawColorFor(useAlternate bool, alternate color.RGBA, fallback color.RGBA) color.RGBA {
	if useAlternate {
		return alternate
	}
	return fallback
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

func demoPickerTitlePoint(rect image.Rectangle) image.Point {
	return image.Pt(rect.Min.X+12, rect.Min.Y+10)
}

func demoPickerRowTextPoint(row image.Rectangle) image.Point {
	return image.Pt(row.Min.X+8, row.Min.Y+4)
}

func demoPickerRowFill(index int) color.RGBA {
	if index%2 == 1 {
		return sectionColor
	}
	return controlColor
}
