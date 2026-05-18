package app

import "springs/internal/sim"

type RenderResult struct {
	Completed                  bool
	Representations            map[string]string
	SpringLinesVisible         bool
	MassesVisible              bool
	FixedMassDistinguishable   bool
	FixedMassRepresentation    string
	MovableMassRepresentation  string
	SelectedMassRepresentation string
}

func (g *Game) RenderWorld() RenderResult {
	g.RenderFrame()
	representations := g.renderRepresentations()
	hasMovable := representations["movable mass"] != ""
	hasFixed := representations["fixed mass"] != ""
	hasSpring := representations["spring"] != ""
	return RenderResult{
		Completed:                  true,
		Representations:            representations,
		SpringLinesVisible:         hasSpring,
		MassesVisible:              hasMovable || hasFixed,
		FixedMassDistinguishable:   hasMovable && hasFixed,
		FixedMassRepresentation:    "red circle",
		MovableMassRepresentation:  "yellow circle",
		SelectedMassRepresentation: "selection outline",
	}
}

func (g *Game) renderRepresentations() map[string]string {
	representations := map[string]string{}
	g.massRepresentations(representations)
	g.springRepresentation(representations)
	g.wallRepresentation(representations)
	g.selectionRepresentation(representations)
	g.centerRepresentation(representations)
	return representations
}

func (g *Game) springRepresentation(representations map[string]string) {
	if len(g.simulation.Springs) > 0 && g.showSprings() {
		representations["spring"] = "cyan line"
	}
}

func (g *Game) wallRepresentation(representations map[string]string) {
	if g.hasEnabledWall() {
		representations["enabled wall"] = "boundary line"
	}
}

func (g *Game) selectionRepresentation(representations map[string]string) {
	if g.selected {
		representations["selection"] = "selection outline"
	}
}

func (g *Game) centerRepresentation(representations map[string]string) {
	if g.simulation.CenterMassID() > 0 {
		representations["force center"] = "center marker"
	}
}

func (r RenderResult) HasVisibleRepresentation(object string) bool {
	if r.Representations == nil {
		return false
	}
	return r.Representations[object] != ""
}

func (g *Game) validSpring(spring sim.Spring) bool {
	_, _, ok := g.springEndpoints(spring)
	return ok
}

func (g *Game) springEndpoints(spring sim.Spring) (sim.Mass, sim.Mass, bool) {
	if spring.MassA != 0 || spring.MassB != 0 {
		a, okA := g.simulation.MassByID(spring.MassA)
		b, okB := g.simulation.MassByID(spring.MassB)
		return a, b, okA && okB
	}
	if spring.A >= 0 && spring.B >= 0 && spring.A < len(g.simulation.Masses) && spring.B < len(g.simulation.Masses) {
		return g.simulation.Masses[spring.A], g.simulation.Masses[spring.B], true
	}
	return sim.Mass{}, sim.Mass{}, false
}

func (g *Game) massRepresentations(representations map[string]string) {
	for _, mass := range g.simulation.Masses {
		if mass.Fixed {
			representations["fixed mass"] = "red circle"
		} else {
			representations["movable mass"] = "yellow circle"
		}
	}
}

func (g *Game) showSprings() bool {
	return g.simulation.Parameters.Value("show springs") == "true"
}

func (g *Game) hasEnabledWall() bool {
	for _, enabled := range g.simulation.Parameters.Walls {
		if enabled {
			return true
		}
	}
	return false
}
