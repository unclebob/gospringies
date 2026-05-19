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
	g.editClipboard = selection
}

func (g *Game) copySelectedMasses(selection *editClipboard) map[int]bool {
	selectedMasses := map[int]bool{}
	for _, mass := range g.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selection.Masses = append(selection.Masses, mass)
			selectedMasses[mass.ID] = true
		}
	}
	return selectedMasses
}

func (g *Game) copySelectedSprings(selection *editClipboard, selectedMasses map[int]bool) {
	for _, spring := range g.simulation.Springs {
		if g.editing().SelectedSprings[spring.ID] || (selectedMasses[spring.MassA] && selectedMasses[spring.MassB]) {
			selection.Springs = append(selection.Springs, spring)
		}
	}
}

func (g *Game) pasteSelectionAt(position sim.Vec2) bool {
	if len(g.editClipboard.Masses) == 0 {
		return false
	}
	idMap := map[int]int{}
	offset := position.Sub(g.editClipboard.origin())
	g.editing().SelectedMasses = map[int]bool{}
	g.editing().SelectedSprings = map[int]bool{}
	g.pasteClipboardMasses(offset, idMap)
	g.pasteClipboardSprings(idMap)
	return true
}

func (g *Game) pasteClipboardMasses(offset sim.Vec2, idMap map[int]int) {
	nextMass := g.nextMassID()
	for _, mass := range g.editClipboard.Masses {
		oldID := mass.ID
		mass.ID = nextMass
		nextMass++
		mass.Position = mass.Position.Add(offset)
		idMap[oldID] = mass.ID
		if err := g.simulation.AddMass(mass); err == nil {
			g.editing().SelectedMasses[mass.ID] = true
		}
	}
}

func (g *Game) pasteClipboardSprings(idMap map[int]int) {
	nextSpring := g.nextSpringID()
	for _, spring := range g.editClipboard.Springs {
		massA, okA := idMap[spring.MassA]
		massB, okB := idMap[spring.MassB]
		if !okA || !okB {
			continue
		}
		spring.ID = nextSpring
		nextSpring++
		spring.MassA = massA
		spring.MassB = massB
		if err := g.simulation.AddSpring(spring); err == nil {
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
	for _, mass := range g.simulation.Masses {
		if mass.ID >= next {
			next = mass.ID + 1
		}
	}
	return next
}

func (g *Game) nextSpringID() int {
	next := 1
	for _, spring := range g.simulation.Springs {
		if spring.ID >= next {
			next = spring.ID + 1
		}
	}
	return next
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:03:32-05:00","module_hash":"7fcb9ab4c4cb83e31d26d542bc260d89366da62f0b6c51e1990ab2aa75f571a7","functions":[{"id":"func/Game.copySelection","name":"Game.copySelection","line":14,"end_line":19,"hash":"d2c828fe96fa3bfe53dc4b4cb7970f1ec2327c5d628025c8906975549c0588f1"},{"id":"func/Game.copySelectedMasses","name":"Game.copySelectedMasses","line":21,"end_line":30,"hash":"9548e21ae623d34c0cdf281e937c2344045383326c400018aabda1b82d561191"},{"id":"func/Game.copySelectedSprings","name":"Game.copySelectedSprings","line":32,"end_line":38,"hash":"96192e16ad03cc66104982a490ad15a171ff89eb3c19d5ccaa7aabb50ec704c4"},{"id":"func/Game.pasteSelectionAt","name":"Game.pasteSelectionAt","line":40,"end_line":51,"hash":"efbe47a966bb549862b0e060300c7fb9bb7c3dd75a1fe4f2f24415a2f9e79e57"},{"id":"func/Game.pasteClipboardMasses","name":"Game.pasteClipboardMasses","line":53,"end_line":65,"hash":"3e0c2a40fd5585767db6e6355f64639477c8745f556e503a541efca3cce95599"},{"id":"func/Game.pasteClipboardSprings","name":"Game.pasteClipboardSprings","line":67,"end_line":83,"hash":"faf6c3906f21c6f55eb77ce4d312ebe5deee025e07dec4b6e2579eef02702158"},{"id":"func/editClipboard.origin","name":"editClipboard.origin","line":85,"end_line":96,"hash":"4fbe0cafdea6f320625536d335d111834f5b5b26278d2e5ebf93558134a716f8"},{"id":"func/Game.nextMassID","name":"Game.nextMassID","line":98,"end_line":106,"hash":"8cb37dcdf29f42bf39fd05607c18043e39294da13080804e0284819b5ef3b3db"},{"id":"func/Game.nextSpringID","name":"Game.nextSpringID","line":108,"end_line":116,"hash":"69d4dea92ace190676544b4259edae058a1fe9f603c2228a133b152d2115e615"}]}
// mutate4go-manifest-end
