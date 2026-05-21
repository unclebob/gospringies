//go:build !appunit

package app

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func drawContextMenu(screen *ebiten.Image, rect image.Rectangle, title string, items []contextMenuItem) {
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, title, rect.Min.X+8, rect.Min.Y+4)
	for i, item := range items {
		row := contextMenuRowRect(rect, i)
		fill := controlColor
		if i%2 == 1 {
			fill = sectionColor
		}
		vector.DrawFilledRect(screen, float32(row.Min.X), float32(row.Min.Y), float32(row.Dx()), float32(row.Dy()), fill, false)
		ebitenutil.DebugPrintAt(screen, item.Label, row.Min.X+8, row.Min.Y+4)
	}
}
