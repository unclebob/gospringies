//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) pollTextDialogKeyboard(open bool, appendInput func([]rune), backspace func(), submit func(), cancel func()) {
	if !open {
		return
	}
	appendInput(ebiten.AppendInputChars(nil))
	handleTextDialogControlKeys(backspace, submit, cancel)
}

func handleTextDialogControlKeys(backspace func(), submit func(), cancel func()) {
	backspace()
	submit()
	cancel()
}

func handleBackspaceKey(action func()) {
	handleJustPressedKey(ebiten.KeyBackspace, action)
}

func handleSubmitKey(action func()) {
	if valueDialogSubmitPressed() {
		action()
	}
}

func handleEscapeKey(action func()) {
	handleJustPressedKey(ebiten.KeyEscape, action)
}

func handleJustPressedKey(key ebiten.Key, action func()) {
	if inpututil.IsKeyJustPressed(key) {
		action()
	}
}

func valueDialogSubmitPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)
}
