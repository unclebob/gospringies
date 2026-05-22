//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) pollNumericTextFieldKeyboard() {
	if g.controls.focusedNumeric == "" {
		return
	}
	g.appendNumericSettingInput(ebiten.AppendInputChars(nil))
	g.handleNumericTextFieldControlKeys()
}

func (g *Game) handleNumericTextFieldControlKeys() {
	g.handleNumericTextFieldBackspace()
	g.handleNumericTextFieldBlur()
}

func (g *Game) handleNumericTextFieldBackspace() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.deleteNumericSettingCharacter()
	}
}

func (g *Game) handleNumericTextFieldCancel() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.cancelNumericSettingInput()
	}
}

func (g *Game) handleNumericTextFieldSubmit() {
	if valueDialogSubmitPressed() {
		g.commitNumericSettingInput()
	}
}

func (g *Game) handleNumericTextFieldBlur() {
	g.handleNumericTextFieldCancel()
	g.handleNumericTextFieldSubmit()
}
