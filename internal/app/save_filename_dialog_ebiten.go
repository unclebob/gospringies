//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawSaveFilenameDialog(screen *ebiten.Image) {
	rect := saveFilenameDialogRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, "Save", rect.Min.X+12, rect.Min.Y+10)
	drawLabeledRect(screen, g.saveFilenameTextRect(), controlColor, g.saveDialog.Text)
	g.drawSaveFilenameCursor(screen)
	drawLabeledRect(screen, g.saveFilenameDialogOKRect(), activeControlColor, "OK")
}

func (g *Game) drawSaveFilenameCursor(screen *ebiten.Image) {
	rect := g.saveFilenameTextRect()
	cursor := clampInt(g.saveDialog.Cursor, 0, len(g.saveDialog.Text))
	x := rect.Min.X + 4 + cursor*debugGlyphWidth
	if x > rect.Max.X-6 {
		x = rect.Max.X - 6
	}
	vector.DrawFilledRect(screen, float32(x), float32(rect.Min.Y+4), 2, float32(debugGlyphHeight-2), selectionColor, false)
}

func (g *Game) pollSaveFilenameDialogKeyboard() {
	if !g.saveDialog.Open {
		return
	}
	g.insertSaveFilenameText(string(ebiten.AppendInputChars(nil)))
	g.handleSaveFilenameDialogControlKeys()
}

func (g *Game) handleSaveFilenameDialogControlKeys() {
	runIfPressed(func() bool { return inpututil.IsKeyJustPressed(ebiten.KeyBackspace) }, g.deleteSaveFilenameCharacter)
	runIfPressed(valueDialogSubmitPressed, func() { _ = g.SubmitSaveFilenameDialog() })
	runIfPressed(func() bool { return inpututil.IsKeyJustPressed(ebiten.KeyEscape) }, func() { g.saveDialog.Open = false })
}
