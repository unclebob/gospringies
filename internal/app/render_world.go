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
	if len(g.world.simulation.Springs) > 0 && g.showSprings() {
		representations["spring"] = "cyan line"
	}
	for _, spring := range g.world.simulation.Springs {
		if spring.Wall && g.showSprings() {
			representations["wall spring"] = "heavy orange line"
			return
		}
	}
}

func (g *Game) wallRepresentation(representations map[string]string) {
	if g.hasEnabledWall() {
		representations["enabled wall"] = "boundary line"
	}
}

func (g *Game) selectionRepresentation(representations map[string]string) {
	if g.editState.selected {
		representations["selection"] = "selection outline"
	}
}

func (g *Game) centerRepresentation(representations map[string]string) {
	if g.world.simulation.CenterMassID() > 0 {
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
		return g.springIDEndpoints(spring)
	}
	return g.springIndexEndpoints(spring)
}

func (g *Game) springIDEndpoints(spring sim.Spring) (sim.Mass, sim.Mass, bool) {
	a, okA := g.world.simulation.MassByID(spring.MassA)
	b, okB := g.world.simulation.MassByID(spring.MassB)
	return a, b, okA && okB
}

func (g *Game) springIndexEndpoints(spring sim.Spring) (sim.Mass, sim.Mass, bool) {
	if validSpringIndex(spring.A, g.world.simulation.Masses) && validSpringIndex(spring.B, g.world.simulation.Masses) {
		return g.world.simulation.Masses[spring.A], g.world.simulation.Masses[spring.B], true
	}
	return sim.Mass{}, sim.Mass{}, false
}

func validSpringIndex(index int, masses []sim.Mass) bool {
	return index >= 0 && index < len(masses)
}

func (g *Game) massRepresentations(representations map[string]string) {
	for _, mass := range g.world.simulation.Masses {
		if mass.Fixed {
			representations["fixed mass"] = "red circle"
		} else {
			representations["movable mass"] = "yellow circle"
		}
	}
}

func (g *Game) showSprings() bool {
	return g.world.simulation.Parameters.Value("show springs") == "true"
}

func (g *Game) hasEnabledWall() bool {
	for _, enabled := range g.world.simulation.Parameters.Walls {
		if enabled {
			return true
		}
	}
	return false
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"fe6bc7a18b454c6d670faa3a38205e9a8d40e6a6e20ace8cdfd17d3e201b66a6","functions":[{"id":"func/Game.RenderWorld","name":"Game.RenderWorld","line":16,"end_line":32,"hash":"2bde05b18a7be040d7562051747da18732595c0daf795e7cc392802d44513103"},{"id":"func/Game.renderRepresentations","name":"Game.renderRepresentations","line":34,"end_line":42,"hash":"2bfa213cc0eadb72a43450d224b129e49f7ad41bd0d2ba4306ab2b15e1dd5d0b"},{"id":"func/Game.springRepresentation","name":"Game.springRepresentation","line":44,"end_line":54,"hash":"cd713f7b6b37b34fe498767357d87bd3eb71952a9f6e7cc9a9c54d72fd7b5a1a"},{"id":"func/Game.wallRepresentation","name":"Game.wallRepresentation","line":56,"end_line":60,"hash":"86732e73fbfe9a694ea9c026d96be90b6be0f9b6506294b5f971d47cad7a3eda"},{"id":"func/Game.selectionRepresentation","name":"Game.selectionRepresentation","line":62,"end_line":66,"hash":"19c1edd1fb987207915404449075bff89ffe4d3f1f113bbce454c6c9b87b1a12"},{"id":"func/Game.centerRepresentation","name":"Game.centerRepresentation","line":68,"end_line":72,"hash":"17d0f0db3aa32d2cbe813ca9089759d2261f659116dcdfddfa09e4791d54d2bd"},{"id":"func/RenderResult.HasVisibleRepresentation","name":"RenderResult.HasVisibleRepresentation","line":74,"end_line":79,"hash":"56c45d99dac804da77ff2103d1e169f6aaf3383a76ac090f42d281f5c894dcd7"},{"id":"func/Game.validSpring","name":"Game.validSpring","line":81,"end_line":84,"hash":"41532802819629c013da56d6fa7ad65140802b38d1a0c9db9cea0b5cb10fcbb8"},{"id":"func/Game.springEndpoints","name":"Game.springEndpoints","line":86,"end_line":91,"hash":"805b4e8628e219609e2dad37e4f9e4a6469a7436f208859e66826ce6a486fd38"},{"id":"func/Game.springIDEndpoints","name":"Game.springIDEndpoints","line":93,"end_line":97,"hash":"c037e7fbab2fd2ec73ba614f9b06820e103dcfa1a22f64948a2c7f8296d1d9ee"},{"id":"func/Game.springIndexEndpoints","name":"Game.springIndexEndpoints","line":99,"end_line":104,"hash":"006c4fdede58c8d97ece257e4e1ad5cc6be2fa77d7b9c9401694ce983f895c37"},{"id":"func/validSpringIndex","name":"validSpringIndex","line":106,"end_line":108,"hash":"ce1344f64662258ab4187ea4e15610caa4fd50c1409fdd372c7098ec2522aa22"},{"id":"func/Game.massRepresentations","name":"Game.massRepresentations","line":110,"end_line":118,"hash":"7d96b578aeaf56d94ec5337b753baa17956bb6e7d7b2816062de2c77f812d5c9"},{"id":"func/Game.showSprings","name":"Game.showSprings","line":120,"end_line":122,"hash":"c0717d8e912ce5229bb7be1815991a468565977e6d20281fe623c4f98067cdfc"},{"id":"func/Game.hasEnabledWall","name":"Game.hasEnabledWall","line":124,"end_line":131,"hash":"4913f9c6ba83c26e065b932d8e90f19c77ead74f428029e9da2beb559352b954"}]}
// mutate4go-manifest-end
