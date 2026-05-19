package app

import (
	"image"
	"os"
)

const demoPickerRowHeight = 24

func demoPickerRect() image.Rectangle {
	return image.Rect(240, 96, screenWidth-240, screenHeight-96)
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

func clampInt(value int, lower int, upper int) int {
	return min(max(value, lower), upper)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T11:59:45-05:00","module_hash":"db754275dfee5ca40b22f7f514530596298e8783ea8b6b37087992f758bef4d2","functions":[{"id":"func/demoPickerRect","name":"demoPickerRect","line":10,"end_line":12,"hash":"fa11d90071c866ffac72f54b4199930a1eb26f85a8eba83b29de7aea8e326444"},{"id":"func/Game.visibleDemoPaths","name":"Game.visibleDemoPaths","line":14,"end_line":19,"hash":"fe38d9297f011742c3001905cf92c6d2f2ee67f34e420c492b6b1a328fa26abc"},{"id":"func/demoPickerVisibleRows","name":"demoPickerVisibleRows","line":21,"end_line":24,"hash":"aded365a5732d1b72592861b7be0d57919432368913a622ae5d1bb4c5088fb3b"},{"id":"func/Game.demoRowRect","name":"Game.demoRowRect","line":26,"end_line":30,"hash":"6163b232a6b823d15f5a2b8cb7b5efe92ffb240351dec15162858de5bfcafe57"},{"id":"func/Game.clickDemoPicker","name":"Game.clickDemoPicker","line":32,"end_line":45,"hash":"670c2f7e66872a3da9487c8ad7307cfedabc7721862e56070803275e1b3e6d0e"},{"id":"func/Game.loadDemoAt","name":"Game.loadDemoAt","line":47,"end_line":57,"hash":"df3c95c8bf486bc97bf1c52e4809196ac863a0368808026def9110f0f20bfb1f"},{"id":"func/clampInt","name":"clampInt","line":59,"end_line":61,"hash":"cad97290430d2f2ba73bc78b479d047fe3c47343fd006a522712ea5683891776"}]}
// mutate4go-manifest-end
