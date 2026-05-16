package app

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"springs/internal/sim"
)

func TestNewGameStartsWithEmptyWorld(t *testing.T) {
	game := NewGame()

	if len(game.World().Masses) != 0 || len(game.World().Springs) != 0 {
		t.Fatalf("world = %#v", game.World())
	}
}

func TestGameLayoutUsesWindowSize(t *testing.T) {
	game := NewGame()
	width, height := game.Layout(1, 1)
	if width != screenWidth || height != screenHeight {
		t.Fatalf("layout = %d, %d", width, height)
	}
}

func TestGameUpdateStepsOnlyWhenUnpaused(t *testing.T) {
	game := NewGame()
	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})

	if err := game.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if game.World().Time == 0 {
		t.Fatal("expected unpaused update to advance simulation time")
	}

	game.SetPaused(true)
	pausedAt := game.World().Time
	if err := game.Update(); err != nil {
		t.Fatalf("paused Update returned error: %v", err)
	}
	if game.World().Time != pausedAt {
		t.Fatalf("paused time = %f, want %f", game.World().Time, pausedAt)
	}
	if !game.InputActive() {
		t.Fatal("expected input handling to remain active")
	}
}

func TestGameDraw(t *testing.T) {
	game := NewGame()
	left := game.World().AddMassAt(sim.Vec2{X: 10, Y: 10}, 1, true)
	right := game.World().AddMassAt(sim.Vec2{X: 30, Y: 10}, 1, false)
	game.World().AddSpringBetween(left, right, 20, 12)
	screen := ebiten.NewImage(screenWidth, screenHeight)

	game.Draw(screen)

	if !game.RenderingActive() {
		t.Fatal("expected rendering to be active")
	}
}

func TestDrawWallsUsesSimulationBounds(t *testing.T) {
	lines := wallDrawLines(sim.Bounds{Width: 20, Height: 16})
	expected := []wallDrawLine{
		{name: "top", x1: 0, y1: 0, x2: 20, y2: 0},
		{name: "bottom", x1: 0, y1: 15, x2: 20, y2: 15},
		{name: "left", x1: 0, y1: 0, x2: 0, y2: 16},
		{name: "right", x1: 19, y1: 0, x2: 19, y2: 16},
	}
	if len(lines) != len(expected) {
		t.Fatalf("line count = %d, want %d", len(lines), len(expected))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("line %d = %#v, want %#v", i, line, expected[i])
		}
	}
}

func TestRenderWorldCompletesForEmptyAndNonEmptyWorlds(t *testing.T) {
	for _, setup := range []func(*Game){func(*Game) {}, addRenderableSpring} {
		game := NewGame()
		setup(game)
		if result := game.RenderWorld(); !result.Completed {
			t.Fatalf("render result = %#v", result)
		}
	}
}

func TestMassDrawCircleCentersOnMassPosition(t *testing.T) {
	x, y, radius := massDrawCircle(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}})

	if x != 30 || y != 40 || radius != 5 {
		t.Fatalf("draw circle = %f,%f radius %f", x, y, radius)
	}
}

func TestSelectionOutlineSurroundsMassPosition(t *testing.T) {
	lines := selectionOutline(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}})
	expected := []selectionLine{
		{22, 32, 38, 32},
		{38, 32, 38, 48},
		{38, 48, 22, 48},
		{22, 48, 22, 32},
	}

	if len(lines) != len(expected) {
		t.Fatalf("line count = %d", len(lines))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("line %d = %#v, want %#v", i, line, expected[i])
		}
	}
}

func TestSelectedMassOutlineIsEmptyWithoutMasses(t *testing.T) {
	if lines := selectedMassOutline(nil); len(lines) != 0 {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestSelectedMassOutlineUsesFirstMass(t *testing.T) {
	lines := selectedMassOutline([]sim.Mass{
		{Position: sim.Vec2{X: 30, Y: 40}},
	})
	expected := selectionOutline(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}})

	if len(lines) != len(expected) {
		t.Fatalf("line count = %d", len(lines))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("line %d = %#v, want %#v", i, line, expected[i])
		}
	}
}

func TestRenderWorldReportsVisibleObjects(t *testing.T) {
	game := NewGame()
	addRenderableSpring(game)
	game.World().Parameters.EnableWall("left")
	game.World().SetForceCenter([]int{1})
	game.SetSelected(true)

	result := game.RenderWorld()

	for _, object := range []string{"movable mass", "fixed mass", "spring", "enabled wall", "selection", "force center"} {
		if !result.HasVisibleRepresentation(object) {
			t.Fatalf("missing representation for %q: %#v", object, result.Representations)
		}
	}
}

