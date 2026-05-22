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
	start := clampInt(g.controls.demoPickerScroll, 0, len(files))
	end := clampInt(start+demoPickerVisibleRows(), start, len(files))
	return files[start:end]
}

func (g *Game) LoadPickerEntries() []string {
	return append([]string{}, g.demoList()...)
}

func (g *Game) ChooseLoadPickerEntry(name string) bool {
	for i, path := range g.demoList() {
		if loadPickerEntryMatches(path, name) {
			return g.loadDemoAt(i)
		}
	}
	return false
}

func loadPickerEntryMatches(path string, name string) bool {
	return path != loadPickerSeparator && (path == name || filepath.Base(path) == name)
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
		g.controls.demoPickerOpen = false
		return
	}
	for i := range g.visibleDemoPaths() {
		if point.In(g.demoRowRect(i)) {
			g.loadDemoAt(g.controls.demoPickerScroll + i)
			return
		}
	}
}

func (g *Game) loadDemoAt(index int) bool {
	path, ok := g.demoPathAt(index)
	if !ok {
		return false
	}
	err := g.loadDemoPath(path)
	g.recordDemoLoadResult(err)
	g.controls.demoPickerOpen = false
	return err == nil
}

func (g *Game) demoPathAt(index int) (string, bool) {
	files := g.demoList()
	if index < 0 || index >= len(files) || files[index] == loadPickerSeparator {
		return "", false
	}
	return files[index], true
}

func (g *Game) loadDemoPath(path string) error {
	content, err := os.ReadFile(path)
	if err == nil {
		err = g.LoadXSPFromFile(path, string(content))
	}
	return err
}

func (g *Game) recordDemoLoadResult(err error) {
	if err != nil {
		g.document.lastFileError = err.Error()
	} else {
		g.document.lastFileError = ""
	}
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
	return g.document.lastFileError
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:11:49-05:00","module_hash":"a10f98430f0297e797cce7aa4f9a62088ff0543a343533847e0312d620092787","functions":[{"id":"func/demoPickerRect","name":"demoPickerRect","line":13,"end_line":15,"hash":"fa11d90071c866ffac72f54b4199930a1eb26f85a8eba83b29de7aea8e326444"},{"id":"func/Game.visibleDemoPaths","name":"Game.visibleDemoPaths","line":17,"end_line":22,"hash":"909e7b5768a8ca470e81c1c85cd43150ae0510e3c4c55f711a60d476a82cf42d"},{"id":"func/Game.LoadPickerEntries","name":"Game.LoadPickerEntries","line":24,"end_line":26,"hash":"261915035d4361d54dba5cc1d5e95fc3457e1c91a4ad34d5e379a31d3a9f8d8e"},{"id":"func/Game.ChooseLoadPickerEntry","name":"Game.ChooseLoadPickerEntry","line":28,"end_line":35,"hash":"0e74fcaf34a0416055e8693475c42100cd248e5be03af251ae27f391b0871633"},{"id":"func/loadPickerEntryMatches","name":"loadPickerEntryMatches","line":37,"end_line":39,"hash":"fb616d9592160b947cdc0a0409e27346938db0ded1e80f03ebb2378155e8dcf3"},{"id":"func/demoPickerVisibleRows","name":"demoPickerVisibleRows","line":41,"end_line":44,"hash":"aded365a5732d1b72592861b7be0d57919432368913a622ae5d1bb4c5088fb3b"},{"id":"func/Game.demoRowRect","name":"Game.demoRowRect","line":46,"end_line":50,"hash":"6163b232a6b823d15f5a2b8cb7b5efe92ffb240351dec15162858de5bfcafe57"},{"id":"func/Game.clickDemoPicker","name":"Game.clickDemoPicker","line":52,"end_line":65,"hash":"daae2cd02ede4aaebf7aea1606057a3851f090bcad9e00f466fe38df5f15eb92"},{"id":"func/Game.loadDemoAt","name":"Game.loadDemoAt","line":67,"end_line":76,"hash":"7185d98a935941de6ae386177c66a990714076753ffe292d5f459f5dec769ead"},{"id":"func/Game.demoPathAt","name":"Game.demoPathAt","line":78,"end_line":84,"hash":"88c632d57e637067eed470c0ee9dd29ee78786d579c50d33c0415744d91cd99c"},{"id":"func/Game.loadDemoPath","name":"Game.loadDemoPath","line":86,"end_line":92,"hash":"dc8f8ffa68c9dcf7fd8f88ccb6273f047433006f7661d748854a5331800e5074"},{"id":"func/Game.recordDemoLoadResult","name":"Game.recordDemoLoadResult","line":94,"end_line":100,"hash":"63a933f784c9415068442184036b6a893ba6fa5b00026f582809cbe014e0f74b"},{"id":"func/Game.buildDemoList","name":"Game.buildDemoList","line":102,"end_line":113,"hash":"1d7ed3497b76ba7fe56277660f4d06c376eb11c53eb0037c7006331714bdf2bf"},{"id":"func/globXSP","name":"globXSP","line":115,"end_line":118,"hash":"d7c08b15c2dcc170661f5c18ca03d0a61b68d5a7074be3ac4deba4793fe166e4"},{"id":"func/groupedLoadPickerEntries","name":"groupedLoadPickerEntries","line":120,"end_line":128,"hash":"4c379e6bb437d8aa4f67ea40472e500bbb53b1adddf79fc0ee18716c3c23cc1c"},{"id":"func/clampInt","name":"clampInt","line":130,"end_line":132,"hash":"cad97290430d2f2ba73bc78b479d047fe3c47343fd006a522712ea5683891776"},{"id":"func/Game.LastFileError","name":"Game.LastFileError","line":134,"end_line":136,"hash":"8a1ba5f2f318c751ac7fc2e0895404769bbfc6e3e58da6e94aa441b59f797b20"}]}
// mutate4go-manifest-end
