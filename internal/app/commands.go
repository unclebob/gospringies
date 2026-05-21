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
	if g.pasteSelectionAt(g.lastCursor) {
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
	g.demoPickerOpen = true
	g.demoPickerScroll = 0
	g.demoFiles = nil
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
// {"version":1,"tested_at":"2026-05-19T11:39:03-05:00","module_hash":"2da3c85f116e71409a3614b289c435fd33be0189c138c31ecbc381a625334ee6","functions":[{"id":"func/Game.RunCommand","name":"Game.RunCommand","line":9,"end_line":16,"hash":"40c33d99c107c931962ad197020a08167e49bbc757b117c3e5109f76e4d6b312"},{"id":"func/Game.resetWorld","name":"Game.resetWorld","line":38,"end_line":41,"hash":"442a95b81114be9994aa54b476f87ef238ed9dc1ba5589b72c7fc24a7746c2b3"},{"id":"func/Game.restoreWorldState","name":"Game.restoreWorldState","line":43,"end_line":46,"hash":"2c7bfb10326dbbe774907dbeb63092bc76685521870c352d2d27d803f46b8183"},{"id":"func/Game.selectAllObjects","name":"Game.selectAllObjects","line":48,"end_line":51,"hash":"edcb26766c21b30a65d2522ee2110e47550a05b01668794e9479a25875458133"},{"id":"func/Game.deleteSelection","name":"Game.deleteSelection","line":53,"end_line":57,"hash":"0950775c64d1b5b335788c81f3dcadd30d0111d0c744742724b7b0024cb445e2"},{"id":"func/Game.cutSelection","name":"Game.cutSelection","line":59,"end_line":62,"hash":"f94aaf7910d0712a18442b0c34c7cb71f8e8a8b1f7ba643c517bb368a141dd4b"},{"id":"func/Game.pasteAtCursor","name":"Game.pasteAtCursor","line":64,"end_line":69,"hash":"40bffa2f7c0eea0e5ee6a508c429ee11fefa1259188738ab3dac78694d20a1dd"},{"id":"func/Game.duplicateSelection","name":"Game.duplicateSelection","line":71,"end_line":76,"hash":"55e12d54ccfd38a2939f1f3e6abb5c66eff9c2e74f2c0c17ddb94d480b41dfa2"},{"id":"func/Game.clearSelection","name":"Game.clearSelection","line":78,"end_line":81,"hash":"e0e598973b68a1130119e6e6b4afef3dc2d33f99d68870b5a022e21c7099caf9"},{"id":"func/Game.syncSelectionState","name":"Game.syncSelectionState","line":83,"end_line":85,"hash":"bd42cf00b3f30a59040fcb888eb6c0e16845e25150323c0aa662aacd8226c439"},{"id":"func/Game.openDemoPicker","name":"Game.openDemoPicker","line":87,"end_line":91,"hash":"72130ee455cd934ee470777faee1430af51282174f5d54c5f321b2ee37774a7d"},{"id":"func/pathEntryLabel","name":"pathEntryLabel","line":93,"end_line":95,"hash":"6ca44d01998a5b0df0580e44ae8af4ba450a842944056541e6e921ed5f386df7"},{"id":"func/Game.SaveXSP","name":"Game.SaveXSP","line":103,"end_line":106,"hash":"baee32060c7ab4076d36d51b7ef89ed1e1352b41c251ff90c30e892e9cfbefbe"},{"id":"func/Game.LoadXSP","name":"Game.LoadXSP","line":108,"end_line":123,"hash":"2c259098a144cf4cc6e5277211467e3632fa5a621eceeb13831af65177c46fb0"},{"id":"func/Game.InsertXSP","name":"Game.InsertXSP","line":125,"end_line":134,"hash":"13b50e53b57f621c301307c57750dd79260493defedb6098a360248685f372eb"},{"id":"func/Game.SaveState","name":"Game.SaveState","line":136,"end_line":138,"hash":"dd9f56918b445b02c95c8e9aa4348eca145a15954013e9ec245553640e584b6e"},{"id":"func/Game.RestoreState","name":"Game.RestoreState","line":140,"end_line":146,"hash":"98deb5f4b4accddc39a45691e81b6339d1cb3d75a7af435e7fc208d1dfdd86d9"},{"id":"func/Game.SetParameter","name":"Game.SetParameter","line":148,"end_line":151,"hash":"56f43c5c92b21b804c8d54e0a1c5cd4d95ab9364ffd9523ee54f1c02e39fd2d5"},{"id":"func/Game.ReplaceWorld","name":"Game.ReplaceWorld","line":153,"end_line":160,"hash":"7f64814bfbaf0fc6530c87bd1c830c3dc83091322563835625400ed318572cad"},{"id":"func/setAppBounds","name":"setAppBounds","line":162,"end_line":164,"hash":"f6e12fbfe7691c9547e6e9530c3479b783bd58849842b6008d515c951775a34d"},{"id":"func/appBounds","name":"appBounds","line":166,"end_line":168,"hash":"fd2879f4622b6f4bd561230b7a429f07bd33d6d9fb77d1d2aa18b47eb1d78029"}]}
// mutate4go-manifest-end
