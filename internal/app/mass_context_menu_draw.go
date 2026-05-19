//go:build !appunit

package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawMassContextMenu(screen *ebiten.Image) {
	rect := g.massContextMenuRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mass #%d", g.massMenu.MassID), rect.Min.X+8, rect.Min.Y+4)
	for i, item := range g.massContextMenuItems() {
		row := g.massContextMenuRowRect(i)
		fill := controlColor
		if i%2 == 1 {
			fill = sectionColor
		}
		vector.DrawFilledRect(screen, float32(row.Min.X), float32(row.Min.Y), float32(row.Dx()), float32(row.Dy()), fill, false)
		ebitenutil.DebugPrintAt(screen, item.Label, row.Min.X+8, row.Min.Y+4)
	}
}
