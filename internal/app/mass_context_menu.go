package app

import (
	"image"

	"springs/internal/sim"
)

type massContextMenu struct {
	Open   bool
	MassID int
	X      int
	Y      int
}

func (g *Game) openContextAt(x int, y int) {
	if g.controls.demoPickerOpen {
		return
	}
	g.overlays.value.Open = false
	if g.openMassContextMenu(x, y) {
		g.overlays.springMenu.Open = false
		return
	}
	if g.openSpringContextMenu(x, y) {
		g.overlays.massMenu.Open = false
		return
	}
	g.overlays.massMenu.Open = false
	g.overlays.springMenu.Open = false
}

func (g *Game) openMassContextMenu(x int, y int) bool {
	id, ok := g.massAt(g.screenToWorld(simVec(x, y)))
	if !ok {
		return false
	}
	g.overlays.massMenu = massContextMenu{Open: true, MassID: id, X: x, Y: y}
	_ = g.editing().SelectMass(id)
	g.setSelected(true)
	return true
}

func simVec(x int, y int) sim.Vec2 {
	return sim.Vec2{X: float64(x), Y: float64(y)}
}

func (g *Game) clickMassContextMenu(x int, y int) {
	clickContextMenu(x, y, g.massContextMenuRect(), g.massContextMenuItems(), func() { g.overlays.massMenu.Open = false })
}

func (g *Game) massContextMenuItems() []contextMenuItem {
	mass, ok := g.world.simulation.MassByID(g.overlays.massMenu.MassID)
	if !ok {
		return nil
	}
	items := []contextMenuItem{{
		Label: fixedToggleLabel(mass.Fixed),
		Action: func() {
			g.setMassFixed(g.overlays.massMenu.MassID, !mass.Fixed)
		},
	}, {
		Label: "Set Mass",
		Action: func() {
			g.openMassValueDialog(g.overlays.massMenu.MassID)
		},
	}, {
		Label: "Set Center",
		Action: func() {
			g.world.simulation.SetForceCenter([]int{g.overlays.massMenu.MassID})
			g.markDirty()
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
	return contextMenuRect(g.overlays.massMenu.X, g.overlays.massMenu.Y, len(g.massContextMenuItems()))
}

func (g *Game) massContextMenuRowRect(index int) image.Rectangle {
	return contextMenuRowRect(g.massContextMenuRect(), index)
}

func (g *Game) setMassFixed(id int, fixed bool) {
	g.updateMass(id, func(mass *sim.Mass) { mass.Fixed = fixed })
}

func (g *Game) setMassValue(id int, value float64) {
	g.updateMass(id, func(mass *sim.Mass) { mass.Mass = value })
}

func (g *Game) updateMass(id int, update func(*sim.Mass)) {
	updateByIDAndMarkDirty(g, g.world.simulation.Masses, id, func(mass *sim.Mass) int { return mass.ID }, update)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:44:43-05:00","module_hash":"26ffe0b8e4e2619305dc9241487bd3758f8d0989e8397d3ce263277663f51fca","functions":[{"id":"func/Game.openContextAt","name":"Game.openContextAt","line":16,"end_line":31,"hash":"1e57f207ed6a870da7a85bed2a4524b67b710d98db5d1de63593ff9e50985245"},{"id":"func/Game.openMassContextMenu","name":"Game.openMassContextMenu","line":33,"end_line":42,"hash":"cac7c0479c524af13aac78d5ec8aa46905438fc364218e1235d55e6c2ea38fa1"},{"id":"func/simVec","name":"simVec","line":44,"end_line":46,"hash":"426903adf9b98bd88fe3ce99be2ed5d79a1e7d37fe60d041fa024fd87aedf7a2"},{"id":"func/Game.clickMassContextMenu","name":"Game.clickMassContextMenu","line":48,"end_line":50,"hash":"73cfa1b3722887ac30be111849ae11cb3ba51248b3474a33d19e15711be21027"},{"id":"func/Game.massContextMenuItems","name":"Game.massContextMenuItems","line":52,"end_line":75,"hash":"5ccef17c3b006cca5caf0cad7deaa3245e3700151ab21c591b8c5c12620c3af6"},{"id":"func/fixedToggleLabel","name":"fixedToggleLabel","line":77,"end_line":82,"hash":"669d646e0a581346a37342be0e36584c8af9f3cdd2188461bfaf351aa7c0d106"},{"id":"func/Game.massContextMenuRect","name":"Game.massContextMenuRect","line":84,"end_line":86,"hash":"cc04353d931ac8f7c3041497efbcda65097707923548e599e196506769294cf7"},{"id":"func/Game.massContextMenuRowRect","name":"Game.massContextMenuRowRect","line":88,"end_line":90,"hash":"cdb1d528831fdb8ae4408c8f42eaa144f1c0ad6dcb700fd34bebe153709ac364"},{"id":"func/Game.setMassFixed","name":"Game.setMassFixed","line":92,"end_line":94,"hash":"53cfbeb3bcfd8e6aaa611ec2ef52aeeb2ddc4857926b1d5ad186b05eb168042a"},{"id":"func/Game.setMassValue","name":"Game.setMassValue","line":96,"end_line":98,"hash":"a96b9235e638c70e3f1595fac6758c7c0daa6a9fd4f56fb67a1be38f63b9ef40"},{"id":"func/Game.updateMass","name":"Game.updateMass","line":100,"end_line":102,"hash":"49a30566bb316408863839748b84ce8f37e0710df55d804496050d060958d3ff"}]}
// mutate4go-manifest-end
