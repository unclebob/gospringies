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
	g.editState.dirty = false
}

func (g *Game) restoreWorldState() {
	g.RestoreState()
	g.editState.dirty = true
}

func (g *Game) selectAllObjects() {
	g.editing().SelectAll()
	g.syncSelectionState()
}

func (g *Game) deleteSelection() {
	g.editing().DeleteSelected()
	g.editState.selected = false
	g.editState.dirty = true
}

func (g *Game) cutSelection() {
	g.copySelection()
	g.deleteSelection()
}

func (g *Game) pasteAtCursor() {
	if g.pasteSelectionAt(g.pointer.lastCursor) {
		g.editState.selected = true
		g.editState.dirty = true
	}
}

func (g *Game) duplicateSelection() {
	if _, err := g.editing().DuplicateSelected(); err == nil {
		g.editState.selected = true
		g.editState.dirty = true
	}
}

func (g *Game) clearSelection() {
	g.editing().ClearSelection()
	g.editState.selected = false
}

func (g *Game) syncSelectionState() {
	g.editState.selected = g.selectedObjectCount() > 0
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
	g.editState.dirty = false
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
	g.world.simulation.LoadFrom(loaded)
	if g.world.editor != nil {
		g.world.editor.World = g.world.simulation
	}
	g.editState.selected = false
	g.editState.dirty = false
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
	g.editState.dirty = true
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
	g.world.simulation.LoadFrom(state)
}

func (g *Game) SetParameter(parameter string, value string) {
	g.world.simulation.Parameters.Set(parameter, value)
	g.editState.dirty = true
}

func (g *Game) ReplaceWorld(world *sim.Simulation) {
	g.run.canvasYUp = false
	setAppBounds(world)
	g.applyCanvasWallBounds(world)
	g.world.simulation.LoadFrom(world)
	if g.world.editor != nil {
		g.world.editor.World = g.world.simulation
	}
}

func setAppBounds(world *sim.Simulation) {
	appcore.ApplyBounds(world, appBounds())
}

func appBounds() sim.Bounds {
	return sim.Bounds{Width: screenWidth, Height: screenHeight}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"09c48b760d18bd39fd834639d75720cc7fff528678ee1d6ab16ed9953d9e95ee","functions":[{"id":"func/Game.RunCommand","name":"Game.RunCommand","line":9,"end_line":16,"hash":"d9d03b55281863f9fff9e59453398529c7b65a7dd54f863854807db946456701"},{"id":"func/Game.resetWorld","name":"Game.resetWorld","line":38,"end_line":41,"hash":"97d68797c21c43151f733c57c89fcdf25e44fc8f5f95800bb6354eac9f01273c"},{"id":"func/Game.restoreWorldState","name":"Game.restoreWorldState","line":43,"end_line":46,"hash":"a718ade042215734505cf752ab396e43b22a0554c37ff37f8e9739fbfd7f2420"},{"id":"func/Game.selectAllObjects","name":"Game.selectAllObjects","line":48,"end_line":51,"hash":"edcb26766c21b30a65d2522ee2110e47550a05b01668794e9479a25875458133"},{"id":"func/Game.deleteSelection","name":"Game.deleteSelection","line":53,"end_line":57,"hash":"d939981dcd82691d4e5ec16d67eac790fb26bfc3a16a7c3f04fed7cf3a5bbb0d"},{"id":"func/Game.cutSelection","name":"Game.cutSelection","line":59,"end_line":62,"hash":"f94aaf7910d0712a18442b0c34c7cb71f8e8a8b1f7ba643c517bb368a141dd4b"},{"id":"func/Game.pasteAtCursor","name":"Game.pasteAtCursor","line":64,"end_line":69,"hash":"cce4c6adb1dc54fb67d588cac63cfe242b3d35d442fd11710645665d24400030"},{"id":"func/Game.duplicateSelection","name":"Game.duplicateSelection","line":71,"end_line":76,"hash":"470ebb5361e86a04f56e6033fde001a3bd52132aa8285746d18d469fba147b95"},{"id":"func/Game.clearSelection","name":"Game.clearSelection","line":78,"end_line":81,"hash":"3a56e598bdaa85885c5bb1d621c593020f324458503ebabb56bc5b0a8e7fb2c5"},{"id":"func/Game.syncSelectionState","name":"Game.syncSelectionState","line":83,"end_line":85,"hash":"b715423a49ddef074f0b1009aaab41d9e02f8b3836d67adb4aab9525fcb9f184"},{"id":"func/Game.openDemoPicker","name":"Game.openDemoPicker","line":87,"end_line":92,"hash":"cb9907d9e1982d1cae2655be0919db980ab1229c501d705fe43aad95fdc136e7"},{"id":"func/pathEntryLabel","name":"pathEntryLabel","line":94,"end_line":96,"hash":"6ca44d01998a5b0df0580e44ae8af4ba450a842944056541e6e921ed5f386df7"},{"id":"func/Game.SaveXSP","name":"Game.SaveXSP","line":104,"end_line":107,"hash":"ed63c6a7620bcb972d7f4e240e6e418ee468c34a93e68f9a65c7d6fab5ffea52"},{"id":"func/Game.LoadXSP","name":"Game.LoadXSP","line":109,"end_line":125,"hash":"59bcc0efb7e0c9971a1cf7bc4a802b5f33fb5441e91d4ef852f7399856b5785b"},{"id":"func/Game.LoadXSPFromFile","name":"Game.LoadXSPFromFile","line":127,"end_line":133,"hash":"116725efa47c8f0d307dd49156fc9bd6d1d0f37a9fa11526f9e5b8854f2daa21"},{"id":"func/Game.InsertXSP","name":"Game.InsertXSP","line":135,"end_line":145,"hash":"7b0b2505d9ce235222740e44e258791e9e79ca4d7dd0a80c918e934c221028b2"},{"id":"func/Game.SaveState","name":"Game.SaveState","line":147,"end_line":149,"hash":"7bacef1e6fdbdf872daf5baacf6a6373cd33f4d4ab06fadcfdfe283ddfa21f95"},{"id":"func/Game.RestoreState","name":"Game.RestoreState","line":151,"end_line":157,"hash":"352e99a3cda9645881c36acacd3545551dc6f081172f872f47cd1a6ecd96ef41"},{"id":"func/Game.SetParameter","name":"Game.SetParameter","line":159,"end_line":162,"hash":"4d077eb5385e321c8b8514f684b7782b2ff3b8f9884b573d832ba81df109618b"},{"id":"func/Game.ReplaceWorld","name":"Game.ReplaceWorld","line":164,"end_line":172,"hash":"da2c6ec8b1ee2355a08877c674cc4613bd2aba9cfb619048fcb8a28076f2590e"},{"id":"func/setAppBounds","name":"setAppBounds","line":174,"end_line":176,"hash":"f6e12fbfe7691c9547e6e9530c3479b783bd58849842b6008d515c951775a34d"},{"id":"func/appBounds","name":"appBounds","line":178,"end_line":180,"hash":"fd2879f4622b6f4bd561230b7a429f07bd33d6d9fb77d1d2aa18b47eb1d78029"}]}
// mutate4go-manifest-end
