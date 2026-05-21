//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) pollNumericTextFieldKeyboard() {
	if g.focusedNumeric == "" {
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

func (g *Game) handleNumericTextFieldBlur() {
	if numericTextFieldBlurPressed() {
		g.focusedNumeric = ""
	}
}

func numericTextFieldBlurPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape) || valueDialogSubmitPressed()
}