func TestEmptyRenderResultHasNoVisibleRepresentation(t *testing.T) {
	var result RenderResult

	if result.HasVisibleRepresentation("mass") {
		t.Fatal("empty render result reported a visible representation")
	}
}

func TestShowSpringsControlsSpringVisibility(t *testing.T) {
	game := NewGame()
	addRenderableSpring(game)
	game.World().Parameters.Set("show springs", "false")

	result := game.RenderWorld()

	if result.SpringLinesVisible {
		t.Fatal("expected spring lines to be hidden")
	}
	if !result.MassesVisible {
		t.Fatal("expected masses to remain visible")
	}
}

func TestMassVisibilityAndFixedDistinctionRequireExpectedMassTypes(t *testing.T) {
	tests := []struct {
		name       string
		masses     []sim.Mass
		wantMasses bool
		wantBoth   bool
	}{
		{name: "only movable", masses: []sim.Mass{{ID: 1, Mass: 1}}, wantMasses: true},
		{name: "only fixed", masses: []sim.Mass{{ID: 1, Mass: 1, Fixed: true}}, wantMasses: true},
		{name: "both", masses: []sim.Mass{{ID: 1, Mass: 1, Fixed: true}, {ID: 2, Mass: 1}}, wantMasses: true, wantBoth: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := NewGame()
			game.World().Masses = append(game.World().Masses, test.masses...)

			result := game.RenderWorld()

			if result.MassesVisible != test.wantMasses {
				t.Fatalf("masses visible = %t, want %t", result.MassesVisible, test.wantMasses)
			}
			if result.FixedMassDistinguishable != test.wantBoth {
				t.Fatalf("fixed distinguishable = %t, want %t", result.FixedMassDistinguishable, test.wantBoth)
			}
		})
	}
}

func TestRenderWorldOmitsAbsentSpringAndWallRepresentations(t *testing.T) {
	game := NewGame()

	result := game.RenderWorld()

	if result.HasVisibleRepresentation("spring") {
		t.Fatalf("unexpected spring representation: %#v", result.Representations)
	}
	if result.HasVisibleRepresentation("enabled wall") {
		t.Fatalf("unexpected wall representation: %#v", result.Representations)
	}
}

func TestValidSpringRejectsOutOfRangeEndpoints(t *testing.T) {
	game := NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})
	_ = game.World().AddMass(sim.Mass{ID: 2, Mass: 1})

	tests := []struct {
		name   string
		spring sim.Spring
		valid  bool
	}{
		{name: "valid", spring: sim.Spring{A: 0, B: 1}, valid: true},
		{name: "valid zero b", spring: sim.Spring{A: 1, B: 0}, valid: true},
		{name: "negative a", spring: sim.Spring{A: -1, B: 1}},
		{name: "negative b", spring: sim.Spring{A: 0, B: -1}},
		{name: "a too high", spring: sim.Spring{A: 2, B: 1}},
		{name: "b too high", spring: sim.Spring{A: 0, B: 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := game.validSpring(test.spring); got != test.valid {
				t.Fatalf("validSpring(%#v) = %t, want %t", test.spring, got, test.valid)
			}
		})
	}
}

func TestFixedMassesAreDistinguishable(t *testing.T) {
	game := NewGame()
	addRenderableSpring(game)

	result := game.RenderWorld()

	if !result.FixedMassDistinguishable {
		t.Fatalf("render result = %#v", result)
	}
	if result.FixedMassRepresentation == result.MovableMassRepresentation {
		t.Fatal("expected distinct fixed and movable mass representations")
	}
}

func TestWindowConfigIsResizable(t *testing.T) {
	config := DefaultWindowConfig()
	if !config.Resizable {
		t.Fatal("window should be resizable")
	}
}

func addRenderableSpring(game *Game) {
	left := game.World().AddMassAt(sim.Vec2{X: 10, Y: 10}, 1, true)
	right := game.World().AddMassAt(sim.Vec2{X: 30, Y: 10}, 1, false)
	game.World().AddSpringBetween(left, right, 20, 12)
}

func TestRenderFrameMarksRenderingActive(t *testing.T) {
	game := NewGame()

	game.RenderFrame()

	if !game.RenderingActive() {
		t.Fatal("expected rendering to be active")
	}
}

