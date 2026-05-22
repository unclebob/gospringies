//go:build !appunit

package app

import (
	"errors"
	"image"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func TestNewGameLoadsPendulumDemoAsStartupWorld(t *testing.T) {
	game := NewGame()
	expected := loadAppTestXSP(t, filepath.Join("..", "..", DefaultStartupScenePath()))
	expected.Bounds = game.World().Bounds

	if !reflect.DeepEqual(game.World(), expected) {
		t.Fatalf("startup world = %#v, want %#v", game.World(), expected)
	}
}

func TestNewGameStartsWithNonblankStarterWorld(t *testing.T) {
	game := NewGame()

	assertStarterObjects(t, game.World())
}

func TestStartupPendulumEnablesGravity(t *testing.T) {
	game := NewGame()
	force, ok := game.World().Parameters.Force("gravity")
	if !ok {
		t.Fatal("missing gravity force")
	}
	if force.Enabled != "true" || force.Values["magnitude"] != "10" || force.Values["direction"] != "0" {
		t.Fatalf("gravity force = %#v", force)
	}
}

func loadAppTestXSP(t *testing.T, path string) *sim.Simulation {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	world, err := xspfmt.LoadXSP(string(content))
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	setAppBounds(world)
	return world
}

func TestGameLayoutUsesWindowSize(t *testing.T) {
	game := NewGame()
	width, height := game.Layout(1, 1)
	if width != screenWidth || height != screenHeight {
		t.Fatalf("layout = %d, %d", width, height)
	}
}

func TestNewGameStartsWithYUpCanvas(t *testing.T) {
	game := NewGame()
	if !game.canvasYUp {
		t.Fatal("new game should use a y-up canvas")
	}
}

func TestAdvanceSimulationFrameRequiresPositiveSpeed(t *testing.T) {
	game := NewGame()
	game.paused = false
	game.simulationSpeed = -1

	game.advanceSimulationFrame()

	if game.World().Time != 0 {
		t.Fatalf("world time = %f, want unchanged", game.World().Time)
	}
}

func TestZeroSpeedDoesNotPinDraggingMass(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.paused = false
	game.simulationSpeed = 0
	game.pointer.draggingMassID = 1
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {X: 1, Y: 1}}
	game.pointer.lastCursor = sim.Vec2{X: 50, Y: 50}

	game.advanceSimulationFrame()

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 10, Y: 10}) {
		t.Fatalf("mass position = %#v, want unchanged", mass.Position)
	}
}

func TestAppWorldBoundsUseWindowSize(t *testing.T) {
	game := NewGame()
	want := sim.Bounds{Width: screenWidth, Height: screenHeight, Left: 72, Right: 1340, Bottom: 0, Top: 964}
	if game.World().Bounds != want {
		t.Fatalf("startup bounds = %#v, want %#v", game.World().Bounds, want)
	}

	if err := game.LoadXSP("#1.0\nmass 1 10 20 1 0\n"); err != nil {
		t.Fatal(err)
	}
	yDownWant := sim.Bounds{Width: screenWidth, Height: screenHeight, Left: 72, Right: 1340, Bottom: 36, Top: 1000}
	if game.World().Bounds != yDownWant {
		t.Fatalf("loaded bounds = %#v, want %#v", game.World().Bounds, yDownWant)
	}

	replacement := sim.NewWorld()
	staleEditorWorld := sim.NewWorld()
	game.editing().World = staleEditorWorld
	game.ReplaceWorld(replacement)
	if game.World().Bounds != yDownWant {
		t.Fatalf("replacement bounds = %#v, want %#v", game.World().Bounds, yDownWant)
	}
	if game.editor.World != game.simulation || game.editor.World == staleEditorWorld {
		t.Fatal("replacement should reattach editor world to game simulation")
	}
}

func TestValueDialogInputKeepsNumericCharacters(t *testing.T) {
	game := NewGame()
	game.valueDialog.Text = "1"

	game.appendValueDialogInput([]rune{'2', 'x', '.', '-', '3'})

	if game.valueDialog.Text != "12.-3" {
		t.Fatalf("dialog text = %q", game.valueDialog.Text)
	}
}

func TestValueDialogFractionClampsParsedText(t *testing.T) {
	game := NewGame()
	game.valueDialog = valueDialog{Text: "15", Min: 0, Max: 10}
	if got := game.valueDialogFraction(); got != 1 {
		t.Fatalf("high fraction = %f", got)
	}

	game.valueDialog.Text = "-5"
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("low fraction = %f", got)
	}

	game.valueDialog.Text = "5"
	if got := game.valueDialogFraction(); got != 0.5 {
		t.Fatalf("mid fraction = %f", got)
	}
}

func TestSelectedMassIDsReturnsOnlySelectedMasses(t *testing.T) {
	game := NewGame()
	editor := game.editing()
	editor.SelectedMasses = map[int]bool{1: true, 2: false, 3: true}

	ids := game.selectedMassIDs()

	if !reflect.DeepEqual(mapFromInts(ids), map[int]bool{1: true, 3: true}) {
		t.Fatalf("selected mass ids = %#v", ids)
	}
}

func mapFromInts(values []int) map[int]bool {
	result := map[int]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}

func TestMoveSelectedMassesMovesOnlySelectedMasses(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 101, Y: 101}},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 105, Y: 105}},
	)
	game.World().Parameters.Set("grid snap", "0")
	_ = game.editing().SelectMass(1)

	game.moveSelectedMasses(sim.Vec2{X: 2, Y: 3})

	if got := game.World().Masses[0].Position; got != (sim.Vec2{X: 103, Y: 104}) {
		t.Fatalf("selected mass position = %#v", got)
	}
	if got := game.World().Masses[1].Position; got != (sim.Vec2{X: 105, Y: 105}) {
		t.Fatalf("unselected mass position = %#v", got)
	}
}

func TestThrowSingleDraggingMassSetsVelocity(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 7})
	game.pointer.draggingMassID = 7

	game.throwSingleDraggingMass(sim.Vec2{X: 4, Y: -2})

	if got := game.World().Masses[0].Velocity; got != (sim.Vec2{X: 4, Y: -2}) {
		t.Fatalf("velocity = %#v", got)
	}
	if !game.dirty {
		t.Fatal("throw should mark game dirty")
	}
}

func TestDrawCoversOpenOverlayBranches(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 120, Y: 120}})
	_ = game.editing().SelectMass(1)
	game.selected = true
	game.demoFiles = []string{"demos/pendulum.xsp", "demos/spring-chain.xsp"}
	game.demoPickerOpen = true
	game.massMenu = massContextMenu{Open: true, MassID: 1, X: 120, Y: 120}
	game.valueDialog = valueDialog{Open: true, Title: "Value", Text: "1", Min: 0, Max: 2}

	game.Draw(ebiten.NewImage(screenWidth, screenHeight))
}

func TestAnyKeyPressedChecksEveryCandidate(t *testing.T) {
	keys := []ebiten.Key{ebiten.KeyA, ebiten.KeyB, ebiten.KeyC}
	checked := map[ebiten.Key]bool{}
	pressed := func(key ebiten.Key) bool {
		checked[key] = true
		return key == ebiten.KeyC
	}

	if !anyKeyPressed(keys, pressed) {
		t.Fatal("expected final candidate to be reported pressed")
	}
	for _, key := range keys {
		if !checked[key] {
			t.Fatalf("key %v was not checked", key)
		}
	}
	if anyKeyPressed(keys[:2], func(ebiten.Key) bool { return false }) {
		t.Fatal("unexpected pressed key")
	}
}

func TestShiftDownBypassesEbitenKeyState(t *testing.T) {
	game := NewGame()
	game.shiftDown = true

	if !game.shiftKeyPressed() {
		t.Fatal("shiftDown should count as shift pressed")
	}
}

func gameWithMasses(masses ...sim.Mass) *Game {
	game := NewGame()
	world := sim.NewWorld()
	world.Masses = append(world.Masses, masses...)
	game.ReplaceWorld(world)
	return game
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
	screen := ebiten.NewImage(screenWidth, screenHeight)

	game.Draw(screen)

	if !game.RenderingActive() {
		t.Fatal("expected rendering to be active")
	}
}

func TestDrawFrameRendersVisibleControlRegions(t *testing.T) {
	report := NewGame().DrawFrameReport()

	for _, region := range []string{"left toolbar", "top command bar", "right inspector"} {
		if report.RegionPixels[region] == 0 {
			t.Fatalf("region %q had no visible pixels: %#v", region, report.RegionPixels)
		}
		if report.RegionControlCounts[region] == 0 {
			t.Fatalf("region %q had no controls: %#v", region, report.RegionControlCounts)
		}
	}
}

func TestDrawFrameRendersReadableControlLabels(t *testing.T) {
	report := NewGame().DrawFrameReport()
	expected := map[string]string{
		"edit menu":               "Edit",
		"run command":             "Run",
		"pause command":           "Pause",
		"reset command":           "Reset",
		"save state command":      "State+",
		"restore state command":   "State",
		"load command":            "Load",
		"insert command":          "Insert",
		"save command":            "Save",
		"quit command":            "Quit",
		"mass label":              "Mass:",
		"elasticity label":        "Elasticity:",
		"gravity force":           "",
		"gravity label":           "Gravity:",
		"center attraction force": "",
		"center mass force":       "",
		"wall repulsion force":    "",
		"wall repulsion label":    "Wall Rep:",
		"mass collision force":    "Collide",
		"top wall toggle":         "T",
		"bottom wall toggle":      "B",
		"left wall toggle":        "L",
		"right wall toggle":       "R",
		"grid snap toggle":        "Grid",
		"show springs toggle":     "Springs",
		"viscosity label":         "Viscosity:",
		"speed label":             "Speed:",
	}
	for control, label := range expected {
		if report.Controls[control] != label {
			t.Fatalf("control %q label = %q, want %q", control, report.Controls[control], label)
		}
	}
	if !report.ControlLabelsFit {
		t.Fatal("expected visible control labels to fit")
	}
}

func TestEditMenuShowsStandardItems(t *testing.T) {
	game := NewGame()
	control, ok := visibleControlWithName("edit menu")
	if !ok {
		t.Fatal("missing edit menu")
	}

	if !game.ClickAt(control.Rect.Min.X+2, control.Rect.Min.Y+2) {
		t.Fatal("edit menu click was not handled")
	}

	expected := map[string]string{
		"cut command":   "Cut     Ctrl+X",
		"copy command":  "Copy    Ctrl+C",
		"paste command": "Paste   Ctrl+V",
	}
	for _, control := range game.editMenuControls() {
		if expected[control.Name] != control.Label {
			t.Fatalf("edit menu control %q label = %q", control.Name, control.Label)
		}
		delete(expected, control.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing edit menu controls: %#v", expected)
	}
}

func TestDrawFrameRendersInspectorAndStatusFields(t *testing.T) {
	report := NewGame().DrawFrameReport()
	for _, section := range []string{
		sectionHeaderLabel("Selected Mass(es)"),
		sectionHeaderLabel("Selected Spring(s)"),
		sectionHeaderLabel("Forces"),
		sectionHeaderLabel("Simulation"),
		sectionHeaderLabel("Display"),
	} {
		if !report.InspectorSections[section] {
			t.Fatalf("missing inspector section %q: %#v", section, report.InspectorSections)
		}
	}
	for field, expected := range map[string]string{
		"run state":     "running",
		"object counts": "object counts",
		"file state":    "saved",
	} {
		if !strings.Contains(report.StatusFields[field], expected) {
			t.Fatalf("status field %q = %q, want it to contain %q", field, report.StatusFields[field], expected)
		}
	}
}

func TestDrawFrameKeepsWorldContentVisible(t *testing.T) {
	report := NewGame().DrawFrameReport()

	if report.CanvasWorldPixels == 0 {
		t.Fatalf("expected visible world pixels in canvas: %#v", report)
	}
}

func TestEditorChromeRectsCoverStartupRegions(t *testing.T) {
	rects := editorChromeRects()
	if len(rects) != 3 {
		t.Fatalf("chrome rect count = %d", len(rects))
	}
	for _, rect := range rects {
		if rect.width <= 0 || rect.height <= 0 {
			t.Fatalf("invalid chrome rect = %#v", rect)
		}
	}
}

func TestDrawWallsUsesSimulationBounds(t *testing.T) {
	lines := wallDrawLines(sim.Bounds{Width: 20, Height: 16})
	expected := []wallDrawLine{
		{name: "top", x1: 0, y1: 15, x2: 20, y2: 15},
		{name: "bottom", x1: 0, y1: 0, x2: 20, y2: 0},
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

func TestDrawWallsUsesExplicitCanvasBounds(t *testing.T) {
	lines := wallDrawLines(sim.Bounds{Width: 100, Height: 80, Left: 10, Right: 70, Bottom: 20, Top: 60})
	expected := []wallDrawLine{
		{name: "top", x1: 10, y1: 59, x2: 70, y2: 59},
		{name: "bottom", x1: 10, y1: 20, x2: 70, y2: 20},
		{name: "left", x1: 10, y1: 20, x2: 10, y2: 60},
		{name: "right", x1: 69, y1: 20, x2: 69, y2: 60},
	}
	if !reflect.DeepEqual(lines, expected) {
		t.Fatalf("wall lines = %#v, want %#v", lines, expected)
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
	x, y, radius := massDrawCircle(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}, Mass: 1})

	if x != 30 || y != 40 || radius != 3 {
		t.Fatalf("draw circle = %f,%f radius %f", x, y, radius)
	}
	_, _, fixedRadius := massDrawCircle(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}, Mass: 1, Fixed: true})
	if fixedRadius != 4 {
		t.Fatalf("fixed draw radius = %f", fixedRadius)
	}
}

