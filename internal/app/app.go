package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"springs/internal/sim"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	simulation *sim.Simulation
}

func NewGame() *Game {
	return &Game{simulation: sim.NewDemoSimulation()}
}

func Run() error {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Springs")
	return ebiten.RunGame(NewGame())
}

func (g *Game) Update() error {
	g.simulation.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 18, G: 20, B: 24, A: 255})
	for _, spring := range g.simulation.Springs {
		a := g.simulation.Masses[spring.A].Position
		b := g.simulation.Masses[spring.B].Position
		ebitenutil.DrawLine(screen, a.X, a.Y, b.X, b.Y, color.RGBA{R: 116, G: 190, B: 222, A: 255})
	}
	for _, mass := range g.simulation.Masses {
		c := color.RGBA{R: 238, G: 212, B: 96, A: 255}
		if mass.Fixed {
			c = color.RGBA{R: 238, G: 116, B: 96, A: 255}
		}
		ebitenutil.DrawRect(screen, mass.Position.X-5, mass.Position.Y-5, 10, 10, c)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS %.0f", ebiten.ActualTPS()))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}
