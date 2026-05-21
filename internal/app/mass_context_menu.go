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
	clickContextMenu(x, y, g.massContextMenuRect(), g.massContextMenuItems(), func() { g.massMenu.Open = false })
}

func (g *Game) massContextMenuItems() []contextMenuItem {
	mass, ok := g.simulation.MassByID(g.massMenu.MassID)
	if !ok {
		return nil
	}
	items := []contextMenuItem{{
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
	return contextMenuRect(g.massMenu.X, g.massMenu.Y, len(g.massContextMenuItems()))
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
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == id {
			update(&g.simulation.Masses[i])
			g.dirty = true
			return
		}
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-21T11:38:39-05:00","module_hash":"17ce10de463d0ad2b9017bb928052f9f0b98904975c4e5ab8ad64852ed617b6f","functions":[{"id":"func/Game.openContextAt","name":"Game.openContextAt","line":16,"end_line":31,"hash":"71a91251742e2e69bfad120defb40948067b1c439d3bab02a70d0e585bec335a"},{"id":"func/Game.openMassContextMenu","name":"Game.openMassContextMenu","line":33,"end_line":42,"hash":"41ac01d73377c3e206abb5c9f995a95e24c3c14c271b38461f1e867ef4717b25"},{"id":"func/simVec","name":"simVec","line":44,"end_line":46,"hash":"426903adf9b98bd88fe3ce99be2ed5d79a1e7d37fe60d041fa024fd87aedf7a2"},{"id":"func/Game.clickMassContextMenu","name":"Game.clickMassContextMenu","line":48,"end_line":50,"hash":"d6536a1fa897faf609f26ddad7c1572e81dcd7f73339b3c8062e6f3852a3eda3"},{"id":"func/Game.massContextMenuItems","name":"Game.massContextMenuItems","line":52,"end_line":75,"hash":"5e382a769f445eff8a69c86d61babcee0051c258e1ab8d09a8700f14a90bf673"},{"id":"func/fixedToggleLabel","name":"fixedToggleLabel","line":77,"end_line":82,"hash":"669d646e0a581346a37342be0e36584c8af9f3cdd2188461bfaf351aa7c0d106"},{"id":"func/Game.massContextMenuRect","name":"Game.massContextMenuRect","line":84,"end_line":86,"hash":"297f0e97499f281d462c510d00ffe4a5e47b2f688b000fe086473dc7f8a8af8a"},{"id":"func/Game.massContextMenuRowRect","name":"Game.massContextMenuRowRect","line":88,"end_line":90,"hash":"cdb1d528831fdb8ae4408c8f42eaa144f1c0ad6dcb700fd34bebe153709ac364"},{"id":"func/Game.setMassFixed","name":"Game.setMassFixed","line":92,"end_line":94,"hash":"53cfbeb3bcfd8e6aaa611ec2ef52aeeb2ddc4857926b1d5ad186b05eb168042a"},{"id":"func/Game.setMassValue","name":"Game.setMassValue","line":96,"end_line":98,"hash":"a96b9235e638c70e3f1595fac6758c7c0daa6a9fd4f56fb67a1be38f63b9ef40"},{"id":"func/Game.updateMass","name":"Game.updateMass","line":100,"end_line":108,"hash":"cad1b972f5de2dc80b60e00d8d791370b21de4625f8e173140ca7055956d08ad"}]}
// mutate4go-manifest-end
