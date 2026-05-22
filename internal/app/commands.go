package app

import (
	"springs/internal/appcore"
	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func (g *Game) RunCommand(command string) {
	g.lastCommand = command
	g.pathEntryCommand = ""
	action, ok := commandActions[command]
	if ok {
		action(g, command)
	}
}

var commandActions = map[string]func(*Game, string){
	"run":             func(g *Game, _ string) { g.paused = false },
	"pause":           func(g *Game, _ string) { g.paused = !g.paused },
	"pause toggle":    func(g *Game, _ string) { g.paused = !g.paused },
	"reset":           func(g *Game, _ string) { g.resetWorld() },
	"save state":      func(g *Game, _ string) { g.SaveState() },
	"restore state":   func(g *Game, _ string) { g.restoreWorldState() },
	"load":            func(g *Game, _ string) { g.openDemoPicker() },
	"insert":          func(g *Game, command string) { g.pathEntryCommand = pathEntryLabel(command) },
	"save":            func(g *Game, _ string) { g.openSaveFilenameDialog() },
	"select all":      func(g *Game, _ string) { g.selectAllObjects() },
	"clear selection": func(g *Game, _ string) { g.clearSelection() },
	"delete":          func(g *Game, _ string) { g.deleteSelection() },
	"cut":             func(g *Game, _ string) { g.cutSelection() },
	"copy":            func(g *Game, _ string) { g.copySelection() },
	"paste":           func(g *Game, _ string) { g.pasteAtCursor() },
	"duplicate":       func(g *Game, _ string) { g.duplicateSelection() },
	"quit":            func(g *Game, _ string) { _ = g.Close() },
}

func (g *Game) resetWorld() {
	g.simulation.Reset()
	g.dirty = false
}

func (g *Game) restoreWorldState() {
	g.RestoreState()
	g.dirty = true
}

func (g *Game) selectAllObjects() {
	g.editing().SelectAll()
	g.syncSelectionState()
}

func (g *Game) deleteSelection() {
	g.editing().DeleteSelected()
	g.selected = false
	g.dirty = true
}

func (g *Game) cutSelection() {
	g.copySelection()
	g.deleteSelection()
}

func (g *Game) pasteAtCursor() {
	if g.pasteSelectionAt(g.pointer.lastCursor) {
		g.selected = true
		g.dirty = true
	}
}

func (g *Game) duplicateSelection() {
	if _, err := g.editing().DuplicateSelected(); err == nil {
		g.selected = true
		g.dirty = true
	}
}

func (g *Game) clearSelection() {
	g.editing().ClearSelection()
	g.selected = false
}

func (g *Game) syncSelectionState() {
	g.selected = g.selectedObjectCount() > 0
}

func (g *Game) openDemoPicker() {
	g.controls.demoPickerOpen = true
	g.controls.demoPickerScroll = 0
	g.controls.demoFiles = nil
	g.demoList()
}

func pathEntryLabel(command string) string {
	return pathEntryLabels[command]
}

var pathEntryLabels = map[string]string{
	"load":   "Load",
	"insert": "Insert",
	"save":   "Save",
}

func (g *Game) SaveXSP() string {
	g.dirty = false
	return xspfmt.SaveXSP(g.simulation)
}

func (g *Game) LoadXSP(input string) error {
	loaded, err := xspfmt.LoadXSP(input)
	if err != nil {
		return err
	}
	setAppBounds(loaded)
	g.canvasYUp = xspfmt.UsesOriginalXSpringiesCoordinates(input)
	g.applyCanvasWallBounds(loaded)
	g.simulation.Reset()
	g.simulation.LoadFrom(loaded)
	if g.editor != nil {
		g.editor.World = g.simulation
	}
	g.selected = false
	g.dirty = false
	return nil
}

func (g *Game) LoadXSPFromFile(path string, input string) error {
	if err := g.LoadXSP(input); err != nil {
		return err
	}
	g.currentFilePath = path
	return nil
}

func (g *Game) InsertXSP(input string) error {
	inserted, err := xspfmt.LoadXSP(input)
	if err != nil {
		return err
	}
	setAppBounds(inserted)
	g.applyCanvasWallBounds(inserted)
	g.simulation.InsertFrom(inserted)
	g.dirty = true
	return nil
}

func (g *Game) SaveState() {
	g.savedState = g.simulation.Clone()
}

func (g *Game) RestoreState() {
	state := g.savedState
	if state == nil {
		state = g.initialState
	}
	g.simulation.LoadFrom(state)
}

func (g *Game) SetParameter(parameter string, value string) {
	g.simulation.Parameters.Set(parameter, value)
	g.dirty = true
}

func (g *Game) ReplaceWorld(world *sim.Simulation) {
	g.canvasYUp = false
	setAppBounds(world)
	g.applyCanvasWallBounds(world)
	g.simulation.LoadFrom(world)
	if g.editor != nil {
		g.editor.World = g.simulation
	}
}

func setAppBounds(world *sim.Simulation) {
	appcore.ApplyBounds(world, appBounds())
}

func appBounds() sim.Bounds {
	return sim.Bounds{Width: screenWidth, Height: screenHeight}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T07:57:49-05:00","module_hash":"fa96ae123302c343e89f0993b2ad7c7fa68d255f59cfded470a4a713e6e2e959","functions":[{"id":"func/Game.RunCommand","name":"Game.RunCommand","line":9,"end_line":16,"hash":"40c33d99c107c931962ad197020a08167e49bbc757b117c3e5109f76e4d6b312"},{"id":"func/Game.resetWorld","name":"Game.resetWorld","line":38,"end_line":41,"hash":"442a95b81114be9994aa54b476f87ef238ed9dc1ba5589b72c7fc24a7746c2b3"},{"id":"func/Game.restoreWorldState","name":"Game.restoreWorldState","line":43,"end_line":46,"hash":"2c7bfb10326dbbe774907dbeb63092bc76685521870c352d2d27d803f46b8183"},{"id":"func/Game.selectAllObjects","name":"Game.selectAllObjects","line":48,"end_line":51,"hash":"edcb26766c21b30a65d2522ee2110e47550a05b01668794e9479a25875458133"},{"id":"func/Game.deleteSelection","name":"Game.deleteSelection","line":53,"end_line":57,"hash":"0950775c64d1b5b335788c81f3dcadd30d0111d0c744742724b7b0024cb445e2"},{"id":"func/Game.cutSelection","name":"Game.cutSelection","line":59,"end_line":62,"hash":"f94aaf7910d0712a18442b0c34c7cb71f8e8a8b1f7ba643c517bb368a141dd4b"},{"id":"func/Game.pasteAtCursor","name":"Game.pasteAtCursor","line":64,"end_line":69,"hash":"a1755c5e4c344d5352b03c5b9826faba37f125e8ddc6848edaa851f9afd120fc"},{"id":"func/Game.duplicateSelection","name":"Game.duplicateSelection","line":71,"end_line":76,"hash":"55e12d54ccfd38a2939f1f3e6abb5c66eff9c2e74f2c0c17ddb94d480b41dfa2"},{"id":"func/Game.clearSelection","name":"Game.clearSelection","line":78,"end_line":81,"hash":"e0e598973b68a1130119e6e6b4afef3dc2d33f99d68870b5a022e21c7099caf9"},{"id":"func/Game.syncSelectionState","name":"Game.syncSelectionState","line":83,"end_line":85,"hash":"bd42cf00b3f30a59040fcb888eb6c0e16845e25150323c0aa662aacd8226c439"},{"id":"func/Game.openDemoPicker","name":"Game.openDemoPicker","line":87,"end_line":92,"hash":"cb9907d9e1982d1cae2655be0919db980ab1229c501d705fe43aad95fdc136e7"},{"id":"func/pathEntryLabel","name":"pathEntryLabel","line":94,"end_line":96,"hash":"6ca44d01998a5b0df0580e44ae8af4ba450a842944056541e6e921ed5f386df7"},{"id":"func/Game.SaveXSP","name":"Game.SaveXSP","line":104,"end_line":107,"hash":"baee32060c7ab4076d36d51b7ef89ed1e1352b41c251ff90c30e892e9cfbefbe"},{"id":"func/Game.LoadXSP","name":"Game.LoadXSP","line":109,"end_line":125,"hash":"22804ffa0b2f5cd329509c2cb526ec06e1ae5c7529fbee8162d7aeb440b21354"},{"id":"func/Game.LoadXSPFromFile","name":"Game.LoadXSPFromFile","line":127,"end_line":133,"hash":"a1def1f01e752b7e16b33e12b8c0ff7b60f9cf6fd9f96be118f8e2bfcd18d5e7"},{"id":"func/Game.InsertXSP","name":"Game.InsertXSP","line":135,"end_line":145,"hash":"8aff4daa19b7a1407b8f7ab1d32cb126ae5fe80ddc1434dc047ef4b515489795"},{"id":"func/Game.SaveState","name":"Game.SaveState","line":147,"end_line":149,"hash":"dd9f56918b445b02c95c8e9aa4348eca145a15954013e9ec245553640e584b6e"},{"id":"func/Game.RestoreState","name":"Game.RestoreState","line":151,"end_line":157,"hash":"98deb5f4b4accddc39a45691e81b6339d1cb3d75a7af435e7fc208d1dfdd86d9"},{"id":"func/Game.SetParameter","name":"Game.SetParameter","line":159,"end_line":162,"hash":"56f43c5c92b21b804c8d54e0a1c5cd4d95ab9364ffd9523ee54f1c02e39fd2d5"},{"id":"func/Game.ReplaceWorld","name":"Game.ReplaceWorld","line":164,"end_line":172,"hash":"2e034b21886ffd1acd5b0bf4dec0d34739ef4cbc8d52560ea8aab663732f95b5"},{"id":"func/setAppBounds","name":"setAppBounds","line":174,"end_line":176,"hash":"f6e12fbfe7691c9547e6e9530c3479b783bd58849842b6008d515c951775a34d"},{"id":"func/appBounds","name":"appBounds","line":178,"end_line":180,"hash":"fd2879f4622b6f4bd561230b7a429f07bd33d6d9fb77d1d2aa18b47eb1d78029"}]}
// mutate4go-manifest-end
