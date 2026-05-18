package app

import (
	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func (g *Game) RunCommand(command string) {
	g.lastCommand = command
	g.pathEntryCommand = ""
	switch command {
	case "run":
		g.paused = false
	case "pause", "pause toggle":
		g.paused = !g.paused
	case "reset":
		g.simulation.Reset()
		g.dirty = false
	case "save state":
		g.SaveState()
	case "restore state":
		g.RestoreState()
		g.dirty = true
	case "load":
		g.openDemoPicker()
	case "insert", "save":
		g.pathEntryCommand = pathEntryLabel(command)
	case "select all":
		g.editing().SelectAll()
		g.syncSelectionState()
	case "clear selection":
		g.clearSelection()
	case "delete":
		g.editing().DeleteSelected()
		g.selected = false
		g.dirty = true
	case "cut":
		g.copySelection()
		g.editing().DeleteSelected()
		g.selected = false
		g.dirty = true
	case "copy":
		g.copySelection()
	case "paste":
		if g.pasteSelectionAt(g.lastCursor) {
			g.selected = true
			g.dirty = true
		}
	case "duplicate":
		if _, err := g.editing().DuplicateSelected(); err == nil {
			g.selected = true
			g.dirty = true
		}
	case "quit":
		_ = g.Close()
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
	switch command {
	case "load":
		return "Load"
	case "insert":
		return "Insert"
	case "save":
		return "Save"
	default:
		return ""
	}
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
	g.simulation.LoadFrom(world)
	if g.editor != nil {
		g.editor.World = g.simulation
	}
}
