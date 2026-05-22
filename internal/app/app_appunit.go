//go:build appunit

package app

func Run() error {
	return nil
}

func (g *Game) Update() error {
	g.inputActive = true
	g.tickNumericTextField()
	g.advanceSimulationFrame()
	return nil
}

func (g *Game) shiftKeyPressed() bool {
	return g.shiftDown
}

func (g *Game) controlKeyPressed() bool {
	return g.controlDown
}

func (g *Game) throwKeyPressed() bool {
	return g.throwDown
}