func TestMassDrawingUsesAntialiasing(t *testing.T) {
	if !massDrawAntiAlias() {
		t.Fatal("expected mass circles to be antialiased")
	}
}

func TestSelectionOutlineSurroundsMassPosition(t *testing.T) {
	lines := selectionOutline(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}, Mass: 1})
	expected := []selectionLine{
		{24, 34, 36, 34},
		{36, 34, 36, 46},
		{36, 46, 24, 46},
		{24, 46, 24, 34},
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

func TestSelectedMassOutlineUsesEveryMass(t *testing.T) {
	lines := selectedMassOutline([]sim.Mass{
		{Position: sim.Vec2{X: 30, Y: 40}},
		{Position: sim.Vec2{X: 80, Y: 90}},
	})
	expected := append(
		selectionOutline(sim.Mass{Position: sim.Vec2{X: 30, Y: 40}}),
		selectionOutline(sim.Mass{Position: sim.Vec2{X: 80, Y: 90}})...,
	)

	if len(lines) != len(expected) {
		t.Fatalf("line count = %d", len(lines))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("line %d = %#v, want %#v", i, line, expected[i])
		}
	}
}

func TestSelectedSpringLinesUseEverySelectedSpring(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 40}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 50, Y: 60}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 1, SpringConstant: 1})
	_ = world.AddSpring(sim.Spring{ID: 2, MassA: 2, MassB: 3, RestLength: 1, SpringConstant: 1})
	game.ReplaceWorld(world)
	game.editing().SelectedSprings[1] = true
	game.editing().SelectedSprings[2] = true

	lines := game.selectedSpringLines()
	expected := []selectionLine{
		{x1: 10, y1: 20, x2: 30, y2: 40},
		{x1: 30, y1: 40, x2: 50, y2: 60},
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

func TestSpringOnlySelectionDoesNotHighlightMasses(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 40}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 1, SpringConstant: 1})
	game.ReplaceWorld(world)
	game.editing().SelectedSprings[3] = true
	game.syncSelectionState()

	if masses := game.selectedMasses(); len(masses) != 0 {
		t.Fatalf("selected masses = %#v, want none", masses)
	}
	if lines := game.selectedSpringLines(); len(lines) != 1 {
		t.Fatalf("selected spring lines = %#v, want one", lines)
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

func TestSpringThicknessIsDoubled(t *testing.T) {
	if springThickness != 2 {
		t.Fatalf("spring thickness = %f, want 2", springThickness)
	}
}

func TestGridPointsFollowGridSnapSetting(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("grid snap", "10")

	points := game.gridPoints()

	if len(points) == 0 {
		t.Fatal("expected grid points when grid snap is enabled")
	}
	for _, point := range points {
		if math.Mod(point.X, 10) != 0 || math.Mod(point.Y, 10) != 0 {
			t.Fatalf("grid point = %#v, want multiples of 10", point)
		}
	}

	game.World().Parameters.Set("grid snap", "0")
	if points := game.gridPoints(); len(points) != 0 {
		t.Fatalf("grid point count = %d, want none when grid snap is disabled", len(points))
	}
}

func TestGridPointsIncludeCanvasEdgesOnlyInsideDrawableRegion(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("grid snap", "10")

	points := game.gridPoints()
	if len(points) == 0 {
		t.Fatal("expected grid points")
	}

	canvas := visibleRegionRects()["canvas"]
	_, _, minY, maxY := game.canvasWorldBounds()
	for _, point := range points {
		if point.X < float64(canvas.Min.X) || point.X >= float64(canvas.Max.X) {
			t.Fatalf("grid point x = %f outside canvas %#v", point.X, canvas)
		}
		if point.Y < minY || point.Y > maxY {
			t.Fatalf("grid point y = %f outside canvas bounds %f..%f", point.Y, minY, maxY)
		}
	}
}

func TestUnitGridSnapStillDrawsGridPoints(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("grid snap", "1")

	if points := game.gridPoints(); len(points) == 0 {
		t.Fatal("expected grid points for unit grid snap")
	}
}

func TestValidGridSnapSizeRequiresPositiveSize(t *testing.T) {
	if validGridSnapSize(0) {
		t.Fatal("zero grid snap should be invalid")
	}
	if validGridSnapSize(-1) {
		t.Fatal("negative grid snap should be invalid")
	}
	if !validGridSnapSize(1) {
		t.Fatal("positive grid snap should be valid")
	}
}

func TestEditorChromeRectsMatchApplicationChrome(t *testing.T) {
	want := []chromeRect{
		{x: 0, y: 0, width: screenWidth, height: topBarHeight, color: chromeColor},
		{x: 0, y: topBarHeight, width: toolbarWidth, height: screenHeight - topBarHeight, color: panelColor},
		{x: screenWidth - inspectorWidth, y: topBarHeight, width: inspectorWidth, height: screenHeight - topBarHeight, color: panelColor},
	}

	if got := editorChromeRects(); !reflect.DeepEqual(got, want) {
		t.Fatalf("editor chrome rects = %#v, want %#v", got, want)
	}
}

func TestDrawHelpersUseCrispUnaliasedGeometry(t *testing.T) {
	if editorChromeAntiAlias() {
		t.Fatal("editor chrome should draw without antialiasing")
	}
	if gridPointPixelSize() != 1 {
		t.Fatalf("grid point size = %f, want 1", gridPointPixelSize())
	}
	if gridPointAntiAlias() {
		t.Fatal("grid points should draw without antialiasing")
	}
	if springLineAntiAlias() {
		t.Fatal("spring lines should draw without antialiasing")
	}

	game := NewGame()
	game.World().Parameters.Set("grid snap", "10")
	rects := game.gridPointRects()
	if len(rects) == 0 {
		t.Fatal("expected grid point draw rects")
	}
	first := rects[0]
	if first.width != 1 || first.height != 1 || first.color != gridPointColor || first.antiAlias {
		t.Fatalf("grid point rect = %#v", first)
	}
}

func TestClickCreatedMassSnapsToGridPoint(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	world.Parameters.Set("grid snap", "10")
	game.ReplaceWorld(world)

	game.handlePointer(true, 123, 87)
	game.handlePointer(false, 123, 87)

	mass, ok := game.World().MassByID(1)
	if !ok {
		t.Fatal("mass was not created")
	}
	if mass.Position != (sim.Vec2{X: 120, Y: 90}) {
		t.Fatalf("mass position = %#v, want snapped grid point 120,90", mass.Position)
	}
}

func TestCreateMassAtRejectsPositionsOutsideCanvasBounds(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	world.Parameters.Set("grid snap", "0")
	game.ReplaceWorld(world)

	if id, ok := game.createMassAt(sim.Vec2{X: -10, Y: -20}, false); ok || id != 0 {
		t.Fatalf("createMassAt = %d, %t; want rejected", id, ok)
	}
	if len(game.World().Masses) != 0 {
		t.Fatalf("mass count = %d, want 0", len(game.World().Masses))
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
			world := sim.NewWorld()
			world.Masses = append(world.Masses, test.masses...)
			game.ReplaceWorld(world)

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
	game.ReplaceWorld(sim.NewWorld())

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
		{name: "valid mass ids", spring: sim.Spring{MassA: 1, MassB: 2}, valid: true},
		{name: "negative a", spring: sim.Spring{A: -1, B: 1}},
		{name: "negative b", spring: sim.Spring{A: 0, B: -1}},
		{name: "a too high", spring: sim.Spring{A: 2, B: 1}},
		{name: "b too high", spring: sim.Spring{A: 0, B: 2}},
		{name: "missing mass a", spring: sim.Spring{MassA: 9, MassB: 2}},
		{name: "missing mass b", spring: sim.Spring{MassA: 1, MassB: 9}},
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

func TestCoordinateTransformsRespectCanvasOrientation(t *testing.T) {
	game := NewGame()
	game.World().Bounds.Height = 100
	point := sim.Vec2{X: 25, Y: 30}

	if got := game.screenToWorld(point); got != (sim.Vec2{X: 25, Y: 70}) {
		t.Fatalf("screenToWorld y-up = %#v", got)
	}
	if got := game.worldToScreen(point); got != (sim.Vec2{X: 25, Y: 70}) {
		t.Fatalf("worldToScreen y-up = %#v", got)
	}

	game.canvasYUp = false
	if got := game.screenToWorld(point); got != point {
		t.Fatalf("screenToWorld y-down = %#v", got)
	}
	if got := game.worldToScreen(point); got != point {
		t.Fatalf("worldToScreen y-down = %#v", got)
	}
}

func TestEditingRebuildsWhenWorldChanges(t *testing.T) {
	game := NewGame()
	editor := game.editing()
	replacement := sim.NewWorld()
	game.simulation = replacement

	if got := game.editing(); got == editor || got.World != replacement {
		t.Fatalf("editor was not rebuilt for replacement world")
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
		"left toolbar":    "run selection commands",
		"top bar":         "run commands and file commands",
		"right inspector": "edit selected objects and world parameters and show simulation state",
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

func TestEditorScreenHasVisibleCommandControls(t *testing.T) {
	screen := NewGame().EditorScreen()
	for _, command := range []string{"run", "pause", "reset", "load", "insert", "save", "quit", "delete", "select all", "cut", "copy", "paste"} {
		if !screen.HasCommandControl(command) {
			t.Fatalf("missing command %q", command)
		}
	}
	if screen.HasCommandControl("export") {
		t.Fatal("unexpected export command")
	}
}

func TestEditorIndicatorsReflectState(t *testing.T) {
	game := NewGame()
	game.SetPaused(true)
	game.SetSelected(true)
	game.SetDirty(true)

	indicators := game.EditorScreen().Indicators

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

func TestSelectSpringsSelectsEveryRequestedSpring(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	_ = world.AddSpring(sim.Spring{ID: 2, MassA: 1, MassB: 2})
	game.ReplaceWorld(world)

	if err := game.SelectSprings(1, 2); err != nil {
		t.Fatalf("SelectSprings returned error: %v", err)
	}
	if !game.editing().SelectedSprings[1] || !game.editing().SelectedSprings[2] {
		t.Fatalf("selected springs = %#v", game.editing().SelectedSprings)
	}
	if err := game.SelectSprings(3); err == nil {
		t.Fatal("missing spring selection should fail")
	}
}

func TestSetSpringTemperatureUpdatesMatchingSpring(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	game.ReplaceWorld(world)

	game.setSpringTemperature(1, 7.5)
	spring, _ := game.World().SpringByID(1)
	if spring.Temperature != 7.5 {
		t.Fatalf("spring temperature = %f", spring.Temperature)
	}
	if game.EditorScreen().Indicators["file state"] != "unsaved changes" {
		t.Fatalf("set temperature should mark file dirty: %#v", game.EditorScreen().Indicators)
	}
	game.setSpringTemperature(2, 3)
	spring, _ = game.World().SpringByID(1)
	if spring.Temperature != 7.5 {
		t.Fatalf("missing spring changed temperature to %f", spring.Temperature)
	}
}

func TestKeyboardShortcutsRunVisibleCommands(t *testing.T) {
	game := NewGame()
	shortcuts := map[string]string{
		"Space":  "pause",
		"Delete": "delete",
		"Ctrl+A": "select all",
		"Ctrl+X": "cut",
		"Ctrl+C": "copy",
		"Ctrl+V": "paste",
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
	if game.HandleShortcut("Ctrl+Z") {
		t.Fatal("unexpected unknown shortcut handling")
	}
}

func TestCommandsAffectApplicationState(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("current mass", "custom")

	game.RunCommand("pause")
	if !game.Paused() {
		t.Fatal("expected pause command to toggle paused state")
	}

	game.RunCommand("reset")
	if len(game.World().Masses) != 0 || game.World().Parameters.Value("current mass") == "custom" {
		t.Fatalf("reset world = %#v", game.World())
	}
	if game.EditorScreen().Indicators["file state"] != "saved" {
		t.Fatalf("reset indicators = %#v", game.EditorScreen().Indicators)
	}

	game.RunCommand("quit")
	if !game.Closed() {
		t.Fatal("expected quit command to close game")
	}
}

func TestCopyPasteDuplicatesSelectedObjects(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 2})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 40}, Mass: 3})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 20, SpringConstant: 12})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.editing().SelectedSprings[1] = true

	game.RunCommand("copy")
	game.pointer.lastCursor = sim.Vec2{X: 100, Y: 120}
	game.RunCommand("paste")

	if len(game.World().Masses) != 4 {
		t.Fatalf("mass count = %d, want 4", len(game.World().Masses))
	}
	if len(game.World().Springs) != 2 {
		t.Fatalf("spring count = %d, want 2", len(game.World().Springs))
	}
	pasted, ok := game.World().MassByID(3)
	if !ok || pasted.Position != (sim.Vec2{X: 100, Y: 120}) {
		t.Fatalf("pasted mass = %#v, ok=%t", pasted, ok)
	}
	secondPasted, ok := game.World().MassByID(4)
	if !ok || secondPasted.Position != (sim.Vec2{X: 120, Y: 140}) {
		t.Fatalf("second pasted mass = %#v, ok=%t", secondPasted, ok)
	}
	if !game.editing().MassSelected(3) || !game.editing().MassSelected(4) || !game.editing().SpringSelected(2) {
		t.Fatalf("pasted selection = %#v %#v", game.editing().SelectedMasses, game.editing().SelectedSprings)
	}
	if !game.selected || !game.dirty {
		t.Fatalf("paste state selected=%t dirty=%t", game.selected, game.dirty)
	}
}

func TestCutCopiesThenDeletesSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 2})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)

	game.RunCommand("cut")

	if len(game.World().Masses) != 0 {
		t.Fatalf("mass count after cut = %d, want 0", len(game.World().Masses))
	}
	if game.selected || !game.dirty {
		t.Fatalf("cut state selected=%t dirty=%t", game.selected, game.dirty)
	}
	game.pointer.lastCursor = sim.Vec2{X: 200, Y: 220}
	game.RunCommand("paste")
	if len(game.World().Masses) != 1 {
		t.Fatalf("mass count after paste = %d, want 1", len(game.World().Masses))
	}
}

func TestSyncSelectionStateClearsEmptySelection(t *testing.T) {
	game := NewGame()
	game.selected = true

	game.syncSelectionState()

	if game.selected {
		t.Fatal("empty selection should clear selected state")
	}
}

func TestClickVisibleCommandControlsRunCommands(t *testing.T) {
	tests := map[string]string{
		"Run":    "run",
		"Pause":  "pause",
		"Reset":  "reset",
		"State+": "save state",
		"State":  "restore state",
		"Quit":   "quit",
	}
	for label, command := range tests {
		game := NewGame()

		if !game.ClickVisibleControl(label) {
			t.Fatalf("click %q was not handled", label)
		}
		if game.LastCommand() != command {
			t.Fatalf("command after %q = %q, want %q", label, game.LastCommand(), command)
		}
	}
}

func TestClickEditMenuRunsCutCopyPasteCommands(t *testing.T) {
	game := NewGame()
	editControl, _ := visibleControlWithName("edit menu")

	if !game.ClickAt(editControl.Rect.Min.X+2, editControl.Rect.Min.Y+2) {
		t.Fatal("edit menu click was not handled")
	}
	for _, expectedCommand := range []string{"cut", "copy", "paste"} {
		var menuItem controlBox
		for _, control := range game.editMenuControls() {
			if visibleControlCommands[control.Name] == expectedCommand {
				menuItem = control
			}
		}
		if !game.ClickAt(menuItem.Rect.Min.X+2, menuItem.Rect.Min.Y+2) {
			t.Fatalf("%s menu click was not handled", expectedCommand)
		}
		if game.LastCommand() != expectedCommand {
			t.Fatalf("last command = %q, want %q", game.LastCommand(), expectedCommand)
		}
		game.ClickAt(editControl.Rect.Min.X+2, editControl.Rect.Min.Y+2)
	}
}

func TestClickVisibleFileControlsOpenPathEntry(t *testing.T) {
	tests := map[string]string{"Insert": "Insert"}
	for label, command := range tests {
		game := NewGame()

		if !game.ClickVisibleControl(label) {
			t.Fatalf("click %q was not handled", label)
		}
		if game.PathEntryCommand() != command {
			t.Fatalf("path entry after %q = %q, want %q", label, game.PathEntryCommand(), command)
		}
	}
}

func TestSaveControlOpensFilenameDialogBeforeExtension(t *testing.T) {
	game := NewGame()

	if !game.ClickVisibleControl("Save") {
		t.Fatal("Save control click was not handled")
	}

	if !game.SaveFilenameDialogOpen() {
		t.Fatal("save filename dialog was not opened")
	}
	if game.SaveFilenameText() != ".xsp" {
		t.Fatalf("save filename text = %q, want .xsp", game.SaveFilenameText())
	}
	if game.SaveFilenameCursor() != 0 {
		t.Fatalf("save filename cursor = %d, want before extension", game.SaveFilenameCursor())
	}
}

