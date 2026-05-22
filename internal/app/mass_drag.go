package app

import "springs/internal/sim"

func (g *Game) finishWorldPointer(position sim.Vec2) {
	if g.pointer.draggingMassID != 0 {
		g.finishMassDrag(position)
	}
	if g.pointer.pendingSpringID != 0 && !g.pointer.springChainActive {
		g.finishSpringAt(position)
	}
	if g.pointer.selectionDrag {
		g.finishSelectGesture(position)
	}
}

func (g *Game) finishMassDrag(position sim.Vec2) {
	if g.pointer.dragMoved && g.throwKeyPressed() {
		g.throwDraggedMasses(position.Sub(g.pointer.draggingStart))
		return
	}
	if g.pointer.dragMoved || g.pointer.selectionAdd {
		return
	}
	if selectionClick(g.pointer.draggingStart, position) {
		_ = g.editing().SelectMass(g.pointer.draggingMassID)
		g.syncSelectionState()
	}
}

func (g *Game) throwDraggedMasses(velocity sim.Vec2) {
	if len(g.pointer.draggingOffsets) > 0 {
		g.throwSelectedDraggingMasses(velocity)
		g.markDirty()
		return
	}
	g.throwSingleDraggingMass(velocity)
}

func (g *Game) throwSelectedDraggingMasses(velocity sim.Vec2) {
	for i := range g.world.simulation.Masses {
		if _, ok := g.pointer.draggingOffsets[g.world.simulation.Masses[i].ID]; ok {
			g.world.simulation.Masses[i].Velocity = velocity
		}
	}
}

func (g *Game) throwSingleDraggingMass(velocity sim.Vec2) {
	for i := range g.world.simulation.Masses {
		if g.world.simulation.Masses[i].ID != g.pointer.draggingMassID {
			continue
		}
		g.world.simulation.Masses[i].Velocity = velocity
		g.markDirty()
		return
	}
}

func (g *Game) beginMassDrag(position sim.Vec2) {
	id, ok := g.massAt(position)
	if !ok {
		return
	}
	g.pointer.draggingMassID = id
	g.pointer.draggingStart = position
	g.pointer.draggingLast = position
	g.captureDraggingOffsets(position)
	g.pointer.dragMoved = false
	g.DragMass(id, position)
}

func (g *Game) captureDraggingOffsets(cursor sim.Vec2) {
	g.pointer.draggingOffsets = map[int]sim.Vec2{}
	if g.captureSelectedDraggingOffsets(cursor) {
		return
	}
	if mass, ok := g.world.simulation.MassByID(g.pointer.draggingMassID); ok {
		g.pointer.draggingOffsets[g.pointer.draggingMassID] = mass.Position.Sub(cursor)
	}
}

func (g *Game) captureSelectedDraggingOffsets(cursor sim.Vec2) bool {
	if len(g.editing().SelectedMasses) == 0 || !g.editing().MassSelected(g.pointer.draggingMassID) {
		return false
	}
	for _, mass := range g.world.simulation.Masses {
		if g.editing().MassSelected(mass.ID) {
			g.pointer.draggingOffsets[mass.ID] = mass.Position.Sub(cursor)
		}
	}
	return true
}

func (g *Game) pinDraggingMasses(cursor sim.Vec2) {
	if g.pointer.draggingMassID == 0 || len(g.pointer.draggingOffsets) == 0 {
		return
	}
	g.applyDraggingOffsets(cursor)
}
