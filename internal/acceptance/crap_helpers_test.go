package acceptance

import (
	"strings"
	"testing"

	"springs/internal/app"
	"springs/internal/sim"
)

func TestClickableVisibleControlActiveAssertion(t *testing.T) {
	w := &world{appGame: app.NewGame()}

	if err := assertVisibleControlActive(w, map[string]string{"control": "Pause"}); err != nil {
		t.Fatal(err)
	}
	if err := assertVisibleControlActive(w, map[string]string{"control": "Missing"}); err == nil {
		t.Fatal("expected missing active control error")
	}
}

func TestSupportedClickableModes(t *testing.T) {
	for _, mode := range []string{"select", "add mass", "add spring", "drag"} {
		if !supportedClickableMode(mode) {
			t.Fatalf("mode %q should be supported", mode)
		}
	}
	if supportedClickableMode("erase") {
		t.Fatal("erase should not be a supported clickable mode")
	}
}

func TestAppControlWithLabelReportsMissing(t *testing.T) {
	if _, ok := appControlWithLabel("Pause"); !ok {
		t.Fatal("Pause control should exist")
	}
	if _, ok := appControlWithLabel("Missing"); ok {
		t.Fatal("Missing control should not exist")
	}
}

func TestClickableHelpersReportPrerequisiteFailures(t *testing.T) {
	w := &world{appGame: nonConcreteAppGame{}}
	for _, tt := range []struct {
		name string
		err  error
		want string
	}{
		{name: "demo picker missing concrete app", err: assertDemoPickerOpen(w, nil), want: "application was not started"},
		{name: "game value missing key", err: assertClickableGameValue(&world{appGame: app.NewGame()}, nil, "command", "path entry", func(*app.Game) string { return "x" }), want: "missing example value command"},
		{name: "game value missing concrete app", err: assertClickableGameValue(w, map[string]string{"command": "x"}, "command", "path entry", clickablePathEntryCommand), want: "application was not started"},
		{name: "click match missing shortcut", err: assertClickMatchesShortcut(&world{}, nil), want: "missing example value shortcut"},
		{name: "click match wrong recorded shortcut", err: assertClickMatchesShortcut(&world{clickShortcut: "Enter"}, map[string]string{"shortcut": "Space"}), want: "recorded shortcut"},
		{name: "click match missing concrete app", err: assertClickMatchesShortcut(&world{appGame: nonConcreteAppGame{}, clickShortcut: "Space"}, map[string]string{"shortcut": "Space"}), want: "application was not started"},
		{name: "click match invalid shortcut", err: assertClickMatchesShortcut(&world{appGame: app.NewGame(), clickShortcut: "Nope"}, map[string]string{"shortcut": "Nope"}), want: "shortcut \"Nope\" was not handled"},
		{name: "outside click missing concrete app", err: clickOutsideVisibleControls(w, nil), want: "application was not started"},
		{name: "unchanged state missing concrete app", err: assertClickableApplicationStateUnchanged(w, nil), want: "application was not started"},
		{name: "set simulation missing state", err: setClickableSimulationState(&world{}, nil), want: "missing example value old_state"},
		{name: "set simulation missing concrete app", err: setClickableSimulationState(w, map[string]string{"old_state": "paused"}), want: "application was not started"},
		{name: "assert simulation missing state", err: assertClickableSimulationState(&world{}, nil), want: "missing example value new_state"},
		{name: "assert simulation missing concrete app", err: assertClickableSimulationState(w, map[string]string{"new_state": "paused"}), want: "application was not started"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil || !strings.Contains(tt.err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", tt.err, tt.want)
			}
		})
	}
}

func TestDragAppMassUpdatesWorld(t *testing.T) {
	game := app.NewGame()
	domain := sim.NewWorld()
	domain.Parameters.Set("grid snap", "0")
	_ = domain.AddMass(sim.Mass{ID: 7, Position: sim.Vec2{}, Mass: 1})
	w := &world{domainWorld: domain}

	if err := dragAppMass(w, game, 7, sim.Vec2{X: 103, Y: 104}); err != nil {
		t.Fatal(err)
	}
	mass, ok := w.domainWorld.MassByID(7)
	if !ok || mass.Position != (sim.Vec2{X: 103, Y: 104}) {
		t.Fatalf("dragged mass = %#v ok=%t", mass, ok)
	}
	if err := dragAppMass(w, game, 99, sim.Vec2{}); err == nil {
		t.Fatal("expected missing mass drag error")
	}
}