func TestSaveFilenameDialogWritesNamedFileUnderSaves(t *testing.T) {
	withWorkingDirectory(t, t.TempDir())
	game := gameWithMasses(sim.Mass{ID: 7, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	game.openSaveFilenameDialog()

	game.EnterSaveFilenamePrefix("lab_scene")
	if err := game.SubmitSaveFilenameDialog(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join("saves", "lab_scene.xsp"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "\nmass 7 ") {
		t.Fatalf("saved content missing mass 7:\n%s", string(content))
	}
	if game.CurrentFilePath() != filepath.Join("saves", "lab_scene.xsp") {
		t.Fatalf("current file path = %q", game.CurrentFilePath())
	}
	if game.SaveFilenameDialogOpen() {
		t.Fatal("save dialog stayed open after submit")
	}
}

func TestSaveFilenameDialogDeleteAndClickPaths(t *testing.T) {
	withWorkingDirectory(t, t.TempDir())
	game := gameWithMasses(sim.Mass{ID: 7, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})

	game.saveDialog = saveFilenameDialog{Open: true, Text: ".xsp", Cursor: 0}
	game.deleteSaveFilenameCharacter()
	if game.SaveFilenameText() != ".xsp" || game.SaveFilenameCursor() != 0 {
		t.Fatalf("delete at start changed dialog text=%q cursor=%d", game.SaveFilenameText(), game.SaveFilenameCursor())
	}

	game.saveDialog = saveFilenameDialog{Open: true, Text: "abc.xsp", Cursor: 3}
	game.deleteSaveFilenameCharacter()
	if game.SaveFilenameText() != "ab.xsp" || game.SaveFilenameCursor() != 2 {
		t.Fatalf("delete middle text=%q cursor=%d", game.SaveFilenameText(), game.SaveFilenameCursor())
	}

	game.openSaveFilenameDialog()
	game.clickSaveFilenameDialog(saveFilenameDialogRect().Max.X+1, saveFilenameDialogRect().Max.Y+1)
	if game.SaveFilenameDialogOpen() {
		t.Fatal("outside click should close save dialog")
	}

	game.openSaveFilenameDialog()
	game.EnterSaveFilenamePrefix("clicked")
	ok := game.saveFilenameDialogOKRect()
	game.clickSaveFilenameDialog(ok.Min.X+1, ok.Min.Y+1)
	if game.SaveFilenameDialogOpen() || game.CurrentFilePath() != filepath.Join("saves", "clicked.xsp") {
		t.Fatalf("ok click left dialog open=%t path=%q", game.SaveFilenameDialogOpen(), game.CurrentFilePath())
	}
}

func TestLoadControlOpensDemoPicker(t *testing.T) {
	game := NewGame()
	game.demoPickerScroll = 3

	if !game.ClickVisibleControl("Load") {
		t.Fatal("Load control click was not handled")
	}

	if !game.demoPickerOpen {
		t.Fatal("demo picker was not opened")
	}
	if len(game.demoList()) == 0 {
		t.Fatal("expected demo files")
	}
	if game.demoPickerScroll != 0 {
		t.Fatalf("demo picker scroll = %d, want reset to 0", game.demoPickerScroll)
	}
}

func TestLoadPickerOrdersSavesThenSeparatorThenDemosAndOriginals(t *testing.T) {
	root := t.TempDir()
	withWorkingDirectory(t, root)
	mustWriteFile(t, filepath.Join(root, "saves", "lab_scene.xsp"), "#1.0\nmass 1 10 20 1 0\n")
	mustWriteFile(t, filepath.Join(root, "demos", "pendulum.xsp"), "#1.0\nmass 2 10 20 1 0\n")
	mustWriteFile(t, filepath.Join(root, "demos", "original", "pend.xsp"), "#1.0\nmass 3 10 20 1 0\n")

	game := NewGame()
	got := game.demoList()
	want := []string{
		filepath.Join("saves", "lab_scene.xsp"),
		loadPickerSeparator,
		filepath.Join("demos", "pendulum.xsp"),
		filepath.Join("demos", "original", "pend.xsp"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("load picker list = %#v, want %#v", got, want)
	}
}

func TestLoadPickerLoadsSavedFileAndTracksCurrentPath(t *testing.T) {
	root := t.TempDir()
	withWorkingDirectory(t, root)
	mustWriteFile(t, filepath.Join(root, "saves", "lab_scene.xsp"), "#1.0\nmass 9 10 20 1 0\n")

	game := NewGame()
	game.demoPickerOpen = true
	if !game.ChooseLoadPickerEntry("lab_scene.xsp") {
		t.Fatalf("saved file did not load: %s", game.LastFileError())
	}

	if _, ok := game.World().MassByID(9); !ok {
		t.Fatalf("saved world was not loaded: %#v", game.World().Masses)
	}
	if game.CurrentFilePath() != filepath.Join("saves", "lab_scene.xsp") {
		t.Fatalf("current file path = %q", game.CurrentFilePath())
	}
	if game.ChooseLoadPickerEntry(loadPickerSeparator) {
		t.Fatal("separator should not load")
	}
	if game.ChooseLoadPickerEntry("missing.xsp") {
		t.Fatal("missing picker entry should not load")
	}
}

func TestNumericSettingTextAndSliderDefaultBuild(t *testing.T) {
	game := NewGame()

	game.appendNumericSettingInput([]rune("7"))
	if game.EnterNumericSettingText("7") {
		t.Fatal("unfocused numeric entry should not be handled")
	}
	if !game.FocusNumericSettingTextField("Mass") {
		t.Fatal("mass text field focus should be handled")
	}
	if game.FocusNumericSettingTextField("Missing") {
		t.Fatal("missing text field focus should not be handled")
	}

	game.appendNumericSettingInput([]rune("x2.5"))
	if got := game.numericInputText; got != "2.5" {
		t.Fatalf("numeric input text = %q, want 2.5", got)
	}
	if got := game.World().Parameters.Value("current mass"); got != "1" {
		t.Fatalf("current mass before commit = %q, want 1", got)
	}
	if !game.CommitNumericSettingText() {
		t.Fatal("numeric input commit should be handled")
	}
	if got := game.World().Parameters.Value("current mass"); got != "2.5" {
		t.Fatalf("current mass after commit = %q, want 2.5", got)
	}
	game.FocusNumericSettingTextField("Mass")
	game.deleteNumericSettingCharacter()
	if got := game.numericInputText; got != "2." {
		t.Fatalf("numeric input after delete = %q, want 2.", got)
	}
	game.focusedNumeric = "Missing"
	game.appendNumericSettingInput([]rune("9"))
	game.deleteNumericSettingCharacter()

	if !game.SetNumericSettingValue("Speed", "99") || game.simulationSpeed != maxSpeed {
		t.Fatalf("speed = %f, want %f", game.simulationSpeed, maxSpeed)
	}
	if game.SetNumericSettingValue("Speed", "not numeric") {
		t.Fatal("invalid numeric value should not be handled")
	}
	if !game.SetNumericSettingValue("Gravity", "12.5") {
		t.Fatal("gravity numeric setting should be handled")
	}
	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "12.5" {
		t.Fatalf("gravity force = %#v", force)
	}
	if game.ChangeNumericSettingWithSlider("Missing", "1") {
		t.Fatal("missing slider setting should not be handled")
	}
	if game.ChangeNumericSettingWithSlider("Mass", "not numeric") {
		t.Fatal("invalid slider value should not be handled")
	}
	if !game.ChangeNumericSettingWithSlider("Mass", "5") {
		t.Fatal("mass slider change should be handled")
	}
}

func TestDemoPickerScrolls(t *testing.T) {
	game := NewGame()
	game.demoFiles = []string{"a.xsp", "b.xsp", "c.xsp", "d.xsp", "e.xsp", "f.xsp", "g.xsp", "h.xsp", "i.xsp", "j.xsp", "k.xsp", "l.xsp", "m.xsp", "n.xsp", "o.xsp", "p.xsp", "q.xsp", "r.xsp", "s.xsp", "t.xsp", "u.xsp", "v.xsp", "w.xsp", "x.xsp", "y.xsp", "z.xsp", "aa.xsp", "ab.xsp", "ac.xsp", "ad.xsp", "ae.xsp", "af.xsp", "ag.xsp", "ah.xsp", "ai.xsp", "aj.xsp", "ak.xsp", "al.xsp", "am.xsp", "an.xsp", "ao.xsp", "ap.xsp"}
	game.demoPickerOpen = true

	game.scrollDemoPicker(3)

	if game.demoPickerScroll != 3 {
		t.Fatalf("demo picker scroll = %d, want 3", game.demoPickerScroll)
	}
}

func TestDemoPickerGeometryAndDrawParameters(t *testing.T) {
	game := NewGame()
	rect := demoPickerRect()
	wantRect := image.Rect(240, 96, screenWidth-240, screenHeight-96)
	if rect != wantRect {
		t.Fatalf("picker rect = %#v, want %#v", rect, wantRect)
	}
	if rows := demoPickerVisibleRows(); rows != (rect.Dy()-48)/demoPickerRowHeight {
		t.Fatalf("visible rows = %d", rows)
	}
	row0 := game.demoRowRect(0)
	row1 := game.demoRowRect(1)
	if row0 != image.Rect(rect.Min.X+12, rect.Min.Y+40, rect.Max.X-12, rect.Min.Y+40+demoPickerRowHeight-2) {
		t.Fatalf("row0 rect = %#v", row0)
	}
	if row1.Min.Y-row0.Min.Y != demoPickerRowHeight || row1.Dx() != row0.Dx() || row1.Dy() != row0.Dy() {
		t.Fatalf("row spacing row0=%#v row1=%#v", row0, row1)
	}
	if got := demoPickerTitlePoint(rect); got != (image.Pt(rect.Min.X+12, rect.Min.Y+10)) {
		t.Fatalf("title point = %#v", got)
	}
	if got := demoPickerRowTextPoint(row0); got != (image.Pt(row0.Min.X+8, row0.Min.Y+4)) {
		t.Fatalf("row text point = %#v", got)
	}
	if demoPickerPanelAntiAlias() || demoPickerRowAntiAlias() {
		t.Fatal("demo picker rectangles should draw without antialiasing")
	}
	if demoPickerRowFill(0) != controlColor || demoPickerRowFill(1) != sectionColor {
		t.Fatal("demo picker row fill should alternate by visible row")
	}
}

func TestDemoPickerClickLoadsSelectedDemo(t *testing.T) {
	game := NewGame()
	game.demoFiles = []string{filepath.Join("..", "..", "demos", "pendulum.xsp")}
	game.demoPickerOpen = true
	oldCount := len(game.World().Masses)
	game.World().Masses = append(game.World().Masses, sim.Mass{ID: 1234, Position: sim.Vec2{X: 1, Y: 1}, Mass: 1})

	row := game.demoRowRect(0)
	game.clickDemoPicker(row.Min.X+2, row.Min.Y+2)

	if game.demoPickerOpen {
		t.Fatal("demo picker stayed open")
	}
	if _, ok := game.World().MassByID(1234); ok {
		t.Fatal("old world data was not cleared")
	}
	if len(game.World().Masses) != oldCount {
		t.Fatal("selected demo was not loaded")
	}
}

func TestDemoPickerClickUsesScrollOffsetAndOutsideClickCloses(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.xsp")
	second := filepath.Join(dir, "second.xsp")
	if err := os.WriteFile(first, []byte("#1.0\nmass 1 10 20 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("#1.0\nmass 2 30 40 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	game := NewGame()
	game.demoFiles = []string{first, second}
	game.demoPickerOpen = true
	game.demoPickerScroll = 1
	row := game.demoRowRect(0)

	game.clickDemoPicker(row.Min.X+2, row.Min.Y+2)

	if _, ok := game.World().MassByID(2); !ok {
		t.Fatalf("scrolled demo was not loaded: %#v", game.World().Masses)
	}

	game.demoPickerOpen = true
	game.clickDemoPicker(demoPickerRect().Max.X+1, demoPickerRect().Max.Y+1)
	if game.demoPickerOpen {
		t.Fatal("outside click should close demo picker")
	}
}

func TestLoadDemoAtOutOfRangeKeepsPickerOpen(t *testing.T) {
	game := NewGame()
	game.demoFiles = []string{filepath.Join("..", "..", "demos", "pendulum.xsp")}
	for _, index := range []int{-1, len(game.demoFiles)} {
		game.demoPickerOpen = true
		game.loadDemoAt(index)
		if !game.demoPickerOpen {
			t.Fatalf("out-of-range index %d closed demo picker", index)
		}
	}
}

func TestClickVisibleGravityControlEnablesGravity(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Forces["gravity"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "0", "direction": "90"}}
	game.dirty = false

	control, ok := visibleControlWithName("gravity force")
	if !ok {
		t.Fatal("missing gravity force checkbox")
	}
	if !game.ClickAt(control.Rect.Min.X+1, control.Rect.Min.Y+1) {
		t.Fatal("Gravity control click was not handled")
	}

	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "10" || force.Values["direction"] != "0" {
		t.Fatalf("gravity force = %#v", force)
	}
	if !game.activeControl("gravity force") {
		t.Fatal("expected Gravity control to show active state")
	}
	if !game.dirty {
		t.Fatal("Gravity control should mark game dirty")
	}
}

func TestClickVisibleMassCollisionControlTogglesCollision(t *testing.T) {
	game := NewGame()

	control, ok := visibleControlWithName("mass collision force")
	if !ok {
		t.Fatal("missing collide toggle")
	}
	if !game.ClickAt(control.Rect.Min.X+1, control.Rect.Min.Y+1) {
		t.Fatal("Collide control click was not handled")
	}

	force, _ := game.World().Parameters.Force("mass collision")
	if force.Enabled != "true" {
		t.Fatalf("mass collision force = %#v", force)
	}
	if !game.activeControl("mass collision force") {
		t.Fatal("expected Collide control to show active state")
	}

	if !game.ClickAt(control.Rect.Min.X+1, control.Rect.Min.Y+1) {
		t.Fatal("second Collide control click was not handled")
	}
	force, _ = game.World().Parameters.Force("mass collision")
	if force.Enabled != "false" {
		t.Fatalf("mass collision force after toggle = %#v", force)
	}
}

func TestToggleForceDisablesForceWithNilValues(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Forces["gravity"] = sim.ForceConfig{Enabled: "true"}

	game.toggleForce("gravity", map[string]string{"magnitude": "10"})

	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "false" || force.Values == nil {
		t.Fatalf("disabled force = %#v, want disabled with initialized values", force)
	}
}

func TestSetForceValueInitializesAndEnablesForce(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Forces["gravity"] = sim.ForceConfig{Enabled: "false"}

	game.setForceValue("gravity", "magnitude", 12)

	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "12" {
		t.Fatalf("force = %#v, want enabled with magnitude 12", force)
	}
}

func TestGravitySliderSetsGravity(t *testing.T) {
	game := NewGame()
	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "0"})
	game.dirty = false
	control, ok := visibleControlWithName("gravity slider")
	if !ok {
		t.Fatal("missing gravity slider")
	}
	track := sliderTrack(control)

	if !game.ClickAt(track.Min.X+track.Dx()/2, track.Min.Y) {
		t.Fatal("gravity slider click was not handled")
	}

	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "25" {
		t.Fatalf("gravity force = %#v", force)
	}
	if !game.dirty {
		t.Fatal("gravity slider should mark game dirty")
	}
}

func TestSpeedSliderSetsSimulationSpeed(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("timestep", "0.016")
	control, ok := visibleControlWithName("speed slider")
	if !ok {
		t.Fatal("missing speed slider")
	}
	track := sliderTrack(control)

	if !game.ClickAt(track.Max.X, track.Min.Y) {
		t.Fatal("speed slider click was not handled")
	}

	if got := game.simulationSpeed; got != maxSpeed {
		t.Fatalf("simulation speed = %f, want %f", got, maxSpeed)
	}
	if got := game.World().Parameters.Value("timestep"); got != "0.016" {
		t.Fatalf("timestep = %q, want unchanged 0.016", got)
	}
}

func TestSpeedSliderMinimumIsZero(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("timestep", "0.016")
	control, ok := visibleControlWithName("speed slider")
	if !ok {
		t.Fatal("missing speed slider")
	}
	track := sliderTrack(control)

	if !game.ClickAt(track.Min.X, track.Min.Y) {
		t.Fatal("speed slider click was not handled")
	}

	if got := game.simulationSpeed; got != 0 {
		t.Fatalf("simulation speed = %f, want 0", got)
	}
	if got := game.World().Parameters.Value("timestep"); got != "0.016" {
		t.Fatalf("timestep = %q, want unchanged 0.016", got)
	}
}

func TestViscositySliderSetsViscosity(t *testing.T) {
	game := NewGame()
	game.dirty = false
	control, ok := visibleControlWithName("viscosity slider")
	if !ok {
		t.Fatal("missing viscosity slider")
	}
	track := sliderTrack(control)

	if !game.ClickAt(track.Min.X+track.Dx()/2, track.Min.Y) {
		t.Fatal("viscosity slider click was not handled")
	}

	if got := game.World().Parameters.Value("viscosity"); got != "1.0" {
		t.Fatalf("viscosity = %q, want 1.0", got)
	}
	if !game.dirty {
		t.Fatal("viscosity slider should mark game dirty")
	}
}

func TestNumericStepButtonsIncrementAndDecrementByTenth(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("damping", "1.0")
	increment, ok := visibleControlWithName("kdamp increment")
	if !ok {
		t.Fatal("missing kdamp increment")
	}
	decrement, ok := visibleControlWithName("kdamp decrement")
	if !ok {
		t.Fatal("missing kdamp decrement")
	}

	if !game.ClickAt(increment.Rect.Min.X+1, increment.Rect.Min.Y+1) {
		t.Fatal("kdamp increment click was not handled")
	}
	if got := game.World().Parameters.Value("damping"); got != "1.1" {
		t.Fatalf("damping after increment = %q, want 1.1", got)
	}
	if !game.ClickAt(decrement.Rect.Min.X+1, decrement.Rect.Min.Y+1) {
		t.Fatal("kdamp decrement click was not handled")
	}
	if got := game.World().Parameters.Value("damping"); got != "1.0" {
		t.Fatalf("damping after decrement = %q, want 1.0", got)
	}
}

func TestNumericStepButtonHoldRepeatsAfterDelay(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("damping", "1.0")
	control, ok := visibleControlWithName("kdamp increment")
	if !ok {
		t.Fatal("missing kdamp increment")
	}
	x := control.Rect.Min.X + 1
	y := control.Rect.Min.Y + 1

	game.handlePointer(true, x, y)
	for range numericStepHoldDelayTicks - 1 {
		game.handlePointer(true, x, y)
	}
	if got := game.World().Parameters.Value("damping"); got != "1.1" {
		t.Fatalf("damping before repeat = %q, want 1.1", got)
	}
	game.handlePointer(true, x, y)
	if got := game.World().Parameters.Value("damping"); got != "1.2" {
		t.Fatalf("damping at first repeat = %q, want 1.2", got)
	}
	for range numericStepRepeatTicks {
		game.handlePointer(true, x, y)
	}
	if got := game.World().Parameters.Value("damping"); got != "1.3" {
		t.Fatalf("damping at second repeat = %q, want 1.3", got)
	}
	game.handlePointer(false, x, y)
}

func TestSliderFractionHandlesTrackBounds(t *testing.T) {
	track := image.Rect(10, 0, 30, 1)
	if got := sliderFractionAt(track, 10); got != 0 {
		t.Fatalf("min fraction = %f, want 0", got)
	}
	if got := sliderFractionAt(track, 30); got != 1 {
		t.Fatalf("max fraction = %f, want 1", got)
	}
	if got := sliderFractionAt(image.Rect(10, 0, 11, 1), 11); got != 1 {
		t.Fatalf("one-pixel fraction = %f, want 1", got)
	}
	if got := sliderFractionAt(image.Rect(10, 0, 10, 1), 20); got != 0 {
		t.Fatalf("zero-width fraction = %f, want 0", got)
	}
}

func TestNumericSettingReportsIncludeCurrentValues(t *testing.T) {
	game := NewGame()
	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "12", "direction": "0"})
	game.World().Parameters.Set("viscosity", "0.5")
	game.World().Parameters.Set("timestep", "0.025")
	game.simulationSpeed = 2

	tests := map[string]string{
		"Gravity":   "12.0",
		"Viscosity": "0.5",
		"Speed":     "2.0",
	}
	for name, expected := range tests {
		frame, ok := game.NumericSettingReport(name)
		if !ok {
			t.Fatalf("missing %s", name)
		}
		if got := frame.Text; got != expected {
			t.Fatalf("%s text = %q, want %q", name, got, expected)
		}
	}
}

func TestNumericTextEditingDoesNotMoveSliderUntilCommit(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("stickiness", "2.0")
	if !game.FocusNumericSettingTextField("Stick") {
		t.Fatal("stick text field focus should be handled")
	}
	if !game.TypeNumericSettingText("7.5") {
		t.Fatal("stick text entry should be handled")
	}

	report, _ := game.NumericSettingReport("Stick")
	if report.Text != "7.5" {
		t.Fatalf("draft text = %q, want 7.5", report.Text)
	}
	if got := game.World().Parameters.Value("stickiness"); got != "2.0" {
		t.Fatalf("stickiness before commit = %q, want 2.0", got)
	}
	if got, _ := game.NumericSettingSliderValue("Stick"); got != "2.0" {
		t.Fatalf("slider value before commit = %q, want 2.0", got)
	}
	if report.SliderFraction != 0.2 {
		t.Fatalf("slider fraction before commit = %f, want 0.2", report.SliderFraction)
	}

	if !game.CommitNumericSettingText() {
		t.Fatal("stick text commit should be handled")
	}
	if got := game.World().Parameters.Value("stickiness"); got != "7.5" {
		t.Fatalf("stickiness after commit = %q, want 7.5", got)
	}
	if game.focusedNumeric != "" {
		t.Fatalf("focused numeric after commit = %q", game.focusedNumeric)
	}
	if got, _ := game.NumericSettingSliderValue("Stick"); got != "7.5" {
		t.Fatalf("slider value after commit = %q, want 7.5", got)
	}
}

func TestNumericTextFieldFocusHighlightsValue(t *testing.T) {
	game := NewGame()
	if !game.FocusNumericSettingTextField("Kdamp") {
		t.Fatal("kdamp text field focus should be handled")
	}
	report, _ := game.NumericSettingReport("Kdamp")
	if !report.TextHighlighted {
		t.Fatal("focused numeric text should be highlighted")
	}
	if !game.TypeNumericSettingText("2.0") {
		t.Fatal("kdamp text entry should be handled")
	}
	report, _ = game.NumericSettingReport("Kdamp")
	if report.TextHighlighted {
		t.Fatal("numeric text should not remain highlighted after editing")
	}
	if !game.CommitNumericSettingText() {
		t.Fatal("kdamp text commit should be handled")
	}
	report, _ = game.NumericSettingReport("Kdamp")
	if report.TextHighlighted || report.TextCursorVisible {
		t.Fatal("numeric highlight and cursor should be cleared after commit")
	}
}

func TestSlidersDragWhileMouseHeld(t *testing.T) {
	game := NewGame()
	control, ok := visibleControlWithName("speed slider")
	if !ok {
		t.Fatal("missing speed slider")
	}
	track := sliderTrack(control)

	game.handlePointer(true, track.Min.X, track.Min.Y)
	game.handlePointer(true, track.Max.X, track.Min.Y)
	game.handlePointer(false, track.Max.X, track.Min.Y)

	if got := game.simulationSpeed; got != maxSpeed {
		t.Fatalf("simulation speed after drag = %f, want %f", got, maxSpeed)
	}
}

func TestReleasePointerClearsTransientDragState(t *testing.T) {
	game := NewGame()
	game.pointer.draggingMassID = 7
	game.pointer.draggingOffsets = map[int]sim.Vec2{7: {X: 1}}
	game.pointer.dragMoved = true
	game.pointer.selectionDrag = true
	game.activeSlider = "speed slider"
	game.activeNumericStep = "speed increment"
	game.numericStepTicks = 12

	game.releasePointer(sim.Vec2{X: 10, Y: 10})

	if game.pointer.draggingMassID != 0 || game.pointer.draggingOffsets != nil || game.pointer.dragMoved || game.pointer.selectionDrag || game.activeSlider != "" || game.activeNumericStep != "" || game.numericStepTicks != 0 {
		t.Fatalf("drag state was not cleared: %#v", game)
	}
}

func TestRightPointerOnlyOpensContextOnInitialPress(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1})

	game.handleRightPointer(true, 20, 20)
	firstMass := game.massMenu.MassID
	game.massMenu = massContextMenu{}
	game.handleRightPointer(true, 20, 20)

	if firstMass != 1 {
		t.Fatalf("first right press mass = %d, want 1", firstMass)
	}
	if game.massMenu.Open {
		t.Fatal("held right press reopened context menu")
	}
}

