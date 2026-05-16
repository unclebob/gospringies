package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"springs/internal/sim"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

var (
	backgroundColor = color.RGBA{R: 18, G: 20, B: 24, A: 255}
	springColor     = color.RGBA{R: 116, G: 190, B: 222, A: 255}
	massColor       = color.RGBA{R: 238, G: 212, B: 96, A: 255}
	fixedMassColor  = color.RGBA{R: 238, G: 116, B: 96, A: 255}
	wallColor       = color.RGBA{R: 180, G: 186, B: 196, A: 255}
	selectionColor  = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

type Game struct {
	simulation      *sim.Simulation
	initialState    *sim.Simulation
	savedState      *sim.Simulation
	mode            string
	selected        bool
	dirty           bool
	lastCommand     string
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
	world := sim.NewWorld()
	return &Game{simulation: world, initialState: world.Clone(), mode: "select"}
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
	result := g.RenderWorld()
	screen.Fill(backgroundColor)
	if result.SpringLinesVisible {
		g.drawSprings(screen)
	}
	g.drawMasses(screen)
	g.drawWalls(screen)
	if g.selected {
		g.drawSelection(screen)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS %.0f", ebiten.ActualTPS()))
}

func (g *Game) drawSprings(screen *ebiten.Image) {
	for _, spring := range g.simulation.Springs {
		if !g.validSpring(spring) {
			continue
		}
		a := g.simulation.Masses[spring.A].Position
		b := g.simulation.Masses[spring.B].Position
		ebitenutil.DrawLine(screen, a.X, a.Y, b.X, b.Y, springColor)
	}
}

func (g *Game) drawMasses(screen *ebiten.Image) {
	for _, mass := range g.simulation.Masses {
		x, y, radius := massDrawCircle(mass)
		vector.DrawFilledCircle(screen, x, y, radius, massDrawColor(mass), true)
	}
}

func massDrawCircle(mass sim.Mass) (float32, float32, float32) {
	return float32(mass.Position.X), float32(mass.Position.Y), 5
}

func massDrawColor(mass sim.Mass) color.RGBA {
	if mass.Fixed {
		return fixedMassColor
	}
	return massColor
}

func (g *Game) drawWalls(screen *ebiten.Image) {
	bounds := g.simulation.Bounds
	drawWallLine := func(name string, x1, y1, x2, y2 float64) {
		if enabled, _ := g.simulation.Parameters.WallEnabled(name); enabled {
			ebitenutil.DrawLine(screen, x1, y1, x2, y2, wallColor)
		}
	}
	drawWallLine("top", 0, 0, bounds.Width, 0)
	drawWallLine("bottom", 0, bounds.Height-1, bounds.Width, bounds.Height-1)
	drawWallLine("left", 0, 0, 0, bounds.Height)
	drawWallLine("right", bounds.Width-1, 0, bounds.Width-1, bounds.Height)
}

func (g *Game) drawSelection(screen *ebiten.Image) {
	for _, line := range selectedMassOutline(g.simulation.Masses) {
		ebitenutil.DrawLine(screen, line.x1, line.y1, line.x2, line.y2, selectionColor)
	}
}

type selectionLine struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

func selectedMassOutline(masses []sim.Mass) []selectionLine {
	if len(masses) == 0 {
		return nil
	}
	return selectionOutline(masses[0])
}

func selectionOutline(mass sim.Mass) []selectionLine {
	x := mass.Position.X
	y := mass.Position.Y
	return []selectionLine{
		{x - 8, y - 8, x + 8, y - 8},
		{x + 8, y - 8, x + 8, y + 8},
		{x + 8, y + 8, x - 8, y + 8},
		{x - 8, y + 8, x - 8, y - 8},
	}
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
