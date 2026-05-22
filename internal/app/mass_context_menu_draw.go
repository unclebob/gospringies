//go:build !appunit

package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) drawMassContextMenu(screen *ebiten.Image) {
	drawContextMenu(screen, g.massContextMenuRect(), fmt.Sprintf("Mass #%d", g.overlays.massMenu.MassID), g.massContextMenuItems())
}