func TestClickOpenOverlayConsumesOpenOverlayClicks(t *testing.T) {
	game := NewGame()
	game.demoFiles = []string{"demos/pendulum.xsp"}
	game.demoPickerOpen = true
	if !game.clickOpenOverlay(0, 0) {
		t.Fatal("demo picker overlay click was not consumed")
	}

	game.demoPickerOpen = false
	game.valueDialog = valueDialog{Open: true}
	if !game.clickOpenOverlay(0, 0) {
		t.Fatal("value dialog overlay click was not consumed")
	}

	game.valueDialog.Open = false
	game.massMenu = massContextMenu{Open: true, MassID: 1}
	if !game.clickOpenOverlay(0, 0) {
		t.Fatal("mass context menu overlay click was not consumed")
	}

	game.massMenu.Open = false
	game.springMenu = springContextMenu{Open: true, SpringID: 1}
	if !game.clickOpenOverlay(0, 0) {
		t.Fatal("spring context menu overlay click was not consumed")
	}

	game.springMenu.Open = false
	game.saveDialog = saveFilenameDialog{Open: true}
	if !game.clickOpenOverlay(0, 0) {
		t.Fatal("save dialog overlay click was not consumed")
	}

	game.saveDialog.Open = false
	if game.clickOpenOverlay(0, 0) {
		t.Fatal("missing overlay click should not be consumed")
	}
}

func TestDemoPickerScrollClampsBothDirections(t *testing.T) {
	game := NewGame()
	game.demoPickerOpen = true
	game.demoFiles = make([]string, demoPickerVisibleRows()+2)

	game.scrollDemoPicker(99)
	if game.demoPickerScroll != 2 {
		t.Fatalf("scroll high = %d, want 2", game.demoPickerScroll)
	}
	game.scrollDemoPicker(-99)
	if game.demoPickerScroll != 0 {
		t.Fatalf("scroll low = %d, want 0", game.demoPickerScroll)
	}
}

func TestUpdateAdvancesBySpeedWithoutChangingTimestep(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Set("timestep", "0.001")
	game.simulationSpeed = 2

	if err := game.Update(); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if got := game.World().Parameters.Value("timestep"); got != "0.001" {
		t.Fatalf("timestep = %q, want unchanged 0.001", got)
	}
	if got := game.World().Time; math.Abs(got-baseFrameTime*2) > 1e-12 {
		t.Fatalf("world time = %f, want %f", got, baseFrameTime*2)
	}
	if game.World().LastAdvanceSteps <= 1 {
		t.Fatalf("last advance steps = %d, want subdivided integration", game.World().LastAdvanceSteps)
	}
}

func TestZeroSpeedPausesSimulationAdvance(t *testing.T) {
	game := NewGame()
	game.simulationSpeed = 0

	if err := game.Update(); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if got := game.World().Time; got != 0 {
		t.Fatalf("world time = %f, want 0", got)
	}
}

func TestQuitControlTerminatesUpdateLoop(t *testing.T) {
	game := NewGame()

	if !game.ClickVisibleControl("Quit") {
		t.Fatal("Quit control click was not handled")
	}

	if err := game.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("Update after Quit = %v, want ebiten.Termination", err)
	}
}

func TestWindowCloseTerminatesUpdateLoop(t *testing.T) {
	game := NewGame()

	game.handleWindowClose(true)

	if err := game.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("Update after window close = %v, want ebiten.Termination", err)
	}
}

func TestInspectorTogglesMapToSimulationParameters(t *testing.T) {
	tests := []struct {
		label  string
		assert func(*testing.T, *Game)
	}{
		{"Springs", func(t *testing.T, game *Game) {
			if game.World().Parameters.Value("show springs") != "false" {
				t.Fatalf("show springs = %q", game.World().Parameters.Value("show springs"))
			}
		}},
		{"Grid", func(t *testing.T, game *Game) {
			if game.World().Parameters.Value("grid snap") != "0" {
				t.Fatalf("grid snap = %q", game.World().Parameters.Value("grid snap"))
			}
		}},
		{"T", func(t *testing.T, game *Game) {
			if enabled, _ := game.World().Parameters.WallEnabled("top"); !enabled {
				t.Fatal("top wall was not enabled")
			}
		}},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			game := NewGame()
			if !game.ClickVisibleControl(test.label) {
				t.Fatalf("click %q was not handled", test.label)
			}
			test.assert(t, game)
		})
	}

	t.Run("Adapt", func(t *testing.T) {
		game := NewGame()
		control, ok := visibleControlWithName("adaptive timestep toggle")
		if !ok {
			t.Fatal("missing adaptive timestep checkbox")
		}
		if !game.ClickAt(control.Rect.Min.X+1, control.Rect.Min.Y+1) {
			t.Fatal("click Adapt checkbox was not handled")
		}
		if game.World().Parameters.Value("adaptive timestep") != "true" {
			t.Fatalf("adaptive timestep = %q", game.World().Parameters.Value("adaptive timestep"))
		}
	})
}

func TestWorldPointerCreatesMassAndSpring(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	game.ReplaceWorld(world)

	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 100, 100)
	game.handlePointer(true, 140, 100)
	game.handlePointer(false, 140, 100)

	if len(game.World().Masses) != 2 {
		t.Fatalf("masses = %#v", game.World().Masses)
	}

	game.controlDown = true
	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 140, 100)

	if len(game.World().Springs) != 1 {
		t.Fatalf("springs = %#v", game.World().Springs)
	}
}

func TestControlDragRubberBandsPendingSpringGesture(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 200, Y: 100}, Mass: 1})
	game.ReplaceWorld(world)
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(true, 150, 140)

	line, ok := game.pendingSpringLine()
	if !ok {
		t.Fatal("expected pending spring line")
	}
	if line != (selectionLine{x1: 100, y1: 100, x2: 150, y2: 140}) {
		t.Fatalf("pending spring line = %#v", line)
	}

	game.handlePointer(false, 200, 100)
	if _, ok := game.pendingSpringLine(); ok {
		t.Fatal("pending spring line remained after release")
	}
}

func TestControlDragCreatesSpringFromAnyMode(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 540, Y: 340}, Mass: 1})
	game.ReplaceWorld(world)
	game.controlDown = true

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.handlePointer(false, 540, 340)

	if len(game.World().Springs) != 1 {
		t.Fatalf("springs = %#v, want one", game.World().Springs)
	}
	spring := game.World().Springs[0]
	if spring.MassA != 1 || spring.MassB != 2 {
		t.Fatalf("spring endpoints = %d,%d, want 1,2", spring.MassA, spring.MassB)
	}
}

func TestControlDragRubberBandsPendingSpring(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 540, Y: 340}, Mass: 1})
	game.ReplaceWorld(world)
	game.controlDown = true

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 520, 330)

	line, ok := game.pendingSpringLine()
	if !ok {
		t.Fatal("expected pending spring line")
	}
	if line != (selectionLine{x1: 500, y1: 300, x2: 520, y2: 330}) {
		t.Fatalf("pending spring line = %#v", line)
	}
}

