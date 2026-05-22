package app

import (
	"math"

	"springs/internal/sim"
)

type editClipboard struct {
	Masses  []sim.Mass
	Springs []sim.Spring
}

func (g *Game) copySelection() {
	selection := editClipboard{}
	selectedMasses := g.copySelectedMasses(&selection)
	g.copySelectedSprings(&selection, selectedMasses)
	g.editState.clipboard = selection
}

func (g *Game) copySelectedMasses(selection *editClipboard) map[int]bool {
	selectedMasses := map[int]bool{}
	for _, mass := range g.world.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selection.Masses = append(selection.Masses, mass)
			selectedMasses[mass.ID] = true
		}
	}
	return selectedMasses
}

func (g *Game) copySelectedSprings(selection *editClipboard, selectedMasses map[int]bool) {
	for _, spring := range g.world.simulation.Springs {
		if g.editing().SelectedSprings[spring.ID] || (selectedMasses[spring.MassA] && selectedMasses[spring.MassB]) {
			selection.Springs = append(selection.Springs, spring)
		}
	}
}

func (g *Game) pasteSelectionAt(position sim.Vec2) bool {
	if len(g.editState.clipboard.Masses) == 0 {
		return false
	}
	idMap := map[int]int{}
	offset := position.Sub(g.editState.clipboard.origin())
	g.editing().SelectedMasses = map[int]bool{}
	g.editing().SelectedSprings = map[int]bool{}
	g.pasteClipboardMasses(offset, idMap)
	g.pasteClipboardSprings(idMap)
	return true
}

func (g *Game) pasteClipboardMasses(offset sim.Vec2, idMap map[int]int) {
	nextMass := g.nextMassID()
	for _, mass := range g.editState.clipboard.Masses {
		oldID := mass.ID
		mass.ID = nextMass
		nextMass++
		mass.Position = g.clampToCanvas(mass.Position.Add(offset))
		idMap[oldID] = mass.ID
		if err := g.world.simulation.AddMass(mass); err == nil {
			g.editing().SelectedMasses[mass.ID] = true
		}
	}
}

func (g *Game) pasteClipboardSprings(idMap map[int]int) {
	nextSpring := g.nextSpringID()
	for _, spring := range g.editState.clipboard.Springs {
		massA, okA := idMap[spring.MassA]
		massB, okB := idMap[spring.MassB]
		if !okA || !okB {
			continue
		}
		spring.ID = nextSpring
		nextSpring++
		spring.MassA = massA
		spring.MassB = massB
		if err := g.world.simulation.AddSpring(spring); err == nil {
			g.editing().SelectedSprings[spring.ID] = true
		}
	}
}

func (c editClipboard) origin() sim.Vec2 {
	if len(c.Masses) == 0 {
		return sim.Vec2{}
	}
	minX := math.MaxFloat64
	minY := math.MaxFloat64
	for _, mass := range c.Masses {
		minX = math.Min(minX, mass.Position.X)
		minY = math.Min(minY, mass.Position.Y)
	}
	return sim.Vec2{X: minX, Y: minY}
}

func (g *Game) nextMassID() int {
	next := 1
	for _, mass := range g.world.simulation.Masses {
		if mass.ID >= next {
			next = mass.ID + 1
		}
	}
	return next
}

func (g *Game) nextSpringID() int {
	next := 1
	for _, spring := range g.world.simulation.Springs {
		if spring.ID >= next {
			next = spring.ID + 1
		}
	}
	return next
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"4a228d494b08320eaf75c89bd1d353726005ffe9d036b592e8a6e552efac3da1","functions":[{"id":"func/Game.copySelection","name":"Game.copySelection","line":14,"end_line":19,"hash":"e0bc60362ae67ee457de5f3dfc98a6d5ade1c01e740721784679d4ef7a675921"},{"id":"func/Game.copySelectedMasses","name":"Game.copySelectedMasses","line":21,"end_line":30,"hash":"19b1ac613f3a74226275a5190a368482feb80dda8b29bc7255b95d259c16d0ef"},{"id":"func/Game.copySelectedSprings","name":"Game.copySelectedSprings","line":32,"end_line":38,"hash":"a950ff664a886b2a700785df07b53fc3b94d2434eb881049eaad11b7d0b32b18"},{"id":"func/Game.pasteSelectionAt","name":"Game.pasteSelectionAt","line":40,"end_line":51,"hash":"8fc26d281da5a958344d39b3bd2f8ef2193c989aadd156bf6837b9a1c2ca19c1"},{"id":"func/Game.pasteClipboardMasses","name":"Game.pasteClipboardMasses","line":53,"end_line":65,"hash":"3553edf8d0b8a381e4c6ead1edfa7b6736be0fd8b6d06e0cc36c7ddbc3fe54f0"},{"id":"func/Game.pasteClipboardSprings","name":"Game.pasteClipboardSprings","line":67,"end_line":83,"hash":"d6c76a4a9e1b45ee509cfb5b52008277e23b3f7e996cf23ec5398582c49a60dc"},{"id":"func/editClipboard.origin","name":"editClipboard.origin","line":85,"end_line":96,"hash":"4fbe0cafdea6f320625536d335d111834f5b5b26278d2e5ebf93558134a716f8"},{"id":"func/Game.nextMassID","name":"Game.nextMassID","line":98,"end_line":106,"hash":"3464fb456a6693be8c23354a0cbf7dae956e31b4b2e680d3b6d789ba2aa82497"},{"id":"func/Game.nextSpringID","name":"Game.nextSpringID","line":108,"end_line":116,"hash":"36f963d5d1cbdf74e3788f947e042469103ca9d572ac788e750a5a4e92b6dcab"}]}
// mutate4go-manifest-end
