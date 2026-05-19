package app

import (
	"image"
	"os"
	"path/filepath"
	"sort"
)

const demoPickerRowHeight = 24
const loadPickerSeparator = "separator"

func demoPickerRect() image.Rectangle {
	return image.Rect(240, 96, screenWidth-240, screenHeight-96)
}

func (g *Game) visibleDemoPaths() []string {
	files := g.demoList()
	start := clampInt(g.demoPickerScroll, 0, len(files))
	end := clampInt(start+demoPickerVisibleRows(), start, len(files))
	return files[start:end]
}

func (g *Game) LoadPickerEntries() []string {
	return append([]string{}, g.demoList()...)
}

func (g *Game) ChooseLoadPickerEntry(name string) bool {
	for i, path := range g.demoList() {
		if path == loadPickerSeparator {
			continue
		}
		if path == name || filepath.Base(path) == name {
			return g.loadDemoAt(i)
		}
	}
	return false
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

func (g *Game) loadDemoAt(index int) bool {
	files := g.demoList()
	if index < 0 || index >= len(files) {
		return false
	}
	path := files[index]
	if path == loadPickerSeparator {
		return false
	}
	content, err := os.ReadFile(path)
	if err == nil {
		err = g.LoadXSPFromFile(path, string(content))
	}
	if err != nil {
		g.lastFileError = err.Error()
	} else {
		g.lastFileError = ""
	}
	g.demoPickerOpen = false
	return err == nil
}

func (g *Game) buildDemoList() []string {
	var saves, starters, originals []string
	for _, root := range []string{".", filepath.Join("..", "..")} {
		saves = append(saves, globXSP(filepath.Join(root, "saves"))...)
		starters = append(starters, globXSP(filepath.Join(root, "demos"))...)
		originals = append(originals, globXSP(filepath.Join(root, "demos", "original"))...)
	}
	sort.Strings(saves)
	sort.Strings(starters)
	sort.Strings(originals)
	return groupedLoadPickerEntries(saves, starters, originals)
}

func globXSP(dir string) []string {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.xsp"))
	return matches
}

func groupedLoadPickerEntries(saves []string, starters []string, originals []string) []string {
	var entries []string
	entries = append(entries, saves...)
	if len(saves) > 0 {
		entries = append(entries, loadPickerSeparator)
	}
	entries = append(entries, starters...)
	return append(entries, originals...)
}

func clampInt(value int, lower int, upper int) int {
	return min(max(value, lower), upper)
}

func (g *Game) LastFileError() string {
	return g.lastFileError
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T16:19:15-05:00","module_hash":"cc07b447aa70dc332d2a1312b912a9b51589935e2c2e829761771af73fef4d15","functions":[{"id":"func/demoPickerRect","name":"demoPickerRect","line":13,"end_line":15,"hash":"fa11d90071c866ffac72f54b4199930a1eb26f85a8eba83b29de7aea8e326444"},{"id":"func/Game.visibleDemoPaths","name":"Game.visibleDemoPaths","line":17,"end_line":22,"hash":"fe38d9297f011742c3001905cf92c6d2f2ee67f34e420c492b6b1a328fa26abc"},{"id":"func/Game.LoadPickerEntries","name":"Game.LoadPickerEntries","line":24,"end_line":26,"hash":"261915035d4361d54dba5cc1d5e95fc3457e1c91a4ad34d5e379a31d3a9f8d8e"},{"id":"func/Game.ChooseLoadPickerEntry","name":"Game.ChooseLoadPickerEntry","line":28,"end_line":38,"hash":"89cd2f5f63031fce320c1e63630ecc990e9b638fd8eb228693855a591c0e6fd3"},{"id":"func/demoPickerVisibleRows","name":"demoPickerVisibleRows","line":40,"end_line":43,"hash":"aded365a5732d1b72592861b7be0d57919432368913a622ae5d1bb4c5088fb3b"},{"id":"func/Game.demoRowRect","name":"Game.demoRowRect","line":45,"end_line":49,"hash":"6163b232a6b823d15f5a2b8cb7b5efe92ffb240351dec15162858de5bfcafe57"},{"id":"func/Game.clickDemoPicker","name":"Game.clickDemoPicker","line":51,"end_line":64,"hash":"670c2f7e66872a3da9487c8ad7307cfedabc7721862e56070803275e1b3e6d0e"},{"id":"func/Game.loadDemoAt","name":"Game.loadDemoAt","line":66,"end_line":86,"hash":"fadc53098e950a41c7f708b6703bb272276d69f23a05a21390ee7082d0ee7326"},{"id":"func/Game.buildDemoList","name":"Game.buildDemoList","line":88,"end_line":99,"hash":"1d7ed3497b76ba7fe56277660f4d06c376eb11c53eb0037c7006331714bdf2bf"},{"id":"func/globXSP","name":"globXSP","line":101,"end_line":104,"hash":"d7c08b15c2dcc170661f5c18ca03d0a61b68d5a7074be3ac4deba4793fe166e4"},{"id":"func/groupedLoadPickerEntries","name":"groupedLoadPickerEntries","line":106,"end_line":114,"hash":"4c379e6bb437d8aa4f67ea40472e500bbb53b1adddf79fc0ee18716c3c23cc1c"},{"id":"func/clampInt","name":"clampInt","line":116,"end_line":118,"hash":"cad97290430d2f2ba73bc78b479d047fe3c47343fd006a522712ea5683891776"},{"id":"func/Game.LastFileError","name":"Game.LastFileError","line":120,"end_line":122,"hash":"97cb1118d4e48b921c525e4801ab58910d600da4519d91dc6ffc154cf680cc47"}]}
// mutate4go-manifest-end
