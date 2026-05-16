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

func TestMassDrawRectCentersOnMassPosition(t *testing.T) {
	x, y, width, height := massDrawRect(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}})

	if x != 25 || y != 35 || width != 10 || height != 10 {
		t.Fatalf("draw rect = %f,%f %fx%f", x, y, width, height)
	}
}

func TestWindowConfigIsResizable(t *testing.T) {
	config := DefaultWindowConfig()
	if !config.Resizable {
		t.Fatal("window should be resizable")
	}
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
