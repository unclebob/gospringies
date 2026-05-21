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
	for _, spring := range g.simulation.Springs {
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
		return g.springIDEndpoints(spring)
	}
	return g.springIndexEndpoints(spring)
}

func (g *Game) springIDEndpoints(spring sim.Spring) (sim.Mass, sim.Mass, bool) {
	a, okA := g.simulation.MassByID(spring.MassA)
	b, okB := g.simulation.MassByID(spring.MassB)
	return a, b, okA && okB
}

func (g *Game) springIndexEndpoints(spring sim.Spring) (sim.Mass, sim.Mass, bool) {
	if validSpringIndex(spring.A, g.simulation.Masses) && validSpringIndex(spring.B, g.simulation.Masses) {
		return g.simulation.Masses[spring.A], g.simulation.Masses[spring.B], true
	}
	return sim.Mass{}, sim.Mass{}, false
}

func validSpringIndex(index int, masses []sim.Mass) bool {
	return index >= 0 && index < len(masses)
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:11:55-05:00","module_hash":"494bdd50f8137d4eac80e3be218d94ff1a9f1dfcc20fbfe468da9c3fef6bf453","functions":[{"id":"func/Game.RenderWorld","name":"Game.RenderWorld","line":16,"end_line":32,"hash":"2bde05b18a7be040d7562051747da18732595c0daf795e7cc392802d44513103"},{"id":"func/Game.renderRepresentations","name":"Game.renderRepresentations","line":34,"end_line":42,"hash":"2bfa213cc0eadb72a43450d224b129e49f7ad41bd0d2ba4306ab2b15e1dd5d0b"},{"id":"func/Game.springRepresentation","name":"Game.springRepresentation","line":44,"end_line":48,"hash":"6c8f33274d66770fef41e9ceca50a8742d5cf7fd952020c69f6e81471ae5f6b2"},{"id":"func/Game.wallRepresentation","name":"Game.wallRepresentation","line":50,"end_line":54,"hash":"86732e73fbfe9a694ea9c026d96be90b6be0f9b6506294b5f971d47cad7a3eda"},{"id":"func/Game.selectionRepresentation","name":"Game.selectionRepresentation","line":56,"end_line":60,"hash":"7ecb67b5f521b5f81b1ff2515ee9c2bb115e1eb914bd190a72238a6553f42c30"},{"id":"func/Game.centerRepresentation","name":"Game.centerRepresentation","line":62,"end_line":66,"hash":"855f9a37bc3e04a0afe6bed04a964d5c7120b4472c9199969b680d63b4f6c00d"},{"id":"func/RenderResult.HasVisibleRepresentation","name":"RenderResult.HasVisibleRepresentation","line":68,"end_line":73,"hash":"56c45d99dac804da77ff2103d1e169f6aaf3383a76ac090f42d281f5c894dcd7"},{"id":"func/Game.validSpring","name":"Game.validSpring","line":75,"end_line":78,"hash":"41532802819629c013da56d6fa7ad65140802b38d1a0c9db9cea0b5cb10fcbb8"},{"id":"func/Game.springEndpoints","name":"Game.springEndpoints","line":80,"end_line":85,"hash":"805b4e8628e219609e2dad37e4f9e4a6469a7436f208859e66826ce6a486fd38"},{"id":"func/Game.springIDEndpoints","name":"Game.springIDEndpoints","line":87,"end_line":91,"hash":"1d06f24b3d7d5f606df7afc9d3ca9cc59d93f8cbbb53a4cd05c6064f39fc6bf0"},{"id":"func/Game.springIndexEndpoints","name":"Game.springIndexEndpoints","line":93,"end_line":98,"hash":"387302bd9faf8afddf19592e5162b26e1eaeafbcd843f54af4c6b70d3d73fa8e"},{"id":"func/validSpringIndex","name":"validSpringIndex","line":100,"end_line":102,"hash":"ce1344f64662258ab4187ea4e15610caa4fd50c1409fdd372c7098ec2522aa22"},{"id":"func/Game.massRepresentations","name":"Game.massRepresentations","line":104,"end_line":112,"hash":"3f2fc352e6aab0f9ea3c533804a532bcb18ee43ac3a8a829cc5326ad88481056"},{"id":"func/Game.showSprings","name":"Game.showSprings","line":114,"end_line":116,"hash":"b55c71324fff938c5c8d054697e121e6e4a8e6cc3191addc588cf3e346ad667f"},{"id":"func/Game.hasEnabledWall","name":"Game.hasEnabledWall","line":118,"end_line":125,"hash":"10cdf2850bd491d91027435cef49667c8da8e2a7b00ca89ad698ce49df08a505"}]}
// mutate4go-manifest-end
