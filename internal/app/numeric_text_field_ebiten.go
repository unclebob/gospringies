//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
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
	handleBackspaceKey(g.deleteNumericSettingCharacter)
}

func (g *Game) handleNumericTextFieldCancel() {
	handleEscapeKey(g.cancelNumericSettingInput)
}

func (g *Game) handleNumericTextFieldSubmit() {
	handleSubmitKey(func() { g.commitNumericSettingInput() })
}

func (g *Game) handleNumericTextFieldBlur() {
	g.handleNumericTextFieldCancel()
	g.handleNumericTextFieldSubmit()
}
