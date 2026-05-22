package app

import (
	"math"

	"springs/internal/sim"
)

func (g *Game) selectNearest(position sim.Vec2) {
	if err := g.editing().SelectNearest(position, false); err == nil {
		g.setSelected(true)
	}
}

func (g *Game) massAt(position sim.Vec2) (int, bool) {
	for _, mass := range g.world.simulation.Masses {
		_, _, radius := massDrawCircle(mass)
		if math.Hypot(mass.Position.X-position.X, mass.Position.Y-position.Y) <= float64(radius) {
			return mass.ID, true
		}
	}
	return 0, false
}

func massDrawCircle(mass sim.Mass) (float32, float32, float32) {
	return float32(mass.Position.X), float32(mass.Position.Y), float32(sim.MassRadius(mass))
}

func (g *Game) screenToWorld(position sim.Vec2) sim.Vec2 {
	return g.canvasCoordinate(position)
}

func (g *Game) worldToScreen(position sim.Vec2) sim.Vec2 {
	return g.canvasCoordinate(position)
}

func (g *Game) canvasCoordinate(position sim.Vec2) sim.Vec2 {
	if !g.run.canvasYUp {
		return position
	}
	return g.flipCanvasY(position)
}

func (g *Game) flipCanvasY(position sim.Vec2) sim.Vec2 {
	return sim.Vec2{X: position.X, Y: g.world.simulation.Bounds.Height - position.Y}
}
