//go:build !appunit

package app

import (
	"image"
	"image/color"

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

func demoPickerTitlePoint(rect image.Rectangle) image.Point {
	return image.Pt(rect.Min.X+12, rect.Min.Y+10)
}

func demoPickerRowTextPoint(row image.Rectangle) image.Point {
	return image.Pt(row.Min.X+8, row.Min.Y+4)
}

func demoPickerRowFill(index int) color.RGBA {
	if index%2 == 1 {
		return sectionColor
	}
	return controlColor
}