func TestGameClosesCleanly(t *testing.T) {
	game := NewGame()
	if err := game.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !game.Closed() {
		t.Fatal("expected game to be closed")
	}
}

func TestEditorScreenHasRequiredRegions(t *testing.T) {
	screen := NewGame().EditorScreen()
	expected := map[string]string{
		"canvas":          "edit and view the simulation world",
		"left toolbar":    "choose editing modes",
		"top bar":         "run commands and file commands",
		"right inspector": "edit selected objects and world parameters",
		"status line":     "show mode, simulation state, counts, and file state",
	}

	if !screen.Editor || screen.LandingPage {
		t.Fatalf("screen = %#v", screen)
	}
	for region, purpose := range expected {
		if got, ok := screen.RegionPurpose(region); !ok || got != purpose {
			t.Fatalf("region %q = %q, %t", region, got, ok)
		}
	}
	if got, ok := screen.RegionPurpose("footer"); ok || got != "" {
		t.Fatalf("missing region = %q, %t", got, ok)
	}
}

func TestEditorScreenHasVisibleModeAndCommandControls(t *testing.T) {
	screen := NewGame().EditorScreen()
	for _, mode := range []string{"select", "add mass", "add spring", "drag"} {
		if !screen.HasModeControl(mode) {
			t.Fatalf("missing mode %q", mode)
		}
	}
	for _, command := range []string{"run", "pause", "reset", "load", "insert", "save", "quit", "delete", "select all"} {
		if !screen.HasCommandControl(command) {
			t.Fatalf("missing command %q", command)
		}
	}
	if screen.HasModeControl("paint") {
		t.Fatal("unexpected paint mode")
	}
	if screen.HasCommandControl("export") {
		t.Fatal("unexpected export command")
	}
}

func TestEditorIndicatorsReflectState(t *testing.T) {
	game := NewGame()
	game.SetMode("select")
	game.SetPaused(true)
	game.SetSelected(true)
	game.SetDirty(true)

	indicators := game.EditorScreen().Indicators

	if indicators["active mode"] != "select mode" {
		t.Fatalf("active mode = %q", indicators["active mode"])
	}
	if indicators["simulation state"] != "paused" {
		t.Fatalf("simulation state = %q", indicators["simulation state"])
	}
	if indicators["selection"] != "object selected" {
		t.Fatalf("selection = %q", indicators["selection"])
	}
	if indicators["file state"] != "unsaved changes" {
		t.Fatalf("file state = %q", indicators["file state"])
	}
}

func TestKeyboardShortcutsRunVisibleCommands(t *testing.T) {
	game := NewGame()
	shortcuts := map[string]string{
		"Space":  "pause",
		"Delete": "delete",
		"Ctrl+A": "select all",
		"R":      "reset",
		"Ctrl+S": "save",
		"Ctrl+O": "load",
		"Ctrl+I": "insert",
		"Q":      "quit",
	}

	for shortcut, command := range shortcuts {
		if !game.EditorScreen().HasCommandControl(command) {
			t.Fatalf("missing visible command %q", command)
		}
		if !game.HandleShortcut(shortcut) {
			t.Fatalf("shortcut %q was not handled", shortcut)
		}
		if got := game.LastCommand(); got != command {
			t.Fatalf("shortcut %q ran %q, want %q", shortcut, got, command)
		}
	}
	if game.HandleShortcut("Ctrl+X") {
		t.Fatal("unexpected unknown shortcut handling")
	}
}

func TestCommandsAffectApplicationState(t *testing.T) {
	game := NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})
	game.World().Parameters.Set("current mass", "custom")

	game.RunCommand("pause")
	if !game.Paused() {
		t.Fatal("expected pause command to toggle paused state")
	}

	game.RunCommand("reset")
	if len(game.World().Masses) != 0 || game.World().Parameters.Value("current mass") == "custom" {
		t.Fatalf("reset world = %#v", game.World())
	}

	game.RunCommand("quit")
	if !game.Closed() {
		t.Fatal("expected quit command to close game")
	}
}