type nonConcreteAppGame struct{}

func (nonConcreteAppGame) Update() error                        { return nil }
func (nonConcreteAppGame) RenderFrame()                         {}
func (nonConcreteAppGame) RenderWorld() renderResult            { return renderResult{} }
func (nonConcreteAppGame) World() *sim.Simulation               { return sim.NewWorld() }
func (nonConcreteAppGame) SetPaused(bool)                       {}
func (nonConcreteAppGame) EditorScreen() editorScreen           { return editorScreen{} }
func (nonConcreteAppGame) SetSelected(bool)                     {}
func (nonConcreteAppGame) SetDirty(bool)                        {}
func (nonConcreteAppGame) HandleShortcut(string) bool           { return false }
func (nonConcreteAppGame) LastCommand() string                  { return "" }
func (nonConcreteAppGame) DrawFrameReport() app.DrawFrameReport { return app.DrawFrameReport{} }
func (nonConcreteAppGame) InputActive() bool                    { return false }
func (nonConcreteAppGame) RenderingActive() bool                { return false }
func (nonConcreteAppGame) Close() error                         { return nil }
func (nonConcreteAppGame) Closed() bool                         { return false }

func TestMouseMassExpectedIDAssertion(t *testing.T) {
	domain := sim.NewWorld()
	_ = domain.AddMass(sim.Mass{ID: 3, Mass: 1})
	w := &world{domainWorld: domain}

	if err := assertMouseMassExpectedID(w, map[string]string{"mass_id": "3", "expected_mass_id": "3"}); err != nil {
		t.Fatal(err)
	}
	if err := assertMouseMassExpectedID(w, map[string]string{"mass_id": "3", "expected_mass_id": "4"}); err == nil {
		t.Fatal("expected mismatched mass id")
	}
}

func TestCollisionMassPropertiesSetter(t *testing.T) {
	domain := sim.NewWorld()
	_ = domain.AddMass(sim.Mass{ID: 2, Mass: 1})
	w := &world{domainWorld: domain}
	example := map[string]string{
		"mass":       "2",
		"mass_value": "5",
		"elasticity": "0.25",
		"fixed":      "true",
	}

	if err := setCollisionMassProperties(w, example, "mass", "mass_value", "elasticity", "fixed"); err != nil {
		t.Fatal(err)
	}
	mass, _ := domain.MassByID(2)
	if mass.Mass != 5 || mass.Elasticity != 0.25 || !mass.Fixed {
		t.Fatalf("mass properties = %#v", mass)
	}
	example["mass"] = "99"
	if err := setCollisionMassProperties(w, example, "mass", "mass_value", "elasticity", "fixed"); err == nil {
		t.Fatal("expected missing mass error")
	}
}

func TestDocumentedCommandRunnerChecksPrerequisites(t *testing.T) {
	if err := runDocumentedCommand(&world{}, map[string]string{"command": "run"}); err == nil {
		t.Fatal("expected clean checkout prerequisite error")
	}
	w := &world{cleanCheckout: true}
	if err := runDocumentedCommand(w, map[string]string{"command": "run"}); err != nil {
		t.Fatal(err)
	}
	if w.documentedCommand != "run" || w.documentedCommandErr != nil {
		t.Fatalf("documented command state = %q %v", w.documentedCommand, w.documentedCommandErr)
	}
	if err := runDocumentedCommand(&world{cleanCheckout: true}, map[string]string{"command": "unknown"}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported command error, got %v", err)
	}
}

func TestControlWorldStateIncludesSeedMassAndParameter(t *testing.T) {
	w := &world{}
	if err := createControlWorldState(w, nil); err != nil {
		t.Fatal(err)
	}
	game := w.appGame.(*app.Game)
	mass, ok := game.World().MassByID(1)
	if !ok {
		t.Fatal("seed mass missing")
	}
	if mass.ID != 1 || mass.Mass != 1 {
		t.Fatalf("seed mass = %#v", mass)
	}
	if got := game.World().Parameters.Value("current mass"); got != controlCurrentValue {
		t.Fatalf("current mass parameter = %q", got)
	}
}

