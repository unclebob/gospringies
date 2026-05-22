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
	g.editState.selected = true
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
			g.editState.dirty = true
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
	for i := range g.world.simulation.Masses {
		if g.world.simulation.Masses[i].ID == id {
			update(&g.world.simulation.Masses[i])
			g.editState.dirty = true
			return
		}
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"2bdbb5935d38584ec7b626d0e7df0997d95b0f3df684c073a3df870011f99b11","functions":[{"id":"func/Game.openContextAt","name":"Game.openContextAt","line":16,"end_line":31,"hash":"1e57f207ed6a870da7a85bed2a4524b67b710d98db5d1de63593ff9e50985245"},{"id":"func/Game.openMassContextMenu","name":"Game.openMassContextMenu","line":33,"end_line":42,"hash":"e060aca4e06c16f63a689cc1a0403d747a5dd0351df6bd84bc2579e696cdd52d"},{"id":"func/simVec","name":"simVec","line":44,"end_line":46,"hash":"426903adf9b98bd88fe3ce99be2ed5d79a1e7d37fe60d041fa024fd87aedf7a2"},{"id":"func/Game.clickMassContextMenu","name":"Game.clickMassContextMenu","line":48,"end_line":50,"hash":"73cfa1b3722887ac30be111849ae11cb3ba51248b3474a33d19e15711be21027"},{"id":"func/Game.massContextMenuItems","name":"Game.massContextMenuItems","line":52,"end_line":75,"hash":"f544c322573532d5c8b2968a8c85232cd9f386e2e02a7f9b34c21d2d79f5e42a"},{"id":"func/fixedToggleLabel","name":"fixedToggleLabel","line":77,"end_line":82,"hash":"669d646e0a581346a37342be0e36584c8af9f3cdd2188461bfaf351aa7c0d106"},{"id":"func/Game.massContextMenuRect","name":"Game.massContextMenuRect","line":84,"end_line":86,"hash":"cc04353d931ac8f7c3041497efbcda65097707923548e599e196506769294cf7"},{"id":"func/Game.massContextMenuRowRect","name":"Game.massContextMenuRowRect","line":88,"end_line":90,"hash":"cdb1d528831fdb8ae4408c8f42eaa144f1c0ad6dcb700fd34bebe153709ac364"},{"id":"func/Game.setMassFixed","name":"Game.setMassFixed","line":92,"end_line":94,"hash":"53cfbeb3bcfd8e6aaa611ec2ef52aeeb2ddc4857926b1d5ad186b05eb168042a"},{"id":"func/Game.setMassValue","name":"Game.setMassValue","line":96,"end_line":98,"hash":"a96b9235e638c70e3f1595fac6758c7c0daa6a9fd4f56fb67a1be38f63b9ef40"},{"id":"func/Game.updateMass","name":"Game.updateMass","line":100,"end_line":108,"hash":"b5706321e2f2fae83d300679e8b285c9495d06661928c15edbcb34c6925db5bd"}]}
// mutate4go-manifest-end
