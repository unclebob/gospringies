package app

import "image"

type springContextMenu struct {
	Open     bool
	SpringID int
	X        int
	Y        int
}

func (g *Game) openSpringContextMenu(x int, y int) bool {
	spring, ok := g.springAtPosition(g.screenToWorld(simVec(x, y)))
	if !ok {
		return false
	}
	g.overlays.springMenu = springContextMenu{Open: true, SpringID: spring.ID, X: x, Y: y}
	return true
}

func (g *Game) clickSpringContextMenu(x int, y int) {
	clickContextMenu(x, y, g.springContextMenuRect(), g.springContextMenuItems(), func() { g.overlays.springMenu.Open = false })
}

func (g *Game) springContextMenuItems() []contextMenuItem {
	if _, ok := g.world.simulation.SpringByID(g.overlays.springMenu.SpringID); !ok {
		return nil
	}
	id := g.overlays.springMenu.SpringID
	return []contextMenuItem{{
		Label:  "Kspring",
		Action: func() { g.openSpringValueDialog(id, springValueKspring) },
	}, {
		Label:  "Kdamp",
		Action: func() { g.openSpringValueDialog(id, springValueKdamp) },
	}, {
		Label:  "RestLen",
		Action: func() { g.openSpringValueDialog(id, springValueRestLen) },
	}, {
		Label:  "Wall",
		Action: func() { g.toggleSpringWall(id) },
	}, {
		Label:  "Temperature",
		Action: func() { g.openSpringValueDialog(id, springValueTemperature) },
	}}
}

func (g *Game) toggleSpringWall(id int) {
	for i := range g.world.simulation.Springs {
		if g.world.simulation.Springs[i].ID == id {
			g.world.simulation.Springs[i].Wall = !g.world.simulation.Springs[i].Wall
			g.editState.dirty = true
			return
		}
	}
}

func (g *Game) SpringContextMenuLabelsForSpring(id int) []string {
	g.overlays.springMenu = springContextMenu{SpringID: id}
	return contextMenuLabels(g.springContextMenuItems())
}

func (g *Game) SelectSpringContextMenuItem(id int, label string) bool {
	g.overlays.springMenu = springContextMenu{SpringID: id}
	return selectContextMenuItem(g.springContextMenuItems(), label, func() { g.overlays.springMenu.Open = false })
}

func (g *Game) springContextMenuRect() image.Rectangle {
	return contextMenuRect(g.overlays.springMenu.X, g.overlays.springMenu.Y, len(g.springContextMenuItems()))
}

func (g *Game) springContextMenuRowRect(index int) image.Rectangle {
	return contextMenuRowRect(g.springContextMenuRect(), index)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"ff87515c81f8ce656f6428af78170a08960bcdeefcd9577115233d0cefab8a04","functions":[{"id":"func/Game.openSpringContextMenu","name":"Game.openSpringContextMenu","line":12,"end_line":19,"hash":"89047e87b09eb101509d4792080c68dc7a7151494dc62f8120f5c85751913b1a"},{"id":"func/Game.clickSpringContextMenu","name":"Game.clickSpringContextMenu","line":21,"end_line":23,"hash":"5463b26271998cf75ab9bdbc5d14212df3bc9186f46934a9d0121f060e6aa35c"},{"id":"func/Game.springContextMenuItems","name":"Game.springContextMenuItems","line":25,"end_line":46,"hash":"fb2b85f7ceff582e675cc30062c26c3e31f0f1beac563e21d4752e8f653263e1"},{"id":"func/Game.toggleSpringWall","name":"Game.toggleSpringWall","line":48,"end_line":56,"hash":"fac988425815e1421bbf13aae2cb4b89a3d06737428f39e617db840206696595"},{"id":"func/Game.SpringContextMenuLabelsForSpring","name":"Game.SpringContextMenuLabelsForSpring","line":58,"end_line":61,"hash":"f957138b99ba5df050d6f0aa17e5253481507480e3d3c3ffe03bc684ec60cc48"},{"id":"func/Game.SelectSpringContextMenuItem","name":"Game.SelectSpringContextMenuItem","line":63,"end_line":66,"hash":"19a7319ef6204ead24c43d06b34b17898131d31795f7d8f1efb3ab1556905f17"},{"id":"func/Game.springContextMenuRect","name":"Game.springContextMenuRect","line":68,"end_line":70,"hash":"b74429fb721acbcf0e42d1917ee9fa759493867ff560313555dd580c5a198ce0"},{"id":"func/Game.springContextMenuRowRect","name":"Game.springContextMenuRowRect","line":72,"end_line":74,"hash":"9b9137af2fa4f7c230063820986f62f11b2c58dec2311e6475854dac1cf25d89"}]}
// mutate4go-manifest-end
