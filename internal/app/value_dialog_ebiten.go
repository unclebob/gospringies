//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawValueDialog(screen *ebiten.Image) {
	rect := valueDialogRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, g.valueDialog.Title, rect.Min.X+12, rect.Min.Y+10)
	drawLabeledRect(screen, g.valueDialogTextRect(), controlColor, g.valueDialog.Text)
	g.drawValueDialogCursor(screen)
	track := g.valueDialogSliderTrack()
	vector.DrawFilledRect(screen, float32(track.Min.X), float32(track.Min.Y), float32(track.Dx()), float32(track.Dy()), sectionColor, false)
	fill := track
	fill.Max.X = fill.Min.X + int(g.valueDialogFraction()*float64(track.Dx()))
	vector.DrawFilledRect(screen, float32(fill.Min.X), float32(fill.Min.Y), float32(fill.Dx()), float32(fill.Dy()), activeControlColor, false)
	drawLabeledRect(screen, g.valueDialogOKRect(), activeControlColor, "OK")
}

func (g *Game) drawValueDialogCursor(screen *ebiten.Image) {
	if !g.valueDialogCursorVisible() {
		return
	}
	rect := g.valueDialogTextRect()
	x := rect.Min.X + 4 + len(g.valueDialog.Text)*debugGlyphWidth
	if x > rect.Max.X-6 {
		x = rect.Max.X - 6
	}
	vector.DrawFilledRect(screen, float32(x), float32(rect.Min.Y+4), 2, float32(debugGlyphHeight-2), selectionColor, false)
}

func (g *Game) pollValueDialogKeyboard() {
	if !g.valueDialog.Open {
		return
	}
	g.appendValueDialogInput(ebiten.AppendInputChars(nil))
	g.handleValueDialogControlKeys()
}

func (g *Game) handleValueDialogControlKeys() {
	g.handleValueDialogBackspace()
	g.handleValueDialogSubmit()
	g.handleValueDialogCancel()
}

func (g *Game) handleValueDialogBackspace() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.deleteValueDialogCharacter()
	}
}

func (g *Game) handleValueDialogSubmit() {
	if valueDialogSubmitPressed() {
		g.applyValueDialog()
	}
}

func (g *Game) handleValueDialogCancel() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.valueDialog.Open = false
	}
}

func valueDialogSubmitPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)
}
