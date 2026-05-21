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
	runIfPressed(func() bool { return inpututil.IsKeyJustPressed(ebiten.KeyBackspace) }, g.deleteNumericSettingCharacter)
	runIfPressed(numericTextFieldBlurPressed, func() { g.focusedNumeric = "" })
}

func numericTextFieldBlurPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape) || valueDialogSubmitPressed()
}
