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
	simulation      *sim.Simulation
	paused          bool
	inputActive     bool
	renderingActive bool
	closed          bool
}

type WindowConfig struct {
	Width     int
	Height    int
	Title     string
	Resizable bool
}

func NewGame() *Game {
	return &Game{simulation: sim.NewWorld()}
}

func DefaultWindowConfig() WindowConfig {
	return WindowConfig{Width: screenWidth, Height: screenHeight, Title: "Springs", Resizable: true}
}

func Run() error {
	config := DefaultWindowConfig()
	ebiten.SetWindowSize(config.Width, config.Height)
	ebiten.SetWindowTitle(config.Title)
	if config.Resizable {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	}
	return ebiten.RunGame(NewGame())
}

func (g *Game) Update() error {
	g.inputActive = true
	if !g.paused {
		g.simulation.Step(1.0 / 60.0)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.RenderFrame()
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

func (g *Game) World() *sim.Simulation {
	return g.simulation
}

func (g *Game) SetPaused(paused bool) {
	g.paused = paused
}

func (g *Game) Paused() bool {
	return g.paused
}

func (g *Game) InputActive() bool {
	return g.inputActive
}

func (g *Game) RenderingActive() bool {
	return g.renderingActive
}

func (g *Game) RenderFrame() {
	g.renderingActive = true
}

func (g *Game) Close() error {
	g.closed = true
	return nil
}

func (g *Game) Closed() bool {
	return g.closed
}
