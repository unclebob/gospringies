package app

import (
	"springs/internal/appcore"
	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func (g *Game) RunCommand(command string) {
	g.editState.lastCommand = command
	g.document.pathEntryCommand = ""
	action, ok := commandActions[command]
	if ok {
		action(g, command)
	}
}

var commandActions = map[string]func(*Game, string){
	"run":             func(g *Game, _ string) { g.run.paused = false },
	"pause":           func(g *Game, _ string) { g.run.paused = !g.run.paused },
	"pause toggle":    func(g *Game, _ string) { g.run.paused = !g.run.paused },
	"reset":           func(g *Game, _ string) { g.resetWorld() },
	"save state":      func(g *Game, _ string) { g.SaveState() },
	"restore state":   func(g *Game, _ string) { g.restoreWorldState() },
	"load":            func(g *Game, _ string) { g.openDemoPicker() },
	"insert":          func(g *Game, command string) { g.document.pathEntryCommand = pathEntryLabel(command) },
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
	g.world.simulation.Reset()
	g.clearDirty()
}

func (g *Game) restoreWorldState() {
	g.RestoreState()
	g.markDirty()
}

func (g *Game) selectAllObjects() {
	g.editing().SelectAll()
	g.syncSelectionState()
}

func (g *Game) deleteSelection() {
	g.editing().DeleteSelected()
	g.setSelected(false)
	g.markDirty()
}

func (g *Game) cutSelection() {
	g.copySelection()
	g.deleteSelection()
}

func (g *Game) pasteAtCursor() {
	if g.pasteSelectionAt(g.pointer.lastCursor) {
		g.setSelected(true)
		g.markDirty()
	}
}

func (g *Game) duplicateSelection() {
	if _, err := g.editing().DuplicateSelected(); err == nil {
		g.setSelected(true)
		g.markDirty()
	}
}

func (g *Game) clearSelection() {
	g.editing().ClearSelection()
	g.setSelected(false)
}

func (g *Game) syncSelectionState() {
	g.setSelected(g.selectedObjectCount() > 0)
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
	g.clearDirty()
	return xspfmt.SaveXSP(g.world.simulation)
}

func (g *Game) LoadXSP(input string) error {
	loaded, err := xspfmt.LoadXSP(input)
	if err != nil {
		return err
	}
	setAppBounds(loaded)
	g.run.canvasYUp = xspfmt.UsesOriginalXSpringiesCoordinates(input)
	g.applyCanvasWallBounds(loaded)
	g.world.simulation.Reset()
	g.loadWorldIntoSession(loaded)
	g.setSelected(false)
	g.clearDirty()
	return nil
}

func (g *Game) LoadXSPFromFile(path string, input string) error {
	if err := g.LoadXSP(input); err != nil {
		return err
	}
	g.document.currentFilePath = path
	return nil
}

func (g *Game) InsertXSP(input string) error {
	inserted, err := xspfmt.LoadXSP(input)
	if err != nil {
		return err
	}
	setAppBounds(inserted)
	g.applyCanvasWallBounds(inserted)
	g.world.simulation.InsertFrom(inserted)
	g.markDirty()
	return nil
}

func (g *Game) SaveState() {
	g.world.savedState = g.world.simulation.Clone()
}

func (g *Game) RestoreState() {
	state := g.world.savedState
	if state == nil {
		state = g.world.initialState
	}
	g.loadWorldIntoSession(state)
}

func (g *Game) SetParameter(parameter string, value string) {
	g.world.simulation.Parameters.Set(parameter, value)
	g.markDirty()
}

func (g *Game) ReplaceWorld(world *sim.Simulation) {
	g.run.canvasYUp = false
	setAppBounds(world)
	g.applyCanvasWallBounds(world)
	g.loadWorldIntoSession(world)
}

func setAppBounds(world *sim.Simulation) {
	appcore.ApplyBounds(world, appBounds())
}

func appBounds() sim.Bounds {
	return sim.Bounds{Width: screenWidth, Height: screenHeight}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:44:32-05:00","module_hash":"f3b1606d7ab5c77d20903a379472bc55ca683a3f60869045e1060bcf65ead05c","functions":[{"id":"func/Game.RunCommand","name":"Game.RunCommand","line":9,"end_line":16,"hash":"d9d03b55281863f9fff9e59453398529c7b65a7dd54f863854807db946456701"},{"id":"func/Game.resetWorld","name":"Game.resetWorld","line":38,"end_line":41,"hash":"dcee62912c6fb5cc65e726ddfe5550b5b240d7acff3ef826d55332b27c4c8160"},{"id":"func/Game.restoreWorldState","name":"Game.restoreWorldState","line":43,"end_line":46,"hash":"b6ec3f31e5383f2e6bb58a1bcff86b15f4c0d3c93adb5019b483ec2e44a4ba61"},{"id":"func/Game.selectAllObjects","name":"Game.selectAllObjects","line":48,"end_line":51,"hash":"edcb26766c21b30a65d2522ee2110e47550a05b01668794e9479a25875458133"},{"id":"func/Game.deleteSelection","name":"Game.deleteSelection","line":53,"end_line":57,"hash":"8449958529aa99dde99f7c4e01915ffe9987d353df23931856a7644eab20b34d"},{"id":"func/Game.cutSelection","name":"Game.cutSelection","line":59,"end_line":62,"hash":"f94aaf7910d0712a18442b0c34c7cb71f8e8a8b1f7ba643c517bb368a141dd4b"},{"id":"func/Game.pasteAtCursor","name":"Game.pasteAtCursor","line":64,"end_line":69,"hash":"073417983cc55cb0258176466fe4922d8cbb63359624f3cc0867bc6c1253ee6f"},{"id":"func/Game.duplicateSelection","name":"Game.duplicateSelection","line":71,"end_line":76,"hash":"82da47b69b31118b77899e4e1eb4ac4db16a3a9e0df9a3e599df5d6a26afedac"},{"id":"func/Game.clearSelection","name":"Game.clearSelection","line":78,"end_line":81,"hash":"b88b303601c82d2fba4e9315b575f42975ad621fcf823720fd52a6daf80e20e6"},{"id":"func/Game.syncSelectionState","name":"Game.syncSelectionState","line":83,"end_line":85,"hash":"fbb27ef1490e2958d713663c5d5a3b995d56eec4c66907cb026416939ab80572"},{"id":"func/Game.openDemoPicker","name":"Game.openDemoPicker","line":87,"end_line":92,"hash":"cb9907d9e1982d1cae2655be0919db980ab1229c501d705fe43aad95fdc136e7"},{"id":"func/pathEntryLabel","name":"pathEntryLabel","line":94,"end_line":96,"hash":"6ca44d01998a5b0df0580e44ae8af4ba450a842944056541e6e921ed5f386df7"},{"id":"func/Game.SaveXSP","name":"Game.SaveXSP","line":104,"end_line":107,"hash":"e3576ebe0eccce468aa723beaf2f7e66505632fab6e9079c8bcad75347a2e983"},{"id":"func/Game.LoadXSP","name":"Game.LoadXSP","line":109,"end_line":122,"hash":"dd275d269afd098ba42f9efb776bf52de43954ba279cfc228bf5b4a305d40c59"},{"id":"func/Game.LoadXSPFromFile","name":"Game.LoadXSPFromFile","line":124,"end_line":130,"hash":"116725efa47c8f0d307dd49156fc9bd6d1d0f37a9fa11526f9e5b8854f2daa21"},{"id":"func/Game.InsertXSP","name":"Game.InsertXSP","line":132,"end_line":142,"hash":"f33ea9d04908585e54ca644162f4b4d343956c835f3aa3e8e240064a6e6dfbff"},{"id":"func/Game.SaveState","name":"Game.SaveState","line":144,"end_line":146,"hash":"7bacef1e6fdbdf872daf5baacf6a6373cd33f4d4ab06fadcfdfe283ddfa21f95"},{"id":"func/Game.RestoreState","name":"Game.RestoreState","line":148,"end_line":154,"hash":"80456fab4a182172997bc3fb7f424ab212094e75f9beaef92c0d4bd969008d24"},{"id":"func/Game.SetParameter","name":"Game.SetParameter","line":156,"end_line":159,"hash":"a9abf4475cc53b6e02d47ca61b8bce31fce28ccf53d25d2663b3aeba4371d086"},{"id":"func/Game.ReplaceWorld","name":"Game.ReplaceWorld","line":161,"end_line":166,"hash":"9ed966b136ffb2d7a179bd9419d238ae95d360529574d143666a35d0ce2eb647"},{"id":"func/setAppBounds","name":"setAppBounds","line":168,"end_line":170,"hash":"f6e12fbfe7691c9547e6e9530c3479b783bd58849842b6008d515c951775a34d"},{"id":"func/appBounds","name":"appBounds","line":172,"end_line":174,"hash":"fd2879f4622b6f4bd561230b7a429f07bd33d6d9fb77d1d2aa18b47eb1d78029"}]}
// mutate4go-manifest-end
