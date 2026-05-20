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
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.deleteNumericSettingCharacter()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.cancelNumericSettingInput()
	}
	if valueDialogSubmitPressed() {
		g.commitNumericSettingInput()
	}
}
