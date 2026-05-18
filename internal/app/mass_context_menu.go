package app

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"springs/internal/sim"
)

const (
	massMenuWidth     = 120
	massMenuTitleRows = 1
	massMenuRowHeight = 24
)

type massContextMenu struct {
	Open   bool
	MassID int
	X      int
	Y      int
}

type massMenuItem struct {
	Label  string
	Action func()
}

func (g *Game) openContextAt(x int, y int) {
	if g.demoPickerOpen {
		return
	}
	g.valueDialog.Open = false
	if g.openMassContextMenu(x, y) {
		return
	}
	if g.openSpringConstantDialogAt(x, y) {
		g.massMenu.Open = false
		return
	}
	g.massMenu.Open = false
}

func (g *Game) openMassContextMenu(x int, y int) bool {
	id, ok := g.massAt(g.screenToWorld(simVec(x, y)))
	if !ok {
		return false
	}
	g.massMenu = massContextMenu{Open: true, MassID: id, X: x, Y: y}
	_ = g.editing().SelectMass(id)
	g.selected = true
	return true
}

func simVec(x int, y int) sim.Vec2 {
	return sim.Vec2{X: float64(x), Y: float64(y)}
}

func (g *Game) drawMassContextMenu(screen *ebiten.Image) {
	rect := g.massContextMenuRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mass #%d", g.massMenu.MassID), rect.Min.X+8, rect.Min.Y+4)
	for i, item := range g.massContextMenuItems() {
		row := g.massContextMenuRowRect(i)
		fill := controlColor
		if i%2 == 1 {
			fill = sectionColor
		}
		vector.DrawFilledRect(screen, float32(row.Min.X), float32(row.Min.Y), float32(row.Dx()), float32(row.Dy()), fill, false)
		ebitenutil.DebugPrintAt(screen, item.Label, row.Min.X+8, row.Min.Y+4)
	}
}

func (g *Game) clickMassContextMenu(x int, y int) {
	point := image.Pt(x, y)
	if !point.In(g.massContextMenuRect()) {
		g.massMenu.Open = false
		return
	}
	for i, item := range g.massContextMenuItems() {
		if point.In(g.massContextMenuRowRect(i)) {
			item.Action()
			g.massMenu.Open = false
			return
		}
	}
}

func (g *Game) massContextMenuItems() []massMenuItem {
	mass, ok := g.simulation.MassByID(g.massMenu.MassID)
	if !ok {
		return nil
	}
	items := []massMenuItem{{
		Label: fixedToggleLabel(mass.Fixed),
		Action: func() {
			g.setMassFixed(g.massMenu.MassID, !mass.Fixed)
		},
	}, {
		Label: "Set Mass",
		Action: func() {
			g.openMassValueDialog(g.massMenu.MassID)
		},
	}}
	return items
}

func fixedToggleLabel(fixed bool) string {
	if fixed {
		return "Set Free"
	}
	return "Set Fixed"
}

func (g *Game) massContextMenuRect() image.Rectangle {
	rows := massMenuTitleRows + len(g.massContextMenuItems())
	height := rows * massMenuRowHeight
	x := clampInt(g.massMenu.X, 0, screenWidth-massMenuWidth)
	y := clampInt(g.massMenu.Y, 0, screenHeight-height)
	return image.Rect(x, y, x+massMenuWidth, y+height)
}

func (g *Game) massContextMenuRowRect(index int) image.Rectangle {
	rect := g.massContextMenuRect()
	top := rect.Min.Y + massMenuTitleRows*massMenuRowHeight + index*massMenuRowHeight
	return image.Rect(rect.Min.X, top, rect.Max.X, top+massMenuRowHeight)
}

func (g *Game) setMassFixed(id int, fixed bool) {
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == id {
			g.simulation.Masses[i].Fixed = fixed
			g.dirty = true
			return
		}
	}
}

func (g *Game) setMassValue(id int, value float64) {
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == id {
			g.simulation.Masses[i].Mass = value
			g.dirty = true
			return
		}
	}
}
