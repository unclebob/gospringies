package app

import "springs/internal/sim"

type editClipboard struct {
	Masses  []sim.Mass
	Springs []sim.Spring
}

func (g *Game) copySelection() {
	selection := editClipboard{}
	selectedMasses := map[int]bool{}
	for _, mass := range g.simulation.Masses {
		if g.editing().SelectedMasses[mass.ID] {
			selection.Masses = append(selection.Masses, mass)
			selectedMasses[mass.ID] = true
		}
	}
	for _, spring := range g.simulation.Springs {
		if g.editing().SelectedSprings[spring.ID] || (selectedMasses[spring.MassA] && selectedMasses[spring.MassB]) {
			selection.Springs = append(selection.Springs, spring)
		}
	}
	g.editClipboard = selection
}

func (g *Game) pasteSelectionAt(position sim.Vec2) bool {
	if len(g.editClipboard.Masses) == 0 {
		return false
	}
	idMap := map[int]int{}
	offset := position.Sub(g.editClipboard.origin())
	g.editing().SelectedMasses = map[int]bool{}
	g.editing().SelectedSprings = map[int]bool{}
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
	return true
}

func (c editClipboard) origin() sim.Vec2 {
	if len(c.Masses) == 0 {
		return sim.Vec2{}
	}
	minX := c.Masses[0].Position.X
	minY := c.Masses[0].Position.Y
	for _, mass := range c.Masses[1:] {
		if mass.Position.X < minX {
			minX = mass.Position.X
		}
		if mass.Position.Y < minY {
			minY = mass.Position.Y
		}
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