func TestControlClickEmptyCanvasStartsMassSpringChain(t *testing.T) {
	game := NewGame()
	game.ReplaceWorld(sim.NewWorld())
	game.World().Parameters.Set("grid snap", "0")
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 100, 100)

	if len(game.World().Masses) != 1 {
		t.Fatalf("masses = %#v, want one start mass", game.World().Masses)
	}
	if game.pointer.pendingSpringID != 1 || !game.pointer.springChainActive {
		t.Fatalf("chain state pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
	if len(game.World().Springs) != 0 {
		t.Fatalf("springs = %#v, want none until next click", game.World().Springs)
	}
}

func TestSpringChainPlacesEndpointMassAndStopsWithoutControl(t *testing.T) {
	game := NewGame()
	game.ReplaceWorld(sim.NewWorld())
	game.World().Parameters.Set("grid snap", "0")
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 100, 100)
	game.controlDown = false
	game.handlePointer(true, 140, 100)
	game.handlePointer(false, 140, 100)

	if len(game.World().Masses) != 2 || len(game.World().Springs) != 1 {
		t.Fatalf("chain objects masses=%#v springs=%#v", game.World().Masses, game.World().Springs)
	}
	spring := game.World().Springs[0]
	if spring.MassA != 1 || spring.MassB != 2 {
		t.Fatalf("spring endpoints = %d,%d, want 1,2", spring.MassA, spring.MassB)
	}
	if game.pointer.pendingSpringID != 0 || game.pointer.springChainActive {
		t.Fatalf("chain remained pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
}

func TestSpringChainContinuesWhileControlIsDown(t *testing.T) {
	game := NewGame()
	game.ReplaceWorld(sim.NewWorld())
	game.World().Parameters.Set("grid snap", "0")
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 100, 100)
	game.handlePointer(true, 140, 100)
	game.handlePointer(false, 140, 100)
	game.handlePointer(true, 180, 100)
	game.handlePointer(false, 180, 100)

	if len(game.World().Masses) != 3 || len(game.World().Springs) != 2 {
		t.Fatalf("chain objects masses=%#v springs=%#v", game.World().Masses, game.World().Springs)
	}
	if game.pointer.pendingSpringID != 3 || !game.pointer.springChainActive {
		t.Fatalf("chain state pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
}

func TestSpringChainTerminatesOnExistingMassEvenWithControlDown(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 7, Position: sim.Vec2{X: 180, Y: 100}, Mass: 1})
	game.ReplaceWorld(world)
	game.World().Parameters.Set("grid snap", "0")
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 100, 100)
	game.handlePointer(true, 180, 100)
	game.handlePointer(false, 180, 100)

	if len(game.World().Masses) != 2 || len(game.World().Springs) != 1 {
		t.Fatalf("chain objects masses=%#v springs=%#v", game.World().Masses, game.World().Springs)
	}
	spring := game.World().Springs[0]
	if spring.MassA != 8 || spring.MassB != 7 {
		t.Fatalf("spring endpoints = %d,%d, want 8,7", spring.MassA, spring.MassB)
	}
	if game.pointer.pendingSpringID != 0 || game.pointer.springChainActive {
		t.Fatalf("chain remained pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
}

func TestSelectionClickThreshold(t *testing.T) {
	if !selectionClick(sim.Vec2{X: 0, Y: 0}, sim.Vec2{X: 2.99, Y: 0}) {
		t.Fatal("short movement should count as click")
	}
	if selectionClick(sim.Vec2{X: 0, Y: 0}, sim.Vec2{X: 3, Y: 0}) {
		t.Fatal("threshold movement should count as drag")
	}
}

func TestThrowKeyWithoutDragDoesNotThrowMass(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Velocity: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	game.pointer.draggingMassID = 1
	game.pointer.draggingStart = sim.Vec2{X: 10, Y: 10}
	game.throwDown = true

	game.finishMassDrag(sim.Vec2{X: 10, Y: 10})

	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 3, Y: 4}) {
		t.Fatalf("velocity = %#v, want unchanged", mass.Velocity)
	}
}

func TestThrowDraggedMassesUsesSingleMassWhenNoSelectionOffsets(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Mass: 1}, sim.Mass{ID: 2, Mass: 1})
	game.pointer.draggingMassID = 1

	game.throwDraggedMasses(sim.Vec2{X: 7, Y: 8})

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	if mass1.Velocity != (sim.Vec2{X: 7, Y: 8}) || mass2.Velocity != (sim.Vec2{}) {
		t.Fatalf("velocities = %#v %#v", mass1.Velocity, mass2.Velocity)
	}
}

func TestThrowDraggedMassesMarksDirtyForSelectionOffsets(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Mass: 1})
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {}}

	game.throwDraggedMasses(sim.Vec2{X: 7, Y: 8})

	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 7, Y: 8}) || !game.dirty {
		t.Fatalf("throw selected state velocity=%#v dirty=%t", mass.Velocity, game.dirty)
	}
}

func TestFinishSelectGestureClickClearsDragAndCreatesMass(t *testing.T) {
	game := NewGame()
	game.ReplaceWorld(sim.NewWorld())
	game.pointer.selectionDrag = true
	game.pointer.selectionStart = sim.Vec2{X: 100, Y: 100}

	game.finishSelectGesture(sim.Vec2{X: 100, Y: 100})

	if game.pointer.selectionDrag {
		t.Fatal("selection drag remained active")
	}
	if len(game.World().Masses) != 1 {
		t.Fatalf("mass count = %d, want 1", len(game.World().Masses))
	}
}

func TestCreateMassAtMarksDirtyAndSelectsCreatedMass(t *testing.T) {
	game := NewGame()
	game.ReplaceWorld(sim.NewWorld())

	id, ok := game.createMassAt(sim.Vec2{X: 130, Y: 140}, false)

	if !ok || id == 0 {
		t.Fatalf("createMassAt = %d, %t", id, ok)
	}
	if !game.dirty || !game.editing().MassSelected(id) || !game.selected {
		t.Fatalf("created mass state dirty=%t selected=%t editor=%#v", game.dirty, game.selected, game.editing().SelectedMasses)
	}
}

func TestFinishSpringAtMarksDirtyOnlyForValidDifferentEndpoint(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 110, Y: 110}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 130, Y: 110}, Mass: 1},
	)
	game.pointer.pendingSpringID = 1

	game.finishSpringAt(sim.Vec2{X: 130, Y: 110})

	if len(game.World().Springs) != 1 || !game.dirty || game.pointer.pendingSpringID != 0 {
		t.Fatalf("finish spring state springs=%#v dirty=%t pending=%d", game.World().Springs, game.dirty, game.pointer.pendingSpringID)
	}

	game.dirty = false
	game.pointer.pendingSpringID = 1
	game.finishSpringAt(sim.Vec2{X: 110, Y: 110})
	if len(game.World().Springs) != 1 || game.dirty {
		t.Fatalf("self spring changed state springs=%#v dirty=%t", game.World().Springs, game.dirty)
	}
}

func TestSpringPlacementOutsideCanvasDoesNotCreateSpring(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 110, Y: 110}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 130, Y: 110}, Mass: 1},
	)

	game.beginSpringAt(sim.Vec2{X: -1, Y: 110})
	if game.pointer.pendingSpringID != 0 {
		t.Fatalf("pending spring id = %d, want none", game.pointer.pendingSpringID)
	}

	game.pointer.pendingSpringID = 1
	game.finishSpringAt(sim.Vec2{X: -1, Y: 110})
	if len(game.World().Springs) != 0 || game.dirty {
		t.Fatalf("spring placement outside canvas changed state springs=%#v dirty=%t", game.World().Springs, game.dirty)
	}
}

func TestMassAtIncludesRadiusBoundaryAndReportsMiss(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_, _, radius := massDrawCircle(game.World().Masses[0])

	if id, ok := game.massAt(sim.Vec2{X: 10 + float64(radius), Y: 10}); !ok || id != 1 {
		t.Fatalf("boundary massAt = %d, %t", id, ok)
	}
	if id, ok := game.massAt(sim.Vec2{X: 10 + float64(radius) + 0.01, Y: 10}); ok || id != 0 {
		t.Fatalf("miss massAt = %d, %t", id, ok)
	}
}

func TestClickVisibleControlsUseRectHitTesting(t *testing.T) {
	game := NewGame()

	if !game.ClickAt(12, 52) {
		t.Fatal("expected click inside All control to be handled")
	}
	if game.ClickAt(500, 300) {
		t.Fatal("unexpected handled click outside controls")
	}
	if control, ok := visibleControlAt(image.Pt(500, 300)); ok || control != (controlBox{}) {
		t.Fatalf("visibleControlAt outside controls = %#v, %t", control, ok)
	}
	if game.VisibleControlActive("missing") {
		t.Fatal("missing visible control should not be active")
	}
}

func TestDragMassWorksWithoutMode(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)

	if !game.DragMass(1, sim.Vec2{X: 40, Y: 50}) {
		t.Fatal("drag was not handled")
	}
}

func TestDragMassSnapsToGridPoint(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	world.Parameters.Set("grid snap", "10")
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.ReplaceWorld(world)

	if !game.DragMass(1, sim.Vec2{X: 123, Y: 87}) {
		t.Fatal("drag was not handled")
	}

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 120, Y: 90}) {
		t.Fatalf("mass position = %#v, want snapped grid point 120,90", mass.Position)
	}
}

func TestDragMassClampsToCanvasBounds(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	world.Parameters.Set("grid snap", "0")
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)

	if !game.DragMass(1, sim.Vec2{X: 2000, Y: 2000}) {
		t.Fatal("drag was not handled")
	}

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 1340, Y: 1000}) {
		t.Fatalf("mass position = %#v, want clamped canvas corner", mass.Position)
	}
}

func TestPointerGestureDragsMass(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.handlePointer(false, 540, 340)
	game.handlePointer(true, 700, 500)

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 540, Y: 340}) {
		t.Fatalf("mass position = %#v, want 540,340", mass.Position)
	}
}

func TestClickOnEmptyCanvasCreatesMassAndReplacesSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 600, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()

	game.handlePointer(true, 500, 300)
	game.handlePointer(false, 500, 300)

	if game.editing().MassSelected(1) {
		t.Fatalf("selection was not cleared: %#v", game.editing().SelectedMasses)
	}
	if len(game.World().Masses) != 2 {
		t.Fatalf("empty click should create a mass, count = %d", len(game.World().Masses))
	}
	if !game.editing().MassSelected(2) || !game.selected {
		t.Fatalf("new mass was not selected: %#v", game.editing().SelectedMasses)
	}
}

func TestPointerGestureOnMassDragsMass(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1, Fixed: true})
	game.ReplaceWorld(world)

	game.handlePointer(true, 500, 300)
	if game.pointer.draggingMassID != 1 {
		t.Fatalf("dragging mass id = %d, want 1", game.pointer.draggingMassID)
	}
	game.handlePointer(true, 540, 340)
	game.handlePointer(false, 540, 340)

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 540, Y: 340}) {
		t.Fatalf("mass position = %#v, want 540,340", mass.Position)
	}
	if !mass.Fixed {
		t.Fatal("dragging should preserve fixed state")
	}
}

func TestDraggingSelectedMassMovesEntireSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 600, Y: 350}, Mass: 1, Fixed: true})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 700, Y: 500}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.syncSelectionState()

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.handlePointer(false, 540, 340)

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	mass3, _ := game.World().MassByID(3)
	if mass1.Position != (sim.Vec2{X: 540, Y: 340}) {
		t.Fatalf("dragged mass position = %#v, want 540,340", mass1.Position)
	}
	if mass2.Position != (sim.Vec2{X: 640, Y: 390}) {
		t.Fatalf("selected fixed mass position = %#v, want 640,390", mass2.Position)
	}
	if mass3.Position != (sim.Vec2{X: 700, Y: 500}) {
		t.Fatalf("unselected mass position = %#v, want unchanged", mass3.Position)
	}
}

func TestDragSelectedMassesWithoutOffsetsMovesByDelta(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 110, Y: 110}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 120, Y: 110}, Mass: 1},
	)
	game.World().Parameters.Set("grid snap", "0")
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.pointer.draggingLast = sim.Vec2{X: 110, Y: 110}
	game.pointer.draggingStart = sim.Vec2{X: 110, Y: 110}

	if !game.dragSelectedMasses(sim.Vec2{X: 113, Y: 114}) {
		t.Fatal("selected drag should report handled")
	}

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	if mass1.Position != (sim.Vec2{X: 113, Y: 114}) || mass2.Position != (sim.Vec2{X: 123, Y: 114}) {
		t.Fatalf("selected drag positions = %#v %#v", mass1.Position, mass2.Position)
	}
	if !game.pointer.dragMoved || !game.dirty {
		t.Fatalf("selected drag state moved=%t dirty=%t", game.pointer.dragMoved, game.dirty)
	}
}

func TestDragSelectedMassesWithSingleOffsetAppliesOffset(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 110, Y: 110}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 120, Y: 110}, Mass: 1},
	)
	game.World().Parameters.Set("grid snap", "0")
	_ = game.editing().SelectMass(1)
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {X: 2, Y: 3}}
	game.pointer.draggingStart = sim.Vec2{X: 110, Y: 110}

	if !game.dragSelectedMasses(sim.Vec2{X: 113, Y: 114}) {
		t.Fatal("selected drag with offsets should report handled")
	}

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	if mass1.Position != (sim.Vec2{X: 115, Y: 117}) || mass2.Position != (sim.Vec2{X: 120, Y: 110}) {
		t.Fatalf("offset drag positions = %#v %#v", mass1.Position, mass2.Position)
	}
}

func TestDragSelectedMassesWithGridSnapAppliesOffsetToGridPoints(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 110, Y: 110}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 120, Y: 110}, Mass: 1},
	)
	game.World().Parameters.Set("grid snap", "10")
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {X: 2, Y: 3}, 2: {X: -3, Y: -2}}
	game.pointer.draggingStart = sim.Vec2{X: 110, Y: 110}

	if !game.dragSelectedMasses(sim.Vec2{X: 113, Y: 114}) {
		t.Fatal("selected drag with offsets should report handled")
	}

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	if mass1.Position != (sim.Vec2{X: 120, Y: 120}) || mass2.Position != (sim.Vec2{X: 110, Y: 110}) {
		t.Fatalf("snapped offset drag positions = %#v %#v", mass1.Position, mass2.Position)
	}
}

