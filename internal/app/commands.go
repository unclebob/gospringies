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
	case "load", "insert", "save":
		g.pathEntryCommand = pathEntryLabel(command)
	case "quit":
		_ = g.Close()
	}
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
	g.simulation.LoadFrom(loaded)
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
	g.simulation.LoadFrom(world)
}
