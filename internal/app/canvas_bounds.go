package app

import "springs/internal/sim"

func (g *Game) canvasWorldBounds() (float64, float64, float64, float64) {
	return g.canvasWorldBoundsForHeight(g.simulation.Bounds.Height)
}

func (g *Game) canvasWorldBoundsForHeight(height float64) (float64, float64, float64, float64) {
	canvas := visibleRegionRects()["canvas"]
	minX := float64(canvas.Min.X)
	maxX := float64(canvas.Max.X)
	if !g.canvasYUp {
		return minX, maxX, float64(canvas.Min.Y), float64(canvas.Max.Y)
	}
	return minX, maxX, height - float64(canvas.Max.Y), height - float64(canvas.Min.Y)
}

func (g *Game) applyCanvasWallBounds(world *sim.Simulation) {
	minX, maxX, minY, maxY := g.canvasWorldBoundsForHeight(world.Bounds.Height)
	world.Bounds.Left = minX
	world.Bounds.Right = maxX
	world.Bounds.Bottom = minY
	world.Bounds.Top = maxY
}

func (g *Game) clampToCanvas(position sim.Vec2) sim.Vec2 {
	minX, maxX, minY, maxY := g.canvasWorldBounds()
	return sim.Vec2{
		X: clampFloat(position.X, minX, maxX),
		Y: clampFloat(position.Y, minY, maxY),
	}
}

func (g *Game) positionInCanvas(position sim.Vec2) bool {
	minX, maxX, minY, maxY := g.canvasWorldBounds()
	return position.X >= minX && position.X <= maxX && position.Y >= minY && position.Y <= maxY
}

func (g *Game) snapToCanvas(position sim.Vec2) sim.Vec2 {
	return g.clampToCanvas(g.snapToGrid(position))
}
