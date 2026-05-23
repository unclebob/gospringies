//go:build !appunit

package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawDemoPicker(screen *ebiten.Image) {
	rect := demoPickerRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, demoPickerPanelAntiAlias())
	title := demoPickerTitlePoint(rect)
	ebitenutil.DebugPrintAt(screen, "Load Demo", title.X, title.Y)
	for i, path := range g.visibleDemoPaths() {
		row := g.demoRowRect(i)
		vector.DrawFilledRect(screen, float32(row.Min.X), float32(row.Min.Y), float32(row.Dx()), float32(row.Dy()), demoPickerRowFill(i), demoPickerRowAntiAlias())
		label := demoPickerRowTextPoint(row)
		ebitenutil.DebugPrintAt(screen, path, label.X, label.Y)
	}
}

func demoPickerPanelAntiAlias() bool {
	return false
}

func demoPickerRowAntiAlias() bool {
	return false
}
