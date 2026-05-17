package app

import (
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const demoPickerRowHeight = 24

func demoPickerRect() image.Rectangle {
	return image.Rect(240, 96, screenWidth-240, screenHeight-96)
}

func (g *Game) drawDemoPicker(screen *ebiten.Image) {
	rect := demoPickerRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, "Load Demo", rect.Min.X+12, rect.Min.Y+10)
	for i, path := range g.visibleDemoPaths() {
		row := g.demoRowRect(i)
		fill := controlColor
		if i%2 == 1 {
			fill = sectionColor
		}
		vector.DrawFilledRect(screen, float32(row.Min.X), float32(row.Min.Y), float32(row.Dx()), float32(row.Dy()), fill, false)
		ebitenutil.DebugPrintAt(screen, path, row.Min.X+8, row.Min.Y+4)
	}
}

func (g *Game) visibleDemoPaths() []string {
	files := g.demoList()
	start := clampInt(g.demoPickerScroll, 0, len(files))
	end := clampInt(start+demoPickerVisibleRows(), start, len(files))
	return files[start:end]
}

func demoPickerVisibleRows() int {
	rect := demoPickerRect()
	return (rect.Dy() - 48) / demoPickerRowHeight
}

func (g *Game) demoRowRect(visibleIndex int) image.Rectangle {
	rect := demoPickerRect()
	top := rect.Min.Y + 40 + visibleIndex*demoPickerRowHeight
	return image.Rect(rect.Min.X+12, top, rect.Max.X-12, top+demoPickerRowHeight-2)
}

func (g *Game) clickDemoPicker(x int, y int) {
	point := image.Pt(x, y)
	rect := demoPickerRect()
	if !point.In(rect) {
		g.demoPickerOpen = false
		return
	}
	for i := range g.visibleDemoPaths() {
		if point.In(g.demoRowRect(i)) {
			g.loadDemoAt(g.demoPickerScroll + i)
			return
		}
	}
}

func (g *Game) loadDemoAt(index int) {
	files := g.demoList()
	if index < 0 || index >= len(files) {
		return
	}
	content, err := os.ReadFile(files[index])
	if err == nil {
		_ = g.LoadXSP(string(content))
	}
	g.demoPickerOpen = false
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
