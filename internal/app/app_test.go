package app

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestGameLayoutAndUpdate(t *testing.T) {
	game := NewGame()

	width, height := game.Layout(1, 1)
	if width != screenWidth || height != screenHeight {
		t.Fatalf("layout = %d, %d", width, height)
	}
	if err := game.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

func TestGameDraw(t *testing.T) {
	game := NewGame()
	screen := ebiten.NewImage(screenWidth, screenHeight)

	game.Draw(screen)
}
