package app

import (
	"springs/internal/edit"
	"springs/internal/sim"
)

func (g *Game) updateSpringChainEnd(position sim.Vec2) {
	if g.springChainActive {
		g.pendingSpringEnd = g.clampToCanvas(position)
	}
}

func (g *Game) beginControlPlacementAt(position sim.Vec2) {
	if !g.positionInCanvas(position) {
		return
	}
	if id, ok := g.massAt(position); ok {
		g.beginSpringAt(position)
		g.pendingSpringID = id
		return
	}
	id, ok := g.createMassAt(position, false)
	if !ok {
		return
	}
	g.pendingSpringID = id
	g.pendingSpringEnd = g.massPosition(id, position)
	g.springChainActive = true
}

func (g *Game) continueSpringChainAt(position sim.Vec2, keepChain bool) {
	if !g.positionInCanvas(position) {
		return
	}
	endID, existing, ok := g.springChainEndpointAt(position)
	if !ok {
		return
	}
	g.createSpringBetween(g.pendingSpringID, endID)
	g.finishSpringChainStep(endID, position, keepChain && !existing)
}

func (g *Game) springChainEndpointAt(position sim.Vec2) (int, bool, bool) {
	if endID, existing := g.massAt(position); existing {
		return endID, true, true
	}
	endID, ok := g.createMassAt(position, false)
	return endID, false, ok
}

func (g *Game) finishSpringChainStep(endID int, position sim.Vec2, keepChain bool) {
	if !keepChain {
		g.clearPendingSpring()
		return
	}
	g.pendingSpringID = endID
	g.pendingSpringEnd = g.massPosition(endID, position)
	g.springChainActive = true
}

func (g *Game) createSpringBetween(startID int, endID int) bool {
	if startID == 0 || endID == 0 || startID == endID {
		return false
	}
	editor := g.editing()
	editor.Mode = edit.ModeAddSpring
	if _, err := editor.CreateSpring(startID, endID); err == nil {
		g.dirty = true
		return true
	}
	return false
}

func (g *Game) clearPendingSpring() {
	g.pendingSpringID = 0
	g.pendingSpringEnd = sim.Vec2{}
	g.springChainActive = false
}

func (g *Game) massPosition(id int, fallback sim.Vec2) sim.Vec2 {
	if mass, ok := g.simulation.MassByID(id); ok {
		return mass.Position
	}
	return fallback
}
