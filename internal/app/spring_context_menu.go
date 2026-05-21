package app

import "image"

const (
	springMenuWidth     = 120
	springMenuTitleRows = 1
	springMenuRowHeight = 24
)

type springContextMenu struct {
	Open     bool
	SpringID int
	X        int
	Y        int
}

type springMenuItem struct {
	Label  string
	Action func()
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
	point := image.Pt(x, y)
	if !point.In(g.springContextMenuRect()) {
		g.springMenu.Open = false
		return
	}
	for i, item := range g.springContextMenuItems() {
		if point.In(g.springContextMenuRowRect(i)) {
			item.Action()
			g.springMenu.Open = false
			return
		}
	}
}

func (g *Game) springContextMenuItems() []springMenuItem {
	if _, ok := g.simulation.SpringByID(g.springMenu.SpringID); !ok {
		return nil
	}
	id := g.springMenu.SpringID
	return []springMenuItem{{
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
	g.springMenu = springContextMenu{Open: true, SpringID: id}
	items := g.springContextMenuItems()
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}

func (g *Game) SelectSpringContextMenuItem(id int, label string) bool {
	g.springMenu = springContextMenu{Open: true, SpringID: id}
	for _, item := range g.springContextMenuItems() {
		if item.Label == label {
			item.Action()
			g.springMenu.Open = false
			return true
		}
	}
	return false
}

func (g *Game) springContextMenuRect() image.Rectangle {
	rows := springMenuTitleRows + len(g.springContextMenuItems())
	height := rows * springMenuRowHeight
	x := clampInt(g.springMenu.X, 0, screenWidth-springMenuWidth)
	y := clampInt(g.springMenu.Y, 0, screenHeight-height)
	return image.Rect(x, y, x+springMenuWidth, y+height)
}

func (g *Game) springContextMenuRowRect(index int) image.Rectangle {
	rect := g.springContextMenuRect()
	top := rect.Min.Y + springMenuTitleRows*springMenuRowHeight + index*springMenuRowHeight
	return image.Rect(rect.Min.X, top, rect.Max.X, top+springMenuRowHeight)
}
