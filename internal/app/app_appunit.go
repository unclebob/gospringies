//go:build appunit

package app

func Run() error {
	return nil
}

func (g *Game) Update() error {
	g.runtime.inputActive = true
	g.tickNumericTextField()
	g.advanceSimulationFrame()
	return nil
}

func (g *Game) shiftKeyPressed() bool {
	return g.keyboard.shiftDown
}

func (g *Game) controlKeyPressed() bool {
	return g.keyboard.controlDown
}

func (g *Game) throwKeyPressed() bool {
	return g.keyboard.throwDown
}