func TestDraggingMassWithThrowKeySetsVelocityFromDragVector(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.throwDown = true
	game.handlePointer(false, 540, 340)

	mass, _ := game.World().MassByID(1)
	if mass.Velocity != (sim.Vec2{X: 40, Y: 40}) {
		t.Fatalf("thrown mass velocity = %#v, want 40,40", mass.Velocity)
	}
}

func TestDraggingSelectedMassesWithThrowKeySetsSelectionVelocity(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 600, Y: 350}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 700, Y: 500}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.syncSelectionState()

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.throwDown = true
	game.handlePointer(false, 540, 340)

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	mass3, _ := game.World().MassByID(3)
	if mass1.Velocity != (sim.Vec2{X: 40, Y: 40}) || mass2.Velocity != (sim.Vec2{X: 40, Y: 40}) {
		t.Fatalf("selected throw velocities = %#v %#v, want 40,40", mass1.Velocity, mass2.Velocity)
	}
	if mass3.Velocity != (sim.Vec2{}) {
		t.Fatalf("unselected mass velocity = %#v, want zero", mass3.Velocity)
	}
}

func TestDraggingMassPinsItToCursorWhileSimulationRuns(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1000", "direction": "90"})
	game.ReplaceWorld(world)
	game.paused = false
	game.simulationSpeed = 1

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.pointer.lastCursor = sim.Vec2{X: 540, Y: 340}
	game.advanceSimulationFrame()

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 540, Y: 340}) {
		t.Fatalf("dragged mass position after advance = %#v, want cursor", mass.Position)
	}
	if mass.Velocity != (sim.Vec2{}) {
		t.Fatalf("dragged mass velocity after advance = %#v, want zero", mass.Velocity)
	}
}

func TestDraggingMassWhileSimulationRunsDoesNotChangeAttachedSpringRestLength(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 600, Y: 300}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 100, SpringConstant: 12})
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1000", "direction": "90"})
	game.ReplaceWorld(world)
	game.paused = false
	game.simulationSpeed = 1

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.pointer.lastCursor = sim.Vec2{X: 540, Y: 340}
	game.advanceSimulationFrame()
	game.handlePointer(false, 540, 340)

	spring, _ := game.World().SpringByID(1)
	if spring.RestLength != 100 {
		t.Fatalf("spring rest length after drag = %f, want 100", spring.RestLength)
	}
}

func TestDraggingSelectedMassesPinsGroupToCursorWhileSimulationRuns(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 600, Y: 350}, Mass: 1})
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1000", "direction": "90"})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.syncSelectionState()

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)
	game.pointer.lastCursor = sim.Vec2{X: 540, Y: 340}
	game.advanceSimulationFrame()

	mass1, _ := game.World().MassByID(1)
	mass2, _ := game.World().MassByID(2)
	if mass1.Position != (sim.Vec2{X: 540, Y: 340}) {
		t.Fatalf("dragged mass position after advance = %#v, want cursor", mass1.Position)
	}
	if mass2.Position != (sim.Vec2{X: 640, Y: 390}) {
		t.Fatalf("selected mass position after advance = %#v, want pinned offset", mass2.Position)
	}
	if mass1.Velocity != (sim.Vec2{}) || mass2.Velocity != (sim.Vec2{}) {
		t.Fatalf("dragged velocities after advance = %#v %#v, want zero", mass1.Velocity, mass2.Velocity)
	}
}

func TestCaptureSelectedDraggingOffsetsRequiresSelectedDraggedMass(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1},
	)
	game.pointer.draggingMassID = 1
	game.pointer.draggingOffsets = map[int]sim.Vec2{}
	_ = game.editing().SelectMass(2)

	if game.captureSelectedDraggingOffsets(sim.Vec2{}) {
		t.Fatal("unselected dragged mass should not capture selection offsets")
	}

	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	if !game.captureSelectedDraggingOffsets(sim.Vec2{X: 5, Y: 5}) {
		t.Fatal("selected dragged mass should capture selection offsets")
	}
	if len(game.pointer.draggingOffsets) != 2 || game.pointer.draggingOffsets[1] != (sim.Vec2{X: 5, Y: 5}) || game.pointer.draggingOffsets[2] != (sim.Vec2{X: 15, Y: 15}) {
		t.Fatalf("dragging offsets = %#v", game.pointer.draggingOffsets)
	}
}

func TestCaptureSelectedDraggingOffsetsWorksForSingleSelectedMass(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.pointer.draggingMassID = 1
	game.pointer.draggingOffsets = map[int]sim.Vec2{}
	_ = game.editing().SelectMass(1)

	if !game.captureSelectedDraggingOffsets(sim.Vec2{X: 4, Y: 5}) {
		t.Fatal("single selected dragged mass should capture selection offset")
	}
	if game.pointer.draggingOffsets[1] != (sim.Vec2{X: 6, Y: 5}) {
		t.Fatalf("dragging offsets = %#v", game.pointer.draggingOffsets)
	}
}

func TestPinDraggingMassesRequiresDraggingIDAndOffsets(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {X: 1, Y: 1}}
	game.pinDraggingMasses(sim.Vec2{X: 20, Y: 20})
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 10, Y: 10}) {
		t.Fatalf("mass moved without dragging id: %#v", mass.Position)
	}

	game.pointer.draggingMassID = 1
	game.pointer.draggingOffsets = nil
	game.pinDraggingMasses(sim.Vec2{X: 20, Y: 20})
	mass, _ = game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 10, Y: 10}) {
		t.Fatalf("mass moved without offsets: %#v", mass.Position)
	}
}

func TestClickingSelectedMassWithoutDraggingReplacesSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 600, Y: 350}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true
	game.syncSelectionState()

	game.handlePointer(true, 500, 300)
	game.handlePointer(false, 500, 300)

	if !game.editing().MassSelected(1) || game.editing().MassSelected(2) {
		t.Fatalf("selection = %#v, want only mass 1", game.editing().SelectedMasses)
	}
}

func TestDragOnEmptyCanvasBoxSelectsMasses(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 700, Y: 500}, Mass: 1})
	game.ReplaceWorld(world)

	game.handlePointer(true, 450, 250)
	game.handlePointer(true, 600, 400)
	if !game.pointer.selectionDrag {
		t.Fatal("selection rectangle was not active during drag")
	}
	game.handlePointer(false, 600, 400)

	if !game.editing().MassSelected(1) {
		t.Fatal("mass inside selection rectangle was not selected")
	}
	if game.editing().MassSelected(2) {
		t.Fatal("mass outside selection rectangle was selected")
	}
	if len(game.World().Masses) != 2 {
		t.Fatal("box selection should not create a mass")
	}
}

func TestShiftClickMassAddsToSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 700, Y: 500}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()
	game.shiftDown = true

	game.handlePointer(true, 700, 500)
	game.handlePointer(false, 700, 500)

	if !game.editing().MassSelected(1) || !game.editing().MassSelected(2) {
		t.Fatalf("shift click selection = %#v", game.editing().SelectedMasses)
	}
}

func TestShiftBoxSelectAddsToSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 700, Y: 500}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(2)
	game.syncSelectionState()
	game.shiftDown = true

	game.handlePointer(true, 450, 250)
	game.handlePointer(true, 600, 400)
	game.handlePointer(false, 600, 400)

	if !game.editing().MassSelected(1) || !game.editing().MassSelected(2) {
		t.Fatalf("shift box selection = %#v", game.editing().SelectedMasses)
	}
}

func TestShiftClickEmptyCanvasAddsCreatedMassToSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 600, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()
	game.shiftDown = true

	game.handlePointer(true, 500, 300)
	game.handlePointer(false, 500, 300)

	if !game.editing().MassSelected(1) || !game.editing().MassSelected(2) || !game.selected {
		t.Fatalf("shift empty click selection = %#v", game.editing().SelectedMasses)
	}
}

func TestEscapeClearsSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()

	if !game.HandleShortcut("Esc") {
		t.Fatal("Esc shortcut was not handled")
	}

	if game.editing().MassSelected(1) || game.selected {
		t.Fatalf("selection after Esc = %#v", game.editing().SelectedMasses)
	}
}

func TestEscapeCancelsRubberbandingSpring(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 100, Y: 100}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 200, Y: 100}, Mass: 1})
	game.ReplaceWorld(world)
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(true, 150, 140)

	if _, ok := game.pendingSpringLine(); !ok {
		t.Fatal("expected pending spring before Esc")
	}
	if !game.HandleShortcut("Esc") {
		t.Fatal("Esc shortcut was not handled")
	}
	if _, ok := game.pendingSpringLine(); ok {
		t.Fatal("pending spring line remained after Esc")
	}
	if len(game.World().Springs) != 0 {
		t.Fatalf("springs = %#v, want none", game.World().Springs)
	}
}

func TestEscapeCancelsSpringChainWithoutClearingSelection(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 300, Y: 100}, Mass: 1})
	game.ReplaceWorld(world)
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()
	game.controlDown = true

	game.handlePointer(true, 100, 100)
	game.handlePointer(false, 100, 100)

	if !game.pointer.springChainActive || game.pointer.pendingSpringID == 0 {
		t.Fatalf("expected active spring chain pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
	if !game.HandleShortcut("Esc") {
		t.Fatal("Esc shortcut was not handled")
	}
	if game.pointer.springChainActive || game.pointer.pendingSpringID != 0 {
		t.Fatalf("spring chain remained pending=%d active=%t", game.pointer.pendingSpringID, game.pointer.springChainActive)
	}
	if !game.editing().MassSelected(2) || !game.selected {
		t.Fatalf("selection was cleared while cancelling spring chain: %#v", game.editing().SelectedMasses)
	}
}

func TestPointerGestureRequiresMouseDownOnMass(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.ReplaceWorld(world)

	game.handlePointer(true, 80, 90)
	game.handlePointer(true, 40, 50)

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 10, Y: 10}) {
		t.Fatalf("mass position = %#v, want 10,10", mass.Position)
	}
}

func TestSelectedMassesImplicitSelectionRequiresNoSpringSelection(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)
	game.World().Springs = []sim.Spring{{ID: 1, MassA: 1, MassB: 2}}
	game.selected = true

	if got := game.selectedMasses(); len(got) != 2 {
		t.Fatalf("implicit selected masses = %#v, want all masses", got)
	}

	game.editing().SelectedSprings[1] = true
	if got := game.selectedMasses(); len(got) != 0 {
		t.Fatalf("spring-only selection masses = %#v, want none", got)
	}
}

func TestSelectedSpringLinesRequireBothEndpoints(t *testing.T) {
	game := gameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.World().Springs = []sim.Spring{
		{ID: 1, MassA: 1, MassB: 2},
		{ID: 2, MassA: 2, MassB: 1},
	}
	game.editing().SelectedSprings[1] = true
	game.editing().SelectedSprings[2] = true

	if got := game.selectedSpringLines(); len(got) != 0 {
		t.Fatalf("selected spring lines with missing endpoints = %#v", got)
	}
}

func TestPointerGestureDragsFixedMass(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 500, Y: 300}, Mass: 1, Fixed: true})
	game.ReplaceWorld(world)

	game.handlePointer(true, 500, 300)
	game.handlePointer(true, 540, 340)

	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 540, Y: 340}) {
		t.Fatalf("mass position = %#v, want 540,340", mass.Position)
	}
	if !mass.Fixed {
		t.Fatal("dragging should not make fixed mass free")
	}
}

func TestRightClickOnMassOpensContextMenu(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 7, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.ReplaceWorld(world)

	game.handleRightPointer(true, 10, 10)

	if !game.massMenu.Open {
		t.Fatal("mass context menu was not opened")
	}
	if game.massMenu.MassID != 7 {
		t.Fatalf("menu mass id = %d, want 7", game.massMenu.MassID)
	}
	if !game.editing().MassSelected(7) {
		t.Fatal("right-clicked mass was not selected")
	}
}

func TestMassContextMenuTogglesFixedState(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.ReplaceWorld(world)

	game.handleRightPointer(true, 10, 10)
	game.handleRightPointer(false, 10, 10)
	row := game.massContextMenuRowRect(0)
	game.handlePointer(true, row.Min.X+2, row.Min.Y+2)
	game.handlePointer(false, row.Min.X+2, row.Min.Y+2)

	mass, _ := game.World().MassByID(1)
	if !mass.Fixed {
		t.Fatal("mass was not fixed")
	}
	if game.massMenu.Open {
		t.Fatal("mass context menu stayed open")
	}
}

