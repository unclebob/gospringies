package app

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"springs/internal/sim"
)

func TestNewGameStartsWithEmptyWorld(t *testing.T) {
	game := NewGame()

	if len(game.World().Masses) != 0 || len(game.World().Springs) != 0 {
		t.Fatalf("world = %#v", game.World())
	}
}

func TestGameLayoutUsesWindowSize(t *testing.T) {
	game := NewGame()
	width, height := game.Layout(1, 1)
	if width != screenWidth || height != screenHeight {
		t.Fatalf("layout = %d, %d", width, height)
	}
}

func TestGameUpdateStepsOnlyWhenUnpaused(t *testing.T) {
	game := NewGame()
	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})

	if err := game.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if game.World().Time == 0 {
		t.Fatal("expected unpaused update to advance simulation time")
	}

	game.SetPaused(true)
	pausedAt := game.World().Time
	if err := game.Update(); err != nil {
		t.Fatalf("paused Update returned error: %v", err)
	}
	if game.World().Time != pausedAt {
		t.Fatalf("paused time = %f, want %f", game.World().Time, pausedAt)
	}
	if !game.InputActive() {
		t.Fatal("expected input handling to remain active")
	}
}

func TestGameDraw(t *testing.T) {
	game := NewGame()
	left := game.World().AddMassAt(sim.Vec2{X: 10, Y: 10}, 1, true)
	right := game.World().AddMassAt(sim.Vec2{X: 30, Y: 10}, 1, false)
	game.World().AddSpringBetween(left, right, 20, 12)
	screen := ebiten.NewImage(screenWidth, screenHeight)

	game.Draw(screen)

	if !game.RenderingActive() {
		t.Fatal("expected rendering to be active")
	}
}

func TestRenderWorldCompletesForEmptyAndNonEmptyWorlds(t *testing.T) {
	for _, setup := range []func(*Game){func(*Game) {}, addRenderableSpring} {
		game := NewGame()
		setup(game)
		if result := game.RenderWorld(); !result.Completed {
			t.Fatalf("render result = %#v", result)
		}
	}
}

func TestMassDrawCircleCentersOnMassPosition(t *testing.T) {
	x, y, radius := massDrawCircle(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}})

	if x != 30 || y != 40 || radius != 5 {
		t.Fatalf("draw circle = %f,%f radius %f", x, y, radius)
	}
}

func TestRenderWorldReportsVisibleObjects(t *testing.T) {
	game := NewGame()
	addRenderableSpring(game)
	game.World().Parameters.EnableWall("left")
	game.SetSelected(true)

	result := game.RenderWorld()

	for _, object := range []string{"movable mass", "fixed mass", "spring", "enabled wall", "selection"} {
		if !result.HasVisibleRepresentation(object) {
			t.Fatalf("missing representation for %q: %#v", object, result.Representations)
		}
	}
}

func TestShowSpringsControlsSpringVisibility(t *testing.T) {
	game := NewGame()
	addRenderableSpring(game)
	game.World().Parameters.Set("show springs", "false")

	result := game.RenderWorld()

	if result.SpringLinesVisible {
		t.Fatal("expected spring lines to be hidden")
	}
	if !result.MassesVisible {
		t.Fatal("expected masses to remain visible")
	}
}

func TestFixedMassesAreDistinguishable(t *testing.T) {
	game := NewGame()
	addRenderableSpring(game)

	result := game.RenderWorld()

	if !result.FixedMassDistinguishable {
		t.Fatalf("render result = %#v", result)
	}
	if result.FixedMassRepresentation == result.MovableMassRepresentation {
		t.Fatal("expected distinct fixed and movable mass representations")
	}
}

func TestWindowConfigIsResizable(t *testing.T) {
	config := DefaultWindowConfig()
	if !config.Resizable {
		t.Fatal("window should be resizable")
	}
}

func addRenderableSpring(game *Game) {
	left := game.World().AddMassAt(sim.Vec2{X: 10, Y: 10}, 1, true)
	right := game.World().AddMassAt(sim.Vec2{X: 30, Y: 10}, 1, false)
	game.World().AddSpringBetween(left, right, 20, 12)
}

func TestRenderFrameMarksRenderingActive(t *testing.T) {
	game := NewGame()

	game.RenderFrame()

	if !game.RenderingActive() {
		t.Fatal("expected rendering to be active")
	}
}

func TestGameClosesCleanly(t *testing.T) {
	game := NewGame()
	if err := game.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !game.Closed() {
		t.Fatal("expected game to be closed")
	}
}

func TestEditorScreenHasRequiredRegions(t *testing.T) {
	screen := NewGame().EditorScreen()
	expected := map[string]string{
		"canvas":          "edit and view the simulation world",
		"left toolbar":    "choose editing modes",
		"top bar":         "run commands and file commands",
		"right inspector": "edit selected objects and world parameters",
		"status line":     "show mode, simulation state, counts, and file state",
	}

	if !screen.Editor || screen.LandingPage {
		t.Fatalf("screen = %#v", screen)
	}
	for region, purpose := range expected {
		if got, ok := screen.RegionPurpose(region); !ok || got != purpose {
			t.Fatalf("region %q = %q, %t", region, got, ok)
		}
	}
}

func TestEditorScreenHasVisibleModeAndCommandControls(t *testing.T) {
	screen := NewGame().EditorScreen()
	for _, mode := range []string{"select", "add mass", "add spring", "drag"} {
		if !screen.HasModeControl(mode) {
			t.Fatalf("missing mode %q", mode)
		}
	}
	for _, command := range []string{"run", "pause", "reset", "load", "insert", "save", "quit"} {
		if !screen.HasCommandControl(command) {
			t.Fatalf("missing command %q", command)
		}
	}
}

func TestEditorIndicatorsReflectState(t *testing.T) {
	game := NewGame()
	game.SetMode("select")
	game.SetPaused(true)
	game.SetSelected(true)
	game.SetDirty(true)

	indicators := game.EditorScreen().Indicators

	if indicators["active mode"] != "select mode" {
		t.Fatalf("active mode = %q", indicators["active mode"])
	}
	if indicators["simulation state"] != "paused" {
		t.Fatalf("simulation state = %q", indicators["simulation state"])
	}
	if indicators["selection"] != "object selected" {
		t.Fatalf("selection = %q", indicators["selection"])
	}
	if indicators["file state"] != "unsaved changes" {
		t.Fatalf("file state = %q", indicators["file state"])
	}
}

func TestKeyboardShortcutsRunVisibleCommands(t *testing.T) {
	game := NewGame()
	shortcuts := map[string]string{
		"Space":  "pause",
		"R":      "reset",
		"Ctrl+S": "save",
		"Ctrl+O": "load",
		"Ctrl+I": "insert",
		"Q":      "quit",
	}

	for shortcut, command := range shortcuts {
		if !game.EditorScreen().HasCommandControl(command) {
			t.Fatalf("missing visible command %q", command)
		}
		if !game.HandleShortcut(shortcut) {
			t.Fatalf("shortcut %q was not handled", shortcut)
		}
		if got := game.LastCommand(); got != command {
			t.Fatalf("shortcut %q ran %q, want %q", shortcut, got, command)
		}
	}
}

func TestEditorControlsRemainUsableWhilePausedOrRunning(t *testing.T) {
	game := NewGame()
	for _, paused := range []bool{true, false} {
		game.SetPaused(paused)
		screen := game.EditorScreen()
		if !screen.CanvasVisible || !screen.ControlsUsable {
			t.Fatalf("paused %t screen = %#v", paused, screen)
		}
	}
}
