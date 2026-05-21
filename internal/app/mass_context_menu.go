package app

import (
	"image"

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
		g.springMenu.Open = false
		return
	}
	if g.openSpringContextMenu(x, y) {
		g.massMenu.Open = false
		return
	}
	g.massMenu.Open = false
	g.springMenu.Open = false
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
	}, {
		Label: "Set Center",
		Action: func() {
			g.simulation.SetForceCenter([]int{g.massMenu.MassID})
			g.dirty = true
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:09:07-05:00","module_hash":"d3ff9193e239f8a82f492e1e2480bf4826cba929e7c3ce8d0faafa9c2efb03b3","functions":[{"id":"func/Game.openContextAt","name":"Game.openContextAt","line":27,"end_line":40,"hash":"94918b44c6841c4bf14bfeed0f6362959d223736b5d7af011448857d719115d4"},{"id":"func/Game.openMassContextMenu","name":"Game.openMassContextMenu","line":42,"end_line":51,"hash":"41ac01d73377c3e206abb5c9f995a95e24c3c14c271b38461f1e867ef4717b25"},{"id":"func/simVec","name":"simVec","line":53,"end_line":55,"hash":"426903adf9b98bd88fe3ce99be2ed5d79a1e7d37fe60d041fa024fd87aedf7a2"},{"id":"func/Game.clickMassContextMenu","name":"Game.clickMassContextMenu","line":57,"end_line":70,"hash":"6f1b40d173852eb8b35137cfbb6e86b255a41459b720f31cf39a0c76228cb02a"},{"id":"func/Game.massContextMenuItems","name":"Game.massContextMenuItems","line":72,"end_line":89,"hash":"b03e27ff56b8ef73bbf3d30181bce8f484f69ed54cd70b3275789c56beaefa04"},{"id":"func/fixedToggleLabel","name":"fixedToggleLabel","line":91,"end_line":96,"hash":"669d646e0a581346a37342be0e36584c8af9f3cdd2188461bfaf351aa7c0d106"},{"id":"func/Game.massContextMenuRect","name":"Game.massContextMenuRect","line":98,"end_line":104,"hash":"4cc0c0824b67af7b0a70ded622c21e56807942174026db3bd0ed2e38c2667cc5"},{"id":"func/Game.massContextMenuRowRect","name":"Game.massContextMenuRowRect","line":106,"end_line":110,"hash":"7ce70e9244fc1795f00682621e6fdba6a2ba4a22bd7fecc2fef90257400aaad6"},{"id":"func/Game.setMassFixed","name":"Game.setMassFixed","line":112,"end_line":120,"hash":"af5459d5e85a813382383f4536d5b17e5e59f62c690e628c1d0716c5f6fdc985"},{"id":"func/Game.setMassValue","name":"Game.setMassValue","line":122,"end_line":130,"hash":"03ba2eb968b5090cfde01ff8c4230032a0fe639bb48732a32a9ae434393e1d97"}]}
// mutate4go-manifest-end
