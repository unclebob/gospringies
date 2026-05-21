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
	g.springMenu = springContextMenu{Open: true, SpringID: spring.ID, X: x, Y: y}
	return true
}

func (g *Game) clickSpringContextMenu(x int, y int) {
	clickContextMenu(x, y, g.springContextMenuRect(), g.springContextMenuItems(), func() { g.springMenu.Open = false })
}

func (g *Game) springContextMenuItems() []contextMenuItem {
	if _, ok := g.simulation.SpringByID(g.springMenu.SpringID); !ok {
		return nil
	}
	id := g.springMenu.SpringID
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
	for i := range g.simulation.Springs {
		if g.simulation.Springs[i].ID == id {
			g.simulation.Springs[i].Wall = !g.simulation.Springs[i].Wall
			g.dirty = true
			return
		}
	}
}

func (g *Game) SpringContextMenuLabelsForSpring(id int) []string {
	g.springMenu = springContextMenu{SpringID: id}
	return contextMenuLabels(g.springContextMenuItems())
}

func (g *Game) SelectSpringContextMenuItem(id int, label string) bool {
	g.springMenu = springContextMenu{SpringID: id}
	return selectContextMenuItem(g.springContextMenuItems(), label, func() { g.springMenu.Open = false })
}

func (g *Game) springContextMenuRect() image.Rectangle {
	return contextMenuRect(g.springMenu.X, g.springMenu.Y, len(g.springContextMenuItems()))
}

func (g *Game) springContextMenuRowRect(index int) image.Rectangle {
	return contextMenuRowRect(g.springContextMenuRect(), index)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-21T11:44:18-05:00","module_hash":"267fd1e2e25e198a4a977cd5dd075d6eb768a18b99950110d9810bd70fdd4666","functions":[{"id":"func/Game.openSpringContextMenu","name":"Game.openSpringContextMenu","line":12,"end_line":19,"hash":"6446d9593043d9fd93669f15cc3256442f8bf4d16259376ff48fbb92c3ff912c"},{"id":"func/Game.clickSpringContextMenu","name":"Game.clickSpringContextMenu","line":21,"end_line":23,"hash":"f18b633e0d1a2828b01838214f4b067090c2526d4e03074c213f0a149d30321e"},{"id":"func/Game.springContextMenuItems","name":"Game.springContextMenuItems","line":25,"end_line":43,"hash":"e27cfb5aeb7c20d31bf15a0affbf7cb7af9251cc2d63457aea51b4546d38e536"},{"id":"func/Game.toggleSpringWall","name":"Game.toggleSpringWall","line":45,"end_line":53,"hash":"4b72ff631abc41d91032b6b0a9df0b1cfb7e31837f1c95f49b29ac5934d7a45f"},{"id":"func/Game.SpringContextMenuLabelsForSpring","name":"Game.SpringContextMenuLabelsForSpring","line":55,"end_line":58,"hash":"9e7c22d38b45e68ab8b3ee15b2fdd9254a51f484213f1cd6dd4b495f806c6fd1"},{"id":"func/Game.SelectSpringContextMenuItem","name":"Game.SelectSpringContextMenuItem","line":60,"end_line":63,"hash":"a4ff18e6d9e91571e38c770f6f9573775e2d10e16462978548b663627631e7cd"},{"id":"func/Game.springContextMenuRect","name":"Game.springContextMenuRect","line":65,"end_line":67,"hash":"d2b507ac04242d2971edaeaaf7c33f322f66bcc4e1fdd7afd93403eb64c28999"},{"id":"func/Game.springContextMenuRowRect","name":"Game.springContextMenuRowRect","line":69,"end_line":71,"hash":"9b9137af2fa4f7c230063820986f62f11b2c58dec2311e6475854dac1cf25d89"}]}
// mutate4go-manifest-end