func TestMassContextMenuOpensSetMassDialog(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 5})
	game.ReplaceWorld(world)

	game.handleRightPointer(true, 10, 10)
	game.handleRightPointer(false, 10, 10)
	row := game.massContextMenuRowRect(1)
	game.handlePointer(true, row.Min.X+2, row.Min.Y+2)
	game.handlePointer(false, row.Min.X+2, row.Min.Y+2)

	if !game.valueDialog.Open {
		t.Fatal("value dialog was not opened")
	}
	if game.valueDialog.Title != "Set Mass #1" || game.valueDialog.Text != "5" {
		t.Fatalf("value dialog = %#v", game.valueDialog)
	}
}

func TestMassContextMenuSetsForceCenter(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 5})
	game.ReplaceWorld(world)

	game.handleRightPointer(true, 10, 10)
	game.handleRightPointer(false, 10, 10)
	row := game.massContextMenuRowRect(2)
	game.handlePointer(true, row.Min.X+2, row.Min.Y+2)
	game.handlePointer(false, row.Min.X+2, row.Min.Y+2)

	if game.World().CenterMassID() != 1 {
		t.Fatalf("center mass id = %d, want 1", game.World().CenterMassID())
	}
	if game.massMenu.Open {
		t.Fatal("mass context menu stayed open")
	}
}

func TestSetMassDialogStaysOpenWhileOpeningClickIsHeld(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 5})
	game.ReplaceWorld(world)

	game.handleRightPointer(true, 10, 10)
	game.handleRightPointer(false, 10, 10)
	row := game.massContextMenuRowRect(1)
	game.handlePointer(true, row.Min.X+2, row.Min.Y+2)
	game.handlePointer(true, row.Min.X+2, row.Min.Y+2)

	if !game.valueDialog.Open {
		t.Fatal("value dialog closed while opening click was still held")
	}
}

func TestSetMassDialogAppliesTypedMassValue(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.ReplaceWorld(world)
	game.openMassValueDialog(1)
	game.valueDialog.Text = "7.5"

	game.applyValueDialog()

	mass, _ := game.World().MassByID(1)
	if mass.Mass != 7.5 {
		t.Fatalf("mass value = %f, want 7.5", mass.Mass)
	}
	if game.valueDialog.Open {
		t.Fatal("value dialog stayed open")
	}
}

func TestSetMassDialogSliderUpdatesText(t *testing.T) {
	game := NewGame()
	game.openMassValueDialog(1)
	track := game.valueDialogSliderTrack()

	game.clickValueDialog(track.Min.X+track.Dx()/2, track.Min.Y)

	if game.valueDialog.Text != "10" {
		t.Fatalf("mass dialog text = %q, want 10", game.valueDialog.Text)
	}
}

func TestValueDialogCursorBlinks(t *testing.T) {
	game := NewGame()
	game.openMassValueDialog(1)

	if !game.valueDialogCursorVisible() {
		t.Fatal("cursor should start visible")
	}
	game.valueDialog.Ticks = valueCursorPeriod
	if game.valueDialogCursorVisible() {
		t.Fatal("cursor should hide after one blink period")
	}
	game.valueDialog.Ticks = valueCursorPeriod * 2
	if !game.valueDialogCursorVisible() {
		t.Fatal("cursor should show after two blink periods")
	}
}

func TestValueDialogTickAdvancesBlinkState(t *testing.T) {
	game := NewGame()
	game.openMassValueDialog(1)

	game.tickValueDialog()

	if game.valueDialog.Ticks != 1 {
		t.Fatalf("dialog ticks = %d, want 1", game.valueDialog.Ticks)
	}
}

func TestRightClickNearSpringOpensSpringContextMenu(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 110, Y: 10}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 100, SpringConstant: 12})
	game.ReplaceWorld(world)

	game.handleRightPointer(true, 60, 12)

	if !game.springMenu.Open {
		t.Fatal("spring context menu was not opened")
	}
	if game.springMenu.SpringID != 3 {
		t.Fatalf("spring menu = %#v", game.springMenu)
	}
}

func TestSpringConstantDialogAppliesValue(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 110, Y: 10}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 100, SpringConstant: 12})
	game.ReplaceWorld(world)
	game.openSpringValueDialog(3, springValueKspring)
	game.valueDialog.Text = "22"

	game.applyValueDialog()

	spring, _ := game.World().SpringByID(3)
	if spring.SpringConstant != 22 || spring.Stiffness != 22 {
		t.Fatalf("spring = %#v, want constant and stiffness 22", spring)
	}
}

func TestSpringContextMenuActionsAndDraw(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 110, Y: 10}, Mass: 1},
	)
	game.World().Springs = []sim.Spring{{ID: 3, MassA: 1, MassB: 2, SpringConstant: 12, Damping: 4, RestLength: 100}}

	labels := game.SpringContextMenuLabelsForSpring(3)
	for _, want := range []string{"Kspring", "Kdamp", "RestLen", "Wall"} {
		if !stringSliceContains(labels, want) {
			t.Fatalf("spring menu labels = %#v, missing %q", labels, want)
		}
	}
	if game.SelectSpringContextMenuItem(3, "missing") {
		t.Fatal("missing spring menu item should not be handled")
	}

	if !game.SelectSpringContextMenuItem(3, "Kdamp") {
		t.Fatal("Kdamp spring menu item was not handled")
	}
	game.valueDialog.Text = "9"
	game.applyValueDialog()
	spring, _ := game.World().SpringByID(3)
	if spring.Damping != 9 {
		t.Fatalf("damping = %f, want 9", spring.Damping)
	}

	if !game.SelectSpringContextMenuItem(3, "RestLen") {
		t.Fatal("RestLen spring menu item was not handled")
	}
	game.valueDialog.Text = "75"
	game.applyValueDialog()
	spring, _ = game.World().SpringByID(3)
	if spring.RestLength != 75 {
		t.Fatalf("rest length = %f, want 75", spring.RestLength)
	}

	if !game.SelectSpringContextMenuItem(3, "Wall") {
		t.Fatal("Wall spring menu item was not handled")
	}
	spring, _ = game.World().SpringByID(3)
	if !spring.Wall {
		t.Fatal("wall menu item did not toggle spring wall state")
	}

	game.springMenu = springContextMenu{Open: true, SpringID: 3, X: 60, Y: 12}
	game.Draw(ebiten.NewImage(screenWidth, screenHeight))
	game.clickSpringContextMenu(game.springContextMenuRowRect(0).Min.X+1, game.springContextMenuRowRect(0).Min.Y+1)
	if !game.valueDialog.Open || game.springMenu.Open {
		t.Fatalf("spring context click dialog=%#v menu=%#v", game.valueDialog, game.springMenu)
	}

	game.springMenu = springContextMenu{Open: true, SpringID: 3, X: 60, Y: 12}
	game.clickSpringContextMenu(screenWidth, screenHeight)
	if game.springMenu.Open {
		t.Fatal("outside spring context click should close menu")
	}
}

func TestValueDialogClickAndHoldBranches(t *testing.T) {
	game := NewGame()
	rect := valueDialogRect()

	game.valueDialog = valueDialog{Open: true, Text: "10", Min: 0, Max: 20}
	game.clickValueDialog(game.valueDialogIncrementRect().Min.X+1, game.valueDialogIncrementRect().Min.Y+1)
	if game.valueDialog.Text != "10.1" || game.activeValueStep != numericStepAmount {
		t.Fatalf("increment text=%q active=%f", game.valueDialog.Text, game.activeValueStep)
	}
	game.valueStepTicks = numericStepHoldDelayTicks - 1
	game.continueValueDialogStepHold()
	if game.valueDialog.Text != "10.2" {
		t.Fatalf("held increment text = %q", game.valueDialog.Text)
	}

	game.clickValueDialog(game.valueDialogDecrementRect().Min.X+1, game.valueDialogDecrementRect().Min.Y+1)
	if game.valueDialog.Text != "10.1" || game.activeValueStep != -numericStepAmount {
		t.Fatalf("decrement text=%q active=%f", game.valueDialog.Text, game.activeValueStep)
	}

	track := game.valueDialogSliderTrack()
	game.clickValueDialog(track.Min.X+track.Dx()/2, track.Min.Y)
	if game.valueDialog.Text != "10" {
		t.Fatalf("slider text = %q, want 10", game.valueDialog.Text)
	}

	applied := 0.0
	game.valueDialog = valueDialog{Open: true, Text: "4", Apply: func(value float64) { applied = value }}
	game.clickValueDialog(game.valueDialogOKRect().Min.X+1, game.valueDialogOKRect().Min.Y+1)
	if applied != 4 || game.valueDialog.Open {
		t.Fatalf("ok click applied=%f dialog=%#v", applied, game.valueDialog)
	}

	game.valueDialog = valueDialog{Open: true}
	game.clickValueDialog(rect.Max.X+1, rect.Max.Y+1)
	if game.valueDialog.Open {
		t.Fatal("outside click should close dialog")
	}

	game.valueDialog = valueDialog{Open: false, Text: "10"}
	game.activeValueStep = numericStepAmount
	game.valueStepTicks = 4
	game.continueValueDialogStepHold()
	if game.activeValueStep != 0 || game.valueStepTicks != 0 {
		t.Fatalf("closed hold state active=%f ticks=%d", game.activeValueStep, game.valueStepTicks)
	}
}

func TestSpringRenderRepresentationsCoverWallState(t *testing.T) {
	game := gameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 110, Y: 10}, Mass: 1},
	)
	game.World().Springs = []sim.Spring{{ID: 3, MassA: 1, MassB: 2}}
	if got := game.RenderWorld().Representations["spring"]; got != "cyan line" {
		t.Fatalf("normal spring representation = %q", got)
	}

	game.World().Springs[0].Wall = true
	report := game.RenderWorld()
	if report.Representations["spring"] != "cyan line" || report.Representations["wall spring"] != "heavy orange line" {
		t.Fatalf("wall spring representations = %#v", report.Representations)
	}
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestRunAndPauseControlsSetSimulationState(t *testing.T) {
	game := NewGame()
	game.SetPaused(false)
	game.ClickVisibleControl("Pause")
	if !game.Paused() {
		t.Fatal("expected Pause click to pause")
	}

	game.ClickVisibleControl("Run")
	if game.Paused() {
		t.Fatal("expected Run click to resume")
	}
}

func TestFileCommandsSaveLoadAndInsertXSP(t *testing.T) {
	game := NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 1, Y: 2}, Mass: 1})
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()

	saved := game.SaveXSP()
	if saved == "" || game.EditorScreen().Indicators["file state"] != "saved" {
		t.Fatalf("saved = %q indicators = %#v", saved, game.EditorScreen().Indicators)
	}

	loaded := "#1.0\ncmas 7\nmass 9 10 20 1 0\n"
	staleEditorWorld := sim.NewWorld()
	game.editing().World = staleEditorWorld
	if err := game.LoadXSP(loaded); err != nil {
		t.Fatal(err)
	}
	if _, ok := game.World().MassByID(9); !ok || game.World().Parameters.Value("current mass") != "7" {
		t.Fatalf("loaded world = %#v", game.World())
	}
	if game.selected {
		t.Fatal("load should clear selected state")
	}
	if game.editor.World != game.simulation || game.editor.World == staleEditorWorld {
		t.Fatal("load should reattach editor world to game simulation")
	}
	if game.EditorScreen().Indicators["file state"] != "saved" {
		t.Fatalf("load indicators = %#v", game.EditorScreen().Indicators)
	}

	game.World().Parameters.Set("current mass", "kept")
	if err := game.InsertXSP("#1.0\ncmas inserted\nmass 10 30 40 1 0\n"); err != nil {
		t.Fatal(err)
	}
	if _, ok := game.World().MassByID(10); !ok || game.World().Parameters.Value("current mass") != "kept" {
		t.Fatalf("inserted world = %#v", game.World())
	}
	if game.EditorScreen().Indicators["file state"] != "unsaved changes" {
		t.Fatalf("insert indicators = %#v", game.EditorScreen().Indicators)
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
	expected := game.World().Clone()
	replaceWithAppTestState(game, "changed")

	game.RestoreState()

	if !reflect.DeepEqual(game.World(), expected) {
		t.Fatalf("restored world = %#v, want %#v", game.World(), expected)
	}
}

func TestRestoreCommandMarksWorldDirty(t *testing.T) {
	game := NewGame()
	game.SaveState()
	game.dirty = false

	game.RunCommand("restore state")

	if !game.dirty {
		t.Fatal("restore command should mark game dirty")
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

func assertStarterObjects(t *testing.T, world *sim.Simulation) {
	t.Helper()
	var fixed, movable int
	for _, mass := range world.Masses {
		if mass.Fixed {
			fixed++
		} else {
			movable++
		}
	}
	if fixed < 1 || movable < 1 || len(world.Springs) < 1 {
		t.Fatalf("starter world fixed=%d movable=%d springs=%d: %#v", fixed, movable, len(world.Springs), world)
	}
}

func withWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	})
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
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
