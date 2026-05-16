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

type EditorScreen struct {
	Editor          bool
	LandingPage     bool
	Regions         []ScreenRegion
	ModeControls    []string
	CommandControls []string
	Indicators      map[string]string
	CanvasVisible   bool
	ControlsUsable  bool
}

type ScreenRegion struct {
	Name    string
	Purpose string
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

func (g *Game) EditorScreen() EditorScreen {
	return EditorScreen{
		Editor:      true,
		LandingPage: false,
		Regions: []ScreenRegion{
			{"canvas", "edit and view the simulation world"},
			{"left toolbar", "choose editing modes"},
			{"top bar", "run commands and file commands"},
			{"right inspector", "edit selected objects and world parameters"},
			{"status line", "show mode, simulation state, counts, and file state"},
		},
		ModeControls:    []string{"select", "add mass", "add spring", "drag"},
		CommandControls: []string{"run", "pause", "reset", "load", "insert", "save", "quit"},
		Indicators: map[string]string{
			"active mode":      g.mode + " mode",
			"simulation state": g.simulationState(),
			"selection":        g.selectionState(),
			"file state":       g.fileState(),
		},
		CanvasVisible:  true,
		ControlsUsable: true,
	}
}

func (s EditorScreen) RegionPurpose(name string) (string, bool) {
	for _, region := range s.Regions {
		if region.Name == name {
			return region.Purpose, true
		}
	}
	return "", false
}

func (s EditorScreen) HasModeControl(mode string) bool {
	return contains(s.ModeControls, mode)
}

func (s EditorScreen) HasCommandControl(command string) bool {
	return contains(s.CommandControls, command)
}

func (g *Game) SetMode(mode string) {
	g.mode = mode
}

func (g *Game) SetSelected(selected bool) {
	g.selected = selected
}

func (g *Game) SetDirty(dirty bool) {
	g.dirty = dirty
}

func (g *Game) HandleShortcut(shortcut string) bool {
	command, ok := map[string]string{
		"Space":  "pause",
		"R":      "reset",
		"Ctrl+S": "save",
		"Ctrl+O": "load",
		"Ctrl+I": "insert",
		"Q":      "quit",
	}[shortcut]
	if !ok {
		return false
	}
	g.RunCommand(command)
	return true
}

func (g *Game) RunCommand(command string) {
	g.lastCommand = command
}

func (g *Game) LastCommand() string {
	return g.lastCommand
}

func (g *Game) simulationState() string {
	return stateLabel(g.paused, "paused", "running")
}

func (g *Game) selectionState() string {
	return stateLabel(g.selected, "object selected", "nothing selected")
}

func (g *Game) fileState() string {
	return stateLabel(g.dirty, "unsaved changes", "saved")
}

func stateLabel(condition bool, trueLabel string, falseLabel string) string {
	if condition {
		return trueLabel
	}
	return falseLabel
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