func TestFileCommandsSaveLoadAndInsertXSP(t *testing.T) {
	game := NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 1, Y: 2}, Mass: 1})

	saved := game.SaveXSP()
	if saved == "" || game.EditorScreen().Indicators["file state"] != "saved" {
		t.Fatalf("saved = %q indicators = %#v", saved, game.EditorScreen().Indicators)
	}

	loaded := "#1.0\ncmas 7\nmass 9 10 20 1 0\n"
	if err := game.LoadXSP(loaded); err != nil {
		t.Fatal(err)
	}
	if _, ok := game.World().MassByID(9); !ok || game.World().Parameters.Value("current mass") != "7" {
		t.Fatalf("loaded world = %#v", game.World())
	}

	game.World().Parameters.Set("current mass", "kept")
	if err := game.InsertXSP("#1.0\ncmas inserted\nmass 10 30 40 1 0\n"); err != nil {
		t.Fatal(err)
	}
	if _, ok := game.World().MassByID(10); !ok || game.World().Parameters.Value("current mass") != "kept" {
		t.Fatalf("inserted world = %#v", game.World())
	}
}

func TestParameterCommandMarksFileDirty(t *testing.T) {
	game := NewGame()
	_ = game.SaveXSP()

	game.SetParameter("current mass", "custom")

	if game.World().Parameters.Value("current mass") != "custom" {
		t.Fatalf("parameters = %#v", game.World().Parameters)
	}
	if game.EditorScreen().Indicators["file state"] != "unsaved changes" {
		t.Fatalf("indicators = %#v", game.EditorScreen().Indicators)
	}
}

func TestSaveStateRestoresObjectsAndParametersRepeatedly(t *testing.T) {
	game := NewGame()
	replaceWithAppTestState(game, "saved")
	game.SaveState()

	replaceWithAppTestState(game, "changed")
	game.RestoreState()
	assertAppTestState(t, game, "saved")

	replaceWithAppTestState(game, "changed")
	game.RestoreState()
	assertAppTestState(t, game, "saved")
}

func TestRestoreStateWithoutSavedStateRestoresInitialWorld(t *testing.T) {
	game := NewGame()
	replaceWithAppTestState(game, "changed")

	game.RestoreState()

	if len(game.World().Masses) != 0 || len(game.World().Springs) != 0 {
		t.Fatalf("world = %#v", game.World())
	}
	if game.World().Parameters.Value("current mass") != sim.DefaultParameters().Value("current mass") {
		t.Fatalf("parameters = %#v", game.World().Parameters)
	}
}

func TestFileOperationsDoNotReplaceSavedState(t *testing.T) {
	game := NewGame()
	replaceWithAppTestState(game, "saved")
	game.SaveState()

	if err := game.LoadXSP("#1.0\ncmas loaded\nmass 9 10 20 1 0\n"); err != nil {
		t.Fatal(err)
	}
	if _, ok := game.World().MassByID(9); !ok {
		t.Fatal("expected file load to replace current world")
	}
	game.RestoreState()

	assertAppTestState(t, game, "saved")
}

func TestEditorControlsRemainUsableWhilePausedOrRunning(t *testing.T) {
	game := NewGame()
	for _, paused := range []bool{true, false} {
		game.SetPaused(paused)
		screen := game.EditorScreen()
		if !screen.CanvasVisible || !screen.ControlsUsable {
			t.Fatalf("paused %t screen = %#v", paused, screen)
		}
	}
}

func replaceWithAppTestState(game *Game, label string) {
	world := sim.NewWorld()
	switch label {
	case "saved":
		_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 2, Elasticity: 0.6, Fixed: true})
		_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 40, Y: 20}, Mass: 3})
		_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 30, SpringConstant: 8, Damping: 0.4})
		world.Parameters.Set("current mass", "saved")
	case "changed":
		_ = world.AddMass(sim.Mass{ID: 7, Position: sim.Vec2{X: 70, Y: 80}, Mass: 4})
		world.Parameters.Set("current mass", "changed")
	}
	game.ReplaceWorld(world)
}

func assertAppTestState(t *testing.T, game *Game, label string) {
	t.Helper()
	switch label {
	case "saved":
		if len(game.World().Masses) != 2 || len(game.World().Springs) != 1 {
			t.Fatalf("saved world = %#v", game.World())
		}
		mass, ok := game.World().MassByID(1)
		if !ok || mass.Position != (sim.Vec2{X: 10, Y: 20}) || !mass.Fixed {
			t.Fatalf("saved mass = %#v, %t", mass, ok)
		}
		if game.World().Parameters.Value("current mass") != "saved" {
			t.Fatalf("saved parameters = %#v", game.World().Parameters)
		}
	default:
		t.Fatalf("unsupported app test state %q", label)
	}
}
