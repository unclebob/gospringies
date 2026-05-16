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

type Game struct {
	simulation      *sim.Simulation
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

type RenderResult struct {
	Completed                  bool
	Representations            map[string]string
	SpringLinesVisible         bool
	MassesVisible              bool
	FixedMassDistinguishable   bool
	FixedMassRepresentation    string
	MovableMassRepresentation  string
	SelectedMassRepresentation string
}

func NewGame() *Game {
	return &Game{simulation: sim.NewWorld(), mode: "select"}
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
	screen.Fill(color.RGBA{R: 18, G: 20, B: 24, A: 255})
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

func (g *Game) RenderWorld() RenderResult {
	g.RenderFrame()
	representations := g.renderRepresentations()
	hasMovable := representations["movable mass"] != ""
	hasFixed := representations["fixed mass"] != ""
	hasSpring := representations["spring"] != ""
	return RenderResult{
		Completed:                  true,
		Representations:            representations,
		SpringLinesVisible:         hasSpring,
		MassesVisible:              hasMovable || hasFixed,
		FixedMassDistinguishable:   hasMovable && hasFixed,
		FixedMassRepresentation:    "red circle",
		MovableMassRepresentation:  "yellow circle",
		SelectedMassRepresentation: "selection outline",
	}
}

func (g *Game) renderRepresentations() map[string]string {
	representations := map[string]string{}
	g.massRepresentations(representations)
	g.springRepresentation(representations)
	g.wallRepresentation(representations)
	g.selectionRepresentation(representations)
	return representations
}

func (g *Game) springRepresentation(representations map[string]string) {
	if len(g.simulation.Springs) > 0 && g.showSprings() {
		representations["spring"] = "cyan line"
	}
}

func (g *Game) wallRepresentation(representations map[string]string) {
	if g.hasEnabledWall() {
		representations["enabled wall"] = "boundary line"
	}
}

func (g *Game) selectionRepresentation(representations map[string]string) {
	if g.selected {
		representations["selection"] = "selection outline"
	}
}

func (r RenderResult) HasVisibleRepresentation(object string) bool {
	if r.Representations == nil {
		return false
	}
	return r.Representations[object] != ""
}

func (g *Game) drawSprings(screen *ebiten.Image) {
	for _, spring := range g.simulation.Springs {
		if !g.validSpring(spring) {
			continue
		}
		a := g.simulation.Masses[spring.A].Position
		b := g.simulation.Masses[spring.B].Position
		ebitenutil.DrawLine(screen, a.X, a.Y, b.X, b.Y, color.RGBA{R: 116, G: 190, B: 222, A: 255})
	}
}

func (g *Game) drawMasses(screen *ebiten.Image) {
	for _, mass := range g.simulation.Masses {
		c := color.RGBA{R: 238, G: 212, B: 96, A: 255}
		if mass.Fixed {
			c = color.RGBA{R: 238, G: 116, B: 96, A: 255}
		}
		x, y, radius := massDrawCircle(mass)
		vector.DrawFilledCircle(screen, x, y, radius, c, true)
	}
}

func massDrawCircle(mass sim.Mass) (float32, float32, float32) {
	return float32(mass.Position.X), float32(mass.Position.Y), 5
}

func (g *Game) drawWalls(screen *ebiten.Image) {
	bounds := g.simulation.Bounds
	wallColor := color.RGBA{R: 180, G: 186, B: 196, A: 255}
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
	if len(g.simulation.Masses) == 0 {
		return
	}
	mass := g.simulation.Masses[0]
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	x := mass.Position.X
	y := mass.Position.Y
	ebitenutil.DrawLine(screen, x-8, y-8, x+8, y-8, c)
	ebitenutil.DrawLine(screen, x+8, y-8, x+8, y+8, c)
	ebitenutil.DrawLine(screen, x+8, y+8, x-8, y+8, c)
	ebitenutil.DrawLine(screen, x-8, y+8, x-8, y-8, c)
}

func (g *Game) validSpring(spring sim.Spring) bool {
	return spring.A >= 0 && spring.B >= 0 && spring.A < len(g.simulation.Masses) && spring.B < len(g.simulation.Masses)
}

func (g *Game) massRepresentations(representations map[string]string) {
	for _, mass := range g.simulation.Masses {
		if mass.Fixed {
			representations["fixed mass"] = "red circle"
		} else {
			representations["movable mass"] = "yellow circle"
		}
	}
}

func (g *Game) showSprings() bool {
	return g.simulation.Parameters.Value("show springs") == "true"
}

func (g *Game) hasEnabledWall() bool {
	for _, enabled := range g.simulation.Parameters.Walls {
		if enabled {
			return true
		}
	}
	return false
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
