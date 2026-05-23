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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-23T11:50:38-05:00","module_hash":"3b74193faab5272bd41eb18afbdfe3a24870e6b15144f41e2751a4de5b1059bc","functions":[{"id":"func/Game.gridPointRects","name":"Game.gridPointRects","line":17,"end_line":31,"hash":"600750169ed58b07bd434dafed0f8a6fd143d81bade263f108fed60c984f43ee"},{"id":"func/gridPointPixelSize","name":"gridPointPixelSize","line":33,"end_line":35,"hash":"6c0a1cf159cdd1cb2dcab7eaf4f8d631777564f781ac420f6d26dab7a0fe6730"},{"id":"func/gridPointAntiAlias","name":"gridPointAntiAlias","line":37,"end_line":39,"hash":"9da19824153c650376ff119c27f1fb0c1582d8b0411321824a1d13b1e594d627"},{"id":"func/Game.gridPoints","name":"Game.gridPoints","line":41,"end_line":57,"hash":"cbd43d3678d2dc7775644dbb6ae63946ed4ef3662f70a838663d69ab3e68fc79"},{"id":"func/validGridSnapSize","name":"validGridSnapSize","line":59,"end_line":61,"hash":"cf2cc0632f5d59e877c981dca062e63dac4c535b77c3065878c5c3a34b22b37e"},{"id":"func/firstGridCoordinateAtOrAfter","name":"firstGridCoordinateAtOrAfter","line":63,"end_line":65,"hash":"8a7c004c8f82076f9188bb1776b097331b802ef718b63f68bdfd2eae4acf9ec7"},{"id":"func/springDrawColor","name":"springDrawColor","line":67,"end_line":72,"hash":"f96e329a29cfaab0418dab4c7ccf7ee1ff0ed0e74d315a57b4aebd3bb12848a4"},{"id":"func/Game.pendingSpringLine","name":"Game.pendingSpringLine","line":74,"end_line":83,"hash":"35e4cb796d7894e574bf1abef437282035d185710dc5de2fb2c700011ce01b12"},{"id":"func/selectionRectangleLines","name":"selectionRectangleLines","line":85,"end_line":96,"hash":"ac6138efe5e031f4de8e884ce3557eafe1b6b8ae5f95fba79c0fd3846af14d8e"},{"id":"func/massDrawColor","name":"massDrawColor","line":98,"end_line":100,"hash":"890b5d5de224ff327a7f9612820334a87d1d55c98a87946eedb0c9a12f73ee25"},{"id":"func/drawColorFor","name":"drawColorFor","line":102,"end_line":107,"hash":"deb424640c409f469cf9100cd980a426225e594d77029431530fc96e63cb5655"},{"id":"func/wallDrawLines","name":"wallDrawLines","line":114,"end_line":121,"hash":"3f62a9ee6f8fb680fb031e79c6de06472479fab01a2b74a8a18731ff5fd99564"},{"id":"func/Game.selectedMasses","name":"Game.selectedMasses","line":123,"end_line":129,"hash":"425a668feb687968b0a698c5a1cb4b9212c6e747cc061cb6e2765985b487b04e"},{"id":"func/Game.explicitSelectedMasses","name":"Game.explicitSelectedMasses","line":131,"end_line":139,"hash":"ce61a71692909056f347ee9ca4e118c7dc813219d96d29060735f2cc717bba35"},{"id":"func/Game.allMassesImplicitlySelected","name":"Game.allMassesImplicitlySelected","line":141,"end_line":143,"hash":"6d552ca010268d6627068ed4fd689846bdbce4a7b74ca907429d953ec7a7db9a"},{"id":"func/Game.selectedSpringLines","name":"Game.selectedSpringLines","line":145,"end_line":158,"hash":"0e630d07b96362560040d393c284e777cbe0d79edb7e05fa929db378c463fd1d"},{"id":"func/selectedMassOutline","name":"selectedMassOutline","line":167,"end_line":173,"hash":"09e522ba39856beab87cf554a1a6ce1af7c4466412f407118cc223529ae81da3"},{"id":"func/selectionOutline","name":"selectionOutline","line":175,"end_line":185,"hash":"d4ec3f8743e56bd8e3bed0d58f60210a5195c24b29e99c152f2e5b09b8806bdf"},{"id":"func/demoPickerTitlePoint","name":"demoPickerTitlePoint","line":187,"end_line":189,"hash":"6464bfdd1c06d73ae89330709204df4e631a504b11a2836853b12b9390a24c36"},{"id":"func/demoPickerRowTextPoint","name":"demoPickerRowTextPoint","line":191,"end_line":193,"hash":"3b3ab6fbeb25fc3953756b46bc3bae8223455de13ffd523cd185387be31a1e85"},{"id":"func/demoPickerRowFill","name":"demoPickerRowFill","line":195,"end_line":200,"hash":"de247040da4d13d92370ece4458f45e0daf0cc77c7aa08f04a07bcd163eea066"}]}
// mutate4go-manifest-end