func TestSetCustomSystemParametersRequiresConcreteGame(t *testing.T) {
	if err := setCustomSystemParameters(&world{}, nil); err == nil || !strings.Contains(err.Error(), "application is not running") {
		t.Fatalf("missing app error = %v", err)
	}
	w := &world{appGame: app.NewGame()}
	if err := setCustomSystemParameters(w, nil); err != nil {
		t.Fatal(err)
	}
	game := w.appGame.(*app.Game)
	if got := game.World().Parameters.Value("current mass"); got != controlCustomValue {
		t.Fatalf("current mass parameter = %q", got)
	}
}

func TestControlParameterHelpersRequireConcreteGameAndPreserveDefault(t *testing.T) {
	if err := assertControlParametersDefault(&world{}, nil); err == nil || !strings.Contains(err.Error(), "application is not running") {
		t.Fatalf("missing app error = %v", err)
	}
	var nilGame *app.Game
	if _, err := concreteGame(&world{appGame: nilGame}); err == nil || !strings.Contains(err.Error(), "application is not running") {
		t.Fatalf("nil concrete app error = %v", err)
	}

	w := &world{}
	example := map[string]string{"parameter": "current mass", "old_value": "default"}
	if err := setControlParameterValue(w, example); err != nil {
		t.Fatal(err)
	}
	game := w.appGame.(*app.Game)
	if got := game.World().Parameters.Value("current mass"); got == "default" {
		t.Fatalf("current mass parameter = %q", got)
	}

	if err := assertControlParameterValue(&world{}, map[string]string{"parameter": "current mass", "new_value": "custom"}); err == nil || !strings.Contains(err.Error(), "application is not running") {
		t.Fatalf("missing app assertion error = %v", err)
	}
	if err := assertControlParameterValue(&world{appGame: app.NewGame()}, nil); err == nil || !strings.Contains(err.Error(), "missing example value parameter") {
		t.Fatalf("missing parameter assertion error = %v", err)
	}
	parameter, value, err := controlParameterAndValue(map[string]string{"parameter": "current mass", "old_value": "custom"}, "old_value")
	if err != nil {
		t.Fatal(err)
	}
	if parameter != "current mass" || value != controlCustomValue {
		t.Fatalf("parameter/value = %q/%q", parameter, value)
	}
	if _, _, err := controlParameterAndValue(nil, "old_value"); err == nil || !strings.Contains(err.Error(), "missing example value parameter") {
		t.Fatalf("missing parameter parse error = %v", err)
	}
}

func TestControlWorldAssertionsDistinguishMassIDs(t *testing.T) {
	loadedWithCurrent := app.NewGame()
	if err := loadedWithCurrent.LoadXSP(controlFileXSP()); err != nil {
		t.Fatal(err)
	}
	if err := loadedWithCurrent.World().AddMass(sim.Mass{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := assertLoadedControlWorld(loadedWithCurrent); err == nil || !strings.Contains(err.Error(), "current mass was not replaced") {
		t.Fatalf("loaded current-mass error = %v", err)
	}

	loadedWithZero := app.NewGame()
	if err := loadedWithZero.LoadXSP(controlFileXSP()); err != nil {
		t.Fatal(err)
	}
	if err := loadedWithZero.World().AddMass(sim.Mass{ID: 0}); err != nil {
		t.Fatal(err)
	}
	if err := assertLoadedControlWorld(loadedWithZero); err != nil {
		t.Fatal(err)
	}

	inserted := app.NewGame()
	if err := inserted.LoadXSP(controlFileXSP()); err != nil {
		t.Fatal(err)
	}
	if err := inserted.World().AddMass(sim.Mass{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := assertInsertedControlWorld(inserted); err != nil {
		t.Fatal(err)
	}

	insertedWithoutCurrent := app.NewGame()
	if err := insertedWithoutCurrent.LoadXSP(controlFileXSP()); err != nil {
		t.Fatal(err)
	}
	if err := assertInsertedControlWorld(insertedWithoutCurrent); err == nil || !strings.Contains(err.Error(), "current mass missing") {
		t.Fatalf("inserted current-mass error = %v", err)
	}
}
