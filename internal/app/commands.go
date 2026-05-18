package app

import (
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
	"save":            func(g *Game, command string) { g.pathEntryCommand = pathEntryLabel(command) },
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
	g.simulation.Reset()
	g.simulation.LoadFrom(loaded)
	if g.editor != nil {
		g.editor.World = g.simulation
	}
	g.selected = false
	g.dirty = false
	return nil
}

func (g *Game) InsertXSP(input string) error {
	inserted, err := xspfmt.LoadXSP(input)
	if err != nil {
		return err
	}
	setAppBounds(inserted)
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
	g.simulation.LoadFrom(world)
	if g.editor != nil {
		g.editor.World = g.simulation
	}
}

func setAppBounds(world *sim.Simulation) {
	world.Bounds = sim.Bounds{Width: screenWidth, Height: screenHeight}
}
