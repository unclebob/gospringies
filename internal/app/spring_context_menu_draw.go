//go:build !appunit

package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) drawSpringContextMenu(screen *ebiten.Image) {
	drawContextMenu(screen, g.springContextMenuRect(), fmt.Sprintf("Spring #%d", g.overlays.springMenu.SpringID), g.springContextMenuItems())
}
