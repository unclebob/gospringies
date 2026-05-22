package app

import (
	"math"

	"springs/internal/sim"
)

func (g *Game) beginCanvasGesture(position sim.Vec2) {
	g.beginSelectGesture(position)
}

func (g *Game) beginSelectGesture(position sim.Vec2) {
	g.pointer.selectionAdd = g.shiftKeyPressed()
	id, ok := g.massAt(position)
	if ok {
		alreadySelected := g.editing().MassSelected(id)
		if g.pointer.selectionAdd {
			_ = g.editing().AddMassSelection(id)
		} else if !alreadySelected {
			_ = g.editing().SelectMass(id)
		}
		g.syncSelectionState()
		g.pointer.draggingMassID = id
		g.pointer.draggingStart = position
		g.pointer.draggingLast = position
		g.captureDraggingOffsets(position)
		g.pointer.dragMoved = false
		return
	}
	g.pointer.selectionDrag = true
	g.pointer.selectionStart = position
	g.pointer.selectionEnd = position
}

func (g *Game) finishSelectGesture(position sim.Vec2) {
	start := g.pointer.selectionStart
	g.pointer.selectionEnd = position
	g.pointer.selectionDrag = false
	if selectionClick(start, position) {
		if !g.pointer.selectionAdd {
			g.clearSelection()
		}
		g.createMassAt(position, g.pointer.selectionAdd)
		return
	}
	g.editing().BoxSelect(start, position, g.pointer.selectionAdd)
	g.syncSelectionState()
}

func selectionClick(start sim.Vec2, end sim.Vec2) bool {
	return math.Hypot(start.X-end.X, start.Y-end.Y) < 3
}
